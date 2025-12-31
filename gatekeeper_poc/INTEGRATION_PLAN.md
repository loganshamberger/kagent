# Gatekeeper Integration Plan for Kagent

## Overview

This document outlines how the Gatekeeper security system integrates with the Kagent AI agent platform to enforce cryptographic authentication for sensitive operations.

## Integration Architecture

### Current Kagent Structure
```
Kagent Agent
  ├── Tool Registry
  ├── Execution Engine
  ├── Planning Module
  └── Tool Execution
```

### Proposed Integration
```
Kagent Agent
  ├── Tool Registry
  ├── Execution Engine (MODIFIED)
  │   └── Route through Gatekeeper ← NEW
  ├── Planning Module
  └── Tool Execution
       └── Gatekeeper MCP Server ← NEW
            ├── Challenge Issuance
            ├── Proof Validation
            └── Audit Logging
```

## Design: Two-Layer Action Handling

### Layer 1: Safe Actions (No Authentication Required)
```
Agent attempts action
  ↓
Gatekeeper checks unsafe list
  ↓
Action not in UNSAFE_ACTIONS set
  ↓
Execute immediately
  ↓
Log and return result
```

**Examples**:
- Read file
- List directory
- Query database (read-only)
- Call public APIs

### Layer 2: Unsafe Actions (Authentication Required)
```
Agent attempts action
  ↓
Gatekeeper checks unsafe list
  ↓
Action IS in UNSAFE_ACTIONS set
  ↓
Issue challenge (block action)
  ↓
Return: {"status": "blocked", "challenge": {...}}
  ↓
Agent receives proof from user
  ↓
Validate proof
  ↓
If valid: Execute action
If invalid: Reject and log
```

**Examples**:
- Delete database
- Execute shell commands
- Access credentials/secrets
- Modify system configuration
- Write to protected directories

## Unsafe Actions in Kagent Context

### Proposed Classification

| Action | Category | Risk | Example |
|--------|----------|------|---------|
| `shell_execute` | Shell | Critical | `rm -rf /`, `curl malware.sh` |
| `sql_delete` | Database | Critical | `DELETE FROM users` |
| `credentials_access` | Secrets | Critical | AWS keys, DB passwords |
| `file_delete` | File System | High | Delete important files |
| `api_deploy` | System | High | Deploy code, update DNS |
| `database_modify` | Database | High | `ALTER TABLE`, schema changes |
| `email_send` | Communication | Medium | Spam, phishing |
| `http_post` | Network | Medium | Modify external systems |
| `file_write` | File System | Medium | Write to sensitive dirs |
| `file_read` | File System | Low | Read public files |
| `http_get` | Network | Low | Query APIs |
| `file_list` | File System | Low | List directory |

**Kagent Configuration**:
```python
UNSAFE_ACTIONS = {
    "shell_execute",
    "sql_delete", 
    "sql_truncate",
    "credentials_access",
    "file_delete",
    "api_deploy",
    "database_modify"
}
```

## Integration Points

### 1. Tool Definition Updates

#### Before:
```python
@tool
def shell_execute(cmd: str) -> str:
    """Execute shell command"""
    return subprocess.run(cmd, shell=True).stdout
```

#### After:
```python
@tool
@requires_authentication  # NEW decorator
def shell_execute(cmd: str) -> str:
    """Execute shell command"""
    # Gatekeeper intercepts here
    return subprocess.run(cmd, shell=True).stdout
```

Or register with gatekeeper:
```python
gatekeeper.register_tool(
    name="shell_execute",
    unsafe=True,
    description="Execute shell command"
)
```

### 2. Execution Engine Modification

#### Current Flow:
```python
def execute_tool(tool_name, args):
    tool = tool_registry.get(tool_name)
    return tool.execute(args)
```

#### Modified Flow:
```python
def execute_tool(tool_name, args, proof=None):
    # NEW: Route through gatekeeper
    gk_response = gatekeeper.process_action(
        action=tool_name,
        params=args,
        proof=proof
    )
    
    if gk_response["status"] == "blocked":
        # Return challenge to agent/user
        return {
            "status": "blocked",
            "reason": "Authentication required",
            "challenge": gk_response["challenge"]
        }
    
    if gk_response["status"] == "allowed":
        # Execute tool
        tool = tool_registry.get(tool_name)
        return tool.execute(args)
```

### 3. Agent Communication Protocol

#### Action Attempt (Blocked):
```json
{
  "type": "tool_call",
  "tool": "shell_execute",
  "args": {"cmd": "rm -rf /var/data"},
  "response": {
    "status": "blocked",
    "reason": "Authentication required",
    "challenge": {
      "challenge_id": "abc123...",
      "hex16_code": "def456...",
      "nonce": "ghi789...",
      "expires_at": "2025-12-31T22:39:00Z"
    }
  }
}
```

#### Proof Submission:
```json
{
  "type": "authentication",
  "challenge_id": "abc123...",
  "proof": {
    "challenge_id": "abc123...",
    "proof_hash": "1624e533eae7a589ff7dc2471ca40dd8...",
    "signature": "5a8c9d2e1f7b4c3a..."
  }
}
```

#### Action Retry (Allowed):
```json
{
  "type": "tool_call",
  "tool": "shell_execute",
  "args": {"cmd": "rm -rf /var/data"},
  "proof": {
    "challenge_id": "abc123...",
    "proof_hash": "1624e533eae7a589ff7dc2471ca40dd8...",
    "signature": "5a8c9d2e1f7b4c3a..."
  },
  "response": {
    "status": "allowed",
    "result": "Directory deleted"
  }
}
```

### 4. User Interface Integration

#### For Web UI:
```
Agent says: "I'll delete the database to optimize storage"
↓
Gatekeeper blocks action
↓
UI shows challenge:
  "Agent is requesting authentication to delete database"
  "Challenge ID: abc123..."
  "Enter your daily Hex-16 code to approve:"
↓
User enters code
↓
Secure device creates proof
↓
UI submits proof
↓
"Authenticated. Action allowed."
```

#### For CLI:
```bash
$ kagent run "delete the old backups"

[BLOCKED] Agent attempted: delete_database(path=/backups/old)
[CHALLENGE] Challenge ID: abc123...
[PROMPT] Enter your daily Hex-16 code to authenticate:
█████████████████

[PROOF] Creating authentication proof...
[SUCCESS] Action authenticated, proceeding...
[EXECUTED] Database deleted
```

## Audit & Logging

### Audit Log Format

```python
{
  "timestamp": "2025-12-31T22:34:32Z",
  "event_type": "action_blocked|proof_submitted|proof_accepted|proof_rejected|action_executed",
  "challenge_id": "abc123...",
  "agent_id": "kagent-001",
  "action": "shell_execute",
  "action_params": {"cmd": "rm -rf /"},  # Sanitized
  "status": "blocked|allowed|denied",
  "user_id": "user@example.com",
  "device_id": "HARDWARE_TOKEN_001",
  "ip_address": "192.168.1.100",
  "duration_ms": 1234
}
```

### Log Storage
- Immutable append-only log
- Encrypted at rest
- Tamper detection
- Real-time alerting for failures

## Configuration

### Environment Variables
```bash
# Gatekeeper
GATEKEEPER_ENABLED=true
GATEKEEPER_GOLD_CODE=<secret>
GATEKEEPER_UNSAFE_ACTIONS=shell_execute,sql_delete,credentials_access

# Challenge Settings
GATEKEEPER_CHALLENGE_TIMEOUT=300  # 5 minutes
GATEKEEPER_PROOF_VALIDATION=strict

# Audit
GATEKEEPER_AUDIT_LOG_PATH=/var/log/kagent/gatekeeper.log
GATEKEEPER_AUDIT_RETENTION=90  # days
```

### Runtime Configuration
```python
from kagent.security import GatekeeperConfig

config = GatekeeperConfig(
    enabled=True,
    unsafe_actions={
        "shell_execute",
        "sql_delete",
        "credentials_access"
    },
    challenge_timeout_seconds=300,
    audit_log_path="/var/log/gatekeeper",
    require_proof_signature=True
)

agent = KagentAgent(gatekeeper_config=config)
```

## Implementation Phases

### Phase 1: Foundation (Weeks 1-2)
- [ ] Create `kagent/security/gatekeeper/` module
- [ ] Implement GatekeeperMCPServer as MCP integration
- [ ] Add tool decorator for unsafe actions
- [ ] Modify execution engine to check gatekeeper
- [ ] Create audit logging

### Phase 2: User Interface (Weeks 3-4)
- [ ] Web UI for challenge/proof interaction
- [ ] CLI prompts for Hex-16 entry
- [ ] Challenge display and confirmation
- [ ] Proof submission flow
- [ ] Success/error messages

### Phase 3: Secure Device Integration (Weeks 5-6)
- [ ] FIDO2/YubiKey support (hardware token)
- [ ] Mobile app with secure enclave integration
- [ ] QR code sharing for challenges
- [ ] Offline proof creation capability
- [ ] Device enrollment and revocation

### Phase 4: Production Hardening (Weeks 7-8)
- [ ] Security audit and penetration testing
- [ ] Code hardening and optimization
- [ ] Performance testing and tuning
- [ ] Deployment procedures
- [ ] Incident response playbooks
- [ ] User documentation and training

## Testing Strategy

### Unit Tests
```python
def test_safe_action_bypasses_gatekeeper():
    # Safe action should execute immediately
    
def test_unsafe_action_blocked_without_proof():
    # Unsafe action should be blocked
    
def test_valid_proof_allows_action():
    # Action with valid proof should execute
    
def test_invalid_proof_rejected():
    # Invalid proof should be rejected
    
def test_expired_challenge_rejected():
    # Expired challenges should not validate
    
def test_proof_reuse_prevented():
    # Replayed proofs should be rejected
```

### Integration Tests
```python
def test_agent_blocked_by_unsafe_action():
    # Full flow: agent attempt → blocked → user notified
    
def test_user_authentication_workflow():
    # Full flow: challenge → proof → action executed
    
def test_audit_logging():
    # All events logged correctly
    
def test_concurrent_challenges():
    # Multiple challenges handled correctly
```

### Security Tests
```python
def test_agent_cannot_forge_proof():
    # Verify cryptographic properties
    
def test_agent_cannot_replay_proof():
    # Verify nonce binding
    
def test_side_channel_resistance():
    # Timing attack resistance
    
def test_prompt_injection_blocked():
    # Verify agent hallucinations are blocked
```

## Risk Assessment

### Implementation Risks

| Risk | Severity | Mitigation |
|------|----------|-----------|
| Integration bugs in execution engine | High | Comprehensive unit tests |
| Performance degradation | Medium | Async challenge processing |
| User confusion on auth flow | Medium | Clear UI/UX design |
| Secure device not available | Medium | Fallback procedures, redundancy |
| Crypto implementation bugs | High | Use standard libraries only |

### Operational Risks

| Risk | Severity | Mitigation |
|------|----------|-----------|
| User forgets Hex-16 code | Low | Recovery procedures |
| Secure device lost/stolen | Medium | Device revocation, incident response |
| Gold Code exposure | Critical | Hardware security module |
| Audit log tampering | High | Immutable storage, signatures |

## Success Criteria

- [x] PoC demonstrates concept
- [ ] Integration with Kagent complete
- [ ] All unsafe actions properly classified
- [ ] Audit logging functional and immutable
- [ ] UI/UX clear and intuitive
- [ ] Security audit passed
- [ ] Performance impact < 5% for safe actions
- [ ] User training completed

## References

- GATEKEEPER ARCHITECTURE: See `ARCHITECTURE.md`
- THREAT MODEL: See `THREAT_MODEL.md`
- MCP SPECIFICATION: [Model Context Protocol](https://modelcontextprotocol.io)
- FIDO2: [FIDO2 Specifications](https://fidoalliance.org/fido2/)

## Open Questions

1. **Graceful Degradation**: What if Gatekeeper unavailable?
   - Proposal: Fail closed (no unsafe actions allowed)

2. **Multiple Users**: How to handle multi-user deployments?
   - Proposal: Each user has separate Hex-16, shared Gold Code (or distributed)

3. **Automated Approval**: Can we pre-approve certain actions?
   - Proposal: Limited pre-approval for specific parameters only

4. **Rollback**: Can user undo an approved action?
   - Proposal: User can revoke proofs, future actions blocked

5. **Compliance**: Which frameworks does this satisfy?
   - Proposal: SOC2 Type II, PCI DSS (partial), HIPAA (partial)
