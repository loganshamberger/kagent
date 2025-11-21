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

package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/kagent-dev/kagent/go/api/v1alpha1"
	"github.com/kagent-dev/kagent/go/pkg/registry/api"
)

func TestServer_ListCards(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)

	card1 := &v1alpha1.AgentCard{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "agent-1",
			Namespace: "default",
		},
		Spec: v1alpha1.AgentCardSpec{
			Name:         "agent-1",
			Capabilities: []string{"kubernetes"},
		},
	}

	card2 := &v1alpha1.AgentCard{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "agent-2",
			Namespace: "kagent",
		},
		Spec: v1alpha1.AgentCardSpec{
			Name:         "agent-2",
			Capabilities: []string{"monitoring"},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(card1, card2).Build()
	server := api.NewServer(client, 8084)

	req := httptest.NewRequest("GET", "/api/v1alpha1/registry/cards", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var cardList v1alpha1.AgentCardList
	err := json.Unmarshal(w.Body.Bytes(), &cardList)
	require.NoError(t, err)
	assert.Len(t, cardList.Items, 2)
}

func TestServer_GetCard(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)

	card := &v1alpha1.AgentCard{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-agent",
			Namespace: "kagent",
		},
		Spec: v1alpha1.AgentCardSpec{
			Name:       "test-agent",
			Version:    "1.0.0",
			PublicCard: `{"name":"test-agent","version":"1.0.0"}`,
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(card).Build()
	server := api.NewServer(client, 8084)

	req := httptest.NewRequest("GET", "/api/v1alpha1/registry/cards/kagent/test-agent", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result v1alpha1.AgentCard
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "test-agent", result.Spec.Name)
}

func TestServer_GetCard_NotFound(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)

	client := fake.NewClientBuilder().WithScheme(scheme).Build()
	server := api.NewServer(client, 8084)

	req := httptest.NewRequest("GET", "/api/v1alpha1/registry/cards/default/nonexistent", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestServer_GetA2ACard(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)

	a2aJSON := `{"name":"test-agent","version":"1.0.0","description":"Test agent"}`
	card := &v1alpha1.AgentCard{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-agent",
			Namespace: "kagent",
		},
		Spec: v1alpha1.AgentCardSpec{
			Name:       "test-agent",
			PublicCard: a2aJSON,
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(card).Build()
	server := api.NewServer(client, 8084)

	req := httptest.NewRequest("GET", "/api/v1alpha1/registry/cards/kagent/test-agent/a2a", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.JSONEq(t, a2aJSON, w.Body.String())
}

func TestServer_GetA2ACard_NoPublicCard(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)

	card := &v1alpha1.AgentCard{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-agent",
			Namespace: "kagent",
		},
		Spec: v1alpha1.AgentCardSpec{
			Name: "test-agent",
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(card).Build()
	server := api.NewServer(client, 8084)

	req := httptest.NewRequest("GET", "/api/v1alpha1/registry/cards/kagent/test-agent/a2a", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
