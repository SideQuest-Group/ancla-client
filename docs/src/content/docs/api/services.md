---
title: "Services"
description: API reference for services endpoints.
---

<!-- Auto-generated from openapi.json — do not edit manually -->

## List Services

```http
GET /workspaces/{workspace}/projects/{project}/envs/{env}/services
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `project` | string |  |
| `env` | string |  |

**Response:**

```json
[
  {}
]
```


## Get Service

```http
GET /workspaces/{workspace}/projects/{project}/envs/{env}/services/{svc}
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `project` | string |  |
| `env` | string |  |
| `svc` | string |  |


## List Builds

```http
GET /workspaces/{workspace}/projects/{project}/envs/{env}/services/{svc}/builds
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `project` | string |  |
| `env` | string |  |
| `svc` | string |  |

**Query parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `page` | integer | No |  |
| `per_page` | integer | No |  |


## Get Build

```http
GET /workspaces/{workspace}/projects/{project}/envs/{env}/services/{svc}/builds/{version}
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `project` | string |  |
| `env` | string |  |
| `svc` | string |  |
| `version` | integer |  |


## Get Build Log

```http
GET /workspaces/{workspace}/projects/{project}/envs/{env}/services/{svc}/builds/{version}/log
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `project` | string |  |
| `env` | string |  |
| `svc` | string |  |
| `version` | integer |  |


## List Config

```http
GET /workspaces/{workspace}/projects/{project}/envs/{env}/services/{svc}/config
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `project` | string |  |
| `env` | string |  |
| `svc` | string |  |

**Response:**

```json
[
  {}
]
```


## Get Resolved Config

```http
GET /workspaces/{workspace}/projects/{project}/envs/{env}/services/{svc}/config/resolved
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `project` | string |  |
| `env` | string |  |
| `svc` | string |  |


## List Deploys

```http
GET /workspaces/{workspace}/projects/{project}/envs/{env}/services/{svc}/deploys
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `project` | string |  |
| `env` | string |  |
| `svc` | string |  |

**Response:**

```json
[
  {}
]
```


## Get Deploy

```http
GET /workspaces/{workspace}/projects/{project}/envs/{env}/services/{svc}/deploys/{deploy_id}
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `project` | string |  |
| `env` | string |  |
| `svc` | string |  |
| `deploy_id` | string |  |


## Get Deploy Log

```http
GET /workspaces/{workspace}/projects/{project}/envs/{env}/services/{svc}/deploys/{deploy_id}/log
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `project` | string |  |
| `env` | string |  |
| `svc` | string |  |
| `deploy_id` | string |  |


## Create Service

```http
POST /workspaces/{workspace}/projects/{project}/envs/{env}/services
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `project` | string |  |
| `env` | string |  |


## Trigger Build

```http
POST /workspaces/{workspace}/projects/{project}/envs/{env}/services/{svc}/builds/trigger
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `project` | string |  |
| `env` | string |  |
| `svc` | string |  |


## Create Config

```http
POST /workspaces/{workspace}/projects/{project}/envs/{env}/services/{svc}/config
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `project` | string |  |
| `env` | string |  |
| `svc` | string |  |


## Deploy Service

```http
POST /workspaces/{workspace}/projects/{project}/envs/{env}/services/{svc}/deploy
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `project` | string |  |
| `env` | string |  |
| `svc` | string |  |


## Update Service

```http
PATCH /workspaces/{workspace}/projects/{project}/envs/{env}/services/{svc}
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `project` | string |  |
| `env` | string |  |
| `svc` | string |  |
