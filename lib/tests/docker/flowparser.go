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
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"log"
	"net/http"
	"sync"
)

func FlowParser(ctx context.Context, wg *sync.WaitGroup, flowRepoIp string) (hostport string, containerip string, err error) {
	pool, err := dockertest.NewPool("")
	container, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "ghcr.io/senergy-platform/analytics-flow-parser",
		Tag:        "latest",
		Env:        []string{"FLOW_API_ENDPOINT=http://" + flowRepoIp + ":5000"},
	}, func(config *docker.HostConfig) {
	})
	if err != nil {
		return hostport, containerip, err
	}
	wg.Add(1)
	go func() {
		<-ctx.Done()
		log.Println("DEBUG: remove container " + container.Container.Name)
		container.Close()
		wg.Done()
	}()
	hostport = container.GetPort("5000/tcp")
	containerip = container.Container.NetworkSettings.IPAddress
	err = pool.Retry(func() error {
		_, err = http.Get("http://localhost:" + hostport)
		return err
	})
	return
}
