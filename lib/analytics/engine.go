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
	"log"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"time"
)

func (this *Analytics) Deploy(label string, user string, deploymentId string, flowId string, eventId string, deviceId string, serviceId string, value string, path string, castFrom string, castTo string) (pipelineId string, err error) {
	shard, err := this.shards.GetShardForUser(user)
	if err != nil {
		return "", err
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
		DeviceId:      deviceId,
		ServiceId:     serviceId,
		ValuePath:     path,
		OperatorValue: value,
		EventId:       eventId,
		DeploymentId:  deploymentId,
	})
	if err != nil {
		debug.PrintStack()
		return "", err
	}

	convertFrom := ""
	convertTo := ""
	converterUrl := ""
	if castFrom != castTo {
		converterUrl = this.config.ConverterUrl
		convertFrom = castFrom
		convertTo = castTo
	}

	pipeline, err, code := this.sendDeployRequest(user, PipelineRequest{
		FlowId:      flowId,
		Name:        label,
		Description: string(description),
		WindowTime:  0,
		Nodes: []PipelineNode{
			{
				NodeId: flowCells[0].Id,
				Inputs: []NodeInput{{
					DeviceId:  deviceId,
					TopicName: ServiceIdToTopic(serviceId),
					Values: []NodeValue{{
						Name: "value",
						Path: path,
					}},
				}},
				Config: []NodeConfig{
					{
						Name:  "value",
						Value: value,
					},
					{
						Name:  "url",
						Value: shard + this.config.CamundaEventTriggerPath,
					},
					{
						Name:  "eventId",
						Value: eventId,
					},
					{
						Name:  "converterUrl",
						Value: converterUrl,
					},
					{
						Name:  "convertFrom",
						Value: convertFrom,
					},
					{
						Name:  "convertTo",
						Value: convertTo,
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

func (this *Analytics) DeployGroup(label string, user string, deploymentId string, flowId string, eventId string, groupId string, value string, serviceIds []string, serviceToDeviceIdsMapping map[string][]string, serviceToPathMapping map[string]string) (pipelineId string, err error) {
	shard, err := this.shards.GetShardForUser(user)
	if err != nil {
		return "", err
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
		DeviceGroupId: groupId,
		OperatorValue: value,
		EventId:       eventId,
		DeploymentId:  deploymentId,
	})
	if err != nil {
		debug.PrintStack()
		return "", err
	}

	inputs := []NodeInput{}
	for _, serviceId := range serviceIds {
		deviceIds := strings.Join(serviceToDeviceIdsMapping[serviceId], ",")
		if deviceIds == "" {
			log.Println("WARNING: missing deviceIds for service in DeployGroup()", serviceId, " --> skip service for group event deployment")
			continue
		}
		path := serviceToPathMapping[serviceId]
		if path == "" {
			log.Println("WARNING: missing path for service in DeployGroup()", serviceId, " --> skip service for group event deployment")
			continue
		}
		inputs = append(inputs, NodeInput{
			DeviceId:  deviceIds,
			TopicName: ServiceIdToTopic(serviceId),
			Values: []NodeValue{{
				Name: "value",
				Path: path,
			}},
		})
	}

	pipeline, err, code := this.sendDeployRequest(user, PipelineRequest{
		FlowId:      flowId,
		Name:        label,
		Description: string(description),
		WindowTime:  0,
		Nodes: []PipelineNode{
			{
				NodeId: flowCells[0].Id,
				Inputs: inputs,
				Config: []NodeConfig{
					{
						Name:  "value",
						Value: value,
					},
					{
						Name:  "url",
						Value: shard + this.config.CamundaEventTriggerPath,
					},
					{
						Name:  "eventId",
						Value: eventId,
					},
					{
						Name:  "converterUrl",
						Value: "",
					},
					{
						Name:  "convertFrom",
						Value: "",
					},
					{
						Name:  "convertTo",
						Value: "",
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
	if this.config.Debug {
		log.Println("DEBUG: deploy event pipeline", string(body))
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
