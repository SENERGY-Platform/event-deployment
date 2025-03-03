/*
 * Copyright (c) 2022 InfAI (CC SES)
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
	"github.com/SENERGY-Platform/event-deployment/lib/model"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"runtime/debug"
)

type DeploymentIndex struct {
	Deployment   deploymentmodel.Deployment `json:"deployment" bson:"deployment"`
	UserId       string                     `json:"user_id" bson:"user_id"`
	DeviceGroups []string                   `json:"device_groups" bson:"device_groups"`
	Id           string                     `json:"id" bson:"id"`
}

type Deployment = model.Deployment

func init() {
	CreateCollections = append(CreateCollections, func(db *Deployments) error {
		var err error
		collection := db.client.Database(db.config.ConditionalEventRepoMongoTable).Collection(db.config.ConditionalEventRepoMongoDeploymentsCollection)
		err = db.ensureIndex(collection, "deployment_id_index", "id", true, true)
		if err != nil {
			debug.PrintStack()
			return err
		}
		err = db.ensureIndex(collection, "deployment_device_group_index", "device_groups", true, false)
		if err != nil {
			debug.PrintStack()
			return err
		}
		return nil
	})
}

func (this *Deployments) deploymentsCollection() *mongo.Collection {
	return this.client.Database(this.config.ConditionalEventRepoMongoTable).Collection(this.config.ConditionalEventRepoMongoDeploymentsCollection)
}

func (this *Deployments) GetDeploymentByDeviceGroupId(deviceGroupId string) (result []Deployment, err error) {
	if deviceGroupId == "" {
		return []Deployment{}, nil
	}
	ctx, _ := this.getTimeoutContext()
	cursor, err := this.deploymentsCollection().Find(ctx, bson.M{"device_groups": deviceGroupId})
	if err != nil {
		return result, err
	}
	temp, err, _ := readCursorResult[DeploymentIndex](ctx, cursor)
	if err != nil {
		return result, err
	}
	for _, e := range temp {
		result = append(result, Deployment{
			Deployment: e.Deployment,
			UserId:     e.UserId,
		})
	}
	return result, err
}

func (this *Deployments) RemoveDeployment(deploymentId string) (err error) {
	ctx, _ := this.getTimeoutContext()
	_, err = this.deploymentsCollection().DeleteMany(ctx, bson.M{"id": deploymentId})
	return err
}

func (this *Deployments) SetDeployment(element Deployment) (err error) {
	ctx, _ := this.getTimeoutContext()
	_, err = this.deploymentsCollection().ReplaceOne(ctx, bson.M{"id": element.Id}, getDeploymentIndex(element), options.Replace().SetUpsert(true))
	return err
}

func getDeploymentIndex(depl Deployment) (result DeploymentIndex) {
	result = DeploymentIndex{
		Id:           depl.Id,
		UserId:       depl.UserId,
		Deployment:   depl.Deployment,
		DeviceGroups: nil,
	}
	for _, element := range depl.Elements {
		if element.ConditionalEvent != nil &&
			element.ConditionalEvent.Selection.SelectedDeviceGroupId != nil &&
			*element.ConditionalEvent.Selection.SelectedDeviceGroupId != "" {
			result.DeviceGroups = append(result.DeviceGroups, *element.ConditionalEvent.Selection.SelectedDeviceGroupId)
		}
	}
	return result
}
