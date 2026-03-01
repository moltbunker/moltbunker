"""Pydantic models for Moltbunker API requests and responses."""

from __future__ import annotations

from datetime import datetime
from typing import Any, Dict, List, Optional

from pydantic import BaseModel, Field


# --- Status ---


class SecurityInfo(BaseModel):
    tls_version: str = ""
    cert_pinning: bool = False
    encryption: str = ""


class StatusResponse(BaseModel):
    node_id: str = ""
    running: bool = False
    version: str = ""
    uptime: str = ""
    peer_count: int = 0
    containers: int = 0
    tor_enabled: bool = False
    tor_address: str = ""
    region: str = ""
    threat_level: float = 0.0
    network_capacity: int = 0
    security: Optional[SecurityInfo] = None
    node_tier: str = ""
    reputation_score: int = 0
    known_nodes: int = 0


# --- Deploy ---


class ResourceSpec(BaseModel):
    cpu_cores: int = Field(default=1, ge=1, le=64)
    memory_mb: int = Field(default=512, ge=128)
    disk_gb: int = Field(default=10, ge=1)
    gpu_type: str = ""


class DeployRequest(BaseModel):
    image: str
    resources: ResourceSpec = Field(default_factory=ResourceSpec)
    tor_only: bool = False
    onion_service: bool = False
    reservation_id: str = ""
    min_provider_tier: str = ""
    spot: bool = False
    expose_ports: List[int] = Field(default_factory=list)
    env: Dict[str, str] = Field(default_factory=dict)


class DeployResponse(BaseModel):
    container_id: str
    status: str
    onion_address: str = ""
    regions: List[str] = Field(default_factory=list)
    replica_count: int = 0
    created_at: Optional[datetime] = None


# --- Containers ---


class ContainerInfo(BaseModel):
    id: str
    image: str = ""
    status: str = ""
    created_at: Optional[datetime] = None
    encrypted: bool = False
    onion_address: str = ""
    regions: List[str] = Field(default_factory=list)
    owner: str = ""
    has_volume: bool = False


# --- Balance ---


class BalanceResponse(BaseModel):
    wallet_address: str
    bunker_balance: str = "0"
    eth_balance: str = "0"
    deposited: str = "0"
    reserved: str = "0"
    available: str = "0"


# --- Snapshots ---


class SnapshotRequest(BaseModel):
    container_id: str
    type: str = "full"
    metadata: Dict[str, str] = Field(default_factory=dict)


class SnapshotResponse(BaseModel):
    snapshot_id: str
    container_id: str
    type: str = ""
    size: int = 0
    checksum: str = ""
    created_at: Optional[datetime] = None


# --- Clones ---


class CloneRequest(BaseModel):
    code_hash: str = ""
    state_snapshot: str = ""
    target_region: str = ""
    reason: str = ""


class CloneResponse(BaseModel):
    clone_id: str
    source_id: str = ""
    target_id: str = ""
    status: str = ""
    target_region: str = ""
    created_at: Optional[datetime] = None


# --- Molts (Serverless Functions) ---


class MoltDeployRequest(BaseModel):
    wasm_bytes: Optional[bytes] = None
    module_cid: str = ""
    memory_limit_mb: int = 64
    timeout_ms: int = 5000
    max_instances: int = 10
    environment: Dict[str, str] = Field(default_factory=dict)


class MoltDeployResponse(BaseModel):
    deployment_id: str
    module_cid: str = ""
    status: str = ""


class MoltInvokeRequest(BaseModel):
    method: str = "GET"
    path: str = "/"
    headers: Dict[str, str] = Field(default_factory=dict)
    body: str = ""


class MoltInvokeResponse(BaseModel):
    status_code: int = 200
    headers: Dict[str, str] = Field(default_factory=dict)
    body: str = ""
    duration_ms: float = 0.0
    error: str = ""


class MoltInfo(BaseModel):
    id: str
    module_cid: str = ""
    status: str = ""
    created_at: Optional[datetime] = None
    owner: str = ""
    memory_limit_mb: int = 0
    timeout_ms: int = 0
    metrics: Dict[str, Any] = Field(default_factory=dict)


# --- Subdomains ---


class SubdomainRegisterRequest(BaseModel):
    name: str
    deployment_id: str


class SubdomainInfo(BaseModel):
    name: str
    deployment_id: str = ""
    owner: str = ""
    url: str = ""
    registered_at: Optional[datetime] = None


# --- Storage ---


class StorageBucket(BaseModel):
    name: str
    owner: str = ""
    created_at: Optional[datetime] = None


class StorageObject(BaseModel):
    bucket: str = ""
    key: str = ""
    size: int = 0
    content_type: str = ""
    etag: str = ""
    owner: str = ""
    created_at: Optional[datetime] = None
    updated_at: Optional[datetime] = None


class StorageUsage(BaseModel):
    wallet_address: str = ""
    total_bytes: int = 0
    object_count: int = 0
    bucket_count: int = 0


# --- Proxy ---


class ProxySession(BaseModel):
    id: str = ""
    wallet: str = ""
    protocol: str = ""
    target: str = ""
    bytes_in: int = 0
    bytes_out: int = 0
    started_at: Optional[datetime] = None


class ProxyUsage(BaseModel):
    wallet: str = ""
    total_bytes_in: int = 0
    total_bytes_out: int = 0
    session_count: int = 0


# --- Threat ---


class ThreatResponse(BaseModel):
    score: float = 0.0
    level: str = ""
    active_signals: List[str] = Field(default_factory=list)
    recommendation: str = ""
    timestamp: Optional[datetime] = None


# --- Crawling ---


class CrawlConfig(BaseModel):
    urls: List[str] = Field(default_factory=list)
    max_depth: int = 0
    max_pages: int = 100
    allowed_domains: List[str] = Field(default_factory=list)
    selectors: List[str] = Field(default_factory=list)
    screenshot: bool = False
    javascript: bool = False
    user_agent: str = ""
    headers: Dict[str, str] = Field(default_factory=dict)
    timeout_sec: int = 0
    respect_robots: bool = False
    use_tor: bool = False
    storage_bucket: str = ""


class CrawlJobRequest(BaseModel):
    urls: List[str]
    max_depth: int = 0
    max_pages: int = 100
    allowed_domains: List[str] = Field(default_factory=list)
    selectors: List[str] = Field(default_factory=list)
    screenshot: bool = False
    javascript: bool = False
    user_agent: str = ""
    headers: Dict[str, str] = Field(default_factory=dict)
    timeout_sec: int = 0
    respect_robots: bool = False
    use_tor: bool = False
    storage_bucket: str = ""


class CrawlResult(BaseModel):
    url: str = ""
    status_code: int = 0
    content_type: str = ""
    title: str = ""
    html: str = ""
    text: str = ""
    links: List[str] = Field(default_factory=list)
    selectors: Dict[str, str] = Field(default_factory=dict)
    screenshot_cid: str = ""
    crawled_at: Optional[datetime] = None
    duration_ms: int = 0
    error: str = ""
    byte_size: int = 0


class CrawlJob(BaseModel):
    id: str
    owner: str = ""
    status: str = ""
    config: Optional[CrawlConfig] = None
    created_at: Optional[datetime] = None
    started_at: Optional[datetime] = None
    completed_at: Optional[datetime] = None
    error: str = ""
    pages_crawled: int = 0
    total_bytes: int = 0
    results: List[CrawlResult] = Field(default_factory=list)


class CrawlPageRequest(BaseModel):
    url: str
    selectors: List[str] = Field(default_factory=list)
    screenshot: bool = False
    javascript: bool = False
    user_agent: str = ""
    headers: Dict[str, str] = Field(default_factory=dict)
    timeout_sec: int = 0


class CrawlStats(BaseModel):
    total_jobs: int = 0
    running_jobs: int = 0
    completed_jobs: int = 0
    failed_jobs: int = 0
    total_pages_crawled: int = 0
    total_bytes: int = 0


# --- Agents ---


class MCPToolDef(BaseModel):
    name: str = ""
    description: str = ""
    parameters: Dict[str, str] = Field(default_factory=dict)


class AgentSpec(BaseModel):
    name: str
    framework: str = "custom"
    image: str = ""
    config: Dict[str, str] = Field(default_factory=dict)
    env: Dict[str, str] = Field(default_factory=dict)
    mcp_tools: List[MCPToolDef] = Field(default_factory=list)
    memory_bucket: str = ""
    max_tokens: int = 0
    timeout_sec: int = 0
    memory_limit_mb: int = 0
    cpu_cores: int = 0


class AgentDeployment(BaseModel):
    id: str
    spec: Optional[AgentSpec] = None
    status: str = ""
    container_id: str = ""
    node_id: str = ""
    created_at: Optional[datetime] = None
    started_at: Optional[datetime] = None
    stopped_at: Optional[datetime] = None
    error: str = ""
    tokens_used: int = 0
    invocation_count: int = 0
    total_cost_wei: str = ""


class AgentInvokeRequest(BaseModel):
    message: str
    context: Dict[str, str] = Field(default_factory=dict)


class AgentInvokeResponse(BaseModel):
    agent_id: str = ""
    response: str = ""
    tokens_used: int = 0
    duration_ms: int = 0
    error: str = ""


class MemoryEntry(BaseModel):
    key: str = ""
    value: str = ""
    updated_at: Optional[datetime] = None
