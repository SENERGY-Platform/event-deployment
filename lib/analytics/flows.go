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
	"io"
	"log"
	"net/http"
	"net/url"
	"runtime/debug"
)

func (this *Analytics) GetFlowInputs(id string, user string) (result []FlowModelCell, err error, code int) {
	client := http.Client{
		Timeout: this.timeout,
	}
	req, err := http.NewRequest(
		"GET",
		this.config.FlowParserUrl+"/flow/getinputs/"+url.PathEscape(id),
		nil,
	)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("X-UserId", user)
	resp, err := client.Do(req)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		debug.PrintStack()
		return result, errors.New("unexpected statuscode"), resp.StatusCode
	}

	temp, err := io.ReadAll(resp.Body)
	err = json.Unmarshal(temp, &result)
	if err != nil {
		log.Println("ERROR:", err, string(temp))
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}
	return result, err, http.StatusOK
}
