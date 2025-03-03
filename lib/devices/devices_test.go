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
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/event-deployment/lib/auth"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/SENERGY-Platform/event-deployment/lib/model"
	"github.com/SENERGY-Platform/event-deployment/lib/tests/docker"
	"github.com/SENERGY-Platform/event-deployment/lib/tests/mocks"
	"reflect"
	"sync"
	"testing"
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

	conf.KafkaUrl, err = docker.Kafka(ctx, wg, zkUrl)
	if err != nil {
		t.Error(err)
		return
	}

	_, mongoIp, err := docker.MongoDB(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	mongoUrl := "mongodb://" + mongoIp + ":27017"

	_, permV2Ip, err := docker.PermissionsV2(ctx, wg, mongoUrl, conf.KafkaUrl)
	if err != nil {
		t.Error(err)
		return
	}
	permv2Url := "http://" + permV2Ip + ":8080"

	_, repoIp, err := docker.DeviceRepo(ctx, wg, conf.KafkaUrl, mongoUrl, permv2Url)
	if err != nil {
		t.Error(err)
		return
	}
	conf.DeviceRepositoryUrl = "http://" + repoIp + ":8080"

	devices, err := New(conf)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("create devices", testCreateDevices(conf.DeviceRepositoryUrl, []model.Device{
		{
			Id:           "ses:infia:device:d1",
			LocalId:      "ses:infia:device:d1",
			Name:         "test-device-1",
			DeviceTypeId: "ses:infai:device-type:dt1",
		},
		{
			Id:           "ses:infia:device:d2",
			LocalId:      "ses:infia:device:d2",
			Name:         "test-device-2",
			DeviceTypeId: "ses:infai:device-type:dt1",
		},
		{
			Id:           "ses:infia:device:d3",
			LocalId:      "ses:infia:device:d3",
			Name:         "test-device-3",
			DeviceTypeId: "ses:infai:device-type:dt1",
		},
		{
			Id:           "ses:infia:device:d4",
			LocalId:      "ses:infia:device:d4",
			Name:         "test-device-4",
			DeviceTypeId: "ses:infai:device-type:dt2",
		},
		{
			Id:           "ses:infia:device:d5",
			LocalId:      "ses:infia:device:d5",
			Name:         "test-device-5",
			DeviceTypeId: "ses:infai:device-type:dt2",
		},
		{
			Id:           "ses:infia:device:d6",
			LocalId:      "ses:infia:device:d6",
			Name:         "test-device-6",
			DeviceTypeId: "ses:infai:device-type:dt2",
		},
		{
			Id:           "ses:infia:device:d7",
			LocalId:      "ses:infia:device:d7",
			Name:         "test-device-7",
			DeviceTypeId: "ses:infai:device-type:dt3",
		},
		{
			Id:           "ses:infia:device:d8",
			LocalId:      "ses:infia:device:d8",
			Name:         "test-device-8",
			DeviceTypeId: "ses:infai:device-type:dt3",
		},
		{
			Id:           "ses:infia:device:d9",
			LocalId:      "ses:infia:device:d9",
			Name:         "test-device-9",
			DeviceTypeId: "ses:infai:device-type:dt3",
		},
	}))

	t.Run("create device-groups", testCreateDeviceGroups(conf.DeviceRepositoryUrl, []model.DeviceGroup{
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

	t.Run("check GetDeviceInfosOfGroup", testCheckGetDeviceInfosOfGroupResult(
		devices,
		"ses:infia:device-group:dg1",
		[]model.Device{
			{
				Id:           "ses:infia:device:d2",
				LocalId:      "ses:infia:device:d2",
				Name:         "test-device-2",
				DeviceTypeId: "ses:infai:device-type:dt1",
				OwnerId:      jwtSubj,
			},
			{
				Id:           "ses:infia:device:d3",
				LocalId:      "ses:infia:device:d3",
				Name:         "test-device-3",
				DeviceTypeId: "ses:infai:device-type:dt1",
				OwnerId:      jwtSubj,
			},
			{
				Id:           "ses:infia:device:d4",
				LocalId:      "ses:infia:device:d4",
				Name:         "test-device-4",
				DeviceTypeId: "ses:infai:device-type:dt2",
				OwnerId:      jwtSubj,
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
			t.Errorf("\na=%#v\ne=%#v\n", actualDevices, expectedDevices)
		}
		if !reflect.DeepEqual(actualDeviceTypeIds, expectedDeviceTypeIds) {
			t.Errorf("\na=%#v\ne=%#v\n", actualDeviceTypeIds, expectedDeviceTypeIds)
		}
	}
}

var jwtSubj = "dd69ea0d-f553-4336-80f3-7f4567f85c7b"

func testCreateDeviceGroups(deviceRepoUrl string, groups []model.DeviceGroup) func(t *testing.T) {
	return func(t *testing.T) {
		c := client.NewClient(deviceRepoUrl, nil)
		token, err := auth.GenerateInternalUserToken(jwtSubj)
		if err != nil {
			t.Error(err)
			return
		}
		for _, group := range groups {
			_, err, _ := c.SetDeviceGroup(token, group)
			if err != nil {
				t.Error(err)
				return
			}
		}
	}
}

func testCreateDevices(deviceRepoUrl string, devices []model.Device) func(t *testing.T) {
	return func(t *testing.T) {
		c := client.NewClient(deviceRepoUrl, nil)
		token, err := auth.GenerateInternalUserToken(jwtSubj)
		if err != nil {
			t.Error(err)
			return
		}
		for _, device := range devices {
			_, err, _ := c.SetDevice(token, device, client.DeviceUpdateOptions{})
			if err != nil {
				t.Error(err)
				return
			}
		}
	}
}
