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
	"github.com/SENERGY-Platform/event-worker/pkg/model"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel"
	"github.com/SENERGY-Platform/process-deployment/lib/model/importmodel"
	"log"
	"net/http"
	"runtime/debug"
)

func (this *Transformer) transformEventForImport(owner string, deployentId string, event *deploymentmodel.ConditionalEvent) (result []model.EventDesc, err error) {
	desc := model.EventDesc{
		UserId:        owner,
		DeploymentId:  deployentId,
		ImportId:      *event.Selection.SelectedImportId,
		Script:        event.Script,
		ValueVariable: event.ValueVariable,
		Variables:     event.Variables,
		Qos:           event.Qos,
		EventId:       event.EventId,
	}

	if event.Selection.FilterCriteria.CharacteristicId != nil {
		desc.CharacteristicId = *event.Selection.FilterCriteria.CharacteristicId
	}
	if event.Selection.FilterCriteria.FunctionId != nil {
		desc.FunctionId = *event.Selection.FilterCriteria.FunctionId
	}
	if event.Selection.FilterCriteria.AspectId != nil {
		desc.AspectId = *event.Selection.FilterCriteria.AspectId
	}
	if event.Selection.SelectedPath != nil {
		desc.Path = event.Selection.SelectedPath.Path
	}

	importInstance, err, code := this.imports.GetImportInstance(owner, desc.ImportId)
	if err != nil {
		if code == http.StatusInternalServerError {
			return []model.EventDesc{}, err
		} else {
			log.Println("ERROR:", code, err)
			debug.PrintStack()
			return []model.EventDesc{}, nil //ignore bad request errors
		}
	}
	importType, err, code := this.imports.GetImportType(owner, importInstance.ImportTypeId)
	if err != nil {
		if code == http.StatusInternalServerError {
			return []model.EventDesc{}, err
		} else {
			log.Println("ERROR:", code, err)
			debug.PrintStack()
			return []model.EventDesc{}, nil //ignore bad request errors
		}
	}
	outputs := importVariablesToContents(importType.Output.SubContentVariables)

	service := models.Service{
		Id:          importType.Id,
		Name:        importType.Name,
		Interaction: models.EVENT,
		Outputs:     outputs,
	}
	desc.ServiceForMarshaller = service

	return []model.EventDesc{desc}, nil
}

func importVariablesToContents(variables []importmodel.ImportContentVariable) (result []models.Content) {
	for _, v := range variables {
		result = append(result, importVariableToContent(v))
	}
	return
}

func importVariableToContent(variable importmodel.ImportContentVariable) (result models.Content) {
	result.ContentVariable = importContentVariableToContentVariable(variable)
	return
}

func importContentVariablesToContentVariables(variables []importmodel.ImportContentVariable) (result []models.ContentVariable) {
	for _, v := range variables {
		result = append(result, importContentVariableToContentVariable(v))
	}
	return
}

func importContentVariableToContentVariable(v importmodel.ImportContentVariable) models.ContentVariable {
	return models.ContentVariable{
		Name:                v.Name,
		Type:                models.Type(v.Type),
		SubContentVariables: importContentVariablesToContentVariables(v.SubContentVariables),
		CharacteristicId:    v.CharacteristicId,
		FunctionId:          v.FunctionId,
		AspectId:            v.AspectId,
	}
}
