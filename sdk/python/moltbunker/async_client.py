"""Asynchronous Moltbunker client."""

from __future__ import annotations

import os
from typing import Any, Dict, List, Optional

import httpx

from moltbunker.auth import APIKeyAuth, AuthStrategy, WalletSessionAuth
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


class AsyncMoltbunkerClient:
    """Async client for the Moltbunker API.

    Usage::

        async with AsyncMoltbunkerClient(api_key="mb_live_...") as client:
            status = await client.get_status()
            print(status.peer_count)
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

        transport = httpx.AsyncHTTPTransport(retries=max_retries)
        self._client = httpx.AsyncClient(timeout=timeout, transport=transport)

    async def close(self) -> None:
        await self._client.aclose()

    async def __aenter__(self):
        # Auto-authenticate wallet sessions
        if isinstance(self._auth, WalletSessionAuth) and not self._auth.is_authenticated:
            await self._auth.authenticate(self._client, self._base_url)
        return self

    async def __aexit__(self, *args):
        await self.close()

    # --- HTTP helpers ---

    async def _request(self, method: str, path: str, **kwargs) -> Dict[str, Any]:
        url = f"{self._base_url}{path}"
        request = self._client.build_request(method, url, **kwargs)
        if self._auth:
            request = self._auth.apply(request)

        response = await self._client.send(request)
        if response.status_code >= 400:
            try:
                body = response.json()
            except Exception:
                body = None
            raise_for_status(response.status_code, body)

        if response.status_code == 204:
            return {}
        return response.json()

    async def _get(self, path: str, **kwargs) -> Dict[str, Any]:
        return await self._request("GET", path, **kwargs)

    async def _post(self, path: str, **kwargs) -> Dict[str, Any]:
        return await self._request("POST", path, **kwargs)

    async def _put(self, path: str, **kwargs) -> Dict[str, Any]:
        return await self._request("PUT", path, **kwargs)

    async def _delete(self, path: str, **kwargs) -> Dict[str, Any]:
        return await self._request("DELETE", path, **kwargs)

    # --- Status ---

    async def get_status(self) -> StatusResponse:
        return StatusResponse.model_validate(await self._get("/v1/status"))

    # --- Balance ---

    async def get_balance(self) -> BalanceResponse:
        return BalanceResponse.model_validate(await self._get("/v1/balance"))

    # --- Threat ---

    async def get_threat(self) -> ThreatResponse:
        return ThreatResponse.model_validate(await self._get("/v1/threat"))

    # --- Deploy ---

    async def deploy(self, request: DeployRequest) -> DeployResponse:
        data = await self._post("/v1/deploy", json=request.model_dump(exclude_none=True))
        return DeployResponse.model_validate(data)

    # --- Containers ---

    async def list_containers(self) -> List[ContainerInfo]:
        data = await self._get("/v1/containers")
        return [ContainerInfo.model_validate(c) for c in data.get("containers", [])]

    async def get_container(self, container_id: str) -> ContainerInfo:
        return ContainerInfo.model_validate(await self._get(f"/v1/containers/{container_id}"))

    async def stop_container(self, container_id: str) -> None:
        await self._post(f"/v1/containers/{container_id}", json={"action": "stop"})

    async def delete_container(self, container_id: str) -> None:
        await self._delete(f"/v1/containers/{container_id}")

    # --- Snapshots ---

    async def create_snapshot(self, request: SnapshotRequest) -> SnapshotResponse:
        data = await self._post("/v1/snapshot", json=request.model_dump())
        return SnapshotResponse.model_validate(data)

    async def list_snapshots(self) -> List[SnapshotResponse]:
        data = await self._get("/v1/snapshots")
        return [SnapshotResponse.model_validate(s) for s in data.get("snapshots", [])]

    # --- Clones ---

    async def clone(self, container_id: str, request: CloneRequest) -> CloneResponse:
        body = request.model_dump()
        body["container_id"] = container_id
        return CloneResponse.model_validate(await self._post("/v1/clone", json=body))

    # --- Molts (Serverless) ---

    async def deploy_molt(self, request: MoltDeployRequest) -> MoltDeployResponse:
        data = await self._post("/v1/molts", json=request.model_dump(exclude_none=True))
        return MoltDeployResponse.model_validate(data)

    async def list_molts(self) -> List[MoltInfo]:
        data = await self._get("/v1/molts")
        return [MoltInfo.model_validate(m) for m in data.get("molts", [])]

    async def get_molt(self, molt_id: str) -> MoltInfo:
        return MoltInfo.model_validate(await self._get(f"/v1/molts/{molt_id}"))

    async def invoke_molt(self, molt_id: str, request: MoltInvokeRequest) -> MoltInvokeResponse:
        data = await self._post(f"/v1/molts/{molt_id}/invoke", json=request.model_dump())
        return MoltInvokeResponse.model_validate(data)

    async def delete_molt(self, molt_id: str) -> None:
        await self._delete(f"/v1/molts/{molt_id}")

    # --- Subdomains ---

    async def list_subdomains(self) -> List[SubdomainInfo]:
        data = await self._get("/v1/subdomains")
        return [SubdomainInfo.model_validate(s) for s in data.get("subdomains", [])]

    async def register_subdomain(self, request: SubdomainRegisterRequest) -> SubdomainInfo:
        return SubdomainInfo.model_validate(await self._post("/v1/subdomains", json=request.model_dump()))

    async def delete_subdomain(self, name: str) -> None:
        await self._delete(f"/v1/subdomains/{name}")

    # --- Storage ---

    async def create_bucket(self, name: str) -> StorageBucket:
        return StorageBucket.model_validate(await self._post("/v1/storage/buckets", json={"name": name}))

    async def list_buckets(self) -> List[StorageBucket]:
        data = await self._get("/v1/storage/buckets")
        return [StorageBucket.model_validate(b) for b in data.get("buckets", data if isinstance(data, list) else [])]

    async def delete_bucket(self, name: str) -> None:
        await self._delete(f"/v1/storage/buckets/{name}")

    async def put_object(self, bucket: str, key: str, data: bytes, content_type: str = "") -> StorageObject:
        import base64

        body = {"bucket": bucket, "key": key, "data": base64.b64encode(data).decode(), "content_type": content_type}
        return StorageObject.model_validate(await self._put(f"/v1/storage/objects/{bucket}/{key}", json=body))

    async def get_object(self, bucket: str, key: str) -> bytes:
        import base64

        data = await self._get(f"/v1/storage/objects/{bucket}/{key}")
        return base64.b64decode(data.get("data", ""))

    async def delete_object(self, bucket: str, key: str) -> None:
        await self._delete(f"/v1/storage/objects/{bucket}/{key}")

    async def list_objects(self, bucket: str, prefix: str = "") -> List[StorageObject]:
        params = {"prefix": prefix} if prefix else {}
        data = await self._get(f"/v1/storage/objects/{bucket}", params=params)
        return [StorageObject.model_validate(o) for o in data.get("objects", [])]

    async def storage_usage(self) -> StorageUsage:
        return StorageUsage.model_validate(await self._get("/v1/storage/usage"))

    # --- Proxy ---

    async def list_proxy_sessions(self) -> List[ProxySession]:
        data = await self._get("/v1/proxy/sessions")
        return [ProxySession.model_validate(s) for s in data.get("sessions", [])]

    async def proxy_usage(self) -> ProxyUsage:
        return ProxyUsage.model_validate(await self._get("/v1/proxy/usage"))

    # --- Crawling ---

    async def create_crawl_job(self, request: CrawlJobRequest) -> CrawlJob:
        data = await self._post("/v1/crawl/jobs", json=request.model_dump())
        return CrawlJob.model_validate(data)

    async def list_crawl_jobs(self) -> List[CrawlJob]:
        data = await self._get("/v1/crawl/jobs")
        return [CrawlJob.model_validate(j) for j in data.get("jobs", [])]

    async def get_crawl_job(self, job_id: str) -> CrawlJob:
        return CrawlJob.model_validate(await self._get(f"/v1/crawl/jobs/{job_id}"))

    async def get_crawl_results(self, job_id: str) -> List[CrawlResult]:
        data = await self._get(f"/v1/crawl/jobs/{job_id}/results")
        return [CrawlResult.model_validate(r) for r in data.get("results", [])]

    async def cancel_crawl_job(self, job_id: str) -> None:
        await self._post(f"/v1/crawl/jobs/{job_id}/cancel")

    async def crawl_page(self, request: CrawlPageRequest) -> CrawlJob:
        data = await self._post("/v1/crawl/pages", json=request.model_dump())
        return CrawlJob.model_validate(data)

    async def get_crawl_stats(self) -> CrawlStats:
        return CrawlStats.model_validate(await self._get("/v1/crawl/stats"))

    # --- Agents ---

    async def deploy_agent(self, spec: AgentSpec) -> AgentDeployment:
        data = await self._post("/v1/agents", json=spec.model_dump())
        return AgentDeployment.model_validate(data)

    async def list_agents(self) -> List[AgentDeployment]:
        data = await self._get("/v1/agents")
        return [AgentDeployment.model_validate(a) for a in data.get("agents", [])]

    async def get_agent(self, agent_id: str) -> AgentDeployment:
        return AgentDeployment.model_validate(await self._get(f"/v1/agents/{agent_id}"))

    async def delete_agent(self, agent_id: str) -> None:
        await self._delete(f"/v1/agents/{agent_id}")

    async def invoke_agent(self, agent_id: str, request: AgentInvokeRequest) -> AgentInvokeResponse:
        data = await self._post(f"/v1/agents/{agent_id}/invoke", json=request.model_dump())
        return AgentInvokeResponse.model_validate(data)

    async def stop_agent(self, agent_id: str) -> None:
        await self._post(f"/v1/agents/{agent_id}/stop")

    async def list_agent_memory(self, agent_id: str) -> List[MemoryEntry]:
        data = await self._get(f"/v1/agents/{agent_id}/memory")
        return [MemoryEntry.model_validate(e) for e in data.get("entries", [])]

    async def set_agent_memory(self, agent_id: str, entry: MemoryEntry) -> None:
        await self._post(f"/v1/agents/{agent_id}/memory", json=entry.model_dump())

    async def delete_agent_memory(self, agent_id: str, key: str) -> None:
        await self._delete(f"/v1/agents/{agent_id}/memory", params={"key": key})
