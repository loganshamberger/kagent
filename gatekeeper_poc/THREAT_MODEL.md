# Gatekeeper Security Threat Model

## Assumptions

### Trust Boundaries
1. **Untrusted**: Agent machine (malware, compromise, hallucination)
2. **Trusted**: Gatekeeper Server (isolated, audited, tamper-evident)
3. **Trusted**: Secure Device (physical security, user-controlled)
4. **Trusted**: Secure Channel (TLS, authenticated, integrity-protected)

### Security Premises
- Gatekeeper code cannot be modified by agent
- Secure device is physically separate and user-controlled
- User controls Hex-16 code (not agent)
- Gold Code is never exposed to agent
- Network channels use authenticated encryption

## Attack Vectors & Mitigations

### Category 1: Direct Action Execution

#### Attack 1.1: Prompt Injection to Execute Unsafe Action
**Threat**: Attacker injects prompt to make agent execute `delete_database`

**Status**: ✓ MITIGATED
- Action blocked at Gatekeeper
- No proof provided = action blocked
- Proof requires human approval
- **Result**: Agent cannot execute regardless of prompt

#### Attack 1.2: Hallucination Leads to Unsafe Action
**Threat**: Agent hallucinates and attempts dangerous operation

**Status**: ✓ MITIGATED
- Same blocking mechanism as prompt injection
- Hallucination doesn't change crypto properties
- **Result**: Hallucination is harmless

#### Attack 1.3: Agent Directly Calls Unsafe Library
**Threat**: Agent executes `rm -rf /` via direct system call

**Status**: ✓ MITIGATED (if Gatekeeper is enforced)
- Gatekeeper intercepts all action requests
- Direct calls bypass Gatekeeper = architectural failure
- **Mitigation**: Use sandboxing, seccomp, eBPF to block direct calls
- **Assumption**: Agent operates through Gatekeeper interface only

### Category 2: Cryptographic Attacks

#### Attack 2.1: Forge Proof Without Gold Code
**Threat**: Agent attempts to create valid proof H(Gold Code || Hex-16 || Nonce)

**Status**: ✓ MITIGATED
- Cannot invert SHA256 hash
- Cannot guess Gold Code (256-bit entropy)
- Without Gold Code, cannot compute HMAC signature
- **Crypto**: Computationally infeasible

#### Attack 2.2: Replay Attack - Reuse Previous Proof
**Threat**: Agent saves proof from one action, replays for different action

**Status**: ✓ MITIGATED
- Each challenge generates unique nonce
- Each proof tied to challenge_id
- Gatekeeper deletes challenge after proof validation
- Proof includes hash(Gold || Hex || Nonce) - nonce differs per challenge
- **Result**: Replay impossible without new authentication

#### Attack 2.3: Proof Modification
**Threat**: Agent intercepts proof, modifies bytes

**Status**: ✓ MITIGATED
- Proof integrity protected by signature
- HMAC is cryptographically authenticated
- Single bit flip invalidates signature
- **Result**: Any modification detected

#### Attack 2.4: Brute Force Proof Generation
**Threat**: Agent generates 2^256 possible proofs to match SHA256 hash

**Status**: ✓ MITIGATED (probabilistically)
- Would require 2^256 attempts (infeasible)
- Gatekeeper rate-limits challenge attempts
- Failed proofs logged and monitored
- **Result**: Rate limiting makes this impractical

### Category 3: Secret Extraction

#### Attack 3.1: Side-Channel Attack on HMAC
**Threat**: Timing analysis, power analysis to extract Gold Code

**Status**: ⚠ PARTIAL MITIGATION
- Python uses constant-time comparison (`hmac.compare_digest`)
- Production: use hardware-backed secure enclave for Gold Code
- **Mitigation**: Gatekeeper runs on tamper-resistant hardware
- **Assumption**: Secure device is physically secure

#### Attack 3.2: Memory Dump of Gatekeeper
**Threat**: Attacker memory-dumps Gatekeeper process, extracts Gold Code

**Status**: ✓ MITIGATED
- Gold Code stored ONLY on secure device
- Gatekeeper never stores Gold Code in memory
- Gatekeeper only validates proofs (knows Gold Code only during validation)
- **Result**: Memory dump yields no secrets (only proofs, nonces, hashes)

#### Attack 3.3: Social Engineering User for Gold Code
**Threat**: Attacker tricks user into revealing Gold Code

**Status**: ✓ MITIGATED by design
- Architecture never requires user to share Gold Code
- Only Hex-16 is shared (changing daily)
- User keeps Gold Code private always
- **Result**: No attack vector here

#### Attack 3.4: Extract Gold Code from Secure Device
**Threat**: Physical attack on secure device (tampering)

**Status**: ✓ MITIGATED
- Assumption: Secure device has tamper detection
- If tampered: device self-destruct or refuse operation
- Hardware token standard: tamper-evident cases, alarm on intrusion
- **Result**: Attack becomes physically expensive, detectable

### Category 4: Protocol Attacks

#### Attack 4.1: Man-in-the-Middle on Challenge Transmission
**Threat**: MITM intercepts challenge, modifies it

**Status**: ✓ MITIGATED
- Challenge sent over authenticated TLS
- Challenge integrity protected
- Modification detected by HMAC validation
- Wrong nonce = invalid proof
- **Result**: Ineffective

#### Attack 4.2: Challenge Expiration Bypass
**Threat**: Agent submits proof for expired challenge

**Status**: ✓ MITIGATED
- Gatekeeper checks timestamp before validation
- Expired challenges deleted from system
- Proof for expired challenge fails validation
- **Result**: Expires prevent old proofs

#### Attack 4.3: Nonce Exhaustion
**Threat**: Agent exhausts nonce space, finds collision

**Status**: ✓ MITIGATED
- Nonce is 256-bit random (2^256 entropy)
- Each challenge gets new nonce
- Collision probability: negligible
- **Result**: Computationally infeasible

#### Attack 4.4: Denial of Service - Challenge Spam
**Threat**: Agent floods with action requests, exhausts resources

**Status**: ✓ PARTIAL MITIGATION
- Rate limiting on challenge issuance
- Challenges auto-expire (5 minutes default)
- Gatekeeper logs spam attempts
- **Mitigation**: Implement request throttling
- **Result**: System remains functional under attack

### Category 5: Implementation Bugs

#### Attack 5.1: Timing Side-Channel in Proof Validation
**Threat**: Time taken to validate proof reveals information about Gold Code

**Status**: ✓ MITIGATED
- Python `hmac.compare_digest()` is constant-time
- All comparisons complete regardless of input
- No early-exit on mismatch
- **Result**: No information leakage

#### Attack 5.2: Off-by-One in Crypto Algorithm
**Threat**: Implementation bug in SHA256 or HMAC

**Status**: ⚠ DEPENDS ON LIBRARY
- Using Python standard library (battle-tested)
- No custom crypto implementation
- Community audited (hashlib, hmac)
- **Mitigation**: Regular updates, security patches

#### Attack 5.3: Proof Validation Logic Bypass
**Threat**: Bug in `validate_proof()` that allows invalid proofs

**Status**: ✓ MITIGATED by code review
- Logic is simple and auditable
- Clear separation of concerns
- Unit tests cover happy/unhappy paths
- **Mitigation**: Formal verification in production

### Category 6: Operational Attacks

#### Attack 6.1: User Forgets Hex-16 Code
**Threat**: User unable to approve actions

**Status**: ⚠ UX ISSUE (not security issue)
- Code can be recovered from secure device
- Fallback mechanisms (backup codes, recovery procedures)
- **Mitigation**: User training, device documentation

#### Attack 6.2: Hex-16 Code Leaked
**Threat**: Daily code disclosed, attacker forges proofs

**Status**: ✓ MITIGATED by design
- Hex-16 alone insufficient to forge proof
- Also requires Gold Code (secret)
- Also requires correct nonce (changes per challenge)
- Attacker needs: Gold Code + Hex-16 + Nonce
- **Result**: Leaked Hex-16 ≠ security breach

#### Attack 6.3: Secure Device Lost/Stolen
**Threat**: Attacker physically possesses secure device

**Status**: ✓ MITIGATED
- Device has PIN/biometric unlock
- Failed auth attempts lock device
- Device may have anti-tampering (alarm, self-wipe)
- User can revoke device from Gatekeeper
- **Mitigation**: Incident response procedure for device loss

#### Attack 6.4: Insider Threat - Admin Modifies Gatekeeper
**Threat**: System admin changes Gatekeeper to allow all actions

**Status**: ✓ MITIGATED by architecture
- Gatekeeper code is audited and version-controlled
- Changes require code review
- Audit log detects unauthorized changes
- **Mitigation**: Immutable infrastructure, signed deployments

## Residual Risks

### Low Risk
- ✓ Prompt injection/hallucination - cryptographically prevented
- ✓ Proof forgery - 256-bit computational hardness
- ✓ Replay attacks - nonce binding + expiration
- ✓ Credential theft - requires human approval per action

### Medium Risk
- ⚠ Side-channel attacks - mitigated by constant-time ops, requires hardware hardening
- ⚠ DoS on challenge system - rate limiting required
- ⚠ Secure device compromise - requires physical attack

### High Risk (Architectural)
- ⚠ Agent bypasses Gatekeeper entirely - requires sandboxing enforcement
- ⚠ Gatekeeper compromise (code modification) - requires signed deployments
- ⚠ User coercion (forced to approve dangerous action) - no technical mitigation (social engineering)

## Security Assumptions Checklist

- [ ] Gatekeeper code is tamper-evident (signed, versioned)
- [ ] Gatekeeper runs on isolated/hardened system
- [ ] Secure device is physically secure
- [ ] TLS used for all communications (1.3+)
- [ ] Audit logs are immutable
- [ ] Rate limiting on challenge issuance
- [ ] Hex-16 code regenerates daily
- [ ] Users receive security training
- [ ] Incident response procedures documented
- [ ] Regular security audits scheduled

## Testing Strategy

### Unit Tests
- ✓ Proof validation (valid/invalid cases)
- ✓ Challenge expiration
- ✓ Nonce uniqueness
- ✓ Signature verification

### Integration Tests
- ✓ Agent → Gatekeeper → Secure Device flow
- ✓ Proof submission and validation
- ✓ Proof rejection scenarios

### Security Tests
- ✓ Brute force resistance
- ✓ Timing attacks (constant-time ops)
- ✓ Invalid proof rejection
- ✓ Replay attack prevention
- ✓ Rate limiting under load

### Threat Modeling Tests
- ✓ Can agent execute action without proof? NO
- ✓ Can agent forge valid proof? NO (computationally infeasible)
- ✓ Can agent replay old proof? NO (nonce binding)
- ✓ Can agent learn Gold Code? NO (zero-knowledge)
