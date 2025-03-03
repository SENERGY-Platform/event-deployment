package mocks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/event-deployment/lib/auth"
	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"log"
	"net/http"
	"net/http/httptest"
	"runtime/debug"
	"time"
)

func MockAuthServer(config config.Config, ctx context.Context) (err error) {
	router, err := getRouter()
	if err != nil {
		return err
	}
	server := httptest.NewServer(router)
	config.AuthEndpoint = server.URL
	go func() {
		<-ctx.Done()
		server.Close()
	}()
	return nil
}

const testAdminToken = `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjEwMDAwMDAwMDAsImlhdCI6MTAwMDAwMDAwMCwiYXV0aF90aW1lIjoxMDAwMDAwMDAwLCJpc3MiOiJpbnRlcm5hbCIsImF1ZCI6W10sInN1YiI6ImRkNjllYTBkLWY1NTMtNDMzNi04MGYzLTdmNDU2N2Y4NWM3YiIsInR5cCI6IkJlYXJlciIsImF6cCI6ImZyb250ZW5kIiwicmVhbG1fYWNjZXNzIjp7InJvbGVzIjpbImFkbWluIiwiZGV2ZWxvcGVyIiwidXNlciJdfSwicmVzb3VyY2VfYWNjZXNzIjp7Im1hc3Rlci1yZWFsbSI6eyJyb2xlcyI6W119LCJCYWNrZW5kLXJlYWxtIjp7InJvbGVzIjpbXX0sImFjY291bnQiOnsicm9sZXMiOltdfX0sInJvbGVzIjpbImFkbWluIiwiZGV2ZWxvcGVyIiwidXNlciJdLCJuYW1lIjoiU2VwbCBBZG1pbiIsInByZWZlcnJlZF91c2VybmFtZSI6InNlcGwiLCJnaXZlbl9uYW1lIjoiU2VwbCIsImxvY2FsZSI6ImVuIiwiZmFtaWx5X25hbWUiOiJBZG1pbiIsImVtYWlsIjoic2VwbEBzZXBsLmRlIn0.HZyG6n-BfpnaPAmcDoSEh0SadxUx-w4sEt2RVlQ9e5I`

func getRouter() (handler http.HandlerFunc, err error) {
	defer func() {
		if r := recover(); r != nil && err == nil {
			log.Printf("%s: %s", r, debug.Stack())
			err = errors.New(fmt.Sprint("Recovered Error: ", r))
		}
	}()
	return func(writer http.ResponseWriter, request *http.Request) {
		tokenpath := "auth/realms/master/protocol/openid-connect/token"
		if request.URL.Path == tokenpath || request.URL.Path == "/"+tokenpath || request.URL.RawPath == tokenpath || request.URL.RawPath == "/"+tokenpath {
			userid := request.FormValue("requested_subject")
			grandType := request.FormValue("grant_type")
			var token string
			if grandType == "client_credentials" {
				//admin token
				token = testAdminToken
			} else {
				//user token
				token, err = auth.GenerateInternalUserToken(userid)
				if err != nil {
					http.Error(writer, err.Error(), 500)
					return
				}
			}
			writer.Header().Set("Content-Type", "application/json; charset=utf-8")
			err = json.NewEncoder(writer).Encode(map[string]interface{}{
				"access_token":       token[7:],
				"expires_in":         10 * float64(time.Hour),
				"refresh_expires_in": 5 * float64(time.Hour),
				"refresh_token":      "",
				"token_type":         "",
			})
			if err != nil {
				log.Println("ERROR: unable to encode response", err)
			}
			return
		}

		http.Error(writer, "mock does not implement given path:"+request.URL.Path, 500)

	}, err
}
