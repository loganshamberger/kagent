# Gatekeeper Security Architecture

## Overview

The Gatekeeper system creates a cryptographic air-gap between an untrusted AI agent and sensitive operations. The agent is treated as an untrusted courier that must route all critical actions through a secure gate that requires human authentication.

## Core Components

### 1. Untrusted Agent Machine
- **AI Agent**: The autonomous system performing tasks
- **Constraints**: 
  - Does not have access to Gold Code
  - Cannot modify Gatekeeper logic
  - Cannot directly execute unsafe actions
  - Subject to hallucinations and prompt injections

### 2. Gatekeeper Server (Trusted Boundary)
- **Role**: Guard all unsafe operations
- **Responsibilities**:
  - Classify actions (safe vs. unsafe)
  - Issue authentication challenges for unsafe actions
  - Validate zero-knowledge proofs
  - Maintain audit log
- **Unsafe Actions**:
  - `delete_database` - Data destruction
  - `execute_shell` - Arbitrary command execution
  - `access_credentials` - Credential access
  - `modify_system` - System configuration changes

### 3. Secure Hardware Device (Human Controlled)
- **Role**: Human-only authentication device
- **Stores**:
  - Gold Code (master secret - NEVER shared with agent)
  - Generates daily Hex-16 codes
  - Creates zero-knowledge proofs
- **Properties**:
  - Physically separate from agent machine
  - User explicitly approves each action
  - Creates proofs without revealing secrets

## Authentication Flow

### Step 1: Unsafe Action Attempt
```
Agent → Gatekeeper: "execute delete_database"
Gatekeeper → Agent: "BLOCKED - Challenge Required"
```

### Step 2: Challenge Issuance
```
Gatekeeper issues:
  - challenge_id: random ID
  - hex16_code: today's human-generated hex
  - nonce: random challenge value
  - expires_at: challenge expiration time
```

### Step 3: Human Authentication
```
User on Secure Device:
  1. Reviews challenge (what action? which resources?)
  2. Enters daily Hex-16 code
  3. Device creates zero-knowledge proof:
     Proof = H(Gold Code || Hex-16 || Nonce)
     Signature = HMAC-SHA256(Gold Code, Proof)
```

### Step 4: Proof Submission
```
Agent receives proof (from secure device via secure channel)
Agent → Gatekeeper: "delete_database + proof"
Gatekeeper validates:
  - Challenge exists and not expired
  - Proof = H(Gold Code || Hex-16 || Nonce)
  - Signature matches HMAC-SHA256
Gatekeeper → Agent: "ALLOWED"
```

## Security Properties

### Cryptographic Guarantees

1. **Gold Code Never Exposed**
   - Only stored on secure device
   - Never transmitted to agent
   - Agent physically cannot access it

2. **Zero-Knowledge Proof**
   - Agent submits H(Gold Code || Hex-16 || Nonce)
   - Cannot reverse-engineer to extract Gold Code
   - Cannot forge proof without knowing Gold Code

3. **Daily Rotation**
   - Hex-16 code changes daily
   - Each authentication requires current code
   - Reduces window of compromise

4. **Proof Signature**
   - HMAC-SHA256(Gold Code, Proof)
   - Proves possession of Gold Code
   - Cannot be forged without accessing device

### Threat Mitigation

| Threat | Mitigation |
|--------|-----------|
| Agent hallucination | Blocked at Gatekeeper - no action without proof |
| Prompt injection | Blocked at Gatekeeper - no action without proof |
| Agent compromise | Agent can't access Gold Code or execute unsafe actions |
| Credential theft | Requires human approval + Hex-16 for each access |
| Network sniffing | Proof only (not secret); proof reuse blocked by nonce |
| Proof forgery | HMAC signature requires Gold Code knowledge |
| Insider threat | Requires physical access to secure device |

## Implementation Details

### Proof Validation Algorithm

```python
def validate_proof(proof: ZKProof, challenge: Challenge):
    # Reconstruct the input (only possible with Gold Code)
    expected_input = f"{GOLD_CODE}||{challenge.hex16_code}||{challenge.nonce}"
    
    # Hash it
    expected_hash = SHA256(expected_input)
    
    # Verify proof hash matches
    if proof.proof_hash != expected_hash:
        return False
    
    # Verify signature (proves possession of Gold Code)
    expected_sig = HMAC-SHA256(GOLD_CODE, expected_hash)
    if proof.signature != expected_sig:
        return False
    
    return True
```

### Key Insight

The agent submits a proof that the Gatekeeper can verify with the Gold Code, but:
- The agent never learns the Gold Code
- The agent cannot compute valid proofs without it
- The agent cannot forge new proofs
- The agent is cryptographically prevented from attacking the system

## Files

- `gatekeeper_server.py` - Core Gatekeeper logic and MCP server interface
- `agent_client.py` - Demonstrates untrusted agent attempting actions
- `secure_device_auth.py` - Simulates human-controlled secure device
- `ARCHITECTURE.md` - This document

## PoC Execution

```bash
# Run complete workflow simulation
python3 secure_device_auth.py

# Run individual components
python3 gatekeeper_server.py      # Server scenarios
python3 agent_client.py            # Agent scenarios (requires server)
```

## Production Considerations

### Secure Device Implementation
- Hardware token (FIDO2, YubiKey with custom applet)
- Mobile app with secure enclave (iOS Secure Enclave, Android Keystore)
- Purpose-built authentication device (e.g., hardware wallet adapted)

### Communication
- TLS 1.3+ for proof transmission
- Mutual authentication on secure channels
- Rate limiting on challenge issuance
- Replay attack prevention via nonces

### Audit & Compliance
- Immutable audit log of all authentications
- Timestamp server for proof timestamps
- Integration with compliance systems
- User activity tracking

### Operational Security
- Key rotation procedures
- Incident response for device compromise
- Secure backup/recovery procedures
- User education on Hex-16 security
