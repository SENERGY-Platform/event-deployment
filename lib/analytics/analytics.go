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
	"github.com/SENERGY-Platform/event-deployment/lib/auth"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/SENERGY-Platform/event-deployment/lib/interfaces"
	"strings"
	"time"
)

type FactoryType struct{}

var Factory = &FactoryType{}

type Analytics struct {
	config  config.Config
	timeout time.Duration
	auth    *auth.Auth
}

func (this *FactoryType) New(ctx context.Context, config config.Config) (interfaces.Analytics, error) {
	timeout, err := time.ParseDuration(config.AnalyticsRequestTimeout)
	if err != nil {
		return nil, err
	}
	a, err := auth.NewAuth(config)
	if err != nil {
		return nil, err
	}
	return &Analytics{config: config, timeout: timeout, auth: a}, nil
}

func trimIdParams(id string) (result string) {
	result, _, _ = strings.Cut(id, "$")
	return result
}
