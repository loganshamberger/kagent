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

package health_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kagent-dev/kagent/go/api/v1alpha1"
	"github.com/kagent-dev/kagent/go/pkg/registry/health"
)

func TestChecker_CheckEndpoints_Healthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	checker := health.NewChecker(5 * time.Second)
	endpoints := []v1alpha1.AgentEndpoint{
		{URL: server.URL, Protocol: "http"},
	}

	healthy, err := checker.CheckEndpoints(context.Background(), endpoints)
	require.NoError(t, err)
	assert.True(t, healthy)
}

func TestChecker_CheckEndpoints_Unhealthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	checker := health.NewChecker(5 * time.Second)
	endpoints := []v1alpha1.AgentEndpoint{
		{URL: server.URL, Protocol: "http"},
	}

	healthy, err := checker.CheckEndpoints(context.Background(), endpoints)
	require.Error(t, err)
	assert.False(t, healthy)
}

func TestChecker_CheckEndpoints_MultipleEndpoints(t *testing.T) {
	healthyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer healthyServer.Close()

	checker := health.NewChecker(5 * time.Second)
	endpoints := []v1alpha1.AgentEndpoint{
		{URL: "http://invalid-host-that-does-not-exist:9999", Protocol: "http"},
		{URL: healthyServer.URL, Protocol: "http"},
	}

	healthy, err := checker.CheckEndpoints(context.Background(), endpoints)
	require.NoError(t, err)
	assert.True(t, healthy)
}

func TestChecker_CheckEndpoints_AllUnreachable(t *testing.T) {
	checker := health.NewChecker(1 * time.Second)
	endpoints := []v1alpha1.AgentEndpoint{
		{URL: "http://invalid-host-1:9999", Protocol: "http"},
		{URL: "http://invalid-host-2:9999", Protocol: "http"},
	}

	healthy, err := checker.CheckEndpoints(context.Background(), endpoints)
	require.Error(t, err)
	assert.False(t, healthy)
	assert.Contains(t, err.Error(), "all endpoints unhealthy")
}

func TestChecker_CheckEndpoints_NoEndpoints(t *testing.T) {
	checker := health.NewChecker(5 * time.Second)
	endpoints := []v1alpha1.AgentEndpoint{}

	healthy, err := checker.CheckEndpoints(context.Background(), endpoints)
	require.Error(t, err)
	assert.False(t, healthy)
	assert.Contains(t, err.Error(), "no endpoints to check")
}

func TestChecker_CheckEndpoints_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	checker := health.NewChecker(100 * time.Millisecond)
	endpoints := []v1alpha1.AgentEndpoint{
		{URL: server.URL, Protocol: "http"},
	}

	healthy, err := checker.CheckEndpoints(context.Background(), endpoints)
	require.Error(t, err)
	assert.False(t, healthy)
}

func TestChecker_CheckEndpoints_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	checker := health.NewChecker(5 * time.Second)
	endpoints := []v1alpha1.AgentEndpoint{
		{URL: server.URL, Protocol: "http"},
	}

	healthy, err := checker.CheckEndpoints(ctx, endpoints)
	require.Error(t, err)
	assert.False(t, healthy)
}
