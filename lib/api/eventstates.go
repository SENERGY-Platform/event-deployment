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
	"github.com/SENERGY-Platform/event-deployment/lib/api/util"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/SENERGY-Platform/event-deployment/lib/interfaces"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
)

func init() {
	endpoints = append(endpoints, EventStatesEndpoints)
}

type EventStates = map[string]bool

// EventStatesEndpoints godoc
// @Summary      get event-states
// @Description  get event-states
// @Tags         event
// @Produce      json
// @Security Bearer
// @Param        ids query string true "comma seperated list of event-ids"
// @Success      200 {object} EventStates
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /event-states [GET]
func EventStatesEndpoints(router *http.ServeMux, config config.Config, ctrl interfaces.Events) {
	router.HandleFunc("GET /event-states", func(writer http.ResponseWriter, request *http.Request) {
		idstring := strings.TrimSpace(request.URL.Query().Get("ids"))
		ids := []string{}
		if idstring != "" {
			ids = strings.Split(strings.Replace(idstring, " ", "", -1), ",")
		}
		states, err, code := ctrl.GetEventStates(util.GetAuthToken(request), ids)
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
