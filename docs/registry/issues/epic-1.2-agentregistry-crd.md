# Epic 1.2: AgentRegistry CRD

**Phase**: 1 - Foundation & CRDs  
**Duration**: 2 days  
**Dependencies**: Epic 1.1 (Development Environment Setup)  
**Status**: ✅ Completed

## Overview

Define and generate the AgentRegistry Custom Resource Definition. This CRD will configure how the Agent Registry controller discovers and catalogs agents within the Kubernetes cluster.

## Tasks

- [x] Task 1.2.1: Create AgentRegistry Types (3 hours)
  - Create `go/api/v1alpha1/agentregistry_types.go`
  - Define `AgentRegistrySpec` with selector, discovery, and observability config
  - Define `AgentRegistryStatus` with conditions and metrics
  - Add kubebuilder markers for CRD generation
  - Verify file compiles without errors

- [x] Task 1.2.2: Add Validation Rules (1 hour)
  - Add kubebuilder validation markers
  - Set default values for SyncInterval and A2AVersion
  - Add enum validation for Phase field
  - Verify CRD includes OpenAPI v3 validation schema

- [x] Task 1.2.3: Generate and Review CRD Manifest (30 minutes)
  - Run `make manifests` to generate CRD
  - Review generated YAML in `config/crd/bases/`
  - Verify preserveUnknownFields: false
  - Check printer columns are present
  - Confirm status subresource enabled

- [x] Task 1.2.4: Create Sample YAML (1 hour)
  - Create `go/config/samples/kagent_v1alpha1_agentregistry.yaml`
  - Include examples for different discovery configurations
  - Add comments explaining each field
  - Test applying sample to Kind cluster

## Success Criteria

- [x] AgentRegistry types file compiles without errors
- [x] CRD generated successfully in `go/config/crd/bases/`
- [x] Follows existing kagent v1alpha1 patterns
- [x] Includes proper kubebuilder markers
- [x] OpenAPI v3 validation schema present
- [x] Invalid specs rejected by API server
- [x] Default values applied correctly (a2aVersion: "0.3.0", syncInterval: "5m")
- [x] Sample YAML applies successfully to cluster
- [x] `make -C go fmt vet` passes
- [ ] `make -C go lint` passes (Go version mismatch - golangci-lint built with 1.24, project uses 1.25.1)

## Review Checklist

### Code Quality
- [ ] Follows AGENTS.md style guide
- [ ] Import order: stdlib → external → internal
- [ ] No unnecessary code comments
- [ ] CGO_ENABLED=0 maintained

### CRD Best Practices
- [ ] Spec/status subresources properly defined
- [ ] Kubebuilder markers present and correct
- [ ] OpenAPI v3 validation included
- [ ] Printer columns configured
- [ ] Default values set appropriately
- [ ] Status conditions include ObservedGeneration
- [ ] Namespace-scoped resource

### Documentation
- [ ] Godoc comments for exported types
- [ ] Sample YAML with explanatory comments
- [ ] Field descriptions clear and complete

## Files to Create

1. `go/api/v1alpha1/agentregistry_types.go` - Main type definitions
2. `go/config/crd/bases/kagent.io_agentregistries.yaml` - Generated CRD manifest
3. `go/config/samples/kagent_v1alpha1_agentregistry.yaml` - Sample resource

## References

- Oracle Review: [docs/registry/oracle-review.md](../oracle-review.md)
- Phase 1 Plan: [docs/registry/phase1-execution-plan.md](../phase1-execution-plan.md)
- Existing CRD patterns: `go/api/v1alpha1/agent_types.go`
- Kubebuilder docs: https://book.kubebuilder.io/

## Notes

- Use `kagent.io` as the API group (as seen in oracle review)
- Note: Current cluster uses `kagent.dev` - may need migration path
- Follow server-side apply patterns for future controller use
- Keep CRD simple in Phase 1, extensible for Phase 2+
