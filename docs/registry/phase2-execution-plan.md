# Phase 2 Execution Plan: A2A Translation & REST API
**Agent Registry Implementation - Weeks 3-4**

## Overview
Phase 2 builds on the foundation from Phase 1 by implementing A2A-compliant card generation, publishing, and a read-only REST API for accessing the registry. This phase focuses on making agent cards consumable by external systems.

## Timeline
- **Duration**: 8-10 working days
- **Parallelization**: Epics 2.2 and 2.3 can overlap
- **Dependencies**: Requires Phase 1 completion

---

## Epic 2.1: A2A Specification Integration
**Duration**: 2 days  
**Goal**: Lock down A2A schema version and implement validator

### Task 2.1.1: Research and Document A2A v1.0 Schema
**Estimated Time**: 2 hours

**File**: `docs/registry/a2a-spec.md`

**Content**:
```markdown
# A2A Specification for kagent

## Version
A2A v1.0

## Schema Definition

### Agent Card Structure
```json
{
  "$schema": "https://a2a.dev/schema/v1.0/agent-card.json",
  "name": "string",
  "version": "string",
  "description": "string",
  "capabilities": ["string"],
  "endpoints": [
    {
      "type": "http|grpc|websocket",
      "url": "string",
      "protocol": "string",
      "metadata": {}
    }
  ],
  "metadata": {
    "namespace": "string",
    "cluster": "string",
    "sourceType": "kubernetes"
  }
}
```

### Required Fields
- `name`: Agent identifier
- `version`: Semantic version
- `endpoints`: At least one endpoint

### Optional Fields
- `description`: Human-readable description
- `capabilities`: List of capabilities
- `metadata`: Custom key-value pairs
```

**Success Criteria**:
- [ ] A2A v1.0 schema documented
- [ ] Required vs optional fields identified
- [ ] Example cards provided

### Task 2.1.2: Create A2A Types Package
**Estimated Time**: 3 hours

**File**: `go/pkg/registry/a2a/types.go`

```go
package a2a

import (
    "encoding/json"
    "fmt"
)

// Version represents the A2A specification version
const Version = "1.0"

// AgentCard represents an A2A-compliant agent card
type AgentCard struct {
    Schema       string       `json:"$schema"`
    Name         string       `json:"name"`
    Version      string       `json:"version"`
    Description  string       `json:"description,omitempty"`
    Capabilities []string     `json:"capabilities,omitempty"`
    Endpoints    []Endpoint   `json:"endpoints"`
    Metadata     Metadata     `json:"metadata,omitempty"`
}

// Endpoint represents an agent endpoint
type Endpoint struct {
    Type     EndpointType       `json:"type"`
    URL      string             `json:"url"`
    Protocol string             `json:"protocol,omitempty"`
    Metadata map[string]string  `json:"metadata,omitempty"`
}

// EndpointType represents the transport type
type EndpointType string

const (
    EndpointTypeHTTP      EndpointType = "http"
    EndpointTypeGRPC      EndpointType = "grpc"
    EndpointTypeWebSocket EndpointType = "websocket"
)

// Metadata contains custom key-value metadata
type Metadata map[string]string

// Validate checks if the card is A2A-compliant
func (c *AgentCard) Validate() error {
    if c.Name == "" {
        return fmt.Errorf("name is required")
    }
    if c.Version == "" {
        return fmt.Errorf("version is required")
    }
    if len(c.Endpoints) == 0 {
        return fmt.Errorf("at least one endpoint is required")
    }
    
    for i, ep := range c.Endpoints {
        if err := ep.Validate(); err != nil {
            return fmt.Errorf("endpoint[%d]: %w", i, err)
        }
    }
    
    return nil
}

// Validate checks if the endpoint is valid
func (e *Endpoint) Validate() error {
    if e.Type == "" {
        return fmt.Errorf("type is required")
    }
    if e.URL == "" {
        return fmt.Errorf("url is required")
    }
    
    switch e.Type {
    case EndpointTypeHTTP, EndpointTypeGRPC, EndpointTypeWebSocket:
        // Valid types
    default:
        return fmt.Errorf("invalid endpoint type: %s", e.Type)
    }
    
    return nil
}

// ToJSON serializes the card to JSON
func (c *AgentCard) ToJSON() ([]byte, error) {
    return json.MarshalIndent(c, "", "  ")
}

// FromJSON deserializes a card from JSON
func FromJSON(data []byte) (*AgentCard, error) {
    var card AgentCard
    if err := json.Unmarshal(data, &card); err != nil {
        return nil, err
    }
    
    if err := card.Validate(); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    return &card, nil
}
```

**Success Criteria**:
- [ ] A2A types compile without errors
- [ ] Validation logic covers all required fields
- [ ] JSON serialization works correctly

### Task 2.1.3: Create A2A Validator with Tests
**Estimated Time**: 2 hours

**File**: `go/pkg/registry/a2a/validator_test.go`

```go
package a2a_test

import (
    "testing"
    
    "github.com/kagent-dev/kagent/go/pkg/registry/a2a"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestAgentCard_Validate(t *testing.T) {
    tests := []struct {
        name    string
        card    *a2a.AgentCard
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid card",
            card: &a2a.AgentCard{
                Schema:  "https://a2a.dev/schema/v1.0/agent-card.json",
                Name:    "test-agent",
                Version: "v1.0.0",
                Endpoints: []a2a.Endpoint{
                    {
                        Type: a2a.EndpointTypeHTTP,
                        URL:  "http://example.com",
                    },
                },
            },
            wantErr: false,
        },
        {
            name: "missing name",
            card: &a2a.AgentCard{
                Version: "v1.0.0",
                Endpoints: []a2a.Endpoint{
                    {Type: a2a.EndpointTypeHTTP, URL: "http://example.com"},
                },
            },
            wantErr: true,
            errMsg:  "name is required",
        },
        {
            name: "missing version",
            card: &a2a.AgentCard{
                Name: "test-agent",
                Endpoints: []a2a.Endpoint{
                    {Type: a2a.EndpointTypeHTTP, URL: "http://example.com"},
                },
            },
            wantErr: true,
            errMsg:  "version is required",
        },
        {
            name: "missing endpoints",
            card: &a2a.AgentCard{
                Name:      "test-agent",
                Version:   "v1.0.0",
                Endpoints: []a2a.Endpoint{},
            },
            wantErr: true,
            errMsg:  "at least one endpoint is required",
        },
        {
            name: "invalid endpoint type",
            card: &a2a.AgentCard{
                Name:    "test-agent",
                Version: "v1.0.0",
                Endpoints: []a2a.Endpoint{
                    {Type: "invalid", URL: "http://example.com"},
                },
            },
            wantErr: true,
            errMsg:  "invalid endpoint type",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.card.Validate()
            if tt.wantErr {
                require.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                require.NoError(t, err)
            }
        })
    }
}

func TestAgentCard_JSON(t *testing.T) {
    original := &a2a.AgentCard{
        Schema:       "https://a2a.dev/schema/v1.0/agent-card.json",
        Name:         "test-agent",
        Version:      "v1.0.0",
        Description:  "Test agent",
        Capabilities: []string{"kubernetes", "monitoring"},
        Endpoints: []a2a.Endpoint{
            {
                Type:     a2a.EndpointTypeHTTP,
                URL:      "http://test-agent.kagent.svc.cluster.local:8080",
                Protocol: "a2a/v1",
            },
        },
        Metadata: a2a.Metadata{
            "namespace": "kagent",
            "cluster":   "production",
        },
    }
    
    // Serialize
    data, err := original.ToJSON()
    require.NoError(t, err)
    
    // Deserialize
    parsed, err := a2a.FromJSON(data)
    require.NoError(t, err)
    
    // Compare
    assert.Equal(t, original.Name, parsed.Name)
    assert.Equal(t, original.Version, parsed.Version)
    assert.Equal(t, original.Capabilities, parsed.Capabilities)
    assert.Len(t, parsed.Endpoints, 1)
}
```

**Validation**:
```bash
cd go
go test -v ./pkg/registry/a2a/...
```

**Success Criteria**:
- [ ] All validator tests pass
- [ ] Edge cases covered (empty strings, nil slices, etc.)
- [ ] JSON round-trip works correctly

### Task 2.1.4: Add A2A Schema Versioning Support
**Estimated Time**: 1 hour

**File**: `go/pkg/registry/a2a/version.go`

```go
package a2a

import (
    "fmt"
)

// SupportedVersions lists all A2A versions we support
var SupportedVersions = []string{"1.0"}

// IsVersionSupported checks if a version is supported
func IsVersionSupported(version string) bool {
    for _, v := range SupportedVersions {
        if v == version {
            return true
        }
    }
    return false
}

// GetSchemaURL returns the schema URL for a version
func GetSchemaURL(version string) (string, error) {
    if !IsVersionSupported(version) {
        return "", fmt.Errorf("unsupported A2A version: %s", version)
    }
    
    return fmt.Sprintf("https://a2a.dev/schema/v%s/agent-card.json", version), nil
}

// LatestVersion returns the latest supported version
func LatestVersion() string {
    return SupportedVersions[len(SupportedVersions)-1]
}
```

**Success Criteria**:
- [ ] Version checking works
- [ ] Future versions can be added easily
- [ ] Schema URLs generated correctly

---

## Epic 2.2: A2A Card Translator
**Duration**: 3 days  
**Goal**: Convert AgentCard CRDs to A2A-compliant JSON documents

### Task 2.2.1: Create Translator Package
**Estimated Time**: 3 hours

**File**: `go/pkg/registry/translator/translator.go`

```go
package translator

import (
    "fmt"
    
    kagentv1alpha1 "github.com/kagent-dev/kagent/go/api/v1alpha1"
    "github.com/kagent-dev/kagent/go/pkg/registry/a2a"
)

// Translator converts kagent AgentCards to A2A format
type Translator struct {
    A2AVersion string
}

// NewTranslator creates a new translator
func NewTranslator(a2aVersion string) (*Translator, error) {
    if !a2a.IsVersionSupported(a2aVersion) {
        return nil, fmt.Errorf("unsupported A2A version: %s", a2aVersion)
    }
    
    return &Translator{A2AVersion: a2aVersion}, nil
}

// Translate converts an AgentCard to A2A format
func (t *Translator) Translate(card *kagentv1alpha1.AgentCard) (*a2a.AgentCard, error) {
    schemaURL, err := a2a.GetSchemaURL(t.A2AVersion)
    if err != nil {
        return nil, err
    }
    
    a2aCard := &a2a.AgentCard{
        Schema:       schemaURL,
        Name:         card.Spec.Name,
        Version:      card.Spec.Version,
        Description:  getDescription(card),
        Capabilities: card.Spec.Capabilities,
        Endpoints:    translateEndpoints(card.Spec.Endpoints),
        Metadata:     enrichMetadata(card),
    }
    
    if err := a2aCard.Validate(); err != nil {
        return nil, fmt.Errorf("translated card is invalid: %w", err)
    }
    
    return a2aCard, nil
}

func translateEndpoints(endpoints []kagentv1alpha1.AgentEndpoint) []a2a.Endpoint {
    result := make([]a2a.Endpoint, len(endpoints))
    for i, ep := range endpoints {
        result[i] = a2a.Endpoint{
            Type:     a2a.EndpointType(ep.Type),
            URL:      ep.URL,
            Protocol: ep.Protocol,
            Metadata: make(map[string]string),
        }
    }
    return result
}

func getDescription(card *kagentv1alpha1.AgentCard) string {
    // Try to get description from metadata
    if desc, ok := card.Spec.Metadata["description"]; ok {
        return desc
    }
    
    // Try to get from annotations
    if desc, ok := card.Annotations["kagent.io/description"]; ok {
        return desc
    }
    
    return fmt.Sprintf("kagent agent: %s", card.Spec.Name)
}

func enrichMetadata(card *kagentv1alpha1.AgentCard) a2a.Metadata {
    metadata := a2a.Metadata{
        "namespace":  card.Namespace,
        "sourceType": "kubernetes",
        "kagent":     "true",
    }
    
    // Add source reference info
    if card.Spec.SourceRef.Kind != "" {
        metadata["sourceKind"] = card.Spec.SourceRef.Kind
        metadata["sourceName"] = card.Spec.SourceRef.Name
    }
    
    // Merge custom metadata
    for k, v := range card.Spec.Metadata {
        metadata[k] = v
    }
    
    return metadata
}
```

**Success Criteria**:
- [ ] Translator compiles without errors
- [ ] All AgentCard fields mapped to A2A
- [ ] Validation runs on translated cards

### Task 2.2.2: Write Translator Tests
**Estimated Time**: 2 hours

**File**: `go/pkg/registry/translator/translator_test.go`

```go
package translator_test

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    
    kagentv1alpha1 "github.com/kagent-dev/kagent/go/api/v1alpha1"
    "github.com/kagent-dev/kagent/go/pkg/registry/a2a"
    "github.com/kagent-dev/kagent/go/pkg/registry/translator"
)

func TestTranslator_Translate(t *testing.T) {
    trans, err := translator.NewTranslator("1.0")
    require.NoError(t, err)
    
    card := &kagentv1alpha1.AgentCard{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "test-agent",
            Namespace: "kagent",
            Annotations: map[string]string{
                "kagent.io/description": "Test agent for translation",
            },
        },
        Spec: kagentv1alpha1.AgentCardSpec{
            Name:    "test-agent",
            Version: "v1.0.0",
            SourceRef: corev1.ObjectReference{
                Kind: "Agent",
                Name: "test-agent",
            },
            Endpoints: []kagentv1alpha1.AgentEndpoint{
                {
                    Type:     "http",
                    URL:      "http://test-agent.kagent.svc.cluster.local:8080",
                    Protocol: "a2a/v1",
                },
            },
            Capabilities: []string{"kubernetes", "monitoring"},
            A2AVersion:   "1.0",
            Metadata: map[string]string{
                "custom": "value",
            },
        },
    }
    
    a2aCard, err := trans.Translate(card)
    require.NoError(t, err)
    
    assert.Equal(t, "test-agent", a2aCard.Name)
    assert.Equal(t, "v1.0.0", a2aCard.Version)
    assert.Equal(t, "Test agent for translation", a2aCard.Description)
    assert.Equal(t, []string{"kubernetes", "monitoring"}, a2aCard.Capabilities)
    assert.Len(t, a2aCard.Endpoints, 1)
    assert.Equal(t, a2a.EndpointTypeHTTP, a2aCard.Endpoints[0].Type)
    assert.Equal(t, "kagent", a2aCard.Metadata["namespace"])
    assert.Equal(t, "value", a2aCard.Metadata["custom"])
}

func TestTranslator_UnsupportedVersion(t *testing.T) {
    _, err := translator.NewTranslator("99.0")
    require.Error(t, err)
    assert.Contains(t, err.Error(), "unsupported A2A version")
}

func TestTranslator_InvalidCard(t *testing.T) {
    trans, err := translator.NewTranslator("1.0")
    require.NoError(t, err)
    
    // Card with no endpoints (invalid)
    card := &kagentv1alpha1.AgentCard{
        Spec: kagentv1alpha1.AgentCardSpec{
            Name:      "invalid",
            Version:   "v1.0.0",
            Endpoints: []kagentv1alpha1.AgentEndpoint{},
        },
    }
    
    _, err = trans.Translate(card)
    require.Error(t, err)
    assert.Contains(t, err.Error(), "translated card is invalid")
}
```

**Validation**:
```bash
cd go
go test -v ./pkg/registry/translator/...
```

### Task 2.2.3: Integrate Translator into Controller
**Estimated Time**: 2 hours

Update `go/internal/controller/agentregistry_controller.go`:

```go
import (
    "github.com/kagent-dev/kagent/go/pkg/registry/translator"
)

func (r *AgentRegistryReconciler) upsertAgentCard(ctx context.Context, gen *registry.CardGenerator, agent *kagentv1alpha1.Agent) error {
    newCard := gen.GenerateCard(agent)
    
    // Translate to A2A format
    trans, err := translator.NewTranslator(newCard.Spec.A2AVersion)
    if err != nil {
        return fmt.Errorf("failed to create translator: %w", err)
    }
    
    a2aCard, err := trans.Translate(newCard)
    if err != nil {
        return fmt.Errorf("failed to translate to A2A: %w", err)
    }
    
    // Store A2A JSON in spec
    a2aJSON, err := a2aCard.ToJSON()
    if err != nil {
        return fmt.Errorf("failed to serialize A2A card: %w", err)
    }
    newCard.Spec.PublicCard = string(a2aJSON)
    
    // ... (rest of upsert logic)
}
```

**Success Criteria**:
- [ ] Controller generates A2A cards
- [ ] PublicCard field populated with valid JSON
- [ ] Validation errors logged appropriately

### Task 2.2.4: Add A2A Publishing to ConfigMap (Optional)
**Estimated Time**: 3 hours

For large A2A documents, store in ConfigMaps instead of inline:

**File**: `go/pkg/registry/publisher/configmap.go`

```go
package publisher

import (
    "context"
    "fmt"
    
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
    
    kagentv1alpha1 "github.com/kagent-dev/kagent/go/api/v1alpha1"
)

// ConfigMapPublisher publishes A2A cards to ConfigMaps
type ConfigMapPublisher struct {
    client.Client
}

// NewConfigMapPublisher creates a new publisher
func NewConfigMapPublisher(c client.Client) *ConfigMapPublisher {
    return &ConfigMapPublisher{Client: c}
}

// Publish creates or updates a ConfigMap with the A2A card
func (p *ConfigMapPublisher) Publish(ctx context.Context, card *kagentv1alpha1.AgentCard, a2aJSON string) error {
    cm := &corev1.ConfigMap{
        ObjectMeta: metav1.ObjectMeta{
            Name:      fmt.Sprintf("%s-a2a", card.Name),
            Namespace: card.Namespace,
            Labels: map[string]string{
                "kagent.io/agent-card": card.Name,
                "kagent.io/a2a":        "true",
            },
        },
        Data: map[string]string{
            "agent-card.json": a2aJSON,
        },
    }
    
    // Set owner reference for garbage collection
    if err := controllerutil.SetControllerReference(card, cm, p.Scheme()); err != nil {
        return fmt.Errorf("failed to set owner reference: %w", err)
    }
    
    // Server-side apply
    if err := p.Patch(ctx, cm, client.Apply, client.ForceOwnership, client.FieldOwner("kagent/registry-publisher")); err != nil {
        return fmt.Errorf("failed to publish ConfigMap: %w", err)
    }
    
    // Update AgentCard status with reference
    card.Status.PublishedRef = &corev1.ObjectReference{
        APIVersion: "v1",
        Kind:       "ConfigMap",
        Name:       cm.Name,
        Namespace:  cm.Namespace,
    }
    
    return nil
}
```

**Success Criteria**:
- [ ] ConfigMaps created with A2A JSON
- [ ] Owner references set correctly
- [ ] AgentCard status references ConfigMap

---

## Epic 2.3: REST API for Registry Access
**Duration**: 3 days  
**Goal**: Expose read-only HTTP API for querying agent cards

### Task 2.3.1: Design REST API Specification
**Estimated Time**: 1 hour

**File**: `docs/registry/api-spec.md`

```markdown
# Agent Registry REST API Specification

## Version
v1alpha1

## Base Path
`/api/v1alpha1/registry`

## Authentication
- In-cluster only (ClusterIP service)
- Kubernetes RBAC via ServiceAccount tokens
- No custom authentication in Phase 2

## Endpoints

### List All Agent Cards
```
GET /api/v1alpha1/registry/cards
```

**Query Parameters**:
- `namespace`: Filter by namespace (optional)
- `capability`: Filter by capability (optional, repeatable)

**Response**:
```json
{
  "apiVersion": "kagent.io/v1alpha1",
  "kind": "AgentCardList",
  "items": [
    {
      "name": "agent-1",
      "namespace": "kagent",
      "a2aCard": { ... }
    }
  ]
}
```

### Get Agent Card by Name
```
GET /api/v1alpha1/registry/cards/{namespace}/{name}
```

**Response**:
```json
{
  "apiVersion": "kagent.io/v1alpha1",
  "kind": "AgentCard",
  "name": "agent-1",
  "namespace": "kagent",
  "a2aCard": { ... }
}
```

### Get A2A Card Only
```
GET /api/v1alpha1/registry/cards/{namespace}/{name}/a2a
```

**Response**: Raw A2A JSON

### Health Check
```
GET /healthz
```

**Response**: `200 OK`
```

**Success Criteria**:
- [ ] API spec documented
- [ ] All endpoints defined
- [ ] Response formats specified

### Task 2.3.2: Implement API Server
**Estimated Time**: 4 hours

**File**: `go/pkg/registry/api/server.go`

```go
package api

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    
    "github.com/gorilla/mux"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/log"
    
    kagentv1alpha1 "github.com/kagent-dev/kagent/go/api/v1alpha1"
)

var tracer = otel.Tracer("kagent-registry-api")

// Server provides HTTP API for the agent registry
type Server struct {
    Client client.Client
    Port   int
    router *mux.Router
}

// NewServer creates a new API server
func NewServer(c client.Client, port int) *Server {
    s := &Server{
        Client: c,
        Port:   port,
        router: mux.NewRouter(),
    }
    
    s.registerRoutes()
    return s
}

func (s *Server) registerRoutes() {
    api := s.router.PathPrefix("/api/v1alpha1/registry").Subrouter()
    
    api.HandleFunc("/cards", s.listCards).Methods("GET")
    api.HandleFunc("/cards/{namespace}/{name}", s.getCard).Methods("GET")
    api.HandleFunc("/cards/{namespace}/{name}/a2a", s.getA2ACard).Methods("GET")
    
    s.router.HandleFunc("/healthz", s.health).Methods("GET")
    
    // Add logging middleware
    s.router.Use(loggingMiddleware)
    s.router.Use(tracingMiddleware)
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
    srv := &http.Server{
        Addr:         fmt.Sprintf(":%d", s.Port),
        Handler:      s.router,
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 10 * time.Second,
    }
    
    log := log.FromContext(ctx)
    log.Info("Starting registry API server", "port", s.Port)
    
    go func() {
        <-ctx.Done()
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        srv.Shutdown(shutdownCtx)
    }()
    
    return srv.ListenAndServe()
}

func (s *Server) listCards(w http.ResponseWriter, r *http.Request) {
    ctx, span := tracer.Start(r.Context(), "API.ListCards")
    defer span.End()
    
    namespace := r.URL.Query().Get("namespace")
    capability := r.URL.Query().Get("capability")
    
    span.SetAttributes(
        attribute.String("namespace", namespace),
        attribute.String("capability", capability),
    )
    
    var cards kagentv1alpha1.AgentCardList
    listOpts := []client.ListOption{}
    
    if namespace != "" {
        listOpts = append(listOpts, client.InNamespace(namespace))
    }
    
    if err := s.Client.List(ctx, &cards, listOpts...); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Filter by capability if requested
    if capability != "" {
        filtered := filterByCapability(cards.Items, capability)
        cards.Items = filtered
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(cards)
}

func (s *Server) getCard(w http.ResponseWriter, r *http.Request) {
    ctx, span := tracer.Start(r.Context(), "API.GetCard")
    defer span.End()
    
    vars := mux.Vars(r)
    namespace := vars["namespace"]
    name := vars["name"]
    
    span.SetAttributes(
        attribute.String("namespace", namespace),
        attribute.String("name", name),
    )
    
    var card kagentv1alpha1.AgentCard
    key := client.ObjectKey{Namespace: namespace, Name: name}
    
    if err := s.Client.Get(ctx, key, &card); err != nil {
        if client.IgnoreNotFound(err) == nil {
            http.Error(w, "not found", http.StatusNotFound)
        } else {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(card)
}

func (s *Server) getA2ACard(w http.ResponseWriter, r *http.Request) {
    ctx, span := tracer.Start(r.Context(), "API.GetA2ACard")
    defer span.End()
    
    vars := mux.Vars(r)
    namespace := vars["namespace"]
    name := vars["name"]
    
    var card kagentv1alpha1.AgentCard
    key := client.ObjectKey{Namespace: namespace, Name: name}
    
    if err := s.Client.Get(ctx, key, &card); err != nil {
        if client.IgnoreNotFound(err) == nil {
            http.Error(w, "not found", http.StatusNotFound)
        } else {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        return
    }
    
    if card.Spec.PublicCard == "" {
        http.Error(w, "A2A card not generated", http.StatusNotFound)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(card.Spec.PublicCard))
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

func filterByCapability(cards []kagentv1alpha1.AgentCard, capability string) []kagentv1alpha1.AgentCard {
    result := []kagentv1alpha1.AgentCard{}
    for _, card := range cards {
        for _, cap := range card.Spec.Capabilities {
            if cap == capability {
                result = append(result, card)
                break
            }
        }
    }
    return result
}

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Log.Info("API request",
            "method", r.Method,
            "path", r.URL.Path,
            "duration", time.Since(start),
        )
    })
}

func tracingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx, span := tracer.Start(r.Context(), fmt.Sprintf("%s %s", r.Method, r.URL.Path))
        defer span.End()
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

**Success Criteria**:
- [ ] Server compiles and starts
- [ ] All routes registered
- [ ] Middleware applied correctly

### Task 2.3.3: Wire API Server into Controller
**Estimated Time**: 1 hour

Update `go/cmd/controller/main.go`:

```go
import (
    "github.com/kagent-dev/kagent/go/pkg/registry/api"
)

// Add flag for API port
var apiPort int

func init() {
    flag.IntVar(&apiPort, "registry-api-port", 8084, "Port for registry REST API")
}

func main() {
    // ... (existing setup)
    
    // Start registry API server
    if apiPort > 0 {
        apiServer := api.NewServer(mgr.GetClient(), apiPort)
        go func() {
            if err := apiServer.Start(context.Background()); err != nil {
                setupLog.Error(err, "failed to start registry API server")
            }
        }()
    }
    
    // ... (start manager)
}
```

Update `go/config/manager/manager.yaml`:

```yaml
containers:
- name: manager
  args:
  - --leader-elect
  - --registry-api-port=8084
  ports:
  - containerPort: 8084
    name: registry-api
    protocol: TCP
```

**Success Criteria**:
- [ ] API server starts with controller
- [ ] Port configurable via flag
- [ ] Graceful shutdown on context cancellation

### Task 2.3.4: Create Service for API Access
**Estimated Time**: 30 minutes

**File**: `helm/kagent/templates/registry-api-service.yaml`

```yaml
{{- if .Values.registry.api.enabled }}
apiVersion: v1
kind: Service
metadata:
  name: {{ include "kagent.fullname" . }}-registry-api
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "kagent.labels" . | nindent 4 }}
    app.kubernetes.io/component: registry-api
spec:
  type: {{ .Values.registry.api.service.type }}
  ports:
    - port: {{ .Values.registry.api.port }}
      targetPort: registry-api
      protocol: TCP
      name: http
  selector:
    {{- include "kagent.selectorLabels" . | nindent 4 }}
    app.kubernetes.io/component: controller
{{- end }}
```

Update `helm/kagent/values.yaml`:

```yaml
registry:
  api:
    enabled: true
    port: 8084
    service:
      type: ClusterIP
```

**Success Criteria**:
- [ ] Service created in cluster
- [ ] ClusterIP accessible from pods
- [ ] Port matches controller configuration

### Task 2.3.5: Write API Integration Tests
**Estimated Time**: 3 hours

**File**: `go/pkg/registry/api/server_test.go`

```go
package api_test

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "sigs.k8s.io/controller-runtime/pkg/client/fake"
    
    kagentv1alpha1 "github.com/kagent-dev/kagent/go/api/v1alpha1"
    "github.com/kagent-dev/kagent/go/pkg/registry/api"
)

func TestServer_ListCards(t *testing.T) {
    // Create fake client with test data
    card1 := &kagentv1alpha1.AgentCard{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "agent-1",
            Namespace: "default",
        },
        Spec: kagentv1alpha1.AgentCardSpec{
            Name:         "agent-1",
            Capabilities: []string{"kubernetes"},
        },
    }
    
    scheme := runtime.NewScheme()
    _ = kagentv1alpha1.AddToScheme(scheme)
    client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(card1).Build()
    
    server := api.NewServer(client, 8084)
    
    // Test list all
    req := httptest.NewRequest("GET", "/api/v1alpha1/registry/cards", nil)
    w := httptest.NewRecorder()
    server.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
    
    var cardList kagentv1alpha1.AgentCardList
    err := json.Unmarshal(w.Body.Bytes(), &cardList)
    require.NoError(t, err)
    assert.Len(t, cardList.Items, 1)
}

func TestServer_GetCard(t *testing.T) {
    card := &kagentv1alpha1.AgentCard{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "test-agent",
            Namespace: "kagent",
        },
        Spec: kagentv1alpha1.AgentCardSpec{
            Name:       "test-agent",
            PublicCard: `{"name":"test-agent"}`,
        },
    }
    
    scheme := runtime.NewScheme()
    _ = kagentv1alpha1.AddToScheme(scheme)
    client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(card).Build()
    
    server := api.NewServer(client, 8084)
    
    req := httptest.NewRequest("GET", "/api/v1alpha1/registry/cards/kagent/test-agent", nil)
    w := httptest.NewRecorder()
    server.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
}

func TestServer_GetA2ACard(t *testing.T) {
    a2aJSON := `{"name":"test-agent","version":"v1.0.0"}`
    card := &kagentv1alpha1.AgentCard{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "test-agent",
            Namespace: "kagent",
        },
        Spec: kagentv1alpha1.AgentCardSpec{
            Name:       "test-agent",
            PublicCard: a2aJSON,
        },
    }
    
    scheme := runtime.NewScheme()
    _ = kagentv1alpha1.AddToScheme(scheme)
    client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(card).Build()
    
    server := api.NewServer(client, 8084)
    
    req := httptest.NewRequest("GET", "/api/v1alpha1/registry/cards/kagent/test-agent/a2a", nil)
    w := httptest.NewRecorder()
    server.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
    assert.Equal(t, a2aJSON, w.Body.String())
}
```

**Validation**:
```bash
cd go
go test -v ./pkg/registry/api/...
```

---

## Epic 2.4: Enhanced Discovery Features
**Duration**: 2 days  
**Goal**: Improve discovery with Service endpoint resolution and health checking

### Task 2.4.1: Implement Service Endpoint Resolution
**Estimated Time**: 3 hours

**File**: `go/pkg/registry/discovery/endpoints.go`

```go
package discovery

import (
    "context"
    "fmt"
    
    corev1 "k8s.io/api/core/v1"
    "sigs.k8s.io/controller-runtime/pkg/client"
    
    kagentv1alpha1 "github.com/kagent-dev/kagent/go/api/v1alpha1"
)

// EndpointResolver resolves Service endpoints for agents
type EndpointResolver struct {
    client.Client
}

// NewEndpointResolver creates a new resolver
func NewEndpointResolver(c client.Client) *EndpointResolver {
    return &EndpointResolver{Client: c}
}

// ResolveEndpoints finds the best endpoint for an agent
func (r *EndpointResolver) ResolveEndpoints(ctx context.Context, agent *kagentv1alpha1.Agent) ([]kagentv1alpha1.AgentEndpoint, error) {
    // Check for explicit annotation
    if endpoint, ok := agent.Annotations["kagent.io/a2a-endpoint"]; ok {
        return []kagentv1alpha1.AgentEndpoint{
            {
                Type:     "http",
                URL:      endpoint,
                Protocol: "a2a/v1",
            },
        }, nil
    }
    
    // Try to find a Service for this agent
    service, err := r.findServiceForAgent(ctx, agent)
    if err == nil && service != nil {
        return r.endpointsFromService(service), nil
    }
    
    // Fallback: construct default endpoint
    return r.defaultEndpoints(agent), nil
}

func (r *EndpointResolver) findServiceForAgent(ctx context.Context, agent *kagentv1alpha1.Agent) (*corev1.Service, error) {
    // Look for service with same name
    service := &corev1.Service{}
    key := client.ObjectKey{
        Namespace: agent.Namespace,
        Name:      agent.Name,
    }
    
    if err := r.Get(ctx, key, service); err != nil {
        return nil, client.IgnoreNotFound(err)
    }
    
    return service, nil
}

func (r *EndpointResolver) endpointsFromService(svc *corev1.Service) []kagentv1alpha1.AgentEndpoint {
    endpoints := []kagentv1alpha1.AgentEndpoint{}
    
    for _, port := range svc.Spec.Ports {
        url := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d",
            svc.Name, svc.Namespace, port.Port)
        
        endpoints = append(endpoints, kagentv1alpha1.AgentEndpoint{
            Type:     "http",
            URL:      url,
            Protocol: "a2a/v1",
        })
    }
    
    return endpoints
}

func (r *EndpointResolver) defaultEndpoints(agent *kagentv1alpha1.Agent) []kagentv1alpha1.AgentEndpoint {
    url := fmt.Sprintf("http://%s.%s.svc.cluster.local:8080",
        agent.Name, agent.Namespace)
    
    return []kagentv1alpha1.AgentEndpoint{
        {
            Type:     "http",
            URL:      url,
            Protocol: "a2a/v1",
        },
    }
}
```

**Success Criteria**:
- [ ] Services preferred over Pod IPs
- [ ] Custom endpoints from annotations work
- [ ] Multiple ports handled correctly

### Task 2.4.2: Add Endpoint Health Checking
**Estimated Time**: 3 hours

**File**: `go/pkg/registry/health/checker.go`

```go
package health

import (
    "context"
    "fmt"
    "net/http"
    "time"
    
    kagentv1alpha1 "github.com/kagent-dev/kagent/go/api/v1alpha1"
)

// Checker performs health checks on agent endpoints
type Checker struct {
    HTTPClient *http.Client
    Timeout    time.Duration
}

// NewChecker creates a new health checker
func NewChecker() *Checker {
    return &Checker{
        HTTPClient: &http.Client{
            Timeout: 5 * time.Second,
        },
        Timeout: 5 * time.Second,
    }
}

// CheckEndpoint verifies an endpoint is reachable
func (c *Checker) CheckEndpoint(ctx context.Context, endpoint kagentv1alpha1.AgentEndpoint) (bool, error) {
    if endpoint.Type != "http" {
        // Only support HTTP health checks for now
        return false, fmt.Errorf("unsupported endpoint type: %s", endpoint.Type)
    }
    
    healthURL := fmt.Sprintf("%s/healthz", endpoint.URL)
    
    req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
    if err != nil {
        return false, err
    }
    
    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return false, err
    }
    defer resp.Body.Close()
    
    return resp.StatusCode == http.StatusOK, nil
}

// CheckCard checks all endpoints for a card
func (c *Checker) CheckCard(ctx context.Context, card *kagentv1alpha1.AgentCard) bool {
    for _, endpoint := range card.Spec.Endpoints {
        healthy, _ := c.CheckEndpoint(ctx, endpoint)
        if healthy {
            return true
        }
    }
    return false
}
```

Integrate into controller:

```go
func (r *AgentRegistryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // ... (existing code)
    
    // Health check agent cards (optional, can be async)
    if registry.Spec.Observability.HealthCheckEnabled {
        checker := health.NewChecker()
        for _, agent := range agents {
            card := // ... get card
            healthy := checker.CheckCard(ctx, card)
            card.Status.EndpointHealthy = &healthy
            r.Status().Update(ctx, card)
        }
    }
}
```

**Success Criteria**:
- [ ] HTTP health checks work
- [ ] Timeouts prevent hanging
- [ ] Health status updated in AgentCard

### Task 2.4.3: Add Namespace Selector Support
**Estimated Time**: 2 hours

Update discovery logic in `go/internal/registry/discovery.go`:

```go
func (d *Discoverer) DiscoverAgents(ctx context.Context, registry *kagentv1alpha1.AgentRegistry) ([]kagentv1alpha1.Agent, error) {
    // Get target namespaces
    namespaces, err := d.getTargetNamespaces(ctx, registry)
    if err != nil {
        return nil, err
    }
    
    var allAgents []kagentv1alpha1.Agent
    for _, ns := range namespaces {
        var agentList kagentv1alpha1.AgentList
        if err := d.List(ctx, &agentList, client.InNamespace(ns)); err != nil {
            return nil, err
        }
        
        for _, agent := range agentList.Items {
            if shouldDiscover(&agent) {
                allAgents = append(allAgents, agent)
            }
        }
    }
    
    return allAgents, nil
}

func (d *Discoverer) getTargetNamespaces(ctx context.Context, registry *kagentv1alpha1.AgentRegistry) ([]string, error) {
    if registry.Spec.Discovery.NamespaceSelector == nil {
        // Default: only watch registry's namespace
        return []string{registry.Namespace}, nil
    }
    
    var nsList corev1.NamespaceList
    selector, err := metav1.LabelSelectorAsSelector(registry.Spec.Discovery.NamespaceSelector)
    if err != nil {
        return nil, err
    }
    
    if err := d.List(ctx, &nsList, client.MatchingLabelsSelector{Selector: selector}); err != nil {
        return nil, err
    }
    
    namespaces := make([]string, len(nsList.Items))
    for i, ns := range nsList.Items {
        namespaces[i] = ns.Name
    }
    
    return namespaces, nil
}
```

**Success Criteria**:
- [ ] Namespace selector filters correctly
- [ ] Multiple namespaces supported
- [ ] Default behavior (same namespace) works

---

## Epic 2.5: Testing & Documentation
**Duration**: 2 days  
**Goal**: Comprehensive testing and user documentation

### Task 2.5.1: Write Integration Tests
**Estimated Time**: 4 hours

**File**: `go/test/integration/registry_integration_test.go`

```go
package integration_test

import (
    "context"
    "testing"
    "time"
    
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    
    kagentv1alpha1 "github.com/kagent-dev/kagent/go/api/v1alpha1"
    "github.com/kagent-dev/kagent/go/pkg/registry/a2a"
)

var _ = Describe("Registry Integration", func() {
    Context("End-to-end A2A workflow", func() {
        It("Should generate valid A2A cards", func() {
            ctx := context.Background()
            
            // Create Agent
            agent := createTestAgent("e2e-agent", "default")
            Expect(k8sClient.Create(ctx, agent)).To(Succeed())
            
            // Create Registry
            registry := createTestRegistry("e2e-registry", "default")
            Expect(k8sClient.Create(ctx, registry)).To(Succeed())
            
            // Wait for AgentCard
            Eventually(func() error {
                card := &kagentv1alpha1.AgentCard{}
                return k8sClient.Get(ctx, 
                    types.NamespacedName{Name: "e2e-agent", Namespace: "default"},
                    card)
            }, "30s", "1s").Should(Succeed())
            
            // Verify A2A card
            card := &kagentv1alpha1.AgentCard{}
            Expect(k8sClient.Get(ctx, 
                types.NamespacedName{Name: "e2e-agent", Namespace: "default"},
                card)).To(Succeed())
            
            Expect(card.Spec.PublicCard).NotTo(BeEmpty())
            
            // Parse and validate A2A JSON
            a2aCard, err := a2a.FromJSON([]byte(card.Spec.PublicCard))
            Expect(err).NotTo(HaveOccurred())
            Expect(a2aCard.Name).To(Equal("e2e-agent"))
        })
    })
})
```

### Task 2.5.2: Create E2E Test Script
**Estimated Time**: 2 hours

**File**: `scripts/test-registry-e2e.sh`

```bash
#!/bin/bash
set -e

echo "=== Agent Registry E2E Test ==="

NAMESPACE="registry-e2e"

echo "Creating test namespace..."
kubectl create namespace $NAMESPACE || true

echo "Deploying AgentRegistry..."
kubectl apply -f - <<EOF
apiVersion: kagent.io/v1alpha1
kind: AgentRegistry
metadata:
  name: test-registry
  namespace: $NAMESPACE
spec:
  discovery:
    enableAutoDiscovery: true
  observability:
    openTelemetryEnabled: true
EOF

echo "Deploying test Agent..."
kubectl apply -f - <<EOF
apiVersion: kagent.io/v1alpha1
kind: Agent
metadata:
  name: test-agent
  namespace: $NAMESPACE
  annotations:
    kagent.io/register-to-registry: "true"
    kagent.io/capabilities: "test,http"
spec:
  description: "Test agent for E2E"
EOF

echo "Waiting for AgentCard..."
kubectl wait --for=condition=Ready agentcard/test-agent \
  -n $NAMESPACE --timeout=60s || true

echo "Verifying AgentCard created..."
kubectl get agentcard test-agent -n $NAMESPACE -o yaml

echo "Checking A2A card content..."
A2A_CARD=$(kubectl get agentcard test-agent -n $NAMESPACE -o jsonpath='{.spec.publicCard}')
echo "$A2A_CARD" | jq .

echo "Testing REST API..."
kubectl port-forward -n kagent svc/kagent-registry-api 8084:8084 &
PF_PID=$!
sleep 2

curl -s http://localhost:8084/api/v1alpha1/registry/cards | jq .
curl -s http://localhost:8084/api/v1alpha1/registry/cards/$NAMESPACE/test-agent/a2a | jq .

kill $PF_PID

echo "Cleanup..."
kubectl delete namespace $NAMESPACE

echo "✅ E2E test passed!"
```

### Task 2.5.3: Update User Documentation
**Estimated Time**: 3 hours

**File**: `docs/registry/user-guide.md` (update)

Add sections:
- A2A card format explanation
- REST API usage examples
- Capability filtering
- Health checking configuration
- Namespace selector examples

### Task 2.5.4: Create API Examples
**Estimated Time**: 1 hour

**File**: `docs/registry/api-examples.md`

```markdown
# Registry API Examples

## List All Agent Cards

```bash
kubectl port-forward -n kagent svc/kagent-registry-api 8084:8084

curl http://localhost:8084/api/v1alpha1/registry/cards
```

## Filter by Namespace

```bash
curl http://localhost:8084/api/v1alpha1/registry/cards?namespace=production
```

## Filter by Capability

```bash
curl http://localhost:8084/api/v1alpha1/registry/cards?capability=kubernetes
```

## Get Specific Agent Card

```bash
curl http://localhost:8084/api/v1alpha1/registry/cards/kagent/my-agent
```

## Get A2A Card Only

```bash
curl http://localhost:8084/api/v1alpha1/registry/cards/kagent/my-agent/a2a | jq .
```

## From Another Pod

```bash
# Using DNS
curl http://kagent-registry-api.kagent.svc.cluster.local:8084/api/v1alpha1/registry/cards
```
```

### Task 2.5.5: Run Full Test Suite
**Estimated Time**: 1 hour

```bash
# Unit tests
cd go
make test

# Integration tests  
make -C go test-integration

# E2E test
./scripts/test-registry-e2e.sh

# Lint
make -C go lint

# Build
make build
```

---

## PR Submission Checklist

### Code Quality
- [ ] All tests pass (`make test`)
- [ ] Lint passes (`make lint`)
- [ ] A2A validation comprehensive
- [ ] API responses match spec
- [ ] OTel spans added to API handlers

### Documentation
- [ ] A2A spec documented
- [ ] API spec complete
- [ ] User guide updated with Phase 2 features
- [ ] API examples provided

### Testing
- [ ] Unit tests >75% coverage
- [ ] Integration tests pass
- [ ] E2E script runs successfully
- [ ] API tested with curl/httpie

### Security
- [ ] API is cluster-internal only (ClusterIP)
- [ ] No authentication bypass
- [ ] Read-only operations enforced
- [ ] No secrets in logs

---

## Success Criteria for Phase 2 Completion

### Functional
- [ ] AgentCards contain valid A2A JSON
- [ ] A2A cards pass validation
- [ ] REST API responds to all endpoints
- [ ] Filtering by namespace/capability works
- [ ] Service endpoints resolved correctly
- [ ] Health checking operational (if enabled)

### Quality
- [ ] Test coverage >75%
- [ ] All integration tests pass
- [ ] E2E test passes in Kind
- [ ] API performance <100ms for list operations
- [ ] No memory leaks in API server

### Documentation
- [ ] A2A specification documented
- [ ] REST API fully documented
- [ ] Usage examples provided
- [ ] Migration guide from Phase 1

---

## Handoff Context for New Thread

Phase 2 adds A2A compliance and REST API access to the registry. The work is organized into 5 epics with 22 tasks.

**Critical Path**:
```
2.1 (A2A Spec) → 2.2 (Translator) → 2.5 (Testing)
                  2.3 (REST API)   ↗
                  2.4 (Discovery)  ↗
```

**Key Deliverables**:
1. A2A types and validator package
2. Translator from AgentCard to A2A
3. REST API server with 4 endpoints
4. Enhanced discovery with Service resolution
5. Comprehensive test suite

**Integration Points**:
- Controller generates A2A cards during reconciliation
- API server reads AgentCards from client
- Translator called automatically for new/updated agents
- Health checker updates AgentCard status

**Testing Strategy**:
- Unit tests for each package
- Integration tests for A2A workflow
- E2E script testing full stack
- API tests using httptest
