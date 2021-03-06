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

package mocks

import (
	"errors"
	"github.com/SENERGY-Platform/event-deployment/lib/model"
)

type DevicesMock struct {
	GetDeviceInfosOfGroupValues map[string][]model.Device //key = groupId
}

func (this *DevicesMock) GetDeviceInfosOfDevices(deviceIds []string) (devices []model.Device, deviceTypeIds []string, err error, code int) {
	allDevices := map[string]model.Device{}
	for _, group := range this.GetDeviceInfosOfGroupValues {
		for _, device := range group {
			allDevices[device.Id] = device
		}
	}
	done := map[string]bool{}
	for _, deviceId := range deviceIds {
		device := allDevices[deviceId]
		devices = append(devices, device)
		if !done[device.DeviceTypeId] {
			done[device.DeviceTypeId] = true
			deviceTypeIds = append(deviceTypeIds, device.DeviceTypeId)
		}
	}
	return devices, deviceTypeIds, nil, 200
}

func (this *DevicesMock) GetDeviceInfosOfGroup(groupId string) (devices []model.Device, deviceTypeIds []string, err error, code int) {
	if this.GetDeviceInfosOfGroupValues == nil {
		return nil, nil, errors.New("DevicesMock.GetDeviceInfosOfGroupValues not set"), 500
	}
	if devices, ok := this.GetDeviceInfosOfGroupValues[groupId]; !ok {
		return nil, nil, errors.New("DevicesMock.GetDeviceInfosOfGroupValues[" + groupId + "] not set"), 500
	} else {
		done := map[string]bool{}
		for _, d := range devices {
			if !done[d.DeviceTypeId] {
				done[d.DeviceTypeId] = true
				deviceTypeIds = append(deviceTypeIds, d.DeviceTypeId)
			}
		}
		return devices, deviceTypeIds, nil, 200
	}
}
