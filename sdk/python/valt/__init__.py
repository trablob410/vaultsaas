"""Valt Python SDK - minimal client for the Valt secret vault API."""

import os
import json
import urllib.request
import urllib.error
from typing import Optional, List, Dict, Any


class ValtError(Exception):
    """Raised when the Valt API returns an error."""

    def __init__(self, status: int, message: str) -> None:
        self.status = status
        super().__init__(f"Valt API error {status}: {message}")


class ValtClient:
    """Client for the Valt secret vault API."""

    def __init__(
        self,
        base_url: str = "http://localhost:8080/api/v1",
        token: Optional[str] = None,
    ) -> None:
        self.base_url = base_url.rstrip("/")
        self.token = token or os.environ.get("VALT_TOKEN", "")

    def _request(self, method: str, path: str, body: Optional[Dict] = None) -> Any:
        """Make an authenticated HTTP request."""
        url = self.base_url + path
        data = json.dumps(body).encode() if body is not None else None
        req = urllib.request.Request(url, data=data, method=method)
        req.add_header("Content-Type", "application/json")
        if self.token:
            req.add_header("Authorization", f"Bearer {self.token}")
        try:
            with urllib.request.urlopen(req, timeout=30) as resp:
                raw = resp.read()
                return json.loads(raw) if raw else {}
        except urllib.error.HTTPError as exc:
            raw = exc.read().decode(errors="replace")
            raise ValtError(exc.code, raw) from exc

    def list_secrets(self) -> List[Dict[str, Any]]:
        """List all secrets."""
        return self._request("GET", "/secrets").get("secrets", [])

    def request_access(
        self, secret_id: str, reason: str, duration_minutes: int = 30
    ) -> str:
        """Request access to a secret. Returns the request ID."""
        result = self._request(
            "POST",
            f"/secrets/{secret_id}/access-requests",
            {"reason": reason, "duration_minutes": duration_minutes},
        )
        return result.get("id", "")

    def get_credential(self, request_id: str) -> Dict[str, Any]:
        """Get an approved credential by request ID."""
        return self._request("GET", f"/credentials/{request_id}")

    def list_agents(self, project_id: str) -> List[Dict[str, Any]]:
        """List agents for a project."""
        return self._request("GET", f"/projects/{project_id}/agents").get("agents", [])

    def create_agent(self, project_id: str, name: str) -> Dict[str, Any]:
        """Create a new agent in a project."""
        return self._request("POST", f"/projects/{project_id}/agents", {"name": name})
