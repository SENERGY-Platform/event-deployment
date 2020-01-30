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

import uuid "github.com/satori/go.uuid"

type PipelineRequest struct {
	Id          string         `json:"id,omitempty"`
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	WindowTime  int            `json:"windowTime,omitempty"`
	Nodes       []PipelineNode `json:"nodes,omitempty"`
}

type PipelineNode struct {
	NodeId string       `json:"nodeId, omitempty"`
	Inputs []NodeInput  `json:"inputs,omitempty"`
	Config []NodeConfig `json:"config,omitempty"`
}

type NodeConfig struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

type NodeInput struct {
	DeviceId  string      `json:"deviceId,omitempty"`
	TopicName string      `json:"topicName,omitempty"`
	Values    []NodeValue `json:"values,omitempty"`
}

type NodeValue struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
}

type Pipeline struct {
	Id          uuid.UUID  `json:"id,omitempty"`
	Name        string     `json:"name,omitempty"`
	Description string     `json:"description,omitempty"`
	Operators   []Operator `json:"operators,omitempty"`
}

type Operator struct {
	Id             string            `json:"id,omitempty"`
	Name           string            `json:"name,omitempty"`
	ImageId        string            `json:"imageId,omitempty"`
	DeploymentType string            `json:"deploymentType,omitempty"`
	OperatorId     string            `json:"operatorId,omitempty"`
	Config         map[string]string `json:"config,omitempty"`
	InputTopics    []InputTopic
}

type InputTopic struct {
	Name        string    `json:"name,omitempty"`
	FilterType  string    `json:"filterType,omitempty"`
	FilterValue string    `json:"filterValue,omitempty"`
	Mappings    []Mapping `json:"mappings,omitempty"`
}

type Mapping struct {
	Dest   string `json:"dest,omitempty"`
	Source string `json:"source,omitempty"`
}

type EventNode struct {
	Id      string       `json:"id"`
	Name    string       `json:"name"`
	Configs []NodeConfig `json:"configs"`
}

type CellConfig struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
}

type FlowModelCell struct {
	Id             string                 `json:"id"`
	Name           string                 `json:"name"`
	DeploymentType string                 `json:"deploymentType"`
	InPorts        []string               `json:"inPorts,omitempty"`
	OutPorts       []string               `json:"outPorts,omitempty"`
	Type           string                 `json:"type"`
	Source         map[string]interface{} `json:"source"`
	Target         map[string]interface{} `json:"target"`
	Image          string                 `json:"image"`
	Config         []CellConfig           `json:"config,omitempty"`
	OperatorId     string                 `json:"operatorId"`
}

type FlowModel struct {
	Cells []FlowModelCell `json:"cells"`
}

type Flow struct {
	Id          string    `json:"_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Model       FlowModel `json:"model"`
}

type EventPipelineDescription struct {
	DeviceId      string `json:"device_id"`
	ServiceId     string `json:"service_id"`
	ValuePath     string `json:"value_path"`
	OperatorValue string `json:"operator_value"`
	EventId       string `json:"event_id"`
	DeploymentId  string `json:"deployment_id"`
}
