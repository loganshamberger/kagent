# Phase 1 Execution Plan: Foundation & CRDs
**Agent Registry Implementation - Weeks 1-2**

## Overview
This document provides a detailed, sequenced breakdown of Phase 1 work. Each epic is broken into atomic tasks that can be executed independently and reviewed incrementally.

## Timeline
- **Duration**: 7-9 working days
- **Parallelization**: Epics 1.2 and 1.3 can run concurrently
- **Dependencies**: 1.1 → (1.2 || 1.3) → 1.4 → 1.5

---

## Epic 1.1: Development Environment Setup
**Duration**: 1 day  
**Goal**: Ensure all tooling is ready and establish development workflow

### Task 1.1.1: Verify Local Development Tools
**Estimated Time**: 30 minutes

```bash
# Commands to run
go version                    # Verify Go 1.24.6+
make -C go controller-gen    # Verify controller-gen v0.19.0
make -C go setup-envtest     # Download envtest binaries
kind version                 # Verify Kind cluster support
kubectl version              # Verify kubectl
```

**Success Criteria**:
- [ ] All tools installed with correct versions
- [ ] controller-gen and setup-envtest functional
- [ ] Kind cluster can be created successfully

### Task 1.1.2: Create Feature Branch Workflow Document
**Estimated Time**: 30 minutes

**Files to Update**:
- `docs/registry/CONTRIBUTING.md` (new file)

**Content**:
- PR submission guidelines
- Branch naming conventions: `feat/registry-<epic-number>`
- Commit message format: conventional commits
- Review checklist template

### Task 1.1.3: Set Up Local Testing Environment
**Estimated Time**: 2 hours

```bash
# Test the full local development workflow
make create-kind-cluster
make helm-install-provider
kubectl get crds | grep kagent.io
```

**Success Criteria**:
- [ ] Kind cluster running
- [ ] Existing kagent CRDs deployed
- [ ] Controller running locally via `make -C go run`
- [ ] Can apply sample Agent resources

### Task 1.1.4: Create Epic Tracking Issues
**Estimated Time**: 30 minutes

Create placeholder GitHub issues for:
- Epic 1.2: AgentRegistry CRD
- Epic 1.3: AgentCard CRD
- Epic 1.4: AgentRegistry Controller
- Epic 1.5: Testing & Docs

**Template**:
```markdown
## Epic 1.X: [Title]
**Phase**: 1 - Foundation & CRDs
**Duration**: X days
**Dependencies**: [List]

### Tasks
- [ ] Task 1.X.1: ...
- [ ] Task 1.X.2: ...

### Success Criteria
- [ ] ...

### Review Checklist
- [ ] Code follows AGENTS.md style guide
- [ ] Unit tests added
- [ ] Documentation updated
```

---

## Epic 1.2: AgentRegistry CRD
**Duration**: 2 days  
**Goal**: Define and generate AgentRegistry custom resource

### Task 1.2.1: Create AgentRegistry Types
**Estimated Time**: 3 hours

**File**: `go/api/v1alpha1/agentregistry_types.go`

**Implementation**:
```go
package v1alpha1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AgentRegistrySpec defines the desired state of AgentRegistry
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
}

type DiscoveryConfig struct {
    // EnableAutoDiscovery enables automatic agent discovery
    // +optional
    EnableAutoDiscovery bool `json:"enableAutoDiscovery,omitempty"`
    
    // NamespaceSelector restricts discovery to matching namespaces
    // +optional
    NamespaceSelector *metav1.LabelSelector `json:"namespaceSelector,omitempty"`
    
    // SyncInterval defines how often to sync agent cards (default: 5m)
    // +optional
    SyncInterval *metav1.Duration `json:"syncInterval,omitempty"`
}

type ObservabilityConfig struct {
    // OpenTelemetryEnabled enables OTel tracing/metrics
    // +optional
    OpenTelemetryEnabled bool `json:"openTelemetryEnabled,omitempty"`
}

// AgentRegistryStatus defines the observed state of AgentRegistry
type AgentRegistryStatus struct {
    // RegisteredAgents is the count of active agent cards
    // +optional
    RegisteredAgents int32 `json:"registeredAgents,omitempty"`
    
    // Phase indicates the current lifecycle phase
    // +optional
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
// +kubebuilder:resource:scope=Namespaced,shortName=ar
// +kubebuilder:printcolumn:name="Agents",type=integer,JSONPath=`.status.registeredAgents`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// AgentRegistry is the Schema for the agentregistries API
type AgentRegistry struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`
    
    Spec   AgentRegistrySpec   `json:"spec,omitempty"`
    Status AgentRegistryStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AgentRegistryList contains a list of AgentRegistry
type AgentRegistryList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []AgentRegistry `json:"items"`
}

func init() {
    SchemeBuilder.Register(&AgentRegistry{}, &AgentRegistryList{})
}
```

**Validation**:
```bash
cd go
make manifests  # Should generate CRD without errors
make fmt vet    # Should pass
```

**Success Criteria**:
- [ ] File compiles without errors
- [ ] CRD generated in `go/config/crd/bases/`
- [ ] Follows existing kagent v1alpha1 patterns
- [ ] Includes proper kubebuilder markers

### Task 1.2.2: Add Validation Rules
**Estimated Time**: 1 hour

Add kubebuilder validation markers:
```go
// +kubebuilder:validation:Minimum=1
// +kubebuilder:validation:Maximum=3600
// +kubebuilder:default=300
SyncInterval *metav1.Duration `json:"syncInterval,omitempty"`

// +kubebuilder:validation:Enum=NotStarted;Discovering;Ready;Error
Phase string `json:"phase,omitempty"`
```

**Success Criteria**:
- [ ] CRD includes OpenAPI v3 validation schema
- [ ] Invalid specs rejected by API server
- [ ] Default values applied correctly

### Task 1.2.3: Generate and Review CRD Manifest
**Estimated Time**: 30 minutes

```bash
cd go
make manifests
cat config/crd/bases/kagent.io_agentregistries.yaml
```

**Manual Review**:
- [ ] Verify `preserveUnknownFields: false`
- [ ] Check printer columns are present
- [ ] Confirm status subresource enabled
- [ ] Validate namespace scope

### Task 1.2.4: Create Sample YAML
**Estimated Time**: 1 hour

**File**: `go/config/samples/kagent_v1alpha1_agentregistry.yaml`

```yaml
apiVersion: kagent.io/v1alpha1
kind: AgentRegistry
metadata:
  name: main-registry
  namespace: kagent
spec:
  selector:
    matchLabels:
      kagent.io/agent-enabled: "true"
  discovery:
    enableAutoDiscovery: true
    namespaceSelector:
      matchLabels:
        kagent.io/agent-enabled: "true"
    syncInterval: 5m
  observability:
    openTelemetryEnabled: true
```

**Validation**:
```bash
kubectl --dry-run=server apply -f go/config/samples/kagent_v1alpha1_agentregistry.yaml
```

### Task 1.2.5: Update Helm CRD Templates
**Estimated Time**: 30 minutes

```bash
cd go
make manifests
cp config/crd/bases/kagent.io_agentregistries.yaml ../helm/kagent-crds/templates/
```

**Success Criteria**:
- [ ] CRD copied to helm chart
- [ ] Helm template renders correctly
- [ ] No validation errors

---

## Epic 1.3: AgentCard CRD
**Duration**: 2 days (parallel with 1.2)  
**Goal**: Define AgentCard resource representing discovered agents

### Task 1.3.1: Create AgentCard Types
**Estimated Time**: 3 hours

**File**: `go/api/v1alpha1/agentcard_types.go`

```go
package v1alpha1

import (
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AgentCardSpec defines the desired state of AgentCard
type AgentCardSpec struct {
    // Name of the agent
    Name string `json:"name"`
    
    // Version of the agent
    // +optional
    Version string `json:"version,omitempty"`
    
    // SourceRef references the source workload (Pod, Deployment, etc)
    SourceRef corev1.ObjectReference `json:"sourceRef"`
    
    // Endpoints for reaching the agent (prefer Service over Pod IPs)
    Endpoints []AgentEndpoint `json:"endpoints"`
    
    // Capabilities lists agent capabilities (e.g., ["kubernetes", "helm"])
    // +optional
    Capabilities []string `json:"capabilities,omitempty"`
    
    // A2AVersion is the A2A spec version (e.g., "1.0")
    A2AVersion string `json:"a2aVersion"`
    
    // Metadata for custom extensibility
    // +optional
    Metadata map[string]string `json:"metadata,omitempty"`
    
    // PublicCard is the full A2A-compliant agent card (JSON blob)
    // +optional
    PublicCard string `json:"publicCard,omitempty"`
}

type AgentEndpoint struct {
    // Type of endpoint (http, grpc, etc)
    Type string `json:"type"`
    
    // URL of the endpoint
    URL string `json:"url"`
    
    // Protocol version
    // +optional
    Protocol string `json:"protocol,omitempty"`
}

// AgentCardStatus defines the observed state of AgentCard
type AgentCardStatus struct {
    // Hash of the spec content for deduplication
    Hash string `json:"hash"`
    
    // LastSeen timestamp when agent was last discovered
    LastSeen metav1.Time `json:"lastSeen"`
    
    // EndpointHealthy indicates if endpoint is reachable
    // +optional
    EndpointHealthy *bool `json:"endpointHealthy,omitempty"`
    
    // Conditions
    // +optional
    Conditions []metav1.Condition `json:"conditions,omitempty"`
    
    // PublishedRef points to ConfigMap containing A2A document (if stored separately)
    // +optional
    PublishedRef *corev1.ObjectReference `json:"publishedRef,omitempty"`
    
    // ObservedGeneration
    // +optional
    ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=ac
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`
// +kubebuilder:printcolumn:name="Healthy",type=boolean,JSONPath=`.status.endpointHealthy`
// +kubebuilder:printcolumn:name="LastSeen",type=date,JSONPath=`.status.lastSeen`

// AgentCard is the Schema for the agentcards API
type AgentCard struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`
    
    Spec   AgentCardSpec   `json:"spec,omitempty"`
    Status AgentCardStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AgentCardList contains a list of AgentCard
type AgentCardList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []AgentCard `json:"items"`
}

func init() {
    SchemeBuilder.Register(&AgentCard{}, &AgentCardList{})
}
```

**Success Criteria**:
- [ ] Types compile and generate CRD
- [ ] Includes printer columns for kubectl output
- [ ] Status subresource enabled

### Task 1.3.2: Add Validation and Defaults
**Estimated Time**: 1 hour

```go
// +kubebuilder:validation:MinLength=1
Name string `json:"name"`

// +kubebuilder:validation:Enum=http;grpc;websocket
Type string `json:"type"`

// +kubebuilder:validation:Format=uri
URL string `json:"url"`

// +kubebuilder:default="1.0"
A2AVersion string `json:"a2aVersion"`
```

### Task 1.3.3: Generate CRD and Create Samples
**Estimated Time**: 1 hour

```bash
cd go
make manifests

# Create sample
cat > config/samples/kagent_v1alpha1_agentcard.yaml <<EOF
apiVersion: kagent.io/v1alpha1
kind: AgentCard
metadata:
  name: example-agent
  namespace: kagent
spec:
  name: example-agent
  version: v1.0.0
  sourceRef:
    apiVersion: kagent.io/v1alpha1
    kind: Agent
    name: example-agent
    namespace: kagent
  endpoints:
    - type: http
      url: http://example-agent.kagent.svc.cluster.local:8080
      protocol: a2a/v1
  capabilities:
    - kubernetes
    - observability
  a2aVersion: "1.0"
EOF
```

### Task 1.3.4: Update Helm Templates
**Estimated Time**: 30 minutes

```bash
cp config/crd/bases/kagent.io_agentcards.yaml ../helm/kagent-crds/templates/
```

---

## Epic 1.4: AgentRegistry Controller
**Duration**: 4 days  
**Goal**: Implement controller reconciliation logic

### Task 1.4.1: Create Controller Scaffold
**Estimated Time**: 2 hours

**File**: `go/internal/controller/agentregistry_controller.go`

```go
package controller

import (
    "context"
    
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/log"
    
    kagentv1alpha1 "github.com/kagent-dev/kagent/go/api/v1alpha1"
)

const (
    AnnotationRegister       = "kagent.io/register-to-registry"
    AnnotationEndpoint       = "kagent.io/a2a-endpoint"
    AnnotationCapabilities   = "kagent.io/capabilities"
    AnnotationDiscoveryDisabled = "kagent.io/discovery-disabled"
)

var tracer = otel.Tracer("kagent-registry-controller")

// AgentRegistryReconciler reconciles an AgentRegistry object
type AgentRegistryReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=kagent.io,resources=agentregistries,verbs=get;list;watch
// +kubebuilder:rbac:groups=kagent.io,resources=agentregistries/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kagent.io,resources=agentcards,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kagent.io,resources=agentcards/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kagent.io,resources=agents,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=pods;services;namespaces,verbs=get;list;watch

func (r *AgentRegistryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    ctx, span := tracer.Start(ctx, "AgentRegistry.Reconcile")
    defer span.End()
    
    log := log.FromContext(ctx)
    
    span.SetAttributes(
        attribute.String("registry.name", req.Name),
        attribute.String("registry.namespace", req.Namespace),
    )
    
    // Fetch AgentRegistry
    var registry kagentv1alpha1.AgentRegistry
    if err := r.Get(ctx, req.NamespacedName, &registry); err != nil {
        span.SetStatus(codes.Error, "failed to get AgentRegistry")
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }
    
    log.Info("Reconciling AgentRegistry", "name", registry.Name)
    
    // TODO: Implement discovery and AgentCard management
    
    // Update status
    if err := r.updateStatus(ctx, &registry); err != nil {
        span.SetStatus(codes.Error, err.Error())
        return ctrl.Result{}, err
    }
    
    span.SetStatus(codes.Ok, "Success")
    return ctrl.Result{RequeueAfter: registry.Spec.Discovery.SyncInterval.Duration}, nil
}

func (r *AgentRegistryReconciler) updateStatus(ctx context.Context, registry *kagentv1alpha1.AgentRegistry) error {
    // Update observedGeneration and conditions
    registry.Status.ObservedGeneration = registry.Generation
    
    // TODO: Set conditions
    
    return r.Status().Update(ctx, registry)
}

func (r *AgentRegistryReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&kagentv1alpha1.AgentRegistry{}).
        Owns(&kagentv1alpha1.AgentCard{}).
        Complete(r)
}
```

**Wire into manager** in `go/cmd/controller/main.go`:
```go
if err = (&controller.AgentRegistryReconciler{
    Client: mgr.GetClient(),
    Scheme: mgr.GetScheme(),
}).SetupWithManager(mgr); err != nil {
    setupLog.Error(err, "unable to create controller", "controller", "AgentRegistry")
    os.Exit(1)
}
```

### Task 1.4.2: Implement Agent Discovery
**Estimated Time**: 4 hours

Create discovery helper:

**File**: `go/internal/registry/discovery.go`

```go
package registry

import (
    "context"
    "strings"
    
    kagentv1alpha1 "github.com/kagent-dev/kagent/go/api/v1alpha1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "sigs.k8s.io/controller-runtime/pkg/client"
)

type Discoverer struct {
    client.Client
}

func NewDiscoverer(c client.Client) *Discoverer {
    return &Discoverer{Client: c}
}

// DiscoverAgents finds all Agents matching the registry's selector
func (d *Discoverer) DiscoverAgents(ctx context.Context, registry *kagentv1alpha1.AgentRegistry) ([]kagentv1alpha1.Agent, error) {
    var agentList kagentv1alpha1.AgentList
    
    listOpts := []client.ListOption{}
    
    // Apply namespace selector if specified
    if registry.Spec.Discovery.NamespaceSelector != nil {
        // TODO: Implement namespace filtering
    }
    
    if err := d.List(ctx, &agentList, listOpts...); err != nil {
        return nil, err
    }
    
    // Filter by annotation
    var discovered []kagentv1alpha1.Agent
    for _, agent := range agentList.Items {
        if shouldDiscover(&agent) {
            discovered = append(discovered, agent)
        }
    }
    
    return discovered, nil
}

func shouldDiscover(agent *kagentv1alpha1.Agent) bool {
    annotations := agent.GetAnnotations()
    if annotations == nil {
        return false
    }
    
    // Check if registration is enabled
    if annotations["kagent.io/register-to-registry"] != "true" {
        return false
    }
    
    // Check if explicitly disabled
    if annotations["kagent.io/discovery-disabled"] == "true" {
        return false
    }
    
    return true
}

func parseCapabilities(agent *kagentv1alpha1.Agent) []string {
    annotations := agent.GetAnnotations()
    if annotations == nil {
        return nil
    }
    
    caps := annotations["kagent.io/capabilities"]
    if caps == "" {
        return nil
    }
    
    return strings.Split(caps, ",")
}
```

### Task 1.4.3: Implement AgentCard Generation
**Estimated Time**: 4 hours

**File**: `go/internal/registry/cardgen.go`

```go
package registry

import (
    "crypto/sha256"
    "encoding/json"
    "fmt"
    
    kagentv1alpha1 "github.com/kagent-dev/kagent/go/api/v1alpha1"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CardGenerator struct {
    Namespace string
}

func NewCardGenerator(namespace string) *CardGenerator {
    return &CardGenerator{Namespace: namespace}
}

// GenerateCard creates an AgentCard from an Agent
func (g *CardGenerator) GenerateCard(agent *kagentv1alpha1.Agent) *kagentv1alpha1.AgentCard {
    card := &kagentv1alpha1.AgentCard{
        ObjectMeta: metav1.ObjectMeta{
            Name:      agent.Name,
            Namespace: agent.Namespace,
            OwnerReferences: []metav1.OwnerReference{
                *metav1.NewControllerRef(agent, kagentv1alpha1.GroupVersion.WithKind("Agent")),
            },
        },
        Spec: kagentv1alpha1.AgentCardSpec{
            Name:         agent.Name,
            Version:      getVersion(agent),
            SourceRef:    makeObjectReference(agent),
            Endpoints:    resolveEndpoints(agent),
            Capabilities: parseCapabilities(agent),
            A2AVersion:   "1.0",
            Metadata:     extractMetadata(agent),
        },
    }
    
    // Calculate hash
    hash := calculateHash(card)
    card.Status.Hash = hash
    card.Status.LastSeen = metav1.Now()
    
    return card
}

func calculateHash(card *kagentv1alpha1.AgentCard) string {
    data, _ := json.Marshal(card.Spec)
    hash := sha256.Sum256(data)
    return fmt.Sprintf("%x", hash)
}

func getVersion(agent *kagentv1alpha1.Agent) string {
    if v, ok := agent.Annotations["kagent.io/version"]; ok {
        return v
    }
    return "v1.0.0"
}

func makeObjectReference(agent *kagentv1alpha1.Agent) corev1.ObjectReference {
    return corev1.ObjectReference{
        APIVersion: kagentv1alpha1.GroupVersion.String(),
        Kind:       "Agent",
        Name:       agent.Name,
        Namespace:  agent.Namespace,
        UID:        agent.UID,
    }
}

func resolveEndpoints(agent *kagentv1alpha1.Agent) []kagentv1alpha1.AgentEndpoint {
    // Check for custom endpoint annotation
    if endpoint, ok := agent.Annotations["kagent.io/a2a-endpoint"]; ok {
        return []kagentv1alpha1.AgentEndpoint{
            {
                Type:     "http",
                URL:      endpoint,
                Protocol: "a2a/v1",
            },
        }
    }
    
    // Default: construct from agent name
    url := fmt.Sprintf("http://%s.%s.svc.cluster.local:8080", agent.Name, agent.Namespace)
    return []kagentv1alpha1.AgentEndpoint{
        {
            Type:     "http",
            URL:      url,
            Protocol: "a2a/v1",
        },
    }
}

func extractMetadata(agent *kagentv1alpha1.Agent) map[string]string {
    metadata := make(map[string]string)
    
    // Copy relevant annotations
    for k, v := range agent.Annotations {
        if strings.HasPrefix(k, "kagent.io/metadata-") {
            key := strings.TrimPrefix(k, "kagent.io/metadata-")
            metadata[key] = v
        }
    }
    
    return metadata
}
```

### Task 1.4.4: Implement AgentCard Upsert Logic
**Estimated Time**: 3 hours

Update controller to use discovery and card generation:

```go
func (r *AgentRegistryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // ... (existing setup code)
    
    // Discover agents
    discoverer := registry.NewDiscoverer(r.Client)
    agents, err := discoverer.DiscoverAgents(ctx, &registry)
    if err != nil {
        return ctrl.Result{}, fmt.Errorf("failed to discover agents: %w", err)
    }
    
    // Generate and upsert agent cards
    generator := registry.NewCardGenerator(registry.Namespace)
    for _, agent := range agents {
        if err := r.upsertAgentCard(ctx, generator, &agent); err != nil {
            log.Error(err, "failed to upsert AgentCard", "agent", agent.Name)
            continue
        }
    }
    
    // Update registry status
    registry.Status.RegisteredAgents = int32(len(agents))
    registry.Status.LastSync = &metav1.Time{Time: time.Now()}
    registry.Status.Phase = "Ready"
    
    // ... (update status)
}

func (r *AgentRegistryReconciler) upsertAgentCard(ctx context.Context, gen *registry.CardGenerator, agent *kagentv1alpha1.Agent) error {
    newCard := gen.GenerateCard(agent)
    
    // Check if card already exists
    existing := &kagentv1alpha1.AgentCard{}
    err := r.Get(ctx, client.ObjectKeyFromObject(newCard), existing)
    
    if err == nil {
        // Card exists, check if update needed
        if existing.Status.Hash == newCard.Status.Hash {
            // No change, skip update
            return nil
        }
    }
    
    // Server-side apply
    return r.Patch(ctx, newCard, client.Apply, 
        client.ForceOwnership, 
        client.FieldOwner("kagent/agent-registry"))
}
```

### Task 1.4.5: Add Status Conditions
**Estimated Time**: 2 hours

```go
func (r *AgentRegistryReconciler) setReadyCondition(registry *kagentv1alpha1.AgentRegistry, status metav1.ConditionStatus, reason, message string) {
    condition := metav1.Condition{
        Type:               "Ready",
        Status:             status,
        ObservedGeneration: registry.Generation,
        LastTransitionTime: metav1.Now(),
        Reason:             reason,
        Message:            message,
    }
    
    meta.SetStatusCondition(&registry.Status.Conditions, condition)
}

// In Reconcile:
if err != nil {
    r.setReadyCondition(&registry, metav1.ConditionFalse, "ReconciliationFailed", err.Error())
    // ...
} else {
    r.setReadyCondition(&registry, metav1.ConditionTrue, "ReconciliationSucceeded", "Successfully discovered agents")
}
```

---

## Epic 1.5: Testing & Documentation
**Duration**: 2 days  
**Goal**: Validate implementation and document usage

### Task 1.5.1: Write Unit Tests
**Estimated Time**: 4 hours

**File**: `go/internal/controller/agentregistry_controller_test.go`

```go
package controller_test

import (
    "context"
    "time"
    
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/types"
    
    kagentv1alpha1 "github.com/kagent-dev/kagent/go/api/v1alpha1"
)

var _ = Describe("AgentRegistry Controller", func() {
    const (
        timeout  = time.Second * 10
        interval = time.Millisecond * 250
    )
    
    Context("When creating an AgentRegistry", func() {
        It("Should create AgentCards for annotated Agents", func() {
            ctx := context.Background()
            
            // Create Agent with annotation
            agent := &kagentv1alpha1.Agent{
                ObjectMeta: metav1.ObjectMeta{
                    Name:      "test-agent",
                    Namespace: "default",
                    Annotations: map[string]string{
                        "kagent.io/register-to-registry": "true",
                    },
                },
                Spec: kagentv1alpha1.AgentSpec{
                    Description: "Test agent",
                },
            }
            Expect(k8sClient.Create(ctx, agent)).To(Succeed())
            
            // Create AgentRegistry
            registry := &kagentv1alpha1.AgentRegistry{
                ObjectMeta: metav1.ObjectMeta{
                    Name:      "test-registry",
                    Namespace: "default",
                },
                Spec: kagentv1alpha1.AgentRegistrySpec{
                    Discovery: kagentv1alpha1.DiscoveryConfig{
                        EnableAutoDiscovery: true,
                    },
                },
            }
            Expect(k8sClient.Create(ctx, registry)).To(Succeed())
            
            // Wait for AgentCard to be created
            card := &kagentv1alpha1.AgentCard{}
            Eventually(func() error {
                return k8sClient.Get(ctx, 
                    types.NamespacedName{Name: "test-agent", Namespace: "default"}, 
                    card)
            }, timeout, interval).Should(Succeed())
            
            // Verify card contents
            Expect(card.Spec.Name).To(Equal("test-agent"))
            Expect(card.Status.Hash).NotTo(BeEmpty())
        })
    })
})
```

### Task 1.5.2: Write E2E Tests
**Estimated Time**: 3 hours

**File**: `go/test/e2e/agentregistry_test.go`

```bash
#!/bin/bash
# E2E test script

set -e

echo "Creating test namespace..."
kubectl create namespace registry-e2e-test || true

echo "Deploying AgentRegistry..."
kubectl apply -f - <<EOF
apiVersion: kagent.io/v1alpha1
kind: AgentRegistry
metadata:
  name: e2e-registry
  namespace: registry-e2e-test
spec:
  discovery:
    enableAutoDiscovery: true
EOF

echo "Deploying test Agent..."
kubectl apply -f - <<EOF
apiVersion: kagent.io/v1alpha1
kind: Agent
metadata:
  name: e2e-agent
  namespace: registry-e2e-test
  annotations:
    kagent.io/register-to-registry: "true"
    kagent.io/capabilities: "test,demo"
spec:
  description: "E2E test agent"
EOF

echo "Waiting for AgentCard creation..."
kubectl wait --for=condition=Ready agentcard/e2e-agent \
  -n registry-e2e-test --timeout=60s

echo "Verifying AgentCard contents..."
kubectl get agentcard e2e-agent -n registry-e2e-test -o yaml

echo "E2E test passed!"
```

### Task 1.5.3: Create User Documentation
**Estimated Time**: 2 hours

**File**: `docs/registry/user-guide.md`

```markdown
# Agent Registry User Guide

## Overview
The Agent Registry provides automatic discovery and cataloging of kagent Agents within your Kubernetes cluster.

## Quick Start

### 1. Create an AgentRegistry
```yaml
apiVersion: kagent.io/v1alpha1
kind: AgentRegistry
metadata:
  name: main-registry
  namespace: kagent
spec:
  discovery:
    enableAutoDiscovery: true
```

### 2. Annotate Agents for Discovery
```yaml
apiVersion: kagent.io/v1alpha1
kind: Agent
metadata:
  name: my-agent
  annotations:
    kagent.io/register-to-registry: "true"
    kagent.io/capabilities: "kubernetes,monitoring"
spec:
  description: "My custom agent"
```

### 3. View Discovered AgentCards
```bash
kubectl get agentcards -n kagent
kubectl describe agentcard my-agent
```

## Configuration Options
[... detailed documentation ...]
```

### Task 1.5.4: Create Developer Documentation
**Estimated Time**: 2 hours

**File**: `docs/registry/developer-guide.md`

```markdown
# Agent Registry Developer Guide

## Architecture
[Diagram of components]

## Controller Flow
1. Watch for AgentRegistry changes
2. Discover Agents matching selector + annotation
3. Generate AgentCard for each discovered Agent
4. Upsert AgentCard using server-side apply
5. Update AgentRegistry status

## Extending the Registry
[How to add custom discovery sources, A2A fields, etc.]
```

### Task 1.5.5: Run Full Test Suite
**Estimated Time**: 1 hour

```bash
# Unit tests
cd go
make test

# Lint
make lint

# Generate manifests
make manifests

# E2E tests (requires Kind cluster)
make e2e

# Verify builds
make build
```

---

## PR Submission Checklist

Before submitting PRs, verify:

### Code Quality
- [ ] `make -C go lint` passes
- [ ] `make -C go test` passes with >70% coverage
- [ ] `make -C go manifests` generates valid CRDs
- [ ] No golangci-lint errors
- [ ] Follows AGENTS.md style guide (stdlib → external → internal imports)

### Documentation
- [ ] User-facing documentation updated
- [ ] Code includes godoc comments for exported types/functions
- [ ] Sample YAML files provided
- [ ] AGENTS.md updated if new commands added

### Testing
- [ ] Unit tests added for new code
- [ ] E2E test scenario documented
- [ ] Manual testing in Kind cluster performed

### CRD & Manifests
- [ ] CRDs copied to `helm/kagent-crds/templates/`
- [ ] `make controller-manifests` runs successfully
- [ ] RBAC rules updated if needed

### Git
- [ ] Commits follow conventional commit format
- [ ] Branch name follows `feat/registry-<epic>`
- [ ] PR description references oracle review

---

## Success Criteria for Phase 1 Completion

### Functional
- [ ] AgentRegistry CRD deployed and validated
- [ ] AgentCard CRD deployed and validated
- [ ] Controller reconciles AgentRegistry resources
- [ ] Annotated Agents automatically create AgentCards
- [ ] AgentCards include correct A2A metadata
- [ ] Status conditions update properly
- [ ] Server-side apply works without conflicts

### Quality
- [ ] Unit test coverage >70%
- [ ] E2E tests pass in Kind cluster
- [ ] All linters pass
- [ ] No race conditions detected
- [ ] OTel spans visible in Jaeger (if enabled)

### Documentation
- [ ] User guide published
- [ ] Developer guide published
- [ ] Sample YAML files available
- [ ] AGENTS.md reflects new capabilities

---

## Handoff Context for New Thread

This plan breaks Phase 1 into 5 epics with 23 discrete tasks. The work should be executed sequentially with the following dependencies:

**Critical Path**:
```
1.1 (Setup) → 1.2 (AgentRegistry CRD) → 1.4 (Controller) → 1.5 (Testing)
              1.3 (AgentCard CRD)    ↗
```

**Parallelization Opportunities**:
- Tasks 1.2 and 1.3 can run concurrently
- Testing tasks (1.5.1, 1.5.2) can run in parallel

**Key Files to Create**:
1. `go/api/v1alpha1/agentregistry_types.go`
2. `go/api/v1alpha1/agentcard_types.go`
3. `go/internal/controller/agentregistry_controller.go`
4. `go/internal/registry/discovery.go`
5. `go/internal/registry/cardgen.go`
6. `go/config/samples/*.yaml`
7. Test files and documentation

**Reference Materials**:
- Oracle review: `docs/registry/oracle-review.md`
- Existing patterns: Check `go/internal/controller/` for similar controllers
- AGENTS.md for build/test commands
- Makefile for codegen workflows
