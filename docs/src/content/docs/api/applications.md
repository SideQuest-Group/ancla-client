---
title: "Applications"
description: API reference for applications endpoints.
---

<!-- Auto-generated from openapi.json — do not edit manually -->

## List Deployments

```http
GET /applications/{application_id}/deployments
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `application_id` | string |  |

**Response:**

```json
{
  "items": [
    {}
  ],
  "pagination": {}
}
```


## Observability

```http
GET /applications/{application_id}/observability
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `application_id` | string |  |

**Response:**

```json
{
  "range": "string",
  "history": [
    {
      "timestamp": {},
      "cpu_usage_m": {},
      "memory_usage_bytes": {},
      "pod_count": {},
      "restart_count": {}
    }
  ]
}
```


## Pipeline Metrics

```http
GET /applications/{application_id}/pipeline-metrics
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `application_id` | string |  |

**Response:**

```json
{
  "build": {},
  "release": {},
  "deploy": {}
}
```


## Pipeline Status

```http
GET /applications/{application_id}/pipeline-status
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `application_id` | string |  |

**Response:**

```json
{
  "build": {},
  "release": {},
  "deploy": {}
}
```


## List Applications

```http
GET /applications/{org_slug}/{project_slug}
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `org_slug` | string |  |
| `project_slug` | string |  |

**Response:**

```json
[
  {
    "id": "string",
    "name": "string",
    "slug": "string",
    "project_id": "string",
    "platform": "string",
    "created": {},
    "latest_image_version": {},
    "latest_release_version": {},
    "latest_deployment_complete": {}
  }
]
```


## Get Application

```http
GET /applications/{org_slug}/{project_slug}/{app_slug}
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `org_slug` | string |  |
| `project_slug` | string |  |
| `app_slug` | string |  |

**Response:**

```json
{
  "id": "string",
  "name": "string",
  "slug": "string",
  "project_id": "string",
  "project_slug": "string",
  "organization_slug": "string",
  "platform": "string",
  "process_counts": {},
  "process_pod_classes": {},
  "github_repository": {},
  "github_repository_is_private": true,
  "auto_deploy_branch": {},
  "subdirectory": {},
  "github_app_installation_id": {},
  "github_environment_name": {},
  "deployment_timeout": {},
  "health_check_path": "string",
  "health_check_host": {},
  "privileged": true,
  "created": {},
  "updated": {},
  "latest_image": {},
  "latest_release": {},
  "latest_deployment": {}
}
```


## Create Application

```http
POST /applications
```

**Request body:**

```json
{
  "name": "string",
  "slug": "string",
  "project_id": "string"
}
```

**Response:**

```json
{
  "id": "string",
  "name": "string",
  "slug": "string",
  "project_id": "string",
  "project_slug": "string",
  "organization_slug": "string",
  "platform": "string",
  "process_counts": {},
  "process_pod_classes": {},
  "github_repository": {},
  "github_repository_is_private": true,
  "auto_deploy_branch": {},
  "subdirectory": {},
  "github_app_installation_id": {},
  "github_environment_name": {},
  "deployment_timeout": {},
  "health_check_path": "string",
  "health_check_host": {},
  "privileged": true,
  "created": {},
  "updated": {},
  "latest_image": {},
  "latest_release": {},
  "latest_deployment": {}
}
```


## Full Deploy

```http
POST /applications/{application_id}/deploy
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `application_id` | string |  |

**Response:**

```json
{
  "ok": true,
  "image_id": "string"
}
```


## Scale Application

```http
POST /applications/{application_id}/scale
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `application_id` | string |  |

**Request body:**

```json
{
  "process_counts": {},
  "process_pod_classes": {}
}
```

**Response:**

```json
{
  "id": "string",
  "name": "string",
  "slug": "string",
  "project_id": "string",
  "project_slug": "string",
  "organization_slug": "string",
  "platform": "string",
  "process_counts": {},
  "process_pod_classes": {},
  "github_repository": {},
  "github_repository_is_private": true,
  "auto_deploy_branch": {},
  "subdirectory": {},
  "github_app_installation_id": {},
  "github_environment_name": {},
  "deployment_timeout": {},
  "health_check_path": "string",
  "health_check_host": {},
  "privileged": true,
  "created": {},
  "updated": {},
  "latest_image": {},
  "latest_release": {},
  "latest_deployment": {}
}
```


## Update Settings

```http
PATCH /applications/{application_id}/settings
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `application_id` | string |  |

**Request body:**

```json
{
  "github_repository": {},
  "github_repository_is_private": {},
  "auto_deploy_branch": {},
  "subdirectory": {},
  "github_app_installation_id": {},
  "github_environment_name": {},
  "deployment_timeout": {},
  "health_check_path": {},
  "health_check_host": {}
}
```

**Response:**

```json
{
  "id": "string",
  "name": "string",
  "slug": "string",
  "project_id": "string",
  "project_slug": "string",
  "organization_slug": "string",
  "platform": "string",
  "process_counts": {},
  "process_pod_classes": {},
  "github_repository": {},
  "github_repository_is_private": true,
  "auto_deploy_branch": {},
  "subdirectory": {},
  "github_app_installation_id": {},
  "github_environment_name": {},
  "deployment_timeout": {},
  "health_check_path": "string",
  "health_check_host": {},
  "privileged": true,
  "created": {},
  "updated": {},
  "latest_image": {},
  "latest_release": {},
  "latest_deployment": {}
}
```

