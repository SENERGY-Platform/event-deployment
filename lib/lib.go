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

package lib

import (
	"context"
	"github.com/SENERGY-Platform/event-deployment/lib/analytics"
	"github.com/SENERGY-Platform/event-deployment/lib/api"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/SENERGY-Platform/event-deployment/lib/devices"
	"github.com/SENERGY-Platform/event-deployment/lib/events"
	"github.com/SENERGY-Platform/event-deployment/lib/imports"
	"github.com/SENERGY-Platform/event-deployment/lib/interfaces"
	"github.com/SENERGY-Platform/event-deployment/lib/kafka"
	"log"
)

func StartDefault(ctx context.Context, config config.Config) error {
	return Start(ctx, config, kafka.Factory, events.Factory, analytics.Factory, devices.Factory, api.Start)
}

type Producer interface {
	Produce(key string, message []byte) error
}

func Start(ctx context.Context, config config.Config, sourcing interfaces.SourcingFactory, events interfaces.EventsFactory, analytics interfaces.AnalyticsFactory, devices interfaces.DevicesFactory, apiFactory func(ctx context.Context, config config.Config, ctrl interfaces.Events) error) error {
	a, err := analytics.New(ctx, config)
	if err != nil {
		return err
	}
	var producer Producer
	if config.DeploymentDoneTopic != "" && config.DeploymentDoneTopic != "-" {
		log.Println("use deployment done producer")
		producer, err = sourcing.NewProducer(ctx, config, config.DeploymentDoneTopic)
		if err != nil {
			return err
		}
	}
	event, err := events.New(ctx, config, a, devices.New(config), imports.New(config), producer)
	if err != nil {
		return err
	}
	err = sourcing.NewConsumer(ctx, config, config.DeploymentTopic, event.HandleCommand)
	if err != nil {
		return err
	}
	if config.DeviceGroupTopic != "" {
		err = sourcing.NewConsumer(ctx, config, config.DeviceGroupTopic, event.HandleDeviceGroupUpdate)
		if err != nil {
			return err
		}
	}
	return apiFactory(ctx, config, event)
}
