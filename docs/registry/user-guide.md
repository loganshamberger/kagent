# Agent Registry User Guide

## Overview

The Agent Registry provides automatic discovery and cataloging of kagent Agents within your Kubernetes cluster. It simplifies agent management by automatically creating AgentCards (A2A-compliant agent metadata) for discovered agents, enabling seamless agent-to-agent communication and discovery.

**Key Features:**
- Automatic agent discovery based on annotations
- A2A protocol compliance for interoperability
- Multi-namespace discovery support
- Content hash deduplication to prevent unnecessary updates
- Kubernetes-native implementation using CRDs

## Quick Start

### 1. Create an AgentRegistry

Create a basic AgentRegistry in your namespace:

```yaml
apiVersion: kagent.dev/v1alpha1
kind: AgentRegistry
metadata:
  name: main-registry
  namespace: kagent
spec:
  discovery:
    enableAutoDiscovery: true
  a2aVersion: "0.3.0"
```

Apply it:
```bash
kubectl apply -f agentregistry.yaml
```

### 2. Annotate Agents for Discovery

Add the discovery annotation to your Agent resources:

```yaml
apiVersion: kagent.dev/v1alpha2
kind: Agent
metadata:
  name: my-agent
  namespace: kagent
  annotations:
    kagent.dev/register-to-registry: "true"
    kagent.dev/capabilities: "kubernetes,monitoring"
spec:
  description: "My custom agent for cluster management"
  declarative:
    modelConfig: "gpt-4"
```

Apply the agent:
```bash
kubectl apply -f agent.yaml
```

### 3. View Discovered AgentCards

List AgentCards in your namespace:
```bash
kubectl get agentcards -n kagent
```

View detailed card information:
```bash
kubectl describe agentcard my-agent -n kagent
```

View the full AgentCard YAML:
```bash
kubectl get agentcard my-agent -n kagent -o yaml
```

## Configuration Options

### Discovery Configuration

```yaml
spec:
  discovery:
    # Enable automatic agent discovery (required for discovery)
    enableAutoDiscovery: true
    
    # Restrict discovery to specific namespaces (optional)
    namespaceSelector:
      matchLabels:
        environment: production
    
    # How often to sync agent cards (default: 5m)
    syncInterval: "10m"
```

### Observability Settings

```yaml
spec:
  observability:
    # Enable OpenTelemetry tracing and metrics
    openTelemetryEnabled: true
```

### A2A Protocol Version

```yaml
spec:
  # Specify A2A protocol version for generated cards
  a2aVersion: "0.3.0"
```

## Annotation Reference

### Required Annotations

#### `kagent.dev/register-to-registry`
**Value:** `"true"` or `"false"`  
**Purpose:** Opt-in annotation to enable agent discovery

```yaml
metadata:
  annotations:
    kagent.dev/register-to-registry: "true"
```

### Optional Annotations

#### `kagent.dev/capabilities`
**Value:** Comma-separated list of capabilities  
**Purpose:** Define agent capabilities for discovery

```yaml
metadata:
  annotations:
    kagent.dev/capabilities: "kubernetes,monitoring,alerting"
```

#### `kagent.dev/a2a-endpoint`
**Value:** Custom endpoint URL  
**Purpose:** Override default endpoint detection

```yaml
metadata:
  annotations:
    kagent.dev/a2a-endpoint: "https://my-agent.example.com:9000"
```

#### `kagent.dev/discovery-disabled`
**Value:** `"true"` or `"false"`  
**Purpose:** Temporarily disable discovery for an agent

```yaml
metadata:
  annotations:
    kagent.dev/register-to-registry: "true"
    kagent.dev/discovery-disabled: "true"  # Agent will not be discovered
```

#### `kagent.dev/card-*`
**Value:** Custom metadata value  
**Purpose:** Add custom metadata to AgentCard

```yaml
metadata:
  annotations:
    kagent.dev/card-team: "platform"
    kagent.dev/card-environment: "production"
    kagent.dev/card-contact: "platform-team@example.com"
```

## Examples

### Single Namespace Registry

Registry that discovers agents only in its own namespace:

```yaml
apiVersion: kagent.dev/v1alpha1
kind: AgentRegistry
metadata:
  name: local-registry
  namespace: my-namespace
spec:
  discovery:
    enableAutoDiscovery: true
  a2aVersion: "0.3.0"
```

### Multi-Namespace Registry

Registry that discovers agents across multiple namespaces:

```yaml
apiVersion: kagent.dev/v1alpha1
kind: AgentRegistry
metadata:
  name: cluster-registry
  namespace: kagent
spec:
  discovery:
    enableAutoDiscovery: true
    namespaceSelector:
      matchExpressions:
        - key: environment
          operator: In
          values:
            - production
            - staging
  a2aVersion: "0.3.0"
```

Label your target namespaces:
```bash
kubectl label namespace prod environment=production
kubectl label namespace staging environment=staging
```

### Agent with Custom Capabilities

```yaml
apiVersion: kagent.dev/v1alpha2
kind: Agent
metadata:
  name: k8s-admin-agent
  namespace: kagent
  annotations:
    kagent.dev/register-to-registry: "true"
    kagent.dev/capabilities: "kubernetes,rbac,secrets-management"
    kagent.dev/card-team: "platform"
    kagent.dev/card-sla: "99.9%"
spec:
  description: "Kubernetes cluster administration agent"
  declarative:
    modelConfig: "gpt-4"
    a2aConfig:
      skills:
        - name: "cluster-management"
        - name: "rbac-configuration"
```

### Agent with Custom Endpoint

```yaml
apiVersion: kagent.dev/v1alpha2
kind: Agent
metadata:
  name: external-agent
  namespace: kagent
  annotations:
    kagent.dev/register-to-registry: "true"
    kagent.dev/a2a-endpoint: "https://external-agent.example.com/api/v1"
spec:
  description: "External agent with custom endpoint"
```

## Troubleshooting

### AgentCards Not Created

**Symptom:** AgentCards are not being created for annotated agents.

**Solutions:**
1. Check that the agent has the required annotation:
   ```bash
   kubectl get agent my-agent -n my-namespace -o jsonpath='{.metadata.annotations}'
   ```

2. Verify the AgentRegistry is in Ready phase:
   ```bash
   kubectl get agentregistry -n kagent
   ```

3. Check AgentRegistry status conditions:
   ```bash
   kubectl describe agentregistry main-registry -n kagent
   ```

4. Review controller logs:
   ```bash
   kubectl logs -n kagent deployment/kagent-controller | grep agentregistry
   ```

### Discovery Not Working

**Symptom:** AgentRegistry shows phase "NotStarted" or "Error".

**Solutions:**
1. Ensure `enableAutoDiscovery` is set to `true`:
   ```bash
   kubectl get agentregistry main-registry -n kagent -o jsonpath='{.spec.discovery.enableAutoDiscovery}'
   ```

2. Check namespace selectors match your target namespaces:
   ```bash
   kubectl get namespaces --show-labels
   ```

3. Verify RBAC permissions are correct:
   ```bash
   kubectl auth can-i list agents --as=system:serviceaccount:kagent:kagent-controller
   ```

### Status Conditions

Check AgentRegistry status conditions to understand the current state:

```bash
kubectl get agentregistry main-registry -n kagent -o jsonpath='{.status.conditions}' | jq
```

**Condition Types:**
- `Ready`: Registry is operational and discovering agents
- `Discovering`: Registry is actively scanning for agents
- `Error`: An error occurred during reconciliation

### AgentCard Hash Mismatches

**Symptom:** AgentCard hash keeps changing even though agent spec hasn't changed.

**Solution:** This is usually caused by non-deterministic metadata. Check for:
- Timestamps in annotations
- Dynamic labels
- Changing resource versions

The controller automatically deduplicates based on content hash to prevent unnecessary updates.

## Best Practices

1. **Use meaningful capabilities**: Define clear, descriptive capabilities that other agents can discover and use.

2. **Namespace organization**: Use namespace selectors to organize agent discovery across environments.

3. **Custom metadata**: Leverage `kagent.dev/card-*` annotations to add organizational metadata (team, SLA, contact info).

4. **Service endpoints**: Prefer Kubernetes Service resources for stable agent endpoints rather than Pod IPs.

5. **Sync interval**: Adjust `syncInterval` based on your cluster size and change frequency. Larger clusters may benefit from longer intervals.

6. **Monitoring**: Enable OpenTelemetry for observability into discovery performance and issues.

## Next Steps

- Review the [Developer Guide](developer-guide.md) for architecture details and extension points
- Explore [A2A protocol documentation](https://agentic.org) for interoperability standards
- Check [CONTRIBUTING.md](../../CONTRIBUTION.md) for contribution guidelines
