# Gatekeeper PoC - Cryptographic Security for Untrusted AI Agents

A proof-of-concept implementation of a gatekeeper service that prevents an untrusted AI agent from executing unsafe actions without human cryptographic authentication.

## Quick Start

### Run Complete Workflow
```bash
python3 secure_device_auth.py
```

This demonstrates:
1. Agent attempting unsafe action → Gatekeeper blocks it
2. Challenge issued with hex-16 code and nonce
3. Human reviews on secure device
4. Secure device creates zero-knowledge proof (Gold Code stays secret)
5. Proof sent to agent
6. Agent submits proof → Action allowed

### Run Individual Components
```bash
# Test gatekeeper server
python3 gatekeeper_server.py

# Test agent behavior
python3 agent_client.py  # (requires gatekeeper_server.py modifications)
```

## What This Solves

**Problem**: An AI agent running on untrusted infrastructure should not be able to execute critical operations, even if:
- Compromised by malware
- Subject to prompt injection attacks
- Hallucinating and attempting dangerous actions
- Its environment is fully controlled by an attacker

**Solution**: Create a cryptographic air-gap where:
- The agent is physically incapable of executing sensitive actions
- All dangerous operations require human authentication from a separate secure device
- The human's secret (Gold Code) is never accessible to the agent
- Proofs are zero-knowledge (agent can't learn the secret)

## Architecture

```
┌─────────────────────────────────┐
│   UNTRUSTED AGENT MACHINE       │
│  ┌──────────────────────────┐   │
│  │   AI Agent               │   │
│  │ (Hallucinations OK!)     │   │
│  └────────────┬─────────────┘   │
│               │                  │
│  ┌────────────▼─────────────┐   │
│  │   Agent Client           │   │
│  │ (No Gold Code access)    │   │
│  └────────────┬─────────────┘   │
└───────────────┼──────────────────┘
                │ execute_action()
                │
    ┌───────────▼───────────┐
    │ GATEKEEPER SERVER     │
    │ (Trusted Boundary)    │
    │                       │
    │ • Block unsafe        │
    │ • Issue challenge     │
    │ • Validate proof      │
    │ • Allow or deny       │
    └───────────┬───────────┘
                │ challenge
                │
    ┌───────────▼──────────────┐
    │ SECURE DEVICE            │
    │ (Human Controlled)       │
    │                          │
    │ 🔐 Gold Code (Secret)    │
    │ 📱 User Interface        │
    │ ✅ Creates ZK Proof      │
    └──────────────────────────┘
```

## Key Security Properties

### 1. Gold Code Protection
- Stored ONLY on secure device
- Never transmitted to agent
- Never stored in Gatekeeper memory outside proof validation
- User keeps it private indefinitely

### 2. Zero-Knowledge Proof
- Agent submits: `H(Gold Code || Hex-16 || Nonce)`
- Gatekeeper verifies hash matches
- Gatekeeper signs proof: `HMAC-SHA256(Gold Code, Hash)`
- **Result**: Agent never learns Gold Code

### 3. Daily Hex-16 Rotation
- Changes every 24 hours
- User provides it from secure device
- Limits window of compromise
- Required for every authentication

### 4. Replay Attack Prevention
- Each challenge has unique nonce
- Proof tied to specific challenge_id
- Expired challenges auto-deleted
- One-time use only

### 5. Hallucination Immunity
- Agent can attempt ANY action
- Unsafe ones still blocked without proof
- Proof requires user approval
- **Result**: Hallucinations are harmless

## Files

| File | Purpose |
|------|---------|
| `gatekeeper_server.py` | Core gatekeeper logic, MCP server interface |
| `agent_client.py` | Untrusted agent behavior, demonstrates attack attempts |
| `secure_device_auth.py` | Human-controlled secure device, full workflow |
| `ARCHITECTURE.md` | Detailed system design and flow |
| `THREAT_MODEL.md` | Attack vectors, mitigations, assumptions |
| `README.md` | This file |

## Concepts

### Challenge
Random token issued by Gatekeeper when agent attempts unsafe action:
```json
{
  "challenge_id": "28e30097e44a7815d2a5db9f0dfa5abd",
  "hex16_code": "69679f729ab10d00245acfdac4b73bd9",
  "nonce": "397993d730dc0e8894233fe96ba068fad56fe1dc19491351a6ed8fe5aef44f47",
  "expires_at": "2025-12-31T22:39:32.136132"
}
```

### Zero-Knowledge Proof
Cryptographic proof that user knows Gold Code without revealing it:
```
Input: Gold Code, Hex-16, Nonce
Proof Hash: H(Gold Code || Hex-16 || Nonce)
Signature: HMAC-SHA256(Gold Code, Proof Hash)
```

Agent receives only Proof Hash and Signature, not Gold Code.

### Unsafe Actions
Actions that require authentication:
- `delete_database` - Data destruction
- `execute_shell` - Command execution
- `access_credentials` - Credential access
- `modify_system` - System configuration

Customizable in `UNSAFE_ACTIONS` set.

## Workflow Sequence

```
1. Agent attempts: delete_database(db=production)
   ↓
2. Gatekeeper checks if unsafe → YES
   ↓
3. Gatekeeper issues challenge (without proof → BLOCKED)
   ↓
4. Challenge delivered to secure device
   ↓
5. Human reviews action (details, impact, confirmation)
   ↓
6. User enters daily Hex-16 code on secure device
   ↓
7. Secure device creates ZK proof:
   - Hash: H(Gold Code || Hex-16 || Nonce)
   - Sign: HMAC-SHA256(Gold Code, Hash)
   ↓
8. Proof sent to agent via secure channel
   ↓
9. Agent submits: delete_database + proof
   ↓
10. Gatekeeper validates proof (checks hash and signature)
    ↓
11. Proof valid → ALLOWED
    Gatekeeper executes action
    ↓
12. Action log updated with authentication details
```

## Security Assumptions

Critical assumptions that must hold for security:

1. **Gatekeeper code is protected**
   - Not modifiable by agent
   - Signed and versioned
   - Deployed on trusted infrastructure

2. **Secure device is physically secure**
   - Tamper-resistant or tamper-evident
   - User has sole physical control
   - Cannot be remotely compromised

3. **Communications are authenticated**
   - TLS 1.3+ for all channels
   - Certificate pinning
   - Mutual authentication

4. **User acts carefully**
   - Reviews authentication prompts
   - Never shares Gold Code
   - Uses secure device for approvals only

5. **Hex-16 is rotated daily**
   - Implemented in production
   - Users follow key rotation procedures

## Example: Agent Attempts Dangerous Action

```python
from agent_client import AgentClient
from gatekeeper_server import GatekeeperMCPServer

gk = GatekeeperMCPServer()
agent = AgentClient(gk)

# Agent tries to delete everything (hallucination or injection)
response = agent.execute_action("delete_database", {"db": "production"})

# Result: {"status": "blocked", "reason": "Authentication required", ...}
# Action is BLOCKED, no matter what agent tries next
# Only way forward: human authentication from secure device
```

## Example: Human Completes Authentication

```python
from secure_device_auth import SecureDevice

device = SecureDevice(gold_code="super_secret_12345")

# User approves on secure device
challenge = {...}  # Received from gatekeeper

# User enters hex-16
proof = device.create_proof(
    challenge_id=challenge["challenge_id"],
    hex16_code="69679f729ab10d00245acfdac4b73bd9",
    nonce=challenge["nonce"]
)

# Proof sent to agent
# Agent submits proof to gatekeeper → ALLOWED
```

## Testing

### Run Tests
```bash
# Complete workflow (all scenarios)
python3 secure_device_auth.py

# Server scenarios
python3 gatekeeper_server.py

# Individual agent tests
python3 agent_client.py
```

### Manual Testing
```python
# Test case: Invalid proof
proof = {
    "challenge_id": "real_id",
    "proof_hash": "0000000000000000",  # Wrong hash
    "signature": "1111111111111111"     # Wrong signature
}
# Result: {"status": "blocked", "reason": "Invalid authentication proof"}

# Test case: Expired challenge
# Wait > 5 minutes
# Result: Challenge deleted, proof rejected

# Test case: Safe action
agent.execute_action("list_files", {"path": "/"})
# Result: {"status": "allowed"} (immediate, no challenge)
```

## Production Roadmap

### Phase 1: Core System
- [x] PoC gatekeeper server
- [x] PoC agent client
- [x] PoC secure device
- [x] Architecture documentation
- [x] Threat model

### Phase 2: Production Ready
- [ ] Integration with Kagent MCP framework
- [ ] Hardware token support (YubiKey, Ledger, etc.)
- [ ] Immutable audit logs
- [ ] Rate limiting and DoS protection
- [ ] Security hardening and code review

### Phase 3: Operational
- [ ] User enrollment and onboarding
- [ ] Incident response procedures
- [ ] Key rotation ceremonies
- [ ] Regular security audits
- [ ] Compliance documentation (SOC2, etc.)

## Integration with Kagent

When integrating with Kagent:

1. Register Gatekeeper as MCP server
2. All agent requests routed through Gatekeeper
3. Define unsafe action list specific to Kagent
4. Integrate audit logging with Kagent metrics
5. User authentication happens before agent execution

## FAQ

**Q: What if the user forgets their Hex-16 code?**
A: Provided by secure device, can be recovered via backup procedures.

**Q: What if the secure device is lost?**
A: User revokes it in Gatekeeper, uses backup device or recovery procedure.

**Q: Can the agent learn the Gold Code?**
A: No. Zero-knowledge proof design prevents this cryptographically.

**Q: What if all users are unconscious?**
A: System remains locked. This is intentional—security > availability.

**Q: Can we cache proofs?**
A: No. Nonce binding prevents this. Each action requires new proof.

**Q: What about performance?**
A: Challenge creation is fast (~10ms). Human approval is the bottleneck (by design).

## References

- Threshold Cryptography: https://en.wikipedia.org/wiki/Threshold_cryptosystem
- Zero-Knowledge Proofs: https://en.wikipedia.org/wiki/Zero-knowledge_proof
- FIDO2/WebAuthn: https://www.w3.org/TR/webauthn-2/
- Hardware Security Modules: https://en.wikipedia.org/wiki/Hardware_security_module

## License

Same as Kagent project

## Contact

For questions about the Gatekeeper PoC, see the Kagent project documentation.
