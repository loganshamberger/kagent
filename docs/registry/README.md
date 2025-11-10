# Agent Registry Documentation

This directory contains the design and implementation documentation for the Agent Registry feature in kagent.

---

## Overview

The Agent Registry is a Kubernetes-native system for discovering, cataloging, and exposing AI agents in an A2A-compliant format. It enables agents to find and communicate with each other both within and across clusters.

**Status**: 🚧 In Development (Feature Branch: `agent-registry`)  
**Timeline**: 7-9 weeks (Phased implementation)  
**Contributing to**: CNCF kagent project

---

## Documentation Index

### Planning & Reviews

| Document | Description | Status |
|----------|-------------|--------|
| [Oracle Review](./oracle-review.md) | Initial plan review with technical recommendations | ✅ Complete |
| [Phase 2 Plan](./phase2-discovery-api.md) | Discovery REST API implementation guide | ✅ Complete |

### Implementation Guides

| Phase | Description | Duration | Status |
|-------|-------------|----------|--------|
| **Phase 1** | CRDs + Controller + Discovery | 2 weeks | 📋 Planned |
| **Phase 2** | Discovery REST API | 2-4 days | 📋 Planned |
| **Phase 3** | Security & Authentication | 1 week | 📋 Planned |
| **Phase 4** | External Agent Registration | 1 week | 📋 Planned |
| **Phase 5** | Testing & Documentation | 2 weeks | 📋 Planned |

---

## Quick Start (After Phase 1 & 2 Complete)

### Discover All Agents
```bash
curl http://localhost:8083/api/agentcards
```

### Filter by Capability
```bash
curl "http://localhost:8083/api/agentcards?capabilities=kubernetes,monitoring"
```

### Get A2A Format
```bash
curl "http://localhost:8083/api/agentcards?format=a2a"
```

### Get Single Agent
```bash
curl http://localhost:8083/api/agentcards/kagent/k8s-agent
```

---

## Architecture

### Components

```
┌─────────────────────────────────────────────────────────────────┐
│                        Kagent Controller                         │
│                                                                   │
│  ┌───────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │ Agent         │  │ AgentRegistry│  │ AgentCard            │  │
│  │ Controller    │  │ Controller   │  │ Controller (opt)     │  │
│  └───────┬───────┘  └──────┬───────┘  └──────────────────────┘  │
│          │                  │                                     │
│          │     ┌────────────▼─────────────┐                     │
│          │     │ Discovery Watcher        │                     │
│          └────►│ - Annotations            │                     │
│                │ - Service Endpoints       │                     │
│                │ - A2A Card Generation    │                     │
│                └────────────┬──────────────┘                     │
│                             │                                     │
│                             ▼                                     │
│                ┌────────────────────────────┐                    │
│                │ AgentCard CRDs             │                    │
│                │ (with A2A PublicCard)      │                    │
│                └────────────┬───────────────┘                    │
└─────────────────────────────┼───────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      HTTP Server (Port 8083)                     │
│                                                                   │
│  /api/agents          │  Existing Agent CRD API                  │
│  /api/agentcards      │  NEW: AgentCard Discovery API            │
│  /api/a2a/{ns}/{name} │  Existing A2A Proxy                      │
│                                                                   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ Middleware: Authn → Authz → Logging → Error Handling    │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

### Data Flow

```
1. Agent Deployed
   └─> Agent CRD created with annotation: kagent.io/register-to-registry=true

2. AgentRegistry Controller
   └─> Discovers annotated Agent
   └─> Resolves Service endpoint
   └─> Generates A2A-compliant card
   └─> Creates/Updates AgentCard CRD (with hash for deduplication)

3. REST API Request
   └─> GET /api/agentcards
   └─> Reads from cached client (controller-runtime)
   └─> Filters & paginates
   └─> Returns CRD or A2A format

4. Agent Discovery
   └─> Other agents query /api/agentcards?capabilities=X
   └─> Get A2A endpoint URLs
   └─> Communicate via /api/a2a/{namespace}/{name}
```

---

## Custom Resource Definitions

### AgentRegistry
```yaml
apiVersion: kagent.io/v1alpha1
kind: AgentRegistry
metadata:
  name: main-registry
  namespace: kagent
spec:
  discovery:
    enableAutoDiscovery: true
    namespaceSelector:
      matchLabels:
        kagent.io/agent-enabled: "true"
  observability:
    openTelemetryEnabled: true
```

**Purpose**: Configures agent discovery behavior and namespace scope.

### AgentCard
```yaml
apiVersion: kagent.io/v1alpha1
kind: AgentCard
metadata:
  name: k8s-agent
  namespace: kagent
  ownerReferences:
    - apiVersion: kagent.io/v1alpha2
      kind: Agent
      name: k8s-agent
      uid: ...
spec:
  agentName: k8s-agent
  agentType: internal
  sourceRef:
    kind: Agent
    name: k8s-agent
    namespace: kagent
  capabilities:
    - kubernetes
    - deployment
    - monitoring
  publicCard:
    name: k8s-agent
    version: "1.0"
    url: "http://k8s-agent.kagent.svc.cluster.local:8080"
    skills:
      - id: k8s-deploy
        name: "Kubernetes Deployment"
        description: "Deploy applications to Kubernetes"
    authentication:
      - type: http
        scheme: bearer
status:
  phase: Ready
  hash: abc123...
  lastUpdated: "2025-11-09T10:30:00Z"
  endpointHealthy: true
```

**Purpose**: Represents a discovered agent with A2A card data.

---

## API Reference

### Endpoints (Phase 2)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/agentcards` | List all agent cards with filtering |
| GET | `/api/agentcards/{namespace}/{name}` | Get single agent card |
| GET | `/api/agentcards/{namespace}/{name}/a2a` | Get A2A document directly |

### Query Parameters

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `namespace` | string | Filter by namespace | `?namespace=production` |
| `labelSelector` | string | Kubernetes label selector | `?labelSelector=team=platform` |
| `capabilities` | string | Comma-separated capabilities (AND) | `?capabilities=k8s,helm` |
| `q` | string | Full-text search | `?q=analytics` |
| `limit` | int | Max results (default: 50, max: 500) | `?limit=100` |
| `offset` | int | Pagination offset | `?offset=50` |
| `format` | string | Output format: "crd" or "a2a" | `?format=a2a` |

---

## Development Workflow

### Feature Branch
```bash
# Current branch (local only, not pushed)
git branch
# * agent-registry

# View commits
git log --oneline
# ad2ff5e Add Phase 2 implementation plan
# e92c91f Add oracle review
# 2323796 Add AGENTS.md
```

### Implementation Phases

#### Phase 1: Foundation (Weeks 1-2)
**Goal**: CRDs, Controller, Auto-discovery

**Tasks**:
1. ✅ Design review (oracle)
2. ⏭️ Create AgentRegistry CRD
3. ⏭️ Create AgentCard CRD
4. ⏭️ Implement AgentRegistry controller
5. ⏭️ Implement discovery watcher
6. ⏭️ Implement A2A card generator
7. ⏭️ Add OpenTelemetry tracing
8. ⏭️ Write unit + integration tests

**Deliverables**:
- Working CRDs
- Controller auto-discovering agents
- AgentCards created with A2A data
- Tests passing

#### Phase 2: Discovery API (Week 3)
**Goal**: REST API for agent discovery

**Tasks**:
1. ⏭️ Create AgentCardsHandler
2. ⏭️ Add routes to HTTP server
3. ⏭️ Implement filtering and pagination
4. ⏭️ Add A2A format support
5. ⏭️ Write handler unit tests
6. ⏭️ Integration tests

**Deliverables**:
- `/api/agentcards` endpoint functional
- Filtering and pagination working
- A2A format output
- Performance <100ms for 100 cards

---

## Key Design Decisions

### 1. **Annotation-Based Discovery**
**Decision**: Use `kagent.io/register-to-registry: "true"` for opt-in  
**Rationale**: Explicit, safe, prevents accidental cluster-wide scans  
**Pattern**: Follows Kagent annotation conventions

### 2. **Server-Side Apply for AgentCards**
**Decision**: Use SSA with field manager `kagent/agent-registry`  
**Rationale**: Idempotent, conflict-free, standard Kubernetes pattern  
**Benefit**: Multiple controllers can update different fields

### 3. **Content Hashing for Deduplication**
**Decision**: SHA256 hash of A2A card in Status.Hash  
**Rationale**: Avoid no-op updates, reduce reconciliation churn  
**Benefit**: Controller skips update if hash unchanged

### 4. **Service Endpoints Over Pod IPs**
**Decision**: Prefer Service DNS names  
**Rationale**: Stable, survives pod restarts, standard Kubernetes  
**Fallback**: Pod IP for headless services

### 5. **Read-Only REST API (Phase 2)**
**Decision**: List/Get only, no POST/PUT/DELETE  
**Rationale**: Simplicity, security, consistency with existing APIs  
**Future**: POST for external agent registration in Phase 4

### 6. **In-Memory Filtering (Phase 2)**
**Decision**: Filter/paginate in-memory from cached client  
**Rationale**: Fast for <5000 cards, no external dependencies  
**Future**: Field indexers or SQLite for >5000 cards

---

## Testing Strategy

### Unit Tests
- CRD validation
- Controller reconciliation logic
- Card generation and hashing
- Handler filtering and pagination
- A2A format conversion

### Integration Tests
- Full controller reconcile cycle
- AgentCard creation from annotated Agents
- REST API endpoints
- Authorization checks

### E2E Tests
- Deploy in KinD cluster
- Create test agents
- Verify auto-discovery
- Query via REST API
- Validate A2A compliance

### Performance Tests
- 100 agents: <100ms list
- 500 agents: <200ms list
- 5000 agents: benchmark for scaling decisions

---

## Performance & Scalability

### Current Design (Phase 2)
- **Target**: 100-500 AgentCards
- **List Latency**: <100ms
- **Strategy**: Controller-runtime cache + in-memory filters
- **Pagination**: Required for >500 results

### Scaling Path (If Needed)
- **500-5000 cards**: Add field indexers
- **5000+ cards**: SQLite materialized view
- **Heavy search**: FTS5 full-text search
- **External traffic**: CDN + ETag caching

---

## Security Considerations

### Authentication (Reused from Kagent)
- Middleware: `auth.AuthnMiddleware(s.authenticator)`
- Interface: `auth.AuthProvider`
- Current: `UnsecureAuthenticator` (dev only)

### Authorization (Reused from Kagent)
- Interface: `auth.Authorizer`
- Per-request checks: `Check(authorizer, request, resource)`
- Future: RBAC integration

### Multi-Tenancy
- Namespace scoping via RBAC
- Label-based filtering
- Optional: Restrict to `WatchedNamespaces`

### A2A Security (Phase 3+)
- mTLS between agents
- OAuth2/OIDC for external agents
- Webhook validation for AgentCards
- Rate limiting and quotas

---

## Migration & Backwards Compatibility

### Compatibility Guarantees
✅ **No Breaking Changes**:
- Existing `/api/agents` unchanged
- Existing `/api/a2a/{namespace}/{name}` unchanged
- New `/api/agentcards` is additive
- All existing Agent CRDs continue to work

✅ **Opt-In**:
- Agents must have annotation to be registered
- No automatic migration of existing agents
- AgentRegistry creation is explicit

✅ **CRD Versioning**:
- v1alpha1 allows schema evolution
- Conversion webhooks planned for beta
- Field deprecation follows Kubernetes standards

---

## Observability

### OpenTelemetry Integration
- **Tracer**: `otel.Tracer("kagent-registry-controller")`
- **Spans**: Reconcile, Discovery, CardGeneration
- **Attributes**: registry.name, agent.name, namespace

### Metrics (Planned)
```
kagent_registry_agent_cards_total{namespace, phase}
kagent_registry_discovery_errors_total{namespace, reason}
kagent_registry_reconcile_duration_seconds{operation}
kagent_registry_api_requests_total{endpoint, status}
kagent_registry_api_latency_seconds{endpoint}
```

### Logging
- Structured logging with `logr`
- Levels: Info, Debug, Error
- Context: registry, agent, operation

---

## Contributing

### Code Review Checklist
- [ ] Follows Kagent code conventions (AGENTS.md)
- [ ] Unit tests with >70% coverage
- [ ] Integration tests pass
- [ ] OTel tracing added
- [ ] RBAC annotations correct
- [ ] Status conditions with ObservedGeneration
- [ ] Documentation updated
- [ ] No golangci-lint errors

### PR Structure (Small, Reviewable)
1. CRDs + Types + Codegen
2. Controller Scaffold + RBAC
3. Discovery + AgentCard Upsert
4. A2A Translator + Publishing
5. REST API Handlers
6. OTel Spans + Metrics
7. Tests + Examples + Docs

---

## Resources

### Internal Documentation
- [AGENTS.md](../../AGENTS.md) - Development guide
- [Oracle Review](./oracle-review.md) - Initial plan review
- [Phase 2 Plan](./phase2-discovery-api.md) - REST API design

### Kagent Codebase
- [HTTP Server](../../go/internal/httpserver/server.go)
- [Handlers Pattern](../../go/internal/httpserver/handlers/)
- [Existing Controllers](../../go/internal/controller/)
- [A2A Integration](../../go/internal/a2a/)

### External Specs
- [Agent2Agent Specification](https://github.com/agent2agent/spec)
- [Kubernetes API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)
- [Kubebuilder Book](https://book.kubebuilder.io/)

---

## Status & Next Steps

### Current Status
- ✅ Feature branch created: `agent-registry`
- ✅ Oracle review complete
- ✅ Phase 1 plan detailed
- ✅ Phase 2 plan detailed
- ⏭️ Begin Phase 1 implementation

### Immediate Next Steps
1. Share design documents with Kagent maintainers
2. Get feedback on CRD schemas
3. Begin CRD implementation
4. Set up local development environment
5. Create first draft PR with CRDs

### Long-Term Roadmap
- **Week 1-2**: Phase 1 (CRDs + Controller)
- **Week 3**: Phase 2 (REST API)
- **Week 4**: Phase 3 (Security)
- **Week 5-6**: Phase 4 (External Agents)
- **Week 7-9**: Phase 5 (Testing + Docs + Review)

---

## Contact & Support

- **Feature Branch**: `agent-registry` (local)
- **Repository**: kagent-dev/kagent (CNCF)
- **Discord**: [Kagent Community](https://discord.gg/Fu3k65f2k3)
- **Issues**: Track in GitHub project board

---

**Last Updated**: 2025-11-09  
**Version**: 0.1.0-alpha  
**Status**: 📋 Planning Complete, Ready for Implementation
