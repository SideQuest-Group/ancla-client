---
title: "Teams"
description: API reference for teams endpoints.
---

<!-- Auto-generated from openapi.json — do not edit manually -->

## List Teams

```http
GET /workspaces/{workspace}/teams
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |

**Response:**

```json
[
  {}
]
```


## List Config

```http
GET /workspaces/{workspace}/teams/{team}/config
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `team` | string |  |

**Response:**

```json
[
  {}
]
```


## Create Team

```http
POST /workspaces/{workspace}/teams
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |


## Create Config

```http
POST /workspaces/{workspace}/teams/{team}/config
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `team` | string |  |
