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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AgentCardConditionTypeReady     = "Ready"
	AgentCardConditionTypePublished = "Published"
)

type AgentEndpoint struct {
	// URL is the full endpoint URL (e.g., http://service-name.namespace.svc.cluster.local:8080)
	// +kubebuilder:validation:MinLength=1
	URL string `json:"url"`

	// Protocol (e.g., http, https, grpc)
	// +optional
	Protocol string `json:"protocol,omitempty"`

	// Port number
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port,omitempty"`
}

// AgentCardSpec defines the desired state of AgentCard.
type AgentCardSpec struct {
	// Name is the agent name
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Version of the agent
	// +optional
	Version string `json:"version,omitempty"`

	// SourceRef tracks which Agent/Deployment/Pod created this card
	SourceRef corev1.ObjectReference `json:"sourceRef"`

	// Endpoints for reaching the agent (prefer Service URLs)
	// +optional
	Endpoints []AgentEndpoint `json:"endpoints,omitempty"`

	// Capabilities provided by this agent
	// +optional
	Capabilities []string `json:"capabilities,omitempty"`

	// A2AVersion specifies the A2A protocol version
	// +optional
	// +kubebuilder:default="0.3.0"
	A2AVersion string `json:"a2aVersion,omitempty"`

	// Metadata for extensibility
	// +optional
	Metadata map[string]string `json:"metadata,omitempty"`

	// PublicCard is the A2A-compliant JSON card for external consumption
	// +optional
	PublicCard string `json:"publicCard,omitempty"`
}

// AgentCardStatus defines the observed state of AgentCard.
type AgentCardStatus struct {
	// Hash is the content hash for deduplication
	// +optional
	Hash string `json:"hash,omitempty"`

	// LastSeen timestamp of last observation
	// +optional
	LastSeen *metav1.Time `json:"lastSeen,omitempty"`

	// EndpointHealthy indicates if endpoints are reachable
	// +optional
	EndpointHealthy *bool `json:"endpointHealthy,omitempty"`

	// Conditions represent the latest available observations
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// PublishedRef points to ConfigMap with full A2A document
	// +optional
	PublishedRef *corev1.ObjectReference `json:"publishedRef,omitempty"`

	// ObservedGeneration reflects the generation observed by the controller
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=ac,categories=kagent
// +kubebuilder:printcolumn:name="Agent",type=string,JSONPath=`.spec.name`
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`
// +kubebuilder:printcolumn:name="Hash",type=string,JSONPath=`.status.hash`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// AgentCard is the Schema for the agentcards API.
type AgentCard struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AgentCardSpec   `json:"spec,omitempty"`
	Status AgentCardStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AgentCardList contains a list of AgentCard.
type AgentCardList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AgentCard `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AgentCard{}, &AgentCardList{})
}
