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
	deploymentmodel "github.com/SENERGY-Platform/process-deployment/lib/model"
	"log"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"time"
)

func (this *Analytics) Deploy(user string, deploymentId string, event deploymentmodel.MsgEvent) (pipelineId string, err error) {
	flowId, ok := this.config.EventOperationFlowMapping[event.Operation]
	if !ok {
		log.Println("WARNING: trying to deploy unknown operation -> ignore ", event.Operation)
		return "", nil
	}
	flowCells, err, code := this.GetFlowInputs(flowId, user)
	if err != nil {
		log.Println("ERROR: unable to get flow inputs", err.Error(), code)
		debug.PrintStack()
		return "", err
	}
	if len(flowCells) != 1 {
		err = errors.New("expect flow to have exact one operator")
		log.Println("ERROR: ", err.Error())
		debug.PrintStack()
		return "", err
	}

	description, err := json.Marshal(EventPipelineDescription{
		DeviceId:      event.Device.Id,
		ServiceId:     event.Service.Id,
		ValuePath:     event.Path,
		OperatorValue: event.Value,
		EventId:       event.EventId,
		DeploymentId:  deploymentId,
	})
	if err != nil {
		debug.PrintStack()
		return "", err
	}

	pipeline, err, code := this.sendDeployRequest(user, PipelineRequest{
		Id:          flowId,
		Name:        event.Label + " (" + event.EventId + ")",
		Description: string(description),
		WindowTime:  0,
		Nodes: []PipelineNode{
			{
				NodeId: flowCells[0].Id,
				Inputs: []NodeInput{{
					DeviceId:  event.Device.Id,
					TopicName: ServiceIdToTopic(event.Service.Id),
					Values: []NodeValue{{
						Name: "value",
						Path: event.Path,
					}},
				}},
				Config: []NodeConfig{
					{
						Name:  "value",
						Value: event.Value,
					},
					{
						Name:  "url",
						Value: this.config.CamundaEventTriggerUrl,
					},
					{
						Name:  "eventId",
						Value: event.EventId,
					},
				},
			},
		},
	})
	if err != nil {
		log.Println("ERROR: unable to deploy pipeline", err.Error(), code)
		debug.PrintStack()
		return "", err
	}
	pipelineId = pipeline.Id.String()
	return pipelineId, nil
}

func ServiceIdToTopic(id string) string {
	id = strings.ReplaceAll(id, "#", "_")
	id = strings.ReplaceAll(id, ":", "_")
	return id
}

func (this *Analytics) Remove(user string, pipelineId string) error {
	client := http.Client{
		Timeout: 5 * time.Second,
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

func (this *Analytics) sendDeployRequest(user string, request PipelineRequest) (result Pipeline, err error, code int) {
	body, err := json.Marshal(request)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	client := http.Client{
		Timeout: 5 * time.Second,
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
	req.Header.Set("X-UserId", user)
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
