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
	"github.com/SENERGY-Platform/event-deployment/lib/model"
	"log"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
)

func (this *Analytics) Deploy(token auth.AuthToken, label string, user string, deploymentId string, flowId string, eventId string, deviceId string, serviceId string, value string, path string, castFrom string, castTo string) (pipelineId string, err error) {
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

	pipeline, err, code := this.sendDeployRequest(token, user, PipelineRequest{
		FlowId:      flowId,
		Name:        label,
		Description: string(description),
		WindowTime:  0,
		Nodes: []PipelineNode{
			{
				NodeId: flowCells[0].Id,
				Inputs: []NodeInput{{
					FilterIds:  deviceId,
					FilterType: DeviceFilterType,
					TopicName:  ServiceIdToTopic(serviceId),
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
						Value: this.config.EventTriggerUrl,
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
					{
						Name:  "userToken",
						Value: string(token),
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

func (this *Analytics) DeployGroup(token auth.AuthToken, label string, user string, desc model.GroupEventDescription, serviceIds []string, serviceToDeviceIdsMapping map[string][]string, serviceToPathsMapping map[string][]string, serviceToPathAndCharacteristic map[string][]model.PathAndCharacteristic) (pipelineId string, err error) {
	if this.config.Debug {
		log.Println("DEBUG: DeployGroup()")
	}
	request, err := this.getPipelineRequestForGroupDeployment(token, label, user, desc, serviceIds, serviceToDeviceIdsMapping, serviceToPathsMapping, serviceToPathAndCharacteristic)
	if err != nil {
		log.Println("ERROR: getPipelineRequestForGroupDeployment()", err.Error())
		debug.PrintStack()
		return "", err
	}
	pipeline, err, code := this.sendDeployRequest(token, user, request)
	if err != nil {
		log.Println("ERROR: unable to deploy pipeline", err.Error(), code)
		debug.PrintStack()
		return "", err
	}
	pipelineId = pipeline.Id.String()
	return pipelineId, nil
}

func (this *Analytics) UpdateGroupDeployment(token auth.AuthToken, pipelineId string, label string, user string, desc model.GroupEventDescription, serviceIds []string, serviceToDeviceIdsMapping map[string][]string, serviceToPathsMapping map[string][]string, serviceToPathAndCharacteristic map[string][]model.PathAndCharacteristic) (err error) {
	request, err := this.getPipelineRequestForGroupDeployment(token, label, user, desc, serviceIds, serviceToDeviceIdsMapping, serviceToPathsMapping, serviceToPathAndCharacteristic)
	if err != nil {
		return err
	}
	request.Id = pipelineId
	_, err, code := this.sendUpdateRequest(user, request)
	if err != nil {
		log.Println("ERROR: unable to deploy pipeline", err.Error(), code)
		debug.PrintStack()
		return err
	}
	return nil
}

func (this *Analytics) getPipelineRequestForGroupDeployment(token auth.AuthToken, label string, user string, desc model.GroupEventDescription, serviceIds []string, serviceToDeviceIdsMapping map[string][]string, serviceToPathsMapping map[string][]string, serviceToPathAndCharacteristic map[string][]model.PathAndCharacteristic) (request PipelineRequest, err error) {
	flowCells, err, code := this.GetFlowInputs(desc.FlowId, user)
	if err != nil {
		log.Println("ERROR: unable to get flow inputs", err.Error(), code)
		if code == http.StatusNotFound || code == http.StatusForbidden || code == http.StatusUnauthorized {
			log.Println("unable to find flow (ignore deployment)", code, err)
			err = nil
		}
		return request, err
	}
	if len(flowCells) != 1 {
		err = errors.New("expect flow to have exact one operator")
		log.Println("ERROR: ", err.Error())
		debug.PrintStack()
		return request, err
	}

	description, err := json.Marshal(EventPipelineDescription{
		DeviceGroupId: desc.DeviceGroupId,
		FunctionId:    desc.FunctionId,
		AspectId:      desc.AspectId,
		OperatorValue: desc.OperatorValue,
		EventId:       desc.EventId,
		DeploymentId:  desc.DeploymentId,
		FlowId:        desc.FlowId,
	})
	if err != nil {
		debug.PrintStack()
		return request, err
	}

	inputs := []NodeInput{}
	for _, serviceId := range serviceIds {
		deviceIds := strings.Join(serviceToDeviceIdsMapping[serviceId], ",")
		if deviceIds == "" {
			log.Println("WARNING: missing deviceIds for service in DeployGroup()", serviceId, " --> skip service for group event deployment")
			continue
		}
		paths := serviceToPathsMapping[serviceId]
		if len(paths) == 0 {
			log.Println("WARNING: missing path for service in DeployGroup()", serviceId, " --> skip service for group event deployment")
			continue
		}
		values := []NodeValue{}
		if this.config.EnableMultiplePaths {
			for _, path := range paths {
				values = append(values, NodeValue{
					Name: "value",
					Path: this.config.GroupPathPrefix + path,
				})
			}
		} else {
			values = []NodeValue{{
				Name: "value",
				Path: this.config.GroupPathPrefix + paths[0],
			}}
		}

		inputs = append(inputs, NodeInput{
			FilterIds:  deviceIds,
			FilterType: DeviceFilterType,
			TopicName:  ServiceIdToTopic(serviceId),
			Values:     values,
		})
	}

	topicToPathAndCharacteristic := map[string][]model.PathAndCharacteristic{}
	for serviceId, list := range serviceToPathAndCharacteristic {
		topicToPathAndCharacteristic[ServiceIdToTopic(serviceId)] = list
	}
	topicToPathAndCharacteristicStr, err := json.Marshal(topicToPathAndCharacteristic)
	if err != nil {
		debug.PrintStack()
		return request, err
	}

	return PipelineRequest{
		FlowId:      desc.FlowId,
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
						Value: desc.OperatorValue,
					},
					{
						Name:  "url",
						Value: this.config.EventTriggerUrl,
					},
					{
						Name:  "eventId",
						Value: desc.EventId,
					},
					{
						Name:  "converterUrl",
						Value: this.config.ConverterUrl,
					},
					{
						Name:  "convertFrom",
						Value: "",
					},
					{
						Name:  "convertTo",
						Value: desc.CharacteristicId,
					},
					{
						Name:  "topicToPathAndCharacteristic",
						Value: string(topicToPathAndCharacteristicStr),
					},
					{
						Name:  "userToken",
						Value: string(token),
					},
				},
			},
		},
	}, nil
}

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

func (this *Analytics) sendUpdateRequest(user string, request PipelineRequest) (result Pipeline, err error, code int) {
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

func (this *Analytics) DeployImport(token auth.AuthToken, label string, user string, desc model.GroupEventDescription, topic string, path string, castFrom string, castTo string) (pipelineId string, err error) {
	request, err := this.getPipelineRequestForImportDeployment(token, label, user, desc, topic, path, castFrom, castTo)
	if err != nil {
		return "", err
	}
	pipeline, err, code := this.sendDeployRequest(token, user, request)
	if err != nil {
		log.Println("ERROR: unable to deploy pipeline", err.Error(), code)
		debug.PrintStack()
		return "", err
	}
	pipelineId = pipeline.Id.String()
	return pipelineId, nil
}

func (this *Analytics) getPipelineRequestForImportDeployment(token auth.AuthToken, label string, user string, desc model.GroupEventDescription, topic string, path string, castFrom string, castTo string) (request PipelineRequest, err error) {
	flowCells, err, code := this.GetFlowInputs(desc.FlowId, user)
	if err != nil {
		log.Println("ERROR: unable to get flow inputs", err.Error(), code)
		if code == http.StatusNotFound || code == http.StatusForbidden || code == http.StatusUnauthorized {
			log.Println("unable to find flow (ignore deployment)", code, err)
			err = nil
		}
		return request, err
	}
	if len(flowCells) != 1 {
		err = errors.New("expect flow to have exact one operator")
		log.Println("ERROR: ", err.Error())
		debug.PrintStack()
		return request, err
	}

	description, err := json.Marshal(EventPipelineDescription{
		ImportId:      desc.ImportId,
		FunctionId:    desc.FunctionId,
		AspectId:      desc.AspectId,
		OperatorValue: desc.OperatorValue,
		EventId:       desc.EventId,
		DeploymentId:  desc.DeploymentId,
		FlowId:        desc.FlowId,
	})
	if err != nil {
		debug.PrintStack()
		return request, err
	}

	inputs := []NodeInput{}

	inputs = append(inputs, NodeInput{
		FilterIds:  desc.ImportId,
		FilterType: ImportFilterType,
		TopicName:  topic,
		Values: []NodeValue{{
			Name: "value",
			Path: path,
		}},
	})

	convertFrom := ""
	convertTo := ""
	converterUrl := ""
	if castFrom != castTo {
		converterUrl = this.config.ConverterUrl
		convertFrom = castFrom
		convertTo = castTo
	}

	return PipelineRequest{
		FlowId:      desc.FlowId,
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
						Value: desc.OperatorValue,
					},
					{
						Name:  "url",
						Value: this.config.EventTriggerUrl,
					},
					{
						Name:  "eventId",
						Value: desc.EventId,
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
					{
						Name:  "userToken",
						Value: string(token),
					},
				},
			},
		},
	}, nil
}
