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

package events

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/SENERGY-Platform/event-deployment/lib/interfaces"
	"github.com/SENERGY-Platform/event-deployment/lib/marshaller"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel/v2"
	"github.com/SENERGY-Platform/process-deployment/lib/model/messages"
	jwt_http_router "github.com/SmartEnergyPlatform/jwt-http-router"
	"log"
	"net/http"
	"runtime/debug"
)

type EventsFactory struct{}

var Factory = &EventsFactory{}

type Events struct {
	config     config.Config
	analytics  interfaces.Analytics
	marshaller interfaces.Marshaller
}

func (this *EventsFactory) New(ctx context.Context, config config.Config, analytics interfaces.Analytics, marshaller interfaces.Marshaller) (interfaces.Events, error) {
	return &Events{config: config, analytics: analytics, marshaller: marshaller}, nil
}

func (this *Events) HandleCommand(msg []byte) error {
	if this.config.Debug {
		log.Println("DEBUG: receive deployment command:", string(msg))
	}
	cmd := messages.DeploymentCommand{}
	err := json.Unmarshal(msg, &cmd)
	if err != nil {
		debug.PrintStack()
		return err
	}
	switch cmd.Command {
	case "PUT":
		if cmd.DeploymentV2 != nil {
			err = this.Deploy(cmd.Owner, *cmd.DeploymentV2)
		}
		return err
	case "DELETE":
		return this.Remove(cmd.Owner, cmd.Id)
	default:
		return errors.New("unknown command " + cmd.Command)
	}
	return nil
}

func (this *Events) Deploy(owner string, deployment deploymentmodel.Deployment) error {
	err := this.Remove(owner, deployment.Id)
	if err != nil {
		return err
	}
	for _, element := range deployment.Elements {
		err = this.deplayElement(owner, deployment.Id, element)
	}
	return nil
}

func (this *Events) deplayElement(owner string, deploymentId string, element deploymentmodel.Element) (err error) {
	event := element.MessageEvent
	if event != nil && event.Selection.FilterCriteria.CharacteristicId != nil {
		path, characteristicId, err := this.GetPathAndCharacteristicForEvent(event)
		if err == ErrMissingCharacteristicInEvent || err == marshaller.ErrCharacteristicNotFoundInService || err == marshaller.ErrServiceNotFound {
			log.Println("WARNING: error on marshaller request;", err, "-> ignore event", event)
			return nil
		}
		if err != nil {
			return err
		}
		pipelineId, err := this.analytics.Deploy(
			element.Name+" ("+event.EventId+")",
			owner,
			deploymentId,
			event.FlowId,
			event.EventId,
			event.Selection.SelectedDeviceId,
			event.Selection.SelectedServiceId,
			event.Value,
			"value."+path,
			characteristicId,
			*event.Selection.FilterCriteria.CharacteristicId)
		if err != nil {
			return err
		}
		if pipelineId == "" {
			log.Println("WARNING: event not deployed in analytics -> ignore event", event)
			return nil
		}
	}
	return nil
}

func (this *Events) Remove(owner string, deploymentId string) error {
	pipelineIds, err := this.analytics.GetPipelinesByDeploymentId(owner, deploymentId)
	if err != nil {
		return err
	}
	for _, id := range pipelineIds {
		err = this.analytics.Remove(owner, id)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *Events) CheckEvent(jwt jwt_http_router.Jwt, id string) int {
	_, exists, err := this.analytics.GetPipelineByEventId(jwt.UserId, id)
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return http.StatusInternalServerError
	}
	if !exists {
		return http.StatusNotFound
	}
	return http.StatusOK
}

func (this *Events) GetEventStates(jwt jwt_http_router.Jwt, ids []string) (states map[string]bool, err error, code int) {
	states = map[string]bool{}
	if len(ids) == 0 {
		return states, nil, http.StatusOK
	}
	states, err = this.analytics.GetEventStates(jwt.UserId, ids)
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return states, err, http.StatusInternalServerError
	}
	return states, nil, http.StatusOK
}

var ErrMissingCharacteristicInEvent = errors.New("missing characteristic id in event")

func (this *Events) GetPathAndCharacteristicForEvent(event *deploymentmodel.MessageEvent) (path string, characteristicId string, err error) {
	if event.Selection.FilterCriteria.CharacteristicId == nil {
		return "", "", ErrMissingCharacteristicInEvent
	}
	return this.marshaller.FindPath(event.Selection.SelectedServiceId, *event.Selection.FilterCriteria.CharacteristicId)
}
