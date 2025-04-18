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

package interfaces

import (
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/SENERGY-Platform/models/go/models"
)

type ImportsFactory interface {
	New(config config.Config) Imports
}

type Imports interface {
	GetTopic(user string, importId string) (topic string, err error, code int)
	GetImportInstance(user string, importId string) (importInstance models.Import, err error, code int)
	GetImportType(user string, importTypeId string) (importInstance models.ImportType, err error, code int)
}
