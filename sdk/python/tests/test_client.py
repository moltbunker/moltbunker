"""Tests for the synchronous MoltbunkerClient."""

import json

import httpx
import pytest

from moltbunker import MoltbunkerClient
from moltbunker.exceptions import AuthenticationError, NotFoundError, RateLimitError, ValidationError
from moltbunker.models import DeployRequest, MoltInvokeRequest, SnapshotRequest, SubdomainRegisterRequest


def make_client(handler):
    """Create a client with a mock transport."""
    transport = httpx.MockTransport(handler)
    client = MoltbunkerClient.__new__(MoltbunkerClient)
    client._base_url = "http://test"
    client._auth = None
    client._client = httpx.Client(transport=transport)
    return client


class TestGetStatus:
    def test_success(self):
        def handler(request):
            return httpx.Response(200, json={
                "node_id": "abc123",
                "running": True,
                "peer_count": 5,
                "version": "0.1.0",
            })

        client = make_client(handler)
        status = client.get_status()
        assert status.node_id == "abc123"
        assert status.running is True
        assert status.peer_count == 5


class TestGetBalance:
    def test_success(self):
        def handler(request):
            return httpx.Response(200, json={
                "wallet_address": "0x1234",
                "bunker_balance": "1000000",
                "available": "500000",
            })

        client = make_client(handler)
        balance = client.get_balance()
        assert balance.wallet_address == "0x1234"
        assert balance.bunker_balance == "1000000"


class TestDeploy:
    def test_success(self):
        def handler(request):
            body = json.loads(request.content)
            assert body["image"] == "nginx:latest"
            return httpx.Response(201, json={
                "container_id": "ctr-001",
                "status": "deploying",
                "regions": ["us-east"],
            })

        client = make_client(handler)
        resp = client.deploy(DeployRequest(image="nginx:latest"))
        assert resp.container_id == "ctr-001"
        assert resp.status == "deploying"


class TestContainers:
    def test_list(self):
        def handler(request):
            return httpx.Response(200, json={
                "containers": [
                    {"id": "c1", "image": "nginx", "status": "running"},
                    {"id": "c2", "image": "redis", "status": "stopped"},
                ]
            })

        client = make_client(handler)
        containers = client.list_containers()
        assert len(containers) == 2
        assert containers[0].id == "c1"

    def test_get(self):
        def handler(request):
            return httpx.Response(200, json={"id": "c1", "status": "running"})

        client = make_client(handler)
        c = client.get_container("c1")
        assert c.id == "c1"

    def test_stop(self):
        def handler(request):
            return httpx.Response(204)

        client = make_client(handler)
        client.stop_container("c1")

    def test_delete(self):
        def handler(request):
            return httpx.Response(204)

        client = make_client(handler)
        client.delete_container("c1")


class TestSnapshots:
    def test_create(self):
        def handler(request):
            return httpx.Response(201, json={
                "snapshot_id": "snap-1",
                "container_id": "c1",
                "type": "full",
                "size": 1024,
            })

        client = make_client(handler)
        snap = client.create_snapshot(SnapshotRequest(container_id="c1"))
        assert snap.snapshot_id == "snap-1"

    def test_list(self):
        def handler(request):
            return httpx.Response(200, json={"snapshots": []})

        client = make_client(handler)
        snaps = client.list_snapshots()
        assert snaps == []


class TestMolts:
    def test_list(self):
        def handler(request):
            return httpx.Response(200, json={
                "molts": [{"id": "m1", "status": "running"}]
            })

        client = make_client(handler)
        molts = client.list_molts()
        assert len(molts) == 1
        assert molts[0].id == "m1"

    def test_invoke(self):
        def handler(request):
            return httpx.Response(200, json={
                "status_code": 200,
                "body": "hello",
                "duration_ms": 12.5,
            })

        client = make_client(handler)
        resp = client.invoke_molt("m1", MoltInvokeRequest(deployment_id="m1"))
        assert resp.status_code == 200
        assert resp.body == "hello"


class TestSubdomains:
    def test_list(self):
        def handler(request):
            return httpx.Response(200, json={
                "subdomains": [{"name": "myapp", "deployment_id": "d1"}]
            })

        client = make_client(handler)
        subs = client.list_subdomains()
        assert len(subs) == 1
        assert subs[0].name == "myapp"

    def test_register(self):
        def handler(request):
            return httpx.Response(201, json={
                "name": "myapp",
                "deployment_id": "d1",
                "url": "https://myapp.moltbunker.dev",
            })

        client = make_client(handler)
        sub = client.register_subdomain(SubdomainRegisterRequest(name="myapp", deployment_id="d1"))
        assert sub.url == "https://myapp.moltbunker.dev"


class TestStorage:
    def test_create_bucket(self):
        def handler(request):
            return httpx.Response(201, json={"name": "test-bucket", "owner": "0x1234"})

        client = make_client(handler)
        bucket = client.create_bucket("test-bucket")
        assert bucket.name == "test-bucket"

    def test_list_buckets(self):
        def handler(request):
            return httpx.Response(200, json={"buckets": [{"name": "b1"}, {"name": "b2"}]})

        client = make_client(handler)
        buckets = client.list_buckets()
        assert len(buckets) == 2

    def test_usage(self):
        def handler(request):
            return httpx.Response(200, json={
                "wallet_address": "0x1234",
                "total_bytes": 1024,
                "object_count": 5,
                "bucket_count": 2,
            })

        client = make_client(handler)
        usage = client.storage_usage()
        assert usage.total_bytes == 1024
        assert usage.object_count == 5


class TestErrorHandling:
    def test_401(self):
        def handler(request):
            return httpx.Response(401, json={"error": "invalid token"})

        client = make_client(handler)
        with pytest.raises(AuthenticationError) as exc_info:
            client.get_status()
        assert exc_info.value.status_code == 401

    def test_404(self):
        def handler(request):
            return httpx.Response(404, json={"error": "not found"})

        client = make_client(handler)
        with pytest.raises(NotFoundError):
            client.get_container("nonexistent")

    def test_429(self):
        def handler(request):
            return httpx.Response(429, json={"error": "rate limited"})

        client = make_client(handler)
        with pytest.raises(RateLimitError):
            client.get_status()

    def test_400(self):
        def handler(request):
            return httpx.Response(400, json={"error": "bad request"})

        client = make_client(handler)
        with pytest.raises(ValidationError):
            client.deploy(DeployRequest(image=""))


class TestAuth:
    def test_api_key_header(self):
        def handler(request):
            assert request.headers["authorization"] == "Bearer mb_test_key_123"
            return httpx.Response(200, json={"node_id": "test"})

        from moltbunker.auth import APIKeyAuth
        transport = httpx.MockTransport(handler)
        client = MoltbunkerClient.__new__(MoltbunkerClient)
        client._base_url = "http://test"
        client._auth = APIKeyAuth("mb_test_key_123")
        client._client = httpx.Client(transport=transport)

        client.get_status()

    def test_invalid_api_key_prefix(self):
        from moltbunker.auth import APIKeyAuth
        with pytest.raises(ValueError, match="must start with"):
            APIKeyAuth("invalid_key")

    def test_no_auth(self):
        def handler(request):
            assert "authorization" not in request.headers
            return httpx.Response(200, json={"node_id": "test"})

        client = make_client(handler)
        client.get_status()
