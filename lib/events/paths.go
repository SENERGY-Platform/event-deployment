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

package events

import (
	"github.com/SENERGY-Platform/event-deployment/lib/model"
	"log"
)

func (this *Events) getDeviceGroupPathOptions(desc model.GroupEventDescription, deviceTypeIds []string) (result map[string][]model.PathOptionsResultElement, err error) {
	result = map[string][]model.PathOptionsResultElement{}
	selectables, err, _ := this.devices.GetDeviceTypeSelectables([]model.FilterCriteria{{
		FunctionId: desc.FunctionId,
		AspectId:   desc.AspectId,
	}})
	if err != nil {
		return result, err
	}
	for _, dtId := range deviceTypeIds {
		for _, selectable := range selectables {
			if selectable.DeviceTypeId == dtId {
				for sid, options := range selectable.ServicePathOptions {
					temp := model.PathOptionsResultElement{
						ServiceId:              sid,
						JsonPath:               []string{},
						PathToCharacteristicId: map[string]string{},
					}
					for _, option := range options {
						if option.ServiceId == sid {
							temp.JsonPath = append(temp.JsonPath, option.Path)
							temp.PathToCharacteristicId[option.Path] = option.CharacteristicId
						} else {
							log.Println("WARNING: unexpected service id in ServicePathOptions")
						}
					}
					result[dtId] = append(result[dtId], temp)
				}
			}
		}
	}
	return result, nil
}
