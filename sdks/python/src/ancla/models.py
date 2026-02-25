"""Pydantic models for Ancla API resources."""

from __future__ import annotations

from pydantic import BaseModel, Field


# ---------------------------------------------------------------------------
# Organizations
# ---------------------------------------------------------------------------

class OrgMember(BaseModel):
    """A member of an organization."""

    username: str
    email: str
    admin: bool = False


class Org(BaseModel):
    """An Ancla organization."""

    id: str = ""
    name: str
    slug: str
    member_count: int = 0
    project_count: int = 0
    application_count: int = 0
    members: list[OrgMember] = Field(default_factory=list)


# ---------------------------------------------------------------------------
# Projects
# ---------------------------------------------------------------------------

class Project(BaseModel):
    """A project within an organization."""

    id: str = ""
    name: str
    slug: str
    organization_slug: str = ""
    organization_name: str = ""
    application_count: int = 0
    created: str = ""
    updated: str = ""


# ---------------------------------------------------------------------------
# Applications
# ---------------------------------------------------------------------------

class App(BaseModel):
    """An application within a project."""

    id: str = ""
    name: str
    slug: str
    platform: str = ""
    github_repository: str = ""
    auto_deploy_branch: str = ""
    process_counts: dict[str, int] = Field(default_factory=dict)


# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------

class ConfigVar(BaseModel):
    """A configuration variable attached to an application."""

    id: str = ""
    name: str
    value: str = ""
    secret: bool = False
    buildtime: bool = False


# ---------------------------------------------------------------------------
# Images
# ---------------------------------------------------------------------------

class Image(BaseModel):
    """A container image built for an application."""

    id: str = ""
    version: int = 0
    built: bool = False
    error: bool = False
    created: str = ""


class ImageList(BaseModel):
    """Paginated wrapper returned by the images list endpoint."""

    items: list[Image] = Field(default_factory=list)


# ---------------------------------------------------------------------------
# Releases
# ---------------------------------------------------------------------------

class Release(BaseModel):
    """A release combining an image and configuration."""

    id: str = ""
    version: int = 0
    platform: str = ""
    built: bool = False
    error: bool = False
    created: str = ""


class ReleaseList(BaseModel):
    """Paginated wrapper returned by the releases list endpoint."""

    items: list[Release] = Field(default_factory=list)


# ---------------------------------------------------------------------------
# Deployments
# ---------------------------------------------------------------------------

class Deployment(BaseModel):
    """A deployment of a release to infrastructure."""

    id: str = ""
    complete: bool = False
    error: bool = False
    error_detail: str = ""
    job_id: str = ""
    created: str = ""
    updated: str = ""


# ---------------------------------------------------------------------------
# Action responses
# ---------------------------------------------------------------------------

class DeployResult(BaseModel):
    """Response from triggering a deploy."""

    image_id: str = ""


class ScaleResult(BaseModel):
    """Response from a scale operation (empty on success)."""


class CreateReleaseResult(BaseModel):
    """Response from creating a release."""

    release_id: str = ""
    version: int = 0


class DeployReleaseResult(BaseModel):
    """Response from deploying a release."""

    deployment_id: str = ""
