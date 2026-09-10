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
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/SENERGY-Platform/event-worker/pkg/model"
	"github.com/SENERGY-Platform/models/go/models"
)

func (this *Transformer) transformEventForImport(owner string, deployentId string, event *models.ConditionalEvent) (result []model.EventDesc, err error) {
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

	setSelectionCriteria(&desc, event.Selection)

	importInstance, err, code := this.imports.GetImportInstance(owner, desc.ImportId)
	if err != nil {
		if code == http.StatusInternalServerError {
			return []model.EventDesc{}, err
		} else {
			slog.Default().Error("ERROR", "code", code, "error", err)
			debug.PrintStack()
			return []model.EventDesc{}, nil //ignore bad request errors
		}
	}
	importType, err, code := this.imports.GetImportType(owner, importInstance.ImportTypeId)
	if err != nil {
		if code == http.StatusInternalServerError {
			return []model.EventDesc{}, err
		} else {
			slog.Default().Error("ERROR", "code", code, "error", err)
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

func importVariablesToContents(variables []models.ImportContentVariable) (result []models.Content) {
	for _, v := range variables {
		result = append(result, importVariableToContent(v))
	}
	return
}

func importVariableToContent(variable models.ImportContentVariable) (result models.Content) {
	result.ContentVariable = importContentVariableToContentVariable(variable)
	return
}

func importContentVariablesToContentVariables(variables []models.ImportContentVariable) (result []models.ContentVariable) {
	for _, v := range variables {
		result = append(result, importContentVariableToContentVariable(v))
	}
	return
}

func importContentVariableToContentVariable(v models.ImportContentVariable) models.ContentVariable {
	return models.ContentVariable{
		Name:                v.Name,
		Type:                models.Type(v.Type),
		SubContentVariables: importContentVariablesToContentVariables(v.SubContentVariables),
		CharacteristicId:    v.CharacteristicId,
		FunctionId:          v.FunctionId,
		AspectId:            v.AspectId,
	}
}
