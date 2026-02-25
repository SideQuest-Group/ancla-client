"""Unit tests for AnclaClient using pytest-httpx to mock HTTP responses."""

from __future__ import annotations

import json
import os

import pytest
from pytest_httpx import HTTPXMock

from ancla import (
    AnclaClient,
    AnclaError,
    AuthenticationError,
    NotFoundError,
    Org,
    ServerError,
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
# Organizations CRUD
# ---------------------------------------------------------------------------


class TestOrgs:
    """Cover list, get, create, update, delete for organizations."""

    def test_list_orgs(self, client: AnclaClient, httpx_mock: HTTPXMock) -> None:
        payload = [
            {
                "id": "org-1",
                "name": "Acme",
                "slug": "acme",
                "member_count": 3,
                "project_count": 2,
            },
        ]
        httpx_mock.add_response(
            url=f"{SERVER}/api/v1/organizations/",
            method="GET",
            json=payload,
        )

        orgs = client.list_orgs()
        assert len(orgs) == 1
        assert isinstance(orgs[0], Org)
        assert orgs[0].slug == "acme"
        assert orgs[0].member_count == 3

    def test_get_org(self, client: AnclaClient, httpx_mock: HTTPXMock) -> None:
        payload = {
            "id": "org-1",
            "name": "Acme",
            "slug": "acme",
            "project_count": 2,
            "application_count": 5,
            "members": [
                {"username": "alice", "email": "alice@example.com", "admin": True},
            ],
        }
        httpx_mock.add_response(
            url=f"{SERVER}/api/v1/organizations/acme",
            method="GET",
            json=payload,
        )

        org = client.get_org("acme")
        assert org.name == "Acme"
        assert len(org.members) == 1
        assert org.members[0].admin is True

    def test_create_org(self, client: AnclaClient, httpx_mock: HTTPXMock) -> None:
        payload = {"id": "org-new", "name": "NewOrg", "slug": "neworg"}
        httpx_mock.add_response(
            url=f"{SERVER}/api/v1/organizations/",
            method="POST",
            json=payload,
        )

        org = client.create_org("NewOrg")
        assert org.slug == "neworg"

        # Verify request body
        request = httpx_mock.get_requests()[0]
        body = json.loads(request.content)
        assert body == {"name": "NewOrg"}

    def test_update_org(self, client: AnclaClient, httpx_mock: HTTPXMock) -> None:
        payload = {"id": "org-1", "name": "Renamed", "slug": "acme"}
        httpx_mock.add_response(
            url=f"{SERVER}/api/v1/organizations/acme",
            method="PATCH",
            json=payload,
        )

        org = client.update_org("acme", "Renamed")
        assert org.name == "Renamed"

    def test_delete_org(self, client: AnclaClient, httpx_mock: HTTPXMock) -> None:
        httpx_mock.add_response(
            url=f"{SERVER}/api/v1/organizations/acme",
            method="DELETE",
            status_code=204,
        )

        # Should not raise
        client.delete_org("acme")


# ---------------------------------------------------------------------------
# Applications
# ---------------------------------------------------------------------------


class TestApps:
    """Spot-check application endpoints."""

    def test_list_apps(self, client: AnclaClient, httpx_mock: HTTPXMock) -> None:
        payload = [
            {"name": "My App", "slug": "my-app", "platform": "docker"},
        ]
        httpx_mock.add_response(
            url=f"{SERVER}/api/v1/applications/acme/web",
            method="GET",
            json=payload,
        )

        apps = client.list_apps("acme", "web")
        assert len(apps) == 1
        assert apps[0].platform == "docker"

    def test_deploy_app(self, client: AnclaClient, httpx_mock: HTTPXMock) -> None:
        httpx_mock.add_response(
            url=f"{SERVER}/api/v1/applications/acme/web/my-app/deploy",
            method="POST",
            json={"image_id": "img-123"},
        )

        result = client.deploy_app("acme", "web", "my-app")
        assert result.image_id == "img-123"

    def test_scale_app(self, client: AnclaClient, httpx_mock: HTTPXMock) -> None:
        httpx_mock.add_response(
            url=f"{SERVER}/api/v1/applications/acme/web/my-app/scale",
            method="POST",
            status_code=200,
            json={},
        )

        client.scale_app("acme", "web", "my-app", {"web": 2, "worker": 1})

        request = httpx_mock.get_requests()[0]
        body = json.loads(request.content)
        assert body == {"process_counts": {"web": 2, "worker": 1}}


# ---------------------------------------------------------------------------
# Images & Releases (list returns {items: [...]})
# ---------------------------------------------------------------------------


class TestImages:
    def test_list_images(self, client: AnclaClient, httpx_mock: HTTPXMock) -> None:
        payload = {
            "items": [
                {
                    "id": "img-1",
                    "version": 1,
                    "built": True,
                    "error": False,
                    "created": "2025-01-01",
                },
            ],
        }
        httpx_mock.add_response(
            url=f"{SERVER}/api/v1/images/acme/web/my-app",
            method="GET",
            json=payload,
        )

        images = client.list_images("acme", "web", "my-app")
        assert len(images) == 1
        assert images[0].built is True


class TestReleases:
    def test_create_release(self, client: AnclaClient, httpx_mock: HTTPXMock) -> None:
        httpx_mock.add_response(
            url=f"{SERVER}/api/v1/releases/acme/web/my-app/create",
            method="POST",
            json={"release_id": "rel-1", "version": 3},
        )

        result = client.create_release("acme", "web", "my-app", "img-1")
        assert result.release_id == "rel-1"
        assert result.version == 3


# ---------------------------------------------------------------------------
# Error handling
# ---------------------------------------------------------------------------


class TestErrors:
    """Verify HTTP error codes map to the correct exception types."""

    def test_401_raises_authentication_error(
        self, client: AnclaClient, httpx_mock: HTTPXMock
    ) -> None:
        httpx_mock.add_response(
            url=f"{SERVER}/api/v1/organizations/",
            status_code=401,
            json={"message": "Invalid API key"},
        )

        with pytest.raises(AuthenticationError, match="Invalid API key"):
            client.list_orgs()

    def test_404_raises_not_found_error(
        self, client: AnclaClient, httpx_mock: HTTPXMock
    ) -> None:
        httpx_mock.add_response(
            url=f"{SERVER}/api/v1/organizations/missing",
            status_code=404,
            json={"detail": "Organization not found"},
        )

        with pytest.raises(NotFoundError, match="Organization not found"):
            client.get_org("missing")

    def test_500_raises_server_error(
        self, client: AnclaClient, httpx_mock: HTTPXMock
    ) -> None:
        httpx_mock.add_response(
            url=f"{SERVER}/api/v1/organizations/",
            status_code=500,
            text="Internal Server Error",
        )

        with pytest.raises(ServerError):
            client.list_orgs()

    def test_422_raises_validation_error(
        self, client: AnclaClient, httpx_mock: HTTPXMock
    ) -> None:
        from ancla import ValidationError as AnclaValidationError

        httpx_mock.add_response(
            url=f"{SERVER}/api/v1/organizations/",
            method="POST",
            status_code=422,
            json={"detail": "Name is required"},
        )

        with pytest.raises(AnclaValidationError, match="Name is required"):
            client.create_org("")

    def test_generic_4xx_raises_ancla_error(
        self, client: AnclaClient, httpx_mock: HTTPXMock
    ) -> None:
        httpx_mock.add_response(
            url=f"{SERVER}/api/v1/organizations/",
            status_code=429,
            json={"message": "Rate limited"},
        )

        with pytest.raises(AnclaError, match="Rate limited"):
            client.list_orgs()

    def test_auth_header_sent(
        self, client: AnclaClient, httpx_mock: HTTPXMock
    ) -> None:
        httpx_mock.add_response(
            url=f"{SERVER}/api/v1/organizations/",
            json=[],
        )

        client.list_orgs()
        request = httpx_mock.get_requests()[0]
        assert request.headers["X-API-Key"] == API_KEY
