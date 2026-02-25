---
title: "Releases"
description: API reference for releases endpoints.
---

<!-- Auto-generated from openapi.json — do not edit manually -->

## List Releases

```http
GET /releases/{application_id}
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `application_id` | string |  |

**Response:**

```json
{
  "items": [
    {
      "id": "string",
      "version": 0,
      "application_id": "string",
      "platform": "string",
      "built": true,
      "error": true,
      "created": {},
      "updated": {}
    }
  ],
  "pagination": {}
}
```


## Get Release

```http
GET /releases/{release_id}/detail
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `release_id` | string |  |

**Response:**

```json
{
  "id": "string",
  "version": 0,
  "application_id": "string",
  "platform": "string",
  "repository_name": "string",
  "release_id": {},
  "image": {},
  "configuration": {},
  "image_changes": {},
  "configuration_changes": {},
  "built": true,
  "error": true,
  "error_detail": {},
  "dockerfile": {},
  "release_metadata": {},
  "commit_sha": {},
  "health_check_path": "string",
  "health_check_host": {},
  "created": {},
  "updated": {}
}
```


## Get Release Log

```http
GET /releases/{release_id}/log
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `release_id` | string |  |

**Response:**

```json
{
  "status": "string",
  "version": 0,
  "log_text": {},
  "error": true,
  "error_detail": {},
  "build_step": {}
}
```


## Create Release

```http
POST /releases/{application_id}/create
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `application_id` | string |  |

**Response:**

```json
{
  "ok": true,
  "release_id": "string",
  "version": 0
}
```


## Deploy Release

```http
POST /releases/{release_id}/deploy
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `release_id` | string |  |

**Response:**

```json
{
  "ok": true,
  "deployment_id": "string"
}
```

