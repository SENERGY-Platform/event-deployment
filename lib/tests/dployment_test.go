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
	"github.com/SENERGY-Platform/event-deployment/lib/model"
	"github.com/SENERGY-Platform/event-deployment/lib/tests/mocks"
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"runtime/debug"
	"sort"
	"sync"
	"testing"
)

const RESOURCES_DIR = "resources/"
const DEPLOYMENT_EXAMPLES_DIR = RESOURCES_DIR + "deployment_examples/"

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

func testDeployment(t *testing.T, testcase string) {
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
	conf.AuthEndpoint = "mocked"
	conf.AuthClientSecret = "mocked"
	conf.AuthClientId = "mocked"
	conf.PermSearchUrl = "mocked"

	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	defer wg.Wait()
	defer cancel()

	err = mocks.MockAuthServer(conf, ctx)
	if err != nil {
		t.Error(err)
		return
	}

	deploymentCmd, err := ioutil.ReadFile(DEPLOYMENT_EXAMPLES_DIR + testcase + "/deploymentcommand.json")
	if err != nil {
		t.Error(err)
		return
	}

	devicesMock := mocks.DevicesMock{
		GetDeviceInfosOfGroupValues: map[string][]model.Device{},
	}

	groupdevicesFilePath := DEPLOYMENT_EXAMPLES_DIR + testcase + "/groupdevices.json"
	if fileExists(groupdevicesFilePath) {
		groupdevicesJson, err := ioutil.ReadFile(groupdevicesFilePath)
		if err != nil {
			t.Error(err)
			return
		}
		err = json.Unmarshal(groupdevicesJson, &devicesMock.GetDeviceInfosOfGroupValues)
		if err != nil {
			t.Error(err)
			return
		}
	}

	deviceTypeSelectablesPath := DEPLOYMENT_EXAMPLES_DIR + testcase + "/devicetypeselectables.json"
	if fileExists(deviceTypeSelectablesPath) {
		deviceTypeSelectablesJson, err := ioutil.ReadFile(deviceTypeSelectablesPath)
		if err != nil {
			t.Error(err)
			return
		}
		err = json.Unmarshal(deviceTypeSelectablesJson, &devicesMock.GetDeviceTypeSelectablesValues)
		if err != nil {
			t.Error(err)
			return
		}
	}

	closeTestPipelineRepoApi := func() {}
	conf.PipelineRepoUrl, closeTestPipelineRepoApi, err = createTestPipelineRepoApi(DEPLOYMENT_EXAMPLES_DIR + testcase + "/knownpipelines.json")
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
	conf.FlowEngineUrl, closeTestFlowEngineApi, err = createTestFlowEngineApi(t, DEPLOYMENT_EXAMPLES_DIR+testcase)
	if err != nil {
		t.Error(err)
		return
	}
	defer closeTestFlowEngineApi()

	a, err := analytics.Factory.New(ctx, conf)
	if err != nil {
		t.Error(err)
		return
	}

	event, err := events.Factory.New(ctx, conf, a, &devicesMock, &mocks.ImportsMock{})
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

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
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
	return files["deploymentcommand.json"] && files["pipelinerequests.json"]
}

func createTestFlowEngineApi(t *testing.T, fullTestCasePath string) (endpointUrl string, close func(), err error) {
	expectedRequestJson, err := ioutil.ReadFile(fullTestCasePath + "/pipelinerequests.json")
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
		if (r.Method == "POST" || r.Method == "PUT") && r.URL.Path == "/pipeline" {
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
			for i, node := range actualRequest.Nodes {
				sort.Slice(node.Inputs, func(i, j int) bool {
					return node.Inputs[i].TopicName < node.Inputs[j].TopicName
				})
				actualRequest.Nodes[i] = node
			}
			for i, node := range expectedRequests[count].Nodes {
				sort.Slice(node.Inputs, func(i, j int) bool {
					return node.Inputs[i].TopicName < node.Inputs[j].TopicName
				})
				expectedRequests[count].Nodes[i] = node
			}
			if !reflect.DeepEqual(expectedRequests[count], actualRequest) {
				expectedJson, _ := json.Marshal(expectedRequests[count])
				actualJson, _ := json.Marshal(actualRequest)
				t.Error(string(expectedJson), "\n\n", string(actualJson))
				return
			}
			count = count + 1
			json.NewEncoder(w).Encode(analytics.Pipeline{
				Id:   uuid.NewV4(),
				Name: "test-result",
			})
		} else {
			t.Error("unknown request endpoint", r.Method, r.URL.Path)
		}
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

func createTestPipelineRepoApi(pipelinesFilePath string) (endpointUrl string, close func(), err error) {
	pipelines := []interface{}{}
	if fileExists(pipelinesFilePath) {
		groupdevicesJson, err := ioutil.ReadFile(pipelinesFilePath)
		if err != nil {
			return endpointUrl, func() {}, err
		}
		err = json.Unmarshal(groupdevicesJson, &pipelines)
		if err != nil {
			return endpointUrl, func() {}, err
		}
	}
	endpointMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(pipelines)
	}))
	endpointUrl = endpointMock.URL
	close = func() {
		endpointMock.Close()
	}
	return
}
