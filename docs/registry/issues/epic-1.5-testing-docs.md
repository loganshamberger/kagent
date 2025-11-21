# Epic 1.5: Testing & Documentation

**Phase**: 1 - Foundation & CRDs  
**Duration**: 2 days  
**Dependencies**: Epic 1.4 (AgentRegistry Controller)  
**Status**: Not Started

## Overview

Comprehensive testing suite and documentation to validate the Agent Registry implementation and enable users and developers to understand and use the feature.

## Tasks

- [ ] Task 1.5.1: Write Unit Tests (3 hours)
  - Create `go/internal/controller/agentregistry_controller_test.go`
  - Test annotation-based discovery
  - Test AgentCard creation via server-side apply
  - Test content hash deduplication
  - Test status condition updates
  - Use envtest framework
  - Target >70% coverage

- [ ] Task 1.5.2: Write E2E Tests (3 hours)
  - Create `go/test/e2e/agentregistry_test.go`
  - Deploy AgentRegistry to Kind cluster
  - Deploy test Agent with annotation
  - Verify AgentCard creation
  - Verify A2A compliance
  - Test multi-namespace discovery
  - Create bash script for manual E2E validation

- [ ] Task 1.5.3: Create User Documentation (2 hours)
  - Create `docs/registry/user-guide.md`
  - Quick start guide
  - Configuration options reference
  - Annotation reference
  - Example use cases
  - Troubleshooting section

- [ ] Task 1.5.4: Create Developer Documentation (2 hours)
  - Create `docs/registry/developer-guide.md`
  - Architecture diagram (using mermaid)
  - Controller flow diagram
  - Extension points
  - How to add custom discovery sources
  - Testing guide

- [ ] Task 1.5.5: Run Full Test Suite (1 hour)
  - Run `make -C go test`
  - Run `make -C go lint`
  - Run `make -C go manifests`
  - Run E2E tests in Kind cluster
  - Verify builds with `make -C go build`
  - Check test coverage report

## Success Criteria

- [ ] Unit test coverage >70%
- [ ] All unit tests pass
- [ ] E2E tests pass in Kind cluster
- [ ] All linters pass
- [ ] CRDs generate without errors
- [ ] User guide complete with examples
- [ ] Developer guide complete with diagrams
- [ ] Sample YAML files validated
- [ ] No race conditions detected
- [ ] Documentation follows kagent style

## Review Checklist

### Testing
- [ ] Unit tests use envtest framework
- [ ] Tests are deterministic and isolated
- [ ] Coverage >70% for new code
- [ ] E2E tests run in Kind successfully
- [ ] Manual testing documented
- [ ] Edge cases tested (missing annotations, invalid config)

### Documentation Quality
- [ ] User guide has clear quick start
- [ ] All configuration options documented
- [ ] Examples are copy-paste ready
- [ ] Developer guide includes architecture diagrams
- [ ] Troubleshooting section addresses common issues
- [ ] Links to oracle review and execution plan

### Code Quality
- [ ] All tests follow AGENTS.md patterns
- [ ] Import order correct
- [ ] No flaky tests
- [ ] Test cleanup properly handled
- [ ] CGO_ENABLED=0 for test builds

## Files to Create

1. `go/internal/controller/agentregistry_controller_test.go` - Controller unit tests
2. `go/internal/registry/discovery/watcher_test.go` - Discovery unit tests
3. `go/internal/registry/cardgen/generator_test.go` - Card generation unit tests
4. `go/test/e2e/agentregistry_test.go` - E2E test suite
5. `scripts/registry/e2e-test.sh` - Manual E2E validation script
6. `docs/registry/user-guide.md` - User documentation
7. `docs/registry/developer-guide.md` - Developer documentation
8. `docs/registry/architecture.md` - Architecture diagrams

## User Guide Outline

```markdown
# Agent Registry User Guide

## Overview
## Quick Start
  ### 1. Create an AgentRegistry
  ### 2. Annotate Agents for Discovery
  ### 3. View Discovered AgentCards
## Configuration Options
  ### Discovery Configuration
  ### Namespace Selectors
  ### Observability Settings
## Annotation Reference
  ### kagent.io/register-to-registry
  ### kagent.io/capabilities
  ### kagent.io/a2a-endpoint
  ### kagent.io/discovery-disabled
## Examples
  ### Single Namespace Registry
  ### Multi-Namespace Registry
  ### Custom Capabilities
## Troubleshooting
  ### AgentCards Not Created
  ### Discovery Not Working
  ### Status Conditions
```

## Developer Guide Outline

```markdown
# Agent Registry Developer Guide

## Architecture
  [Mermaid diagram of components]
## Controller Flow
  [Sequence diagram of reconciliation]
## Code Organization
  ### CRD Definitions
  ### Controller Logic
  ### Discovery Package
  ### Card Generation Package
## Extending the Registry
  ### Adding Custom Discovery Sources
  ### Customizing A2A Fields
  ### Adding New Annotations
## Testing
  ### Running Unit Tests
  ### Running E2E Tests
  ### Adding New Tests
## Contributing
  [Link to CONTRIBUTING.md]
```

## References

- Oracle Review: [docs/registry/oracle-review.md](../oracle-review.md)
- Phase 1 Plan: [docs/registry/phase1-execution-plan.md](../phase1-execution-plan.md)
- Existing test patterns: `go/internal/controller/*_test.go`
- Envtest docs: https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/envtest

## Test Coverage Goals

- Controller: >80%
- Discovery: >75%
- Card Generation: >85%
- Overall: >70%

## Notes

- Use table-driven tests for multiple scenarios
- Mock external dependencies where appropriate
- E2E tests should be runnable via `make -C go e2e`
- Documentation should reference real sample files
- Include troubleshooting for common user errors
- Architecture diagrams created with mermaid tool
