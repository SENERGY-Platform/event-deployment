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
	GetDeviceInfosOfGroupValues map[string][]model.DevicePerm
}

func (this *DevicesMock) GetDeviceInfosOfGroup(groupId string) (devices []model.DevicePerm, deviceTypeIds []string, err error, code int) {
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
