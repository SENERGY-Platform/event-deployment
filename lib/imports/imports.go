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
	"github.com/SENERGY-Platform/event-deployment/lib/auth"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/SENERGY-Platform/event-deployment/lib/interfaces"
	"github.com/SENERGY-Platform/process-deployment/lib/model/importmodel"
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
	auth   *auth.Auth
}

func New(config config.Config) *Imports {
	return &Imports{
		config: config,
		auth:   auth.NewAuth(config),
	}
}

func (this *Imports) GetTopic(user string, importId string) (topic string, err error, code int) {
	token, err := this.auth.GetUserToken(user)
	if err != nil {
		debug.PrintStack()
		return "", err, http.StatusInternalServerError
	}
	var importInstance importmodel.Import
	err = token.GetJSON(this.config.ImportDeployUrl+"/instances/"+importId, &importInstance)
	if err != nil {
		debug.PrintStack()
		return "", err, http.StatusBadGateway
	}

	topic = importInstance.KafkaTopic
	return topic, nil, 200
}
