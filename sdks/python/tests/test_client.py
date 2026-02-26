"""Unit tests for AnclaClient using pytest-httpx to mock HTTP responses."""

from __future__ import annotations

import json

import pytest
from pytest_httpx import HTTPXMock

from ancla import (
    AnclaClient,
    AnclaError,
    AuthenticationError,
    Build,
    Deploy,
    NotFoundError,
    ServerError,
    Service,
    Workspace,
)

# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------

SERVER = "https://test.ancla.dev"
API_KEY = "test-key-abc123"


@pytest.fixture()
def client() -> AnclaClient:
    """Return a client pointed at a fake server."""
    return AnclaClient(api_key=API_KEY, server=SERVER)


# ---------------------------------------------------------------------------
# Client initialisation
# ---------------------------------------------------------------------------


class TestClientInit:
    """Verify constructor defaults and env-var fallback."""

    def test_default_server(self) -> None:
        c = AnclaClient(api_key="k")
        assert c.server == "https://ancla.dev"

    def test_custom_server_strips_trailing_slash(self) -> None:
        c = AnclaClient(api_key="k", server="https://example.com/")
        assert c.server == "https://example.com"

    def test_api_key_from_env(self, monkeypatch: pytest.MonkeyPatch) -> None:
        monkeypatch.setenv("ANCLA_API_KEY", "env-key")
        c = AnclaClient()
        assert c.api_key == "env-key"

    def test_server_from_env(self, monkeypatch: pytest.MonkeyPatch) -> None:
        monkeypatch.setenv("ANCLA_SERVER", "https://env.ancla.dev")
        c = AnclaClient(api_key="k")
        assert c.server == "https://env.ancla.dev"

    def test_explicit_overrides_env(self, monkeypatch: pytest.MonkeyPatch) -> None:
        monkeypatch.setenv("ANCLA_API_KEY", "env-key")
        c = AnclaClient(api_key="explicit-key")
        assert c.api_key == "explicit-key"

    def test_context_manager(self) -> None:
        with AnclaClient(api_key="k") as c:
            assert c.api_key == "k"


# ---------------------------------------------------------------------------
# Workspaces
# ---------------------------------------------------------------------------


class TestWorkspaces:
    """Cover list, get, create for workspaces."""

    def test_list_workspaces(self, client: AnclaClient, httpx_mock: HTTPXMock) -> None:
        payload = [
            {
                "id": "ws-1",
                "name": "Acme",
                "slug": "acme",
                "member_count": 3,
                "project_count": 2,
                "service_count": 5,
            },
        ]
        httpx_mock.add_response(
            url=f"{SERVER}/api/v1/workspaces/",
            method="GET",
            json=payload,
        )

        workspaces = client.list_workspaces()
        assert len(workspaces) == 1
        assert isinstance(workspaces[0], Workspace)
        assert workspaces[0].slug == "acme"
        assert workspaces[0].member_count == 3
        assert workspaces[0].service_count == 5

    def test_get_workspace(self, client: AnclaClient, httpx_mock: HTTPXMock) -> None:
        payload = {
            "id": "ws-1",
            "name": "Acme",
            "slug": "acme",
            "project_count": 2,
            "service_count": 5,
            "members": [
                {"username": "alice", "email": "alice@example.com", "admin": True},
            ],
        }
        httpx_mock.add_response(
            url=f"{SERVER}/api/v1/workspaces/acme",
            method="GET",
            json=payload,
        )

        ws = client.get_workspace("acme")
        assert ws.name == "Acme"
        assert len(ws.members) == 1
        assert ws.members[0].admin is True

    def test_create_workspace(self, client: AnclaClient, httpx_mock: HTTPXMock) -> None:
        payload = {"id": "ws-new", "name": "NewWS", "slug": "newws"}
        httpx_mock.add_response(
            url=f"{SERVER}/api/v1/workspaces/",
            method="POST",
            json=payload,
        )

        ws = client.create_workspace("NewWS")
        assert ws.slug == "newws"

        request = httpx_mock.get_requests()[0]
        body = json.loads(request.content)
        assert body == {"name": "NewWS"}


# ---------------------------------------------------------------------------
# Services
# ---------------------------------------------------------------------------


class TestServices:
    """Spot-check service endpoints."""

    def test_list_services(self, client: AnclaClient, httpx_mock: HTTPXMock) -> None:
        payload = [
            {"name": "Web API", "slug": "web-api", "platform": "docker"},
        ]
        httpx_mock.add_response(
            url=f"{SERVER}/api/v1/workspaces/acme/projects/web/envs/production/services/",
            method="GET",
            json=payload,
        )

        services = client.list_services("acme", "web", "production")
        assert len(services) == 1
        assert isinstance(services[0], Service)
        assert services[0].platform == "docker"

    def test_deploy_service(self, client: AnclaClient, httpx_mock: HTTPXMock) -> None:
        httpx_mock.add_response(
            url=f"{SERVER}/api/v1/workspaces/acme/projects/web/envs/production/services/web-api/deploy",
            method="POST",
            json={"build_id": "bld-123"},
        )

        result = client.deploy_service("acme", "web", "production", "web-api")
        assert result.build_id == "bld-123"


# ---------------------------------------------------------------------------
# Builds
# ---------------------------------------------------------------------------


class TestBuilds:
    """Spot-check build endpoints."""

    def test_list_builds(self, client: AnclaClient, httpx_mock: HTTPXMock) -> None:
        payload = {
            "items": [
                {
                    "id": "bld-1",
                    "version": 1,
                    "built": True,
                    "error": False,
                    "created": "2025-01-01",
                },
            ],
        }
        httpx_mock.add_response(
            url=f"{SERVER}/api/v1/workspaces/acme/projects/web/envs/production/services/web-api/builds/",
            method="GET",
            json=payload,
        )

        builds = client.list_builds("acme", "web", "production", "web-api")
        assert len(builds) == 1
        assert isinstance(builds[0], Build)
        assert builds[0].built is True


# ---------------------------------------------------------------------------
# Deploys
# ---------------------------------------------------------------------------


class TestDeploys:
    """Spot-check deploy endpoints."""

    def test_get_deploy(self, client: AnclaClient, httpx_mock: HTTPXMock) -> None:
        payload = {
            "id": "dpl-1",
            "complete": True,
            "error": False,
            "error_detail": "",
            "job_id": "job-abc",
            "created": "2025-01-01",
            "updated": "2025-01-01",
        }
        httpx_mock.add_response(
            url=f"{SERVER}/api/v1/deploys/dpl-1/detail",
            method="GET",
            json=payload,
        )

        deploy = client.get_deploy("dpl-1")
        assert isinstance(deploy, Deploy)
        assert deploy.complete is True
        assert deploy.job_id == "job-abc"


# ---------------------------------------------------------------------------
# Error handling
# ---------------------------------------------------------------------------


class TestErrors:
    """Verify HTTP error codes map to the correct exception types."""

    def test_401_raises_authentication_error(self, client: AnclaClient, httpx_mock: HTTPXMock) -> None:
        httpx_mock.add_response(
            url=f"{SERVER}/api/v1/workspaces/",
            status_code=401,
            json={"message": "Invalid API key"},
        )

        with pytest.raises(AuthenticationError, match="Invalid API key"):
            client.list_workspaces()

    def test_404_raises_not_found_error(self, client: AnclaClient, httpx_mock: HTTPXMock) -> None:
        httpx_mock.add_response(
            url=f"{SERVER}/api/v1/workspaces/missing",
            status_code=404,
            json={"detail": "Workspace not found"},
        )

        with pytest.raises(NotFoundError, match="Workspace not found"):
            client.get_workspace("missing")

    def test_500_raises_server_error(self, client: AnclaClient, httpx_mock: HTTPXMock) -> None:
        httpx_mock.add_response(
            url=f"{SERVER}/api/v1/workspaces/",
            status_code=500,
            text="Internal Server Error",
        )

        with pytest.raises(ServerError):
            client.list_workspaces()

    def test_422_raises_validation_error(self, client: AnclaClient, httpx_mock: HTTPXMock) -> None:
        from ancla import ValidationError as AnclaValidationError

        httpx_mock.add_response(
            url=f"{SERVER}/api/v1/workspaces/",
            method="POST",
            status_code=422,
            json={"detail": "Name is required"},
        )

        with pytest.raises(AnclaValidationError, match="Name is required"):
            client.create_workspace("")

    def test_generic_4xx_raises_ancla_error(self, client: AnclaClient, httpx_mock: HTTPXMock) -> None:
        httpx_mock.add_response(
            url=f"{SERVER}/api/v1/workspaces/",
            status_code=429,
            json={"message": "Rate limited"},
        )

        with pytest.raises(AnclaError, match="Rate limited"):
            client.list_workspaces()

    def test_auth_header_sent(self, client: AnclaClient, httpx_mock: HTTPXMock) -> None:
        httpx_mock.add_response(
            url=f"{SERVER}/api/v1/workspaces/",
            json=[],
        )

        client.list_workspaces()
        request = httpx_mock.get_requests()[0]
        assert request.headers["X-API-Key"] == API_KEY
