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

package events

import (
	"encoding/json"
	"errors"
	"log"
	"runtime/debug"
)

func (this *Events) HandleDeviceGroupUpdate(msg []byte) error {
	if this.config.Debug {
		log.Println("DEBUG: receive device-group command:", string(msg))
	}
	cmd := DeviceGroupCommand{}
	err := json.Unmarshal(msg, &cmd)
	if err != nil {
		debug.PrintStack()
		return err
	}
	switch cmd.Command {
	case "PUT":
		return this.updateDeviceGroup(cmd.Owner, cmd.Id)
	case "DELETE":
		log.Println("ignore device-group delete")
		return nil
	default:
		return errors.New("unknown command " + cmd.Command)
	}
}

type DeviceGroupCommand struct {
	Command string `json:"command"`
	Id      string `json:"id"`
	Owner   string `json:"owner"`
}

func (this *Events) updateDeviceGroup(owner string, groupId string) error {
	pipelines, groupInfos, labels, err := this.analytics.GetPipelinesByDeviceGroupId(owner, groupId)
	if err != nil {
		log.Println("unable to get pipelines for device-group", owner, groupId, err)
		return err
	}
	for _, pipeline := range pipelines {
		err = this.analytics.Remove(owner, pipeline)
		if err != nil {
			log.Println("unable to remove pipelines for device-group", owner, groupId, pipeline, err)
			return err
		}
		name := labels[pipeline]
		info := groupInfos[pipeline]
		err = this.deployEventForDeviceGroupWithDescription(name, owner, info)
		if err != nil {
			return err
		}
	}
	return nil
}
