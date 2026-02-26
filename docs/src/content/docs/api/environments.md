---
title: "Environments"
description: API reference for environments endpoints.
---

<!-- Auto-generated from openapi.json — do not edit manually -->

## List Environments

```http
GET /workspaces/{workspace}/projects/{project}/envs
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `project` | string |  |

**Response:**

```json
[
  {}
]
```


## Get Environment

```http
GET /workspaces/{workspace}/projects/{project}/envs/{env}
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `project` | string |  |
| `env` | string |  |


## List Config

```http
GET /workspaces/{workspace}/projects/{project}/envs/{env}/config
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


## List Deploys

```http
GET /workspaces/{workspace}/projects/{project}/envs/{env}/deploys
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


## Get Deploy

```http
GET /workspaces/{workspace}/projects/{project}/envs/{env}/deploys/{deploy_id}
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `project` | string |  |
| `env` | string |  |
| `deploy_id` | string |  |


## Get Deploy Log

```http
GET /workspaces/{workspace}/projects/{project}/envs/{env}/deploys/{deploy_id}/log
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `project` | string |  |
| `env` | string |  |
| `deploy_id` | string |  |


## Create Environment

```http
POST /workspaces/{workspace}/projects/{project}/envs
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `project` | string |  |


## Create Config

```http
POST /workspaces/{workspace}/projects/{project}/envs/{env}/config
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `project` | string |  |
| `env` | string |  |


## Bulk Create Config

```http
POST /workspaces/{workspace}/projects/{project}/envs/{env}/config/bulk
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `project` | string |  |
| `env` | string |  |

**Request body:**

```json
[
  {}
]
```

**Response:**

```json
[
  {}
]
```


## Deploy Env

```http
POST /workspaces/{workspace}/projects/{project}/envs/{env}/deploy
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `project` | string |  |
| `env` | string |  |

**Query parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `service` | string | No |  |


## Update Environment

```http
PATCH /workspaces/{workspace}/projects/{project}/envs/{env}
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `project` | string |  |
| `env` | string |  |


## Delete Config

```http
DELETE /workspaces/{workspace}/projects/{project}/envs/{env}/config/{config_id}
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `project` | string |  |
| `env` | string |  |
| `config_id` | string |  |
