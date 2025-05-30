/*
 * Copyright 2020 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package api

import (
	"context"
	"github.com/SENERGY-Platform/event-deployment/lib/api/util"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/SENERGY-Platform/event-deployment/lib/interfaces"
	"github.com/SENERGY-Platform/service-commons/pkg/accesslog"
	"log"
	"net/http"
	"reflect"
	"runtime"
	"time"
)

//go:generate go tool swag init -o ../../docs --parseDependency -d .. -g api/api.go

var endpoints []func(*http.ServeMux, config.Config, interfaces.Events)

func Start(ctx context.Context, config config.Config, ctrl interfaces.Events) error {
	log.Println("start api")
	router := Router(config, ctrl)
	timeout, err := time.ParseDuration(config.HttpServerTimeout)
	if err != nil {
		log.Println("WARNING: invalid http server timeout --> no timeouts\n", err)
		err = nil
	}

	readtimeout, err := time.ParseDuration(config.HttpServerReadTimeout)
	if err != nil {
		log.Println("WARNING: invalid http server read timeout --> no timeouts\n", err)
		err = nil
	}
	server := &http.Server{Addr: ":" + config.ApiPort, Handler: router, WriteTimeout: timeout, ReadTimeout: readtimeout}
	go func() {
		log.Println("Listening on ", server.Addr)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Println("ERROR: api server error", err)
			log.Fatal(err)
		}
	}()
	go func() {
		<-ctx.Done()
		log.Println("DEBUG: api shutdown", server.Shutdown(context.Background()))
	}()
	return nil
}

// Router doc
// @title         Event-Deployment
// @version       0.1
// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath  /
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func Router(config config.Config, ctrl interfaces.Events) http.Handler {
	router := http.NewServeMux()
	for _, e := range endpoints {
		log.Println("add endpoints: " + runtime.FuncForPC(reflect.ValueOf(e).Pointer()).Name())
		e(router, config, ctrl)
	}
	log.Println("add logging and cors")
	corsHandler := util.NewCors(router)
	return accesslog.New(corsHandler)
}
