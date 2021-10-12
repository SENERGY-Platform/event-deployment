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
	"github.com/SENERGY-Platform/event-deployment/lib/model"
	"net/http"
	"runtime/debug"
	"strconv"
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

func (this *Analytics) GetPipelinesByDeviceGroupId(owner string, groupId string) (pipelineIds []string, pipelineToGroupDescription map[string]model.GroupEventDescription, pipelineNames map[string]string, err error) {
	pipelineToGroupDescription = map[string]model.GroupEventDescription{}
	pipelineNames = map[string]string{}
	pipelines, err := this.getPipelines(owner)
	if err != nil {
		return pipelineIds, pipelineToGroupDescription, pipelineNames, err
	}
	for _, pipeline := range pipelines {
		desc := EventPipelineDescription{}
		err = json.Unmarshal([]byte(pipeline.Description), &desc)
		if err != nil {
			//candidate does not use event pipeline description format -> is not event pipeline -> is not searched pipeline
			err = nil
			continue
		}
		if desc.DeviceGroupId == groupId {
			id := pipeline.Id.String()
			pipelineNames[id] = pipeline.Name
			pipelineIds = append(pipelineIds, id)
			pipelineToGroupDescription[id] = model.GroupEventDescription{
				DeviceGroupId: desc.DeviceGroupId,
				EventId:       desc.EventId,
				DeploymentId:  desc.DeploymentId,
				FunctionId:    desc.FunctionId,
				AspectId:      desc.AspectId,
				FlowId:        desc.FlowId,
				OperatorValue: desc.OperatorValue,
			}
		}
	}
	return pipelineIds, pipelineToGroupDescription, pipelineNames, err
}

func (this *Analytics) GetEventStates(owner string, eventIds []string) (states map[string]bool, err error) {
	states = map[string]bool{}
	pipelines, err := this.getPipelines(owner)
	if err != nil {
		return states, err
	}
	allEvents := map[string]bool{}
	for _, pipeline := range pipelines {
		desc := EventPipelineDescription{}
		err = json.Unmarshal([]byte(pipeline.Description), &desc)
		if err != nil {
			//candidate does not use event pipeline description format -> is not event pipeline -> is not searched pipeline
			err = nil
			continue
		}
		if desc.EventId != "" {
			allEvents[desc.EventId] = true
		}
	}
	for _, eventId := range eventIds {
		if allEvents[eventId] {
			states[eventId] = true
		} else {
			states[eventId] = false
		}
	}
	return states, nil
}

func (this *Analytics) getPipelines(user string) (pipelines []Pipeline, err error) {
	limit := 500
	offset := 0
	for {
		temp, err := this.getSomePipelines(user, limit, offset)
		if err != nil {
			return pipelines, err
		}
		if temp != nil {
			pipelines = append(pipelines, temp...)
		}
		if len(temp) < limit {
			return pipelines, nil
		} else {
			offset = offset + limit
		}
	}
}

func (this *Analytics) getSomePipelines(user string, limit int, offset int) (pipelines []Pipeline, err error) {
	client := http.Client{
		Timeout: this.timeout,
	}
	req, err := http.NewRequest(
		"GET",
		this.config.PipelineRepoUrl+"/pipeline?limit="+strconv.Itoa(limit)+"&offset="+strconv.Itoa(offset),
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
