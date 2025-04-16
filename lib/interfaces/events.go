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

package interfaces

import (
	"context"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/SENERGY-Platform/event-deployment/lib/metrics"
	"github.com/SENERGY-Platform/event-deployment/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
)

type EventsFactory interface {
	New(ctx context.Context, config config.Config, analytics Analytics, devices Devices, imports Imports, doneProducer Producer, m *metrics.Metrics) (Events, error)
}

type Events interface {
	HandleCommand(msg []byte) error
	Remove(owner string, deploymentId string) (err error)
	Deploy(owner string, deployment models.Deployment) (err error)
	HandleDeviceGroupUpdate(msg []byte) error
	UpdateDeviceGroup(group model.DeviceGroup) (err error)
	CheckEvent(token string, id string) int
	GetEventStates(token string, ids []string) (states map[string]bool, err error, code int)
}
