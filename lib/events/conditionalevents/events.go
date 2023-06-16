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
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/SENERGY-Platform/event-deployment/lib/events/conditionalevents/deployments"
	"github.com/SENERGY-Platform/event-deployment/lib/events/conditionalevents/idmodifier"
	"github.com/SENERGY-Platform/event-deployment/lib/interfaces"
	"github.com/SENERGY-Platform/event-deployment/lib/metrics"
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
	config      config.Config
	db          *mongo.Mongo
	transformer *Transformer
	deployments *deployments.Deployments
	mux         sync.Mutex
	metrics     *metrics.Metrics
}

func New(ctx context.Context, config config.Config, devices interfaces.Devices, imports interfaces.Imports, m *metrics.Metrics) (result *Events, err error) {
	result = &Events{config: config, transformer: NewTransformer(devices, imports), metrics: m}
	result.deployments, err = deployments.New(ctx, &sync.WaitGroup{}, config)
	if err != nil {
		return result, err
	}
	result.db, err = mongo.New(ctx, &sync.WaitGroup{}, configuration.Config{
		CloudEventRepoMongoUrl:            config.ConditionalEventRepoMongoUrl,
		CloudEventRepoMongoTable:          config.ConditionalEventRepoMongoTable,
		CloudEventRepoMongoDescCollection: config.ConditionalEventRepoMongoDescCollection,
	})
	return result, err
}

func (this *Events) Deploy(owner string, deployment deploymentmodel.Deployment) error {
	this.mux.Lock()
	defer this.mux.Unlock()
	err := this.deployments.SetDeployment(deployment)
	if err != nil {
		return err
	}
	return this.deployEvents(owner, deployment)
}

func (this *Events) deployEvents(owner string, deployment deploymentmodel.Deployment) error {
	err := this.removeEvents(owner, deployment.Id)
	if err != nil {
		return err
	}
	descriptions, err := this.transformer.Transform(owner, deployment)
	if err != nil {
		return err
	}
	for _, element := range descriptions {
		err = this.deployDescription(element)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *Events) Remove(owner string, deploymentId string) error {
	this.mux.Lock()
	defer this.mux.Unlock()
	err := this.deployments.RemoveDeployment(deploymentId)
	if err != nil {
		return err
	}
	return this.removeEvents(owner, deploymentId)
}

func (this *Events) removeEvents(owner string, deploymentId string) error {
	count, err := this.db.RemoveEventDescriptionsByDeploymentId(deploymentId)
	this.metrics.RemovedConditionalEvents.Add(float64(count))
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

func (this *Events) deployDescription(desc model.EventDesc) error {
	this.metrics.DeployedConditionalEvents.Inc()
	desc.DeviceId, _ = idmodifier.SplitModifier(desc.DeviceId)
	desc.ServiceId, _ = idmodifier.SplitModifier(desc.ServiceId)
	return this.db.SetEventDescription(desc)
}
