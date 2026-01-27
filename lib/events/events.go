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
	"net/http"
	"runtime/debug"

	"github.com/SENERGY-Platform/event-deployment/lib/auth"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/SENERGY-Platform/event-deployment/lib/events/conditionalevents"
	"github.com/SENERGY-Platform/event-deployment/lib/interfaces"
	"github.com/SENERGY-Platform/event-deployment/lib/metrics"
	"github.com/SENERGY-Platform/models/go/models"
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
	Deploy(owner string, deployment models.Deployment) error
	UpdateDeviceGroup(groupId string) error
}

func (this *EventsFactory) New(ctx context.Context, config config.Config, devices interfaces.Devices, imports interfaces.Imports, doneProducer interfaces.Producer, m *metrics.Metrics) (result interfaces.Events, err error) {
	handlers := []Handler{}
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

type DeploymentCommand struct {
	Command    string             `json:"command"`
	Id         string             `json:"id"`
	Owner      string             `json:"owner"`
	Deployment *models.Deployment `json:"deployment"`
	Source     string             `json:"source,omitempty"`
	Version    int64              `json:"version"`
}

func (this *Events) HandleCommand(msg []byte) error {
	this.config.GetLogger().Debug("received deployment command", "msg", string(msg))

	version := VersionWrapper{}
	err := json.Unmarshal(msg, &version)
	if err != nil {
		this.config.GetLogger().Error("invalid message --> ignore", "error", err)
		debug.PrintStack()
		return nil
	}
	if version.Version != models.CurrentDeploymentModelVersion {
		this.config.GetLogger().Error("consumed unexpected deployment version", "version", version.Version)
		if version.Command == "DELETE" {
			this.config.GetLogger().Warn("handle legacy delete")
			return this.Remove(version.Owner, version.Id)
		}
		return nil
	}

	cmd := DeploymentCommand{}
	err = json.Unmarshal(msg, &cmd)
	if err != nil {
		this.config.GetLogger().Error("invalid message --> ignore", "error", err)
		debug.PrintStack()
		return nil
	}
	switch cmd.Command {
	case "RIGHTS":
		return nil
	case "PUT":
		if cmd.Version != models.CurrentDeploymentModelVersion {
			this.config.GetLogger().Error("consumed unexpected deployment version", "version", cmd.Version)
			return nil
		}
		if cmd.Owner == "" {
			this.config.GetLogger().Error("missing owner --> ignore deployment command", "error", string(msg))
			return nil
		}
		if cmd.Deployment != nil {
			err = this.Deploy(cmd.Owner, *cmd.Deployment)
		}
		if errors.Is(err, auth.ErrUserDoesNotExist) {
			this.config.GetLogger().Warn("user does not exist -> ignore deployment command", "error", err, "owner", cmd.Owner)
			return nil
		}
		return err
	case "DELETE":
		if cmd.Owner == "" {
			this.config.GetLogger().Error("missing owner --> ignore deployment delete command", "error", string(msg))
			return nil
		}
		err = this.Remove(cmd.Owner, cmd.Id)
		if errors.Is(err, auth.ErrUserDoesNotExist) {
			this.config.GetLogger().Warn("user does not exist -> ignore deployment delete command", "error", err, "owner", cmd.Owner)
			return nil
		}
		return err
	default:
		return errors.New("unknown command " + cmd.Command)
	}
	return nil
}

func (this *Events) Deploy(owner string, deployment models.Deployment) (err error) {
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
		this.config.GetLogger().Debug("send deployment done", "message", message)
		msg, err := json.Marshal(message)
		if err != nil {
			this.config.GetLogger().Error("unable to marshal deployment done message", "error", err)
			debug.PrintStack()
			return
		}
		err = this.doneProducer.Produce(id, msg)
		if err != nil {
			this.config.GetLogger().Error("unable to send deployment done message", "error", err)
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
