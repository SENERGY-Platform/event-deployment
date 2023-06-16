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

package metrics

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"runtime/debug"
)

type Metrics struct {
	DeployedProcesses         prometheus.Counter
	RemovedProcesses          prometheus.Counter
	DeployedConditionalEvents prometheus.Counter
	RemovedConditionalEvents  prometheus.Counter
	DeployedAnalyticsEvents   prometheus.Counter
	RemovedAnalyticsEvents    prometheus.Counter
	httphandler               http.Handler
}

func New() *Metrics {
	reg := prometheus.NewRegistry()
	m := &Metrics{
		DeployedProcesses: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "event_manager_deployed_processes",
			Help: "count of deployed processes since startup",
		}),
		DeployedConditionalEvents: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "event_manager_deployed_conditional_events",
			Help: "count of deployed conditional events since startup",
		}),
		DeployedAnalyticsEvents: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "event_manager_deployed_analytics_events",
			Help: "count of deployed analytics events since startup",
		}),
		RemovedProcesses: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "event_manager_removed_processes",
			Help: "count of removed processes since startup",
		}),
		RemovedConditionalEvents: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "event_manager_removed_conditional_events",
			Help: "count of removed conditional events since startup",
		}),
		RemovedAnalyticsEvents: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "event_manager_removed_analytics_events",
			Help: "count of removed analytics events since startup",
		}),
		httphandler: promhttp.HandlerFor(
			reg,
			promhttp.HandlerOpts{
				Registry: reg,
			},
		),
	}

	reg.MustRegister(m.DeployedProcesses)
	reg.MustRegister(m.DeployedConditionalEvents)
	reg.MustRegister(m.DeployedAnalyticsEvents)
	reg.MustRegister(m.RemovedProcesses)
	reg.MustRegister(m.RemovedConditionalEvents)
	reg.MustRegister(m.RemovedAnalyticsEvents)

	return m
}

func (this *Metrics) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	log.Printf("%v [%v] %v \n", request.RemoteAddr, request.Method, request.URL)
	this.httphandler.ServeHTTP(writer, request)
}

func (this *Metrics) Serve(ctx context.Context, port string) *Metrics {
	if port == "" || port == "-" {
		return this
	}
	router := http.NewServeMux()

	router.Handle("/metrics", this)

	server := &http.Server{Addr: ":" + port, Handler: router}
	go func() {
		log.Println("listening on ", server.Addr, "for /metrics")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			debug.PrintStack()
			log.Fatal("FATAL:", err)
		}
	}()
	go func() {
		<-ctx.Done()
		log.Println("metrics shutdown", server.Shutdown(context.Background()))
	}()
	return this
}
