---
title: "Pipeline"
description: API reference for pipeline endpoints.
---

<!-- Auto-generated from openapi.json — do not edit manually -->

## Pipeline History

```http
GET /workspaces/{workspace}/projects/{project}/pipeline/history
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `project` | string |  |

**Query parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `service` | string | Yes |  |
| `env` | string | Yes |  |
| `limit` | integer | No |  |

**Response:**

```json
[
  {}
]
```


## Pipeline Metrics

```http
GET /workspaces/{workspace}/projects/{project}/pipeline/metrics
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `project` | string |  |

**Query parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `service` | string | Yes |  |
| `env` | string | Yes |  |


## Pipeline Status

```http
GET /workspaces/{workspace}/projects/{project}/pipeline/status
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `project` | string |  |

**Query parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `service` | string | Yes |  |
| `env` | string | Yes |  |


## Pipeline Deploy

```http
POST /workspaces/{workspace}/projects/{project}/pipeline/deploy
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `project` | string |  |

**Query parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `service` | string | Yes |  |
| `env` | string | Yes |  |


## Pipeline Rollback

```http
POST /workspaces/{workspace}/projects/{project}/pipeline/rollback
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `project` | string |  |

**Query parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `service` | string | Yes |  |
| `env` | string | Yes |  |
