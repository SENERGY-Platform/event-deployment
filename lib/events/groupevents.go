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
	"github.com/SENERGY-Platform/event-deployment/lib/auth"
	"github.com/SENERGY-Platform/event-deployment/lib/model"
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
	case "RIGHTS":
		return nil
	case "PUT":
		err = this.updateDeviceGroup(cmd.DeviceGroup)
		if errors.Is(err, auth.ErrUserDoesNotExist) {
			log.Printf("WARNING: user does not exist -> device-group update will be ignored\n")
			return nil
		}
		return err
	case "DELETE":
		log.Println("ignore device-group delete")
		return nil
	default:
		return errors.New("unknown command " + cmd.Command)
	}
}

type DeviceGroupCommand struct {
	Command     string            `json:"command"`
	Id          string            `json:"id"`
	DeviceGroup model.DeviceGroup `json:"device_group"`
}

func (this *Events) updateDeviceGroup(group model.DeviceGroup) (err error) {
	for _, h := range this.handlers {
		err = h.UpdateDeviceGroup(group)
		if err != nil {
			return err
		}
	}
	return nil
}
