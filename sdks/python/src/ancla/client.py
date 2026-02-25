"""Synchronous HTTP client for the Ancla PaaS API."""

from __future__ import annotations

import os
from typing import Any

import httpx

from ancla.exceptions import (
    AnclaError,
    AuthenticationError,
    NotFoundError,
    ServerError,
    ValidationError,
)
from ancla.models import (
    App,
    ConfigVar,
    CreateReleaseResult,
    Deployment,
    DeployReleaseResult,
    DeployResult,
    Image,
    ImageList,
    Org,
    Project,
    Release,
    ReleaseList,
)

_DEFAULT_SERVER = "https://ancla.dev"


class AnclaClient:
    """Client for the Ancla PaaS REST API.

    Args:
        api_key: API key for authentication.  Falls back to the
            ``ANCLA_API_KEY`` environment variable when *None*.
        server: Base URL of the Ancla server.  Falls back to the
            ``ANCLA_SERVER`` environment variable, then to
            ``https://ancla.dev``.
        timeout: Request timeout in seconds.
    """

    def __init__(
        self,
        api_key: str | None = None,
        server: str | None = None,
        timeout: float = 30.0,
    ) -> None:
        self.api_key = api_key or os.environ.get("ANCLA_API_KEY", "")
        self.server = (
            server
            or os.environ.get("ANCLA_SERVER")
            or _DEFAULT_SERVER
        ).rstrip("/")
        self._client = httpx.Client(
            base_url=f"{self.server}/api/v1",
            headers=self._build_headers(),
            timeout=timeout,
        )

    # ------------------------------------------------------------------
    # Internal helpers
    # ------------------------------------------------------------------

    def _build_headers(self) -> dict[str, str]:
        headers: dict[str, str] = {}
        if self.api_key:
            headers["X-API-Key"] = self.api_key
        return headers

    def _request(
        self,
        method: str,
        path: str,
        *,
        json: Any | None = None,
    ) -> httpx.Response:
        """Perform an HTTP request and raise on error status codes."""
        response = self._client.request(method, path, json=json)
        if response.status_code >= 400:
            self._raise_for_status(response)
        return response

    @staticmethod
    def _raise_for_status(response: httpx.Response) -> None:
        """Map HTTP error codes to SDK exceptions."""
        status = response.status_code
        detail: str | None = None
        try:
            body = response.json()
            detail = body.get("message") or body.get("detail")
        except Exception:
            detail = response.text or None

        message = detail or f"API request failed ({status})"

        if status == 401:
            raise AuthenticationError(message, status_code=status, detail=detail)
        if status == 404:
            raise NotFoundError(message, status_code=status, detail=detail)
        if status == 422:
            raise ValidationError(message, status_code=status, detail=detail)
        if status >= 500:
            raise ServerError(message, status_code=status, detail=detail)
        raise AnclaError(message, status_code=status, detail=detail)

    def close(self) -> None:
        """Close the underlying HTTP client."""
        self._client.close()

    def __enter__(self) -> AnclaClient:
        return self

    def __exit__(self, *args: object) -> None:
        self.close()

    # ------------------------------------------------------------------
    # Organizations
    # ------------------------------------------------------------------

    def list_orgs(self) -> list[Org]:
        """List all organizations the authenticated user belongs to."""
        resp = self._request("GET", "/organizations/")
        return [Org.model_validate(item) for item in resp.json()]

    def get_org(self, slug: str) -> Org:
        """Get details for a single organization by slug."""
        resp = self._request("GET", f"/organizations/{slug}")
        return Org.model_validate(resp.json())

    def create_org(self, name: str) -> Org:
        """Create a new organization."""
        resp = self._request("POST", "/organizations/", json={"name": name})
        return Org.model_validate(resp.json())

    def update_org(self, slug: str, name: str) -> Org:
        """Rename an organization."""
        resp = self._request("PATCH", f"/organizations/{slug}", json={"name": name})
        return Org.model_validate(resp.json())

    def delete_org(self, slug: str) -> None:
        """Delete an organization by slug."""
        self._request("DELETE", f"/organizations/{slug}")

    # ------------------------------------------------------------------
    # Projects
    # ------------------------------------------------------------------

    def list_projects(self, org: str) -> list[Project]:
        """List all projects in an organization."""
        resp = self._request("GET", f"/projects/{org}")
        return [Project.model_validate(item) for item in resp.json()]

    def get_project(self, org: str, slug: str) -> Project:
        """Get details for a single project."""
        resp = self._request("GET", f"/projects/{org}/{slug}")
        return Project.model_validate(resp.json())

    def create_project(self, org: str, name: str) -> Project:
        """Create a new project in an organization."""
        resp = self._request("POST", f"/projects/{org}", json={"name": name})
        return Project.model_validate(resp.json())

    def update_project(self, org: str, slug: str, name: str) -> Project:
        """Rename a project."""
        resp = self._request(
            "PATCH", f"/projects/{org}/{slug}", json={"name": name}
        )
        return Project.model_validate(resp.json())

    def delete_project(self, org: str, slug: str) -> None:
        """Delete a project."""
        self._request("DELETE", f"/projects/{org}/{slug}")

    # ------------------------------------------------------------------
    # Applications
    # ------------------------------------------------------------------

    def list_apps(self, org: str, project: str) -> list[App]:
        """List all applications in a project."""
        resp = self._request("GET", f"/applications/{org}/{project}")
        return [App.model_validate(item) for item in resp.json()]

    def get_app(self, org: str, project: str, slug: str) -> App:
        """Get details for a single application."""
        resp = self._request("GET", f"/applications/{org}/{project}/{slug}")
        return App.model_validate(resp.json())

    def create_app(
        self, org: str, project: str, name: str, platform: str
    ) -> App:
        """Create a new application in a project."""
        resp = self._request(
            "POST",
            f"/applications/{org}/{project}",
            json={"name": name, "platform": platform},
        )
        return App.model_validate(resp.json())

    def update_app(
        self, org: str, project: str, slug: str, **kwargs: Any
    ) -> App:
        """Update application attributes (name, platform, etc.)."""
        resp = self._request(
            "PATCH",
            f"/applications/{org}/{project}/{slug}",
            json=kwargs,
        )
        return App.model_validate(resp.json())

    def delete_app(self, org: str, project: str, slug: str) -> None:
        """Delete an application."""
        self._request("DELETE", f"/applications/{org}/{project}/{slug}")

    def deploy_app(self, org: str, project: str, slug: str) -> DeployResult:
        """Trigger a full deploy for an application."""
        resp = self._request(
            "POST", f"/applications/{org}/{project}/{slug}/deploy"
        )
        return DeployResult.model_validate(resp.json())

    def scale_app(
        self, org: str, project: str, slug: str, counts: dict[str, int]
    ) -> None:
        """Scale application processes.

        Args:
            counts: Mapping of process name to desired count,
                e.g. ``{"web": 2, "worker": 1}``.
        """
        self._request(
            "POST",
            f"/applications/{org}/{project}/{slug}/scale",
            json={"process_counts": counts},
        )

    # ------------------------------------------------------------------
    # Configuration
    # ------------------------------------------------------------------

    def list_config(self, org: str, project: str, app: str) -> list[ConfigVar]:
        """List configuration variables for an application."""
        resp = self._request(
            "GET", f"/configurations/{org}/{project}/{app}"
        )
        return [ConfigVar.model_validate(item) for item in resp.json()]

    def get_config(
        self, org: str, project: str, app: str, key: str
    ) -> ConfigVar:
        """Get a single configuration variable by key name."""
        resp = self._request(
            "GET", f"/configurations/{org}/{project}/{app}/{key}"
        )
        return ConfigVar.model_validate(resp.json())

    def set_config(
        self,
        org: str,
        project: str,
        app: str,
        key: str,
        value: str,
        secret: bool = False,
    ) -> ConfigVar:
        """Set (create or update) a configuration variable."""
        resp = self._request(
            "POST",
            f"/configurations/{org}/{project}/{app}",
            json={"name": key, "value": value, "secret": secret},
        )
        return ConfigVar.model_validate(resp.json())

    def delete_config(
        self, org: str, project: str, app: str, key: str
    ) -> None:
        """Delete a configuration variable."""
        self._request(
            "DELETE", f"/configurations/{org}/{project}/{app}/{key}"
        )

    # ------------------------------------------------------------------
    # Images
    # ------------------------------------------------------------------

    def list_images(
        self, org: str, project: str, app: str
    ) -> list[Image]:
        """List images for an application."""
        resp = self._request("GET", f"/images/{org}/{project}/{app}")
        result = ImageList.model_validate(resp.json())
        return result.items

    def get_image(
        self, org: str, project: str, app: str, image_id: str
    ) -> Image:
        """Get a single image by ID."""
        resp = self._request(
            "GET", f"/images/{org}/{project}/{app}/{image_id}"
        )
        return Image.model_validate(resp.json())

    # ------------------------------------------------------------------
    # Releases
    # ------------------------------------------------------------------

    def list_releases(
        self, org: str, project: str, app: str
    ) -> list[Release]:
        """List releases for an application."""
        resp = self._request("GET", f"/releases/{org}/{project}/{app}")
        result = ReleaseList.model_validate(resp.json())
        return result.items

    def get_release(
        self, org: str, project: str, app: str, release_id: str
    ) -> Release:
        """Get a single release by ID."""
        resp = self._request(
            "GET", f"/releases/{org}/{project}/{app}/{release_id}"
        )
        return Release.model_validate(resp.json())

    def create_release(
        self, org: str, project: str, app: str, image_id: str
    ) -> CreateReleaseResult:
        """Create a new release from an image."""
        resp = self._request(
            "POST",
            f"/releases/{org}/{project}/{app}/create",
            json={"image_id": image_id},
        )
        return CreateReleaseResult.model_validate(resp.json())

    # ------------------------------------------------------------------
    # Deployments
    # ------------------------------------------------------------------

    def list_deployments(
        self, org: str, project: str, app: str
    ) -> list[Deployment]:
        """List deployments for an application."""
        resp = self._request(
            "GET", f"/deployments/{org}/{project}/{app}"
        )
        return [Deployment.model_validate(item) for item in resp.json()]

    def get_deployment(
        self, org: str, project: str, app: str, deployment_id: str
    ) -> Deployment:
        """Get deployment details by ID."""
        resp = self._request(
            "GET", f"/deployments/{deployment_id}/detail"
        )
        return Deployment.model_validate(resp.json())
