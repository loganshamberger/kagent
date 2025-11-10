# Phase 2 Planning Summary
## Oracle Analysis & Recommendations

**Date**: 2025-11-10  
**Feature Branch**: `agent-registry` (local)  
**Phase**: Transition from Phase 1 → Phase 2

---

## Executive Summary

The oracle has analyzed the kagent codebase and provided detailed recommendations for **Phase 2: Discovery REST API** implementation. Phase 1 (CRDs + Controller) appears to be in progress with initial implementations already present.

### Key Findings

✅ **Phase 1 Status**: Implementation in progress
- AgentRegistry CRD types defined
- AgentCard CRD types defined  
- AgentRegistry controller scaffolded
- Discovery logic partially implemented

✅ **Phase 2 Recommendations**: Detailed and actionable
- Reuse existing HTTP server infrastructure
- Follow established handler patterns
- Simple, performant API design
- 2-4 day implementation timeline

---

## Oracle-Recommended Phase 2 Approach

### Core Principles

1. **Reuse Existing Infrastructure**
   - Leverage `go/internal/httpserver/server.go`
   - Follow `handlers/` pattern with Base struct
   - Use existing middleware (authn, authz, logging)
   - No new dependencies

2. **Simple & Fast**
   - Read from controller-runtime cache
   - In-memory filtering for <5000 cards
   - Pagination with limit/offset
   - A2A format via query parameter

3. **Kagent-Native**
   - Consistent with `/api/agents` patterns
   - Uses `api.NewResponse()` envelope
   - Same auth/authz mechanisms
   - Compatible middleware stack

---

## API Design (Oracle-Approved)

### Endpoints

```
GET  /api/agentcards                      List all agent cards
GET  /api/agentcards/{namespace}/{name}   Get single agent card
GET  /api/agentcards/{namespace}/{name}/a2a   Get A2A document (optional)
```

### Query Parameters

| Parameter | Type | Purpose | Example |
|-----------|------|---------|---------|
| `namespace` | string | Filter by namespace | `?namespace=prod` |
| `labelSelector` | string | Kubernetes label selector | `?labelSelector=team=platform` |
| `capabilities` | string | Filter by capabilities (AND) | `?capabilities=k8s,helm` |
| `q` | string | Full-text search | `?q=analytics` |
| `limit` | int | Max results (default: 50, max: 500) | `?limit=100` |
| `offset` | int | Pagination offset | `?offset=50` |
| `format` | string | "crd" or "a2a" | `?format=a2a` |

### Response Format

**Standard (CRD)**:
```json
{
  "data": [...AgentCard CRDs...],
  "message": "Successfully listed agent cards",
  "error": false,
  "metadata": {
    "total": 42,
    "limit": 50,
    "offset": 0
  }
}
```

**A2A Format** (`?format=a2a`):
```json
{
  "data": [...A2A documents from card.Spec.PublicCard...],
  "message": "Successfully listed agent cards in A2A format",
  "error": false
}
```

---

## Implementation Plan

### File Structure

```
go/internal/httpserver/
├── server.go                      # ADD: APIPathAgentCards constant
│                                  # MODIFY: setupRoutes() add 3 routes
├── handlers/
│   ├── handlers.go               # ADD: AgentCards field
│   │                             # MODIFY: NewHandlers() wire up
│   └── agentcards.go             # NEW: Full handler implementation
```

### Code Changes

#### 1. server.go
```go
// Add constant
const APIPathAgentCards = "/api/agentcards"

// In setupRoutes()
s.router.HandleFunc(APIPathAgentCards, 
    adaptHandler(s.handlers.AgentCards.HandleListAgentCards)).
    Methods(http.MethodGet)
s.router.HandleFunc(APIPathAgentCards+"/{namespace}/{name}", 
    adaptHandler(s.handlers.AgentCards.HandleGetAgentCard)).
    Methods(http.MethodGet)
s.router.HandleFunc(APIPathAgentCards+"/{namespace}/{name}/a2a", 
    adaptHandler(s.handlers.AgentCards.HandleGetAgentCardA2A)).
    Methods(http.MethodGet)
```

#### 2. handlers.go
```go
type Handlers struct {
    // ... existing ...
    AgentCards *AgentCardsHandler  // NEW
}

func NewHandlers(...) *Handlers {
    // ... existing ...
    return &Handlers{
        // ... existing ...
        AgentCards: NewAgentCardsHandler(base),
    }
}
```

#### 3. agentcards.go (NEW)
See full implementation in [phase2-discovery-api.md](./phase2-discovery-api.md#4-agentcardsgo-new-handler)

**Key Methods**:
- `HandleListAgentCards(w, r)` - List with filtering/pagination
- `HandleGetAgentCard(w, r)` - Get single card
- `HandleGetAgentCardA2A(w, r)` - Convenience A2A endpoint
- `filterAgentCards()` - In-memory filtering
- `toA2AFormat()` - Extract A2A documents

---

## Performance Strategy (Oracle-Validated)

### Current Design (Good for <5000 cards)

```
Request → Parse Query → List from Cache → Filter in Memory → 
Paginate → Format (CRD/A2A) → Respond
```

**Optimizations**:
1. **Cache**: Uses controller-runtime's shared informer cache
2. **Label Selector**: Pushed to API server (fast)
3. **Pagination**: Default limit=50, max=500
4. **Minimal Transforms**: Read A2A from Spec, don't regenerate

### Performance Targets

| Operation | Target | Max Acceptable |
|-----------|--------|----------------|
| List 100 cards | <50ms | <100ms |
| List with labelSelector | <30ms | <80ms |
| Get single card | <10ms | <30ms |
| A2A format conversion | <5ms | <15ms |

### Scaling Path (If Needed)

| Card Count | Strategy |
|------------|----------|
| 100-500 | ✅ Current design (in-memory) |
| 500-5000 | Add field indexers for capabilities |
| 5000+ | SQLite read replica with FTS |

---

## Security & Authorization

### Authentication
```go
// Reuses existing middleware
s.router.Use(auth.AuthnMiddleware(s.authenticator))
```

### Authorization
```go
// List: General access check
Check(h.Authorizer, r, auth.Resource{Type: "AgentCard"})

// Get: Specific resource check
Check(h.Authorizer, r, auth.Resource{
    Type: "AgentCard",
    Name: types.NamespacedName{Namespace: ns, Name: name}.String(),
})
```

### Multi-Tenancy
- Namespace filtering respects RBAC
- Optional: Scope to WatchedNamespaces
- Label selectors for tenant isolation

---

## Testing Requirements

### Unit Tests (`handlers/agentcards_test.go`)
- [ ] List all cards
- [ ] Filter by namespace
- [ ] Filter by labelSelector
- [ ] Filter by capabilities
- [ ] Search query (q parameter)
- [ ] Pagination (limit/offset)
- [ ] A2A format output
- [ ] Get single card
- [ ] Get A2A document
- [ ] Authorization failures
- [ ] Not found errors

### Integration Tests (`test/integration/agentcards_api_test.go`)
- [ ] Full API workflow in envtest
- [ ] Real AgentCard CRDs
- [ ] Cache integration
- [ ] Performance benchmarks

### E2E Tests (`test/e2e/agentregistry_test.go`)
- [ ] Deploy in KinD
- [ ] Create test agents
- [ ] Verify API endpoints
- [ ] A2A compliance validation

---

## Effort Estimate (Oracle-Validated)

| Task | Complexity | Duration |
|------|------------|----------|
| Handler implementation | Medium | 1-2 days |
| Server integration | Small | 0.5 day |
| A2A format handling | Small | 0.5 day |
| Unit tests | Medium | 1 day |
| Integration tests | Small-Medium | 0.5-1 day |
| Documentation | Small | 0.5 day |
| **Total** | **Medium-Large** | **2-4 days** |

---

## Migration Checklist

### Prerequisites from Phase 1
- [x] AgentRegistry CRD exists (`go/api/v1alpha1/agentregistry_types.go`)
- [x] AgentCard CRD exists (`go/api/v1alpha1/agentcard_types.go`)
- [ ] Controller reconciling successfully
- [ ] AgentCards created with Spec.PublicCard populated
- [ ] Status.Hash implemented for deduplication
- [ ] Tests passing

### Phase 2 Implementation Steps
1. [ ] Create `handlers/agentcards.go` with full implementation
2. [ ] Add APIPathAgentCards constant to `server.go`
3. [ ] Add routes in `setupRoutes()`
4. [ ] Register handler in `handlers.go`
5. [ ] Write unit tests
6. [ ] Write integration tests
7. [ ] Test with 100+ sample AgentCards
8. [ ] Verify performance (<100ms)
9. [ ] Documentation and examples
10. [ ] Code review

---

## Key Integration Points

### 1. Existing HTTP Server
```
go/internal/httpserver/server.go
├── Router: *mux.Router (gorilla/mux)
├── Middleware: authn, authz, logging, errors
├── Patterns: adaptHandler(), ErrorResponseWriter
└── Constants: APIPath* for all endpoints
```

### 2. Handler Base Pattern
```go
type Base struct {
    KubeClient         client.Client
    DefaultModelConfig types.NamespacedName
    DatabaseService    database.Client
    Authorizer         auth.Authorizer
}
```
Reuse for AgentCardsHandler.

### 3. Authorization Pattern
```go
func Check(authz auth.Authorizer, r *http.Request, resource auth.Resource) error
```
Used in all handlers.

### 4. Response Pattern
```go
api.NewResponse(data interface{}, message string, isError bool)
RespondWithJSON(w ErrorResponseWriter, status int, data interface{})
```

---

## Oracle's Risk Mitigation Recommendations

### 1. Large List Protection
```go
// Enforce pagination limits
limit := 50  // default
if limit > 500 {
    limit = 500  // hard cap
}
```

### 2. Label Selector Validation
```go
selector, err := labels.Parse(labelSelectorStr)
if err != nil {
    return errors.NewBadRequestError("Invalid labelSelector", err)
}
```

### 3. A2A Document Availability
```go
if format == "a2a" && agentCard.Spec.PublicCard == nil {
    return errors.NewNotFoundError("A2A card not available", nil)
}
```

### 4. Cross-Namespace Leakage Prevention
- Rely on Authorizer for RBAC
- Optional: Filter to WatchedNamespaces if non-empty
- Document multi-tenant considerations

---

## Backwards Compatibility

✅ **No Breaking Changes**
- `/api/agents` - Unchanged (lists Agent CRDs)
- `/api/a2a/{namespace}/{name}` - Unchanged (A2A proxy)
- `/api/agentcards` - New, additive

✅ **Opt-In Discovery**
- Agents need `kagent.io/register-to-registry: "true"`
- No automatic migration

✅ **Independent Versioning**
- v1alpha1 for registry CRDs
- Separate from v1alpha2 Agent CRD
- Conversion webhooks planned for beta

---

## Success Criteria

### Phase 2 Complete When:
- [ ] `/api/agentcards` returns list of AgentCards
- [ ] All query parameters functional (namespace, labelSelector, capabilities, q, limit, offset, format)
- [ ] `format=a2a` returns A2A documents
- [ ] `/api/agentcards/{namespace}/{name}` works
- [ ] Authorization enforced
- [ ] Unit tests >80% coverage
- [ ] Integration tests pass
- [ ] Performance <100ms for 100 cards
- [ ] Documentation complete
- [ ] Code reviewed

### Ready for Phase 3 (Security) When:
- [ ] API tested with 100+ AgentCards
- [ ] External agents can query registry
- [ ] A2A format validated against spec
- [ ] No performance degradation
- [ ] Multi-namespace scenarios tested

---

## Next Actions

### Immediate (This Week)
1. ✅ Oracle review complete
2. ✅ Phase 2 plan documented
3. ⏭️ Verify Phase 1 implementation status
4. ⏭️ Review existing controller code
5. ⏭️ Ensure AgentCard.Spec.PublicCard is populated

### Week 3 (Phase 2 Implementation)
1. ⏭️ Day 1: Create `agentcards.go` handler
2. ⏭️ Day 2: Integrate with HTTP server
3. ⏭️ Day 3: Write tests
4. ⏭️ Day 4: Documentation and review

### Week 4 (Phase 3 Planning)
1. ⏭️ Oracle review for security layer
2. ⏭️ External agent registration design
3. ⏭️ Authentication and authorization strategy

---

## Documentation Index

| Document | Purpose | Status |
|----------|---------|--------|
| [README.md](./README.md) | Overview and index | ✅ Complete |
| [oracle-review.md](./oracle-review.md) | Phase 1 review | ✅ Complete |
| [phase2-discovery-api.md](./phase2-discovery-api.md) | Phase 2 detailed design | ✅ Complete |
| **PHASE2-SUMMARY.md** | **This document** | ✅ Complete |
| [phase1-execution-plan.md](./phase1-execution-plan.md) | Phase 1 tasks | 📝 In Progress |
| [developer-guide.md](./developer-guide.md) | Dev setup and workflows | 📝 In Progress |

---

## Oracle Quotes

> "The feature is feasible in 7–9 weeks if you scope it to: CRDs + controller-driven discovery + A2A card generation + read-only in-cluster API with basic authn via Kubernetes, and incremental OTel hooks."

> "Keep it simple: list from the cached Kube client, filter by namespace/labels/capabilities in-memory, paginate with limit/offset, and return either CRD or A2A JSON via a query flag."

> "Reuse the existing handler Base pattern, controller-runtime cached client, and middleware."

> "Add a new read-only 'AgentCards' REST API under /api/agentcards that lists and retrieves AgentCard CRDs, with built-in filtering, pagination, and optional A2A output."

---

## Conclusion

Phase 2 has a clear, oracle-validated implementation path that:
- ✅ Reuses existing kagent infrastructure
- ✅ Follows established patterns and conventions  
- ✅ Delivers in 2-4 days with 1 engineer
- ✅ Scales to 5000+ agents with simple enhancements
- ✅ Maintains backwards compatibility
- ✅ Provides comprehensive testing strategy

**Status**: Ready to implement upon Phase 1 completion 🚀

---

**Created**: 2025-11-10  
**Oracle Session**: Complete  
**Reviewed By**: Oracle (AI Planning System)  
**Approved For**: Implementation
