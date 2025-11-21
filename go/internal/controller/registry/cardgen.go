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
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kagent-dev/kagent/go/api/v1alpha1"
	"github.com/kagent-dev/kagent/go/api/v1alpha2"
	"github.com/kagent-dev/kagent/go/pkg/registry/a2a"
)

type CardGenerator struct {
	Client     client.Client
	Translator *a2a.Translator
}

func (g *CardGenerator) GenerateCard(ctx context.Context, agentRegistry *v1alpha1.AgentRegistry, agent *v1alpha2.Agent) (*v1alpha1.AgentCard, error) {
	card := &v1alpha1.AgentCard{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kagent.dev/v1alpha1",
			Kind:       "AgentCard",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      agent.Name,
			Namespace: agent.Namespace,
		},
		Spec: v1alpha1.AgentCardSpec{
			Name:         agent.Name,
			Version:      g.extractVersion(agent),
			SourceRef:    g.buildSourceRef(agent),
			Endpoints:    g.extractEndpoints(ctx, agent),
			Capabilities: g.extractCapabilities(agent),
			A2AVersion:   agentRegistry.Spec.A2AVersion,
			Metadata:     g.extractMetadata(agent),
		},
	}

	if g.Translator != nil {
		publicCard, err := g.Translator.ToJSON(card)
		if err == nil {
			card.Spec.PublicCard = publicCard
		}
	}

	hash := g.calculateHash(card.Spec)
	now := metav1.Now()
	card.Status = v1alpha1.AgentCardStatus{
		Hash:     hash,
		LastSeen: &now,
		Conditions: []metav1.Condition{
			{
				Type:               v1alpha1.AgentCardConditionTypeReady,
				Status:             metav1.ConditionTrue,
				ObservedGeneration: agent.Generation,
				LastTransitionTime: now,
				Reason:             "Generated",
				Message:            "AgentCard generated successfully",
			},
		},
		ObservedGeneration: agent.Generation,
	}

	return card, nil
}

func (g *CardGenerator) extractVersion(agent *v1alpha2.Agent) string {
	if version, ok := agent.Labels["version"]; ok {
		return version
	}
	if version, ok := agent.Labels["app.kubernetes.io/version"]; ok {
		return version
	}
	return agent.ResourceVersion
}

func (g *CardGenerator) buildSourceRef(agent *v1alpha2.Agent) corev1.ObjectReference {
	return corev1.ObjectReference{
		APIVersion: agent.APIVersion,
		Kind:       agent.Kind,
		Name:       agent.Name,
		Namespace:  agent.Namespace,
		UID:        agent.UID,
	}
}

func (g *CardGenerator) extractEndpoints(ctx context.Context, agent *v1alpha2.Agent) []v1alpha1.AgentEndpoint {
	if customEndpoint, ok := agent.Annotations[AnnotationA2AEndpoint]; ok {
		return []v1alpha1.AgentEndpoint{
			{
				URL:      customEndpoint,
				Protocol: "http",
			},
		}
	}

	if agent.Spec.Declarative != nil && agent.Spec.Declarative.A2AConfig != nil {
		endpoint := v1alpha1.AgentEndpoint{
			URL:      fmt.Sprintf("http://kagent-controller.kagent.svc.cluster.local:8083/api/a2a/%s/%s", agent.Namespace, agent.Name),
			Protocol: "http",
			Port:     8083,
		}
		return []v1alpha1.AgentEndpoint{endpoint}
	}

	service, err := g.findServiceForAgent(ctx, agent)
	if err == nil && service != nil {
		endpoints := make([]v1alpha1.AgentEndpoint, 0)
		for _, port := range service.Spec.Ports {
			endpoint := v1alpha1.AgentEndpoint{
				URL:      fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", service.Name, service.Namespace, port.Port),
				Protocol: strings.ToLower(string(port.Protocol)),
				Port:     port.Port,
			}
			endpoints = append(endpoints, endpoint)
		}
		return endpoints
	}

	return nil
}

func (g *CardGenerator) findServiceForAgent(ctx context.Context, agent *v1alpha2.Agent) (*corev1.Service, error) {
	service := &corev1.Service{}
	err := g.Client.Get(ctx, types.NamespacedName{
		Name:      agent.Name,
		Namespace: agent.Namespace,
	}, service)
	if err != nil {
		return nil, err
	}
	return service, nil
}

func (g *CardGenerator) extractCapabilities(agent *v1alpha2.Agent) []string {
	if capStr, ok := agent.Annotations[AnnotationCapabilities]; ok {
		caps := strings.Split(capStr, ",")
		for i := range caps {
			caps[i] = strings.TrimSpace(caps[i])
		}
		return caps
	}

	if agent.Spec.Declarative != nil && agent.Spec.Declarative.A2AConfig != nil && len(agent.Spec.Declarative.A2AConfig.Skills) > 0 {
		caps := make([]string, 0, len(agent.Spec.Declarative.A2AConfig.Skills))
		for _, skill := range agent.Spec.Declarative.A2AConfig.Skills {
			if skill.Name != "" {
				caps = append(caps, skill.Name)
			}
		}
		return caps
	}

	return nil
}

func (g *CardGenerator) extractMetadata(agent *v1alpha2.Agent) map[string]string {
	metadata := make(map[string]string)

	if agent.Spec.Description != "" {
		metadata["description"] = agent.Spec.Description
	}

	if agent.Spec.Declarative != nil {
		if agent.Spec.Declarative.ModelConfig != "" {
			metadata["modelConfig"] = agent.Spec.Declarative.ModelConfig
		}

		if len(agent.Spec.Declarative.Tools) > 0 {
			toolNames := make([]string, 0, len(agent.Spec.Declarative.Tools))
			for _, tool := range agent.Spec.Declarative.Tools {
				if tool.McpServer != nil && tool.McpServer.Name != "" {
					toolNames = append(toolNames, tool.McpServer.Name)
				} else if tool.Agent != nil && tool.Agent.Name != "" {
					toolNames = append(toolNames, tool.Agent.Name)
				}
			}
			if len(toolNames) > 0 {
				metadata["tools"] = strings.Join(toolNames, ",")
			}
		}
	}

	for key, value := range agent.Annotations {
		if strings.HasPrefix(key, AnnotationCardMetadataPrefix) {
			metaKey := strings.TrimPrefix(key, AnnotationCardMetadataPrefix)
			metadata[metaKey] = value
		}
	}

	return metadata
}

func (g *CardGenerator) calculateHash(spec v1alpha1.AgentCardSpec) string {
	data, _ := json.Marshal(spec)
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}
