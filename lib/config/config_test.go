/*
 * Copyright 2026 InfAI (CC SES)
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
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type mongoFields struct {
	Url, User, Password, AuthSource, Database, DescCollection, DeploymentsCollection string
}

func mongoOf(c Config) mongoFields {
	return mongoFields{c.MongoUrl, c.MongoUser, c.MongoPassword, c.MongoAuthSource, c.MongoDatabase, c.ConditionalEventRepoMongoDescCollection, c.ConditionalEventRepoMongoDeploymentsCollection}
}

// clearMongoEnv empties the variables for this test; the loader ignores empty values.
func clearMongoEnv(t *testing.T) {
	for _, k := range []string{"MONGO_URL", "MONGO_USER", "MONGO_PASSWORD", "MONGO_AUTH_SOURCE", "MONGO_DATABASE", "CONDITIONAL_EVENT_REPO_MONGO_DESC_COLLECTION", "CONDITIONAL_EVENT_REPO_MONGO_DEPLOYMENTS_COLLECTION"} {
		t.Setenv(k, "")
	}
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLoad_RepoConfigMongoDefaults(t *testing.T) {
	clearMongoEnv(t)
	cfg, err := LoadConfig("../../config.json")
	if err != nil {
		t.Fatal(err)
	}
	want := mongoFields{Url: "", AuthSource: "admin", Database: "event_descriptions", DescCollection: "event_descriptions", DeploymentsCollection: "deployments"}
	if got := mongoOf(cfg); got != want {
		t.Errorf("mongo = %+v, want %+v", got, want)
	}
}

func TestLoad_MongoDefaultsWhenFileOmitsThem(t *testing.T) {
	clearMongoEnv(t)
	cfg, err := LoadConfig(writeConfig(t, `{}`))
	if err != nil {
		t.Fatal(err)
	}
	want := mongoFields{Url: "mongodb://localhost:27017", AuthSource: "admin", Database: "event_descriptions"}
	if got := mongoOf(cfg); got != want {
		t.Errorf("mongo = %+v, want %+v", got, want)
	}
}

func TestLoad_MongoEnvNames(t *testing.T) {
	clearMongoEnv(t)
	t.Setenv("MONGO_URL", "mongodb://mongo-0.mongo:27017,mongo-1.mongo:27017/?replicaSet=rs0")
	t.Setenv("MONGO_USER", "event-deployment")
	t.Setenv("MONGO_PASSWORD", "p@ss:w/rd")
	t.Setenv("MONGO_AUTH_SOURCE", "users")
	t.Setenv("MONGO_DATABASE", "event_descriptions_test")
	cfg, err := LoadConfig("../../config.json")
	if err != nil {
		t.Fatal(err)
	}
	want := mongoFields{
		Url:                   "mongodb://mongo-0.mongo:27017,mongo-1.mongo:27017/?replicaSet=rs0",
		User:                  "event-deployment",
		Password:              "p@ss:w/rd",
		AuthSource:            "users",
		Database:              "event_descriptions_test",
		DescCollection:        "event_descriptions",
		DeploymentsCollection: "deployments",
	}
	if got := mongoOf(cfg); got != want {
		t.Errorf("mongo = %+v, want %+v", got, want)
	}
}

func TestLoad_MongoConfigFile(t *testing.T) {
	clearMongoEnv(t)
	cfg, err := LoadConfig(writeConfig(t, `{"mongo_url": "mongodb://file:27017", "mongo_user": "u", "mongo_password": "s3cr3t", "mongo_auth_source": "a", "mongo_database": "d"}`))
	if err != nil {
		t.Fatal(err)
	}
	want := mongoFields{Url: "mongodb://file:27017", User: "u", Password: "s3cr3t", AuthSource: "a", Database: "d"}
	if got := mongoOf(cfg); got != want {
		t.Errorf("mongo = %+v, want %+v", got, want)
	}
}

// The loader prints every environment variable it applies.
func TestLoad_EnvPrintMasksMongoPassword(t *testing.T) {
	clearMongoEnv(t)
	t.Setenv("MONGO_USER", "event-deployment")
	t.Setenv("MONGO_PASSWORD", "s3cr3t-pw")
	out := captureStdout(t, func() {
		if _, err := LoadConfig("../../config.json"); err != nil {
			t.Error(err)
		}
	})
	if strings.Contains(out, "s3cr3t-pw") {
		t.Errorf("printed environment leaks the password: %s", out)
	}
	if !strings.Contains(out, "MONGO_USER  =  event-deployment") {
		t.Errorf("expected the applied variables to be printed, got: %s", out)
	}
}

func TestConfigFormattingMasksMongoPassword(t *testing.T) {
	clearMongoEnv(t)
	t.Setenv("MONGO_PASSWORD", "s3cr3t-pw")
	var cfg Config
	captureStdout(t, func() {
		var err error
		if cfg, err = LoadConfig("../../config.json"); err != nil {
			t.Error(err)
		}
	})
	b, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	outputs := map[string]string{
		"json": string(b),
		"%v":   fmt.Sprintf("%v", cfg),
		"%+v":  fmt.Sprintf("%+v", *cfg),
		"%#v":  fmt.Sprintf("%#v", *cfg),
		"%s":   fmt.Sprintf("%s", cfg),
	}
	for name, s := range outputs {
		if strings.Contains(s, "s3cr3t-pw") {
			t.Errorf("%s leaks the password: %s", name, s)
		}
		if !strings.Contains(s, "event_descriptions") {
			t.Errorf("%s lost the other fields: %s", name, s)
		}
	}
	if !strings.Contains(string(b), `"mongo_password":"***"`) {
		t.Errorf("json does not show the password as masked: %s", b)
	}
	if cfg.MongoPassword != "s3cr3t-pw" {
		t.Errorf("masking changed the loaded password to %q", cfg.MongoPassword)
	}
}

func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w
	done := make(chan string)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()
	defer func() { os.Stdout = orig }()
	f()
	_ = w.Close()
	return <-done
}
