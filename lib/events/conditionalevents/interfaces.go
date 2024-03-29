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
	"github.com/SENERGY-Platform/event-deployment/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
)

type Devices interface {
	GetDeviceInfosOfGroup(groupId string) (devices []model.Device, deviceTypeIds []string, err error, code int)
	GetDeviceInfosOfDevices(deviceIds []string) (devices []model.Device, deviceTypeIds []string, err error, code int)
	GetDeviceTypeSelectables(criteria []model.FilterCriteria) (result []model.DeviceTypeSelectable, err error, code int)
	GetService(serviceId string) (result models.Service, err error, code int)
}
