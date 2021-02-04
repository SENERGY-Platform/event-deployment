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
	"github.com/SENERGY-Platform/event-deployment/lib/marshaller"
	"github.com/SENERGY-Platform/event-deployment/lib/model"
)

type MarshallerMock struct {
	FindPathValues map[string]map[string]marshaller.Response
}

func (this *MarshallerMock) FindPathOptions(deviceTypeIds []string, functionId string, aspectId string, characteristicsIdFilter []string, withEnvelope bool) (result map[string][]model.PathOptionsResultElement, err error) {
	return result, errors.New("not implemented")
}

func (this *MarshallerMock) FindPath(serviceId string, characteristicId string) (path string, serviceCharacteristicId string, err error) {
	if this == nil {
		return "", "", errors.New("missing mock data")
	}
	temp, ok := this.FindPathValues[serviceId]
	if !ok {
		return "", "", marshaller.ErrServiceNotFound
	}
	resp, ok := temp[characteristicId]
	if !ok {
		return "", "", marshaller.ErrCharacteristicNotFoundInService
	}
	return resp.Path, resp.ServiceCharacteristicId, nil
}
