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

package analytics

import (
	"context"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/SENERGY-Platform/event-deployment/lib/interfaces"
	deploymentmodel "github.com/SENERGY-Platform/process-deployment/lib/model"
	"github.com/SENERGY-Platform/process-deployment/lib/model/devicemodel"
	uuid "github.com/satori/go.uuid"
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestAnalytics(t *testing.T) {
	config, err := config.LoadConfig("test.config.json")
	if err != nil {
		t.Error(err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer time.Sleep(1 * time.Second) //wait for goroutines with context
	defer cancel()

	config, err = createAnalyticsProxyServer(ctx, config)
	if err != nil {
		t.Error(err)
		return
	}

	analytics, err := Factory.New(ctx, config)
	if err != nil {
		t.Error(err)
		return
	}

	var pipelineId string
	deploymentId := uuid.NewV4().String()
	eventId := uuid.NewV4().String()

	t.Run("deploy", func(t *testing.T) {
		pipelineId = testDeploy(t, analytics, deploymentId, deploymentmodel.MsgEvent{
			Label: "test event //TODO: delete",
			Device: devicemodel.Device{
				Id: "d1",
			},
			Service: devicemodel.Service{
				Id: "s1",
			},
			Path:      "path/to/value",
			Value:     "42",
			Operation: "math-interval",
			EventId:   eventId,
		})
	})

	t.Run("readByDeploymentId", func(t *testing.T) {
		testReadByDeploymentId(t, analytics, deploymentId, []string{pipelineId})
	})

	t.Run("readByEventIdId", func(t *testing.T) {
		testReadByEventId(t, analytics, eventId, pipelineId, true)
	})

	t.Run("remove", func(t *testing.T) {
		testRemove(t, analytics, pipelineId)
	})

	t.Run("readByEventIdId", func(t *testing.T) {
		testReadByEventId(t, analytics, eventId, "", false)
	})

	t.Run("readByDeploymentId", func(t *testing.T) {
		testReadByDeploymentId(t, analytics, deploymentId, []string{})
	})
}

func testReadByDeploymentId(t *testing.T, analytics interfaces.Analytics, deploymentId string, expectedPipelineIds []string) {
	pipelineIds, err := analytics.GetPipelinesByDeploymentId("", deploymentId)
	if err != nil {
		t.Fatal(err)
	}

	sort.Strings(expectedPipelineIds)
	sort.Strings(pipelineIds)

	if !reflect.DeepEqual(pipelineIds, expectedPipelineIds) {
		t.Fatal(pipelineIds, expectedPipelineIds)
	}
}

func testReadByEventId(t *testing.T, analytics interfaces.Analytics, event string, expectedPipelineId string, expectedExists bool) {
	pipelineId, exists, err := analytics.GetPipelineByEventId("", event)
	if err != nil {
		t.Fatal(err)
	}
	if exists != expectedExists {
		t.Fatal(exists, expectedExists)
	}
	if pipelineId != expectedPipelineId {
		t.Fatal(pipelineId, expectedPipelineId)
	}
}

func testDeploy(t *testing.T, analytics interfaces.Analytics, deploymentId string, event deploymentmodel.MsgEvent) (pipelineId string) {
	var err error
	pipelineId, err = analytics.Deploy("", deploymentId, event)
	if err != nil {
		t.Fatal(err)
	}
	return pipelineId
}

func testRemove(t *testing.T, analytics interfaces.Analytics, pipelineId string) {
	err := analytics.Remove("", pipelineId)
	if err != nil {
		t.Fatal(err)
	}
}
