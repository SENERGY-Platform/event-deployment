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

package events

import (
	"errors"
	"github.com/golang-jwt/jwt"
	"strings"
)

type Claims struct {
	Sub         string              `json:"sub,omitempty"`
	RealmAccess map[string][]string `json:"realm_access,omitempty"`
}

func (this *Claims) Valid() error {
	if this.Sub == "" {
		return errors.New("missing subject")
	}
	return nil
}

func parse(token string) (claims Claims, err error) {
	if strings.HasPrefix(token, "Bearer ") {
		token = token[7:]
	}
	_, _, err = new(jwt.Parser).ParseUnverified(token, &claims)
	return
}

func IsAdmin(token string) bool {
	claims, err := parse(token)
	if err != nil {
		panic(err)
	}
	return contains(claims.RealmAccess["roles"], "admin")
}

func GetUserId(token string) (userId string, err error) {
	claims, err := parse(token)
	if err != nil {
		return userId, err
	}
	return claims.Sub, err
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
