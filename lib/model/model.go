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

type DeviceGroup struct {
	Id        string   `json:"id"`
	Name      string   `json:"name"`
	DeviceIds []string `json:"device_ids"`
}

type Device struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	DeviceTypeId string `json:"device_type_id"`
}

type PathOptionsResultElement struct {
	ServiceId              string            `json:"service_id"`
	JsonPath               []string          `json:"json_path"`
	PathToCharacteristicId map[string]string `json:"path_to_characteristic_id"`
}

type GroupEventDescription struct {
	ImportId         string
	Path             string
	DeviceGroupId    string
	DeviceIds        []string //optional
	EventId          string
	DeploymentId     string
	FunctionId       string
	AspectId         string
	FlowId           string
	OperatorValue    string
	CharacteristicId string
}

type PathAndCharacteristic struct {
	JsonPath         string `json:"json_path"`
	CharacteristicId string `json:"characteristic_id"`
}
