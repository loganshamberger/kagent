# Epic 1.4: AgentRegistry Controller

**Phase**: 1 - Foundation & CRDs  
**Duration**: 3 days  
**Dependencies**: Epic 1.2 (AgentRegistry CRD), Epic 1.3 (AgentCard CRD)  
**Status**: ✅ Completed (2025-11-10)

## Overview

Implement the AgentRegistry controller that watches for AgentRegistry resources, discovers annotated agents, generates A2A-compliant AgentCards, and maintains registry status.

## Tasks

- [ ] Task 1.4.1: Scaffold Controller (2 hours)
  - Create `go/internal/controller/agentregistry_controller.go`
  - Implement basic Reconcile function
  - Add RBAC markers
  - Wire up controller in `cmd/controller/main.go`

- [ ] Task 1.4.2: Implement Discovery Package (3 hours)
  - Create `go/internal/registry/discovery/watcher.go`
  - Implement agent discovery based on annotations
  - Support namespace selector filtering
  - Add label selector support
  - Create `go/internal/registry/discovery/resolver.go`
  - Resolve Service endpoints (prefer over Pod IPs)

- [ ] Task 1.4.3: Implement AgentCard Generator (3 hours)
  - Create `go/internal/registry/cardgen/generator.go`
  - Extract capabilities from annotations
  - Build A2A-compliant card structure
  - Create `go/internal/registry/cardgen/hash.go`
  - Implement content hashing (SHA-256)

- [ ] Task 1.4.4: Implement Server-Side Apply (2 hours)
  - Use server-side apply for AgentCard upserts
  - Set field manager: "kagent/agent-registry"
  - Add ownerReferences for garbage collection
  - Skip updates when content hash unchanged

- [ ] Task 1.4.5: Implement Status Updates (2 hours)
  - Update AgentRegistry status conditions
  - Track registeredAgents count
  - Set phase: NotStarted/Discovering/Ready/Error
  - Update lastSync timestamp
  - Set observedGeneration

- [ ] Task 1.4.6: Add OpenTelemetry Tracing (1 hour)
  - Integrate existing OTel tracer
  - Add spans for reconciliation phases
  - Add attributes: registry name/namespace
  - Track discovery and card generation metrics

- [ ] Task 1.4.7: Implement RBAC (1 hour)
  - Generate RBAC manifests via kubebuilder markers
  - Read permissions: pods, services, endpoints, namespaces
  - Read/write permissions: agentcards
  - Read/write permissions: agentregistries/status
  - Copy to `helm/kagent-crds/templates/`

## Success Criteria

- [ ] Controller reconciles AgentRegistry resources
- [ ] Discovers agents with annotation: `kagent.io/register-to-registry: "true"`
- [ ] Creates AgentCards via server-side apply
- [ ] AgentCards include correct A2A metadata
- [ ] Status conditions update properly
- [ ] Content hashing prevents no-op updates
- [ ] OTel spans visible in Jaeger
- [ ] RBAC follows least privilege
- [ ] `make -C go fmt vet lint` passes
- [ ] Unit tests written (>70% coverage)

## Review Checklist

### Code Quality
- [ ] Follows AGENTS.md style guide
- [ ] Import order: stdlib → external → internal
- [ ] Idempotent reconciliation logic
- [ ] Proper error handling and requeueing
- [ ] CGO_ENABLED=0 maintained

### Controller Patterns
- [ ] Uses controller-runtime properly
- [ ] Server-side apply for updates
- [ ] Content hashing to avoid no-ops
- [ ] OwnerReferences for garbage collection
- [ ] No finalizers (relies on garbage collection)
- [ ] OTel tracing integrated
- [ ] MaxConcurrentReconciles set appropriately

### Discovery Logic
- [ ] Annotation-based opt-in: `kagent.io/register-to-registry`
- [ ] Namespace selector respected
- [ ] Label selector respected
- [ ] Prefers Service endpoints over Pod IPs
- [ ] Handles missing/deleted agents gracefully

### A2A Card Generation
- [ ] Capabilities extracted from annotations
- [ ] Endpoints include Service DNS names
- [ ] Version information included
- [ ] Metadata map for extensibility
- [ ] SourceRef tracks origin workload

### RBAC
- [ ] Least privilege principle
- [ ] Separate spec/status permissions
- [ ] Read-only for discovery sources
- [ ] Write access scoped to AgentCards only

### Testing
- [ ] Unit tests with envtest
- [ ] Test annotation discovery
- [ ] Test content hash deduplication
- [ ] Test status condition updates
- [ ] Coverage >70%

## Files to Create

1. `go/internal/controller/agentregistry_controller.go` - Main controller
2. `go/internal/registry/discovery/watcher.go` - Discovery logic
3. `go/internal/registry/discovery/resolver.go` - Endpoint resolution
4. `go/internal/registry/cardgen/generator.go` - Card generation
5. `go/internal/registry/cardgen/hash.go` - Content hashing
6. `go/internal/registry/cardgen/validator.go` - A2A validation
7. `go/config/rbac/agentregistry_*.yaml` - RBAC manifests
8. `go/internal/controller/agentregistry_controller_test.go` - Unit tests

## References

- Oracle Review: [docs/registry/oracle-review.md](../oracle-review.md)
- Phase 1 Plan: [docs/registry/phase1-execution-plan.md](../phase1-execution-plan.md)
- Existing controller patterns: `go/internal/controller/`
- Controller-runtime docs: https://pkg.go.dev/sigs.k8s.io/controller-runtime

## Annotation Constants

```go
const (
    AnnotationRegister     = "kagent.io/register-to-registry"
    AnnotationEndpoint     = "kagent.io/a2a-endpoint"
    AnnotationCapabilities = "kagent.io/capabilities"
    AnnotationDisabled     = "kagent.io/discovery-disabled"
)
```

## Notes

- Use existing OTel tracer: `otel.Tracer("kagent-registry-controller")`
- Requeue interval: 5-10 minutes for debouncing
- MaxConcurrentReconciles: 5 for scale
- Use field indexers for fast annotation lookups
- Prefer Service endpoints to avoid Pod IP flapping
