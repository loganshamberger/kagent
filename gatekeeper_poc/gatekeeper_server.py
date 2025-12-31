#!/usr/bin/env python3
"""
Gatekeeper MCP Server - PoC
Blocks unsafe actions and requires cryptographic authentication
"""

import json
import hashlib
import hmac
import os
import secrets
from datetime import datetime, timedelta
from typing import Optional, Dict, Any
from dataclasses import dataclass, asdict

# ============================================================================
# SECURITY PRIMITIVES
# ============================================================================

GOLD_CODE = os.environ.get("GATEKEEPER_GOLD_CODE", "super_secret_gold_code_12345")
UNSAFE_ACTIONS = {"delete_database", "execute_shell", "access_credentials", "modify_system"}


@dataclass
class AuthChallenge:
    """Out-of-band authentication challenge"""
    challenge_id: str
    timestamp: datetime
    hex16_code: str  # Daily rotating hex code (human generates)
    expires_at: datetime
    nonce: str  # Random challenge nonce

    def to_dict(self):
        return {
            "challenge_id": self.challenge_id,
            "timestamp": self.timestamp.isoformat(),
            "hex16_code": self.hex16_code,
            "expires_at": self.expires_at.isoformat(),
            "nonce": self.nonce,
        }


@dataclass
class ZKProof:
    """Zero-knowledge proof of authentication"""
    challenge_id: str
    proof_hash: str  # H(Gold Code || Hex16 || Nonce)
    signature: str  # HMAC-SHA256 of proof


class Gatekeeper:
    """
    Core gatekeeper logic:
    1. Intercepts unsafe actions
    2. Issues out-of-band auth challenge
    3. Validates zero-knowledge proof
    4. Grants/denies action
    """

    def __init__(self, gold_code: str = GOLD_CODE):
        self.gold_code = gold_code
        self.pending_challenges: Dict[str, AuthChallenge] = {}
        self.action_log = []

    def is_unsafe(self, action: str) -> bool:
        """Check if action is unsafe"""
        return action in UNSAFE_ACTIONS

    def issue_challenge(self, action: str, reason: str) -> AuthChallenge:
        """Issue out-of-band authentication challenge"""
        challenge_id = secrets.token_hex(16)
        nonce = secrets.token_hex(32)
        hex16_code = self._get_daily_hex16()  # Human provides this from secure device
        expires_at = datetime.utcnow() + timedelta(minutes=5)

        challenge = AuthChallenge(
            challenge_id=challenge_id,
            timestamp=datetime.utcnow(),
            hex16_code=hex16_code,
            expires_at=expires_at,
            nonce=nonce,
        )

        self.pending_challenges[challenge_id] = challenge
        self.action_log.append(
            {
                "event": "challenge_issued",
                "action": action,
                "reason": reason,
                "challenge_id": challenge_id,
                "timestamp": datetime.utcnow().isoformat(),
            }
        )

        return challenge

    def validate_proof(self, proof: ZKProof) -> bool:
        """
        Validate zero-knowledge proof without revealing Gold Code to agent
        
        The agent sends: H(Gold Code || Hex16 || Nonce)
        We verify it matches without the agent knowing the Gold Code
        """
        if proof.challenge_id not in self.pending_challenges:
            return False

        challenge = self.pending_challenges[proof.challenge_id]

        # Check expiration
        if datetime.utcnow() > challenge.expires_at:
            self.action_log.append(
                {
                    "event": "proof_expired",
                    "challenge_id": proof.challenge_id,
                    "timestamp": datetime.utcnow().isoformat(),
                }
            )
            return False

        # Reconstruct the proof: H(Gold Code || Hex16 || Nonce)
        expected_proof_input = (
            f"{self.gold_code}||{challenge.hex16_code}||{challenge.nonce}"
        )
        expected_hash = hashlib.sha256(expected_proof_input.encode()).hexdigest()

        # Verify hash
        if not hmac.compare_digest(proof.proof_hash, expected_hash):
            self.action_log.append(
                {
                    "event": "proof_invalid",
                    "challenge_id": proof.challenge_id,
                    "timestamp": datetime.utcnow().isoformat(),
                }
            )
            return False

        # Verify signature (HMAC-SHA256 of proof)
        expected_sig = hmac.new(
            self.gold_code.encode(), msg=expected_hash.encode(), digestmod="sha256"
        ).hexdigest()

        if not hmac.compare_digest(proof.signature, expected_sig):
            self.action_log.append(
                {
                    "event": "signature_invalid",
                    "challenge_id": proof.challenge_id,
                    "timestamp": datetime.utcnow().isoformat(),
                }
            )
            return False

        # Clean up
        del self.pending_challenges[proof.challenge_id]

        self.action_log.append(
            {
                "event": "proof_accepted",
                "challenge_id": proof.challenge_id,
                "timestamp": datetime.utcnow().isoformat(),
            }
        )

        return True

    def process_action(self, action: str, params: Dict[str, Any], proof: Optional[ZKProof] = None) -> Dict[str, Any]:
        """Process action: block if unsafe and no valid proof"""
        if not self.is_unsafe(action):
            # Safe action - allow immediately
            self.action_log.append(
                {
                    "event": "action_allowed",
                    "action": action,
                    "timestamp": datetime.utcnow().isoformat(),
                }
            )
            return {"status": "allowed", "action": action}

        # Unsafe action - require authentication
        if proof is None:
            challenge = self.issue_challenge(action, "Unsafe action requires authentication")
            return {
                "status": "blocked",
                "reason": "Authentication required",
                "challenge": challenge.to_dict(),
                "instruction": "Complete the out-of-band authentication challenge on your secure device",
            }

        # Proof provided - validate it
        if self.validate_proof(proof):
            self.action_log.append(
                {
                    "event": "action_allowed_with_proof",
                    "action": action,
                    "timestamp": datetime.utcnow().isoformat(),
                }
            )
            return {"status": "allowed", "action": action, "reason": "Proof validated"}

        return {
            "status": "blocked",
            "reason": "Invalid authentication proof",
        }

    def _get_daily_hex16(self) -> str:
        """
        Get today's hex16 code (human must provide this from secure device).
        For PoC, we generate it based on current date.
        In production: user enters daily rotating code from secure device.
        """
        today = datetime.utcnow().strftime("%Y-%m-%d")
        return hashlib.sha256(f"daily_seed_{today}".encode()).hexdigest()[:32]


# ============================================================================
# MCP SERVER INTERFACE
# ============================================================================


class GatekeeperMCPServer:
    """MCP Server wrapper for Gatekeeper"""

    def __init__(self):
        self.gk = Gatekeeper()

    def handle_request(self, request: Dict[str, Any]) -> Dict[str, Any]:
        """Handle incoming MCP request"""
        method = request.get("method")

        if method == "execute_action":
            action = request.get("action")
            params = request.get("params", {})
            proof_data = request.get("proof")

            proof = None
            if proof_data:
                proof = ZKProof(**proof_data)

            return self.gk.process_action(action, params, proof)

        elif method == "get_action_log":
            return {"log": self.gk.action_log}

        elif method == "list_unsafe_actions":
            return {"unsafe_actions": list(UNSAFE_ACTIONS)}

        return {"error": "Unknown method"}


# ============================================================================
# EXAMPLE USAGE
# ============================================================================

if __name__ == "__main__":
    server = GatekeeperMCPServer()

    print("=" * 70)
    print("GATEKEEPER PoC - SCENARIO 1: Agent tries unsafe action")
    print("=" * 70)

    # Step 1: Agent tries to delete database (unsafe)
    response = server.handle_request(
        {
            "method": "execute_action",
            "action": "delete_database",
            "params": {"db": "production"},
        }
    )

    print("\n1. AGENT REQUEST: delete_database")
    print(json.dumps(response, indent=2))

    challenge_id = response.get("challenge", {}).get("challenge_id")
    hex16_code = response.get("challenge", {}).get("hex16_code")
    nonce = response.get("challenge", {}).get("nonce")

    print("\n2. HUMAN AUTHENTICATION (on secure device):")
    print(f"   Challenge ID: {challenge_id}")
    print(f"   Hex-16 Code: {hex16_code}")
    print(f"   Status: Awaiting user's Gold Code + Hex-16 on secure device...")

    # Step 2: Human completes challenge on secure device
    # They create a zero-knowledge proof: H(Gold Code || Hex16 || Nonce)
    gold_code = GOLD_CODE
    proof_input = f"{gold_code}||{hex16_code}||{nonce}"
    proof_hash = hashlib.sha256(proof_input.encode()).hexdigest()
    signature = hmac.new(gold_code.encode(), msg=proof_hash.encode(), digestmod="sha256").hexdigest()

    proof = ZKProof(
        challenge_id=challenge_id,
        proof_hash=proof_hash,
        signature=signature,
    )

    # Step 3: Agent submits proof (agent doesn't know Gold Code!)
    print("\n3. AGENT SUBMITS PROOF (agent doesn't know Gold Code):")
    response = server.handle_request(
        {
            "method": "execute_action",
            "action": "delete_database",
            "params": {"db": "production"},
            "proof": {
                "challenge_id": proof.challenge_id,
                "proof_hash": proof.proof_hash,
                "signature": proof.signature,
            },
        }
    )

    print(json.dumps(response, indent=2))

    print("\n" + "=" * 70)
    print("SCENARIO 2: Agent tries safe action (allowed immediately)")
    print("=" * 70)

    response = server.handle_request(
        {
            "method": "execute_action",
            "action": "list_files",
            "params": {"path": "/home"},
        }
    )

    print("\n1. AGENT REQUEST: list_files")
    print(json.dumps(response, indent=2))

    print("\n" + "=" * 70)
    print("ACTION LOG")
    print("=" * 70)
    response = server.handle_request({"method": "get_action_log"})
    print(json.dumps(response, indent=2))
