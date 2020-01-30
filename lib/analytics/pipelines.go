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
	"encoding/json"
	"errors"
	"net/http"
	"runtime/debug"
	"time"
)

func (this *Analytics) GetPipelinesByDeploymentId(owner string, deploymentId string) (pipelineIds []string, err error) {
	pipelineIds = []string{}
	pipelines, err := this.getPipelines(owner)
	if err != nil {
		return pipelineIds, err
	}
	for _, pipeline := range pipelines {
		desc := EventPipelineDescription{}
		err = json.Unmarshal([]byte(pipeline.Description), &desc)
		if err != nil {
			//candidate does not use event pipeline description format -> is not event pipeline -> is not searched pipeline
			err = nil
			continue
		}
		if desc.DeploymentId == deploymentId {
			pipelineIds = append(pipelineIds, pipeline.Id.String())
		}
	}
	return pipelineIds, nil
}

func (this *Analytics) GetPipelineByEventId(owner string, eventId string) (pipelineId string, exists bool, err error) {
	pipelines, err := this.getPipelines(owner)
	if err != nil {
		return pipelineId, exists, err
	}
	for _, pipeline := range pipelines {
		desc := EventPipelineDescription{}
		err = json.Unmarshal([]byte(pipeline.Description), &desc)
		if err != nil {
			//candidate does not use event pipeline description format -> is not event pipeline -> is not searched pipeline
			err = nil
			continue
		}
		if desc.EventId == eventId {
			return pipeline.Id.String(), true, nil
		}
	}
	return "", false, nil
}

func (this *Analytics) getPipelines(user string) (pipelines []Pipeline, err error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest(
		"GET",
		this.config.PipelineRepoUrl+"/pipeline",
		nil,
	)
	if err != nil {
		debug.PrintStack()
		return pipelines, err
	}
	req.Header.Set("X-UserId", user)
	resp, err := client.Do(req)
	if err != nil {
		debug.PrintStack()
		return pipelines, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		debug.PrintStack()
		return pipelines, errors.New("unexpected statuscode")
	}

	err = json.NewDecoder(resp.Body).Decode(&pipelines)
	return pipelines, err
}
