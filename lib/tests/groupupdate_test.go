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
	"github.com/SENERGY-Platform/event-deployment/lib/metrics"
	"github.com/SENERGY-Platform/event-deployment/lib/model"
	"github.com/SENERGY-Platform/event-deployment/lib/tests/mocks"
	"io/ioutil"
	"runtime/debug"
	"sync"
	"testing"
)

const GROUPUPDATE_EXAMPLES_DIR = RESOURCES_DIR + "groupupdate_examples/"

func TestGroupUpdate(t *testing.T) {
	infos, err := ioutil.ReadDir(GROUPUPDATE_EXAMPLES_DIR)
	if err != nil {
		t.Error(err)
		return
	}
	for _, info := range infos {
		name := info.Name()
		if info.IsDir() && isValidForGroupUpdateTest(GROUPUPDATE_EXAMPLES_DIR+name) {
			t.Run(name, func(t *testing.T) {
				testGroupUpdate(t, name)
			})
		}
	}
}

func testGroupUpdate(t *testing.T, testcase string) {
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
	conf.ImportPathPrefix = ""
	conf.DevicePathPrefix = ""
	conf.GroupPathPrefix = ""

	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	defer wg.Wait()
	defer cancel()

	err = mocks.MockAuthServer(conf, ctx)
	if err != nil {
		t.Error(err)
		return
	}

	updateCmd, err := ioutil.ReadFile(GROUPUPDATE_EXAMPLES_DIR + testcase + "/groupcommand.json")
	if err != nil {
		t.Error(err)
		return
	}

	devicesMock := mocks.DevicesMock{
		GetDeviceInfosOfGroupValues: map[string][]model.Device{},
	}

	groupdevicesFilePath := GROUPUPDATE_EXAMPLES_DIR + testcase + "/groupdevices.json"
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

	deviceTypeSelectablesPath := GROUPUPDATE_EXAMPLES_DIR + testcase + "/devicetypeselectables.json"
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

	functionsPath := GROUPUPDATE_EXAMPLES_DIR + testcase + "/functions.json"
	if fileExists(functionsPath) {
		functionsJson, err := ioutil.ReadFile(functionsPath)
		if err != nil {
			t.Error(err)
			return
		}
		err = json.Unmarshal(functionsJson, &devicesMock.Functions)
		if err != nil {
			t.Error(err)
			return
		}
	}

	conceptsPath := GROUPUPDATE_EXAMPLES_DIR + testcase + "/concepts.json"
	if fileExists(conceptsPath) {
		conceptsJson, err := ioutil.ReadFile(conceptsPath)
		if err != nil {
			t.Error(err)
			return
		}
		err = json.Unmarshal(conceptsJson, &devicesMock.Concepts)
		if err != nil {
			t.Error(err)
			return
		}
	}

	closeTestPipelineRepoApi := func() {}
	conf.PipelineRepoUrl, closeTestPipelineRepoApi, err = createTestPipelineRepoApi(GROUPUPDATE_EXAMPLES_DIR + testcase + "/knownpipelines.json")
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
	conf.FlowEngineUrl, closeTestFlowEngineApi, err = createTestFlowEngineApi(t, GROUPUPDATE_EXAMPLES_DIR+testcase)
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

	event, err := events.Factory.New(ctx, conf, a, &devicesMock, &mocks.ImportsMock{}, nil, metrics.New())
	if err != nil {
		t.Error(err)
		return
	}

	err = event.HandleDeviceGroupUpdate(updateCmd)
	if err != nil {
		t.Error(err)
		return
	}
}

func isValidForGroupUpdateTest(dir string) bool {
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
	return files["groupcommand.json"] && files["pipelinerequests.json"]
}
