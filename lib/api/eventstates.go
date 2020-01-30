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
	"encoding/json"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/SENERGY-Platform/event-deployment/lib/interfaces"
	jwt_http_router "github.com/SmartEnergyPlatform/jwt-http-router"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
)

func init() {
	endpoints = append(endpoints, EventStatesEndpoints)
}

func EventStatesEndpoints(router *jwt_http_router.Router, config config.Config, ctrl interfaces.Events) {

	router.GET("/event-states", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		idstring := request.URL.Query().Get("ids")
		ids := strings.Split(strings.Replace(idstring, " ", "", -1), ",")
		states, err, code := ctrl.GetEventStates(jwt, ids)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(states)
		if err != nil {
			log.Println("ERROR:", err)
			debug.PrintStack()
		}
	})
}
