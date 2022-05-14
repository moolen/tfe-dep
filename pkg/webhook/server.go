/*
Copyright 2020 The Flux authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package webhook

import (
	"context"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	prommetrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"
)

// ReceiverServer handles webhook POST requests
type ReceiverServer struct {
	port   string
	logger log.FieldLogger
}

// NewEventServer returns an HTTP server that handles webhooks
func NewReceiverServer(port string) *ReceiverServer {
	return &ReceiverServer{
		port:   port,
		logger: log.New().WithField("component", "webhook"),
	}
}

// ListenAndServe starts the HTTP server on the specified port
func (s *ReceiverServer) ListenAndServe(stopCh <-chan struct{}) {
	mux := http.DefaultServeMux
	mux.Handle("/webhook/", http.HandlerFunc(s.handlePayload()))
	mdlw := middleware.New(middleware.Config{
		Recorder: prommetrics.NewRecorder(prommetrics.Config{
			Prefix: "tdep_webhook",
		}),
	})

	h := std.Handler("", mdlw, mux)
	srv := &http.Server{
		Addr:    s.port,
		Handler: h,
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			s.logger.Error(err, "Receiver server crashed")
			os.Exit(1)
		}
	}()

	<-stopCh
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		s.logger.Error(err, "Receiver server graceful shutdown failed")
	} else {
		s.logger.Info("Receiver server stopped")
	}
}
