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
	eventmodel "github.com/SENERGY-Platform/event-deployment/lib/model"
	"github.com/SENERGY-Platform/event-worker/pkg/model"
	"github.com/SENERGY-Platform/models/go/models"
	"log"
	"net/http"
	"runtime/debug"
)

func (this *Transformer) transformEventForDeviceGroup(owner string, deployentId string, event *models.ConditionalEvent) (result []model.EventDesc, err error) {
	desc := model.EventDesc{
		UserId:        owner,
		DeploymentId:  deployentId,
		DeviceGroupId: *event.Selection.SelectedDeviceGroupId,
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

	devices, _, err, code := this.devices.GetDeviceInfosOfGroup(desc.DeviceGroupId)
	if err != nil {
		if code == http.StatusInternalServerError {
			return []model.EventDesc{}, err
		} else {
			log.Println("ERROR:", code, err)
			debug.PrintStack()
			return []model.EventDesc{}, nil //ignore bad request errors
		}
	}

	deviceCache := map[string]models.Device{}

	for _, device := range devices {
		deviceCache[device.Id] = device
	}

	dtSelectables, err, code := this.devices.GetDeviceTypeSelectables([]eventmodel.FilterCriteria{{
		FunctionId: desc.FunctionId,
		AspectId:   desc.AspectId,
	}})
	if err != nil {
		if code == http.StatusInternalServerError {
			return []model.EventDesc{}, err
		} else {
			log.Println("ERROR:", code, err)
			debug.PrintStack()
			return []model.EventDesc{}, nil //ignore bad request errors
		}
	}

	result = []model.EventDesc{}
	for _, device := range devices {
		partial := desc
		partial.DeviceId = device.Id
		temp, err := this.transformPartialDescription(partial, dtSelectables, &deviceCache)
		if err != nil {
			return []model.EventDesc{}, err
		}
		result = append(result, temp...)
	}
	return result, nil
}

func (this *Transformer) transformEventForDeviceWithoutService(owner string, deployentId string, event *models.ConditionalEvent) (result []model.EventDesc, err error) {
	desc := model.EventDesc{
		UserId:        owner,
		DeploymentId:  deployentId,
		DeviceId:      *event.Selection.SelectedDeviceId,
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

	devices, _, err, code := this.devices.GetDeviceInfosOfDevices([]string{desc.DeviceId})
	if err != nil {
		if code == http.StatusInternalServerError {
			return []model.EventDesc{}, err
		} else {
			log.Println("ERROR:", code, err)
			debug.PrintStack()
			return []model.EventDesc{}, nil //ignore bad request errors
		}
	}

	deviceCache := map[string]models.Device{}

	for _, device := range devices {
		deviceCache[device.Id] = device
	}

	dtSelectables, err, code := this.devices.GetDeviceTypeSelectables([]eventmodel.FilterCriteria{{
		FunctionId: desc.FunctionId,
		AspectId:   desc.AspectId,
	}})
	if err != nil {
		if code == http.StatusInternalServerError {
			return []model.EventDesc{}, err
		} else {
			log.Println("ERROR:", code, err)
			debug.PrintStack()
			return []model.EventDesc{}, nil //ignore bad request errors
		}
	}

	result = []model.EventDesc{}
	for _, device := range devices {
		partial := desc
		partial.DeviceId = device.Id
		temp, err := this.transformPartialDescription(partial, dtSelectables, &deviceCache)
		if err != nil {
			return []model.EventDesc{}, err
		}
		result = append(result, temp...)
	}
	return result, nil
}

func (this *Transformer) transformPartialDescription(partialDesc model.EventDesc, selectables []eventmodel.DeviceTypeSelectable, deviceCache *map[string]models.Device) (result []model.EventDesc, err error) {
	dtId := (*deviceCache)[partialDesc.DeviceId].DeviceTypeId
	if dtId == "" {
		devices, _, err, code := this.devices.GetDeviceInfosOfDevices([]string{partialDesc.DeviceId})
		if err != nil {
			if code == http.StatusInternalServerError {
				return []model.EventDesc{}, err
			} else {
				log.Println("ERROR:", code, err)
				debug.PrintStack()
				return []model.EventDesc{}, nil //ignore bad request errors
			}
		}
		if len(devices) == 0 {
			log.Println("ERROR: unexpected GetDeviceInfosOfDevices() result", devices)
			debug.PrintStack()
			return []model.EventDesc{}, nil
		}
		device := devices[0]
		(*deviceCache)[device.Id] = device
		dtId = device.DeviceTypeId
	}

	result = []model.EventDesc{}
	for _, selectable := range selectables {
		if selectable.DeviceTypeId == dtId {
			for _, service := range selectable.Services {
				temp := partialDesc
				temp.ServiceId = service.Id
				temp.ServiceForMarshaller = service
				result = append(result, temp)
			}
		}
	}
	return result, nil
}
