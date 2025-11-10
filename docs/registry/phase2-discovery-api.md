# Phase 2: Discovery API - Implementation Plan
## Agent Registry REST API for kagent

**Date**: 2025-11-09  
**Phase**: 2 of 5 (Weeks 3-4)  
**Dependencies**: Phase 1 Complete (CRDs + Controller)  
**Reviewer**: Oracle (AI Planning & Review System)

---

## Executive Summary

Phase 2 adds a **read-only REST API** for agent discovery that:
- Lists and retrieves AgentCard CRDs with filtering and pagination
- Supports A2A format output via query parameter
- Integrates seamlessly with kagent's existing HTTP server
- Reuses established patterns (handlers, middleware, auth)
- Optimized for 100-500 agents with room to scale to 5000+

**Timeline**: 2-4 days (1 engineer)  
**Effort**: Medium  
**Risk Level**: Low (reuses existing infrastructure)

---

## API Design

### Endpoints

#### 1. List AgentCards
```http
GET /api/agentcards
```

**Query Parameters**:
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `namespace` | string | all | Filter by namespace |
| `labelSelector` | string | none | Kubernetes-style label selector |
| `capabilities` | string | none | Comma-separated; match if card has all |
| `q` | string | none | Substring search (name, metadata, capabilities) |
| `limit` | int | 50 | Max results (max: 500) |
| `offset` | int | 0 | Pagination offset |
| `format` | string | "crd" | Output format: "crd" or "a2a" |

**Examples**:
```bash
# List all agent cards
GET /api/agentcards

# Filter by namespace
GET /api/agentcards?namespace=production

# Search by capability
GET /api/agentcards?capabilities=kubernetes,monitoring

# Label selector
GET /api/agentcards?labelSelector=kagent.io/team=platform

# Full-text search
GET /api/agentcards?q=analytics

# A2A format output
GET /api/agentcards?format=a2a

# Pagination
GET /api/agentcards?limit=20&offset=40

# Combined filters
GET /api/agentcards?namespace=prod&capabilities=helm&format=a2a&limit=10
```

**Response** (format=crd):
```json
{
  "data": [
    {
      "metadata": {
        "name": "k8s-agent",
        "namespace": "kagent"
      },
      "spec": {
        "agentName": "k8s-agent",
        "agentType": "internal",
        "publicCard": {
          "name": "k8s-agent",
          "version": "1.0",
          "url": "http://k8s-agent.kagent.svc.cluster.local:8080",
          "capabilities": {
            "streaming": true
          },
          "skills": [
            {
              "id": "k8s-deploy",
              "name": "Kubernetes Deployment",
              "description": "Deploy applications to Kubernetes"
            }
          ]
        }
      },
      "status": {
        "phase": "Ready",
        "hash": "abc123...",
        "lastUpdated": "2025-11-09T10:30:00Z"
      }
    }
  ],
  "message": "Successfully listed agent cards",
  "error": false,
  "metadata": {
    "total": 42,
    "limit": 50,
    "offset": 0
  }
}
```

**Response** (format=a2a):
```json
{
  "data": [
    {
      "name": "k8s-agent",
      "version": "1.0",
      "url": "http://k8s-agent.kagent.svc.cluster.local:8080",
      "capabilities": {
        "streaming": true,
        "pushNotifications": false
      },
      "defaultInputModes": ["text/plain", "application/json"],
      "defaultOutputModes": ["text/plain", "application/json"],
      "skills": [
        {
          "id": "k8s-deploy",
          "name": "Kubernetes Deployment",
          "description": "Deploy applications to Kubernetes",
          "tags": ["kubernetes", "deployment"]
        }
      ],
      "authentication": [
        {
          "type": "http",
          "scheme": "bearer"
        }
      ]
    }
  ],
  "message": "Successfully listed agent cards in A2A format",
  "error": false
}
```

---

#### 2. Get Single AgentCard
```http
GET /api/agentcards/{namespace}/{name}
```

**Query Parameters**:
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `format` | string | "crd" | Output format: "crd" or "a2a" |

**Examples**:
```bash
# Get AgentCard as CRD
GET /api/agentcards/kagent/k8s-agent

# Get A2A format
GET /api/agentcards/kagent/k8s-agent?format=a2a
```

**Response** (format=crd):
```json
{
  "data": {
    "metadata": {
      "name": "k8s-agent",
      "namespace": "kagent"
    },
    "spec": { ... },
    "status": { ... }
  },
  "message": "Successfully retrieved agent card",
  "error": false
}
```

**Response** (format=a2a):
```json
{
  "data": {
    "name": "k8s-agent",
    "version": "1.0",
    "url": "http://k8s-agent.kagent.svc.cluster.local:8080",
    ...
  },
  "message": "Successfully retrieved agent card in A2A format",
  "error": false
}
```

---

#### 3. Get A2A Document (Optional Convenience)
```http
GET /api/agentcards/{namespace}/{name}/a2a
```

**Response**:
```json
{
  "data": {
    "name": "k8s-agent",
    "version": "1.0",
    ...
  },
  "message": "Successfully retrieved A2A document",
  "error": false
}
```

---

## Architecture Integration

### Files Changed/Added

```
go/
├── internal/
│   └── httpserver/
│       ├── server.go                    # MODIFY: Add APIPathAgentCards constant
│       │                                 # MODIFY: Add routes in setupRoutes()
│       └── handlers/
│           ├── handlers.go              # MODIFY: Add AgentCards field
│           └── agentcards.go            # NEW: AgentCardsHandler implementation
│
└── api/
    └── v1alpha1/
        ├── agentcard_types.go           # Already exists from Phase 1
        └── agentregistry_types.go       # Already exists from Phase 1
```

### Code Changes

#### 1. server.go (Constants)
```go
// Add to existing constants
const (
    // ... existing constants ...
    APIPathAgentCards = "/api/agentcards"
)
```

#### 2. server.go (Route Setup)
```go
func (s *HTTPServer) setupRoutes() {
    // ... existing routes ...
    
    // AgentCards - NEW
    s.router.HandleFunc(APIPathAgentCards, 
        adaptHandler(s.handlers.AgentCards.HandleListAgentCards)).
        Methods(http.MethodGet)
    s.router.HandleFunc(APIPathAgentCards+"/{namespace}/{name}", 
        adaptHandler(s.handlers.AgentCards.HandleGetAgentCard)).
        Methods(http.MethodGet)
    // Optional convenience endpoint
    s.router.HandleFunc(APIPathAgentCards+"/{namespace}/{name}/a2a", 
        adaptHandler(s.handlers.AgentCards.HandleGetAgentCardA2A)).
        Methods(http.MethodGet)
    
    // ... middleware setup ...
}
```

#### 3. handlers.go (Registry)
```go
type Handlers struct {
    // ... existing handlers ...
    AgentCards      *AgentCardsHandler  // NEW
}

func NewHandlers(...) *Handlers {
    // ... existing code ...
    
    return &Handlers{
        // ... existing handlers ...
        AgentCards:      NewAgentCardsHandler(base),  // NEW
    }
}
```

#### 4. agentcards.go (NEW Handler)
```go
package handlers

import (
    "context"
    "net/http"
    "strconv"
    "strings"
    
    "github.com/gorilla/mux"
    registryv1alpha1 "github.com/kagent-dev/kagent/go/api/v1alpha1"
    "github.com/kagent-dev/kagent/go/internal/httpserver/errors"
    "github.com/kagent-dev/kagent/go/pkg/auth"
    "github.com/kagent-dev/kagent/go/pkg/client/api"
    k8serrors "k8s.io/apimachinery/pkg/api/errors"
    "k8s.io/apimachinery/pkg/labels"
    "k8s.io/apimachinery/pkg/types"
    "sigs.k8s.io/controller-runtime/pkg/client"
    ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// AgentCardsHandler handles agent card discovery requests
type AgentCardsHandler struct {
    *Base
}

// NewAgentCardsHandler creates a new AgentCardsHandler
func NewAgentCardsHandler(base *Base) *AgentCardsHandler {
    return &AgentCardsHandler{Base: base}
}

// HandleListAgentCards handles GET /api/agentcards
func (h *AgentCardsHandler) HandleListAgentCards(w ErrorResponseWriter, r *http.Request) {
    log := ctrllog.FromContext(r.Context()).
        WithName("agentcards-handler").
        WithValues("operation", "list")
    
    // Authorization check
    if err := Check(h.Authorizer, r, auth.Resource{Type: "AgentCard"}); err != nil {
        w.RespondWithError(err)
        return
    }
    
    // Parse query parameters
    query := r.URL.Query()
    namespace := query.Get("namespace")
    labelSelectorStr := query.Get("labelSelector")
    capabilitiesStr := query.Get("capabilities")
    searchQuery := query.Get("q")
    format := query.Get("format")
    if format == "" {
        format = "crd"
    }
    
    // Pagination
    limit := 50
    if limitStr := query.Get("limit"); limitStr != "" {
        if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
            limit = l
            if limit > 500 {
                limit = 500
            }
        }
    }
    
    offset := 0
    if offsetStr := query.Get("offset"); offsetStr != "" {
        if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
            offset = o
        }
    }
    
    // Build list options
    listOpts := []client.ListOption{}
    
    // Namespace filter
    if namespace != "" {
        listOpts = append(listOpts, client.InNamespace(namespace))
    }
    
    // Label selector
    if labelSelectorStr != "" {
        selector, err := labels.Parse(labelSelectorStr)
        if err != nil {
            w.RespondWithError(errors.NewBadRequestError(
                "Invalid labelSelector", err))
            return
        }
        listOpts = append(listOpts, client.MatchingLabelsSelector{Selector: selector})
    }
    
    // List from cached client
    agentCardList := &registryv1alpha1.AgentCardList{}
    if err := h.KubeClient.List(r.Context(), agentCardList, listOpts...); err != nil {
        log.Error(err, "Failed to list AgentCards")
        w.RespondWithError(errors.NewInternalServerError(
            "Failed to list agent cards", err))
        return
    }
    
    // In-memory filtering
    filtered := h.filterAgentCards(agentCardList.Items, 
        capabilitiesStr, searchQuery)
    
    // Pagination
    total := len(filtered)
    start := offset
    if start > total {
        start = total
    }
    end := start + limit
    if end > total {
        end = total
    }
    
    pagedCards := filtered[start:end]
    
    // Format response
    var responseData interface{}
    if format == "a2a" {
        responseData = h.toA2AFormat(pagedCards)
    } else {
        responseData = pagedCards
    }
    
    log.Info("Successfully listed agent cards", 
        "total", total, "returned", len(pagedCards))
    
    response := api.NewResponse(responseData, 
        "Successfully listed agent cards", false)
    
    // Add metadata
    if respMap, ok := response.(map[string]interface{}); ok {
        respMap["metadata"] = map[string]interface{}{
            "total":  total,
            "limit":  limit,
            "offset": offset,
        }
    }
    
    RespondWithJSON(w, http.StatusOK, response)
}

// HandleGetAgentCard handles GET /api/agentcards/{namespace}/{name}
func (h *AgentCardsHandler) HandleGetAgentCard(w ErrorResponseWriter, r *http.Request) {
    log := ctrllog.FromContext(r.Context()).
        WithName("agentcards-handler").
        WithValues("operation", "get")
    
    // Parse path parameters
    vars := mux.Vars(r)
    namespace := vars["namespace"]
    name := vars["name"]
    
    // Authorization check
    if err := Check(h.Authorizer, r, auth.Resource{
        Type: "AgentCard",
        Name: types.NamespacedName{Namespace: namespace, Name: name}.String(),
    }); err != nil {
        w.RespondWithError(err)
        return
    }
    
    // Get format
    format := r.URL.Query().Get("format")
    if format == "" {
        format = "crd"
    }
    
    // Fetch AgentCard
    agentCard := &registryv1alpha1.AgentCard{}
    key := types.NamespacedName{Namespace: namespace, Name: name}
    
    if err := h.KubeClient.Get(r.Context(), key, agentCard); err != nil {
        if k8serrors.IsNotFound(err) {
            w.RespondWithError(errors.NewNotFoundError(
                "AgentCard not found", err))
        } else {
            log.Error(err, "Failed to get AgentCard")
            w.RespondWithError(errors.NewInternalServerError(
                "Failed to retrieve agent card", err))
        }
        return
    }
    
    // Format response
    var responseData interface{}
    if format == "a2a" {
        if agentCard.Spec.PublicCard == nil {
            w.RespondWithError(errors.NewNotFoundError(
                "A2A card not available for this agent", nil))
            return
        }
        responseData = agentCard.Spec.PublicCard
    } else {
        responseData = agentCard
    }
    
    log.Info("Successfully retrieved agent card", 
        "namespace", namespace, "name", name)
    
    RespondWithJSON(w, http.StatusOK, 
        api.NewResponse(responseData, "Successfully retrieved agent card", false))
}

// HandleGetAgentCardA2A handles GET /api/agentcards/{namespace}/{name}/a2a
func (h *AgentCardsHandler) HandleGetAgentCardA2A(w ErrorResponseWriter, r *http.Request) {
    // Reuse HandleGetAgentCard with forced a2a format
    q := r.URL.Query()
    q.Set("format", "a2a")
    r.URL.RawQuery = q.Encode()
    h.HandleGetAgentCard(w, r)
}

// filterAgentCards applies in-memory filters
func (h *AgentCardsHandler) filterAgentCards(
    cards []registryv1alpha1.AgentCard,
    capabilitiesStr string,
    searchQuery string,
) []registryv1alpha1.AgentCard {
    
    filtered := make([]registryv1alpha1.AgentCard, 0, len(cards))
    
    // Parse capabilities
    var requiredCapabilities []string
    if capabilitiesStr != "" {
        requiredCapabilities = strings.Split(capabilitiesStr, ",")
        for i := range requiredCapabilities {
            requiredCapabilities[i] = strings.TrimSpace(requiredCapabilities[i])
        }
    }
    
    for _, card := range cards {
        // Capabilities filter
        if len(requiredCapabilities) > 0 {
            if !hasAllCapabilities(card.Spec.Capabilities, requiredCapabilities) {
                continue
            }
        }
        
        // Search query filter
        if searchQuery != "" {
            if !matchesSearchQuery(card, searchQuery) {
                continue
            }
        }
        
        filtered = append(filtered, card)
    }
    
    return filtered
}

// hasAllCapabilities checks if card has all required capabilities
func hasAllCapabilities(cardCapabilities []string, required []string) bool {
    capMap := make(map[string]bool)
    for _, cap := range cardCapabilities {
        capMap[cap] = true
    }
    
    for _, req := range required {
        if !capMap[req] {
            return false
        }
    }
    return true
}

// matchesSearchQuery performs substring search
func matchesSearchQuery(card registryv1alpha1.AgentCard, query string) bool {
    query = strings.ToLower(query)
    
    // Search in name
    if strings.Contains(strings.ToLower(card.Name), query) {
        return true
    }
    
    // Search in capabilities
    for _, cap := range card.Spec.Capabilities {
        if strings.Contains(strings.ToLower(cap), query) {
            return true
        }
    }
    
    // Search in metadata
    for k, v := range card.Spec.Metadata {
        if strings.Contains(strings.ToLower(k), query) ||
           strings.Contains(strings.ToLower(v), query) {
            return true
        }
    }
    
    // Search in A2A card fields
    if card.Spec.PublicCard != nil {
        if strings.Contains(strings.ToLower(card.Spec.PublicCard.Name), query) ||
           strings.Contains(strings.ToLower(card.Spec.PublicCard.Description), query) {
            return true
        }
        
        for _, skill := range card.Spec.PublicCard.Skills {
            if strings.Contains(strings.ToLower(skill.Name), query) ||
               strings.Contains(strings.ToLower(skill.Description), query) {
                return true
            }
        }
    }
    
    return false
}

// toA2AFormat extracts A2A documents from AgentCards
func (h *AgentCardsHandler) toA2AFormat(
    cards []registryv1alpha1.AgentCard,
) []interface{} {
    result := make([]interface{}, 0, len(cards))
    
    for _, card := range cards {
        if card.Spec.PublicCard != nil {
            result = append(result, card.Spec.PublicCard)
        }
    }
    
    return result
}
```

---

## Testing Strategy

### Unit Tests

**File**: `go/internal/httpserver/handlers/agentcards_test.go`

```go
package handlers

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    
    registryv1alpha1 "github.com/kagent-dev/kagent/go/api/v1alpha1"
    "github.com/gorilla/mux"
    "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestHandleListAgentCards(t *testing.T) {
    tests := []struct {
        name           string
        cards          []registryv1alpha1.AgentCard
        query          string
        expectedCount  int
        expectedStatus int
    }{
        {
            name: "list all",
            cards: []registryv1alpha1.AgentCard{
                makeTestCard("default", "agent1", []string{"k8s"}),
                makeTestCard("default", "agent2", []string{"helm"}),
            },
            query:          "",
            expectedCount:  2,
            expectedStatus: http.StatusOK,
        },
        {
            name: "filter by capability",
            cards: []registryv1alpha1.AgentCard{
                makeTestCard("default", "agent1", []string{"k8s", "monitoring"}),
                makeTestCard("default", "agent2", []string{"helm"}),
            },
            query:          "?capabilities=k8s",
            expectedCount:  1,
            expectedStatus: http.StatusOK,
        },
        {
            name: "pagination",
            cards: make10TestCards(),
            query:          "?limit=5&offset=3",
            expectedCount:  5,
            expectedStatus: http.StatusOK,
        },
        {
            name: "search query",
            cards: []registryv1alpha1.AgentCard{
                makeTestCardWithName("default", "k8s-agent", []string{}),
                makeTestCardWithName("default", "helm-agent", []string{}),
            },
            query:          "?q=k8s",
            expectedCount:  1,
            expectedStatus: http.StatusOK,
        },
        {
            name: "a2a format",
            cards: []registryv1alpha1.AgentCard{
                makeTestCardWithA2A("default", "agent1"),
            },
            query:          "?format=a2a",
            expectedCount:  1,
            expectedStatus: http.StatusOK,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup fake client
            fakeClient := fake.NewClientBuilder().
                WithLists(&registryv1alpha1.AgentCardList{Items: tt.cards}).
                Build()
            
            base := &Base{KubeClient: fakeClient}
            handler := NewAgentCardsHandler(base)
            
            // Create request
            req := httptest.NewRequest("GET", "/api/agentcards"+tt.query, nil)
            req = req.WithContext(context.Background())
            w := httptest.NewRecorder()
            
            // Execute
            handler.HandleListAgentCards(w, req)
            
            // Verify
            if w.Code != tt.expectedStatus {
                t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
            }
            
            // Parse response and verify count
            // ... (additional assertions)
        })
    }
}

func TestHandleGetAgentCard(t *testing.T) {
    // ... similar test structure
}

// Helper functions
func makeTestCard(ns, name string, capabilities []string) registryv1alpha1.AgentCard {
    return registryv1alpha1.AgentCard{
        ObjectMeta: metav1.ObjectMeta{
            Name:      name,
            Namespace: ns,
        },
        Spec: registryv1alpha1.AgentCardSpec{
            Capabilities: capabilities,
        },
    }
}
```

### Integration Tests

**File**: `test/integration/agentcards_api_test.go`

```go
package integration

import (
    "context"
    "io"
    "net/http"
    "testing"
    
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
)

var _ = Describe("AgentCards API", func() {
    var (
        ctx    context.Context
        apiURL string
    )
    
    BeforeEach(func() {
        ctx = context.Background()
        apiURL = "http://localhost:8083/api/agentcards"
    })
    
    It("should list agent cards", func() {
        resp, err := http.Get(apiURL)
        Expect(err).NotTo(HaveOccurred())
        defer resp.Body.Close()
        
        Expect(resp.StatusCode).To(Equal(http.StatusOK))
        
        body, _ := io.ReadAll(resp.Body)
        Expect(body).To(ContainSubstring("data"))
    })
    
    It("should filter by capabilities", func() {
        resp, err := http.Get(apiURL + "?capabilities=kubernetes")
        Expect(err).NotTo(HaveOccurred())
        defer resp.Body.Close()
        
        Expect(resp.StatusCode).To(Equal(http.StatusOK))
    })
    
    It("should return A2A format", func() {
        resp, err := http.Get(apiURL + "?format=a2a")
        Expect(err).NotTo(HaveOccurred())
        defer resp.Body.Close()
        
        Expect(resp.StatusCode).To(Equal(http.StatusOK))
        // Verify A2A schema compliance
    })
})
```

---

## Performance Considerations

### Optimization Strategy

1. **Leverage Controller-Runtime Cache**
   - All List/Get operations use the shared informer cache
   - No database queries or external calls
   - Sub-millisecond response for cached data

2. **Pagination Enforcement**
   - Default: 50 results
   - Maximum: 500 results
   - Clients must paginate for larger sets

3. **Filter Pipeline**
   ```
   List (with labelSelector) → In-memory capability filter → 
   Search query filter → Sort → Paginate → Format
   ```

4. **Label Selector Optimization**
   - Pushed down to Kubernetes API server
   - Reduces data transfer and in-memory processing
   - Encourage users to use labels over search

5. **Content Hashing (Phase 1)**
   - AgentCard Status.Hash prevents unnecessary updates
   - Reduces churn in the cache

### Performance Targets

| Metric | Target | Max Acceptable |
|--------|--------|----------------|
| List 100 cards (no filter) | <50ms | <100ms |
| List 100 cards (labelSelector) | <30ms | <80ms |
| List 100 cards (capabilities filter) | <60ms | <120ms |
| Get single card | <10ms | <30ms |
| Format conversion (A2A) | <5ms | <15ms |

### Scaling Limits

| Cards | Strategy |
|-------|----------|
| <500 | Current design (in-memory filtering) |
| 500-5000 | Add field indexers for capabilities |
| 5000+ | Consider SQLite materialized view |

---

## Migration Path from Phase 1

### Prerequisites
✅ Phase 1 complete:
- AgentRegistry CRD deployed
- AgentCard CRD deployed
- Controller reconciling and creating AgentCards
- AgentCard.Spec.PublicCard populated with A2A data

### Integration Steps

1. **Week 3, Day 1**: Handler Implementation
   - Create `agentcards.go` handler
   - Implement list and get methods
   - Write unit tests

2. **Week 3, Day 2**: Server Integration
   - Add routes to `server.go`
   - Update `handlers.go` registry
   - Test locally with `make run`

3. **Week 3, Day 3**: Testing & Refinement
   - Integration tests with KinD cluster
   - Performance testing with 100-500 sample cards
   - Fix any issues

4. **Week 3, Day 4**: Documentation & Review
   - API documentation
   - Example usage
   - Code review prep

### Backwards Compatibility

✅ **No Breaking Changes**:
- Existing `/api/agents` unchanged
- Existing `/api/a2a/{namespace}/{name}` unchanged
- New `/api/agentcards` is additive
- All existing CRDs compatible

---

## Security & Authorization

### Authentication
- Reuses existing `auth.AuthProvider` middleware
- Same authentication flow as other API endpoints
- No new auth mechanisms required

### Authorization
```go
// List: Check general access
Check(h.Authorizer, r, auth.Resource{Type: "AgentCard"})

// Get: Check specific resource
Check(h.Authorizer, r, auth.Resource{
    Type: "AgentCard",
    Name: types.NamespacedName{Namespace: ns, Name: name}.String(),
})
```

### Multi-Tenancy Considerations
- Namespace filtering respects RBAC
- labelSelector can be used to scope to tenant labels
- Optional: Restrict to `WatchedNamespaces` if configured

---

## Effort Estimate

| Task | Effort | Duration |
|------|--------|----------|
| Handler implementation | Medium | 1-2 days |
| Server integration | Small | 0.5 day |
| A2A format handling | Small | 0.5 day |
| Unit tests | Medium | 1 day |
| Integration tests | Small-Medium | 0.5-1 day |
| Documentation | Small | 0.5 day |
| **Total** | **Medium-Large** | **2-4 days** |

---

## Success Criteria

### Phase 2 Complete When:
- [ ] `/api/agentcards` endpoint returns list of AgentCards
- [ ] Filtering works: namespace, labelSelector, capabilities, search
- [ ] Pagination with limit/offset functional
- [ ] `format=a2a` returns A2A-compliant documents
- [ ] `/api/agentcards/{namespace}/{name}` returns single card
- [ ] Authorization checks enforced
- [ ] Unit tests >80% coverage
- [ ] Integration tests pass
- [ ] Performance targets met (<100ms for 100 cards)
- [ ] Documentation complete
- [ ] Code reviewed and merged to feature branch

### Ready for Phase 3 When:
- [ ] API tested with 100+ real AgentCards
- [ ] No performance degradation
- [ ] External agents can discover internal agents via API
- [ ] A2A format validated against spec

---

## Future Enhancements (Not Phase 2)

### Advanced Path (if >5k cards or heavy traffic)
- **Field Indexers**: Index capabilities and metadata keys
- **ETag Support**: Cache control with If-None-Match
- **SQLite Read Replica**: Materialized view for complex queries
- **Full-Text Search**: FTS5 for advanced search
- **GraphQL**: Alternative query interface
- **Content Negotiation**: Accept: application/vnd.a2a+json

### Phase 3+ Features
- External agent registration API (POST /api/agentcards)
- Webhook validation for AgentCard
- Health checks for registered agents
- Agent discovery webhooks (notify on new agents)

---

## References

- [Phase 1 Oracle Review](./oracle-review.md)
- [Kagent HTTP Server](../../go/internal/httpserver/server.go)
- [Existing Handlers Pattern](../../go/internal/httpserver/handlers/)
- [Agent2Agent Specification](https://github.com/agent2agent/spec)

---

## Next Steps

1. ✅ Review this design document
2. ⏭️ Begin handler implementation (Week 3, Day 1)
3. ⏭️ Integrate with HTTP server (Week 3, Day 2)
4. ⏭️ Testing and refinement (Week 3, Day 3-4)
5. ⏭️ Documentation and review
6. ⏭️ Prepare for Phase 3 (Security & Authentication)

**Status**: Ready for implementation 🚀
