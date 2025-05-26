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
	"github.com/SENERGY-Platform/event-deployment/lib/events/conditionalevents/deployments"
	eventworkermodel "github.com/SENERGY-Platform/event-worker/pkg/model"
)

func (this *Events) UpdateDeviceGroup(groupId string) error {
	this.mux.Lock()
	defer this.mux.Unlock()
	deploymentList, err := this.deployments.GetDeploymentByDeviceGroupId(groupId)
	if err != nil {
		return err
	}
	descr, err := this.db.GetEventDescriptionsByDeviceGroup(groupId)
	if err != nil {
		return err
	}
	for _, depl := range deploymentList {
		err = this.removeEvents(depl.Id)
		if err != nil {
			return err
		}
		if depl.UserId == "" {
			depl.UserId = getFallbackUser(descr, depl)
		}
		err = this.deployEvents(depl.UserId, depl.Deployment)
		if err != nil {
			return err
		}
	}
	return nil
}

func getFallbackUser(descr []eventworkermodel.EventDesc, depl deployments.Deployment) string {
	for _, d := range descr {
		if d.DeploymentId == depl.Id {
			return d.UserId
		}
	}
	return ""
}
