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

package kafka

import (
	"context"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/SENERGY-Platform/event-deployment/lib/tests/docker"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestKafka(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, err := config.LoadConfig("../../config.json")
	if err != nil {
		t.Error(err)
		return
	}
	config.Debug = false
	config.InitTopics = true

	_, zkIp, err := Zookeeper(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	zookeeperUrl := zkIp + ":2181"

	//kafka
	config.KafkaUrl, err = Kafka(ctx, wg, zookeeperUrl)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(1 * time.Second)

	consumed := []string{}
	mux := sync.Mutex{}

	wait := sync.WaitGroup{}

	err = Factory.NewConsumer(ctx, config, "test", func(delivery []byte) error {
		mux.Lock()
		defer mux.Unlock()
		consumed = append(consumed, string(delivery))
		wait.Done()
		return nil
	})

	if err != nil {
		t.Error(err)
		return
	}

	producer, err := Factory.NewProducer(ctx, config, "test")
	if err != nil {
		t.Error(err)
		return
	}

	wait.Add(1)
	err = producer.Produce("key", []byte("foo"))
	if err != nil {
		t.Error(err)
		return
	}

	wait.Add(1)
	err = producer.Produce("key", []byte("bar"))
	if err != nil {
		t.Error(err)
		return
	}

	wait.Wait()
	mux.Lock()
	defer mux.Unlock()

	if !reflect.DeepEqual(consumed, []string{"foo", "bar"}) {
		t.Error(consumed)
		return
	}
}

var Kafka = docker.Kafka

var Zookeeper = docker.Zookeeper
