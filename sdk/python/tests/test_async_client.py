"""Tests for the async MoltbunkerClient."""

import json

import httpx
import pytest

from moltbunker.async_client import AsyncMoltbunkerClient
from moltbunker.exceptions import AuthenticationError, NotFoundError
from moltbunker.models import DeployRequest


def make_async_client(handler):
    """Create an async client with a mock transport."""
    transport = httpx.MockTransport(handler)
    client = AsyncMoltbunkerClient.__new__(AsyncMoltbunkerClient)
    client._base_url = "http://test"
    client._auth = None
    client._client = httpx.AsyncClient(transport=transport)
    return client


@pytest.mark.asyncio
async def test_get_status():
    def handler(request):
        return httpx.Response(200, json={
            "node_id": "abc123",
            "running": True,
            "peer_count": 3,
        })

    client = make_async_client(handler)
    status = await client.get_status()
    assert status.node_id == "abc123"
    assert status.peer_count == 3
    await client.close()


@pytest.mark.asyncio
async def test_deploy():
    def handler(request):
        body = json.loads(request.content)
        assert body["image"] == "redis:latest"
        return httpx.Response(201, json={
            "container_id": "ctr-002",
            "status": "deploying",
        })

    client = make_async_client(handler)
    resp = await client.deploy(DeployRequest(image="redis:latest"))
    assert resp.container_id == "ctr-002"
    await client.close()


@pytest.mark.asyncio
async def test_list_containers():
    def handler(request):
        return httpx.Response(200, json={
            "containers": [{"id": "c1", "status": "running"}]
        })

    client = make_async_client(handler)
    containers = await client.list_containers()
    assert len(containers) == 1
    await client.close()


@pytest.mark.asyncio
async def test_error_handling():
    def handler(request):
        return httpx.Response(401, json={"error": "unauthorized"})

    client = make_async_client(handler)
    with pytest.raises(AuthenticationError):
        await client.get_status()
    await client.close()


@pytest.mark.asyncio
async def test_list_molts():
    def handler(request):
        return httpx.Response(200, json={
            "molts": [
                {"id": "m1", "status": "running", "module_cid": "Qm..."},
                {"id": "m2", "status": "stopped"},
            ]
        })

    client = make_async_client(handler)
    molts = await client.list_molts()
    assert len(molts) == 2
    assert molts[0].module_cid == "Qm..."
    await client.close()


@pytest.mark.asyncio
async def test_storage_create_bucket():
    def handler(request):
        return httpx.Response(201, json={"name": "data", "owner": "0xabc"})

    client = make_async_client(handler)
    bucket = await client.create_bucket("data")
    assert bucket.name == "data"
    await client.close()


@pytest.mark.asyncio
async def test_storage_usage():
    def handler(request):
        return httpx.Response(200, json={
            "wallet_address": "0x1234",
            "total_bytes": 2048,
            "object_count": 10,
            "bucket_count": 3,
        })

    client = make_async_client(handler)
    usage = await client.storage_usage()
    assert usage.total_bytes == 2048
    assert usage.bucket_count == 3
    await client.close()


@pytest.mark.asyncio
async def test_proxy_usage():
    def handler(request):
        return httpx.Response(200, json={
            "wallet": "0x1234",
            "total_bytes_in": 1000,
            "total_bytes_out": 2000,
            "session_count": 5,
        })

    client = make_async_client(handler)
    usage = await client.proxy_usage()
    assert usage.total_bytes_in == 1000
    assert usage.session_count == 5
    await client.close()


@pytest.mark.asyncio
async def test_subdomains():
    def handler(request):
        return httpx.Response(200, json={
            "subdomains": [{"name": "myapp", "deployment_id": "d1", "url": "https://myapp.moltbunker.dev"}]
        })

    client = make_async_client(handler)
    subs = await client.list_subdomains()
    assert len(subs) == 1
    assert subs[0].name == "myapp"
    await client.close()


@pytest.mark.asyncio
async def test_404():
    def handler(request):
        return httpx.Response(404, json={"error": "container not found"})

    client = make_async_client(handler)
    with pytest.raises(NotFoundError):
        await client.get_container("nonexistent")
    await client.close()
