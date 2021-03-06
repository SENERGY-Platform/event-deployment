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

package interfaces

import (
	"context"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/SENERGY-Platform/event-deployment/lib/model"
)

type MarshallerFactory interface {
	New(ctx context.Context, config config.Config) (Marshaller, error)
}

type Marshaller interface {
	FindPath(serviceId string, characteristicId string) (path string, serviceCharacteristicId string, err error)
	FindPathOptions(deviceTypeIds []string, functionId string, aspectId string, characteristicsIdFilter []string, withEnvelope bool) (result map[string][]model.PathOptionsResultElement, err error)
}
