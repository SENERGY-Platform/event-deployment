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
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/event-deployment/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"net/http"
)

type DevicesMock struct {
	GetDeviceInfosOfGroupValues    map[string][]model.Device //key = groupId
	GetDeviceTypeSelectablesValues map[string]map[string][]model.DeviceTypeSelectable
	Functions                      map[string]model.Function
	Concepts                       map[string]model.Concept
}

func (this *DevicesMock) GetService(serviceId string) (result models.Service, err error, code int) {
	str := `{
                "attributes": [],
                "description": "",
                "id": "urn:infai:ses:service:137fdd3e-bf55-4129-b71a-f33df8223097",
                "inputs": [],
                "interaction": "event+request",
                "local_id": "getStatus",
                "name": "getStatusService",
                "outputs": [
                    {
                        "content_variable": {
                            "aspect_id": "urn:infai:ses:aspect:a7470d73-dde3-41fc-92bd-f16bb28f2da6",
                            "characteristic_id": "urn:infai:ses:characteristic:64928e9f-98ca-42bb-a1e5-adf2a760a2f9",
                            "function_id": "urn:infai:ses:measuring-function:bdb6a7c8-4a3d-4fe0-bab3-ce02e09b5869",
                            "id": "urn:infai:ses:content-variable:a3579952-2068-4353-9243-4d0d7f5eabca",
                            "is_void": false,
                            "name": "struct",
                            "serialization_options": null,
                            "sub_content_variables": [
                                {
                                    "characteristic_id": "urn:infai:ses:characteristic:d840607c-c8f9-45d6-b9bd-2c2d444e2899",
                                    "id": "urn:infai:ses:content-variable:fb9d3125-91a7-42ea-866c-def28cef7123",
                                    "is_void": false,
                                    "name": "brightness",
                                    "serialization_options": null,
                                    "sub_content_variables": null,
                                    "type": "https://schema.org/Integer",
                                    "value": null
                                },
                                {
                                    "characteristic_id": "urn:infai:ses:characteristic:6ec70e99-8c6a-4909-8d5a-7cc12af76b9a",
                                    "id": "urn:infai:ses:content-variable:6e25d312-1140-4210-9a12-e18ed72804fd",
                                    "is_void": false,
                                    "name": "hue",
                                    "serialization_options": null,
                                    "sub_content_variables": null,
                                    "type": "https://schema.org/Integer",
                                    "value": null
                                },
                                {
                                    "aspect_id": "urn:infai:ses:aspect:a7470d73-dde3-41fc-92bd-f16bb28f2da6",
                                    "characteristic_id": "",
                                    "id": "urn:infai:ses:content-variable:6534ae08-e920-4f94-8be7-043260581584",
                                    "is_void": false,
                                    "name": "kelvin",
                                    "serialization_options": null,
                                    "sub_content_variables": null,
                                    "type": "https://schema.org/Float",
                                    "value": null
                                },
                                {
                                    "aspect_id": "urn:infai:ses:aspect:a7470d73-dde3-41fc-92bd-f16bb28f2da6",
                                    "characteristic_id": "urn:infai:ses:characteristic:7621686a-56bc-402d-b4cc-5b266d39736f",
                                    "function_id": "urn:infai:ses:measuring-function:20d3c1d3-77d7-4181-a9f3-b487add58cd0",
                                    "id": "urn:infai:ses:content-variable:16007de9-0700-4cf3-95fb-7a54ae2b6924",
                                    "is_void": false,
                                    "name": "power",
                                    "serialization_options": null,
                                    "sub_content_variables": null,
                                    "type": "https://schema.org/Text",
                                    "value": null
                                },
                                {
                                    "characteristic_id": "urn:infai:ses:characteristic:a66dc568-c0e0-420f-b513-18e8df405538",
                                    "id": "urn:infai:ses:content-variable:2d010c30-279a-45b8-a973-708283a6dec3",
                                    "is_void": false,
                                    "name": "saturation",
                                    "serialization_options": null,
                                    "sub_content_variables": null,
                                    "type": "https://schema.org/Integer",
                                    "value": null
                                },
                                {
                                    "characteristic_id": "",
                                    "id": "urn:infai:ses:content-variable:aa01de36-1240-4dab-8634-cf19859438d5",
                                    "is_void": false,
                                    "name": "status",
                                    "serialization_options": null,
                                    "sub_content_variables": null,
                                    "type": "https://schema.org/Integer",
                                    "value": null
                                },
                                {
                                    "aspect_id": "urn:infai:ses:aspect:a7470d73-dde3-41fc-92bd-f16bb28f2da6",
                                    "characteristic_id": "urn:infai:ses:characteristic:6bc41b45-a9f3-4d87-9c51-dd3e11257800",
                                    "function_id": "urn:infai:ses:measuring-function:3b4e0766-0d67-4658-b249-295902cd3290",
                                    "id": "urn:infai:ses:content-variable:6afe6df0-8fb5-41e2-99fe-e5b69d3f9e42",
                                    "is_void": false,
                                    "name": "time",
                                    "serialization_options": null,
                                    "sub_content_variables": null,
                                    "type": "https://schema.org/Text",
                                    "value": null
                                },
                                {
                                    "characteristic_id": "",
                                    "id": "urn:infai:ses:content-variable:df153ef4-f351-41ab-8748-5689496fd006",
                                    "is_void": false,
                                    "name": "brightness_unit",
                                    "serialization_options": null,
                                    "sub_content_variables": null,
                                    "type": "https://schema.org/Text",
                                    "unit_reference": "brightness",
                                    "value": null
                                },
                                {
                                    "characteristic_id": "",
                                    "id": "urn:infai:ses:content-variable:5d3ab5cf-fda1-4218-aad5-6fc0ae48f05f",
                                    "is_void": false,
                                    "name": "hue_unit",
                                    "serialization_options": null,
                                    "sub_content_variables": null,
                                    "type": "https://schema.org/Text",
                                    "unit_reference": "hue",
                                    "value": null
                                },
                                {
                                    "characteristic_id": "",
                                    "id": "urn:infai:ses:content-variable:87baee56-33c3-4985-888e-40c27f063329",
                                    "is_void": false,
                                    "name": "kelvin_unit",
                                    "serialization_options": null,
                                    "sub_content_variables": null,
                                    "type": "https://schema.org/Text",
                                    "unit_reference": "kelvin",
                                    "value": null
                                },
                                {
                                    "characteristic_id": "",
                                    "id": "urn:infai:ses:content-variable:656d4a95-b3bb-4df3-909d-910491c8cc11",
                                    "is_void": false,
                                    "name": "power_unit",
                                    "serialization_options": null,
                                    "sub_content_variables": null,
                                    "type": "https://schema.org/Text",
                                    "unit_reference": "power",
                                    "value": null
                                },
                                {
                                    "characteristic_id": "",
                                    "id": "urn:infai:ses:content-variable:6ac39c66-f726-4e31-8683-74af81ae0acf",
                                    "is_void": false,
                                    "name": "saturation_unit",
                                    "serialization_options": null,
                                    "sub_content_variables": null,
                                    "type": "https://schema.org/Text",
                                    "unit_reference": "saturation",
                                    "value": null
                                },
                                {
                                    "characteristic_id": "",
                                    "id": "urn:infai:ses:content-variable:ea1260e0-c166-45f7-8d3a-140d285c569d",
                                    "is_void": false,
                                    "name": "time_unit",
                                    "serialization_options": null,
                                    "sub_content_variables": null,
                                    "type": "https://schema.org/Text",
                                    "unit_reference": "time",
                                    "value": null
                                }
                            ],
                            "type": "https://schema.org/StructuredValue",
                            "value": null
                        },
                        "id": "urn:infai:ses:content:9fa5b0db-61fd-4502-b9af-98da93229260",
                        "protocol_segment_id": "urn:infai:ses:protocol-segment:0d211842-cef8-41ec-ab6b-9dbc31bc3a65",
                        "serialization": "json"
                    }
                ],
                "protocol_id": "urn:infai:ses:protocol:f3a63aeb-187e-4dd9-9ef5-d97a6eb6292b",
                "service_group_key": ""
            }`
	err = json.Unmarshal([]byte(str), &result)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	result.Id = serviceId
	return result, nil, http.StatusOK
}

func (this *DevicesMock) GetDeviceTypeSelectables(criteria []model.FilterCriteria) (result []model.DeviceTypeSelectable, err error, code int) {
	if len(criteria) != 1 {
		return nil, errors.New("expect exactly 1 criteria"), http.StatusInternalServerError
	}
	functionId := criteria[0].FunctionId
	aspectId := criteria[0].AspectId
	functionMap, ok := this.GetDeviceTypeSelectablesValues[functionId]
	if !ok {
		//no function found
		return result, nil, http.StatusOK
	}
	result, ok = functionMap[aspectId]
	if !ok {
		//no aspect found
		return result, nil, http.StatusOK
	}
	return result, nil, http.StatusOK
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

func (this *DevicesMock) GetConcept(conceptId string) (result model.Concept, err error, code int) {
	if result, ok := this.Concepts[conceptId]; ok {
		return result, nil, http.StatusOK
	} else {
		return result, errors.New("not found"), http.StatusNotFound
	}
}

func (this *DevicesMock) GetFunction(functionId string) (result model.Function, err error, code int) {
	if result, ok := this.Functions[functionId]; ok {
		return result, nil, http.StatusOK
	} else {
		return result, errors.New("not found"), http.StatusNotFound
	}
}
