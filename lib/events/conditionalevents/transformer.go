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
	"github.com/SENERGY-Platform/event-deployment/lib/interfaces"
	"github.com/SENERGY-Platform/event-worker/pkg/model"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel"
	"log"
)

func NewTransformer(devices Devices, imports interfaces.Imports) *Transformer {
	return &Transformer{
		devices: devices,
		imports: imports,
	}
}

type Transformer struct {
	devices Devices
	imports interfaces.Imports
}

func (this *Transformer) Transform(token auth.AuthToken, owner string, deployment deploymentmodel.Deployment) (result []model.EventDesc, err error) {
	for _, element := range deployment.Elements {
		temp, err := this.TransformElement(token, owner, deployment.Id, element)
		if err != nil {
			return result, err
		}
		result = append(result, temp...)
	}
	return result, nil
}

func (this *Transformer) TransformElement(token auth.AuthToken, owner string, deploymentId string, element deploymentmodel.Element) (result []model.EventDesc, err error) {
	event := element.ConditionalEvent
	if event != nil && event.Selection.FilterCriteria.CharacteristicId != nil {
		if event.Selection.SelectedDeviceGroupId != nil && *event.Selection.SelectedDeviceGroupId != "" {
			return this.transformEventForDeviceGroup(token, owner, deploymentId, event)
		}
		if event.Selection.SelectedDeviceId != nil && event.Selection.SelectedServiceId != nil && *event.Selection.SelectedServiceId != "" {
			return this.transformEventForDevice(token, owner, deploymentId, event)
		}
		if event.Selection.SelectedDeviceId != nil && !(event.Selection.SelectedServiceId != nil && *event.Selection.SelectedServiceId != "") {
			return this.transformEventForDeviceWithoutService(token, owner, deploymentId, event)
		}
		if event.Selection.SelectedImportId != nil {
			return this.transformEventForImport(token, owner, deploymentId, event)
		}
		if event.Selection.SelectedGenericEventSource != nil {
			log.Println("WARNING: generic event sources not supported for conditional events")
			return []model.EventDesc{}, nil
		}
	}
	return []model.EventDesc{}, nil
}
