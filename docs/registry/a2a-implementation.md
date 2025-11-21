# A2A Implementation for Agent Registry

## Overview

The Agent Registry implementation includes A2A (Agent-to-Agent) protocol support for exposing agent metadata in a standardized format. This enables external systems and other agents to discover and interact with kagent agents.

## Architecture

### Components

1. **A2A Translator** (`go/pkg/registry/a2a/translator.go`)
   - Converts kagent `AgentCard` CRD → `server.AgentCard` (A2A v0.2.5 format)
   - Handles metadata mapping and default value population
   - Generates JSON serialization for external consumption

2. **AgentCard CRD** (`go/api/v1alpha1/agentcard_types.go`)
   - `Spec.PublicCard`: Stores A2A-compliant JSON document
   - `Spec.A2AVersion`: Tracks protocol version (default: "0.3.0")
   - `Spec.Metadata`: Extensible key-value pairs for A2A fields

3. **Card Generator** (`go/internal/controller/registry/cardgen.go`)
   - Enhanced to automatically generate A2A JSON during card creation
   - Integrates translator for PublicCard field population

4. **REST API** (`go/pkg/registry/api/server.go`)
   - `/api/v1alpha1/registry/cards`: List all agent cards
   - `/api/v1alpha1/registry/cards/{namespace}/{name}`: Get specific card
   - `/api/v1alpha1/registry/cards/{namespace}/{name}/a2a`: Get A2A JSON

## A2A Protocol Version

kagent uses **A2A v0.2.5** as defined by `trpc.group/trpc-go/trpc-a2a-go`.

### AgentCard Schema

```json
{
  "name": "string",
  "description": "string",
  "version": "string",
  "url": "string",
  "provider": {
    "organization": "string",
    "url": "string"
  },
  "iconUrl": "string",
  "documentationUrl": "string",
  "capabilities": {
    "streaming": true,
    "pushNotifications": false,
    "stateTransitionHistory": false
  },
  "defaultInputModes": ["text"],
  "defaultOutputModes": ["text"],
  "skills": [
    {
      "id": "string",
      "name": "string",
      "description": "string",
      "tags": []
    }
  ],
  "protocolVersion": "0.3.0"
}
```

## Metadata Mapping

The translator maps `AgentCard.Spec.Metadata` fields to A2A card properties:

| Metadata Key | A2A Field | Required |
|--------------|-----------|----------|
| `description` | `.description` | No (defaults to "Agent {name}") |
| `provider.organization` | `.provider.organization` | No |
| `provider.url` | `.provider.url` | No |
| `iconUrl` | `.iconUrl` | No |
| `documentationUrl` | `.documentationUrl` | No |

## Usage

### Annotating an Agent for A2A Discovery

```yaml
apiVersion: kagent.dev/v1alpha2
kind: Agent
metadata:
  name: my-agent
  namespace: default
  annotations:
    kagent.dev/register-to-registry: "true"
    kagent.dev/capabilities: "kubernetes,monitoring"
spec:
  description: "My custom agent"
```

### Setting A2A Metadata

```yaml
apiVersion: kagent.dev/v1alpha2
kind: Agent
metadata:
  name: my-agent
  annotations:
    kagent.dev/register-to-registry: "true"
    kagent.dev/card-metadata/provider.organization: "My Company"
    kagent.dev/card-metadata/provider.url: "https://company.com"
    kagent.dev/card-metadata/iconUrl: "https://company.com/icon.png"
    kagent.dev/card-metadata/documentationUrl: "https://docs.company.com"
```

### Retrieving A2A Cards via REST API

```bash
# List all agent cards
curl http://kagent-controller.kagent.svc.cluster.local:8084/api/v1alpha1/registry/cards

# Get specific card (full CRD)
curl http://kagent-controller.kagent.svc.cluster.local:8084/api/v1alpha1/registry/cards/default/my-agent

# Get A2A-compliant JSON
curl http://kagent-controller.kagent.svc.cluster.local:8084/api/v1alpha1/registry/cards/default/my-agent/a2a
```

### Example A2A Card Output

```json
{
  "name": "my-agent",
  "description": "My custom agent",
  "version": "1.0.0",
  "url": "http://my-agent.default.svc.cluster.local:8080",
  "provider": {
    "organization": "My Company",
    "url": "https://company.com"
  },
  "iconUrl": "https://company.com/icon.png",
  "documentationUrl": "https://docs.company.com",
  "capabilities": {
    "streaming": true,
    "pushNotifications": false,
    "stateTransitionHistory": false
  },
  "defaultInputModes": ["text"],
  "defaultOutputModes": ["text"],
  "skills": [
    {
      "id": "skill-0",
      "name": "kubernetes",
      "tags": []
    },
    {
      "id": "skill-1",
      "name": "monitoring",
      "tags": []
    }
  ],
  "protocolVersion": "0.3.0"
}
```

## Implementation Details

### Translator Behavior

1. **Default Values**:
   - Description: "Agent {name}" if not provided
   - Version: "1.0.0" if not specified
   - Skills: Single "General Purpose" skill if no capabilities defined
   - Capabilities: Streaming enabled, push notifications disabled

2. **Endpoint Selection**:
   - Primary URL from first endpoint in `AgentCard.Spec.Endpoints`
   - Falls back to empty string if no endpoints

3. **Skills Generation**:
   - Each capability becomes a skill
   - Skill ID: `skill-{index}`
   - Skill name: capability string

### Content Hashing

The card generator calculates a SHA256 hash of the entire `AgentCardSpec` including the `PublicCard` field. This enables no-op update detection in the controller.

## Testing

Run A2A translation tests:
```bash
cd go
go test ./pkg/registry/a2a/... -v
```

Run REST API tests:
```bash
cd go
go test ./pkg/registry/api/... -v
```

## Security

### Phase 1 (Current)
- REST API exposed via ClusterIP Service (cluster-internal only)
- Authentication: Kubernetes RBAC via API server
- Authorization: Read-only operations
- No custom authentication layer

### Future Enhancements
- Service mesh mTLS (Istio/Linkerd)
- External API Gateway with OAuth2/OIDC
- Rate limiting and quota
- Audit logging

## References

- A2A Protocol: https://github.com/google/A2A
- trpc-a2a-go: https://github.com/trpc-group/trpc-a2a-go
- AgentCard CRD: [agentcard_types.go](file:///Users/loganshamberger/Development/kagent/go/api/v1alpha1/agentcard_types.go)
- Translator: [translator.go](file:///Users/loganshamberger/Development/kagent/go/pkg/registry/a2a/translator.go)
