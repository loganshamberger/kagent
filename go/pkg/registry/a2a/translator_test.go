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

package a2a_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"trpc.group/trpc-go/trpc-a2a-go/server"

	"github.com/kagent-dev/kagent/go/api/v1alpha1"
	"github.com/kagent-dev/kagent/go/pkg/registry/a2a"
)

func TestTranslator_ToA2ACard(t *testing.T) {
	translator := a2a.NewTranslator()

	tests := []struct {
		name    string
		input   *v1alpha1.AgentCard
		wantErr bool
		check   func(t *testing.T, card *server.AgentCard)
	}{
		{
			name:    "nil card",
			input:   nil,
			wantErr: true,
		},
		{
			name: "minimal card",
			input: &v1alpha1.AgentCard{
				Spec: v1alpha1.AgentCardSpec{
					Name:       "test-agent",
					Version:    "1.0.0",
					A2AVersion: "0.3.0",
					Endpoints: []v1alpha1.AgentEndpoint{
						{URL: "http://test.example.com", Protocol: "http"},
					},
				},
			},
			wantErr: false,
			check: func(t *testing.T, card *server.AgentCard) {
				assert.Equal(t, "test-agent", card.Name)
				assert.Equal(t, "1.0.0", card.Version)
				assert.Equal(t, "http://test.example.com", card.URL)
				assert.NotNil(t, card.ProtocolVersion)
				assert.Equal(t, "0.3.0", *card.ProtocolVersion)
				assert.Len(t, card.Skills, 1)
			},
		},
		{
			name: "card with capabilities",
			input: &v1alpha1.AgentCard{
				Spec: v1alpha1.AgentCardSpec{
					Name:         "capable-agent",
					Version:      "2.0.0",
					Capabilities: []string{"kubernetes", "monitoring", "alerting"},
					Endpoints: []v1alpha1.AgentEndpoint{
						{URL: "http://capable.example.com"},
					},
				},
			},
			wantErr: false,
			check: func(t *testing.T, card *server.AgentCard) {
				assert.Len(t, card.Skills, 3)
				assert.Equal(t, "kubernetes", card.Skills[0].Name)
				assert.Equal(t, "monitoring", card.Skills[1].Name)
				assert.Equal(t, "alerting", card.Skills[2].Name)
			},
		},
		{
			name: "card with metadata",
			input: &v1alpha1.AgentCard{
				Spec: v1alpha1.AgentCardSpec{
					Name:    "meta-agent",
					Version: "1.0.0",
					Endpoints: []v1alpha1.AgentEndpoint{
						{URL: "http://meta.example.com"},
					},
					Metadata: map[string]string{
						"description":            "A metadata-rich agent",
						"provider.organization":  "Kagent Labs",
						"provider.url":           "https://kagent.dev",
						"iconUrl":                "https://kagent.dev/icon.png",
						"documentationUrl":       "https://docs.kagent.dev",
					},
				},
			},
			wantErr: false,
			check: func(t *testing.T, card *server.AgentCard) {
				assert.Equal(t, "A metadata-rich agent", card.Description)
				assert.NotNil(t, card.Provider)
				assert.Equal(t, "Kagent Labs", card.Provider.Organization)
				assert.NotNil(t, card.Provider.URL)
				assert.Equal(t, "https://kagent.dev", *card.Provider.URL)
				assert.NotNil(t, card.IconURL)
				assert.Equal(t, "https://kagent.dev/icon.png", *card.IconURL)
				assert.NotNil(t, card.DocumentationURL)
				assert.Equal(t, "https://docs.kagent.dev", *card.DocumentationURL)
			},
		},
		{
			name: "card without explicit version uses default",
			input: &v1alpha1.AgentCard{
				Spec: v1alpha1.AgentCardSpec{
					Name: "versionless-agent",
					Endpoints: []v1alpha1.AgentEndpoint{
						{URL: "http://test.example.com"},
					},
				},
			},
			wantErr: false,
			check: func(t *testing.T, card *server.AgentCard) {
				assert.Equal(t, "1.0.0", card.Version)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := translator.ToA2ACard(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.check != nil {
					tt.check(t, result)
				}
			}
		})
	}
}

func TestTranslator_ToJSON(t *testing.T) {
	translator := a2a.NewTranslator()

	card := &v1alpha1.AgentCard{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "json-agent",
			Namespace: "default",
		},
		Spec: v1alpha1.AgentCardSpec{
			Name:         "json-agent",
			Version:      "1.5.0",
			A2AVersion:   "0.3.0",
			Capabilities: []string{"json-processing"},
			Endpoints: []v1alpha1.AgentEndpoint{
				{
					URL:      "http://json-agent.default.svc.cluster.local:8080",
					Protocol: "http",
					Port:     8080,
				},
			},
			Metadata: map[string]string{
				"description": "JSON processing agent",
			},
			SourceRef: corev1.ObjectReference{
				APIVersion: "kagent.dev/v1alpha2",
				Kind:       "Agent",
				Name:       "json-agent",
				Namespace:  "default",
			},
		},
	}

	jsonStr, err := translator.ToJSON(card)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonStr)

	var a2aCard server.AgentCard
	err = json.Unmarshal([]byte(jsonStr), &a2aCard)
	require.NoError(t, err)

	assert.Equal(t, "json-agent", a2aCard.Name)
	assert.Equal(t, "1.5.0", a2aCard.Version)
	assert.Equal(t, "JSON processing agent", a2aCard.Description)
	assert.Equal(t, "http://json-agent.default.svc.cluster.local:8080", a2aCard.URL)
	assert.Len(t, a2aCard.Skills, 1)
	assert.Equal(t, "json-processing", a2aCard.Skills[0].Name)
}
