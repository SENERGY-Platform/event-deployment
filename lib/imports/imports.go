/*
 * Copyright 2021 InfAI (CC SES)
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

package imports

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/SENERGY-Platform/event-deployment/lib/interfaces"
	"github.com/SENERGY-Platform/process-deployment/lib/model/importmodel"
	"log"
	"net/http"
	"runtime/debug"
)

type FactoryType struct{}

func (this *FactoryType) New(config config.Config) interfaces.Imports {
	return New(config)
}

var Factory = &FactoryType{}

type Imports struct {
	config config.Config
}

func New(config config.Config) *Imports {
	return &Imports{
		config: config,
	}
}

func (this *Imports) GetTopic(user string, importId string) (topic string, err error, code int) {
	req, err := http.NewRequest("GET", this.config.ImportDeployUrl+"/instances/"+importId, nil)
	if err != nil {
		debug.PrintStack()
		return "", err, http.StatusInternalServerError
	}
	req.Header.Set("X-UserId", user)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		debug.PrintStack()
		return "", err, http.StatusInternalServerError
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(resp.Body)
		if err != nil {
			return "", err, resp.StatusCode
		}
		err = errors.New(buf.String())
		log.Println("ERROR: ", resp.StatusCode, err)
		debug.PrintStack()
		return "", err, resp.StatusCode
	}
	var importInstance importmodel.Import
	err = json.NewDecoder(resp.Body).Decode(&importInstance)
	if err != nil {
		return "", err, http.StatusInternalServerError
	}
	topic = importInstance.KafkaTopic
	return topic, nil, 200
}
