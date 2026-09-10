/*
 * Copyright 2026 InfAI (CC SES)
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
	eventmodel "github.com/SENERGY-Platform/event-deployment/lib/model"
	"github.com/SENERGY-Platform/event-worker/pkg/model"
	"github.com/SENERGY-Platform/models/go/models"
)

// setSelectionCriteria copies the criteria of an event selection into the description the
// worker evaluates later. Both aspect spellings are passed on as they were selected:
// AspectId is deprecated and an alias for an AspectIds list with a single element, and a
// description that carries it alone stays readable for a worker that only knows the single
// field. The worker folds the two on read.
func setSelectionCriteria(desc *model.EventDesc, selection models.Selection) {
	if selection.FilterCriteria.CharacteristicId != nil {
		desc.CharacteristicId = *selection.FilterCriteria.CharacteristicId
	}
	if selection.FilterCriteria.FunctionId != nil {
		desc.FunctionId = *selection.FilterCriteria.FunctionId
	}
	if selection.FilterCriteria.AspectId != nil {
		desc.AspectId = *selection.FilterCriteria.AspectId
	}
	desc.AspectIds = selection.FilterCriteria.AspectIds
	if selection.SelectedPath != nil {
		desc.Path = selection.SelectedPath.Path
	}
}

// selectableCriteria asks the device-repository for the services that produce what the
// event description needs. The deprecated AspectId is sent next to the list, because the
// device-repository normalizes it into AspectIds at its own boundary and a description may
// carry either spelling.
func selectableCriteria(desc model.EventDesc) []eventmodel.FilterCriteria {
	return []eventmodel.FilterCriteria{{
		FunctionId: desc.FunctionId,
		AspectId:   desc.AspectId,
		AspectIds:  desc.AspectIds,
	}}
}
