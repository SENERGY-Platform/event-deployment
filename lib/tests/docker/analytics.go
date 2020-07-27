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

package docker

import (
	"context"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"sync"
)

func Analytics(ctx context.Context, wg *sync.WaitGroup, conf config.Config) (err error) {
	_, mongoIp, err := Mongo(ctx, wg)
	if err != nil {
		return err
	}
	_, flowrepoIp, err := FlowRepo(ctx, wg, mongoIp)
	if err != nil {
		return err
	}

	flowPort, _, err := FlowParser(ctx, wg, flowrepoIp)
	if err != nil {
		return err
	}
	conf.FlowParserUrl = "http://localhost:" + flowPort
	return nil
}
