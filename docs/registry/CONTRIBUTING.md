# Contributing to Agent Registry

This document provides guidelines for contributing to the Agent Registry feature implementation.

## Branch Naming Convention

Use the following naming pattern for feature branches:

```
feat/registry-<epic-number>
```

Examples:
- `feat/registry-1.2` - AgentRegistry CRD
- `feat/registry-1.3` - AgentCard CRD
- `feat/registry-1.4` - Controller implementation

**Note**: The main development branch `agent-registry` remains local-only and is not pushed to GitHub.

## Commit Message Format

Follow [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only changes
- `style`: Code style changes (formatting, missing semicolons, etc.)
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `test`: Adding or updating tests
- `chore`: Changes to build process or auxiliary tools

### Scopes
- `crd`: Custom Resource Definition changes
- `controller`: Controller logic
- `api`: API types and validation
- `docs`: Documentation
- `test`: Test-related changes

### Examples

```
feat(crd): add AgentRegistry CRD with discovery configuration

- Added AgentRegistrySpec with selector and discovery config
- Implemented status subresource with conditions
- Added kubebuilder markers for validation

Refs: Epic 1.2
```

```
test(controller): add unit tests for AgentCard reconciliation

- Test annotation-based discovery
- Test server-side apply updates
- Test content hash deduplication

Coverage: 85%
```

## PR Submission Guidelines

### Before Submitting

1. **Run all checks locally**:
   ```bash
   cd go
   make fmt vet lint
   make test
   make manifests
   ```

2. **Verify CRD generation**:
   ```bash
   cd go
   make manifests
   git diff config/crd/bases/
   ```

3. **Update Helm templates if needed**:
   ```bash
   cp go/config/crd/bases/* helm/kagent-crds/templates/
   ```

4. **Test in Kind cluster**:
   ```bash
   make create-kind-cluster
   make helm-install-provider
   # Apply your changes and verify
   ```

### PR Description Template

```markdown
## Description
Brief description of what this PR does.

## Epic
Epic 1.X: [Epic Name]

## Changes
- Change 1
- Change 2
- Change 3

## Testing
- [ ] Unit tests added/updated
- [ ] Manual testing in Kind cluster
- [ ] E2E tests updated (if applicable)

## Checklist
- [ ] Code follows AGENTS.md style guide
- [ ] Imports organized: stdlib → external → internal
- [ ] `make -C go lint` passes
- [ ] `make -C go test` passes
- [ ] `make -C go manifests` generates valid CRDs
- [ ] CRDs copied to helm/kagent-crds/templates/
- [ ] Documentation updated
- [ ] Sample YAML files provided (if new CRD)
- [ ] No golangci-lint errors
- [ ] CGO_ENABLED=0 for builds

## Oracle Review
Link to oracle review (if applicable): docs/registry/oracle-review.md

## Breaking Changes
List any breaking changes or migration steps required.
```

## Review Checklist

Reviewers should verify:

### Code Quality
- [ ] Follows existing kagent patterns
- [ ] Import order: stdlib → external → internal
- [ ] No code comments unless complex logic requires explanation
- [ ] No suppressed type errors without justification
- [ ] CGO_ENABLED=0 maintained for builds

### CRD Best Practices
- [ ] Spec/status subresources properly defined
- [ ] Kubebuilder markers present and correct
- [ ] OpenAPI v3 validation included
- [ ] Printer columns configured
- [ ] Default values set appropriately
- [ ] Status conditions include ObservedGeneration

### Controller Patterns
- [ ] Uses controller-runtime
- [ ] Idempotent reconciliation logic
- [ ] Server-side apply for updates
- [ ] Content hashing to avoid no-ops
- [ ] OwnerReferences for garbage collection
- [ ] Minimal finalizer use (only if managing external artifacts)
- [ ] OTel tracing integrated
- [ ] Proper error handling and requeueing

### RBAC
- [ ] Least privilege principle followed
- [ ] Separate permissions for spec and status
- [ ] Read-only where appropriate
- [ ] Namespace-scoped where possible

### Testing
- [ ] Unit tests with envtest
- [ ] Coverage >70%
- [ ] E2E tests in Kind
- [ ] Manual testing documented

### Documentation
- [ ] Godoc comments for exported types/functions
- [ ] User-facing documentation updated
- [ ] Sample YAML files provided
- [ ] AGENTS.md updated if new commands added

### Security
- [ ] No secrets in code or logs
- [ ] RBAC properly scoped
- [ ] Input validation present
- [ ] No credential exposure

## Development Workflow

### 1. Start New Work

```bash
# Ensure you're on agent-registry branch
git checkout agent-registry
git status

# Create issue tracking file
mkdir -p docs/registry/issues
cat > docs/registry/issues/epic-X.X-description.md <<EOF
## Epic X.X: [Title]
**Phase**: 1 - Foundation & CRDs
**Duration**: X days
**Dependencies**: [List]

### Tasks
- [ ] Task X.X.1: ...

### Success Criteria
- [ ] ...
EOF
```

### 2. Implement Changes

```bash
# Make your changes following AGENTS.md patterns
# Run tests frequently
cd go
make fmt vet
make test

# Generate CRDs after type changes
make manifests
```

### 3. Test Locally

```bash
# Create/update Kind cluster
make create-kind-cluster
make helm-install-provider

# Apply your changes
kubectl apply -f go/config/crd/bases/
kubectl apply -f go/config/samples/

# Run controller locally
cd go
make run
```

### 4. Prepare for Review

```bash
# Final checks
cd go
make lint
make test
make manifests

# Copy CRDs to Helm
cp config/crd/bases/* ../helm/kagent-crds/templates/

# Commit with conventional commit message
git add .
git commit -m "feat(crd): add AgentRegistry CRD

- Added discovery configuration
- Implemented status conditions
- Added validation rules

Refs: Epic 1.2"
```

### 5. Submit PR (When Ready to Push)

**Note**: During initial development, PRs are NOT created as work remains local-only on the `agent-registry` branch.

When ready to contribute upstream:
1. Consult with maintainers about contribution process
2. Follow their specific PR submission guidelines
3. Reference oracle review: docs/registry/oracle-review.md
4. Include epic tracking documentation

## Getting Help

- Review [AGENTS.md](../../AGENTS.md) for build commands and architecture
- Check [oracle-review.md](oracle-review.md) for design decisions
- See [phase1-execution-plan.md](phase1-execution-plan.md) for task breakdown
- Refer to existing controllers in `go/internal/controller/` for patterns

## Common Issues

### Controller-gen fails
```bash
cd go
make controller-gen
./bin/controller-gen --version  # Should be v0.19.0
```

### Envtest setup fails
```bash
cd go
make setup-envtest
# Check bin/k8s/ directory for binaries
```

### Kind cluster issues
```bash
make delete-kind-cluster
make create-kind-cluster
```

### Import order
Always use this order:
```go
import (
    // Standard library
    "context"
    "fmt"
    
    // External dependencies
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    ctrl "sigs.k8s.io/controller-runtime"
    
    // Internal packages
    kagentv1alpha1 "github.com/kagent-dev/kagent/go/api/v1alpha1"
)
```
