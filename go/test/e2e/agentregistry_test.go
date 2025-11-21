package e2e_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/kagent-dev/kagent/go/api/v1alpha1"
	"github.com/kagent-dev/kagent/go/api/v1alpha2"
)

func TestE2E_AgentRegistryDiscovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	k8sClient := setupK8sClient(t, true)

	testNamespace := "registry-e2e-test-" + time.Now().Format("20060102150405")

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNamespace,
			Labels: map[string]string{
				"kagent.dev/agent-enabled": "true",
			},
		},
	}
	err := k8sClient.Create(ctx, ns)
	require.NoError(t, err)
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		k8sClient.Delete(cleanupCtx, ns)
	}()

	agentRegistry := &v1alpha1.AgentRegistry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "e2e-registry",
			Namespace: testNamespace,
		},
		Spec: v1alpha1.AgentRegistrySpec{
			Discovery: v1alpha1.DiscoveryConfig{
				EnableAutoDiscovery: true,
				SyncInterval:        &metav1.Duration{Duration: 30 * time.Second},
			},
			A2AVersion: "0.3.0",
		},
	}
	err = k8sClient.Create(ctx, agentRegistry)
	require.NoError(t, err)

	agent := &v1alpha2.Agent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "e2e-test-agent",
			Namespace: testNamespace,
			Annotations: map[string]string{
				"kagent.dev/register-to-registry": "true",
				"kagent.dev/capabilities":         "kubernetes,test,e2e",
			},
		},
		Spec: v1alpha2.AgentSpec{
			Description: "E2E test agent for registry discovery",
			Declarative: &v1alpha2.DeclarativeAgentSpec{
				ModelConfig: "test-model",
			},
		},
	}
	err = k8sClient.Create(ctx, agent)
	require.NoError(t, err)

	err = wait.PollUntilContextTimeout(ctx, 2*time.Second, 60*time.Second, true, func(ctx context.Context) (bool, error) {
		var card v1alpha1.AgentCard
		err := k8sClient.Get(ctx, types.NamespacedName{
			Name:      "e2e-test-agent",
			Namespace: testNamespace,
		}, &card)
		if err != nil {
			t.Logf("Waiting for AgentCard to be created: %v", err)
			return false, nil
		}
		return true, nil
	})
	require.NoError(t, err, "AgentCard should be created within timeout")

	var agentCard v1alpha1.AgentCard
	err = k8sClient.Get(ctx, types.NamespacedName{
		Name:      "e2e-test-agent",
		Namespace: testNamespace,
	}, &agentCard)
	require.NoError(t, err)

	assert.Equal(t, "e2e-test-agent", agentCard.Spec.Name)
	assert.Equal(t, "0.3.0", agentCard.Spec.A2AVersion)
	assert.Contains(t, agentCard.Spec.Capabilities, "kubernetes")
	assert.Contains(t, agentCard.Spec.Capabilities, "test")
	assert.Contains(t, agentCard.Spec.Capabilities, "e2e")
	assert.Equal(t, "E2E test agent for registry discovery", agentCard.Spec.Metadata["description"])
	assert.Equal(t, "test-model", agentCard.Spec.Metadata["modelConfig"])
	assert.NotEmpty(t, agentCard.Status.Hash)
	assert.NotNil(t, agentCard.Status.LastSeen)
	assert.Equal(t, "Agent", agentCard.Spec.SourceRef.Kind)
	assert.Equal(t, "e2e-test-agent", agentCard.Spec.SourceRef.Name)

	var updatedRegistry v1alpha1.AgentRegistry
	err = k8sClient.Get(ctx, types.NamespacedName{
		Name:      "e2e-registry",
		Namespace: testNamespace,
	}, &updatedRegistry)
	require.NoError(t, err)

	assert.Equal(t, "Ready", updatedRegistry.Status.Phase)
	assert.Equal(t, int32(1), updatedRegistry.Status.RegisteredAgents)
	assert.NotNil(t, updatedRegistry.Status.LastSync)
	assert.NotEmpty(t, updatedRegistry.Status.Conditions)
}

func TestE2E_AgentRegistry_MultiNamespace(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	k8sClient := setupK8sClient(t, true)

	timestamp := time.Now().Format("20060102150405")
	ns1Name := "registry-e2e-ns1-" + timestamp
	ns2Name := "registry-e2e-ns2-" + timestamp
	registryNsName := "registry-e2e-main-" + timestamp

	ns1 := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ns1Name,
			Labels: map[string]string{
				"environment": "test",
			},
		},
	}
	err := k8sClient.Create(ctx, ns1)
	require.NoError(t, err)
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		k8sClient.Delete(cleanupCtx, ns1)
	}()

	ns2 := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ns2Name,
			Labels: map[string]string{
				"environment": "test",
			},
		},
	}
	err = k8sClient.Create(ctx, ns2)
	require.NoError(t, err)
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		k8sClient.Delete(cleanupCtx, ns2)
	}()

	registryNs := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: registryNsName,
		},
	}
	err = k8sClient.Create(ctx, registryNs)
	require.NoError(t, err)
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		k8sClient.Delete(cleanupCtx, registryNs)
	}()

	agentRegistry := &v1alpha1.AgentRegistry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "multi-ns-registry",
			Namespace: registryNsName,
		},
		Spec: v1alpha1.AgentRegistrySpec{
			Discovery: v1alpha1.DiscoveryConfig{
				EnableAutoDiscovery: true,
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"environment": "test",
					},
				},
			},
			A2AVersion: "0.3.0",
		},
	}
	err = k8sClient.Create(ctx, agentRegistry)
	require.NoError(t, err)

	agent1 := &v1alpha2.Agent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "agent-ns1",
			Namespace: ns1Name,
			Annotations: map[string]string{
				"kagent.dev/register-to-registry": "true",
			},
		},
		Spec: v1alpha2.AgentSpec{
			Description: "Agent in namespace 1",
		},
	}
	err = k8sClient.Create(ctx, agent1)
	require.NoError(t, err)

	agent2 := &v1alpha2.Agent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "agent-ns2",
			Namespace: ns2Name,
			Annotations: map[string]string{
				"kagent.dev/register-to-registry": "true",
			},
		},
		Spec: v1alpha2.AgentSpec{
			Description: "Agent in namespace 2",
		},
	}
	err = k8sClient.Create(ctx, agent2)
	require.NoError(t, err)

	err = wait.PollUntilContextTimeout(ctx, 2*time.Second, 60*time.Second, true, func(ctx context.Context) (bool, error) {
		var updatedRegistry v1alpha1.AgentRegistry
		err := k8sClient.Get(ctx, types.NamespacedName{
			Name:      "multi-ns-registry",
			Namespace: registryNsName,
		}, &updatedRegistry)
		if err != nil {
			return false, nil
		}
		return updatedRegistry.Status.RegisteredAgents == 2, nil
	})
	require.NoError(t, err, "Both agents should be discovered")

	var card1 v1alpha1.AgentCard
	err = k8sClient.Get(ctx, types.NamespacedName{
		Name:      "agent-ns1",
		Namespace: ns1Name,
	}, &card1)
	require.NoError(t, err)

	var card2 v1alpha1.AgentCard
	err = k8sClient.Get(ctx, types.NamespacedName{
		Name:      "agent-ns2",
		Namespace: ns2Name,
	}, &card2)
	require.NoError(t, err)
}

func TestE2E_AgentRegistry_AnnotationDisabled(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	k8sClient := setupK8sClient(t, true)

	testNamespace := "registry-e2e-disabled-" + time.Now().Format("20060102150405")

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNamespace,
		},
	}
	err := k8sClient.Create(ctx, ns)
	require.NoError(t, err)
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		k8sClient.Delete(cleanupCtx, ns)
	}()

	agentRegistry := &v1alpha1.AgentRegistry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "disabled-registry",
			Namespace: testNamespace,
		},
		Spec: v1alpha1.AgentRegistrySpec{
			Discovery: v1alpha1.DiscoveryConfig{
				EnableAutoDiscovery: true,
			},
		},
	}
	err = k8sClient.Create(ctx, agentRegistry)
	require.NoError(t, err)

	agentDisabled := &v1alpha2.Agent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "disabled-agent",
			Namespace: testNamespace,
			Annotations: map[string]string{
				"kagent.dev/register-to-registry": "true",
				"kagent.dev/discovery-disabled":   "true",
			},
		},
		Spec: v1alpha2.AgentSpec{
			Description: "This agent should not be discovered",
		},
	}
	err = k8sClient.Create(ctx, agentDisabled)
	require.NoError(t, err)

	time.Sleep(10 * time.Second)

	var card v1alpha1.AgentCard
	err = k8sClient.Get(ctx, types.NamespacedName{
		Name:      "disabled-agent",
		Namespace: testNamespace,
	}, &card)
	assert.Error(t, err)

	var updatedRegistry v1alpha1.AgentRegistry
	err = k8sClient.Get(ctx, types.NamespacedName{
		Name:      "disabled-registry",
		Namespace: testNamespace,
	}, &updatedRegistry)
	require.NoError(t, err)
	assert.Equal(t, int32(0), updatedRegistry.Status.RegisteredAgents)
}
