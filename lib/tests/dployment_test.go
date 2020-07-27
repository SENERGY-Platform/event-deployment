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

package tests

import (
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/event-deployment/lib/analytics"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/SENERGY-Platform/event-deployment/lib/events"
	"github.com/SENERGY-Platform/event-deployment/lib/iot/model"
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"runtime/debug"
	"strings"
	"testing"
)

const RESOURCES_DIR = "resources/"
const DEPLOYMENT_EXAMPLES_DIR = RESOURCES_DIR + "deployment_examples/"
const IOT_DESC_DIR = RESOURCES_DIR + "/iot/"

func TestDeployment(t *testing.T) {
	infos, err := ioutil.ReadDir(DEPLOYMENT_EXAMPLES_DIR)
	if err != nil {
		t.Error(err)
		return
	}
	for _, info := range infos {
		name := info.Name()
		if info.IsDir() && isValidForDeploymentTest(DEPLOYMENT_EXAMPLES_DIR+name) {
			t.Run(name, func(t *testing.T) {
				testDeployment(t, name)
			})
		}
	}
}

func testDeployment(t *testing.T, name string) {
	t.Parallel()
	defer func() {
		if r := recover(); r != nil {
			t.Error(r, string(debug.Stack()))
		}
	}()
	conf, err := config.LoadConfig("../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	deploymentCmd, err := ioutil.ReadFile(DEPLOYMENT_EXAMPLES_DIR + name + "/deploymentcommand.json")
	if err != nil {
		t.Error(err)
		return
	}

	closeTestIotApi := func() {}
	conf.DeviceRepoUrl, closeTestIotApi, err = createTestIotApi(t, name)
	if err != nil {
		t.Error(err)
		return
	}
	defer closeTestIotApi()

	closeTestPipelineRepoApi := func() {}
	conf.PipelineRepoUrl, closeTestPipelineRepoApi, err = createTestPipelineRepoApi()
	if err != nil {
		t.Error(err)
		return
	}
	defer closeTestPipelineRepoApi()

	closeTestFlowParserApi := func() {}
	conf.FlowParserUrl, closeTestFlowParserApi, err = createTestFlowParserApi()
	if err != nil {
		t.Error(err)
		return
	}
	defer closeTestFlowParserApi()

	closeTestFlowEngineApi := func() {}
	conf.FlowEngineUrl, closeTestFlowEngineApi, err = createTestFlowEngineApi(t, name)
	if err != nil {
		t.Error(err)
		return
	}
	defer closeTestFlowEngineApi()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	a, err := analytics.Factory.New(ctx, conf)
	if err != nil {
		t.Error(err)
		return
	}
	event, err := events.Factory.New(ctx, conf, a)
	if err != nil {
		t.Error(err)
		return
	}

	err = event.HandleCommand(deploymentCmd)
	if err != nil {
		t.Error(err)
		return
	}
}

func isValidForDeploymentTest(dir string) bool {
	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	files := map[string]bool{}
	for _, info := range infos {
		if !info.IsDir() {
			files[info.Name()] = true
		}
	}
	return files["deploymentcommand.json"] && files["pipelinerequests.json"] && files["devices.json"]
}

func createTestIotApi(t *testing.T, example string) (deviceRepoUrl string, close func(), err error) {
	devicesDesc, err := ioutil.ReadFile(DEPLOYMENT_EXAMPLES_DIR + example + "/devices.json")
	if err != nil {
		t.Error(err)
		return deviceRepoUrl, close, err
	}

	deviceTypesDesc, err := ioutil.ReadFile(IOT_DESC_DIR + "devicetypes.json")
	if err != nil {
		t.Error(err)
		return deviceRepoUrl, close, err
	}

	deviceTypes := []model.DeviceType{}
	err = json.Unmarshal(deviceTypesDesc, &deviceTypes)
	if err != nil {
		t.Error(err)
		return deviceRepoUrl, close, err
	}

	testDevices := []model.Device{}
	err = json.Unmarshal(devicesDesc, &testDevices)
	if err != nil {
		t.Error(err)
		return deviceRepoUrl, close, err
	}

	endpointMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		devicePrefix := "/devices/"
		if strings.HasPrefix(r.URL.Path, devicePrefix) {
			for _, device := range testDevices {
				if device.Id == r.URL.Path[len(devicePrefix):] {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					json.NewEncoder(w).Encode(device)
					return
				}
			}
		}

		deviceTypePrefix := "/device-types/"
		if strings.HasPrefix(r.URL.Path, deviceTypePrefix) {
			for _, dt := range deviceTypes {
				if dt.Id == r.URL.Path[len(deviceTypePrefix):] {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					json.NewEncoder(w).Encode(dt)
					return
				}
			}
		}
	}))

	deviceRepoUrl = endpointMock.URL
	close = func() {
		endpointMock.Close()
	}
	return
}

func createTestFlowEngineApi(t *testing.T, example string) (endpointUrl string, close func(), err error) {
	expectedRequestJson, err := ioutil.ReadFile(DEPLOYMENT_EXAMPLES_DIR + example + "/pipelinerequests.json")
	if err != nil {
		return endpointUrl, close, err
	}

	expectedRequests := []analytics.PipelineRequest{}
	err = json.Unmarshal(expectedRequestJson, &expectedRequests)
	if err != nil {
		return endpointUrl, close, err
	}

	count := 0
	endpointMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if count >= len(expectedRequests) {
			t.Error("to many requests to flow engine \n\n", len(expectedRequests))
			return
		}
		actualRequest := analytics.PipelineRequest{}
		err = json.NewDecoder(r.Body).Decode(&actualRequest)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(expectedRequests[count], actualRequest) {
			t.Error(expectedRequests[count], "\n\n", actualRequest)
			return
		}
		count = count + 1
		json.NewEncoder(w).Encode(analytics.Pipeline{
			Id:   uuid.NewV4(),
			Name: "test-result",
		})
	}))

	endpointUrl = endpointMock.URL
	close = func() {
		endpointMock.Close()
		if count < len(expectedRequests) {
			t.Error("missing requests to flow engine", count, len(expectedRequests))
		}
	}
	return
}

func createTestFlowParserApi() (endpointUrl string, close func(), err error) {
	endpointMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]analytics.FlowModelCell{{Id: "test-flow-cell-id"}})
	}))
	endpointUrl = endpointMock.URL
	close = func() {
		endpointMock.Close()
	}
	return
}

func createTestPipelineRepoApi() (endpointUrl string, close func(), err error) {
	endpointMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]interface{}{})
	}))
	endpointUrl = endpointMock.URL
	close = func() {
		endpointMock.Close()
	}
	return
}
