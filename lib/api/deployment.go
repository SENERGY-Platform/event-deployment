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
	"github.com/SENERGY-Platform/event-deployment/lib/model"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"net/http"
)

func init() {
	endpoints = append(endpoints,
		SetDeploymentEndpoint,
		DeleteDeploymentEndpoint,
	)
}

// SetDeploymentEndpoint godoc
// @Summary      deploy process
// @Description  deploy process, meant for internal use by the process-deployment service, only admins may access this endpoint
// @Tags         deployment
// @Security Bearer
// @Param        message body model.Deployment true "deployment"
// @Success      200
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /process-deployments [PUT]
func SetDeploymentEndpoint(router *http.ServeMux, config config.Config, ctrl interfaces.Events) {
	router.HandleFunc("PUT /process-deployments", func(writer http.ResponseWriter, request *http.Request) {
		token, err := jwt.Parse(request.Header.Get("Authorization"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusUnauthorized)
			return
		}
		if !token.IsAdmin() {
			http.Error(writer, "only admins may use this endpoint", http.StatusUnauthorized)
			return
		}
		var deployment model.Deployment
		err = json.NewDecoder(request.Body).Decode(&deployment)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if deployment.Id == "" {
			http.Error(writer, "missing deployment id", http.StatusBadRequest)
			return
		}
		if deployment.UserId == "" {
			http.Error(writer, "missing deployment userid", http.StatusBadRequest)
			return
		}
		err = ctrl.Deploy(deployment.UserId, deployment.Deployment)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
}

// DeleteDeploymentEndpoint godoc
// @Summary      delete deployment
// @Description  delete deployment, meant for internal use by the process-deployment service, only admins may access this endpoint
// @Tags         deployment
// @Security Bearer
// @Param        deplid path string true "deployment id"
// @Param        userid path string true "user id"
// @Success      200
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /process-deployments/{userid}/{deplid} [DELETE]
func DeleteDeploymentEndpoint(router *http.ServeMux, config config.Config, ctrl interfaces.Events) {
	router.HandleFunc("DELETE /process-deployments/{userid}/{deplid}", func(writer http.ResponseWriter, request *http.Request) {
		userid := request.PathValue("userid")
		deplid := request.PathValue("deplid")
		token, err := jwt.Parse(request.Header.Get("Authorization"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusUnauthorized)
			return
		}
		if !token.IsAdmin() {
			http.Error(writer, "only admins may use this endpoint", http.StatusUnauthorized)
			return
		}
		err = ctrl.Remove(userid, deplid)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
}
