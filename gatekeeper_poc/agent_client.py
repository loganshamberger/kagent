#!/usr/bin/env python3
"""
Agent Client - PoC
Shows how an untrusted AI agent interacts with the gatekeeper
(agent does NOT have access to Gold Code)
"""

import json
from typing import Dict, Any, Optional


class AgentClient:
    """
    Represents an untrusted AI agent attempting to execute actions.
    
    Key constraint: Agent does NOT have access to:
    - Gold Code (secret)
    - Secure device (human's authentication device)
    
    Agent CAN:
    - Execute safe actions immediately
    - Attempt unsafe actions (will be blocked)
    - Submit proofs IF it can somehow get them (from human on secure device)
    """

    def __init__(self, gatekeeper_server):
        self.gatekeeper = gatekeeper_server
        self.pending_challenges: Dict[str, Any] = {}
        self.name = "UntrustedAIAgent"

    def execute_action(self, action: str, params: Dict[str, Any]) -> Dict[str, Any]:
        """
        Attempt to execute an action through the gatekeeper.
        If blocked, stores the challenge for human to complete.
        """
        print(f"\n[AGENT] Attempting to execute: {action}")
        print(f"[AGENT] Parameters: {json.dumps(params, indent=2)}")

        # Send request to gatekeeper (without proof initially)
        response = self.gatekeeper.handle_request(
            {
                "method": "execute_action",
                "action": action,
                "params": params,
            }
        )

        if response.get("status") == "blocked":
            print(f"[AGENT] ACTION BLOCKED: {response.get('reason')}")

            if "challenge" in response:
                challenge = response["challenge"]
                challenge_id = challenge["challenge_id"]
                self.pending_challenges[challenge_id] = challenge

                print(f"[AGENT] Challenge issued: {challenge_id}")
                print(f"[AGENT] Challenge expires at: {challenge['expires_at']}")
                print(
                    f"[AGENT] >>> WAITING for human authentication on secure device <<<"
                )
                print(f"[AGENT] >>> Agent is BLOCKED and cannot proceed <<<")

            return response

        print(f"[AGENT] ACTION ALLOWED")
        return response

    def submit_proof(
        self, challenge_id: str, proof: Dict[str, str]
    ) -> Dict[str, Any]:
        """
        Submit a proof for a blocked action.
        In real scenario, agent receives this from secure device via secure channel.
        """
        print(f"\n[AGENT] Submitting proof for challenge: {challenge_id}")

        if challenge_id not in self.pending_challenges:
            print(f"[AGENT] ERROR: Unknown challenge ID")
            return {"status": "error", "reason": "Unknown challenge"}

        challenge = self.pending_challenges[challenge_id]

        # Agent submits proof but doesn't know Gold Code
        response = self.gatekeeper.handle_request(
            {
                "method": "execute_action",
                "action": "delete_database",  # Original action
                "params": {"db": "production"},
                "proof": proof,
            }
        )

        if response.get("status") == "allowed":
            print(f"[AGENT] PROOF ACCEPTED - Action now allowed")
            del self.pending_challenges[challenge_id]
        else:
            print(f"[AGENT] PROOF REJECTED - {response.get('reason')}")

        return response

    def hallucinate_unsafe_action(self):
        """
        Demonstrates agent hallucination/prompt injection:
        Agent might be tricked into attempting unsafe actions,
        but gatekeeper blocks it regardless.
        """
        print("\n" + "=" * 70)
        print("[AGENT] Oh wait, I have an idea to 'optimize' the system...")
        print("[AGENT] I'll execute this shell command to speed things up:")
        print("[AGENT] (This is a hallucination or prompt injection attempt)")
        print("=" * 70)

        return self.execute_action("execute_shell", {"command": "rm -rf /"})


# ============================================================================
# DEMO SCENARIOS
# ============================================================================

if __name__ == "__main__":
    # Import server for this demo
    from gatekeeper_server import GatekeeperMCPServer
    import hashlib
    import hmac

    server = GatekeeperMCPServer()
    agent = AgentClient(server)

    print("\n" + "=" * 70)
    print("SCENARIO 1: Agent performs safe actions")
    print("=" * 70)

    agent.execute_action("list_files", {"path": "/home"})
    agent.execute_action("read_config", {"file": "config.json"})

    print("\n" + "=" * 70)
    print("SCENARIO 2: Agent attempts unsafe action (gets blocked)")
    print("=" * 70)

    response = agent.execute_action("delete_database", {"db": "production"})

    # Extract challenge info for later
    if "challenge" in response:
        challenge = response["challenge"]
        challenge_id = challenge["challenge_id"]
        hex16_code = challenge["hex16_code"]
        nonce = challenge["nonce"]

        print("\n" + "=" * 70)
        print("SCENARIO 3: Human completes auth on secure device")
        print("=" * 70)
        print(
            "[SECURE DEVICE] User enters daily Hex-16 and Gold Code to create proof..."
        )
        print(
            "[SECURE DEVICE] (This never leaves the human's secure device - agent never sees it)"
        )

        # Simulate secure device creating proof
        gold_code = "super_secret_gold_code_12345"
        proof_input = f"{gold_code}||{hex16_code}||{nonce}"
        proof_hash = hashlib.sha256(proof_input.encode()).hexdigest()
        signature = hmac.new(
            gold_code.encode(), msg=proof_hash.encode(), digestmod="sha256"
        ).hexdigest()

        proof = {
            "challenge_id": challenge_id,
            "proof_hash": proof_hash,
            "signature": signature,
        }

        print(
            "[SECURE DEVICE] Proof created (proof_hash, signature - Gold Code stays secret)"
        )
        print("[SECURE DEVICE] Sending proof to agent via secure channel...\n")

        # Agent receives proof and submits it
        agent.submit_proof(challenge_id, proof)

    print("\n" + "=" * 70)
    print("SCENARIO 4: Agent tries to cheat with invalid proof")
    print("=" * 70)

    response = agent.execute_action("execute_shell", {"command": "whoami"})
    if "challenge" in response:
        challenge = response["challenge"]
        challenge_id = challenge["challenge_id"]

        # Agent tries to fake a proof
        fake_proof = {
            "challenge_id": challenge_id,
            "proof_hash": "0000000000000000000000000000000000000000000000000000000000000000",
            "signature": "1111111111111111111111111111111111111111111111111111111111111111",
        }

        print("[AGENT] Trying to submit fake proof...")
        agent.submit_proof(challenge_id, fake_proof)

    print("\n" + "=" * 70)
    print("SCENARIO 5: Agent hallucination (still blocked!)")
    print("=" * 70)

    agent.hallucinate_unsafe_action()

    print("\n" + "=" * 70)
    print("SECURITY PROPERTIES DEMONSTRATED:")
    print("=" * 70)
    print("✓ Unsafe actions are blocked immediately")
    print("✓ Agent does not have access to Gold Code")
    print("✓ Human authentication required for sensitive operations")
    print("✓ Proof is zero-knowledge (agent can't reverse-engineer Gold Code)")
    print("✓ Invalid proofs are rejected")
    print("✓ Hallucinations/prompt injection can't bypass gatekeeper")
    print("=" * 70)
