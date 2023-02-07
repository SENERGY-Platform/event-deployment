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
	"github.com/SENERGY-Platform/event-deployment/lib/model"
)

func (this *Events) UpdateDeviceGroup(owner string, group model.DeviceGroup) error {
	this.mux.Lock()
	defer this.mux.Unlock()
	deployments, err := this.deployments.GetDeploymentByDeviceGroupId(group.Id)
	if err != nil {
		return err
	}
	for _, depl := range deployments {
		err = this.removeEvents(owner, depl.Id)
		if err != nil {
			return err
		}
		err = this.deployEvents(owner, depl)
		if err != nil {
			return err
		}
	}
	return nil
}
