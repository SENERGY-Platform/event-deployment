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
	"bytes"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/event-deployment/lib/model"
	"io/ioutil"
	"log"
	"net/http"
	"runtime/debug"
)

type PathOptionsQuery struct {
	DeviceTypeIds          []string `json:"device_type_ids"`
	FunctionId             string   `json:"function_id"`
	AspectId               string   `json:"aspect_id"`
	CharacteristicIdFilter []string `json:"characteristic_id_filter"`
	WithoutEnvelope        bool     `json:"without_envelope"`
}

func (this *Marshaller) FindPathOptions(deviceTypeIds []string, functionId string, aspectId string, characteristicsIdFilter []string, withEnvelope bool) (result map[string][]model.PathOptionsResultElement, err error) {
	query := &bytes.Buffer{}
	err = json.NewEncoder(query).Encode(PathOptionsQuery{
		DeviceTypeIds:          deviceTypeIds,
		FunctionId:             functionId,
		AspectId:               aspectId,
		CharacteristicIdFilter: characteristicsIdFilter,
		WithoutEnvelope:        !withEnvelope,
	})
	if err != nil {
		debug.PrintStack()
		return result, err
	}

	req, err := http.NewRequest(
		"POST",
		this.config.MarshallerUrl+"/query/path-options",
		query,
	)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		temp, err := ioutil.ReadAll(resp.Body)
		log.Println("WARNING: invalid message event with device group", string(temp), err)
		debug.PrintStack()
		return map[string][]model.PathOptionsResultElement{}, nil
	}

	if resp.StatusCode >= 300 {
		debug.PrintStack()
		return result, errors.New("unexpected status code")
	}

	temp, err := ioutil.ReadAll(resp.Body)
	result = map[string][]model.PathOptionsResultElement{}
	err = json.Unmarshal(temp, &result)
	if err != nil {
		log.Println("ERROR:", err, string(temp))
		debug.PrintStack()
		return result, err
	}
	return result, err
}
