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
	"runtime/debug"
	"strings"
)

func (this *Analytics) DeployDeviceWithMarshaller(token auth.AuthToken, label string, user string, deploymentId string, flowId string, eventId string, deviceId string, serviceId string, value string, path string, functionId string, aspectNodeId string) (pipelineId string, err error) {
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
		UseMarshaller: true,
	})
	if err != nil {
		debug.PrintStack()
		return "", err
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
						Path: strings.TrimSuffix(this.config.DevicePathPrefix, "."),
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
						Name:  "marshallerUrl",
						Value: this.config.MarshallerUrl,
					},
					{
						Name:  "path",
						Value: path,
					},
					{
						Name:  "functionId",
						Value: functionId,
					},
					{
						Name:  "aspectNodeId",
						Value: aspectNodeId,
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

func (this *Analytics) DeployDevice(token auth.AuthToken, label string, user string, deploymentId string, flowId string, eventId string, deviceId string, serviceId string, value string, path string, castFrom string, castTo string, castExtensions []model.ConverterExtension) (pipelineId string, err error) {
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
	extendedConverterUrl := ""
	castExtensionsJson := ""
	if castFrom != castTo {
		converterUrl = this.config.ConverterUrl
		extendedConverterUrl = this.config.ExtendedConverterUrl
		convertFrom = castFrom
		convertTo = castTo
		if len(castExtensions) > 0 {
			castExtensionsJsonTemp, err := json.Marshal(castExtensions)
			if err != nil {
				debug.PrintStack()
				return "", err
			}
			castExtensionsJson = string(castExtensionsJsonTemp)
		}
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
						Name:  "extendedConverterUrl",
						Value: extendedConverterUrl,
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
						Name:  "castExtensions",
						Value: castExtensionsJson,
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
