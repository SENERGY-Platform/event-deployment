/*
 * Copyright 2021 InfAI (CC SES)
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

package mocks

import (
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/process-deployment/lib/model/importmodel"
	"net/http"
	"strings"
)

type ImportsMock struct{}

func (this *ImportsMock) GetImportInstance(user string, importId string) (importInstance importmodel.Import, err error, code int) {
	return importmodel.Import{
		Id:           importId,
		ImportTypeId: "urn:infai:ses:import-type:a93420ae-ff5f-4c44-ee6b-5d3313f946d2",
	}, nil, http.StatusOK
}

func (this *ImportsMock) GetImportType(user string, importTypeId string) (importType importmodel.ImportType, err error, code int) {
	str := `{
   "id":"urn:infai:ses:import-type:a93420ae-ff5f-4c44-ee6b-5d3313f946d2",
   "name":"yr-forecast",
   "description":"Weather forecast by yr.no",
   "image":"ghcr.io/senergy-platform/import-yr-forecast:prod",
   "default_restart":true,
   "configs":[
      {
         "name":"lat",
         "description":"Coordinate latitude",
         "type":"https://schema.org/Float",
         "default_value":51.34
      },
      {
         "name":"long",
         "description":"Coordinate longitude",
         "type":"https://schema.org/Float",
         "default_value":12.38
      },
      {
         "name":"altitude",
         "description":"altitude above sea level (optional, use -1 to indicate no value)",
         "type":"https://schema.org/Integer",
         "default_value":-1
      },
      {
         "name":"max_forecasts",
         "description":"Maximum number of forecasts you wish to import at a given time. A value of 1 indicates that you are only interested in the current data.",
         "type":"https://schema.org/Integer",
         "default_value":1
      }
   ],
   "aspect_ids":null,
   "output":{
      "name":"root",
      "type":"https://schema.org/StructuredValue",
      "sub_content_variables":[
         {
            "name":"import_id",
            "type":"https://schema.org/Text"
         },
         {
            "name":"time",
            "type":"https://schema.org/Text",
            "characteristic_id":"urn:infai:ses:characteristic:6bc41b45-a9f3-4d87-9c51-dd3e11257800"
         },
         {
            "name":"value",
            "type":"https://schema.org/StructuredValue",
            "sub_content_variables":[
               {
                  "name":"units",
                  "type":"https://schema.org/StructuredValue",
                  "sub_content_variables":[
                     {
                        "name":"air_pressure_at_sea_level",
                        "type":"https://schema.org/Text"
                     },
                     {
                        "name":"air_temperature",
                        "type":"https://schema.org/Text"
                     },
                     {
                        "name":"air_temperature_max",
                        "type":"https://schema.org/Text"
                     },
                     {
                        "name":"air_temperature_min",
                        "type":"https://schema.org/Text"
                     },
                     {
                        "name":"cloud_area_fraction",
                        "type":"https://schema.org/Text"
                     },
                     {
                        "name":"cloud_area_fraction_high",
                        "type":"https://schema.org/Text"
                     },
                     {
                        "name":"cloud_area_fraction_low",
                        "type":"https://schema.org/Text"
                     },
                     {
                        "name":"cloud_area_fraction_medium",
                        "type":"https://schema.org/Text"
                     },
                     {
                        "name":"dew_point_temperature",
                        "type":"https://schema.org/Text"
                     },
                     {
                        "name":"fog_area_fraction",
                        "type":"https://schema.org/Text"
                     },
                     {
                        "name":"precipitation_amount",
                        "type":"https://schema.org/Text"
                     },
                     {
                        "name":"precipitation_amount_max",
                        "type":"https://schema.org/Text"
                     },
                     {
                        "name":"precipitation_amount_min",
                        "type":"https://schema.org/Text"
                     },
                     {
                        "name":"probability_of_precipitation",
                        "type":"https://schema.org/Text"
                     },
                     {
                        "name":"probability_of_thunder",
                        "type":"https://schema.org/Text"
                     },
                     {
                        "name":"relative_humidity",
                        "type":"https://schema.org/Text"
                     },
                     {
                        "name":"ultraviolet_index_clear_sky",
                        "type":"https://schema.org/Text"
                     },
                     {
                        "name":"wind_from_direction",
                        "type":"https://schema.org/Text"
                     },
                     {
                        "name":"wind_speed",
                        "type":"https://schema.org/Text"
                     },
                     {
                        "name":"wind_speed_of_gust",
                        "type":"https://schema.org/Text"
                     }
                  ]
               },
               {
                  "name":"forecasted_for",
                  "type":"https://schema.org/Text",
                  "characteristic_id":"urn:infai:ses:characteristic:6bc41b45-a9f3-4d87-9c51-dd3e11257800",
                  "use_as_tag":true
               },
               {
                  "name":"instant_air_pressure_at_sea_level",
                  "type":"https://schema.org/Float",
                  "characteristic_id":"urn:infai:ses:characteristic:eb33cf65-b0a2-413d-891d-cada05be01ed"
               },
               {
                  "name":"instant_air_temperature",
                  "type":"https://schema.org/Float",
                  "characteristic_id":"urn:infai:ses:characteristic:5ba31623-0ccb-4488-bfb7-f73b50e03b5a"
               },
               {
                  "name":"instant_cloud_area_fraction",
                  "type":"https://schema.org/Float",
                  "characteristic_id":"urn:infai:ses:characteristic:46f808f4-bb9e-4cc2-bd50-dc33ca74f273"
               },
               {
                  "name":"instant_cloud_area_fraction_high",
                  "type":"https://schema.org/Float",
                  "characteristic_id":"urn:infai:ses:characteristic:46f808f4-bb9e-4cc2-bd50-dc33ca74f273"
               },
               {
                  "name":"instant_cloud_area_fraction_low",
                  "type":"https://schema.org/Float",
                  "characteristic_id":"urn:infai:ses:characteristic:46f808f4-bb9e-4cc2-bd50-dc33ca74f273"
               },
               {
                  "name":"instant_cloud_area_fraction_medium",
                  "type":"https://schema.org/Float",
                  "characteristic_id":"urn:infai:ses:characteristic:46f808f4-bb9e-4cc2-bd50-dc33ca74f273"
               },
               {
                  "name":"instant_dew_point_temperature",
                  "type":"https://schema.org/Float",
                  "characteristic_id":"urn:infai:ses:characteristic:5ba31623-0ccb-4488-bfb7-f73b50e03b5a"
               },
               {
                  "name":"instant_fog_area_fraction",
                  "type":"https://schema.org/Float",
                  "characteristic_id":"urn:infai:ses:characteristic:46f808f4-bb9e-4cc2-bd50-dc33ca74f273"
               },
               {
                  "name":"instant_relative_humidity",
                  "type":"https://schema.org/Float",
                  "characteristic_id":"urn:infai:ses:characteristic:46f808f4-bb9e-4cc2-bd50-dc33ca74f273"
               },
               {
                  "name":"instant_ultraviolet_index_clear_sky",
                  "type":"https://schema.org/Integer",
                  "characteristic_id":"urn:infai:ses:characteristic:0a61343d-c0d1-4af8-9329-3829c30ba59f"
               },
               {
                  "name":"instant_wind_from_direction",
                  "type":"https://schema.org/Float",
                  "characteristic_id":"urn:infai:ses:characteristic:005ca5d3-44da-4fc8-b2b8-a88e3209e9f7"
               },
               {
                  "name":"instant_wind_speed",
                  "type":"https://schema.org/Float",
                  "characteristic_id":"urn:infai:ses:characteristic:04bce19c-839d-45de-8cfb-bd470715f4cd"
               },
               {
                  "name":"instant_wind_speed_of_gust",
                  "type":"https://schema.org/Float",
                  "characteristic_id":"urn:infai:ses:characteristic:04bce19c-839d-45de-8cfb-bd470715f4cd"
               },
               {
                  "name":"12_hours_probability_of_precipitation",
                  "type":"https://schema.org/Float",
                  "characteristic_id":"urn:infai:ses:characteristic:46f808f4-bb9e-4cc2-bd50-dc33ca74f273"
               },
               {
                  "name":"1_hours_precipitation_amount",
                  "type":"https://schema.org/Float",
                  "characteristic_id":"urn:infai:ses:characteristic:7424a147-fb0e-4e71-a666-6c4997928e61"
               },
               {
                  "name":"1_hours_precipitation_amount_max",
                  "type":"https://schema.org/Float",
                  "characteristic_id":"urn:infai:ses:characteristic:7424a147-fb0e-4e71-a666-6c4997928e61"
               },
               {
                  "name":"1_hours_precipitation_amount_min",
                  "type":"https://schema.org/Float",
                  "characteristic_id":"urn:infai:ses:characteristic:7424a147-fb0e-4e71-a666-6c4997928e61"
               },
               {
                  "name":"1_hours_probability_of_precipitation",
                  "type":"https://schema.org/Float",
                  "characteristic_id":"urn:infai:ses:characteristic:46f808f4-bb9e-4cc2-bd50-dc33ca74f273"
               },
               {
                  "name":"1_hours_probability_of_thunder",
                  "type":"https://schema.org/Float",
                  "characteristic_id":"urn:infai:ses:characteristic:46f808f4-bb9e-4cc2-bd50-dc33ca74f273"
               },
               {
                  "name":"6_hours_air_temperature_max",
                  "type":"https://schema.org/Float",
                  "characteristic_id":"urn:infai:ses:characteristic:5ba31623-0ccb-4488-bfb7-f73b50e03b5a"
               },
               {
                  "name":"6_hours_air_temperature_min",
                  "type":"https://schema.org/Float",
                  "characteristic_id":"urn:infai:ses:characteristic:5ba31623-0ccb-4488-bfb7-f73b50e03b5a"
               },
               {
                  "name":"6_hours_precipitation_amount",
                  "type":"https://schema.org/Float",
                  "characteristic_id":"urn:infai:ses:characteristic:7424a147-fb0e-4e71-a666-6c4997928e61"
               },
               {
                  "name":"6_hours_precipitation_amount_max",
                  "type":"https://schema.org/Float",
                  "characteristic_id":"urn:infai:ses:characteristic:7424a147-fb0e-4e71-a666-6c4997928e61"
               },
               {
                  "name":"6_hours_precipitation_amount_min",
                  "type":"https://schema.org/Float",
                  "characteristic_id":"urn:infai:ses:characteristic:7424a147-fb0e-4e71-a666-6c4997928e61"
               },
               {
                  "name":"6_hours_probability_of_precipitation",
                  "type":"https://schema.org/Float",
                  "characteristic_id":"urn:infai:ses:characteristic:46f808f4-bb9e-4cc2-bd50-dc33ca74f273"
               },
               {
                  "name":"test",
                  "type":"https://schema.org/Float",
                  "characteristic_id":"urn:infai:ses:characteristic:5b4eea52-e8e5-4e80-9455-0382f81a1b43",
                  "function_id": "urn:infai:ses:measuring-function:bdb6a7c8-4a3d-4fe0-bab3-ce02e09b5869",
				  "aspect_id": "urn:infai:ses:aspect:a7470d73-dde3-41fc-92bd-f16bb28f2da6"
				}
            ]
         }
      ]
   },
   "function_ids":null,
   "owner":"ca4d1149-e3ed-4e0b-9e49-3bda908de436"
}`
	id := "urn:infai:ses:import-type:a93420ae-ff5f-4c44-ee6b-5d3313f946d2"
	if importTypeId == id {
		err = json.Unmarshal([]byte(str), &importType)
		if err != nil {
			return importType, err, http.StatusInternalServerError
		}
		return importType, nil, http.StatusOK
	}
	return importType, errors.New("not found"), http.StatusNotFound
}

func (this *ImportsMock) GetTopic(_ string, importId string) (topic string, err error, code int) {
	return strings.ReplaceAll(importId, ":", "_"), nil, http.StatusOK
}
