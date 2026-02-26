---
title: "Observability"
description: API reference for observability endpoints.
---

<!-- Auto-generated from openapi.json — do not edit manually -->

## Get Observability

```http
GET /workspaces/{workspace}/projects/{project}/observability
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `workspace` | string |  |
| `project` | string |  |

**Query parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `service` | string | No |  |
| `env` | string | No |  |
| `range` | string | No |  |
