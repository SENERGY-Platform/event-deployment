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
	"net/url"
	"runtime/debug"
)

type FactoryType struct{}

func (this *FactoryType) New(config config.Config) (interfaces.Imports, error) {
	return New(config)
}

var Factory = &FactoryType{}

type Imports struct {
	config config.Config
	auth   *auth.Auth
}

func New(config config.Config) (*Imports, error) {
	a, err := auth.NewAuth(config)
	if err != nil {
		return nil, err
	}
	return &Imports{
		config: config,
		auth:   a,
	}, nil
}

func (this *Imports) GetTopic(user string, importId string) (topic string, err error, code int) {
	instance, err, code := this.GetImportInstance(user, importId)
	if err != nil {
		debug.PrintStack()
		return "", err, code
	}
	topic = instance.KafkaTopic
	return topic, nil, 200
}

func (this *Imports) GetImportInstance(user string, importId string) (importInstance importmodel.Import, err error, code int) {
	token, err := this.auth.GetUserToken(user)
	if err != nil {
		debug.PrintStack()
		return importInstance, err, http.StatusInternalServerError
	}
	err = token.GetJSON(this.config.ImportDeployUrl+"/instances/"+url.PathEscape(importId), &importInstance)
	if err != nil {
		debug.PrintStack()
		return importInstance, err, http.StatusBadGateway
	}
	return importInstance, nil, 200
}

func (this *Imports) GetImportType(user string, importTypeId string) (importInstance importmodel.ImportType, err error, code int) {
	token, err := this.auth.GetUserToken(user)
	if err != nil {
		debug.PrintStack()
		return importInstance, err, http.StatusInternalServerError
	}
	err = token.GetJSON(this.config.ImportRepositoryUrl+"/import-types/"+url.PathEscape(importTypeId), &importInstance)
	if err != nil {
		debug.PrintStack()
		return importInstance, err, http.StatusBadGateway
	}
	return importInstance, nil, 200
}
