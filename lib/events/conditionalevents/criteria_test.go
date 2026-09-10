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
	"reflect"
	"testing"

	eventmodel "github.com/SENERGY-Platform/event-deployment/lib/model"
	"github.com/SENERGY-Platform/event-worker/pkg/model"
	"github.com/SENERGY-Platform/models/go/models"
)

func TestSetSelectionCriteriaAspects(t *testing.T) {
	t.Run("carries an aspect list into the description", func(t *testing.T) {
		desc := model.EventDesc{}
		setSelectionCriteria(&desc, models.Selection{FilterCriteria: models.ProcessFilterCriteria{
			AspectIds: []string{"aid1", "aid2"},
		}})
		if !reflect.DeepEqual(desc.AspectIds, []string{"aid1", "aid2"}) {
			t.Error(desc.AspectIds)
		}
		if desc.AspectId != "" {
			t.Error(desc.AspectId)
		}
	})

	t.Run("keeps a deprecated single aspect id in the field the old workers read", func(t *testing.T) {
		desc := model.EventDesc{}
		setSelectionCriteria(&desc, models.Selection{FilterCriteria: models.ProcessFilterCriteria{
			AspectId: ptr("aid"),
		}})
		if desc.AspectId != "aid" {
			t.Error(desc.AspectId)
		}
		if len(desc.AspectIds) != 0 {
			t.Error(desc.AspectIds)
		}
	})

	t.Run("gives the same aspects for a deprecated id and a one element list", func(t *testing.T) {
		fromSingle := model.EventDesc{}
		setSelectionCriteria(&fromSingle, models.Selection{FilterCriteria: models.ProcessFilterCriteria{
			AspectId: ptr("aid"),
		}})
		fromList := model.EventDesc{}
		setSelectionCriteria(&fromList, models.Selection{FilterCriteria: models.ProcessFilterCriteria{
			AspectIds: []string{"aid"},
		}})
		if !reflect.DeepEqual(fromSingle.GetAspectIds(), fromList.GetAspectIds()) {
			t.Error(fromSingle.GetAspectIds(), fromList.GetAspectIds())
		}
	})

	t.Run("carries both spellings when the selection names both", func(t *testing.T) {
		desc := model.EventDesc{}
		setSelectionCriteria(&desc, models.Selection{FilterCriteria: models.ProcessFilterCriteria{
			AspectId:  ptr("aid1"),
			AspectIds: []string{"aid2"},
		}})
		if desc.AspectId != "aid1" {
			t.Error(desc.AspectId)
		}
		if !reflect.DeepEqual(desc.AspectIds, []string{"aid2"}) {
			t.Error(desc.AspectIds)
		}
		if !reflect.DeepEqual(desc.GetAspectIds(), []string{"aid2", "aid1"}) {
			t.Error(desc.GetAspectIds())
		}
	})

	t.Run("leaves the aspects unset when the selection names none", func(t *testing.T) {
		desc := model.EventDesc{}
		setSelectionCriteria(&desc, models.Selection{FilterCriteria: models.ProcessFilterCriteria{
			FunctionId: ptr("fid"),
		}})
		if desc.AspectId != "" || len(desc.AspectIds) != 0 {
			t.Error(desc.AspectId, desc.AspectIds)
		}
		if desc.FunctionId != "fid" {
			t.Error(desc.FunctionId)
		}
	})
}

func TestSelectableCriteria(t *testing.T) {
	t.Run("asks the device-repository with the aspect list", func(t *testing.T) {
		actual := selectableCriteria(model.EventDesc{FunctionId: "fid", AspectIds: []string{"aid1", "aid2"}})
		expected := []eventmodel.FilterCriteria{{FunctionId: "fid", AspectIds: []string{"aid1", "aid2"}}}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("\na=%#v\ne=%#v", actual, expected)
		}
	})

	t.Run("asks with the deprecated aspect id when the description carries only that", func(t *testing.T) {
		actual := selectableCriteria(model.EventDesc{FunctionId: "fid", AspectId: "aid"})
		expected := []eventmodel.FilterCriteria{{FunctionId: "fid", AspectId: "aid"}}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("\na=%#v\ne=%#v", actual, expected)
		}
	})
}

func ptr[T any](value T) *T {
	return &value
}
