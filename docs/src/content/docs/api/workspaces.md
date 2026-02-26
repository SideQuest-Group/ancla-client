---
title: "Workspaces"
description: API reference for workspaces endpoints.
---

<!-- Auto-generated from openapi.json — do not edit manually -->

## List Workspaces

```http
GET /workspaces
```

**Response:**

```json
[
  {}
]
```


## Get Workspace

```http
GET /workspaces/{workspace}
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |


## List Config

```http
GET /workspaces/{workspace}/config
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


## List Members

```http
GET /workspaces/{workspace}/members
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


## Create Workspace

```http
POST /workspaces
```


## Create Config

```http
POST /workspaces/{workspace}/config
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |


## Add Member

```http
POST /workspaces/{workspace}/members
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
