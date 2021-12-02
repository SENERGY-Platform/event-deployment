/*
 * Copyright 2019 InfAI (CC SES)
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

package auth

import (
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"github.com/golang-jwt/jwt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"net/url"

	"io/ioutil"
)

type AuthToken string

func (this AuthToken) UseInRequest(req *http.Request) {
	req.Header.Set("Authorization", string(this))
}

type OpenidToken struct {
	AccessToken      string    `json:"access_token"`
	ExpiresIn        float64   `json:"expires_in"`
	RefreshExpiresIn float64   `json:"refresh_expires_in"`
	RefreshToken     string    `json:"refresh_token"`
	TokenType        string    `json:"token_type"`
	RequestTime      time.Time `json:"-"`
}

type Auth struct {
	openid *OpenidToken
	config config.Config
}

func NewAuth(config config.Config) *Auth {
	return &Auth{config: config}
}

func (this *Auth) Ensure() (token AuthToken, err error) {
	if this.openid == nil {
		this.openid = &OpenidToken{}
	}
	duration := time.Since(this.openid.RequestTime).Seconds()

	if this.openid.AccessToken != "" && this.openid.ExpiresIn-this.config.AuthExpirationTimeBuffer > duration {
		token = AuthToken("Bearer " + this.openid.AccessToken)
		return
	}

	if this.openid.RefreshToken != "" && this.openid.RefreshExpiresIn-this.config.AuthExpirationTimeBuffer > duration {
		log.Println("refresh token", this.openid.RefreshExpiresIn, duration)
		err = refreshOpenidToken(this.openid, this.config)
		if err != nil {
			log.Println("WARNING: unable to use refreshtoken", err)
		} else {
			token = AuthToken("Bearer " + this.openid.AccessToken)
			return
		}
	}

	log.Println("get new access token")
	err = getOpenidToken(this.openid, this.config)
	if err != nil {
		log.Println("ERROR: unable to get new access token", err)
		this.openid = &OpenidToken{}
	}
	token = AuthToken("Bearer " + this.openid.AccessToken)
	return
}

func (this *Auth) GetUserToken(userid string) (token AuthToken, err error) {
	resp, err := http.PostForm(this.config.AuthEndpoint+"/auth/realms/master/protocol/openid-connect/token", url.Values{
		"client_id":         {this.config.AuthClientId},
		"client_secret":     {this.config.AuthClientSecret},
		"grant_type":        {"urn:ietf:params:oauth:grant-type:token-exchange"},
		"requested_subject": {userid},
	})
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println("ERROR: GetUserToken()", resp.StatusCode, string(body))
		err = errors.New("access denied")
		resp.Body.Close()
		return
	}
	var openIdToken OpenidToken
	err = json.NewDecoder(resp.Body).Decode(&openIdToken)
	if err != nil {
		return
	}
	return AuthToken("Bearer " + openIdToken.AccessToken), nil
}

func (this *AuthToken) GetJSON(url string, result interface{}) (err error) {
	resp, err := this.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("unexpected status code " + strconv.Itoa(resp.StatusCode))
	}
	return json.NewDecoder(resp.Body).Decode(&result)
}

func (this *AuthToken) Get(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", string(*this))
	resp, err = http.DefaultClient.Do(req)
	return
}

func getOpenidToken(token *OpenidToken, config config.Config) (err error) {
	requesttime := time.Now()
	resp, err := http.PostForm(config.AuthEndpoint+"/auth/realms/master/protocol/openid-connect/token", url.Values{
		"client_id":     {config.AuthClientId},
		"client_secret": {config.AuthClientSecret},
		"grant_type":    {"client_credentials"},
	})

	if err != nil {
		log.Println("ERROR: getOpenidToken::PostForm()", err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println("ERROR: getOpenidToken()", resp.StatusCode, string(body))
		err = errors.New("access denied")
		resp.Body.Close()
		return
	}
	err = json.NewDecoder(resp.Body).Decode(token)
	token.RequestTime = requesttime
	return
}

func refreshOpenidToken(token *OpenidToken, config config.Config) (err error) {
	requesttime := time.Now()
	resp, err := http.PostForm(config.AuthEndpoint+"/auth/realms/master/protocol/openid-connect/token", url.Values{
		"client_id":     {config.AuthClientId},
		"client_secret": {config.AuthClientSecret},
		"refresh_token": {token.RefreshToken},
		"grant_type":    {"refresh_token"},
	})

	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println("ERROR: refreshOpenidToken()", resp.StatusCode, string(body))
		err = errors.New("access denied")
		resp.Body.Close()
		return
	}
	err = json.NewDecoder(resp.Body).Decode(token)
	token.RequestTime = requesttime
	return
}

func (this *Auth) GenerateInternalUserToken(userid string) (token AuthToken, err error) {
	claims := jwt.StandardClaims{
		ExpiresAt: time.Time{}.Unix(),
		Issuer:    "internal",
		Subject:   userid,
	}

	jwtoken := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	unsignedTokenString, err := jwtoken.SigningString()
	if err != nil {
		log.Println("ERROR: GenerateUserTokenById::SigningString()", err, userid)
		return token, err
	}
	tokenString := strings.Join([]string{unsignedTokenString, ""}, ".")
	token = AuthToken("Bearer " + tokenString)
	return token, err
}
