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
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/SENERGY-Platform/event-deployment/lib/interfaces"
	"io"
	"log"
	"net/http"
	"time"
)

const connectivityTestToken = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJjb25uZWN0aXZpdHktdGVzdCJ9.OnihzQ7zwSq0l1Za991SpdsxkktfrdlNl-vHHpYpXQw"

func init() {
	endpoints = append(endpoints, HealthEndpoints)
}

// HealthEndpoints godoc
// @Summary      health
// @Description  check service health
// @Tags         health
// @Security Bearer
// @Success      200
// @Router       /health [POST]
func HealthEndpoints(router *http.ServeMux, config config.Config, ctrl interfaces.Events) {
	router.HandleFunc("POST /health", func(writer http.ResponseWriter, request *http.Request) {
		msg, err := io.ReadAll(request.Body)
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
				io.ReadAll(resp.Body)
				resp.Body.Close()
			}
		}()
	}
}
