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

package mocks

import (
	"github.com/SENERGY-Platform/process-deployment/lib/model/importmodel"
	"net/http"
	"strings"
)

type ImportsMock struct{}

func (this *ImportsMock) GetImportInstance(user string, importId string) (importInstance importmodel.Import, err error, code int) {
	//TODO implement me
	panic("implement me")
}

func (this *ImportsMock) GetImportType(user string, importTypeId string) (importInstance importmodel.ImportType, err error, code int) {
	//TODO implement me
	panic("implement me")
}

func (this *ImportsMock) GetTopic(_ string, importId string) (topic string, err error, code int) {
	return strings.ReplaceAll(importId, ":", "_"), nil, http.StatusOK
}
