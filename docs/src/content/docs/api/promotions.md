---
title: "Promotions"
description: API reference for promotions endpoints.
---

<!-- Auto-generated from openapi.json — do not edit manually -->

## Preview Promotion

```http
GET /workspaces/{workspace}/projects/{project}/promote/preview
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
| `source_env` | string | Yes |  |
| `target_env` | string | Yes |  |


## Execute Promotion

```http
POST /workspaces/{workspace}/projects/{project}/promote
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `project` | string |  |
