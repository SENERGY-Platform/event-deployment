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
	"fmt"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/SENERGY-Platform/permission-search/lib/tests/docker"
	"log"
	"sync"
	"time"
)

func ExampleKafka() {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, err := config.LoadConfig("../../config.json")
	if err != nil {
		log.Println(err)
		return
	}
	config.Debug = false

	_, zkIp, err := Zookeeper(ctx, wg)
	if err != nil {
		log.Println(err)
		return
	}
	zookeeperUrl := zkIp + ":2181"

	//kafka
	config.KafkaUrl, err = Kafka(ctx, wg, zookeeperUrl)
	if err != nil {
		log.Println(err)
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
		log.Println(err)
		return
	}

	producer, err := Factory.NewProducer(ctx, config, "test")
	if err != nil {
		log.Println(err)
		return
	}

	wait.Add(1)
	err = producer.Produce("key", []byte("foo"))
	if err != nil {
		log.Println(err)
		return
	}

	wait.Add(1)
	err = producer.Produce("key", []byte("bar"))
	if err != nil {
		log.Println(err)
		return
	}

	wait.Wait()
	mux.Lock()
	defer mux.Unlock()
	fmt.Println("CONSUMED:", consumed)
	//wait for finished commits
	time.Sleep(1 * time.Second)

	//output:
	//CONSUMED: [foo bar]
}

var Kafka = docker.Kafka

var Zookeeper = docker.Zookeeper
