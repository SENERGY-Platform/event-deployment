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

package devices

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/SENERGY-Platform/event-deployment/lib/model"
)

// TestGetDeviceTypeSelectablesCriteria covers the request as it leaves this service. The
// aspects only reach the device-repository if they are serialized into the criteria, and
// the deprecated aspect_id has to keep its place in the payload, because it is the alias a
// deployment may still be written with.
func TestGetDeviceTypeSelectablesCriteria(t *testing.T) {
	t.Run("sends the aspect list", func(t *testing.T) {
		actual := requestedCriteria(t, []model.FilterCriteria{{
			FunctionId: "fid",
			AspectIds:  []string{"aid1", "aid2"},
		}})
		expected := []map[string]interface{}{{
			"function_id":     "fid",
			"device_class_id": "",
			"aspect_id":       "",
			"aspect_ids":      []interface{}{"aid1", "aid2"},
		}}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("\na=%#v\ne=%#v", actual, expected)
		}
	})

	t.Run("omits the aspect list when only the deprecated aspect id is asked for", func(t *testing.T) {
		actual := requestedCriteria(t, []model.FilterCriteria{{
			FunctionId: "fid",
			AspectId:   "aid",
		}})
		expected := []map[string]interface{}{{
			"function_id":     "fid",
			"device_class_id": "",
			"aspect_id":       "aid",
		}}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("\na=%#v\ne=%#v", actual, expected)
		}
	})
}

// requestedCriteria returns the criteria as the device-repository receives them, read from
// the request body instead of from the struct, so that the json tags are part of the test.
func requestedCriteria(t *testing.T, criteria []model.FilterCriteria) []map[string]interface{} {
	t.Helper()
	var body []byte
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var err error
		body, err = io.ReadAll(request.Body)
		if err != nil {
			t.Error(err)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, err = writer.Write([]byte("[]"))
		if err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()

	devices := NewWithAuth(&config.ConfigStruct{DeviceRepositoryUrl: server.URL}, InternalAdminTokenAuth{})
	_, err, code := devices.GetDeviceTypeSelectables(criteria)
	if err != nil {
		t.Error(err, code)
		return nil
	}

	result := []map[string]interface{}{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		t.Error(err, string(body))
		return nil
	}
	return result
}
