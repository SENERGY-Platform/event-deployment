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

package devices

import (
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/SENERGY-Platform/event-deployment/lib/model"
	"github.com/SENERGY-Platform/event-deployment/lib/tests/docker"
	"github.com/SENERGY-Platform/event-deployment/lib/tests/mocks"
	"github.com/SENERGY-Platform/event-deployment/lib/tests/util"
	"github.com/segmentio/kafka-go"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestDevices(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf := &config.ConfigStruct{
		AuthExpirationTimeBuffer: 0,
		AuthEndpoint:             "",
		AuthClientId:             "ignored",
		AuthClientSecret:         "ignored",
		PermSearchUrl:            "",
	}

	err := mocks.MockAuthServer(conf, ctx)
	if err != nil {
		t.Error(err)
		return
	}

	_, zk, err := docker.Zookeeper(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	zkUrl := zk + ":2181"

	kport, err := docker.Kafka(ctx, wg, zkUrl)
	if err != nil {
		t.Error(err)
		return
	}

	_, esIp, err := docker.ElasticSearch(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	permSearchPort, _, err := docker.PermSearch(ctx, wg, zkUrl, esIp)
	if err != nil {
		t.Error(err)
		return
	}
	conf.PermSearchUrl = "http://localhost:" + permSearchPort

	devices := New(conf)

	time.Sleep(5 * time.Second)

	t.Run("create devices", testCreateDevices(kport, []model.Device{
		{
			Id:           "ses:infia:device:d1",
			Name:         "test-device-1",
			DeviceTypeId: "ses:infai:device-type:dt1",
		},
		{
			Id:           "ses:infia:device:d2",
			Name:         "test-device-2",
			DeviceTypeId: "ses:infai:device-type:dt1",
		},
		{
			Id:           "ses:infia:device:d3",
			Name:         "test-device-3",
			DeviceTypeId: "ses:infai:device-type:dt1",
		},
		{
			Id:           "ses:infia:device:d4",
			Name:         "test-device-4",
			DeviceTypeId: "ses:infai:device-type:dt2",
		},
		{
			Id:           "ses:infia:device:d5",
			Name:         "test-device-5",
			DeviceTypeId: "ses:infai:device-type:dt2",
		},
		{
			Id:           "ses:infia:device:d6",
			Name:         "test-device-6",
			DeviceTypeId: "ses:infai:device-type:dt2",
		},
		{
			Id:           "ses:infia:device:d7",
			Name:         "test-device-7",
			DeviceTypeId: "ses:infai:device-type:dt3",
		},
		{
			Id:           "ses:infia:device:d8",
			Name:         "test-device-8",
			DeviceTypeId: "ses:infai:device-type:dt3",
		},
		{
			Id:           "ses:infia:device:d9",
			Name:         "test-device-9",
			DeviceTypeId: "ses:infai:device-type:dt3",
		},
	}))

	t.Run("create device-groups", testCreateDeviceGroups(kport, []model.DeviceGroup{
		{
			Id:   "ses:infia:device-group:dg1",
			Name: "test-group-1",
			DeviceIds: []string{
				"ses:infia:device:d2",
				"ses:infia:device:d3",
				"ses:infia:device:d4",
			},
		},
		{
			Id:   "ses:infia:device-group:dg2",
			Name: "test-group-2",
			DeviceIds: []string{
				"ses:infia:device:d9",
			},
		},
	}))

	time.Sleep(10 * time.Second) // wait for consumption of kafka messages

	t.Run("check GetDeviceInfosOfGroup", testCheckGetDeviceInfosOfGroupResult(
		devices,
		"ses:infia:device-group:dg1",
		[]model.Device{
			{
				Id:           "ses:infia:device:d2",
				Name:         "test-device-2",
				DeviceTypeId: "ses:infai:device-type:dt1",
			},
			{
				Id:           "ses:infia:device:d3",
				Name:         "test-device-3",
				DeviceTypeId: "ses:infai:device-type:dt1",
			},
			{
				Id:           "ses:infia:device:d4",
				Name:         "test-device-4",
				DeviceTypeId: "ses:infai:device-type:dt2",
			},
		}, []string{
			"ses:infai:device-type:dt1",
			"ses:infai:device-type:dt2",
		}))
}

func testCheckGetDeviceInfosOfGroupResult(repo *Devices, deviceGroupId string, expectedDevices []model.Device, expectedDeviceTypeIds []string) func(t *testing.T) {
	return func(t *testing.T) {
		actualDevices, actualDeviceTypeIds, err, _ := repo.GetDeviceInfosOfGroup(deviceGroupId)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(actualDevices, expectedDevices) {
			actualJson, _ := json.Marshal(actualDevices)
			expectedJson, _ := json.Marshal(expectedDevices)
			t.Error(string(actualJson), "\n", string(expectedJson))
		}
		if !reflect.DeepEqual(actualDeviceTypeIds, expectedDeviceTypeIds) {
			actualJson, _ := json.Marshal(actualDeviceTypeIds)
			expectedJson, _ := json.Marshal(expectedDeviceTypeIds)
			t.Error(string(actualJson), "\n", string(expectedJson))
		}
	}
}

var jwtSubj = "dd69ea0d-f553-4336-80f3-7f4567f85c7b"

func testCreateDeviceGroups(kafkaPort int, groups []model.DeviceGroup) func(t *testing.T) {
	return func(t *testing.T) {
		topic := "device-groups"
		producer, err := util.GetKafkaProducer([]string{"127.0.0.1:" + strconv.Itoa(kafkaPort)}, topic)
		if err != nil {
			t.Error(err)
			return
		}
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		defer producer.Close()
		for _, group := range groups {
			msg, err := json.Marshal(map[string]interface{}{
				"command":      "PUT",
				"id":           group.Id,
				"owner":        jwtSubj,
				"device_group": group,
			})
			if err != nil {
				t.Error(err)
				return
			}
			err = producer.WriteMessages(ctx, kafka.Message{
				Key:   []byte(group.Id),
				Value: msg,
			})
			if err != nil {
				t.Error(err)
				return
			}
		}
	}
}

func testCreateDevices(kafkaPort int, devices []model.Device) func(t *testing.T) {
	return func(t *testing.T) {
		topic := "devices"
		producer, err := util.GetKafkaProducer([]string{"127.0.0.1:" + strconv.Itoa(kafkaPort)}, topic)
		if err != nil {
			t.Error(err)
			return
		}
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		defer producer.Close()
		for _, device := range devices {
			msg, err := json.Marshal(map[string]interface{}{
				"command": "PUT",
				"id":      device.Id,
				"owner":   jwtSubj,
				"device":  device,
			})
			if err != nil {
				t.Error(err)
				return
			}
			err = producer.WriteMessages(ctx, kafka.Message{
				Key:   []byte(device.Id),
				Value: msg,
			})
			if err != nil {
				t.Error(err)
				return
			}
		}
	}
}
