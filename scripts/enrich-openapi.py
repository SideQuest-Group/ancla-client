#!/usr/bin/env python3
"""Enrich the Ancla OpenAPI spec with typed schemas and clean operationIds.

Reads the bare spec (openapi.json) where workspace-domain endpoints have
``{"type": "object"}`` for request/response bodies and verbose operationIds,
then writes an enriched copy with proper ``$ref`` pointers and short names.

Usage::

    python3 scripts/enrich-openapi.py                          # defaults
    python3 scripts/enrich-openapi.py --spec openapi.json --out openapi.enriched.json
"""

from __future__ import annotations

import argparse
import copy
import json
import sys
from pathlib import Path

# ---------------------------------------------------------------------------
# Schema definitions (derived from sdks/go/models.go)
# ---------------------------------------------------------------------------

SCHEMAS: dict[str, dict] = {
    "Workspace": {
        "type": "object",
        "properties": {
            "id": {"type": "string"},
            "name": {"type": "string"},
            "slug": {"type": "string"},
            "member_count": {"type": "integer"},
            "project_count": {"type": "integer"},
            "service_count": {"type": "integer"},
            "members": {
                "type": "array",
                "items": {"$ref": "#/components/schemas/WorkspaceMember"},
            },
        },
        "required": ["id", "name", "slug"],
    },
    "WorkspaceMember": {
        "type": "object",
        "properties": {
            "username": {"type": "string"},
            "email": {"type": "string"},
            "admin": {"type": "boolean"},
        },
        "required": ["username", "email", "admin"],
    },
    "Project": {
        "type": "object",
        "properties": {
            "id": {"type": "string"},
            "name": {"type": "string"},
            "slug": {"type": "string"},
            "workspace_slug": {"type": "string"},
            "workspace_name": {"type": "string"},
            "service_count": {"type": "integer"},
            "created": {"type": "string", "format": "date-time"},
            "updated": {"type": "string", "format": "date-time"},
        },
        "required": ["id", "name", "slug"],
    },
    "Environment": {
        "type": "object",
        "properties": {
            "id": {"type": "string"},
            "name": {"type": "string"},
            "slug": {"type": "string"},
            "service_count": {"type": "integer"},
            "created": {"type": "string", "format": "date-time"},
        },
        "required": ["id", "name", "slug"],
    },
    "Service": {
        "type": "object",
        "properties": {
            "id": {"type": "string"},
            "name": {"type": "string"},
            "slug": {"type": "string"},
            "platform": {"type": "string"},
            "github_repository": {"type": "string"},
            "auto_deploy_branch": {"type": "string"},
            "process_counts": {
                "type": "object",
                "additionalProperties": {"type": "integer"},
            },
        },
        "required": ["id", "name", "slug", "platform"],
    },
    "Build": {
        "type": "object",
        "properties": {
            "id": {"type": "string"},
            "version": {"type": "integer"},
            "built": {"type": "boolean"},
            "error": {"type": "boolean"},
            "created": {"type": "string", "format": "date-time"},
        },
        "required": ["id", "version", "built", "error", "created"],
    },
    "BuildList": {
        "type": "object",
        "properties": {
            "items": {
                "type": "array",
                "items": {"$ref": "#/components/schemas/Build"},
            },
        },
        "required": ["items"],
    },
    "BuildResult": {
        "type": "object",
        "properties": {
            "build_id": {"type": "string"},
            "version": {"type": "integer"},
        },
        "required": ["build_id", "version"],
    },
    "BuildLog": {
        "type": "object",
        "properties": {
            "status": {"type": "string"},
            "version": {"type": "integer"},
            "log_text": {"type": "string"},
        },
        "required": ["status", "version", "log_text"],
    },
    "Deploy": {
        "type": "object",
        "properties": {
            "id": {"type": "string"},
            "complete": {"type": "boolean"},
            "error": {"type": "boolean"},
            "error_detail": {"type": "string"},
            "job_id": {"type": "string"},
            "created": {"type": "string", "format": "date-time"},
            "updated": {"type": "string", "format": "date-time"},
        },
        "required": ["id", "complete", "error", "created"],
    },
    "DeployList": {
        "type": "object",
        "properties": {
            "items": {
                "type": "array",
                "items": {"$ref": "#/components/schemas/Deploy"},
            },
        },
        "required": ["items"],
    },
    "DeployLog": {
        "type": "object",
        "properties": {
            "status": {"type": "string"},
            "log_text": {"type": "string"},
        },
        "required": ["status", "log_text"],
    },
    "ConfigVar": {
        "type": "object",
        "properties": {
            "id": {"type": "string"},
            "name": {"type": "string"},
            "value": {"type": "string"},
            "secret": {"type": "boolean"},
            "buildtime": {"type": "boolean"},
            "scope": {"type": "string"},
        },
        "required": ["id", "name", "value", "secret", "buildtime", "scope"],
    },
    "Team": {
        "type": "object",
        "properties": {
            "id": {"type": "string"},
            "name": {"type": "string"},
            "slug": {"type": "string"},
        },
        "required": ["id", "name", "slug"],
    },
    "PipelineStatus": {
        "type": "object",
        "properties": {
            "build": {"$ref": "#/components/schemas/StageStatus"},
            "deploy": {"$ref": "#/components/schemas/StageStatus"},
        },
    },
    "StageStatus": {
        "type": "object",
        "properties": {
            "status": {"type": "string"},
        },
        "required": ["status"],
    },
    "PipelineHistory": {
        "type": "object",
        "properties": {
            "items": {"type": "array", "items": {"type": "object"}},
        },
    },
    "PipelineMetrics": {
        "type": "object",
        "properties": {
            "metrics": {"type": "object"},
        },
    },
    "ObservabilityData": {
        "type": "object",
        "properties": {
            "metrics": {"type": "object"},
            "logs": {"type": "object"},
        },
    },
    "PromotionPreview": {
        "type": "object",
        "properties": {
            "changes": {"type": "array", "items": {"type": "object"}},
        },
    },
    "PromotionResult": {
        "type": "object",
        "properties": {
            "success": {"type": "boolean"},
            "message": {"type": "string"},
        },
        "required": ["success"],
    },
    # --- Request body schemas ---
    "CreateWorkspaceRequest": {
        "type": "object",
        "properties": {"name": {"type": "string"}},
        "required": ["name"],
    },
    "UpdateWorkspaceRequest": {
        "type": "object",
        "properties": {"name": {"type": "string"}},
        "required": ["name"],
    },
    "CreateProjectRequest": {
        "type": "object",
        "properties": {"name": {"type": "string"}},
        "required": ["name"],
    },
    "UpdateProjectRequest": {
        "type": "object",
        "properties": {"name": {"type": "string"}},
        "required": ["name"],
    },
    "CreateEnvironmentRequest": {
        "type": "object",
        "properties": {"name": {"type": "string"}},
        "required": ["name"],
    },
    "UpdateEnvironmentRequest": {
        "type": "object",
        "properties": {"name": {"type": "string"}},
        "required": ["name"],
    },
    "CreateServiceRequest": {
        "type": "object",
        "properties": {
            "name": {"type": "string"},
            "platform": {"type": "string"},
        },
        "required": ["name", "platform"],
    },
    "UpdateServiceRequest": {
        "type": "object",
        "properties": {
            "name": {"type": "string"},
            "github_repository": {"type": "string"},
            "auto_deploy_branch": {"type": "string"},
        },
    },
    "SetConfigRequest": {
        "type": "object",
        "properties": {
            "name": {"type": "string"},
            "value": {"type": "string"},
            "secret": {"type": "boolean"},
        },
        "required": ["name", "value"],
    },
    "BulkSetConfigRequest": {
        "type": "object",
        "properties": {
            "vars": {
                "type": "array",
                "items": {"$ref": "#/components/schemas/SetConfigRequest"},
            },
        },
        "required": ["vars"],
    },
    "ScaleRequest": {
        "type": "object",
        "properties": {
            "process_counts": {
                "type": "object",
                "additionalProperties": {"type": "integer"},
            },
        },
        "required": ["process_counts"],
    },
    "DeployRequest": {
        "type": "object",
        "properties": {
            "version": {"type": "integer"},
        },
    },
    "AddMemberRequest": {
        "type": "object",
        "properties": {
            "username": {"type": "string"},
            "admin": {"type": "boolean"},
        },
        "required": ["username"],
    },
    "CreateTeamRequest": {
        "type": "object",
        "properties": {"name": {"type": "string"}},
        "required": ["name"],
    },
    "PipelineDeployRequest": {
        "type": "object",
        "properties": {
            "version": {"type": "integer"},
        },
    },
    "PipelineRollbackRequest": {
        "type": "object",
        "properties": {
            "version": {"type": "integer"},
        },
    },
}

# ---------------------------------------------------------------------------
# operationId mapping: old verbose name → clean name
# ---------------------------------------------------------------------------

OPERATION_ID_MAP: dict[str, str] = {
    # Workspaces
    "ApiV1WorkspacesListWorkspaces": "listWorkspaces",
    "ApiV1WorkspacesCreateWorkspace": "createWorkspace",
    "ApiV1WorkspacesGetWorkspace": "getWorkspace",
    # Members
    "ApiV1WorkspacesMembersListMembers": "listMembers",
    "ApiV1WorkspacesMembersAddMember": "addMember",
    # Config (workspace scope)
    "ApiV1WorkspacesConfigListConfig": "listWorkspaceConfig",
    "ApiV1WorkspacesConfigCreateConfig": "createWorkspaceConfig",
    # Projects
    "ApiV1WorkspacesProjectsListProjects": "listProjects",
    "ApiV1WorkspacesProjectsCreateProject": "createProject",
    "ApiV1WorkspacesProjectsGetProject": "getProject",
    "ApiV1WorkspacesProjectsUpdateProject": "updateProject",
    # Config (project scope)
    "ApiV1WorkspacesProjectsConfigListConfig": "listProjectConfig",
    "ApiV1WorkspacesProjectsConfigCreateConfig": "createProjectConfig",
    # Environments
    "ApiV1WorkspacesProjectsEnvsListEnvironments": "listEnvironments",
    "ApiV1WorkspacesProjectsEnvsCreateEnvironment": "createEnvironment",
    "ApiV1WorkspacesProjectsEnvsGetEnvironment": "getEnvironment",
    "ApiV1WorkspacesProjectsEnvsUpdateEnvironment": "updateEnvironment",
    # Config (env scope)
    "ApiV1WorkspacesProjectsEnvsConfigListConfig": "listEnvConfig",
    "ApiV1WorkspacesProjectsEnvsConfigCreateConfig": "createEnvConfig",
    "ApiV1WorkspacesProjectsEnvsConfigConfigIdDeleteConfig": "deleteEnvConfig",
    "ApiV1WorkspacesProjectsEnvsConfigBulkBulkCreateConfig": "bulkCreateEnvConfig",
    # Env deploys
    "ApiV1WorkspacesProjectsEnvsDeployDeployEnv": "deployEnvironment",
    "ApiV1WorkspacesProjectsEnvsDeploysListDeploys": "listEnvDeploys",
    "ApiV1WorkspacesProjectsEnvsDeploysDeployIdGetDeploy": "getEnvDeploy",
    "ApiV1WorkspacesProjectsEnvsDeploysDeployIdLogGetDeployLog": "getEnvDeployLog",
    # Services
    "ApiV1WorkspacesProjectsEnvsServicesListServices": "listServices",
    "ApiV1WorkspacesProjectsEnvsServicesCreateService": "createService",
    "ApiV1WorkspacesProjectsEnvsServicesSvcGetService": "getService",
    "ApiV1WorkspacesProjectsEnvsServicesSvcUpdateService": "updateService",
    "ApiV1WorkspacesProjectsEnvsServicesSvcDeployDeployService": "deployService",
    # Config (service scope)
    "ApiV1WorkspacesProjectsEnvsServicesSvcConfigListConfig": "listServiceConfig",
    "ApiV1WorkspacesProjectsEnvsServicesSvcConfigCreateConfig": "createServiceConfig",
    "ApiV1WorkspacesProjectsEnvsServicesSvcConfigResolvedGetResolvedConfig": "getResolvedConfig",
    # Builds
    "ApiV1WorkspacesProjectsEnvsServicesSvcBuildsListBuilds": "listBuilds",
    "ApiV1WorkspacesProjectsEnvsServicesSvcBuildsVersionGetBuild": "getBuild",
    "ApiV1WorkspacesProjectsEnvsServicesSvcBuildsVersionLogGetBuildLog": "getBuildLog",
    "ApiV1WorkspacesProjectsEnvsServicesSvcBuildsTriggerTriggerBuild": "triggerBuild",
    # Deploys (service scope)
    "ApiV1WorkspacesProjectsEnvsServicesSvcDeploysListDeploys": "listServiceDeploys",
    "ApiV1WorkspacesProjectsEnvsServicesSvcDeploysDeployIdGetDeploy": "getServiceDeploy",
    "ApiV1WorkspacesProjectsEnvsServicesSvcDeploysDeployIdLogGetDeployLog": "getServiceDeployLog",
    # Promotion
    "ApiV1WorkspacesProjectsPromoteExecutePromotion": "executePromotion",
    "ApiV1WorkspacesProjectsPromotePreviewPreviewPromotion": "previewPromotion",
    # Teams
    "ApiV1WorkspacesTeamsListTeams": "listTeams",
    "ApiV1WorkspacesTeamsCreateTeam": "createTeam",
    "ApiV1WorkspacesTeamsConfigListConfig": "listTeamConfig",
    "ApiV1WorkspacesTeamsConfigCreateConfig": "createTeamConfig",
    # Pipeline
    "ApiV1WorkspacesProjectsPipelineDeployPipelineDeploy": "pipelineDeploy",
    "ApiV1WorkspacesProjectsPipelineHistoryPipelineHistory": "pipelineHistory",
    "ApiV1WorkspacesProjectsPipelineMetricsPipelineMetrics": "pipelineMetrics",
    "ApiV1WorkspacesProjectsPipelineRollbackPipelineRollback": "pipelineRollback",
    "ApiV1WorkspacesProjectsPipelineStatusPipelineStatus": "pipelineStatus",
    # Observability
    "ApiV1WorkspacesProjectsObservabilityGetObservability": "getObservability",
    # Auth
    "ApiV1AuthSessionGetSession": "getSession",
    "ApiV1AuthLoginLogin": "login",
    "ApiV1AuthLogoutLogout": "logout",
    # Integrations
    "ApiV1IntegrationsDockerAuthGetDockerAuth": "getDockerAuth",
    "ApiV1IntegrationsDockerAuthPostDockerAuth": "postDockerAuth",
    "ApiV1IntegrationsGithubHooksGithubHooks": "githubHooks",
    "ApiV1IntegrationsHealthHealth": "health",
    "ApiV1IntegrationsSigningCertSigningCert": "getSigningCert",
}

# ---------------------------------------------------------------------------
# Response/request schema mapping: operationId → (response_schema, request_schema)
# Values are either a schema name (string) or a dict for inline array wrapping.
# None means "leave as-is".
# ---------------------------------------------------------------------------

def _ref(name: str) -> dict:
    return {"$ref": f"#/components/schemas/{name}"}


def _array_of(name: str) -> dict:
    return {"type": "array", "items": _ref(name)}


# Maps NEW operationId → (response_schema_override, request_body_override)
# Use string for $ref, dict for inline, None for no change.
SCHEMA_MAP: dict[str, tuple] = {
    # Workspaces
    "listWorkspaces": (_array_of("Workspace"), None),
    "createWorkspace": (_ref("Workspace"), _ref("CreateWorkspaceRequest")),
    "getWorkspace": (_ref("Workspace"), None),
    # Members
    "listMembers": (_array_of("WorkspaceMember"), None),
    "addMember": (_ref("WorkspaceMember"), _ref("AddMemberRequest")),
    # Config (workspace)
    "listWorkspaceConfig": (_array_of("ConfigVar"), None),
    "createWorkspaceConfig": (_ref("ConfigVar"), _ref("SetConfigRequest")),
    # Projects
    "listProjects": (_array_of("Project"), None),
    "createProject": (_ref("Project"), _ref("CreateProjectRequest")),
    "getProject": (_ref("Project"), None),
    "updateProject": (_ref("Project"), _ref("UpdateProjectRequest")),
    # Config (project)
    "listProjectConfig": (_array_of("ConfigVar"), None),
    "createProjectConfig": (_ref("ConfigVar"), _ref("SetConfigRequest")),
    # Environments
    "listEnvironments": (_array_of("Environment"), None),
    "createEnvironment": (_ref("Environment"), _ref("CreateEnvironmentRequest")),
    "getEnvironment": (_ref("Environment"), None),
    "updateEnvironment": (_ref("Environment"), _ref("UpdateEnvironmentRequest")),
    # Config (env)
    "listEnvConfig": (_array_of("ConfigVar"), None),
    "createEnvConfig": (_ref("ConfigVar"), _ref("SetConfigRequest")),
    "deleteEnvConfig": (None, None),
    "bulkCreateEnvConfig": (_array_of("ConfigVar"), _ref("BulkSetConfigRequest")),
    # Env deploys
    "deployEnvironment": (_ref("Deploy"), _ref("DeployRequest")),
    "listEnvDeploys": (_ref("DeployList"), None),
    "getEnvDeploy": (_ref("Deploy"), None),
    "getEnvDeployLog": (_ref("DeployLog"), None),
    # Services
    "listServices": (_array_of("Service"), None),
    "createService": (_ref("Service"), _ref("CreateServiceRequest")),
    "getService": (_ref("Service"), None),
    "updateService": (_ref("Service"), _ref("UpdateServiceRequest")),
    "deployService": (_ref("Deploy"), _ref("DeployRequest")),
    # Config (service)
    "listServiceConfig": (_array_of("ConfigVar"), None),
    "createServiceConfig": (_ref("ConfigVar"), _ref("SetConfigRequest")),
    "getResolvedConfig": (_array_of("ConfigVar"), None),
    # Builds
    "listBuilds": (_ref("BuildList"), None),
    "getBuild": (_ref("Build"), None),
    "getBuildLog": (_ref("BuildLog"), None),
    "triggerBuild": (_ref("BuildResult"), None),
    # Deploys (service)
    "listServiceDeploys": (_ref("DeployList"), None),
    "getServiceDeploy": (_ref("Deploy"), None),
    "getServiceDeployLog": (_ref("DeployLog"), None),
    # Promotion
    "executePromotion": (_ref("PromotionResult"), None),
    "previewPromotion": (_ref("PromotionPreview"), None),
    # Teams
    "listTeams": (_array_of("Team"), None),
    "createTeam": (_ref("Team"), _ref("CreateTeamRequest")),
    "listTeamConfig": (_array_of("ConfigVar"), None),
    "createTeamConfig": (_ref("ConfigVar"), _ref("SetConfigRequest")),
    # Pipeline
    "pipelineDeploy": (_ref("Deploy"), _ref("PipelineDeployRequest")),
    "pipelineHistory": (_ref("PipelineHistory"), None),
    "pipelineMetrics": (_ref("PipelineMetrics"), None),
    "pipelineRollback": (_ref("Deploy"), _ref("PipelineRollbackRequest")),
    "pipelineStatus": (_ref("PipelineStatus"), None),
    # Observability
    "getObservability": (_ref("ObservabilityData"), None),
    # Auth
    "getSession": (None, None),
    "login": (None, None),
    "logout": (None, None),
    # Integrations
    "getDockerAuth": (None, None),
    "postDockerAuth": (None, None),
    "githubHooks": (None, None),
    "health": (None, None),
    "getSigningCert": (None, None),
}


def _is_bare_object(schema: dict) -> bool:
    """Return True if schema is effectively ``{"type": "object"}`` with no properties."""
    if schema.get("type") == "object" and "properties" not in schema and "$ref" not in schema:
        return True
    return False


def _is_bare_array_of_object(schema: dict) -> bool:
    """Return True for ``{"type": "array", "items": {"type": "object"}}``."""
    if schema.get("type") == "array":
        items = schema.get("items", {})
        return _is_bare_object(items)
    return False


def enrich(spec: dict) -> dict:
    """Return a deep-copied spec with enriched schemas and operationIds."""
    spec = copy.deepcopy(spec)

    # 1. Inject schemas into components/schemas
    if "components" not in spec:
        spec["components"] = {}
    if "schemas" not in spec["components"]:
        spec["components"]["schemas"] = {}
    for name, schema_def in SCHEMAS.items():
        spec["components"]["schemas"][name] = schema_def

    # 2. Walk all operations, remap operationId and patch schemas
    for _path, path_item in spec.get("paths", {}).items():
        for _method in ("get", "post", "put", "patch", "delete"):
            op = path_item.get(_method)
            if op is None:
                continue

            old_id = op.get("operationId", "")
            new_id = OPERATION_ID_MAP.get(old_id)
            if new_id:
                op["operationId"] = new_id

                mapping = SCHEMA_MAP.get(new_id)
                if mapping:
                    resp_schema, req_schema = mapping

                    # Patch success response schema
                    if resp_schema is not None:
                        for code, resp in op.get("responses", {}).items():
                            if not code.startswith("2"):
                                continue
                            content = resp.get("content", {})
                            json_content = content.get("application/json", {})
                            existing = json_content.get("schema", {})
                            if _is_bare_object(existing) or _is_bare_array_of_object(existing):
                                json_content["schema"] = resp_schema

                    # Patch request body schema
                    if req_schema is not None:
                        rb = op.get("requestBody", {})
                        content = rb.get("content", {})
                        json_content = content.get("application/json", {})
                        existing = json_content.get("schema", {})
                        if _is_bare_object(existing):
                            json_content["schema"] = req_schema

    return spec


def main() -> None:
    parser = argparse.ArgumentParser(description="Enrich the Ancla OpenAPI spec")
    parser.add_argument(
        "--spec",
        default="openapi.json",
        help="Input OpenAPI spec path (default: openapi.json)",
    )
    parser.add_argument(
        "--out",
        default="openapi.enriched.json",
        help="Output enriched spec path (default: openapi.enriched.json)",
    )
    args = parser.parse_args()

    spec_path = Path(args.spec)
    if not spec_path.exists():
        print(f"Error: {spec_path} not found", file=sys.stderr)
        sys.exit(1)

    with open(spec_path) as f:
        spec = json.load(f)

    enriched = enrich(spec)

    with open(args.out, "w") as f:
        json.dump(enriched, f, indent=4)
        f.write("\n")

    # Summary
    mapped = sum(
        1
        for _path, pi in enriched.get("paths", {}).items()
        for m in ("get", "post", "put", "patch", "delete")
        if (op := pi.get(m)) and op.get("operationId") in SCHEMA_MAP
    )
    print(f"Enriched spec written to {args.out}")
    print(f"  Schemas injected: {len(SCHEMAS)}")
    print(f"  operationIds remapped: {len(OPERATION_ID_MAP)}")
    print(f"  Endpoints with schema overrides: {mapped}")


if __name__ == "__main__":
    main()
