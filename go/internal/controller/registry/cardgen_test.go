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

package registry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/kagent-dev/kagent/go/api/v1alpha1"
	"github.com/kagent-dev/kagent/go/api/v1alpha2"
)

func TestGenerateCard_BasicAgent(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, v1alpha1.AddToScheme(scheme))
	require.NoError(t, v1alpha2.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	generator := &CardGenerator{
		Client: fakeClient,
	}

	agentRegistry := &v1alpha1.AgentRegistry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-registry",
			Namespace: "default",
		},
		Spec: v1alpha1.AgentRegistrySpec{
			A2AVersion: "0.3.0",
		},
	}

	agent := &v1alpha2.Agent{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-agent",
			Namespace:  "default",
			Generation: 1,
			UID:        "test-uid-123",
		},
		Spec: v1alpha2.AgentSpec{
			Description: "Test agent description",
		},
	}

	ctx := context.Background()
	card, err := generator.GenerateCard(ctx, agentRegistry, agent)
	require.NoError(t, err)
	require.NotNil(t, card)

	assert.Equal(t, "test-agent", card.Name)
	assert.Equal(t, "default", card.Namespace)
	assert.Equal(t, "test-agent", card.Spec.Name)
	assert.Equal(t, "0.3.0", card.Spec.A2AVersion)
	assert.Equal(t, "test-agent", card.Spec.SourceRef.Name)
	assert.Equal(t, "default", card.Spec.SourceRef.Namespace)
	assert.Equal(t, "Test agent description", card.Spec.Metadata["description"])
	assert.NotEmpty(t, card.Status.Hash)
	assert.NotNil(t, card.Status.LastSeen)
	assert.Equal(t, int64(1), card.Status.ObservedGeneration)
	require.NotEmpty(t, card.Status.Conditions)
	assert.Equal(t, v1alpha1.AgentCardConditionTypeReady, card.Status.Conditions[0].Type)
}

func TestGenerateCard_WithCapabilitiesAnnotation(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, v1alpha1.AddToScheme(scheme))
	require.NoError(t, v1alpha2.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	generator := &CardGenerator{
		Client: fakeClient,
	}

	agentRegistry := &v1alpha1.AgentRegistry{
		Spec: v1alpha1.AgentRegistrySpec{
			A2AVersion: "0.3.0",
		},
	}

	agent := &v1alpha2.Agent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-agent",
			Namespace: "default",
			Annotations: map[string]string{
				AnnotationCapabilities: "kubernetes, monitoring, alerting",
			},
		},
		Spec: v1alpha2.AgentSpec{
			Description: "Agent with capabilities",
		},
	}

	ctx := context.Background()
	card, err := generator.GenerateCard(ctx, agentRegistry, agent)
	require.NoError(t, err)

	require.Len(t, card.Spec.Capabilities, 3)
	assert.Contains(t, card.Spec.Capabilities, "kubernetes")
	assert.Contains(t, card.Spec.Capabilities, "monitoring")
	assert.Contains(t, card.Spec.Capabilities, "alerting")
}

func TestGenerateCard_WithA2AConfig(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, v1alpha1.AddToScheme(scheme))
	require.NoError(t, v1alpha2.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	generator := &CardGenerator{
		Client: fakeClient,
	}

	agentRegistry := &v1alpha1.AgentRegistry{
		Spec: v1alpha1.AgentRegistrySpec{
			A2AVersion: "0.3.0",
		},
	}

	agent := &v1alpha2.Agent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-agent",
			Namespace: "default",
		},
		Spec: v1alpha2.AgentSpec{
			Description: "Agent with A2A config",
			Declarative: &v1alpha2.DeclarativeAgentSpec{
				A2AConfig: &v1alpha2.A2AConfig{
					Skills: []v1alpha2.AgentSkill{
						{Name: "skill1"},
						{Name: "skill2"},
					},
				},
			},
		},
	}

	ctx := context.Background()
	card, err := generator.GenerateCard(ctx, agentRegistry, agent)
	require.NoError(t, err)

	require.Len(t, card.Spec.Capabilities, 2)
	assert.Contains(t, card.Spec.Capabilities, "skill1")
	assert.Contains(t, card.Spec.Capabilities, "skill2")

	require.Len(t, card.Spec.Endpoints, 1)
	assert.Contains(t, card.Spec.Endpoints[0].URL, "kagent-controller.kagent.svc.cluster.local:8083")
	assert.Equal(t, int32(8083), card.Spec.Endpoints[0].Port)
}

func TestGenerateCard_WithCustomEndpoint(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, v1alpha1.AddToScheme(scheme))
	require.NoError(t, v1alpha2.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	generator := &CardGenerator{
		Client: fakeClient,
	}

	agentRegistry := &v1alpha1.AgentRegistry{
		Spec: v1alpha1.AgentRegistrySpec{
			A2AVersion: "0.3.0",
		},
	}

	agent := &v1alpha2.Agent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-agent",
			Namespace: "default",
			Annotations: map[string]string{
				AnnotationA2AEndpoint: "https://custom-endpoint.example.com:9000",
			},
		},
		Spec: v1alpha2.AgentSpec{
			Description: "Agent with custom endpoint",
		},
	}

	ctx := context.Background()
	card, err := generator.GenerateCard(ctx, agentRegistry, agent)
	require.NoError(t, err)

	require.Len(t, card.Spec.Endpoints, 1)
	assert.Equal(t, "https://custom-endpoint.example.com:9000", card.Spec.Endpoints[0].URL)
	assert.Equal(t, "http", card.Spec.Endpoints[0].Protocol)
}

func TestGenerateCard_WithService(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, v1alpha1.AddToScheme(scheme))
	require.NoError(t, v1alpha2.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-agent",
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     "http",
					Port:     8080,
					Protocol: corev1.ProtocolTCP,
				},
				{
					Name:     "grpc",
					Port:     9090,
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(service).
		Build()

	generator := &CardGenerator{
		Client: fakeClient,
	}

	agentRegistry := &v1alpha1.AgentRegistry{
		Spec: v1alpha1.AgentRegistrySpec{
			A2AVersion: "0.3.0",
		},
	}

	agent := &v1alpha2.Agent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-agent",
			Namespace: "default",
		},
		Spec: v1alpha2.AgentSpec{
			Description: "Agent with service",
		},
	}

	ctx := context.Background()
	card, err := generator.GenerateCard(ctx, agentRegistry, agent)
	require.NoError(t, err)

	require.Len(t, card.Spec.Endpoints, 2)
	assert.Contains(t, card.Spec.Endpoints[0].URL, "test-agent.default.svc.cluster.local")
	assert.Equal(t, "tcp", card.Spec.Endpoints[0].Protocol)
	assert.Equal(t, int32(8080), card.Spec.Endpoints[0].Port)
	assert.Equal(t, int32(9090), card.Spec.Endpoints[1].Port)
}

func TestGenerateCard_WithMetadata(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, v1alpha1.AddToScheme(scheme))
	require.NoError(t, v1alpha2.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	generator := &CardGenerator{
		Client: fakeClient,
	}

	agentRegistry := &v1alpha1.AgentRegistry{
		Spec: v1alpha1.AgentRegistrySpec{
			A2AVersion: "0.3.0",
		},
	}

	agent := &v1alpha2.Agent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-agent",
			Namespace: "default",
			Annotations: map[string]string{
				AnnotationCardMetadataPrefix + "team":        "platform",
				AnnotationCardMetadataPrefix + "environment": "production",
			},
		},
		Spec: v1alpha2.AgentSpec{
			Description: "Agent with metadata",
			Declarative: &v1alpha2.DeclarativeAgentSpec{
				ModelConfig: "gpt-4",
				Tools: []*v1alpha2.Tool{
					{
						McpServer: &v1alpha2.McpServerTool{
							TypedLocalReference: v1alpha2.TypedLocalReference{
								Name: "tool1",
							},
						},
					},
					{
						Agent: &v1alpha2.TypedLocalReference{
							Name: "agent1",
						},
					},
				},
			},
		},
	}

	ctx := context.Background()
	card, err := generator.GenerateCard(ctx, agentRegistry, agent)
	require.NoError(t, err)

	assert.Equal(t, "Agent with metadata", card.Spec.Metadata["description"])
	assert.Equal(t, "gpt-4", card.Spec.Metadata["modelConfig"])
	assert.Equal(t, "platform", card.Spec.Metadata["team"])
	assert.Equal(t, "production", card.Spec.Metadata["environment"])
	assert.Equal(t, "tool1,agent1", card.Spec.Metadata["tools"])
}

func TestExtractVersion(t *testing.T) {
	generator := &CardGenerator{}

	tests := []struct {
		name     string
		agent    *v1alpha2.Agent
		expected string
	}{
		{
			name: "Version from version label",
			agent: &v1alpha2.Agent{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"version": "v1.2.3",
					},
				},
			},
			expected: "v1.2.3",
		},
		{
			name: "Version from app.kubernetes.io/version label",
			agent: &v1alpha2.Agent{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/version": "v2.0.0",
					},
				},
			},
			expected: "v2.0.0",
		},
		{
			name: "Version from ResourceVersion",
			agent: &v1alpha2.Agent{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "12345",
				},
			},
			expected: "12345",
		},
		{
			name: "Prefer version label over app.kubernetes.io/version",
			agent: &v1alpha2.Agent{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"version":                   "v1.0.0",
						"app.kubernetes.io/version": "v2.0.0",
					},
				},
			},
			expected: "v1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version := generator.extractVersion(tt.agent)
			assert.Equal(t, tt.expected, version)
		})
	}
}

func TestCalculateHash(t *testing.T) {
	generator := &CardGenerator{}

	spec1 := v1alpha1.AgentCardSpec{
		Name:         "test-agent",
		Version:      "v1.0.0",
		Capabilities: []string{"kubernetes", "monitoring"},
	}

	spec2 := v1alpha1.AgentCardSpec{
		Name:         "test-agent",
		Version:      "v1.0.0",
		Capabilities: []string{"kubernetes", "monitoring"},
	}

	spec3 := v1alpha1.AgentCardSpec{
		Name:         "test-agent",
		Version:      "v2.0.0",
		Capabilities: []string{"kubernetes", "monitoring"},
	}

	hash1 := generator.calculateHash(spec1)
	hash2 := generator.calculateHash(spec2)
	hash3 := generator.calculateHash(spec3)

	assert.Equal(t, hash1, hash2)
	assert.NotEqual(t, hash1, hash3)
	assert.NotEmpty(t, hash1)
}

func TestExtractCapabilities_EmptySkills(t *testing.T) {
	generator := &CardGenerator{}

	agent := &v1alpha2.Agent{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-agent",
		},
		Spec: v1alpha2.AgentSpec{
			Declarative: &v1alpha2.DeclarativeAgentSpec{
				A2AConfig: &v1alpha2.A2AConfig{
					Skills: []v1alpha2.AgentSkill{
						{Name: ""},
					},
				},
			},
		},
	}

	capabilities := generator.extractCapabilities(agent)
	assert.Empty(t, capabilities)
}

func TestExtractMetadata_NoDeclarative(t *testing.T) {
	generator := &CardGenerator{}

	agent := &v1alpha2.Agent{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-agent",
		},
		Spec: v1alpha2.AgentSpec{
			Description: "Simple agent",
		},
	}

	metadata := generator.extractMetadata(agent)
	assert.Equal(t, "Simple agent", metadata["description"])
	assert.Len(t, metadata, 1)
}

func TestBuildSourceRef(t *testing.T) {
	generator := &CardGenerator{}

	agent := &v1alpha2.Agent{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kagent.dev/v1alpha2",
			Kind:       "Agent",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-agent",
			Namespace: "default",
			UID:       "test-uid-123",
		},
	}

	sourceRef := generator.buildSourceRef(agent)
	assert.Equal(t, "kagent.dev/v1alpha2", sourceRef.APIVersion)
	assert.Equal(t, "Agent", sourceRef.Kind)
	assert.Equal(t, "test-agent", sourceRef.Name)
	assert.Equal(t, "default", sourceRef.Namespace)
	assert.Equal(t, "test-uid-123", string(sourceRef.UID))
}
