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
	"github.com/SENERGY-Platform/event-deployment/lib/auth"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/SENERGY-Platform/event-deployment/lib/events/analyticsevents"
	"github.com/SENERGY-Platform/event-deployment/lib/events/conditionalevents"
	"github.com/SENERGY-Platform/event-deployment/lib/interfaces"
	"github.com/SENERGY-Platform/event-deployment/lib/metrics"
	"github.com/SENERGY-Platform/event-deployment/lib/model"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel"
	"github.com/SENERGY-Platform/process-deployment/lib/model/messages"
	"log"
	"net/http"
	"runtime/debug"
)

type EventsFactory struct{}

var Factory = &EventsFactory{}

type Events struct {
	config       config.Config
	handlers     []Handler
	doneProducer interfaces.Producer
	metrics      *metrics.Metrics
}

type Handler interface {
	GetEventStates(token string, ids []string) (states map[string]bool, err error, code int)
	CheckEvent(token string, id string) int
	Remove(owner string, deploymentId string) error
	Deploy(owner string, deployment deploymentmodel.Deployment) error
	UpdateDeviceGroup(owner string, group model.DeviceGroup) error
}

func (this *EventsFactory) New(ctx context.Context, config config.Config, analytics interfaces.Analytics, devices interfaces.Devices, imports interfaces.Imports, doneProducer interfaces.Producer, m *metrics.Metrics) (result interfaces.Events, err error) {
	handlers := []Handler{}
	if config.EnableAnalyticsEvents {
		analyticsEvents, err := analyticsevents.New(ctx, config, analytics, devices, imports, m)
		if err != nil {
			return nil, err
		}
		handlers = append(handlers, analyticsEvents)
	}
	if config.ConditionalEventRepoMongoUrl != "" && config.ConditionalEventRepoMongoUrl != "-" {
		conditionalEvents, err := conditionalevents.New(ctx, config, devices, imports, m)
		if err != nil {
			return nil, err
		}
		handlers = append(handlers, conditionalEvents)
	}
	return &Events{config: config, handlers: handlers, doneProducer: doneProducer, metrics: m}, err
}

type VersionWrapper struct {
	Command string `json:"command"`
	Id      string `json:"id"`
	Version int64  `json:"version"`
	Owner   string `json:"owner"`
}

func (this *Events) HandleCommand(msg []byte) error {
	if this.config.Debug {
		log.Println("DEBUG: receive deployment command:", string(msg))
	}

	version := VersionWrapper{}
	err := json.Unmarshal(msg, &version)
	if err != nil {
		log.Println("ERROR: consumed invalid message --> ignore", err)
		debug.PrintStack()
		return nil
	}
	if version.Version != deploymentmodel.CurrentVersion {
		log.Println("ERROR: consumed unexpected deployment version", version.Version)
		if version.Command == "DELETE" {
			log.Println("handle legacy delete")
			return this.Remove(version.Owner, version.Id)
		}
		return nil
	}

	cmd := messages.DeploymentCommand{}
	err = json.Unmarshal(msg, &cmd)
	if err != nil {
		log.Println("ERROR: invalid message --> ignore", err)
		debug.PrintStack()
		return nil
	}
	switch cmd.Command {
	case "RIGHTS":
		return nil
	case "PUT":
		if cmd.Version != deploymentmodel.CurrentVersion {
			log.Println("ERROR: unexpected deployment version", cmd.Version)
			return nil
		}
		if cmd.Owner == "" {
			log.Printf("ERROR: missing owner --> ignore deployment command %#v\n", cmd)
			return nil
		}
		if cmd.Deployment != nil {
			err = this.Deploy(cmd.Owner, *cmd.Deployment)
		}
		if errors.Is(err, auth.ErrUserDoesNotExist) {
			log.Printf("WARNING: user %v does not exist -> DEPLOYMENT WILL BE IGNORED\n", cmd.Owner)
			return nil
		}
		return err
	case "DELETE":
		if cmd.Owner == "" {
			log.Printf("ERROR: missing owner --> ignore deployment delete command %#v\n", cmd)
			return nil
		}
		err = this.Remove(cmd.Owner, cmd.Id)
		if errors.Is(err, auth.ErrUserDoesNotExist) {
			log.Printf("WARNING: user %v does not exist -> DEPLOYMENT WILL BE IGNORED\n", cmd.Owner)
			return nil
		}
		return err
	default:
		return errors.New("unknown command " + cmd.Command)
	}
	return nil
}

func (this *Events) Deploy(owner string, deployment deploymentmodel.Deployment) (err error) {
	for _, h := range this.handlers {
		err = h.Deploy(owner, deployment)
		if err != nil {
			return err
		}
	}
	this.metrics.DeployedProcesses.Inc()
	this.notifyProcessDeploymentDone(deployment.Id)
	return nil
}

func (this *Events) Remove(owner string, deploymentId string) (err error) {
	for _, h := range this.handlers {
		err = h.Remove(owner, deploymentId)
		if err != nil {
			return err
		}
	}
	this.metrics.RemovedProcesses.Inc()
	return nil
}

func (this *Events) CheckEvent(token string, id string) (result int) {
	for _, h := range this.handlers {
		result = h.CheckEvent(token, id)
		if result == http.StatusOK || result == http.StatusBadRequest || result == http.StatusInternalServerError {
			return result
		}
	}
	return http.StatusNotFound
}

func (this *Events) GetEventStates(token string, ids []string) (states map[string]bool, err error, code int) {
	states = map[string]bool{}
	for _, h := range this.handlers {
		temp, err, code := h.GetEventStates(token, ids)
		if err != nil {
			return states, err, code
		}
		for key, value := range temp {
			if !states[key] {
				states[key] = value
			}
		}
	}
	return states, nil, http.StatusOK
}

func (this *Events) notifyProcessDeploymentDone(id string) {
	if this.doneProducer != nil {
		message := DoneNotification{
			Command: "PUT",
			Id:      id,
			Handler: "github.com/SENERGY-Platform/event-deployment",
		}
		log.Println("send deployment done", message)
		msg, err := json.Marshal(message)
		if err != nil {
			log.Println("ERROR:", err)
			debug.PrintStack()
			return
		}
		err = this.doneProducer.Produce(id, msg)
		if err != nil {
			log.Println("ERROR:", err)
			debug.PrintStack()
			return
		}
	}
}

type DoneNotification struct {
	Command string `json:"command"`
	Id      string `json:"id"`
	Handler string `json:"handler"`
}
