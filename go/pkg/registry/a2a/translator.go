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

package a2a

import (
	"encoding/json"
	"fmt"

	"trpc.group/trpc-go/trpc-a2a-go/server"

	"github.com/kagent-dev/kagent/go/api/v1alpha1"
)

type Translator struct{}

func NewTranslator() *Translator {
	return &Translator{}
}

func (t *Translator) ToA2ACard(card *v1alpha1.AgentCard) (*server.AgentCard, error) {
	if card == nil {
		return nil, fmt.Errorf("AgentCard cannot be nil")
	}

	a2aCard := &server.AgentCard{
		Name:        card.Spec.Name,
		Description: t.extractDescription(card),
		Version:     t.extractVersion(card),
		URL:         t.extractPrimaryURL(card),
		Capabilities: server.AgentCapabilities{
			Streaming:              boolPtr(true),
			PushNotifications:      boolPtr(false),
			StateTransitionHistory: boolPtr(false),
		},
		DefaultInputModes:  []string{"text"},
		DefaultOutputModes: []string{"text"},
		Skills:             t.extractSkills(card),
	}

	if provider := t.extractProvider(card); provider != nil {
		a2aCard.Provider = provider
	}

	if iconURL := t.extractIconURL(card); iconURL != "" {
		a2aCard.IconURL = &iconURL
	}

	if docURL := t.extractDocURL(card); docURL != "" {
		a2aCard.DocumentationURL = &docURL
	}

	if card.Spec.A2AVersion != "" {
		a2aCard.ProtocolVersion = &card.Spec.A2AVersion
	}

	return a2aCard, nil
}

func (t *Translator) ToJSON(card *v1alpha1.AgentCard) (string, error) {
	a2aCard, err := t.ToA2ACard(card)
	if err != nil {
		return "", err
	}

	data, err := json.Marshal(a2aCard)
	if err != nil {
		return "", fmt.Errorf("failed to marshal A2A card: %w", err)
	}

	return string(data), nil
}

func (t *Translator) extractDescription(card *v1alpha1.AgentCard) string {
	if desc, ok := card.Spec.Metadata["description"]; ok && desc != "" {
		return desc
	}
	return fmt.Sprintf("Agent %s", card.Spec.Name)
}

func (t *Translator) extractVersion(card *v1alpha1.AgentCard) string {
	if card.Spec.Version != "" {
		return card.Spec.Version
	}
	return "1.0.0"
}

func (t *Translator) extractPrimaryURL(card *v1alpha1.AgentCard) string {
	if len(card.Spec.Endpoints) > 0 {
		return card.Spec.Endpoints[0].URL
	}
	return ""
}

func (t *Translator) extractProvider(card *v1alpha1.AgentCard) *server.AgentProvider {
	orgName, hasOrg := card.Spec.Metadata["provider.organization"]
	orgURL, hasURL := card.Spec.Metadata["provider.url"]

	if !hasOrg {
		return nil
	}

	provider := &server.AgentProvider{
		Organization: orgName,
	}
	if hasURL {
		provider.URL = &orgURL
	}

	return provider
}

func (t *Translator) extractIconURL(card *v1alpha1.AgentCard) string {
	if iconURL, ok := card.Spec.Metadata["iconUrl"]; ok {
		return iconURL
	}
	return ""
}

func (t *Translator) extractDocURL(card *v1alpha1.AgentCard) string {
	if docURL, ok := card.Spec.Metadata["documentationUrl"]; ok {
		return docURL
	}
	return ""
}

func (t *Translator) extractSkills(card *v1alpha1.AgentCard) []server.AgentSkill {
	if len(card.Spec.Capabilities) == 0 {
		return []server.AgentSkill{
			{
				ID:          "default",
				Name:        "General Purpose",
				Description: stringPtr("General purpose agent"),
				Tags:        []string{},
			},
		}
	}

	skills := make([]server.AgentSkill, 0, len(card.Spec.Capabilities))
	for i, cap := range card.Spec.Capabilities {
		skill := server.AgentSkill{
			ID:   fmt.Sprintf("skill-%d", i),
			Name: cap,
			Tags: []string{},
		}
		skills = append(skills, skill)
	}

	return skills
}

func boolPtr(b bool) *bool {
	return &b
}

func stringPtr(s string) *string {
	return &s
}
