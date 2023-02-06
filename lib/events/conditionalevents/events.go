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

package conditionalevents

import (
	"context"
	"errors"
	"github.com/SENERGY-Platform/event-deployment/lib/auth"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/SENERGY-Platform/event-deployment/lib/events/conditionalevents/idmodifier"
	"github.com/SENERGY-Platform/event-deployment/lib/interfaces"
	"github.com/SENERGY-Platform/event-worker/pkg/configuration"
	"github.com/SENERGY-Platform/event-worker/pkg/eventrepo/cloud/mongo"
	"github.com/SENERGY-Platform/event-worker/pkg/model"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel"
	"log"
	"net/http"
	"runtime/debug"
	"sync"
)

type Events struct {
	config  config.Config
	db      *mongo.Mongo
	devices interfaces.Devices
	imports interfaces.Imports
}

func New(ctx context.Context, config config.Config, devices interfaces.Devices, imports interfaces.Imports) (result *Events, err error) {
	result = &Events{config: config, devices: devices, imports: imports}
	result.db, err = mongo.New(ctx, &sync.WaitGroup{}, configuration.Config{
		CloudEventRepoMongoUrl:            config.ConditionalEventRepoMongoUrl,
		CloudEventRepoMongoTable:          config.ConditionalEventRepoMongoTable,
		CloudEventRepoMongoDescCollection: config.ConditionalEventRepoMongoDescCollection,
	})
	return result, err
}

func (this *Events) Deploy(owner string, deployment deploymentmodel.Deployment) error {
	err := this.Remove(owner, deployment.Id)
	if err != nil {
		return err
	}
	token, err := auth.NewAuth(this.config).GetUserToken(owner)
	if err != nil {
		return err
	}
	for _, element := range deployment.Elements {
		err = this.deployElement(token, owner, deployment.Id, element)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *Events) deployElement(token auth.AuthToken, owner string, deploymentId string, element deploymentmodel.Element) (err error) {
	event := element.ConditionalEvent
	if event != nil && event.Selection.FilterCriteria.CharacteristicId != nil {
		if event.Selection.SelectedDeviceGroupId != nil && *event.Selection.SelectedDeviceGroupId != "" {
			return this.deployEventForDeviceGroup(token, owner, deploymentId, event)
		}
		if event.Selection.SelectedDeviceId != nil && event.Selection.SelectedServiceId != nil && *event.Selection.SelectedServiceId != "" {
			return this.deployEventForDevice(token, owner, deploymentId, event)
		}
		if event.Selection.SelectedDeviceId != nil && !(event.Selection.SelectedServiceId != nil && *event.Selection.SelectedServiceId != "") {
			return this.deployEventForDeviceWithoutService(token, owner, deploymentId, event)
		}
		if event.Selection.SelectedImportId != nil {
			return this.deployEventForImport(token, owner, deploymentId, event)
		}
		if event.Selection.SelectedGenericEventSource != nil {
			log.Println("WARNING: generic event sources not supported for conditional events")
			return nil
		}
	}
	return nil
}

func (this *Events) Remove(owner string, deploymentId string) error {
	err := this.db.RemoveEventDescriptionsByDeploymentId(deploymentId)
	return err
}

func (this *Events) CheckEvent(token string, id string) int {
	desc, err := this.db.GetEventDescriptionsByEventId(id)
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return http.StatusInternalServerError
	}
	if len(desc) == 0 {
		return http.StatusNotFound
	}
	return http.StatusOK
}

func (this *Events) GetEventStates(token string, ids []string) (states map[string]bool, err error, code int) {
	for _, id := range ids {
		state := this.CheckEvent(token, id)
		if state == http.StatusInternalServerError {
			return states, errors.New("unable to get event state"), state
		}
		if state == http.StatusNotFound {
			states[id] = false
		}
		if state == http.StatusOK {
			states[id] = true
		}
	}
	return states, nil, http.StatusOK
}

func (this *Events) DeviceTypeUpdate() {
	//TODO
}

func (this *Events) deployDescription(desc model.EventDesc) error {
	desc.DeviceId, _ = idmodifier.SplitModifier(desc.DeviceId)
	desc.ServiceId, _ = idmodifier.SplitModifier(desc.ServiceId)
	return this.db.SetEventDescription(desc)
}
