"""Synchronous Moltbunker client."""

from __future__ import annotations

import os
from typing import Any, Dict, List, Optional

import httpx

from moltbunker.auth import APIKeyAuth, AuthStrategy
from moltbunker.exceptions import raise_for_status
from moltbunker.models import (
    AgentDeployment,
    AgentInvokeRequest,
    AgentInvokeResponse,
    AgentSpec,
    BalanceResponse,
    CloneRequest,
    CloneResponse,
    ContainerInfo,
    CrawlJob,
    CrawlJobRequest,
    CrawlPageRequest,
    CrawlResult,
    CrawlStats,
    DeployRequest,
    DeployResponse,
    MemoryEntry,
    MoltDeployRequest,
    MoltDeployResponse,
    MoltInfo,
    MoltInvokeRequest,
    MoltInvokeResponse,
    ProxySession,
    ProxyUsage,
    SnapshotRequest,
    SnapshotResponse,
    StatusResponse,
    StorageBucket,
    StorageObject,
    StorageUsage,
    SubdomainInfo,
    SubdomainRegisterRequest,
    ThreatResponse,
)

DEFAULT_BASE_URL = "https://api.moltbunker.com"
DEFAULT_TIMEOUT = 30.0


class MoltbunkerClient:
    """Synchronous client for the Moltbunker API.

    Usage::

        client = MoltbunkerClient(api_key="mb_live_...")
        status = client.get_status()
        print(status.peer_count)
        client.close()

    Or as a context manager::

        with MoltbunkerClient(api_key="mb_live_...") as client:
            status = client.get_status()
    """

    def __init__(
        self,
        *,
        base_url: Optional[str] = None,
        api_key: Optional[str] = None,
        auth: AuthStrategy | None = None,
        timeout: float = DEFAULT_TIMEOUT,
        max_retries: int = 3,
    ):
        self._base_url = (base_url or os.environ.get("MOLTBUNKER_API_URL") or DEFAULT_BASE_URL).rstrip("/")

        if auth is not None:
            self._auth = auth
        elif api_key or os.environ.get("MOLTBUNKER_API_KEY"):
            self._auth = APIKeyAuth(api_key or os.environ["MOLTBUNKER_API_KEY"])
        else:
            self._auth = None

        transport = httpx.HTTPTransport(retries=max_retries)
        self._client = httpx.Client(timeout=timeout, transport=transport)

    def close(self) -> None:
        self._client.close()

    def __enter__(self):
        return self

    def __exit__(self, *args):
        self.close()

    # --- HTTP helpers ---

    def _request(self, method: str, path: str, **kwargs) -> Dict[str, Any]:
        url = f"{self._base_url}{path}"
        request = self._client.build_request(method, url, **kwargs)
        if self._auth:
            request = self._auth.apply(request)

        response = self._client.send(request)
        if response.status_code >= 400:
            try:
                body = response.json()
            except Exception:
                body = None
            raise_for_status(response.status_code, body)

        if response.status_code == 204:
            return {}
        return response.json()

    def _get(self, path: str, **kwargs) -> Dict[str, Any]:
        return self._request("GET", path, **kwargs)

    def _post(self, path: str, **kwargs) -> Dict[str, Any]:
        return self._request("POST", path, **kwargs)

    def _put(self, path: str, **kwargs) -> Dict[str, Any]:
        return self._request("PUT", path, **kwargs)

    def _delete(self, path: str, **kwargs) -> Dict[str, Any]:
        return self._request("DELETE", path, **kwargs)

    # --- Status ---

    def get_status(self) -> StatusResponse:
        return StatusResponse.model_validate(self._get("/v1/status"))

    # --- Balance ---

    def get_balance(self) -> BalanceResponse:
        return BalanceResponse.model_validate(self._get("/v1/balance"))

    # --- Threat ---

    def get_threat(self) -> ThreatResponse:
        return ThreatResponse.model_validate(self._get("/v1/threat"))

    # --- Deploy ---

    def deploy(self, request: DeployRequest) -> DeployResponse:
        data = self._post("/v1/deploy", json=request.model_dump(exclude_none=True))
        return DeployResponse.model_validate(data)

    # --- Containers ---

    def list_containers(self) -> List[ContainerInfo]:
        data = self._get("/v1/containers")
        return [ContainerInfo.model_validate(c) for c in data.get("containers", [])]

    def get_container(self, container_id: str) -> ContainerInfo:
        return ContainerInfo.model_validate(self._get(f"/v1/containers/{container_id}"))

    def stop_container(self, container_id: str) -> None:
        self._post(f"/v1/containers/{container_id}", json={"action": "stop"})

    def delete_container(self, container_id: str) -> None:
        self._delete(f"/v1/containers/{container_id}")

    # --- Snapshots ---

    def create_snapshot(self, request: SnapshotRequest) -> SnapshotResponse:
        data = self._post("/v1/snapshot", json=request.model_dump())
        return SnapshotResponse.model_validate(data)

    def list_snapshots(self) -> List[SnapshotResponse]:
        data = self._get("/v1/snapshots")
        return [SnapshotResponse.model_validate(s) for s in data.get("snapshots", [])]

    # --- Clones ---

    def clone(self, container_id: str, request: CloneRequest) -> CloneResponse:
        body = request.model_dump()
        body["container_id"] = container_id
        return CloneResponse.model_validate(self._post("/v1/clone", json=body))

    # --- Molts (Serverless) ---

    def deploy_molt(self, request: MoltDeployRequest) -> MoltDeployResponse:
        data = self._post("/v1/molts", json=request.model_dump(exclude_none=True))
        return MoltDeployResponse.model_validate(data)

    def list_molts(self) -> List[MoltInfo]:
        data = self._get("/v1/molts")
        return [MoltInfo.model_validate(m) for m in data.get("molts", [])]

    def get_molt(self, molt_id: str) -> MoltInfo:
        return MoltInfo.model_validate(self._get(f"/v1/molts/{molt_id}"))

    def invoke_molt(self, molt_id: str, request: MoltInvokeRequest) -> MoltInvokeResponse:
        data = self._post(f"/v1/molts/{molt_id}/invoke", json=request.model_dump())
        return MoltInvokeResponse.model_validate(data)

    def delete_molt(self, molt_id: str) -> None:
        self._delete(f"/v1/molts/{molt_id}")

    # --- Subdomains ---

    def list_subdomains(self) -> List[SubdomainInfo]:
        data = self._get("/v1/subdomains")
        return [SubdomainInfo.model_validate(s) for s in data.get("subdomains", [])]

    def register_subdomain(self, request: SubdomainRegisterRequest) -> SubdomainInfo:
        return SubdomainInfo.model_validate(self._post("/v1/subdomains", json=request.model_dump()))

    def delete_subdomain(self, name: str) -> None:
        self._delete(f"/v1/subdomains/{name}")

    # --- Storage ---

    def create_bucket(self, name: str) -> StorageBucket:
        return StorageBucket.model_validate(self._post("/v1/storage/buckets", json={"name": name}))

    def list_buckets(self) -> List[StorageBucket]:
        data = self._get("/v1/storage/buckets")
        return [StorageBucket.model_validate(b) for b in data.get("buckets", data if isinstance(data, list) else [])]

    def delete_bucket(self, name: str) -> None:
        self._delete(f"/v1/storage/buckets/{name}")

    def put_object(self, bucket: str, key: str, data: bytes, content_type: str = "") -> StorageObject:
        import base64

        body = {"bucket": bucket, "key": key, "data": base64.b64encode(data).decode(), "content_type": content_type}
        return StorageObject.model_validate(self._put(f"/v1/storage/objects/{bucket}/{key}", json=body))

    def get_object(self, bucket: str, key: str) -> bytes:
        import base64

        data = self._get(f"/v1/storage/objects/{bucket}/{key}")
        return base64.b64decode(data.get("data", ""))

    def delete_object(self, bucket: str, key: str) -> None:
        self._delete(f"/v1/storage/objects/{bucket}/{key}")

    def list_objects(self, bucket: str, prefix: str = "") -> List[StorageObject]:
        params = {"prefix": prefix} if prefix else {}
        data = self._get(f"/v1/storage/objects/{bucket}", params=params)
        return [StorageObject.model_validate(o) for o in data.get("objects", [])]

    def storage_usage(self) -> StorageUsage:
        return StorageUsage.model_validate(self._get("/v1/storage/usage"))

    # --- Proxy ---

    def list_proxy_sessions(self) -> List[ProxySession]:
        data = self._get("/v1/proxy/sessions")
        return [ProxySession.model_validate(s) for s in data.get("sessions", [])]

    def proxy_usage(self) -> ProxyUsage:
        return ProxyUsage.model_validate(self._get("/v1/proxy/usage"))

    # --- Crawling ---

    def create_crawl_job(self, request: CrawlJobRequest) -> CrawlJob:
        data = self._post("/v1/crawl/jobs", json=request.model_dump())
        return CrawlJob.model_validate(data)

    def list_crawl_jobs(self) -> List[CrawlJob]:
        data = self._get("/v1/crawl/jobs")
        return [CrawlJob.model_validate(j) for j in data.get("jobs", [])]

    def get_crawl_job(self, job_id: str) -> CrawlJob:
        return CrawlJob.model_validate(self._get(f"/v1/crawl/jobs/{job_id}"))

    def get_crawl_results(self, job_id: str) -> List[CrawlResult]:
        data = self._get(f"/v1/crawl/jobs/{job_id}/results")
        return [CrawlResult.model_validate(r) for r in data.get("results", [])]

    def cancel_crawl_job(self, job_id: str) -> None:
        self._post(f"/v1/crawl/jobs/{job_id}/cancel")

    def crawl_page(self, request: CrawlPageRequest) -> CrawlJob:
        data = self._post("/v1/crawl/pages", json=request.model_dump())
        return CrawlJob.model_validate(data)

    def get_crawl_stats(self) -> CrawlStats:
        return CrawlStats.model_validate(self._get("/v1/crawl/stats"))

    # --- Agents ---

    def deploy_agent(self, spec: AgentSpec) -> AgentDeployment:
        data = self._post("/v1/agents", json=spec.model_dump())
        return AgentDeployment.model_validate(data)

    def list_agents(self) -> List[AgentDeployment]:
        data = self._get("/v1/agents")
        return [AgentDeployment.model_validate(a) for a in data.get("agents", [])]

    def get_agent(self, agent_id: str) -> AgentDeployment:
        return AgentDeployment.model_validate(self._get(f"/v1/agents/{agent_id}"))

    def delete_agent(self, agent_id: str) -> None:
        self._delete(f"/v1/agents/{agent_id}")

    def invoke_agent(self, agent_id: str, request: AgentInvokeRequest) -> AgentInvokeResponse:
        data = self._post(f"/v1/agents/{agent_id}/invoke", json=request.model_dump())
        return AgentInvokeResponse.model_validate(data)

    def stop_agent(self, agent_id: str) -> None:
        self._post(f"/v1/agents/{agent_id}/stop")

    def list_agent_memory(self, agent_id: str) -> List[MemoryEntry]:
        data = self._get(f"/v1/agents/{agent_id}/memory")
        return [MemoryEntry.model_validate(e) for e in data.get("entries", [])]

    def set_agent_memory(self, agent_id: str, entry: MemoryEntry) -> None:
        self._post(f"/v1/agents/{agent_id}/memory", json=entry.model_dump())

    def delete_agent_memory(self, agent_id: str, key: str) -> None:
        self._delete(f"/v1/agents/{agent_id}/memory", params={"key": key})
