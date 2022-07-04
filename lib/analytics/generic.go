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
)

func (this *Analytics) DeployGenericSource(token auth.AuthToken, label string, user string, desc model.GroupEventDescription, path string, castFrom string, castTo string, castExtensions []model.ConverterExtension) (pipelineId string, err error) {
	flowCells, err, code := this.GetFlowInputs(desc.FlowId, user)
	if err != nil {
		log.Println("ERROR: unable to get flow inputs", err.Error(), code)
		if code == http.StatusNotFound || code == http.StatusForbidden || code == http.StatusUnauthorized {
			log.Println("unable to find flow (ignore deployment)", code, err)
			err = nil
		}
		return "", err
	}
	if len(flowCells) != 1 {
		err = errors.New("expect flow to have exact one operator")
		log.Println("ERROR: ", err.Error())
		debug.PrintStack()
		return "", err
	}

	description, err := json.Marshal(EventPipelineDescription{
		GenericEventSource: desc.GenericEventSource,
		FunctionId:         desc.FunctionId,
		AspectId:           desc.AspectId,
		OperatorValue:      desc.OperatorValue,
		EventId:            desc.EventId,
		DeploymentId:       desc.DeploymentId,
		FlowId:             desc.FlowId,
	})
	if err != nil {
		debug.PrintStack()
		return "", err
	}

	inputs := []NodeInput{}

	inputs = append(inputs, NodeInput{
		FilterIds:  desc.GenericEventSource.FilterIds,
		FilterType: desc.GenericEventSource.FilterType,
		TopicName:  desc.GenericEventSource.Topic,
		Values: []NodeValue{{
			Name: "value",
			Path: path,
		}},
	})

	convertFrom := ""
	convertTo := ""
	converterUrl := ""
	extendedConverterUrl := ""
	castExtensionsJson := ""
	if castFrom != castTo && castFrom != "" && castTo != "" {
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

	request := PipelineRequest{
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
