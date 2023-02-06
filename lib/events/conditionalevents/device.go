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

package conditionalevents

import (
	"github.com/SENERGY-Platform/event-deployment/lib/auth"
	"github.com/SENERGY-Platform/event-deployment/lib/events/conditionalevents/idmodifier"
	"github.com/SENERGY-Platform/event-worker/pkg/model"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel"
	"log"
	"net/http"
	"runtime/debug"
)

func (this *Events) deployEventForDevice(token auth.AuthToken, owner string, deployentId string, event *deploymentmodel.ConditionalEvent) error {
	desc := model.EventDesc{
		UserId:        owner,
		DeploymentId:  deployentId,
		DeviceId:      *event.Selection.SelectedDeviceId,
		ServiceId:     *event.Selection.SelectedServiceId,
		Script:        event.Script,
		ValueVariable: event.ValueVariable,
		Variables:     event.Variables,
		Qos:           event.Qos,
		EventId:       event.EventId,
	}

	if event.Selection.FilterCriteria.CharacteristicId != nil {
		desc.CharacteristicId = *event.Selection.FilterCriteria.CharacteristicId
	}
	if event.Selection.FilterCriteria.FunctionId != nil {
		desc.FunctionId = *event.Selection.FilterCriteria.FunctionId
	}
	if event.Selection.FilterCriteria.AspectId != nil {
		desc.AspectId = *event.Selection.FilterCriteria.AspectId
	}
	if event.Selection.SelectedPath != nil {
		desc.Path = event.Selection.SelectedPath.Path
	}

	desc.DeviceId, _ = idmodifier.SplitModifier(desc.DeviceId)
	desc.ServiceId, _ = idmodifier.SplitModifier(desc.ServiceId)

	service, err, code := this.devices.GetService(desc.ServiceId)
	if err != nil {
		if code == http.StatusInternalServerError {
			return err
		} else {
			log.Println("ERROR:", code, err)
			debug.PrintStack()
			return nil //ignore bad request errors
		}
	}
	desc.ServiceForMarshaller = service

	return this.deployDescription(desc)
}
