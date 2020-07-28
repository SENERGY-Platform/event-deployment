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

package marshaller

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/SENERGY-Platform/event-deployment/lib/interfaces"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"runtime/debug"
	"time"
)

type FactoryType struct{}

var Factory = &FactoryType{}

type Marshaller struct {
	config config.Config
}

func (this *FactoryType) New(ctx context.Context, config config.Config) (interfaces.Marshaller, error) {
	return &Marshaller{config: config}, nil
}

type Response struct {
	Path                    string `json:"path"`
	ServiceCharacteristicId string `json:"service_characteristic_id"`
}

var ErrServiceNotFound = errors.New("service not found")
var ErrCharacteristicNotFoundInService = errors.New("characteristic not in service")

func (this *Marshaller) FindPath(serviceId string, characteristicId string) (path string, serviceCharacteristicId string, err error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest(
		"GET",
		this.config.MarshallerUrl+"/characteristic-paths/"+url.PathEscape(serviceId)+"/"+url.PathEscape(characteristicId),
		nil,
	)
	if err != nil {
		debug.PrintStack()
		return path, serviceCharacteristicId, err
	}
	resp, err := client.Do(req)
	if err != nil {
		debug.PrintStack()
		return path, serviceCharacteristicId, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		temp, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("ERROR:", err, string(temp))
			debug.PrintStack()
			return path, serviceCharacteristicId, err
		}
		if string(temp) == ErrServiceNotFound.Error() {
			return "", "", ErrServiceNotFound
		}
		if string(temp) == ErrCharacteristicNotFoundInService.Error() {
			return "", "", ErrCharacteristicNotFoundInService
		}
		return "", "", errors.New(string(temp))
	}

	if resp.StatusCode >= 300 {
		debug.PrintStack()
		return path, serviceCharacteristicId, errors.New("unexpected status code")
	}

	temp, err := ioutil.ReadAll(resp.Body)
	result := Response{}
	err = json.Unmarshal(temp, &result)
	if err != nil {
		log.Println("ERROR:", err, string(temp))
		debug.PrintStack()
		return path, serviceCharacteristicId, err
	}
	return result.Path, result.ServiceCharacteristicId, err
}
