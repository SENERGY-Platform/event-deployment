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
	"github.com/SENERGY-Platform/event-deployment/lib/tests/docker"
	"github.com/SENERGY-Platform/event-deployment/lib/tests/mocks"
	eventworkermodel "github.com/SENERGY-Platform/event-worker/pkg/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"os"
	"reflect"
	"runtime/debug"
	"sort"
	"sync"
	"testing"
)

const CONDITIONAL_EVENT_EXAMPLES_DIR = RESOURCES_DIR + "conditional_events/"

func TestConditionalEvent(t *testing.T) {
	infos, err := os.ReadDir(CONDITIONAL_EVENT_EXAMPLES_DIR)
	if err != nil {
		t.Error(err)
		return
	}
	for _, info := range infos {
		name := info.Name()
		if info.IsDir() && isValidForConditionalEventTest(CONDITIONAL_EVENT_EXAMPLES_DIR+name) {
			t.Run(name, func(t *testing.T) {
				testConditionalEvent(t, name)
			})
		}
	}
}

func testConditionalEvent(t *testing.T, testcase string) {
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

	_, mongoIp, err := docker.Mongo(ctx, &wg)
	if err != nil {
		t.Error(err)
		return
	}

	conf.ConditionalEventRepoMongoUrl = "mongodb://" + mongoIp + ":27017"

	err = mocks.MockAuthServer(conf, ctx)
	if err != nil {
		t.Error(err)
		return
	}

	deploymentCmd, err := os.ReadFile(CONDITIONAL_EVENT_EXAMPLES_DIR + testcase + "/deploymentcommand.json")
	if err != nil {
		t.Error(err)
		return
	}

	expectedEventDescriptionsJson, err := os.ReadFile(CONDITIONAL_EVENT_EXAMPLES_DIR + testcase + "/expected_event_descriptions.json")
	if err != nil {
		t.Error(err)
		return
	}
	expectedEventDescriptions := []eventworkermodel.EventDesc{}
	err = json.Unmarshal(expectedEventDescriptionsJson, &expectedEventDescriptions)
	if err != nil {
		t.Error(err)
		return
	}

	devicesMock := mocks.DevicesMock{
		GetDeviceInfosOfGroupValues: map[string][]model.Device{},
	}

	groupdevicesFilePath := CONDITIONAL_EVENT_EXAMPLES_DIR + testcase + "/groupdevices.json"
	if fileExists(groupdevicesFilePath) {
		groupdevicesJson, err := os.ReadFile(groupdevicesFilePath)
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

	deviceTypeSelectablesPath := CONDITIONAL_EVENT_EXAMPLES_DIR + testcase + "/devicetypeselectables.json"
	if fileExists(deviceTypeSelectablesPath) {
		deviceTypeSelectablesJson, err := os.ReadFile(deviceTypeSelectablesPath)
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

	functionsPath := CONDITIONAL_EVENT_EXAMPLES_DIR + testcase + "/functions.json"
	if fileExists(functionsPath) {
		functionsJson, err := os.ReadFile(functionsPath)
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

	conceptsPath := CONDITIONAL_EVENT_EXAMPLES_DIR + testcase + "/concepts.json"
	if fileExists(conceptsPath) {
		conceptsJson, err := os.ReadFile(conceptsPath)
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
	conf.PipelineRepoUrl, closeTestPipelineRepoApi, err = createTestPipelineRepoApi(DEPLOYMENT_EXAMPLES_DIR + testcase + "/knownpipelines.json")
	if err != nil {
		t.Error(err)
		return
	}
	defer closeTestPipelineRepoApi()

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

	err = event.HandleCommand(deploymentCmd)
	if err != nil {
		t.Error(err)
		return
	}

	reg := bson.NewRegistryBuilder().RegisterTypeMapEntry(bsontype.EmbeddedDocument, reflect.TypeOf(bson.M{})).Build()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(conf.ConditionalEventRepoMongoUrl), options.Client().SetRegistry(reg))
	if err != nil {
		t.Error(err)
		return
	}

	cursor, err := client.Database(conf.ConditionalEventRepoMongoTable).Collection(conf.ConditionalEventRepoMongoDescCollection).Find(ctx, bson.M{})
	if err != nil {
		t.Error(err)
		return
	}
	actualEventDescriptions, err, _ := readCursorResult[eventworkermodel.EventDesc](ctx, cursor)
	if err != nil {
		t.Error(err)
		return
	}

	sortEventDesc := func(list []eventworkermodel.EventDesc) []eventworkermodel.EventDesc {
		sort.Slice(list, func(i, j int) bool {
			return list[i].DeploymentId < list[j].DeploymentId
		})
		sort.Slice(list, func(i, j int) bool {
			return list[i].EventId < list[j].EventId
		})
		sort.Slice(list, func(i, j int) bool {
			return list[i].DeviceId < list[j].DeviceId
		})
		sort.Slice(list, func(i, j int) bool {
			return list[i].ServiceId < list[j].ServiceId
		})
		return list
	}
	expectedEventDescriptions = sortEventDesc(expectedEventDescriptions)
	actualEventDescriptions = sortEventDesc(actualEventDescriptions)

	if !reflect.DeepEqual(expectedEventDescriptions, actualEventDescriptions) {
		t.Errorf("\n%#v\n%#v", expectedEventDescriptions, actualEventDescriptions)
		temp, _ := json.Marshal(actualEventDescriptions)
		t.Log(string(temp))
		temp, _ = json.Marshal(expectedEventDescriptions)
		t.Log(string(temp))
	}

}

func isValidForConditionalEventTest(dir string) bool {
	infos, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	files := map[string]bool{}
	for _, info := range infos {
		if !info.IsDir() {
			files[info.Name()] = true
		}
	}
	return files["deploymentcommand.json"] && files["expected_event_descriptions.json"]
}

func readCursorResult[T any](ctx context.Context, cursor *mongo.Cursor) (result []T, err error, code int) {
	result = []T{}
	for cursor.Next(ctx) {
		element := new(T)
		err = cursor.Decode(element)
		if err != nil {
			return result, err, http.StatusInternalServerError
		}
		result = append(result, *element)
	}
	err = cursor.Err()
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return result, nil, http.StatusOK
}
