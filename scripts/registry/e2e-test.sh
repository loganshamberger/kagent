#!/bin/bash
set -e

echo "========================================="
echo "Agent Registry E2E Test Script"
echo "========================================="
echo ""

NAMESPACE="registry-e2e-manual"
REGISTRY_NAME="manual-test-registry"
AGENT_NAME="manual-test-agent"

cleanup() {
    echo ""
    echo "Cleaning up resources..."
    kubectl delete namespace ${NAMESPACE} --ignore-not-found=true --timeout=60s || true
    echo "Cleanup complete"
}

trap cleanup EXIT

echo "Step 1: Creating test namespace..."
kubectl create namespace ${NAMESPACE} || true
kubectl label namespace ${NAMESPACE} kagent.dev/agent-enabled=true --overwrite

echo ""
echo "Step 2: Deploying AgentRegistry..."
kubectl apply -f - <<EOF
apiVersion: kagent.dev/v1alpha1
kind: AgentRegistry
metadata:
  name: ${REGISTRY_NAME}
  namespace: ${NAMESPACE}
spec:
  discovery:
    enableAutoDiscovery: true
    syncInterval: "30s"
  observability:
    openTelemetryEnabled: true
  a2aVersion: "0.3.0"
EOF

echo ""
echo "Step 3: Waiting for AgentRegistry to be created..."
sleep 2

echo ""
echo "Step 4: Deploying test Agent with annotations..."
kubectl apply -f - <<EOF
apiVersion: kagent.dev/v1alpha2
kind: Agent
metadata:
  name: ${AGENT_NAME}
  namespace: ${NAMESPACE}
  annotations:
    kagent.dev/register-to-registry: "true"
    kagent.dev/capabilities: "kubernetes,monitoring,test"
    kagent.dev/card-team: "platform"
    kagent.dev/card-environment: "e2e-test"
spec:
  description: "Manual E2E test agent for Agent Registry"
  declarative:
    modelConfig: "test-model"
    a2aConfig:
      skills:
        - name: "kubernetes-management"
        - name: "cluster-monitoring"
EOF

echo ""
echo "Step 5: Waiting for AgentCard to be created (max 60 seconds)..."
for i in {1..30}; do
    if kubectl get agentcard ${AGENT_NAME} -n ${NAMESPACE} &>/dev/null; then
        echo "✓ AgentCard created successfully!"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "✗ Timeout waiting for AgentCard creation"
        exit 1
    fi
    echo "  Waiting... ($i/30)"
    sleep 2
done

echo ""
echo "Step 6: Verifying AgentCard contents..."
echo "----------------------------------------"
kubectl get agentcard ${AGENT_NAME} -n ${NAMESPACE} -o yaml

echo ""
echo "========================================="
echo "AgentCard Verification"
echo "========================================="

CARD_NAME=$(kubectl get agentcard ${AGENT_NAME} -n ${NAMESPACE} -o jsonpath='{.spec.name}')
CARD_VERSION=$(kubectl get agentcard ${AGENT_NAME} -n ${NAMESPACE} -o jsonpath='{.spec.a2aVersion}')
CARD_CAPABILITIES=$(kubectl get agentcard ${AGENT_NAME} -n ${NAMESPACE} -o jsonpath='{.spec.capabilities}')
CARD_HASH=$(kubectl get agentcard ${AGENT_NAME} -n ${NAMESPACE} -o jsonpath='{.status.hash}')
CARD_DESCRIPTION=$(kubectl get agentcard ${AGENT_NAME} -n ${NAMESPACE} -o jsonpath='{.spec.metadata.description}')

echo "Name: ${CARD_NAME}"
echo "A2A Version: ${CARD_VERSION}"
echo "Capabilities: ${CARD_CAPABILITIES}"
echo "Content Hash: ${CARD_HASH}"
echo "Description: ${CARD_DESCRIPTION}"

if [ -z "${CARD_NAME}" ]; then
    echo "✗ FAILED: AgentCard name is empty"
    exit 1
fi

if [ "${CARD_VERSION}" != "0.3.0" ]; then
    echo "✗ FAILED: Expected A2A version 0.3.0, got ${CARD_VERSION}"
    exit 1
fi

if [ -z "${CARD_HASH}" ]; then
    echo "✗ FAILED: Content hash is empty"
    exit 1
fi

echo ""
echo "========================================="
echo "AgentRegistry Status Verification"
echo "========================================="

REGISTRY_PHASE=$(kubectl get agentregistry ${REGISTRY_NAME} -n ${NAMESPACE} -o jsonpath='{.status.phase}')
REGISTRY_AGENTS=$(kubectl get agentregistry ${REGISTRY_NAME} -n ${NAMESPACE} -o jsonpath='{.status.registeredAgents}')
REGISTRY_LAST_SYNC=$(kubectl get agentregistry ${REGISTRY_NAME} -n ${NAMESPACE} -o jsonpath='{.status.lastSync}')

echo "Phase: ${REGISTRY_PHASE}"
echo "Registered Agents: ${REGISTRY_AGENTS}"
echo "Last Sync: ${REGISTRY_LAST_SYNC}"

if [ "${REGISTRY_PHASE}" != "Ready" ]; then
    echo "✗ FAILED: Expected phase Ready, got ${REGISTRY_PHASE}"
    exit 1
fi

if [ "${REGISTRY_AGENTS}" != "1" ]; then
    echo "✗ FAILED: Expected 1 registered agent, got ${REGISTRY_AGENTS}"
    exit 1
fi

echo ""
echo "========================================="
echo "Testing Content Hash Deduplication"
echo "========================================="

HASH_BEFORE=$(kubectl get agentcard ${AGENT_NAME} -n ${NAMESPACE} -o jsonpath='{.status.hash}')
echo "Hash before update: ${HASH_BEFORE}"

echo "Triggering controller reconciliation by annotating registry..."
kubectl annotate agentregistry ${REGISTRY_NAME} -n ${NAMESPACE} test-trigger="$(date +%s)" --overwrite

sleep 5

HASH_AFTER=$(kubectl get agentcard ${AGENT_NAME} -n ${NAMESPACE} -o jsonpath='{.status.hash}')
echo "Hash after update: ${HASH_AFTER}"

if [ "${HASH_BEFORE}" != "${HASH_AFTER}" ]; then
    echo "✗ FAILED: Content hash changed unexpectedly"
    exit 1
fi

echo "✓ Content hash deduplication working correctly"

echo ""
echo "========================================="
echo "Testing Discovery Disabled Annotation"
echo "========================================="

echo "Creating agent with discovery-disabled annotation..."
kubectl apply -f - <<EOF
apiVersion: kagent.dev/v1alpha2
kind: Agent
metadata:
  name: disabled-agent
  namespace: ${NAMESPACE}
  annotations:
    kagent.dev/register-to-registry: "true"
    kagent.dev/discovery-disabled: "true"
spec:
  description: "Agent that should not be discovered"
EOF

sleep 10

if kubectl get agentcard disabled-agent -n ${NAMESPACE} &>/dev/null; then
    echo "✗ FAILED: AgentCard should not be created for disabled agent"
    exit 1
fi

REGISTRY_AGENTS_AFTER=$(kubectl get agentregistry ${REGISTRY_NAME} -n ${NAMESPACE} -o jsonpath='{.status.registeredAgents}')
if [ "${REGISTRY_AGENTS_AFTER}" != "1" ]; then
    echo "✗ FAILED: Registered agents count changed unexpectedly: ${REGISTRY_AGENTS_AFTER}"
    exit 1
fi

echo "✓ Discovery disabled annotation working correctly"

echo ""
echo "========================================="
echo "✓ ALL TESTS PASSED"
echo "========================================="
echo ""
echo "Manual verification complete. Resources will be cleaned up automatically."
