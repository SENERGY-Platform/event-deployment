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
	"github.com/SENERGY-Platform/service-commons/pkg/cache"
	"github.com/golang-jwt/jwt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"net/url"
)

const CacheExpiration = 600 * time.Second

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
	openid         *OpenidToken
	config         config.Config
	userTokenCache *cache.Cache
}

func NewAuth(config config.Config) (*Auth, error) {
	c, err := cache.New(cache.Config{})
	if err != nil {
		return nil, err
	}
	return &Auth{config: config, userTokenCache: c}, nil
}

func NewAuthWithoutCache(config config.Config) *Auth {
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

var ErrUserDoesNotExist = errors.New("user does not exist")

func (this *Auth) GetUserToken(userid string) (token AuthToken, err error) {
	return cache.Use(this.userTokenCache, "user_token."+userid, func() (AuthToken, error) {
		return this.getUserTokenCheckExistence(userid)
	}, func(token AuthToken) error {
		return nil
	}, time.Duration(this.config.UserTokenCacheLifespanInSec)*time.Second)
}

func (this *Auth) getUserTokenCheckExistence(userid string) (token AuthToken, err error) {
	token, err = this.getUserToken(userid)
	if err != nil {
		exists, err := this.UserExists(userid)
		if err != nil {
			return token, err
		}
		if !exists {
			return token, ErrUserDoesNotExist
		}
	}
	return token, err
}

func (this *Auth) getUserToken(userid string) (token AuthToken, err error) {
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
		body, _ := io.ReadAll(resp.Body)
		log.Println("ERROR: GetUserToken()", userid, resp.StatusCode, string(body))
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
		body, _ := io.ReadAll(resp.Body)
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
		body, _ := io.ReadAll(resp.Body)
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
	temp, err := GenerateInternalUserToken(userid)
	return AuthToken(temp), err
}

func GenerateInternalUserToken(userid string) (token string, err error) {
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
	token = "Bearer " + tokenString
	return token, err
}

type User struct {
	Id         string                 `json:"id"`
	Name       string                 `json:"username"`
	Attributes map[string]interface{} `json:"attributes"`
	//Enabled    bool                   `json:"enabled"`
	//FirstName  string                 `json:"firstName"`
	//LastName   string                 `json:"lastName"`
}

func (this *Auth) GetUserById(id string) (user User, err error) {
	token, err := this.Ensure()
	if err != nil {
		return user, err
	}
	err = token.GetJSON(this.config.AuthEndpoint+"/auth/admin/realms/master/users/"+url.QueryEscape(id), &user)
	return
}

func (this *Auth) UserExists(id string) (exists bool, err error) {
	token, err := this.Ensure()
	if err != nil {
		return false, err
	}
	resp, err := token.Get(this.config.AuthEndpoint + "/auth/admin/realms/master/users/" + url.QueryEscape(id))
	if err != nil {
		return false, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, errors.New(string(body))
	}
	return true, nil
}
