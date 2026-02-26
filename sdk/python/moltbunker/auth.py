"""Authentication strategies for Moltbunker SDK."""

from __future__ import annotations

import time
from abc import ABC, abstractmethod
from typing import Optional

import httpx


class AuthStrategy(ABC):
    """Base authentication strategy."""

    @abstractmethod
    def apply(self, request: httpx.Request) -> httpx.Request:
        """Apply authentication to an outgoing request."""


class APIKeyAuth(AuthStrategy):
    """Authenticate using an API key (mb_live_* or mb_*)."""

    def __init__(self, api_key: str):
        if not api_key.startswith("mb_"):
            raise ValueError("API key must start with 'mb_'")
        self._api_key = api_key

    def apply(self, request: httpx.Request) -> httpx.Request:
        request.headers["Authorization"] = f"Bearer {self._api_key}"
        return request


class WalletSessionAuth(AuthStrategy):
    """Authenticate using a wallet challenge-verify session (wt_* tokens).

    Requires eth-account: pip install moltbunker[wallet]
    """

    def __init__(self, private_key: str):
        self._private_key = private_key
        self._token: Optional[str] = None
        self._expires_at: float = 0

    @property
    def is_authenticated(self) -> bool:
        return self._token is not None and time.time() < self._expires_at

    async def authenticate(self, client: httpx.AsyncClient, base_url: str) -> None:
        """Perform challenge-verify flow to obtain a session token."""
        try:
            from eth_account import Account
            from eth_account.messages import encode_defunct
        except ImportError:
            raise ImportError("Install eth-account: pip install moltbunker[wallet]")

        account = Account.from_key(self._private_key)

        # Step 1: Request challenge
        resp = await client.post(f"{base_url}/v1/auth/challenge", json={"address": account.address})
        resp.raise_for_status()
        challenge = resp.json()

        # Step 2: Sign challenge
        message = encode_defunct(text=challenge["message"])
        signed = account.sign_message(message)

        # Step 3: Verify signature
        resp = await client.post(
            f"{base_url}/v1/auth/verify",
            json={
                "address": account.address,
                "message": challenge["message"],
                "signature": signed.signature.hex(),
            },
        )
        resp.raise_for_status()
        result = resp.json()

        self._token = result["access_token"]
        self._expires_at = time.time() + result.get("expires_in", 3600) - 60  # 60s buffer

    def authenticate_sync(self, client: httpx.Client, base_url: str) -> None:
        """Synchronous version of authenticate."""
        try:
            from eth_account import Account
            from eth_account.messages import encode_defunct
        except ImportError:
            raise ImportError("Install eth-account: pip install moltbunker[wallet]")

        account = Account.from_key(self._private_key)

        resp = client.post(f"{base_url}/v1/auth/challenge", json={"address": account.address})
        resp.raise_for_status()
        challenge = resp.json()

        message = encode_defunct(text=challenge["message"])
        signed = account.sign_message(message)

        resp = client.post(
            f"{base_url}/v1/auth/verify",
            json={
                "address": account.address,
                "message": challenge["message"],
                "signature": signed.signature.hex(),
            },
        )
        resp.raise_for_status()
        result = resp.json()

        self._token = result["access_token"]
        self._expires_at = time.time() + result.get("expires_in", 3600) - 60

    def apply(self, request: httpx.Request) -> httpx.Request:
        if self._token:
            request.headers["Authorization"] = f"Bearer {self._token}"
        return request


class InlineWalletAuth(AuthStrategy):
    """Authenticate by signing every request with a wallet key.

    Requires eth-account: pip install moltbunker[wallet]
    """

    def __init__(self, private_key: str):
        try:
            from eth_account import Account
        except ImportError:
            raise ImportError("Install eth-account: pip install moltbunker[wallet]")

        self._account = Account.from_key(private_key)

    def apply(self, request: httpx.Request) -> httpx.Request:
        from eth_account.messages import encode_defunct

        timestamp = str(int(time.time()))
        message = f"moltbunker-auth:{timestamp}"
        signed = self._account.sign_message(encode_defunct(text=message))

        request.headers["X-Wallet-Address"] = self._account.address
        request.headers["X-Wallet-Signature"] = signed.signature.hex()
        request.headers["X-Wallet-Message"] = message
        return request
