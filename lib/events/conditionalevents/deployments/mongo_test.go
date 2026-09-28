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

package deployments

import (
	"context"
	"errors"
	"net"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/SENERGY-Platform/event-deployment/lib/config"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const replicaSetURL = "mongodb://mongo-0.mongo:27017,mongo-1.mongo:27017/?replicaSet=rs0&readPreference=primary"

func TestClientOptions_AuthWhenUserGiven(t *testing.T) {
	opts := clientOptions(&config.ConfigStruct{
		MongoUrl:        replicaSetURL,
		MongoUser:       "event-deployment",
		MongoPassword:   "s3cr3t",
		MongoAuthSource: "admin",
		MongoDatabase:   "event_descriptions",
	})
	if err := opts.Validate(); err != nil {
		t.Fatal(err)
	}
	want := &options.Credential{Username: "event-deployment", Password: "s3cr3t", AuthSource: "admin"}
	if !reflect.DeepEqual(opts.Auth, want) {
		t.Errorf("auth = %+v, want %+v", opts.Auth, want)
	}
}

func TestClientOptions_NoAuthWhenUserEmpty(t *testing.T) {
	// A password without a user must not switch auth on.
	opts := clientOptions(&config.ConfigStruct{
		MongoUrl:        "mongodb://localhost:27017",
		MongoPassword:   "s3cr3t",
		MongoAuthSource: "admin",
		MongoDatabase:   "event_descriptions",
	})
	if err := opts.Validate(); err != nil {
		t.Fatal(err)
	}
	if opts.Auth != nil {
		t.Errorf("auth = %+v, want nil", opts.Auth)
	}
}

func TestClientOptions_ConfiguredCredentialsReplaceURICredentials(t *testing.T) {
	opts := clientOptions(&config.ConfigStruct{
		MongoUrl:        "mongodb://old:oldpw@localhost:27017/?authSource=other&authMechanism=SCRAM-SHA-1",
		MongoUser:       "event-deployment",
		MongoPassword:   "newpw",
		MongoAuthSource: "admin",
		MongoDatabase:   "event_descriptions",
	})
	if err := opts.Validate(); err != nil {
		t.Fatal(err)
	}
	want := &options.Credential{Username: "event-deployment", Password: "newpw", AuthSource: "admin"}
	if !reflect.DeepEqual(opts.Auth, want) {
		t.Errorf("auth = %+v, want %+v", opts.Auth, want)
	}
}

func TestClientOptions_URIPassedUnchanged(t *testing.T) {
	opts := clientOptions(&config.ConfigStruct{MongoUrl: replicaSetURL, MongoDatabase: "event_descriptions"})
	if err := opts.Validate(); err != nil {
		t.Fatal(err)
	}
	if got := opts.GetURI(); got != replicaSetURL {
		t.Errorf("uri = %q, want %q", got, replicaSetURL)
	}
	if want := []string{"mongo-0.mongo:27017", "mongo-1.mongo:27017"}; !reflect.DeepEqual(opts.Hosts, want) {
		t.Errorf("hosts = %v, want %v", opts.Hosts, want)
	}
	if opts.ReplicaSet == nil || *opts.ReplicaSet != "rs0" {
		t.Errorf("replica set = %v, want rs0", opts.ReplicaSet)
	}
}

func TestClientOptions_NoSchemeAdded(t *testing.T) {
	opts := clientOptions(&config.ConfigStruct{MongoUrl: "localhost:27017", MongoDatabase: "event_descriptions"})
	if err := opts.Validate(); err == nil {
		t.Fatal("expected an error for a url without scheme")
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.ConfigStruct
		wantErr error
	}{
		{"no auth", config.ConfigStruct{MongoDatabase: "event_descriptions"}, nil},
		{"user and password", config.ConfigStruct{MongoDatabase: "event_descriptions", MongoUser: "u", MongoPassword: "p"}, nil},
		{"password without user", config.ConfigStruct{MongoDatabase: "event_descriptions", MongoPassword: "p"}, nil},
		{"user without password", config.ConfigStruct{MongoDatabase: "event_descriptions", MongoUser: "u"}, errMissingPassword},
		{"empty database", config.ConfigStruct{}, errEmptyDatabase},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateConfig(&tt.cfg); !errors.Is(err, tt.wantErr) {
				t.Errorf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeploymentsCollectionUsesConfiguredDatabase(t *testing.T) {
	// mongo.Connect does not contact the server, so no running instance is needed.
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })
	db := &Deployments{config: &config.ConfigStruct{MongoDatabase: "custom_db", ConditionalEventRepoMongoDeploymentsCollection: "deployments"}, client: client}
	coll := db.deploymentsCollection()
	if coll.Database().Name() != "custom_db" || coll.Name() != "deployments" {
		t.Errorf("collection = %s.%s, want custom_db.deployments", coll.Database().Name(), coll.Name())
	}
}

// unreachableURL points at a port that was just free, so only the startup check can fail.
func unreachableURL(t *testing.T) string {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	if err = l.Close(); err != nil {
		t.Fatal(err)
	}
	return "mongodb://" + addr + "/?directConnection=true"
}

// The startup check would fail as well, so these check for the specific validation error.
func TestNew_RejectsBeforeConnecting(t *testing.T) {
	cases := map[string]config.ConfigStruct{
		"empty database":        {MongoUser: "event-deployment", MongoPassword: "s3cr3t"},
		"user without password": {MongoUser: "event-deployment", MongoDatabase: "event_descriptions"},
	}
	want := map[string]error{"empty database": errEmptyDatabase, "user without password": errMissingPassword}
	for name, cfg := range cases {
		t.Run(name, func(t *testing.T) {
			cfg.MongoUrl = unreachableURL(t)
			wg := &sync.WaitGroup{}
			db, err := New(context.Background(), wg, &cfg)
			if !errors.Is(err, want[name]) {
				t.Fatalf("err = %v, want %v", err, want[name])
			}
			if db != nil {
				t.Error("expected no db on failure")
			}
			if strings.Contains(err.Error(), "s3cr3t") {
				t.Errorf("error leaks the password: %v", err)
			}
		})
	}
}

// poolCounter counts connection pools; Disconnect closes every pool Connect created.
type poolCounter struct{ created, closed atomic.Int32 }

func (p *poolCounter) monitor() *event.PoolMonitor {
	return &event.PoolMonitor{Event: func(e *event.PoolEvent) {
		switch e.Type {
		case event.PoolCreated:
			p.created.Add(1)
		case event.PoolClosedEvent:
			p.closed.Add(1)
		}
	}}
}

func (p *poolCounter) assertAllClosed(t *testing.T) {
	t.Helper()
	created, closed := p.created.Load(), p.closed.Load()
	if created == 0 || closed != created {
		t.Errorf("%d of %d connection pools closed, the client was left connected", closed, created)
	}
}

func assertNoWaiters(t *testing.T, wg *sync.WaitGroup) {
	t.Helper()
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("a failed startup left a shutdown routine registered")
	}
}

func TestStart_StartupCheckFailsWithoutServer(t *testing.T) {
	const password = "pw-must-not-appear-7f3a"
	conf := &config.ConfigStruct{
		MongoUrl:        unreachableURL(t),
		MongoUser:       "event-deployment",
		MongoPassword:   password,
		MongoAuthSource: "admin",
		MongoDatabase:   "event_descriptions",
		ConditionalEventRepoMongoDeploymentsCollection: "deployments",
	}
	pools := &poolCounter{}
	wg := &sync.WaitGroup{}
	begin := time.Now()
	db, err := start(context.Background(), wg, conf, clientOptions(conf).SetPoolMonitor(pools.monitor()), 500*time.Millisecond)
	if err == nil {
		t.Fatal("expected an error when the server is unreachable")
	}
	if db != nil {
		t.Error("expected no db on failure")
	}
	if !strings.HasPrefix(err.Error(), "mongo startup check failed: ") {
		t.Errorf("unexpected error: %v", err)
	}
	if strings.Contains(err.Error(), password) {
		t.Error("error text contains the password")
	}
	if elapsed := time.Since(begin); elapsed > 5*time.Second {
		t.Errorf("start took %v, the timeout was not applied", elapsed)
	}
	pools.assertAllClosed(t)
	assertNoWaiters(t, wg)
}
