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

package analytics

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/event-deployment/lib/auth"
	"log"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
)

func ServiceIdToTopic(id string) string {
	id = strings.ReplaceAll(id, "#", "_")
	id = strings.ReplaceAll(id, ":", "_")
	return id
}

func (this *Analytics) Remove(user string, pipelineId string) error {
	client := http.Client{
		Timeout: this.timeout,
	}
	req, err := http.NewRequest(
		"DELETE",
		this.config.FlowEngineUrl+"/pipeline/"+url.PathEscape(pipelineId),
		nil,
	)
	if err != nil {
		debug.PrintStack()
		return err
	}
	req.Header.Set("X-UserId", user)
	resp, err := client.Do(req)
	if err != nil {
		debug.PrintStack()
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		debug.PrintStack()
		return errors.New("unexpected statuscode")
	}
	return nil
}

func (this *Analytics) sendDeployRequest(token auth.AuthToken, user string, request PipelineRequest) (result Pipeline, err error, code int) {
	body, err := json.Marshal(request)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if this.config.Debug {
		log.Println("DEBUG: deploy event pipeline", string(body))
	}
	client := http.Client{
		Timeout: this.timeout,
	}
	req, err := http.NewRequest(
		"POST",
		this.config.FlowEngineUrl+"/pipeline",
		bytes.NewBuffer(body),
	)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}
	token.UseInRequest(req)
	req.Header.Set("X-UserId", user)
	if this.config.Debug {
		log.Println("DEBUG: send analytics deployment with token:", req.Header.Get("Authorization"))
	}
	resp, err := client.Do(req)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		debug.PrintStack()
		return result, errors.New("unexpected statuscode"), resp.StatusCode
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err, http.StatusOK
}

func (this *Analytics) sendUpdateRequest(token auth.AuthToken, user string, request PipelineRequest) (result Pipeline, err error, code int) {
	body, err := json.Marshal(request)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if this.config.Debug {
		log.Println("DEBUG: deploy event pipeline", string(body))
	}
	client := http.Client{
		Timeout: this.timeout,
	}
	req, err := http.NewRequest(
		"PUT",
		this.config.FlowEngineUrl+"/pipeline",
		bytes.NewBuffer(body),
	)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("X-UserId", user)
	token.UseInRequest(req)
	req.Header.Set("X-UserId", user)
	if this.config.Debug {
		log.Println("DEBUG: send analytics deployment with token:", req.Header.Get("Authorization"))
	}
	resp, err := client.Do(req)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		debug.PrintStack()
		return result, errors.New("unexpected statuscode"), resp.StatusCode
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err, http.StatusOK
}
