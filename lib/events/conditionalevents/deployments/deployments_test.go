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

package deployments

import (
	"context"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/SENERGY-Platform/event-deployment/lib/model"
	"github.com/SENERGY-Platform/event-deployment/lib/tests/docker"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel"
	"reflect"
	"sort"
	"sync"
	"testing"
)

func TestDeployments(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, err := config.LoadConfig("../../../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	_, mongoIp, err := docker.Mongo(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	config.ConditionalEventRepoMongoUrl = "mongodb://" + mongoIp + ":27017"

	deployments, err := New(ctx, wg, config)
	if err != nil {
		t.Error(err)
		return
	}

	deploy := func(d deploymentmodel.Deployment) error {
		return deployments.SetDeployment(model.Deployment{
			Deployment: d,
			UserId:     "userid",
		})
	}

	t.Run("set deployments", func(t *testing.T) {
		err = deploy(deploymentmodel.Deployment{
			Id:   "duplicate",
			Name: "duplicate",
		})
		if err != nil {
			t.Error(err)
			return
		}
		err = deploy(deploymentmodel.Deployment{
			Id:   "duplicate",
			Name: "duplicate",
			Elements: []deploymentmodel.Element{
				{
					ConditionalEvent: &deploymentmodel.ConditionalEvent{
						Selection: deploymentmodel.Selection{
							SelectedDeviceGroupId: ptr("g1"),
						},
					},
				},
				{
					ConditionalEvent: &deploymentmodel.ConditionalEvent{
						Selection: deploymentmodel.Selection{
							SelectedDeviceGroupId: ptr("g2"),
						},
					},
				},
			},
		})
		if err != nil {
			t.Error(err)
			return
		}
		err = deploy(deploymentmodel.Deployment{
			Id:   "deleted",
			Name: "deleted",
			Elements: []deploymentmodel.Element{
				{
					ConditionalEvent: &deploymentmodel.ConditionalEvent{
						Selection: deploymentmodel.Selection{
							SelectedDeviceGroupId: ptr("g1"),
						},
					},
				},
				{
					ConditionalEvent: &deploymentmodel.ConditionalEvent{
						Selection: deploymentmodel.Selection{
							SelectedDeviceGroupId: ptr("g2"),
						},
					},
				},
			},
		})
		if err != nil {
			t.Error(err)
			return
		}
		err = deploy(deploymentmodel.Deployment{
			Id:   "second",
			Name: "second",
			Elements: []deploymentmodel.Element{
				{
					ConditionalEvent: &deploymentmodel.ConditionalEvent{
						Selection: deploymentmodel.Selection{
							SelectedDeviceGroupId: ptr("g2"),
						},
					},
				},
			},
		})
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("remove", func(t *testing.T) {
		err = deployments.RemoveDeployment("deleted")
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("read g1", func(t *testing.T) {
		result, err := deployments.GetDeploymentByDeviceGroupId("g1")
		if err != nil {
			t.Error(err)
			return
		}
		ids := []string{}
		for _, e := range result {
			ids = append(ids, e.Id)
		}
		expeted := []string{"duplicate"}
		sort.Strings(expeted)
		sort.Strings(ids)
		if !reflect.DeepEqual(ids, expeted) {
			t.Errorf("\n%#v\n%#v", expeted, ids)
			return
		}
	})

	t.Run("read g2", func(t *testing.T) {
		result, err := deployments.GetDeploymentByDeviceGroupId("g2")
		if err != nil {
			t.Error(err)
			return
		}
		ids := []string{}
		for _, e := range result {
			ids = append(ids, e.Id)
		}
		expeted := []string{"duplicate", "second"}
		sort.Strings(expeted)
		sort.Strings(ids)
		if !reflect.DeepEqual(ids, expeted) {
			t.Errorf("\n%#v\n%#v", expeted, ids)
			return
		}
	})

}

func ptr(s string) *string {
	return &s
}
