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
	"github.com/SENERGY-Platform/event-deployment/lib"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/SENERGY-Platform/event-deployment/lib/interfaces"
	"github.com/SENERGY-Platform/event-deployment/lib/kafka"
	deploymentmodel "github.com/SENERGY-Platform/process-deployment/lib/model"
	"github.com/SENERGY-Platform/process-deployment/lib/model/devicemodel"
	"github.com/ory/dockertest"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"net/url"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestLib(t *testing.T) {
	config, err := config.LoadConfig("test.config.json")
	if err != nil {
		debug.PrintStack()
		t.Error(err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer time.Sleep(10 * time.Second) //wait for goroutines with context
	defer cancel()

	config, err = createAnalyticsProxyServer(ctx, config)
	if err != nil {
		debug.PrintStack()
		t.Error(err)
		return
	}

	apiPort, err := getFreePort()
	if err != nil {
		debug.PrintStack()
		t.Error(err)
		return
	}
	config.ApiPort = strconv.Itoa(apiPort)

	pool, err := dockertest.NewPool("")
	if err != nil {
		debug.PrintStack()
		t.Error(err)
		return
	}

	_, zk, err := Zookeeper(pool, ctx)
	if err != nil {
		debug.PrintStack()
		t.Error(err)
		return
	}
	config.ZookeeperUrl = zk + ":2181"

	err = Kafka(pool, ctx, config.ZookeeperUrl)
	if err != nil {
		debug.PrintStack()
		t.Error(err)
		return
	}

	err = lib.StartDefault(ctx, config)
	if err != nil {
		debug.PrintStack()
		t.Error(err)
		return
	}

	producer, err := kafka.NewProducer(ctx, config, config.DeploymentTopic)
	if err != nil {
		debug.PrintStack()
		t.Error(err)
		return
	}

	deploymentId := uuid.NewV4().String()
	eventId := uuid.NewV4().String()

	t.Run("deploy", func(t *testing.T) {
		testDeployToKafka(t, producer, deploymentId, deploymentmodel.MsgEvent{
			Label: "test event //TODO: delete",
			Device: devicemodel.Device{
				Id: "d1",
			},
			Service: devicemodel.Service{
				Id: "s1",
			},
			Path:      "path/to/value",
			Value:     "(*,*)",
			Operation: "math-interval",
			EventId:   eventId,
		})
	})

	time.Sleep(10 * time.Second)

	t.Run("eventExists", func(t *testing.T) {
		apiCheckEvent(t, config, eventId, true)
	})

	t.Run("eventStates", func(t *testing.T) {
		apiEventStates(t, config, []string{eventId}, map[string]bool{eventId: true})
	})
	t.Run("eventStates", func(t *testing.T) {
		apiEventStates(t, config, []string{}, map[string]bool{})
	})
	t.Run("remove", func(t *testing.T) {
		testRemoveByKafka(t, producer, deploymentId)
	})

	time.Sleep(10 * time.Second)

	t.Run("eventExistsNot", func(t *testing.T) {
		apiCheckEvent(t, config, eventId, false)
	})

	t.Run("eventStates", func(t *testing.T) {
		apiEventStates(t, config, []string{eventId}, map[string]bool{eventId: false})
	})
}

func apiEventStates(t *testing.T, config config.Config, eventIds []string, expected map[string]bool) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest(
		"GET",
		"http://localhost:"+config.ApiPort+"/event-states?ids="+url.QueryEscape(strings.Join(eventIds, ",")),
		nil,
	)
	if err != nil {
		debug.PrintStack()
		t.Error(err)
		return
	}
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICIzaUtabW9aUHpsMmRtQnBJdS1vSkY4ZVVUZHh4OUFIckVOcG5CcHM5SjYwIn0.eyJqdGkiOiI0ZjY4MThmZi0wYTRiLTQ4YjYtYTdkYi00NTk1ZjY5Y2RmYWEiLCJleHAiOjE1ODA0MDQwNjMsIm5iZiI6MCwiaWF0IjoxNTgwNDAwNDYzLCJpc3MiOiJodHRwOi8vZmdzZWl0c3JhbmNoZXIud2lmYS5pbnRlcm4udW5pLWxlaXB6aWcuZGU6ODA4Ny9hdXRoL3JlYWxtcy9tYXN0ZXIiLCJhdWQiOiJhY2NvdW50Iiwic3ViIjoiNjIxOWRjNDItYjhkMC00YjQyLTg1MWEtMWM1OTU2MTQ5OTQ0IiwidHlwIjoiQmVhcmVyIiwiYXpwIjoiZnJvbnRlbmQiLCJub25jZSI6IjMzNzEyYzA2LTA1YTctNDhkNy04ZjZhLTg5OGU0N2Q5OTM0MSIsImF1dGhfdGltZSI6MTU4MDM5NjA5NSwic2Vzc2lvbl9zdGF0ZSI6IjcyNGMyMjIzLWNkMjQtNDQ5MC05OTM4LWIzNDc1NjBkYmI0ZSIsImFjciI6IjAiLCJhbGxvd2VkLW9yaWdpbnMiOlsiKiJdLCJyZWFsbV9hY2Nlc3MiOnsicm9sZXMiOlsib2ZmbGluZV9hY2Nlc3MiLCJkZXZlbG9wZXIiLCJ1bWFfYXV0aG9yaXphdGlvbiIsInVzZXIiXX0sInJlc291cmNlX2FjY2VzcyI6eyJhY2NvdW50Ijp7InJvbGVzIjpbIm1hbmFnZS1hY2NvdW50IiwibWFuYWdlLWFjY291bnQtbGlua3MiLCJ2aWV3LXByb2ZpbGUiXX19LCJzY29wZSI6Im9wZW5pZCIsInJvbGVzIjpbIm9mZmxpbmVfYWNjZXNzIiwiZGV2ZWxvcGVyIiwidW1hX2F1dGhvcml6YXRpb24iLCJ1c2VyIl0sIm5hbWUiOiJEZW1vIFVzZXIiLCJwcmVmZXJyZWRfdXNlcm5hbWUiOiJkZW1vLnVzZXIiLCJnaXZlbl9uYW1lIjoiRGVtbyIsImZhbWlseV9uYW1lIjoiVXNlciIsImVtYWlsIjoiIn0.ZDgUrqwoW2NxL2kDias7DyU4AAv8p0m2y69l8s1aE5LbGkMMzljoaERJiMpFlQLBrM3hL57XI1FpJcVgKb4VvAjxZqBUqSFIiyB38GvmvXs5xIrwUBLMYu8uicnqYKW0hZi9Gfr7Fiwzsk_t9KA-YaxeGEdOJ5K6VQV-eV8Prs5bXkyacW_Lu5ZzTAbHSZllNqRxppsjD6mBOPficeKHFhoAw-_EsT2d4DmD2DyYis32olLYaOWGBGI5y5X7yf92S_mtzy1Amy_yDqVWR2WXKfG7uPj6Tfw8yczElVRxV1OXKblzrR1E28Mui5Ll6eL5n0VOuMs3NLNW62rDFMRZ7Q")
	resp, err := client.Do(req)
	if err != nil {
		debug.PrintStack()
		t.Error(err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		debug.PrintStack()
		t.Error(resp.StatusCode)
		return
	}
	result := map[string]bool{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if resp.StatusCode != http.StatusOK {
		debug.PrintStack()
		t.Error(err)
		return
	}
	if !reflect.DeepEqual(expected, result) {
		t.Error(expected, result)
		return
	}
}

func testRemoveByKafka(t *testing.T, producer interfaces.Producer, deploymentId string) {
	cmd := deploymentmodel.DeploymentCommand{Id: deploymentId, Command: "DELETE"}
	msg, err := json.Marshal(cmd)
	if err != nil {
		debug.PrintStack()
		t.Error(err)
		return
	}
	err = producer.Produce("test", msg)
	if err != nil {
		debug.PrintStack()
		t.Error(err)
		return
	}
}

func apiCheckEvent(t *testing.T, config config.Config, eventId string, expectetToExist bool) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest(
		"HEAD",
		"http://localhost:"+config.ApiPort+"/events/"+url.PathEscape(eventId),
		nil,
	)
	if err != nil {
		debug.PrintStack()
		t.Error(err)
		return
	}
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICIzaUtabW9aUHpsMmRtQnBJdS1vSkY4ZVVUZHh4OUFIckVOcG5CcHM5SjYwIn0.eyJqdGkiOiI0ZjY4MThmZi0wYTRiLTQ4YjYtYTdkYi00NTk1ZjY5Y2RmYWEiLCJleHAiOjE1ODA0MDQwNjMsIm5iZiI6MCwiaWF0IjoxNTgwNDAwNDYzLCJpc3MiOiJodHRwOi8vZmdzZWl0c3JhbmNoZXIud2lmYS5pbnRlcm4udW5pLWxlaXB6aWcuZGU6ODA4Ny9hdXRoL3JlYWxtcy9tYXN0ZXIiLCJhdWQiOiJhY2NvdW50Iiwic3ViIjoiNjIxOWRjNDItYjhkMC00YjQyLTg1MWEtMWM1OTU2MTQ5OTQ0IiwidHlwIjoiQmVhcmVyIiwiYXpwIjoiZnJvbnRlbmQiLCJub25jZSI6IjMzNzEyYzA2LTA1YTctNDhkNy04ZjZhLTg5OGU0N2Q5OTM0MSIsImF1dGhfdGltZSI6MTU4MDM5NjA5NSwic2Vzc2lvbl9zdGF0ZSI6IjcyNGMyMjIzLWNkMjQtNDQ5MC05OTM4LWIzNDc1NjBkYmI0ZSIsImFjciI6IjAiLCJhbGxvd2VkLW9yaWdpbnMiOlsiKiJdLCJyZWFsbV9hY2Nlc3MiOnsicm9sZXMiOlsib2ZmbGluZV9hY2Nlc3MiLCJkZXZlbG9wZXIiLCJ1bWFfYXV0aG9yaXphdGlvbiIsInVzZXIiXX0sInJlc291cmNlX2FjY2VzcyI6eyJhY2NvdW50Ijp7InJvbGVzIjpbIm1hbmFnZS1hY2NvdW50IiwibWFuYWdlLWFjY291bnQtbGlua3MiLCJ2aWV3LXByb2ZpbGUiXX19LCJzY29wZSI6Im9wZW5pZCIsInJvbGVzIjpbIm9mZmxpbmVfYWNjZXNzIiwiZGV2ZWxvcGVyIiwidW1hX2F1dGhvcml6YXRpb24iLCJ1c2VyIl0sIm5hbWUiOiJEZW1vIFVzZXIiLCJwcmVmZXJyZWRfdXNlcm5hbWUiOiJkZW1vLnVzZXIiLCJnaXZlbl9uYW1lIjoiRGVtbyIsImZhbWlseV9uYW1lIjoiVXNlciIsImVtYWlsIjoiIn0.ZDgUrqwoW2NxL2kDias7DyU4AAv8p0m2y69l8s1aE5LbGkMMzljoaERJiMpFlQLBrM3hL57XI1FpJcVgKb4VvAjxZqBUqSFIiyB38GvmvXs5xIrwUBLMYu8uicnqYKW0hZi9Gfr7Fiwzsk_t9KA-YaxeGEdOJ5K6VQV-eV8Prs5bXkyacW_Lu5ZzTAbHSZllNqRxppsjD6mBOPficeKHFhoAw-_EsT2d4DmD2DyYis32olLYaOWGBGI5y5X7yf92S_mtzy1Amy_yDqVWR2WXKfG7uPj6Tfw8yczElVRxV1OXKblzrR1E28Mui5Ll6eL5n0VOuMs3NLNW62rDFMRZ7Q")
	resp, err := client.Do(req)
	if err != nil {
		debug.PrintStack()
		t.Error(err)
		return
	}
	defer resp.Body.Close()
	if expectetToExist && resp.StatusCode != http.StatusOK {
		debug.PrintStack()
		t.Error(resp.StatusCode)
		return
	}
	if !expectetToExist && resp.StatusCode != http.StatusNotFound {
		debug.PrintStack()
		t.Error(resp.StatusCode)
		return
	}
}

func testDeployToKafka(t *testing.T, producer interfaces.Producer, deploymentId string, event deploymentmodel.MsgEvent) {
	cmd := deploymentmodel.DeploymentCommand{
		Id:      deploymentId,
		Command: "PUT",
		Deployment: deploymentmodel.Deployment{
			Id:   deploymentId,
			Name: "test-deployment",
			Elements: []deploymentmodel.Element{
				{MsgEvent: &event},
			},
		},
	}

	msg, err := json.Marshal(cmd)
	if err != nil {
		debug.PrintStack()
		t.Error(err)
		return
	}
	err = producer.Produce("test", msg)
	if err != nil {
		debug.PrintStack()
		t.Error(err)
		return
	}
}
