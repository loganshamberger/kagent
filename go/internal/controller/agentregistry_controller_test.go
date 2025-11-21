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

package controller_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kagent-dev/kagent/go/api/v1alpha1"
	"github.com/kagent-dev/kagent/go/api/v1alpha2"
	"github.com/kagent-dev/kagent/go/internal/controller"
	"github.com/kagent-dev/kagent/go/internal/controller/registry"
	"github.com/kagent-dev/kagent/go/pkg/registry/health"
)

func TestAgentRegistryController_Reconcile_AutoDiscoveryDisabled(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)
	_ = v1alpha2.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	agentRegistry := &v1alpha1.AgentRegistry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-registry",
			Namespace: "default",
		},
		Spec: v1alpha1.AgentRegistrySpec{
			Discovery: v1alpha1.DiscoveryConfig{
				EnableAutoDiscovery: false,
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(agentRegistry).
		WithStatusSubresource(agentRegistry).
		Build()

	reconciler := &controller.AgentRegistryController{
		Client: fakeClient,
		Scheme: scheme,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-registry",
			Namespace: "default",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 5*time.Minute, result.RequeueAfter)

	var updated v1alpha1.AgentRegistry
	err = fakeClient.Get(context.Background(), req.NamespacedName, &updated)
	require.NoError(t, err)
	assert.Equal(t, registry.PhaseNotStarted, updated.Status.Phase)
}

func TestAgentRegistryController_Reconcile_DiscoverAgents(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)
	_ = v1alpha2.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	agentRegistry := &v1alpha1.AgentRegistry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-registry",
			Namespace: "default",
		},
		Spec: v1alpha1.AgentRegistrySpec{
			Discovery: v1alpha1.DiscoveryConfig{
				EnableAutoDiscovery: true,
			},
		},
	}

	agent1 := &v1alpha2.Agent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "agent-1",
			Namespace: "default",
			Annotations: map[string]string{
				registry.AnnotationRegisterToRegistry: "true",
			},
		},
		Spec: v1alpha2.AgentSpec{
			Description: "Test Agent 1",
		},
	}

	agent2 := &v1alpha2.Agent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "agent-2",
			Namespace: "default",
			Annotations: map[string]string{
				registry.AnnotationRegisterToRegistry: "true",
			},
		},
		Spec: v1alpha2.AgentSpec{
			Description: "Test Agent 2",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(agentRegistry, agent1, agent2).
		WithStatusSubresource(agentRegistry).
		Build()

	reconciler := &controller.AgentRegistryController{
		Client:        fakeClient,
		Scheme:        scheme,
		HealthChecker: health.NewChecker(5 * time.Second),
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-registry",
			Namespace: "default",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 5*time.Minute, result.RequeueAfter)

	var updated v1alpha1.AgentRegistry
	err = fakeClient.Get(context.Background(), req.NamespacedName, &updated)
	require.NoError(t, err)
	assert.Equal(t, registry.PhaseReady, updated.Status.Phase)
	assert.Equal(t, int32(2), updated.Status.RegisteredAgents)
}

func TestAgentRegistryController_Reconcile_SkipAgentWithoutAnnotation(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)
	_ = v1alpha2.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	agentRegistry := &v1alpha1.AgentRegistry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-registry",
			Namespace: "default",
		},
		Spec: v1alpha1.AgentRegistrySpec{
			Discovery: v1alpha1.DiscoveryConfig{
				EnableAutoDiscovery: true,
			},
		},
	}

	agentWithAnnotation := &v1alpha2.Agent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "agent-with-annotation",
			Namespace: "default",
			Annotations: map[string]string{
				registry.AnnotationRegisterToRegistry: "true",
			},
		},
	}

	agentWithoutAnnotation := &v1alpha2.Agent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "agent-without-annotation",
			Namespace: "default",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(agentRegistry, agentWithAnnotation, agentWithoutAnnotation).
		WithStatusSubresource(agentRegistry).
		Build()

	reconciler := &controller.AgentRegistryController{
		Client: fakeClient,
		Scheme: scheme,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-registry",
			Namespace: "default",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 5*time.Minute, result.RequeueAfter)

	var updated v1alpha1.AgentRegistry
	err = fakeClient.Get(context.Background(), req.NamespacedName, &updated)
	require.NoError(t, err)
	assert.Equal(t, int32(1), updated.Status.RegisteredAgents)
}

func TestAgentRegistryController_Reconcile_SkipDisabledAgent(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)
	_ = v1alpha2.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	agentRegistry := &v1alpha1.AgentRegistry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-registry",
			Namespace: "default",
		},
		Spec: v1alpha1.AgentRegistrySpec{
			Discovery: v1alpha1.DiscoveryConfig{
				EnableAutoDiscovery: true,
			},
		},
	}

	disabledAgent := &v1alpha2.Agent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "disabled-agent",
			Namespace: "default",
			Annotations: map[string]string{
				registry.AnnotationRegisterToRegistry: "true",
				registry.AnnotationDiscoveryDisabled:  "true",
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(agentRegistry, disabledAgent).
		WithStatusSubresource(agentRegistry).
		Build()

	reconciler := &controller.AgentRegistryController{
		Client: fakeClient,
		Scheme: scheme,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-registry",
			Namespace: "default",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 5*time.Minute, result.RequeueAfter)

	var updated v1alpha1.AgentRegistry
	err = fakeClient.Get(context.Background(), req.NamespacedName, &updated)
	require.NoError(t, err)
	assert.Equal(t, int32(0), updated.Status.RegisteredAgents)
}

func TestAgentRegistryController_Reconcile_CustomSyncInterval(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)
	_ = v1alpha2.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	customInterval := 10 * time.Minute
	agentRegistry := &v1alpha1.AgentRegistry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-registry",
			Namespace: "default",
		},
		Spec: v1alpha1.AgentRegistrySpec{
			Discovery: v1alpha1.DiscoveryConfig{
				EnableAutoDiscovery: true,
				SyncInterval:        &metav1.Duration{Duration: customInterval},
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(agentRegistry).
		WithStatusSubresource(agentRegistry).
		Build()

	reconciler := &controller.AgentRegistryController{
		Client: fakeClient,
		Scheme: scheme,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-registry",
			Namespace: "default",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, customInterval, result.RequeueAfter)
}
