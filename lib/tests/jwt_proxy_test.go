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

package tests

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"runtime/debug"
	"testing"
	"time"
)

func createAnalyticsProxyServer(t *testing.T, ctx context.Context, config config.Config) (resultConfig config.Config, err error) {
	t.Skip("needs al local test.config.json \n will contact senergy test platform analytics")

	testConfig := TestConfig{}

	file, err := os.Open("test.config.json")
	if err != nil {
		return resultConfig, err
	}
	err = json.NewDecoder(file).Decode(&testConfig)
	if err != nil {
		return resultConfig, err
		return
	}

	token, err := testConfig.Access()
	if err != nil {
		return resultConfig, err
		return
	}

	resultConfig = config
	resultConfig.FlowEngineUrl = createJwtProxyServer(ctx, token, config.FlowEngineUrl)
	resultConfig.FlowParserUrl = createJwtProxyServer(ctx, token, config.FlowParserUrl)
	resultConfig.PipelineRepoUrl = createJwtProxyServer(ctx, token, config.PipelineRepoUrl)
	return
}

func createJwtProxyServer(ctx context.Context, token string, to string) (url string) {
	server := httptest.NewServer(&JwtProxy{token: token, to: to})
	url = server.URL
	go func() {
		<-ctx.Done()
		server.Close()
	}()
	return
}

type JwtProxy struct {
	token string
	to    string
}

func (this *JwtProxy) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	url := this.to + req.URL.Path
	log.Println("DEBUG: call", req.Method, url)
	remoteReq, err := http.NewRequest(
		req.Method,
		url,
		req.Body,
	)
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return
	}
	remoteReq.Header.Set("Authorization", this.token)
	remoteResp, err := client.Do(remoteReq)
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return
	}
	defer remoteResp.Body.Close()
	resp.WriteHeader(remoteResp.StatusCode)
	temp, err := ioutil.ReadAll(remoteResp.Body)
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return
	}
	resp.Write(temp)
	return
}

type TestConfig struct {
	User         string `json:"user"`
	Pw           string `json:"pw"`
	AuthEndpoint string `json:"auth_endpoint"`
	AuthClient   string `json:"auth_client"`
}

type OpenidToken struct {
	AccessToken string `json:"access_token"`
}

func (config *TestConfig) Access() (token string, err error) {
	values := url.Values{
		"client_id":  {config.AuthClient},
		"username":   {config.User},
		"password":   {config.Pw},
		"grant_type": {"password"},
	}
	resp, err := http.PostForm(config.AuthEndpoint+"/auth/realms/master/protocol/openid-connect/token", values)

	if err != nil {
		log.Println("ERROR: getOpenidToken::PostForm()", err)
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println("ERROR: getOpenidToken()", resp.StatusCode, string(body))
		err = errors.New("access denied")
		resp.Body.Close()
		return
	}
	oidToken := &OpenidToken{}
	err = json.NewDecoder(resp.Body).Decode(oidToken)
	return oidToken.AccessToken, err
}
