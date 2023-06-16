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
	"context"
	"errors"
	"github.com/SENERGY-Platform/event-deployment/lib/auth"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/SENERGY-Platform/event-deployment/lib/interfaces"
	"github.com/SENERGY-Platform/event-deployment/lib/metrics"
	"github.com/SENERGY-Platform/event-deployment/lib/model"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel"
	"log"
	"net/http"
	"runtime/debug"
	"sort"
)

type Events struct {
	config    config.Config
	analytics interfaces.Analytics
	devices   interfaces.Devices
	imports   interfaces.Imports
	metrics   *metrics.Metrics
}

func New(ctx context.Context, config config.Config, analytics interfaces.Analytics, devices interfaces.Devices, imports interfaces.Imports, m *metrics.Metrics) (result *Events, err error) {
	return &Events{config: config, analytics: analytics, devices: devices, imports: imports, metrics: m}, err
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
	this.metrics.DeployedAnalyticsEvents.Inc()
	event := element.MessageEvent
	if event != nil && event.Selection.FilterCriteria.CharacteristicId != nil {
		label := element.Name + " (" + event.EventId + ")"
		if event.Selection.SelectedDeviceGroupId != nil && *event.Selection.SelectedDeviceGroupId != "" {
			return this.deployEventForDeviceGroup(token, label, owner, deploymentId, event)
		}
		if event.Selection.SelectedDeviceId != nil && event.Selection.SelectedServiceId != nil && *event.Selection.SelectedServiceId != "" {
			return this.deployEventForDevice(token, label, owner, deploymentId, event)
		}
		if event.Selection.SelectedDeviceId != nil && !(event.Selection.SelectedServiceId != nil && *event.Selection.SelectedServiceId != "") {
			return this.deployEventForDeviceWithoutService(token, label, owner, deploymentId, event)
		}
		if event.Selection.SelectedImportId != nil {
			return this.deployEventForImport(token, label, owner, deploymentId, event)
		}
		if event.Selection.SelectedGenericEventSource != nil {
			return this.deployEventForGenericSource(token, label, owner, deploymentId, event)
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
		this.metrics.RemovedAnalyticsEvents.Inc()
		err = this.analytics.Remove(owner, id)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *Events) CheckEvent(token string, id string) int {
	userId, err := GetUserId(token)
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return http.StatusBadRequest
	}
	_, exists, err := this.analytics.GetPipelineByEventId(userId, id)
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

func (this *Events) GetEventStates(token string, ids []string) (states map[string]bool, err error, code int) {
	userId, err := GetUserId(token)
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return states, err, http.StatusBadRequest
	}
	states = map[string]bool{}
	if len(ids) == 0 {
		return states, nil, http.StatusOK
	}
	states, err = this.analytics.GetEventStates(userId, ids)
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return states, err, http.StatusInternalServerError
	}
	return states, nil, http.StatusOK
}

var ErrMissingCharacteristicInEvent = errors.New("missing characteristic id in event")

// expects event.Selection.SelectedDeviceId and event.Selection.SelectedServiceId to be set
func (this *Events) deployEventForDevice(token auth.AuthToken, label string, owner string, deploymentId string, event *deploymentmodel.MessageEvent) error {
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
	castExtensions := []model.ConverterExtension{}

	if event.Selection.SelectedPath != nil {
		castFrom = event.Selection.SelectedPath.CharacteristicId
		path = this.config.DevicePathPrefix + event.Selection.SelectedPath.Path

		//find cast extensions
		if event.Selection.FilterCriteria.FunctionId != nil {
			function, err, code := this.devices.GetFunction(*event.Selection.FilterCriteria.FunctionId)
			if err != nil {
				if code != http.StatusNotFound {
					//ignore not found errors to prevent unresolvable kafka consumption loop
					return err
				}
			} else if function.ConceptId != "" {
				concept, err, code := this.devices.GetConcept(function.ConceptId)
				if err != nil {
					if code != http.StatusNotFound {
						//ignore not found errors to prevent unresolvable kafka consumption loop
						return err
					}
				} else {
					castExtensions = concept.Conversions
				}
			}
		}

	} else {
		log.Println("WARNING: missing SelectedPath --> ignore event", event)
		return nil
	}
	var pipelineId string
	var err error
	if event.UseMarshaller {
		functionId := ""
		if event.Selection.FilterCriteria.FunctionId != nil {
			functionId = *event.Selection.FilterCriteria.FunctionId
		}
		aspectNodeId := ""
		if event.Selection.FilterCriteria.AspectId != nil {
			aspectNodeId = *event.Selection.FilterCriteria.AspectId
		}
		serializedPath := ""
		if event.Selection.SelectedPath != nil {
			serializedPath = event.Selection.SelectedPath.Path
		}
		pipelineId, err = this.analytics.DeployDeviceWithMarshaller(
			token,
			label,
			owner,
			deploymentId,
			event.FlowId,
			event.EventId,
			*event.Selection.SelectedDeviceId,
			*event.Selection.SelectedServiceId,
			event.Value,
			serializedPath,
			functionId,
			aspectNodeId,
			*event.Selection.FilterCriteria.CharacteristicId)
	} else {
		pipelineId, err = this.analytics.DeployDevice(
			token,
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
			*event.Selection.FilterCriteria.CharacteristicId,
			castExtensions)
	}

	if err != nil {
		return err
	}
	if pipelineId == "" {
		log.Println("WARNING: event not deployed in analytics -> ignore event", event)
		return nil
	}
	return nil
}

func (this *Events) deployEventForDeviceGroup(token auth.AuthToken, label string, owner string, deploymentId string, event *deploymentmodel.MessageEvent) error {
	if !this.DeviceGroupsAndImportsEnabled() {
		log.Println("WARNING: DeviceGroupsAndImportsEnabled() = false; configure AuthClientId, AuthClientSecret, AuthEndpoint, PermSearchUrl")
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

	return this.deployEventForDeviceGroupWithDescription(token, label, owner, model.GroupEventDescription{
		DeviceGroupId:    *event.Selection.SelectedDeviceGroupId,
		EventId:          event.EventId,
		DeploymentId:     deploymentId,
		CharacteristicId: characteristicId,
		FunctionId:       *event.Selection.FilterCriteria.FunctionId,
		AspectId:         *event.Selection.FilterCriteria.AspectId,
		FlowId:           event.FlowId,
		OperatorValue:    event.Value,
		UseMarshaller:    event.UseMarshaller,
	})
}

func (this *Events) deployEventForDeviceGroupWithDescription(token auth.AuthToken, label string, owner string, desc model.GroupEventDescription) error {
	if !this.DeviceGroupsAndImportsEnabled() {
		log.Println("WARNING: DeviceGroupsAndImportsEnabled() = false; configure AuthClientId, AuthClientSecret, AuthEndpoint, PermSearchUrl")
		return nil
	}
	if desc.DeviceGroupId == "" && len(desc.DeviceIds) == 0 {
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
		log.Println("WARNING: getServicesPathsAndDevicesForEvent()", code, err)
		if code == http.StatusNotFound {
			return nil //ignore
		}
		return err
	}

	//find cast extensions
	castExtensions := []model.ConverterExtension{}
	function, err, code := this.devices.GetFunction(desc.FunctionId)
	if err != nil {
		if code != http.StatusNotFound {
			//ignore not found errors to prevent unresolvable kafka consumption loop
			return err
		}
	} else if function.ConceptId != "" {
		concept, err, code := this.devices.GetConcept(function.ConceptId)
		if err != nil {
			if code != http.StatusNotFound {
				//ignore not found errors to prevent unresolvable kafka consumption loop
				return err
			}
		} else {
			castExtensions = concept.Conversions
		}
	}

	pipelineId, err := this.analytics.DeployGroup(
		token,
		label,
		owner,
		desc,
		serviceIds,
		serviceToDevices,
		serviceToPath,
		serviceToPathAndCharacteristic,
		castExtensions,
		desc.UseMarshaller)
	if err != nil {
		return err
	}
	if pipelineId == "" {
		log.Println("WARNING: event not deployed in analytics -> ignore event", desc)
		return nil
	}
	return nil
}

func (this *Events) updateEventPipelineForDeviceGroup(token auth.AuthToken, pipelineId string, label string, owner string, desc model.GroupEventDescription) error {
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
		log.Println("WARNING: getServicesPathsAndDevicesForEvent()", code, err)
		if code == http.StatusNotFound {
			return nil //ignore
		}
		return err
	}

	//find cast extensions
	castExtensions := []model.ConverterExtension{}
	function, err, code := this.devices.GetFunction(desc.FunctionId)
	if err != nil {
		if code != http.StatusNotFound {
			//ignore not found errors to prevent unresolvable kafka consumption loop
			return err
		}
	} else if function.ConceptId != "" {
		concept, err, code := this.devices.GetConcept(function.ConceptId)
		if err != nil {
			if code != http.StatusNotFound {
				//ignore not found errors to prevent unresolvable kafka consumption loop
				return err
			}
		} else {
			castExtensions = concept.Conversions
		}
	}

	err = this.analytics.UpdateGroupDeployment(
		token,
		pipelineId,
		label,
		owner,
		desc,
		serviceIds,
		serviceToDevices,
		serviceToPath,
		serviceToPathAndCharacteristic,
		castExtensions,
		desc.UseMarshaller)
	if err != nil {
		return err
	}
	return nil
}

const IdParameterSeperator = "$"

func (this *Events) getServicesPathsAndDevicesForEvent(desc model.GroupEventDescription) (serviceIds []string, serviceToDevices map[string][]string, serviceToPath map[string][]string, serviceToPathAndCharacteristic map[string][]model.PathAndCharacteristic, err error, code int) {
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
	options, err := this.getDeviceGroupPathOptions(desc, deviceTypeIds)
	if err != nil {
		log.Println("ERROR: unable to find path options", err)
		return nil, nil, nil, serviceToPathAndCharacteristic, err, http.StatusInternalServerError
	}
	serviceIds = []string{}
	serviceToDevices = map[string][]string{}
	serviceToPath = map[string][]string{}
	serviceToPathToCharacteristic := map[string]map[string]string{}
	for _, device := range devices {
		for _, option := range options[device.DeviceTypeId] {
			if len(option.JsonPath) > 0 {
				serviceToDevices[option.ServiceId] = append(serviceToDevices[option.ServiceId], device.Id)
				if _, ok := serviceToPath[option.ServiceId]; !ok {
					serviceIds = append(serviceIds, option.ServiceId)
				}
				for _, path := range option.JsonPath {
					serviceToPath[option.ServiceId] = append(serviceToPath[option.ServiceId], path)
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
	sort.Strings(serviceIds)
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

func (this *Events) deployEventForImport(token auth.AuthToken, label string, owner string, deploymentId string, event *deploymentmodel.MessageEvent) error {
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
	path := ""
	castExtensions := []model.ConverterExtension{}

	if event.Selection.SelectedPath != nil {
		castFrom = event.Selection.SelectedPath.CharacteristicId
		path = event.Selection.SelectedPath.Path

		//find cast extensions
		if event.Selection.FilterCriteria.FunctionId != nil {
			function, err, code := this.devices.GetFunction(*event.Selection.FilterCriteria.FunctionId)
			if err != nil {
				if code != http.StatusNotFound {
					//ignore not found errors to prevent unresolvable kafka consumption loop
					return err
				}
			} else if function.ConceptId != "" {
				concept, err, code := this.devices.GetConcept(function.ConceptId)
				if err != nil {
					if code != http.StatusNotFound {
						//ignore not found errors to prevent unresolvable kafka consumption loop
						return err
					}
				} else {
					castExtensions = concept.Conversions
				}
			}
		}
	}
	return this.deployEventForImportWithDescription(token, label, owner, model.GroupEventDescription{
		ImportId:      *event.Selection.SelectedImportId,
		EventId:       event.EventId,
		DeploymentId:  deploymentId,
		FunctionId:    *event.Selection.FilterCriteria.FunctionId,
		AspectId:      *event.Selection.FilterCriteria.AspectId,
		FlowId:        event.FlowId,
		OperatorValue: event.Value,
		Path:          path,
	}, castFrom, *event.Selection.FilterCriteria.CharacteristicId, castExtensions)
}

func (this *Events) deployEventForGenericSource(token auth.AuthToken, label string, owner string, deploymentId string, event *deploymentmodel.MessageEvent) error {
	if event == nil {
		debug.PrintStack()
		return errors.New("missing event element") //programming error -> dont ignore
	}
	if event.Selection.SelectedGenericEventSource == nil {
		debug.PrintStack()
		return errors.New("missing selected_generic_source") //programming error -> dont ignore
	}
	if event.Selection.SelectedGenericEventSource.FilterType == "" {
		log.Println("WARNING: try to deploy selected_generic_source event without filter_type --> ignore", label, deploymentId, event)
		return nil
	}
	if event.Selection.SelectedGenericEventSource.FilterIds == "" {
		log.Println("WARNING: try to deploy selected_generic_source event without filter_ids --> ignore", label, deploymentId, event)
		return nil
	}
	if event.Selection.SelectedGenericEventSource.Topic == "" {
		log.Println("WARNING: try to deploy selected_generic_source event without topic --> ignore", label, deploymentId, event)
		return nil
	}
	if event.Selection.SelectedPath == nil || event.Selection.SelectedPath.Path == "" {
		log.Println("WARNING: try to deploy selected_generic_source event without path --> ignore", label, deploymentId, event)
		return nil
	}

	var castFrom string
	path := event.Selection.SelectedPath.Path
	castExtensions := []model.ConverterExtension{}

	if event.Selection.SelectedPath.CharacteristicId != "" {
		castFrom = event.Selection.SelectedPath.CharacteristicId

		//find cast extensions
		if event.Selection.FilterCriteria.FunctionId != nil {
			function, err, code := this.devices.GetFunction(*event.Selection.FilterCriteria.FunctionId)
			if err != nil {
				if code != http.StatusNotFound {
					//ignore not found errors to prevent unresolvable kafka consumption loop
					return err
				}
			} else if function.ConceptId != "" {
				concept, err, code := this.devices.GetConcept(function.ConceptId)
				if err != nil {
					if code != http.StatusNotFound {
						//ignore not found errors to prevent unresolvable kafka consumption loop
						return err
					}
				} else {
					castExtensions = concept.Conversions
				}
			}
		}
	}
	return this.deployEventForGenericSourceWithDescription(token, label, owner, model.GroupEventDescription{
		GenericEventSource: event.Selection.SelectedGenericEventSource,
		EventId:            event.EventId,
		DeploymentId:       deploymentId,
		FunctionId:         *event.Selection.FilterCriteria.FunctionId,
		AspectId:           *event.Selection.FilterCriteria.AspectId,
		FlowId:             event.FlowId,
		OperatorValue:      event.Value,
		Path:               path,
	}, castFrom, *event.Selection.FilterCriteria.CharacteristicId, castExtensions)
}

func (this *Events) deployEventForImportWithDescription(token auth.AuthToken, label string, owner string, desc model.GroupEventDescription, castFrom string, castTo string, castExtensions []model.ConverterExtension) error {
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
		token,
		label,
		owner,
		desc,
		topic,
		this.config.ImportPathPrefix+desc.Path,
		castFrom,
		castTo,
		castExtensions)
	if err != nil {
		return err
	}
	if pipelineId == "" {
		log.Println("WARNING: event not deployed in analytics -> ignore event", desc)
		return nil
	}
	return nil
}

func (this *Events) deployEventForGenericSourceWithDescription(token auth.AuthToken, label string, owner string, desc model.GroupEventDescription, castFrom string, castTo string, castExtensions []model.ConverterExtension) error {
	if !this.DeviceGroupsAndImportsEnabled() {
		return nil
	}
	if desc.GenericEventSource == nil {
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
	pipelineId, err := this.analytics.DeployGenericSource(
		token,
		label,
		owner,
		desc,
		this.config.GenericSourcePathPrefix+desc.Path,
		castFrom,
		castTo,
		castExtensions)
	if err != nil {
		return err
	}
	if pipelineId == "" {
		log.Println("WARNING: event not deployed in analytics -> ignore event", desc)
		return nil
	}
	return nil
}

func (this *Events) deployEventForDeviceWithoutService(token auth.AuthToken, label string, owner string, deploymentId string, event *deploymentmodel.MessageEvent) error {
	if !this.DeviceGroupsAndImportsEnabled() {
		log.Println("WARNING: DeviceGroupsAndImportsEnabled() = false; configure AuthClientId, AuthClientSecret, AuthEndpoint, PermSearchUrl")
		return nil
	}
	if event == nil {
		debug.PrintStack()
		return errors.New("missing event element") //programming error -> dont ignore
	}
	if event.Selection.SelectedDeviceId == nil {
		debug.PrintStack()
		return errors.New("missing device id") //programming error -> dont ignore
	}
	if event.Selection.FilterCriteria.FunctionId == nil {
		log.Println("WARNING: try to deploy device (no service selection) event without function id --> ignore", label, deploymentId, event)
		return nil
	}
	if event.Selection.FilterCriteria.AspectId == nil {
		log.Println("WARNING: try to deploy (no service selection) event without aspect id --> ignore", label, deploymentId, event)
		return nil
	}
	characteristicId := ""
	if event.Selection.FilterCriteria.CharacteristicId != nil {
		characteristicId = *event.Selection.FilterCriteria.CharacteristicId
	}

	return this.deployEventForDeviceGroupWithDescription(token, label, owner, model.GroupEventDescription{
		DeviceIds:        []string{*event.Selection.SelectedDeviceId},
		EventId:          event.EventId,
		DeploymentId:     deploymentId,
		CharacteristicId: characteristicId,
		FunctionId:       *event.Selection.FilterCriteria.FunctionId,
		AspectId:         *event.Selection.FilterCriteria.AspectId,
		FlowId:           event.FlowId,
		OperatorValue:    event.Value,
		UseMarshaller:    event.UseMarshaller,
	})
}
