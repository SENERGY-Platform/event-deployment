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
	deploymentmodel "github.com/SENERGY-Platform/process-deployment/lib/model"
	jwt_http_router "github.com/SmartEnergyPlatform/jwt-http-router"
	"log"
	"net/http"
	"runtime/debug"
)

type EventsFactory struct{}

var Factory = &EventsFactory{}

type Events struct {
	config    config.Config
	analytics interfaces.Analytics
}

func (this *EventsFactory) New(ctx context.Context, config config.Config, analytics interfaces.Analytics) (interfaces.Events, error) {
	return &Events{config: config, analytics: analytics}, nil
}

func (this *Events) HandleCommand(msg []byte) error {
	if this.config.Debug {
		log.Println("DEBUG: receive deployment command:", string(msg))
	}
	cmd := deploymentmodel.DeploymentCommand{}
	err := json.Unmarshal(msg, &cmd)
	if err != nil {
		debug.PrintStack()
		return err
	}
	switch cmd.Command {
	case "PUT":
		return this.Deploy(cmd.Owner, cmd.Deployment)
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
	events := this.deploymentToMsgEvents(deployment)
	for _, event := range events {
		pipelineId, err := this.analytics.Deploy(owner, deployment.Id, event)
		if err != nil {
			return err
		}
		if pipelineId == "" {
			log.Println("WARNING: event not deployed in analytics -> ignore event", event)
			return nil
		}
		if err != nil {
			return err
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

func (this *Events) deploymentToMsgEvents(deployment deploymentmodel.Deployment) (result []deploymentmodel.MsgEvent) {
	for _, lane := range deployment.Lanes {
		if lane.Lane != nil {
			for _, element := range lane.Lane.Elements {
				if element.MsgEvent != nil {
					result = append(result, *element.MsgEvent)
				}
				if element.ReceiveTaskEvent != nil {
					result = append(result, *element.ReceiveTaskEvent)
				}
			}
		}
		if lane.MultiLane != nil {
			for _, element := range lane.MultiLane.Elements {
				if element.MsgEvent != nil {
					result = append(result, *element.MsgEvent)
				}
				if element.ReceiveTaskEvent != nil {
					result = append(result, *element.ReceiveTaskEvent)
				}
			}
		}
	}
	for _, element := range deployment.Elements {
		if element.MsgEvent != nil {
			result = append(result, *element.MsgEvent)
		}
		if element.ReceiveTaskEvent != nil {
			result = append(result, *element.ReceiveTaskEvent)
		}
	}
	return result
}
