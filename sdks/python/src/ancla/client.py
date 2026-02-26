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
    Build,
    BuildList,
    ConfigVar,
    Deploy,
    DeployList,
    DeployLog,
    DeployResult,
    Environment,
    Project,
    Service,
    Workspace,
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
        self.server = (server or os.environ.get("ANCLA_SERVER") or _DEFAULT_SERVER).rstrip("/")
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
    # Workspaces
    # ------------------------------------------------------------------

    def list_workspaces(self) -> list[Workspace]:
        """List all workspaces the authenticated user belongs to."""
        resp = self._request("GET", "/workspaces/")
        return [Workspace.model_validate(item) for item in resp.json()]

    def get_workspace(self, slug: str) -> Workspace:
        """Get details for a single workspace by slug."""
        resp = self._request("GET", f"/workspaces/{slug}")
        return Workspace.model_validate(resp.json())

    def create_workspace(self, name: str) -> Workspace:
        """Create a new workspace."""
        resp = self._request("POST", "/workspaces/", json={"name": name})
        return Workspace.model_validate(resp.json())

    def update_workspace(self, slug: str, name: str) -> Workspace:
        """Rename a workspace."""
        resp = self._request("PATCH", f"/workspaces/{slug}", json={"name": name})
        return Workspace.model_validate(resp.json())

    def delete_workspace(self, slug: str) -> None:
        """Delete a workspace by slug."""
        self._request("DELETE", f"/workspaces/{slug}")

    # ------------------------------------------------------------------
    # Projects
    # ------------------------------------------------------------------

    def list_projects(self, ws: str) -> list[Project]:
        """List all projects in a workspace."""
        resp = self._request("GET", f"/workspaces/{ws}/projects/")
        return [Project.model_validate(item) for item in resp.json()]

    def get_project(self, ws: str, slug: str) -> Project:
        """Get details for a single project."""
        resp = self._request("GET", f"/workspaces/{ws}/projects/{slug}")
        return Project.model_validate(resp.json())

    def create_project(self, ws: str, name: str) -> Project:
        """Create a new project in a workspace."""
        resp = self._request("POST", f"/workspaces/{ws}/projects/", json={"name": name})
        return Project.model_validate(resp.json())

    def update_project(self, ws: str, slug: str, name: str) -> Project:
        """Rename a project."""
        resp = self._request("PATCH", f"/workspaces/{ws}/projects/{slug}", json={"name": name})
        return Project.model_validate(resp.json())

    def delete_project(self, ws: str, slug: str) -> None:
        """Delete a project."""
        self._request("DELETE", f"/workspaces/{ws}/projects/{slug}")

    # ------------------------------------------------------------------
    # Environments
    # ------------------------------------------------------------------

    def _env_base(self, ws: str, proj: str) -> str:
        return f"/workspaces/{ws}/projects/{proj}/envs"

    def list_envs(self, ws: str, proj: str) -> list[Environment]:
        """List all environments in a project."""
        resp = self._request("GET", f"{self._env_base(ws, proj)}/")
        return [Environment.model_validate(item) for item in resp.json()]

    def get_env(self, ws: str, proj: str, slug: str) -> Environment:
        """Get details for a single environment."""
        resp = self._request("GET", f"{self._env_base(ws, proj)}/{slug}")
        return Environment.model_validate(resp.json())

    def create_env(self, ws: str, proj: str, name: str) -> Environment:
        """Create a new environment in a project."""
        resp = self._request("POST", f"{self._env_base(ws, proj)}/", json={"name": name})
        return Environment.model_validate(resp.json())

    # ------------------------------------------------------------------
    # Services
    # ------------------------------------------------------------------

    def _svc_base(self, ws: str, proj: str, env: str) -> str:
        return f"/workspaces/{ws}/projects/{proj}/envs/{env}/services"

    def list_services(self, ws: str, proj: str, env: str) -> list[Service]:
        """List all services in an environment."""
        resp = self._request("GET", f"{self._svc_base(ws, proj, env)}/")
        return [Service.model_validate(item) for item in resp.json()]

    def get_service(self, ws: str, proj: str, env: str, slug: str) -> Service:
        """Get details for a single service."""
        resp = self._request("GET", f"{self._svc_base(ws, proj, env)}/{slug}")
        return Service.model_validate(resp.json())

    def create_service(self, ws: str, proj: str, env: str, name: str, platform: str) -> Service:
        """Create a new service in an environment."""
        resp = self._request(
            "POST",
            f"{self._svc_base(ws, proj, env)}/",
            json={"name": name, "platform": platform},
        )
        return Service.model_validate(resp.json())

    def update_service(self, ws: str, proj: str, env: str, slug: str, **kwargs: Any) -> Service:
        """Update service attributes (name, platform, etc.)."""
        resp = self._request(
            "PATCH",
            f"{self._svc_base(ws, proj, env)}/{slug}",
            json=kwargs,
        )
        return Service.model_validate(resp.json())

    def delete_service(self, ws: str, proj: str, env: str, slug: str) -> None:
        """Delete a service."""
        self._request("DELETE", f"{self._svc_base(ws, proj, env)}/{slug}")

    def deploy_service(self, ws: str, proj: str, env: str, slug: str) -> DeployResult:
        """Trigger a full deploy for a service."""
        resp = self._request("POST", f"{self._svc_base(ws, proj, env)}/{slug}/deploy")
        return DeployResult.model_validate(resp.json())

    def scale_service(
        self,
        ws: str,
        proj: str,
        env: str,
        slug: str,
        counts: dict[str, int],
    ) -> None:
        """Scale service processes.

        Args:
            counts: Mapping of process name to desired count,
                e.g. ``{"web": 2, "worker": 1}``.
        """
        self._request(
            "POST",
            f"{self._svc_base(ws, proj, env)}/{slug}/scale",
            json={"process_counts": counts},
        )

    # ------------------------------------------------------------------
    # Builds
    # ------------------------------------------------------------------

    def list_builds(self, ws: str, proj: str, env: str, svc: str) -> list[Build]:
        """List builds for a service."""
        resp = self._request("GET", f"{self._svc_base(ws, proj, env)}/{svc}/builds/")
        result = BuildList.model_validate(resp.json())
        return result.items

    def get_build(self, build_id: str) -> Build:
        """Get a single build by ID."""
        resp = self._request("GET", f"/builds/{build_id}")
        return Build.model_validate(resp.json())

    def get_build_log(self, build_id: str) -> DeployLog:
        """Get log output for a build."""
        resp = self._request("GET", f"/builds/{build_id}/log")
        return DeployLog.model_validate(resp.json())

    # ------------------------------------------------------------------
    # Deploys
    # ------------------------------------------------------------------

    def list_deploys(self, ws: str, proj: str, env: str, svc: str) -> list[Deploy]:
        """List deploys for a service."""
        resp = self._request("GET", f"{self._svc_base(ws, proj, env)}/{svc}/deploys/")
        result = DeployList.model_validate(resp.json())
        return result.items

    def get_deploy(self, deploy_id: str) -> Deploy:
        """Get deploy details by ID."""
        resp = self._request("GET", f"/deploys/{deploy_id}/detail")
        return Deploy.model_validate(resp.json())

    # ------------------------------------------------------------------
    # Configuration
    # ------------------------------------------------------------------

    def list_config(self, ws: str, proj: str, env: str, svc: str) -> list[ConfigVar]:
        """List configuration variables for a service."""
        resp = self._request("GET", f"{self._svc_base(ws, proj, env)}/{svc}/config/")
        return [ConfigVar.model_validate(item) for item in resp.json()]

    def set_config(
        self,
        ws: str,
        proj: str,
        env: str,
        svc: str,
        key: str,
        value: str,
        secret: bool = False,
    ) -> ConfigVar:
        """Set (create or update) a configuration variable."""
        resp = self._request(
            "POST",
            f"{self._svc_base(ws, proj, env)}/{svc}/config/",
            json={"name": key, "value": value, "secret": secret},
        )
        return ConfigVar.model_validate(resp.json())

    def delete_config(self, ws: str, proj: str, env: str, svc: str, key: str) -> None:
        """Delete a configuration variable."""
        self._request(
            "DELETE",
            f"{self._svc_base(ws, proj, env)}/{svc}/config/{key}",
        )
