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
	"context"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"reflect"
	"sync"
	"time"
)

type Deployments struct {
	config config.Config
	client *mongo.Client
	ctx    context.Context
}

var CreateCollections = []func(db *Deployments) error{}

func New(ctx context.Context, wg *sync.WaitGroup, conf config.Config) (*Deployments, error) {
	timeout, _ := getTimeoutContext(ctx)
	reg := bson.NewRegistryBuilder().RegisterTypeMapEntry(bsontype.EmbeddedDocument, reflect.TypeOf(bson.M{})).Build() //ensure map marshalling to interface
	client, err := mongo.Connect(timeout, options.Client().ApplyURI(conf.ConditionalEventRepoMongoUrl), options.Client().SetRegistry(reg))
	if err != nil {
		return nil, err
	}
	db := &Deployments{config: conf, client: client, ctx: ctx}
	for _, creators := range CreateCollections {
		err = creators(db)
		if err != nil {
			client.Disconnect(context.Background())
			return nil, err
		}
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		client.Disconnect(context.Background())
	}()
	return db, nil
}

func getTimeoutContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, 10*time.Second)
}

func (this *Deployments) getTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(this.ctx, 10*time.Second)
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
