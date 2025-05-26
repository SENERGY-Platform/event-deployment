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
	"bytes"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/event-deployment/lib/auth"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/SENERGY-Platform/event-deployment/lib/interfaces"
	"github.com/SENERGY-Platform/event-deployment/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"log"
	"net/http"
	"net/url"
	"runtime/debug"
)

type FactoryType struct{}

func (this *FactoryType) New(config config.Config) (interfaces.Devices, error) {
	return New(config)
}

var Factory = &FactoryType{}

type Devices struct {
	config     config.Config
	auth       Auth
	devicerepo client.Interface
}

type Auth interface {
	Ensure() (token auth.AuthToken, err error)
}

func New(config config.Config) (result *Devices, err error) {
	var a Auth = InternalAdminTokenAuth{}
	if config.AuthEndpoint != "" && config.AuthEndpoint != "-" {
		a, err = auth.NewAuth(config)
		if err != nil {
			return nil, err
		}
	}
	return NewWithAuth(config, a), nil
}

func NewWithAuth(config config.Config, auth Auth) *Devices {
	return &Devices{
		config: config,
		auth:   auth,
		devicerepo: client.NewClient(config.DeviceRepositoryUrl, func() (token string, err error) {
			temp, err := auth.Ensure()
			return string(temp), err
		}),
	}
}

type InternalAdminTokenAuth struct{}

func (this InternalAdminTokenAuth) Ensure() (token auth.AuthToken, err error) {
	return client.InternalAdminToken, nil
}

func (this *Devices) GetDeviceInfosOfGroup(groupId string) (devices []model.Device, deviceTypeIds []string, err error, code int) {
	token, err := this.auth.Ensure()
	if err != nil {
		return devices, nil, err, http.StatusInternalServerError
	}
	group, err, code := this.GetDeviceGroup(token, groupId)
	if err != nil {
		if code == http.StatusNotFound || code == http.StatusForbidden || code == http.StatusUnauthorized {
			return nil, nil, nil, http.StatusOK
		}
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

func (this *Devices) GetDeviceGroup(token auth.AuthToken, groupId string) (result model.DeviceGroup, err error, code int) {
	return this.devicerepo.ReadDeviceGroup(groupId, string(token), false)
}

func (this *Devices) GetDevicesWithIds(token auth.AuthToken, ids []string) (result []model.Device, err error, code int) {
	return this.devicerepo.ListDevices(string(token), client.DeviceListOptions{Ids: ids})
}

func (this *Devices) GetService(serviceId string) (result models.Service, err error, code int) {
	token, err := this.auth.Ensure()
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}
	req, err := http.NewRequest("GET", this.config.DeviceRepositoryUrl+"/services/"+url.PathEscape(serviceId), nil)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}
	token.UseInRequest(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		err = errors.New(buf.String())
		log.Println("ERROR: ", resp.StatusCode, err)
		debug.PrintStack()
		return result, err, resp.StatusCode
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}

	return result, nil, http.StatusOK
}

func (this *Devices) GetDeviceTypeSelectables(criteria []model.FilterCriteria) (result []model.DeviceTypeSelectable, err error, code int) {
	token, err := this.auth.Ensure()
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}
	requestBody := new(bytes.Buffer)
	err = json.NewEncoder(requestBody).Encode(criteria)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req, err := http.NewRequest("POST", this.config.DeviceRepositoryUrl+"/query/device-type-selectables?interactions-filter=event&include_id_modified=true", requestBody)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}
	token.UseInRequest(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		err = errors.New(buf.String())
		log.Println("ERROR: ", resp.StatusCode, err)
		debug.PrintStack()
		return result, err, resp.StatusCode
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}

	return result, nil, http.StatusOK
}
