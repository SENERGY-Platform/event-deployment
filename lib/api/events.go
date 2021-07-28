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
	"github.com/SENERGY-Platform/event-deployment/lib/api/util"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/SENERGY-Platform/event-deployment/lib/interfaces"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func init() {
	endpoints = append(endpoints, EventsEndpoints)
}

func EventsEndpoints(router *httprouter.Router, config config.Config, ctrl interfaces.Events) {

	router.HEAD("/events/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		code := ctrl.CheckEvent(util.GetAuthToken(request), id)
		writer.WriteHeader(code)
	})
}
