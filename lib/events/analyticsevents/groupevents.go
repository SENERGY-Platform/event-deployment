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

package analyticsevents

import (
	"github.com/SENERGY-Platform/event-deployment/lib/auth"
	"github.com/SENERGY-Platform/event-deployment/lib/model"
	"log"
)

func (this *Events) UpdateDeviceGroup(owner string, group model.DeviceGroup) error {
	pipelines, groupInfos, labels, err := this.analytics.GetPipelinesByDeviceGroupId(owner, group.Id)
	if err != nil {
		log.Println("unable to get pipelines for device-group", owner, group.Id, err)
		return err
	}
	token, err := auth.NewAuth(this.config).GetUserToken(owner)
	if err != nil {
		return err
	}
	for _, pipeline := range pipelines {
		name := labels[pipeline]
		info := groupInfos[pipeline]
		info.DeviceIds = group.DeviceIds
		err = this.updateEventPipelineForDeviceGroup(token, pipeline, name, owner, info)
		if err != nil {
			return err
		}
	}
	return nil
}
