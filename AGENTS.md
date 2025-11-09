# AGENTS.md - kagent Development Guide

## Build/Test/Lint Commands
- **Full build**: `make build` (builds all components: controller, UI, app)
- **Go lint**: `make -C go lint` or `cd go && make lint`
- **Go test**: `make -C go test` or `cd go && make test` (unit tests), `make -C go e2e` (e2e tests)
- **Python lint**: `make -C python lint` or `cd python && make lint` (uses ruff)
- **Python test**: `make -C python test` or `cd python && uv run pytest tests packages/*/tests`
- **UI lint**: `cd ui && npm run lint`
- **UI test**: `cd ui && npm run test` (jest), `npm run test:e2e` (cypress)

## Architecture
kagent is a Kubernetes-native AI agent framework with 4 components:
- **Controller** (go/): Kubernetes controller managing Agent, ModelConfig, ToolServer CRDs
- **Engine** (python/): ADK-based agent runtime; UV workspace with multiple packages
- **UI** (ui/): Next.js web UI using React, Tailwind, shadcn/ui
- **CLI** (go/cli/): Command-line REPL tool

## Code Style
- **Go**: Standard golangci-lint rules; imports organized stdlib → external → internal; CGO_ENABLED=0 for builds
- **Python**: Ruff formatting; UV for dependency management; Python 3.13+
- **TypeScript**: Next.js + ESLint; functional components with hooks; colocated tests
- **General**: No code comments unless complex; follow existing patterns; never suppress type errors without reason

## Feature Branch Workflow
- **Current branch**: `agent-registry` (local only, not pushed to remote)
- **Feature**: Agent Registry implementation with A2A-compliant discovery
- **Patterns to follow**: 
  - CRDs in `api/v1alpha1/` with spec/status subresources
  - Controllers use controller-runtime with OTel tracing
  - Server-side apply for AgentCard updates
  - Annotation-based discovery: `kagent.io/register-to-registry: "true"`
  - Status conditions: Ready/Discovering/Error with ObservedGeneration
  - Minimal finalizers (only if managing external artifacts)
  - RBAC: least privilege, separate status/spec permissions
