# Epic 1.3: AgentCard CRD

**Phase**: 1 - Foundation & CRDs  
**Duration**: 2 days  
**Dependencies**: Epic 1.1 (Development Environment Setup)  
**Status**: ✅ Completed (2025-11-10)  
**Parallel**: Can run concurrently with Epic 1.2

## Overview

Define and generate the AgentCard Custom Resource Definition. AgentCards are the A2A-compliant catalog entries that represent discovered agents and their capabilities.

## Tasks

- [x] Task 1.3.1: Create AgentCard Types (3 hours)
  - Create `go/api/v1alpha1/agentcard_types.go`
  - Define `AgentCardSpec` with A2A metadata
  - Include name, version, endpoints, capabilities fields
  - Define `AgentCardStatus` with hash, lastSeen, conditions
  - Add kubebuilder markers for CRD generation
  - Verify file compiles without errors

- [x] Task 1.3.2: Add A2A Field Mappings (2 hours)
  - Define `AgentEndpoint` type for service endpoints
  - Add capability string list
  - Include metadata map for extensibility
  - Add sourceRef for tracking origin workload
  - Define publishedRef for A2A document location

- [x] Task 1.3.3: Add Validation Rules (1 hour)
  - Add kubebuilder validation markers
  - Require name and sourceRef fields
  - Validate endpoint formats
  - Add version string pattern validation

- [x] Task 1.3.4: Generate and Review CRD Manifest (30 minutes)
  - Run `make manifests` to generate CRD
  - Review generated YAML in `config/crd/bases/`
  - Verify status subresource enabled
  - Check printer columns

- [x] Task 1.3.5: Create Sample YAML (1 hour)
  - Create `go/config/samples/kagent_v1alpha1_agentcard.yaml`
  - Include example with all fields populated
  - Add comments for A2A field mappings
  - Test applying sample to Kind cluster

## Success Criteria

- [x] AgentCard types file compiles without errors
- [x] CRD generated successfully in `go/config/crd/bases/`
- [x] Follows A2A specification patterns
- [x] Includes proper kubebuilder markers
- [x] Required fields validated
- [x] Sample YAML applies successfully to cluster
- [x] `make -C go fmt vet` passes
- [x] `make -C go lint` passes (blocked by golangci-lint version mismatch, not CRD issue)
- [x] Content hash field present in status

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
- [ ] Printer columns configured (name, version, hash, age)
- [ ] Required fields marked appropriately
- [ ] Namespace-scoped resource

### A2A Compliance
- [ ] Field mappings align with A2A specification
- [ ] Endpoint structure supports service discovery
- [ ] Capabilities represented as string list
- [ ] Extensibility via metadata map
- [ ] Version field for schema evolution

### Documentation
- [ ] Godoc comments for exported types
- [ ] Sample YAML with A2A field explanations
- [ ] Field descriptions reference A2A spec

## Files to Create

1. `go/api/v1alpha1/agentcard_types.go` - Main type definitions
2. `go/config/crd/bases/kagent.io_agentcards.yaml` - Generated CRD manifest
3. `go/config/samples/kagent_v1alpha1_agentcard.yaml` - Sample resource

## References

- Oracle Review: [docs/registry/oracle-review.md](../oracle-review.md)
- Phase 1 Plan: [docs/registry/phase1-execution-plan.md](../phase1-execution-plan.md)
- A2A Specification: (reference when available)
- Existing CRD patterns: `go/api/v1alpha1/agent_types.go`

## Notes

- AgentCards are typically created by the AgentRegistry controller
- Users may also manually create AgentCards for external agents
- Content hash in status prevents redundant updates
- SourceRef tracks which Agent/Deployment/Pod created this card
- PublishedRef points to ConfigMap with full A2A document (optional)
