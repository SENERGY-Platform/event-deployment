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
	"github.com/SENERGY-Platform/event-deployment/lib/auth"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/SENERGY-Platform/event-deployment/lib/model"
)

type AnalyticsFactory interface {
	New(ctx context.Context, config config.Config) (Analytics, error)
}

type Analytics interface {
	UpdateGroupDeployment(token auth.AuthToken, pipelineId string, label string, owner string, desc model.GroupEventDescription, serviceIds []string, serviceToDeviceIdsMapping map[string][]string, serviceToPathsMapping map[string][]string, serviceToPathAndCharacteristic map[string][]model.PathAndCharacteristic) (err error)
	DeployGroup(token auth.AuthToken, label string, user string, desc model.GroupEventDescription, serviceIds []string, serviceToDeviceIdsMapping map[string][]string, serviceToPathsMapping map[string][]string, serviceToPathAndCharacteristic map[string][]model.PathAndCharacteristic) (pipelineId string, err error)
	DeployImport(token auth.AuthToken, label string, user string, desc model.GroupEventDescription, topic string, path string, castFrom string, castTo string, castExtensions []model.ConverterExtension) (pipelineId string, err error)
	Deploy(token auth.AuthToken, label string, user string, deploymentId string, flowId string, eventId string, deviceId string, serviceId string, value string, path string, castFrom string, castTo string, castExtensions []model.ConverterExtension) (pipelineId string, err error)
	Remove(user string, pipelineId string) error
	GetPipelinesByDeploymentId(owner string, deploymentId string) (pipelineIds []string, err error)
	GetPipelineByEventId(owner string, eventId string) (pipelineId string, exists bool, err error)
	GetPipelinesByDeviceGroupId(owner string, groupId string) (pipelineIds []string, pipelineToGroupDescription map[string]model.GroupEventDescription, pipelineNames map[string]string, err error)
	GetEventStates(userId string, eventIds []string) (states map[string]bool, err error)
}
