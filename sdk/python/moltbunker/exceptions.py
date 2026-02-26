"""Exception hierarchy for Moltbunker SDK."""

from __future__ import annotations

from typing import Optional


class MoltbunkerError(Exception):
    """Base exception for all Moltbunker SDK errors."""

    def __init__(self, message: str, status_code: Optional[int] = None, response_body: Optional[dict] = None):
        super().__init__(message)
        self.status_code = status_code
        self.response_body = response_body


class AuthenticationError(MoltbunkerError):
    """Authentication or authorization failed (401/403)."""


class NotFoundError(MoltbunkerError):
    """Requested resource not found (404)."""


class RateLimitError(MoltbunkerError):
    """Rate limit exceeded (429). Check retry_after attribute."""

    def __init__(self, message: str, retry_after: float = 60.0, **kwargs):
        super().__init__(message, **kwargs)
        self.retry_after = retry_after


class ValidationError(MoltbunkerError):
    """Request validation failed (400/422)."""


class InsufficientBalanceError(MoltbunkerError):
    """Insufficient BUNKER token balance for the operation."""


class ServerError(MoltbunkerError):
    """Server-side error (5xx)."""


class TimeoutError(MoltbunkerError):
    """Request timed out."""


def raise_for_status(status_code: int, body: Optional[dict] = None):
    """Raise the appropriate exception for an HTTP error status."""
    msg = ""
    if body and "error" in body:
        msg = body["error"]
    elif body and "message" in body:
        msg = body["message"]

    if status_code == 401:
        raise AuthenticationError(msg or "Authentication required", status_code=status_code, response_body=body)
    if status_code == 403:
        raise AuthenticationError(msg or "Permission denied", status_code=status_code, response_body=body)
    if status_code == 404:
        raise NotFoundError(msg or "Not found", status_code=status_code, response_body=body)
    if status_code == 422 or status_code == 400:
        raise ValidationError(msg or "Validation error", status_code=status_code, response_body=body)
    if status_code == 429:
        raise RateLimitError(msg or "Rate limit exceeded", status_code=status_code, response_body=body)
    if status_code >= 500:
        raise ServerError(msg or f"Server error ({status_code})", status_code=status_code, response_body=body)
    if status_code >= 400:
        raise MoltbunkerError(msg or f"HTTP {status_code}", status_code=status_code, response_body=body)
