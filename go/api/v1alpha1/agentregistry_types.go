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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AgentRegistryConditionTypeReady       = "Ready"
	AgentRegistryConditionTypeDiscovering = "Discovering"
	AgentRegistryConditionTypeError       = "Error"
)

type DiscoveryConfig struct {
	// EnableAutoDiscovery enables automatic agent discovery
	// +optional
	EnableAutoDiscovery bool `json:"enableAutoDiscovery,omitempty"`

	// NamespaceSelector restricts discovery to matching namespaces
	// +optional
	NamespaceSelector *metav1.LabelSelector `json:"namespaceSelector,omitempty"`

	// SyncInterval defines how often to sync agent cards (default: 5m)
	// +optional
	// +kubebuilder:default="5m"
	SyncInterval *metav1.Duration `json:"syncInterval,omitempty"`
}

type ObservabilityConfig struct {
	// OpenTelemetryEnabled enables OTel tracing/metrics
	// +optional
	OpenTelemetryEnabled bool `json:"openTelemetryEnabled,omitempty"`
}

// AgentRegistrySpec defines the desired state of AgentRegistry.
type AgentRegistrySpec struct {
	// Selector for discovering agents across namespaces
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty"`

	// Discovery configuration
	// +optional
	Discovery DiscoveryConfig `json:"discovery,omitempty"`

	// Observability settings
	// +optional
	Observability ObservabilityConfig `json:"observability,omitempty"`

	// A2AVersion specifies the A2A protocol version for agent cards
	// +optional
	// +kubebuilder:default="0.3.0"
	A2AVersion string `json:"a2aVersion,omitempty"`
}

// AgentRegistryStatus defines the observed state of AgentRegistry.
type AgentRegistryStatus struct {
	// RegisteredAgents is the count of active agent cards
	// +optional
	RegisteredAgents int32 `json:"registeredAgents,omitempty"`

	// Phase indicates the current lifecycle phase
	// +optional
	// +kubebuilder:validation:Enum=NotStarted;Discovering;Ready;Error
	Phase string `json:"phase,omitempty"`

	// LastSync timestamp of last successful sync
	// +optional
	LastSync *metav1.Time `json:"lastSync,omitempty"`

	// Conditions represent the latest available observations
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration reflects the generation observed by the controller
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=ar,categories=kagent
// +kubebuilder:printcolumn:name="Agents",type=integer,JSONPath=`.status.registeredAgents`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// AgentRegistry is the Schema for the agentregistries API.
type AgentRegistry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AgentRegistrySpec   `json:"spec,omitempty"`
	Status AgentRegistryStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AgentRegistryList contains a list of AgentRegistry.
type AgentRegistryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AgentRegistry `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AgentRegistry{}, &AgentRegistryList{})
}
