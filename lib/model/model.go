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

package model

import (
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel"
)

type DeviceGroup = models.DeviceGroup

type Device = models.Device

type Deployment struct {
	deploymentmodel.Deployment
	UserId string
}

type PathOptionsResultElement struct {
	ServiceId              string            `json:"service_id"`
	JsonPath               []string          `json:"json_path"`
	PathToCharacteristicId map[string]string `json:"path_to_characteristic_id"`
}

type GroupEventDescription struct {
	GenericEventSource *deploymentmodel.GenericEventSource
	ImportId           string
	Path               string
	DeviceGroupId      string
	DeviceIds          []string //optional
	EventId            string
	DeploymentId       string
	FunctionId         string
	AspectId           string
	FlowId             string
	OperatorValue      string
	CharacteristicId   string
	UseMarshaller      bool
}

type PathAndCharacteristic struct {
	JsonPath         string `json:"json_path"`
	CharacteristicId string `json:"characteristic_id"`
}

type FilterCriteria = struct {
	FunctionId    string `json:"function_id"`
	DeviceClassId string `json:"device_class_id"`
	AspectId      string `json:"aspect_id"`
}

type DeviceTypeSelectable struct {
	DeviceTypeId       string                         `json:"device_type_id,omitempty"`
	Services           []models.Service               `json:"services,omitempty"`
	ServicePathOptions map[string][]ServicePathOption `json:"service_path_options,omitempty"`
}

type ServicePathOption struct {
	ServiceId        string `json:"service_id"`
	Path             string `json:"path"`
	CharacteristicId string `json:"characteristic_id"`
	//AspectNode            AspectNode     `json:"aspect_node"`
	FunctionId            string      `json:"function_id"`
	IsVoid                bool        `json:"is_void"`
	Value                 interface{} `json:"value,omitempty"`
	IsControllingFunction bool        `json:"is_controlling_function"`
	//Configurables         []Configurable `json:"configurables,omitempty"`
	//Type                  Type           `json:"type,omitempty"`
}

type Function = models.Function

type Concept = models.Concept

type ConverterExtension = models.ConverterExtension
