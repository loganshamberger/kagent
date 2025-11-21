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

package controller

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kagent-dev/kagent/go/api/v1alpha1"
	"github.com/kagent-dev/kagent/go/api/v1alpha2"
	"github.com/kagent-dev/kagent/go/internal/controller/registry"
	"github.com/kagent-dev/kagent/go/pkg/registry/a2a"
	"github.com/kagent-dev/kagent/go/pkg/registry/health"
)

var (
	agentRegistryControllerLog = ctrl.Log.WithName("agentregistry-controller")
)

type AgentRegistryController struct {
	Client        client.Client
	Scheme        *runtime.Scheme
	HealthChecker *health.Checker
}

// +kubebuilder:rbac:groups=kagent.dev,resources=agentregistries,verbs=get;list;watch
// +kubebuilder:rbac:groups=kagent.dev,resources=agentregistries/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kagent.dev,resources=agents,verbs=get;list;watch
// +kubebuilder:rbac:groups=kagent.dev,resources=agentcards,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kagent.dev,resources=agentcards/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch

func (r *AgentRegistryController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues(
		"registry", req.Name,
		"namespace", req.Namespace,
	)

	agentRegistry := &v1alpha1.AgentRegistry{}
	if err := r.Client.Get(ctx, req.NamespacedName, agentRegistry); err != nil {
		if apierrors.IsNotFound(err) {
			logger.V(1).Info("AgentRegistry not found, likely deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "failed to get AgentRegistry")
		return ctrl.Result{}, err
	}

	logger.V(1).Info("reconciling AgentRegistry",
		"generation", agentRegistry.Generation,
		"observedGeneration", agentRegistry.Status.ObservedGeneration,
	)

	if err := r.reconcileAgentRegistry(ctx, agentRegistry); err != nil {
		logger.Error(err, "failed to reconcile AgentRegistry")
		r.updateStatus(ctx, agentRegistry, registry.PhaseError, err.Error(), 0)
		return ctrl.Result{}, err
	}

	syncInterval := 5 * time.Minute
	if agentRegistry.Spec.Discovery.SyncInterval != nil {
		syncInterval = agentRegistry.Spec.Discovery.SyncInterval.Duration
	}

	logger.V(1).Info("AgentRegistry reconciliation complete", "requeueAfter", syncInterval)
	return ctrl.Result{RequeueAfter: syncInterval}, nil
}

func (r *AgentRegistryController) reconcileAgentRegistry(ctx context.Context, agentRegistry *v1alpha1.AgentRegistry) error {
	logger := log.FromContext(ctx)

	if !agentRegistry.Spec.Discovery.EnableAutoDiscovery {
		logger.V(1).Info("auto-discovery disabled, skipping")
		r.updateStatus(ctx, agentRegistry, registry.PhaseNotStarted, "Auto-discovery disabled", 0)
		return nil
	}

	logger.V(1).Info("starting agent discovery")
	r.updateStatus(ctx, agentRegistry, registry.PhaseDiscovering, "Discovering agents", 0)

	agents, err := r.discoverAgents(ctx, agentRegistry)
	if err != nil {
		return fmt.Errorf("failed to discover agents: %w", err)
	}

	logger.Info("discovered agents", "count", len(agents))
	registeredCount := 0
	for _, agent := range agents {
		if err := r.reconcileAgentCard(ctx, agentRegistry, agent); err != nil {
			logger.Error(err, "failed to reconcile AgentCard", "agent", agent.Name, "namespace", agent.Namespace)
			continue
		}
		registeredCount++
	}

	logger.Info("agent discovery complete", "registered", registeredCount, "total", len(agents))
	r.updateStatus(ctx, agentRegistry, registry.PhaseReady, "Discovery complete", int32(registeredCount))
	return nil
}

func (r *AgentRegistryController) discoverAgents(ctx context.Context, agentRegistry *v1alpha1.AgentRegistry) ([]v1alpha2.Agent, error) {
	logger := log.FromContext(ctx)

	listOpts := []client.ListOption{}

	if agentRegistry.Spec.Discovery.NamespaceSelector != nil {
		namespaces, err := r.getMatchingNamespaces(ctx, agentRegistry.Spec.Discovery.NamespaceSelector)
		if err != nil {
			return nil, fmt.Errorf("failed to get matching namespaces: %w", err)
		}

		var allAgents []v1alpha2.Agent
		for _, ns := range namespaces {
			var agentList v1alpha2.AgentList
			if err := r.Client.List(ctx, &agentList, client.InNamespace(ns)); err != nil {
				logger.Error(err, "failed to list agents", "namespace", ns)
				continue
			}
			allAgents = append(allAgents, agentList.Items...)
		}

		return r.filterAgentsByAnnotation(allAgents), nil
	}

	listOpts = append(listOpts, client.InNamespace(agentRegistry.Namespace))

	var agentList v1alpha2.AgentList
	if err := r.Client.List(ctx, &agentList, listOpts...); err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}

	return r.filterAgentsByAnnotation(agentList.Items), nil
}

func (r *AgentRegistryController) getMatchingNamespaces(ctx context.Context, selector *metav1.LabelSelector) ([]string, error) {
	labelSelector, err := metav1.LabelSelectorAsSelector(selector)
	if err != nil {
		return nil, fmt.Errorf("invalid label selector: %w", err)
	}

	var namespaceList corev1.NamespaceList
	if err := r.Client.List(ctx, &namespaceList, client.MatchingLabelsSelector{Selector: labelSelector}); err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	namespaces := make([]string, 0, len(namespaceList.Items))
	for _, ns := range namespaceList.Items {
		namespaces = append(namespaces, ns.Name)
	}

	return namespaces, nil
}

func (r *AgentRegistryController) filterAgentsByAnnotation(agents []v1alpha2.Agent) []v1alpha2.Agent {
	filtered := make([]v1alpha2.Agent, 0)
	for _, agent := range agents {
		if agent.Annotations[registry.AnnotationRegisterToRegistry] == "true" {
			if agent.Annotations[registry.AnnotationDiscoveryDisabled] != "true" {
				filtered = append(filtered, agent)
			}
		}
	}
	return filtered
}

func (r *AgentRegistryController) reconcileAgentCard(ctx context.Context, agentRegistry *v1alpha1.AgentRegistry, agent v1alpha2.Agent) error {
	logger := log.FromContext(ctx)

	translator := a2a.NewTranslator()
	generator := &registry.CardGenerator{
		Client:     r.Client,
		Translator: translator,
	}
	card, err := generator.GenerateCard(ctx, agentRegistry, &agent)
	if err != nil {
		return fmt.Errorf("failed to generate agent card: %w", err)
	}

	existing := &v1alpha1.AgentCard{}
	err = r.Client.Get(ctx, types.NamespacedName{
		Name:      card.Name,
		Namespace: card.Namespace,
	}, existing)

	if err == nil {
		if existing.Status.Hash == card.Status.Hash {
			logger.V(1).Info("agent card unchanged, skipping update", "card", card.Name)
			return nil
		}
	} else if !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to get existing agent card: %w", err)
	}

	if r.HealthChecker != nil && len(card.Spec.Endpoints) > 0 {
		healthy, err := r.HealthChecker.CheckEndpoints(ctx, card.Spec.Endpoints)
		if err != nil {
			logger.V(1).Info("health check failed", "card", card.Name, "error", err)
		}
		card.Status.EndpointHealthy = &healthy
	}

	card.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion: agentRegistry.APIVersion,
			Kind:       agentRegistry.Kind,
			Name:       agentRegistry.Name,
			UID:        agentRegistry.UID,
			Controller: ptr.To(true),
		},
	})

	if err := r.Client.Patch(ctx, card, client.Apply, client.ForceOwnership, client.FieldOwner(registry.FieldManagerAgentRegistry)); err != nil {
		return fmt.Errorf("failed to apply agent card: %w", err)
	}

	logger.Info("reconciled agent card", "card", card.Name, "agent", agent.Name, "healthy", card.Status.EndpointHealthy)
	return nil
}

func (r *AgentRegistryController) updateStatus(ctx context.Context, agentRegistry *v1alpha1.AgentRegistry, phase, message string, registeredAgents int32) {
	agentRegistry.Status.Phase = phase
	agentRegistry.Status.RegisteredAgents = registeredAgents
	agentRegistry.Status.ObservedGeneration = agentRegistry.Generation
	now := metav1.Now()
	agentRegistry.Status.LastSync = &now

	condition := metav1.Condition{
		Type:               v1alpha1.AgentRegistryConditionTypeReady,
		Status:             metav1.ConditionTrue,
		ObservedGeneration: agentRegistry.Generation,
		LastTransitionTime: now,
		Reason:             "ReconciliationSucceeded",
		Message:            message,
	}

	if phase == registry.PhaseError {
		condition.Type = v1alpha1.AgentRegistryConditionTypeError
		condition.Status = metav1.ConditionTrue
		condition.Reason = "ReconciliationFailed"
	} else if phase == registry.PhaseDiscovering {
		condition.Type = v1alpha1.AgentRegistryConditionTypeDiscovering
		condition.Reason = "DiscoveringAgents"
	}

	existingConditions := agentRegistry.Status.Conditions
	found := false
	for i, c := range existingConditions {
		if c.Type == condition.Type {
			existingConditions[i] = condition
			found = true
			break
		}
	}
	if !found {
		existingConditions = append(existingConditions, condition)
	}
	agentRegistry.Status.Conditions = existingConditions

	if err := r.Client.Status().Update(ctx, agentRegistry); err != nil {
		agentRegistryControllerLog.Error(err, "failed to update AgentRegistry status")
	}
}

func (r *AgentRegistryController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			NeedLeaderElection:      ptr.To(true),
			MaxConcurrentReconciles: 5,
		}).
		For(&v1alpha1.AgentRegistry{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Watches(
			&v1alpha2.Agent{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				if obj.GetAnnotations()[registry.AnnotationRegisterToRegistry] != "true" {
					return nil
				}

				var registries v1alpha1.AgentRegistryList
				if err := r.Client.List(ctx, &registries); err != nil {
					agentRegistryControllerLog.Error(err, "failed to list registries for agent watch")
					return nil
				}

				requests := make([]reconcile.Request, 0, len(registries.Items))
				for _, reg := range registries.Items {
					requests = append(requests, reconcile.Request{
						NamespacedName: types.NamespacedName{
							Name:      reg.Name,
							Namespace: reg.Namespace,
						},
					})
				}
				return requests
			}),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Watches(
			&corev1.Service{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				var registries v1alpha1.AgentRegistryList
				if err := r.Client.List(ctx, &registries); err != nil {
					agentRegistryControllerLog.Error(err, "failed to list registries for service watch")
					return nil
				}

				requests := make([]reconcile.Request, 0, len(registries.Items))
				for _, reg := range registries.Items {
					requests = append(requests, reconcile.Request{
						NamespacedName: types.NamespacedName{
							Name:      reg.Name,
							Namespace: reg.Namespace,
						},
					})
				}
				return requests
			}),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Named("agentregistry").
		Complete(r)
}
