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
	"context"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/SENERGY-Platform/event-deployment/lib/interfaces"
	"github.com/SENERGY-Platform/event-deployment/lib/marshaller"
	"github.com/SENERGY-Platform/event-deployment/lib/model"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel/v2"
	"github.com/SENERGY-Platform/process-deployment/lib/model/messages"
	jwt_http_router "github.com/SmartEnergyPlatform/jwt-http-router"
	"log"
	"net/http"
	"runtime/debug"
	"sort"
)

type EventsFactory struct{}

var Factory = &EventsFactory{}

type Events struct {
	config     config.Config
	analytics  interfaces.Analytics
	marshaller interfaces.Marshaller
	devices    interfaces.Devices
	imports    interfaces.Imports
}

func (this *EventsFactory) New(ctx context.Context, config config.Config, analytics interfaces.Analytics, marshaller interfaces.Marshaller, devices interfaces.Devices, imports interfaces.Imports) (interfaces.Events, error) {
	return &Events{config: config, analytics: analytics, marshaller: marshaller, devices: devices, imports: imports}, nil
}

func (this *Events) HandleCommand(msg []byte) error {
	if this.config.Debug {
		log.Println("DEBUG: receive deployment command:", string(msg))
	}
	cmd := messages.DeploymentCommand{}
	err := json.Unmarshal(msg, &cmd)
	if err != nil {
		debug.PrintStack()
		return err
	}
	switch cmd.Command {
	case "PUT":
		if cmd.DeploymentV2 != nil {
			err = this.Deploy(cmd.Owner, *cmd.DeploymentV2)
		}
		return err
	case "DELETE":
		return this.Remove(cmd.Owner, cmd.Id)
	default:
		return errors.New("unknown command " + cmd.Command)
	}
	return nil
}

func (this *Events) Deploy(owner string, deployment deploymentmodel.Deployment) error {
	err := this.Remove(owner, deployment.Id)
	if err != nil {
		return err
	}
	for _, element := range deployment.Elements {
		err = this.deployElement(owner, deployment.Id, element)
	}
	return nil
}

func (this *Events) deployElement(owner string, deploymentId string, element deploymentmodel.Element) (err error) {
	event := element.MessageEvent
	if event != nil && event.Selection.FilterCriteria.CharacteristicId != nil {
		label := element.Name + " (" + event.EventId + ")"
		if event.Selection.SelectedDeviceGroupId != nil {
			return this.deployEventForDeviceGroup(label, owner, deploymentId, event)
		}
		if event.Selection.SelectedDeviceId != nil && event.Selection.SelectedServiceId != nil {
			return this.deployEventForDevice(label, owner, deploymentId, event)
		}
		if event.Selection.SelectedImportId != nil {
			return this.deployEventForImport(label, owner, deploymentId, event)
		}
	}
	return nil
}

func (this *Events) Remove(owner string, deploymentId string) error {
	pipelineIds, err := this.analytics.GetPipelinesByDeploymentId(owner, deploymentId)
	if err != nil {
		return err
	}
	for _, id := range pipelineIds {
		err = this.analytics.Remove(owner, id)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *Events) CheckEvent(jwt jwt_http_router.Jwt, id string) int {
	_, exists, err := this.analytics.GetPipelineByEventId(jwt.UserId, id)
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return http.StatusInternalServerError
	}
	if !exists {
		return http.StatusNotFound
	}
	return http.StatusOK
}

func (this *Events) GetEventStates(jwt jwt_http_router.Jwt, ids []string) (states map[string]bool, err error, code int) {
	states = map[string]bool{}
	if len(ids) == 0 {
		return states, nil, http.StatusOK
	}
	states, err = this.analytics.GetEventStates(jwt.UserId, ids)
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return states, err, http.StatusInternalServerError
	}
	return states, nil, http.StatusOK
}

var ErrMissingCharacteristicInEvent = errors.New("missing characteristic id in event")

func (this *Events) GetPathAndCharacteristicForEvent(event *deploymentmodel.MessageEvent) (path string, characteristicId string, err error) {
	if event.Selection.FilterCriteria.CharacteristicId == nil {
		return "", "", ErrMissingCharacteristicInEvent
	}
	if event.Selection.SelectedServiceId == nil {
		err = errors.New("missing service id")
		debug.PrintStack()
		return
	}
	return this.marshaller.FindPath(*event.Selection.SelectedServiceId, *event.Selection.FilterCriteria.CharacteristicId)
}

//expects event.Selection.SelectedDeviceId and event.Selection.SelectedServiceId to be set
func (this *Events) deployEventForDevice(label string, owner string, deploymentId string, event *deploymentmodel.MessageEvent) error {
	if event == nil {
		debug.PrintStack()
		return errors.New("missing event element") //programming error -> dont ignore
	}
	if event.Selection.SelectedDeviceId == nil {
		debug.PrintStack()
		return errors.New("missing device id") //programming error -> dont ignore
	}
	if event.Selection.SelectedServiceId == nil {
		debug.PrintStack()
		return errors.New("missing service id") //programming error -> dont ignore
	}
	var path string
	var castFrom string
	var err error
	if event.Selection.SelectedCharacteristicId != nil && event.Selection.SelectedPath != nil {
		castFrom = *event.Selection.SelectedCharacteristicId
		path = *event.Selection.SelectedPath
	} else {
		path, castFrom, err = this.GetPathAndCharacteristicForEvent(event)
	}
	if err == ErrMissingCharacteristicInEvent || err == marshaller.ErrCharacteristicNotFoundInService || err == marshaller.ErrServiceNotFound {
		log.Println("WARNING: error on marshaller request;", err, "-> ignore event", event)
		return nil
	}
	if err != nil {
		return err
	}
	pipelineId, err := this.analytics.Deploy(
		label,
		owner,
		deploymentId,
		event.FlowId,
		event.EventId,
		*event.Selection.SelectedDeviceId,
		*event.Selection.SelectedServiceId,
		event.Value,
		path,
		castFrom,
		*event.Selection.FilterCriteria.CharacteristicId)
	if err != nil {
		return err
	}
	if pipelineId == "" {
		log.Println("WARNING: event not deployed in analytics -> ignore event", event)
		return nil
	}
	return nil
}

func (this *Events) deployEventForDeviceGroup(label string, owner string, deploymentId string, event *deploymentmodel.MessageEvent) error {
	if !this.DeviceGroupsAndImportsEnabled() {
		return nil
	}
	if event == nil {
		debug.PrintStack()
		return errors.New("missing event element") //programming error -> dont ignore
	}
	if event.Selection.SelectedDeviceGroupId == nil {
		debug.PrintStack()
		return errors.New("missing device group id") //programming error -> dont ignore
	}
	if event.Selection.FilterCriteria.FunctionId == nil {
		log.Println("WARNING: try to deploy group event without function id --> ignore", label, deploymentId, event)
		return nil
	}
	if event.Selection.FilterCriteria.AspectId == nil {
		log.Println("WARNING: try to deploy group event without aspect id --> ignore", label, deploymentId, event)
		return nil
	}
	characteristicId := ""
	if event.Selection.FilterCriteria.CharacteristicId != nil {
		characteristicId = *event.Selection.FilterCriteria.CharacteristicId
	}

	return this.deployEventForDeviceGroupWithDescription(label, owner, model.GroupEventDescription{
		DeviceGroupId:    *event.Selection.SelectedDeviceGroupId,
		EventId:          event.EventId,
		DeploymentId:     deploymentId,
		CharacteristicId: characteristicId,
		FunctionId:       *event.Selection.FilterCriteria.FunctionId,
		AspectId:         *event.Selection.FilterCriteria.AspectId,
		FlowId:           event.FlowId,
		OperatorValue:    event.Value,
	})
}

func (this *Events) deployEventForDeviceGroupWithDescription(label string, owner string, desc model.GroupEventDescription) error {
	if !this.DeviceGroupsAndImportsEnabled() {
		return nil
	}
	if desc.DeviceGroupId == "" {
		debug.PrintStack()
		return errors.New("missing device group id") //programming error -> dont ignore
	}
	if desc.FunctionId == "" {
		log.Println("WARNING: try to deploy group event without function id --> ignore", label, desc)
		return nil
	}
	if desc.AspectId == "" {
		log.Println("WARNING: try to deploy group event without aspect id --> ignore", label, desc)
		return nil
	}
	if desc.DeploymentId == "" {
		log.Println("WARNING: try to deploy group event without deployment id --> ignore", label, desc)
		return nil
	}
	serviceIds, serviceToDevices, serviceToPath, serviceToPathAndCharacteristic, err, code := this.getServicesPathsAndDevicesForEvent(desc)
	if err != nil {
		if code == http.StatusNotFound {
			return nil //ignore
		}
		return err
	}
	pipelineId, err := this.analytics.DeployGroup(
		label,
		owner,
		desc,
		serviceIds,
		serviceToDevices,
		serviceToPath,
		serviceToPathAndCharacteristic)
	if err != nil {
		return err
	}
	if pipelineId == "" {
		log.Println("WARNING: event not deployed in analytics -> ignore event", desc)
		return nil
	}
	return nil
}

func (this *Events) updateEventPipelineForDeviceGroup(pipelineId string, label string, owner string, desc model.GroupEventDescription) error {
	if !this.DeviceGroupsAndImportsEnabled() {
		return nil
	}
	if desc.DeviceGroupId == "" {
		debug.PrintStack()
		return errors.New("missing device group id") //programming error -> dont ignore
	}
	if desc.FunctionId == "" {
		log.Println("WARNING: try to deploy group event without function id --> ignore", label, desc)
		return nil
	}
	if desc.AspectId == "" {
		log.Println("WARNING: try to deploy group event without aspect id --> ignore", label, desc)
		return nil
	}
	if desc.DeploymentId == "" {
		log.Println("WARNING: try to deploy group event without deployment id --> ignore", label, desc)
		return nil
	}

	serviceIds, serviceToDevices, serviceToPath, serviceToPathAndCharacteristic, err, code := this.getServicesPathsAndDevicesForEvent(desc)
	if err != nil {
		if code == http.StatusNotFound {
			return nil //ignore
		}
		return err
	}

	err = this.analytics.UpdateGroupDeployment(
		pipelineId,
		label,
		owner,
		desc,
		serviceIds,
		serviceToDevices,
		serviceToPath,
		serviceToPathAndCharacteristic)
	if err != nil {
		return err
	}
	return nil
}

func (this *Events) getServicesPathsAndDevicesForEvent(desc model.GroupEventDescription) (serviceIds []string, serviceToDevices map[string][]string, serviceToPath map[string]string, serviceToPathAndCharacteristic map[string][]model.PathAndCharacteristic, err error, code int) {
	serviceToPathAndCharacteristic = map[string][]model.PathAndCharacteristic{}
	var devices []model.Device
	var deviceTypeIds []string
	if desc.DeviceIds != nil {
		devices, deviceTypeIds, err, code = this.devices.GetDeviceInfosOfDevices(desc.DeviceIds)
	} else {
		devices, deviceTypeIds, err, code = this.devices.GetDeviceInfosOfGroup(desc.DeviceGroupId)
	}
	if err != nil {
		return nil, nil, nil, serviceToPathAndCharacteristic, err, code
	}
	options, err := this.marshaller.FindPathOptions(
		deviceTypeIds,
		desc.FunctionId,
		desc.AspectId,
		[]string{},
		true)
	if err != nil {
		log.Println("ERROR: unable to find path options", err)
		return nil, nil, nil, serviceToPathAndCharacteristic, err, http.StatusInternalServerError
	}
	serviceIds = []string{}
	serviceToDevices = map[string][]string{}
	serviceToPath = map[string]string{}
	serviceToPathToCharacteristic := map[string]map[string]string{}
	for _, device := range devices {
		for _, option := range options[device.DeviceTypeId] {
			if len(option.JsonPath) > 0 {
				serviceToDevices[option.ServiceId] = append(serviceToDevices[option.ServiceId], device.Id)
				if _, ok := serviceToPath[option.ServiceId]; !ok {
					serviceIds = append(serviceIds, option.ServiceId)
					serviceToPath[option.ServiceId] = option.JsonPath[0]
				}
				for _, path := range option.JsonPath {
					if _, ok := serviceToPathToCharacteristic[option.ServiceId]; !ok {
						serviceToPathToCharacteristic[option.ServiceId] = map[string]string{}
					}
					serviceToPathToCharacteristic[option.ServiceId][path] = option.PathToCharacteristicId[path]
				}
			}
		}
	}
	for serviceId, pathToCharacteristic := range serviceToPathToCharacteristic {
		for path, characteristic := range pathToCharacteristic {
			serviceToPathAndCharacteristic[serviceId] = append(serviceToPathAndCharacteristic[serviceId], model.PathAndCharacteristic{
				JsonPath:         path,
				CharacteristicId: characteristic,
			})
		}
		sort.Slice(serviceToPathAndCharacteristic[serviceId], func(i, j int) bool {
			return serviceToPathAndCharacteristic[serviceId][i].JsonPath < serviceToPathAndCharacteristic[serviceId][j].JsonPath
		})
	}
	return serviceIds, serviceToDevices, serviceToPath, serviceToPathAndCharacteristic, nil, http.StatusOK
}

func (this *Events) DeviceGroupsAndImportsEnabled() bool {
	if this.config.AuthClientId == "" {
		return false
	}
	if this.config.AuthClientSecret == "" {
		return false
	}
	if this.config.AuthEndpoint == "" {
		return false
	}
	if this.config.PermSearchUrl == "" {
		return false
	}
	return true
}

func (this *Events) deployEventForImport(label string, owner string, deploymentId string, event *deploymentmodel.MessageEvent) error {
	if !this.DeviceGroupsAndImportsEnabled() {
		return nil
	}
	if event == nil {
		debug.PrintStack()
		return errors.New("missing event element") //programming error -> dont ignore
	}
	if event.Selection.SelectedImportId == nil {
		debug.PrintStack()
		return errors.New("missing import id") //programming error -> dont ignore
	}
	if event.Selection.FilterCriteria.FunctionId == nil {
		log.Println("WARNING: try to deploy group event without function id --> ignore", label, deploymentId, event)
		return nil
	}
	if event.Selection.FilterCriteria.AspectId == nil {
		log.Println("WARNING: try to deploy group event without aspect id --> ignore", label, deploymentId, event)
		return nil
	}
	var castFrom string
	if event.Selection.SelectedCharacteristicId != nil {
		castFrom = *event.Selection.SelectedCharacteristicId
	}
	return this.deployEventForImportWithDescription(label, owner, model.GroupEventDescription{
		ImportId:      *event.Selection.SelectedImportId,
		EventId:       event.EventId,
		DeploymentId:  deploymentId,
		FunctionId:    *event.Selection.FilterCriteria.FunctionId,
		AspectId:      *event.Selection.FilterCriteria.AspectId,
		FlowId:        event.FlowId,
		OperatorValue: event.Value,
		Path:          *event.Selection.SelectedPath,
	}, castFrom, *event.Selection.FilterCriteria.CharacteristicId)
}

func (this *Events) deployEventForImportWithDescription(label string, owner string, desc model.GroupEventDescription, castFrom string, castTo string) error {
	if !this.DeviceGroupsAndImportsEnabled() {
		return nil
	}
	if desc.ImportId == "" {
		debug.PrintStack()
		return errors.New("missing import id") //programming error -> dont ignore
	}
	if desc.DeploymentId == "" {
		log.Println("WARNING: try to deploy import event without deployment id --> ignore", label, desc)
		return nil
	}
	if desc.Path == "" {
		return errors.New("missing path") //programming error -> dont ignore
	}
	topic, err, _ := this.imports.GetTopic(owner, desc.ImportId)
	if err != nil {
		return err
	}
	pipelineId, err := this.analytics.DeployImport(
		label,
		owner,
		desc,
		topic,
		desc.Path,
		castFrom,
		castTo)
	if err != nil {
		return err
	}
	if pipelineId == "" {
		log.Println("WARNING: event not deployed in analytics -> ignore event", desc)
		return nil
	}
	return nil
}
