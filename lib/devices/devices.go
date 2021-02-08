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

package devices

import (
	"errors"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/SENERGY-Platform/event-deployment/lib/interfaces"
	"github.com/SENERGY-Platform/event-deployment/lib/model"
	"net/http"
)

type FactoryType struct{}

func (this *FactoryType) New(config config.Config) interfaces.Devices {
	return New(config)
}

var Factory = &FactoryType{}

type Devices struct {
	config config.Config
	auth   *Auth
}

func New(config config.Config) *Devices {
	return &Devices{
		config: config,
		auth:   NewAuth(config),
	}
}

func (this *Devices) GetDeviceInfosOfGroup(groupId string) (devices []model.Device, deviceTypeIds []string, err error, code int) {
	token, err := this.auth.Ensure()
	if err != nil {
		return devices, nil, err, http.StatusInternalServerError
	}
	group, err, code := this.GetDeviceGroup(token, groupId)
	if err != nil {
		return devices, nil, err, code
	}
	return this.GetDeviceInfosOfDevices(group.DeviceIds)
}

func (this *Devices) GetDeviceInfosOfDevices(deviceIds []string) (devices []model.Device, deviceTypeIds []string, err error, code int) {
	token, err := this.auth.Ensure()
	if err != nil {
		return devices, nil, err, http.StatusInternalServerError
	}
	devices, err, code = this.GetDevicesWithIds(token, deviceIds)
	if err != nil {
		return devices, nil, err, code
	}
	deviceTypeIsUsed := map[string]bool{}
	for _, d := range devices {
		if !deviceTypeIsUsed[d.DeviceTypeId] {
			deviceTypeIsUsed[d.DeviceTypeId] = true
			deviceTypeIds = append(deviceTypeIds, d.DeviceTypeId)
		}
	}
	return devices, deviceTypeIds, nil, http.StatusOK
}

func (this *Devices) GetDeviceGroup(token AuthToken, groupId string) (result model.DeviceGroup, err error, code int) {
	groups := []model.DeviceGroup{}
	err, code = this.Search(token, QueryMessage{
		Resource: "device-groups",
		ListIds: &QueryListIds{
			QueryListCommons: QueryListCommons{
				Limit:    1,
				Offset:   0,
				Rights:   "r",
				SortBy:   "name",
				SortDesc: false,
			},
			Ids: []string{groupId},
		},
	}, &groups)
	if err != nil {
		return result, err, code
	}
	if len(groups) == 0 {
		return result, errors.New("not found"), http.StatusNotFound
	}
	return groups[0], nil, http.StatusOK
}

func (this *Devices) GetDevicesWithIds(token AuthToken, ids []string) (result []model.Device, err error, code int) {
	err, code = this.Search(token, QueryMessage{
		Resource: "devices",
		ListIds: &QueryListIds{
			QueryListCommons: QueryListCommons{
				Limit:    len(ids),
				Offset:   0,
				Rights:   "r",
				SortBy:   "name",
				SortDesc: false,
			},
			Ids: ids,
		},
	}, &result)
	return
}
