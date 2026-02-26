"""Pydantic models for Ancla API resources."""

from __future__ import annotations

from pydantic import BaseModel, Field

# ---------------------------------------------------------------------------
# Workspaces
# ---------------------------------------------------------------------------


class WorkspaceMember(BaseModel):
    """A member of a workspace."""

    username: str
    email: str
    admin: bool = False


class Workspace(BaseModel):
    """An Ancla workspace (formerly organization)."""

    id: str = ""
    name: str
    slug: str
    member_count: int = 0
    project_count: int = 0
    service_count: int = 0
    members: list[WorkspaceMember] = Field(default_factory=list)


# ---------------------------------------------------------------------------
# Projects
# ---------------------------------------------------------------------------


class Project(BaseModel):
    """A project within a workspace."""

    id: str = ""
    name: str
    slug: str
    workspace_slug: str = ""
    workspace_name: str = ""
    service_count: int = 0
    created: str = ""
    updated: str = ""


# ---------------------------------------------------------------------------
# Environments
# ---------------------------------------------------------------------------


class Environment(BaseModel):
    """An environment within a project (e.g. production, staging)."""

    id: str = ""
    name: str
    slug: str
    service_count: int = 0
    created: str = ""


# ---------------------------------------------------------------------------
# Services
# ---------------------------------------------------------------------------


class Service(BaseModel):
    """A service within an environment (formerly application)."""

    id: str = ""
    name: str
    slug: str
    platform: str = ""
    github_repository: str = ""
    auto_deploy_branch: str = ""
    process_counts: dict[str, int] = Field(default_factory=dict)


# ---------------------------------------------------------------------------
# Builds
# ---------------------------------------------------------------------------


class Build(BaseModel):
    """A container build for a service (formerly image)."""

    id: str = ""
    version: int = 0
    built: bool = False
    error: bool = False
    created: str = ""


class BuildList(BaseModel):
    """Paginated wrapper returned by the builds list endpoint."""

    items: list[Build] = Field(default_factory=list)


# ---------------------------------------------------------------------------
# Deploys
# ---------------------------------------------------------------------------


class Deploy(BaseModel):
    """A deploy combining build and rollout (formerly release + deployment)."""

    id: str = ""
    complete: bool = False
    error: bool = False
    error_detail: str = ""
    job_id: str = ""
    created: str = ""
    updated: str = ""


class DeployLog(BaseModel):
    """Log output for a deploy."""

    status: str = ""
    log_text: str = ""


class DeployList(BaseModel):
    """Paginated wrapper returned by the deploys list endpoint."""

    items: list[Deploy] = Field(default_factory=list)


# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------


class ConfigVar(BaseModel):
    """A configuration variable attached to a service."""

    id: str = ""
    name: str
    value: str = ""
    secret: bool = False
    buildtime: bool = False


# ---------------------------------------------------------------------------
# Pipeline status
# ---------------------------------------------------------------------------


class StageStatus(BaseModel):
    """Status of a single pipeline stage."""

    status: str = ""


class PipelineStatus(BaseModel):
    """Status of the build/deploy pipeline (no release stage)."""

    build: StageStatus | None = None
    deploy: StageStatus | None = None


# ---------------------------------------------------------------------------
# Action responses
# ---------------------------------------------------------------------------


class DeployResult(BaseModel):
    """Response from triggering a deploy."""

    build_id: str = ""


class ScaleResult(BaseModel):
    """Response from a scale operation (empty on success)."""


class BuildResult(BaseModel):
    """Response from creating a build."""

    build_id: str = ""
    version: int = 0
