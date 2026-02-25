---
title: "Deployments"
description: API reference for deployments endpoints.
---

<!-- Auto-generated from openapi.json — do not edit manually -->

## Get Deployment

```http
GET /deployments/{deployment_id}/detail
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `deployment_id` | string |  |

**Response:**

```json
{
  "id": "string",
  "application_id": "string",
  "release": {},
  "complete": true,
  "error": true,
  "error_detail": {},
  "deploy_metadata": {},
  "deploy_log": {},
  "job_id": {},
  "created": {},
  "updated": {}
}
```


## Get Deployment Log

```http
GET /deployments/{deployment_id}/log
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `deployment_id` | string |  |

**Response:**

```json
{
  "status": "string",
  "log_text": {},
  "error": true,
  "error_detail": {},
  "deploy_step": {}
}
```

