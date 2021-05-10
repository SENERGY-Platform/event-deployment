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

package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type ConfigStruct struct {
	LogLevel                string `json:"log_level"`
	ApiPort                 string `json:"api_port"`
	MarshallerUrl           string `json:"marshaller_url"`
	ConverterUrl            string `json:"converter_url"`
	KafkaUrl                string `json:"kafka_url"`
	FlowEngineUrl           string `json:"flow_engine_url"`
	FlowParserUrl           string `json:"flow_parser_url"`
	PipelineRepoUrl         string `json:"pipeline_repo_url"`
	ImportDeployUrl         string `json:"import_deploy_url"`
	ConsumerGroup           string `json:"consumer_group"`
	Debug                   bool   `json:"debug"`
	DeploymentTopic         string `json:"deployment_topic"`
	ConnectivityTest        bool   `json:"connectivity_test"`
	CamundaEventTriggerPath string `json:"camunda_event_trigger_path"`
	ShardsDb                string `json:"shards_db"`
	AnalyticsRequestTimeout string `json:"analytics_request_timeout"`

	//if not configured: no device-group updates handled
	DeviceGroupTopic string `json:"device_group_topic"`

	//if not configured: events with groups not handled
	PermSearchUrl            string  `json:"perm_search_url"`
	AuthExpirationTimeBuffer float64 `json:"auth_expiration_time_buffer"`
	AuthEndpoint             string  `json:"auth_endpoint"`
	AuthClientId             string  `json:"auth_client_id"`
	AuthClientSecret         string  `json:"auth_client_secret"`
}

type Config = *ConfigStruct

func LoadConfig(location string) (config Config, err error) {
	file, error := os.Open(location)
	if error != nil {
		log.Println("error on config load: ", error)
		return config, error
	}
	decoder := json.NewDecoder(file)
	error = decoder.Decode(&config)
	if error != nil {
		log.Println("invalid config json: ", error)
		return config, error
	}
	HandleEnvironmentVars(config)
	return config, nil
}

var camel = regexp.MustCompile("(^[^A-Z]*|[A-Z]*)([A-Z][^A-Z]+|$)")

func fieldNameToEnvName(s string) string {
	var a []string
	for _, sub := range camel.FindAllStringSubmatch(s, -1) {
		if sub[1] != "" {
			a = append(a, sub[1])
		}
		if sub[2] != "" {
			a = append(a, sub[2])
		}
	}
	return strings.ToUpper(strings.Join(a, "_"))
}

// preparations for docker
func HandleEnvironmentVars(config Config) {
	configValue := reflect.Indirect(reflect.ValueOf(config))
	configType := configValue.Type()
	for index := 0; index < configType.NumField(); index++ {
		fieldName := configType.Field(index).Name
		envName := fieldNameToEnvName(fieldName)
		envValue := os.Getenv(envName)
		if envValue != "" {
			fmt.Println("use environment variable: ", envName, " = ", envValue)
			if configValue.FieldByName(fieldName).Kind() == reflect.Int64 {
				i, _ := strconv.ParseInt(envValue, 10, 64)
				configValue.FieldByName(fieldName).SetInt(i)
			}
			if configValue.FieldByName(fieldName).Kind() == reflect.Float64 {
				f, _ := strconv.ParseFloat(envValue, 64)
				configValue.FieldByName(fieldName).SetFloat(f)
			}
			if configValue.FieldByName(fieldName).Kind() == reflect.String {
				configValue.FieldByName(fieldName).SetString(envValue)
			}
			if configValue.FieldByName(fieldName).Kind() == reflect.Bool {
				b, _ := strconv.ParseBool(envValue)
				configValue.FieldByName(fieldName).SetBool(b)
			}
			if configValue.FieldByName(fieldName).Kind() == reflect.Slice {
				val := []string{}
				for _, element := range strings.Split(envValue, ",") {
					val = append(val, strings.TrimSpace(element))
				}
				configValue.FieldByName(fieldName).Set(reflect.ValueOf(val))
			}
			if configValue.FieldByName(fieldName).Kind() == reflect.Map {
				value := map[string]string{}
				for _, element := range strings.Split(envValue, ",") {
					keyVal := strings.Split(element, ":")
					key := strings.TrimSpace(keyVal[0])
					val := strings.TrimSpace(keyVal[1])
					value[key] = val
				}
				configValue.FieldByName(fieldName).Set(reflect.ValueOf(value))
			}
		}
	}
}
