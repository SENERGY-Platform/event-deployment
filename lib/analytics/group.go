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
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/event-deployment/lib/auth"
	"github.com/SENERGY-Platform/event-deployment/lib/model"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
)

func (this *Analytics) DeployGroup(token auth.AuthToken, label string, user string, desc model.GroupEventDescription, serviceIds []string, serviceToDeviceIdsMapping map[string][]string, serviceToPathsMapping map[string][]string, serviceToPathAndCharacteristic map[string][]model.PathAndCharacteristic, castExtensions []model.ConverterExtension, useMarshaller bool) (pipelineId string, err error) {
	if this.config.Debug {
		log.Println("DEBUG: DeployGroup()")
	}
	request, err := this.getPipelineRequestForGroupDeployment(token, label, user, desc, serviceIds, serviceToDeviceIdsMapping, serviceToPathsMapping, serviceToPathAndCharacteristic, castExtensions, useMarshaller)
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

func (this *Analytics) UpdateGroupDeployment(token auth.AuthToken, pipelineId string, label string, user string, desc model.GroupEventDescription, serviceIds []string, serviceToDeviceIdsMapping map[string][]string, serviceToPathsMapping map[string][]string, serviceToPathAndCharacteristic map[string][]model.PathAndCharacteristic, castExtensions []model.ConverterExtension, useMarshaller bool) (err error) {
	request, err := this.getPipelineRequestForGroupDeployment(token, label, user, desc, serviceIds, serviceToDeviceIdsMapping, serviceToPathsMapping, serviceToPathAndCharacteristic, castExtensions, useMarshaller)
	if err != nil {
		return err
	}
	request.Id = pipelineId
	_, err, code := this.sendUpdateRequest(token, user, request)
	if err != nil {
		log.Println("ERROR: unable to deploy pipeline", err.Error(), code)
		debug.PrintStack()
		return err
	}
	return nil
}

func (this *Analytics) getPipelineRequestForGroupDeployment(token auth.AuthToken, label string, user string, desc model.GroupEventDescription, serviceIds []string, serviceToDeviceIdsMapping map[string][]string, serviceToPathsMapping map[string][]string, serviceToPathAndCharacteristic map[string][]model.PathAndCharacteristic, castExtensions []model.ConverterExtension, useMarshaller bool) (request PipelineRequest, err error) {
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
		UseMarshaller: desc.UseMarshaller,
	})
	if err != nil {
		debug.PrintStack()
		return request, err
	}

	inputs := []NodeInput{}
	for _, serviceId := range serviceIds {
		deviceIdList := []string{}
		for _, id := range serviceToDeviceIdsMapping[serviceId] {
			deviceIdList = append(deviceIdList, trimIdParams(id))
		}
		deviceIds := strings.Join(deviceIdList, ",")
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
		if useMarshaller {
			values = []NodeValue{{
				Name: "value",
				Path: strings.TrimSuffix(this.config.GroupPathPrefix, "."),
			}}
		} else {
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

	castExtensionsJson := ""
	if len(castExtensions) > 0 {
		castExtensionsJsonTemp, err := json.Marshal(castExtensions)
		if err != nil {
			debug.PrintStack()
			return request, err
		}
		castExtensionsJson = string(castExtensionsJsonTemp)
	}

	if useMarshaller {
		topicToServiceId := map[string]string{}
		for _, id := range serviceIds {
			topicToServiceId[ServiceIdToTopic(id)] = id
		}
		topicToServiceIdJson, err := json.Marshal(topicToServiceId)
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
							Name:  "marshallerUrl",
							Value: this.config.MarshallerUrl,
						},
						{
							Name:  "functionId",
							Value: desc.FunctionId,
						},
						{
							Name:  "aspectNodeId",
							Value: desc.AspectId,
						},
						{
							Name:  "targetCharacteristicId",
							Value: desc.CharacteristicId,
						},
						{
							Name:  "topicToServiceId",
							Value: string(topicToServiceIdJson),
						},
						{
							Name:  "userToken",
							Value: string(token),
						},
					},
				},
			},
		}, nil
	} else {
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
							Name:  "extendedConverterUrl",
							Value: this.config.ExtendedConverterUrl,
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
							Name:  "castExtensions",
							Value: castExtensionsJson,
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
}
