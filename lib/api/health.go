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
	"bytes"
	"github.com/SENERGY-Platform/event-manager/lib/config"
	"github.com/SENERGY-Platform/event-manager/lib/interfaces"
	jwt_http_router "github.com/SmartEnergyPlatform/jwt-http-router"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const connectivityTestToken = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJjb25uZWN0aXZpdHktdGVzdCJ9.OnihzQ7zwSq0l1Za991SpdsxkktfrdlNl-vHHpYpXQw"

func init() {
	endpoints = append(endpoints, HealthEndpoints)
}

func HealthEndpoints(router *jwt_http_router.Router, config config.Config, ctrl interfaces.Events) {
	router.POST("/health", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		msg, err := ioutil.ReadAll(request.Body)
		log.Println("INFO: /health", err, string(msg))
		writer.WriteHeader(http.StatusOK)
	})

	if config.ConnectivityTest {
		go func() {
			ticker := time.NewTicker(1 * time.Minute)
			for t := range ticker.C {
				log.Println("INFO: connectivity test: " + t.String())
				client := http.Client{
					Timeout: 5 * time.Second,
				}

				req, err := http.NewRequest(
					"POST",
					"http://localhost:"+config.ApiPort+"/health",
					bytes.NewBuffer([]byte("local connection test: "+t.String())),
				)

				if err != nil {
					log.Fatal("FATAL: connection test unable to build request:", err)
				}
				req.Header.Set("Authorization", connectivityTestToken)

				resp, err := client.Do(req)
				if err != nil {
					log.Fatal("FATAL: connection test:", err)
				}
				ioutil.ReadAll(resp.Body)
				resp.Body.Close()
			}
		}()
	}
}
