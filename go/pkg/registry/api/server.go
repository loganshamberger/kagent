/*
Copyright 2025.

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

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kagent-dev/kagent/go/api/v1alpha1"
)

type Server struct {
	client client.Client
	router *mux.Router
	port   int
	server *http.Server
}

func NewServer(c client.Client, port int) *Server {
	s := &Server{
		client: c,
		port:   port,
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.router = mux.NewRouter()
	api := s.router.PathPrefix("/api/v1alpha1/registry").Subrouter()

	api.HandleFunc("/cards", s.listCards).Methods("GET")
	api.HandleFunc("/cards/{namespace}/{name}", s.getCard).Methods("GET")
	api.HandleFunc("/cards/{namespace}/{name}/a2a", s.getA2ACard).Methods("GET")
}

func (s *Server) listCards(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cardList := &v1alpha1.AgentCardList{}
	if err := s.client.List(ctx, cardList); err != nil {
		http.Error(w, fmt.Sprintf("failed to list agent cards: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(cardList); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

func (s *Server) getCard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	name := vars["name"]

	card := &v1alpha1.AgentCard{}
	key := client.ObjectKey{Namespace: namespace, Name: name}
	if err := s.client.Get(ctx, key, card); err != nil {
		http.Error(w, fmt.Sprintf("agent card not found: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(card); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

func (s *Server) getA2ACard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	name := vars["name"]

	card := &v1alpha1.AgentCard{}
	key := client.ObjectKey{Namespace: namespace, Name: name}
	if err := s.client.Get(ctx, key, card); err != nil {
		http.Error(w, fmt.Sprintf("agent card not found: %v", err), http.StatusNotFound)
		return
	}

	if card.Spec.PublicCard == "" {
		http.Error(w, "A2A card not available", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(card.Spec.PublicCard))
}

func (s *Server) Start(ctx context.Context) error {
	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.server.Shutdown(shutdownCtx)
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
