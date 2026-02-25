---
title: "Images"
description: API reference for images endpoints.
---

<!-- Auto-generated from openapi.json — do not edit manually -->

## List Images

```http
GET /images/{application_id}
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
      "repository_name": "string",
      "built": true,
      "error": true,
      "created": {},
      "updated": {}
    }
  ],
  "pagination": {}
}
```


## Get Image

```http
GET /images/{image_id}/detail
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `image_id` | string |  |

**Response:**

```json
{
  "id": "string",
  "version": 0,
  "application_id": "string",
  "repository_name": "string",
  "image_id": {},
  "built": true,
  "error": true,
  "error_detail": {},
  "build_slug": {},
  "build_ref": {},
  "dockerfile": {},
  "procfile": {},
  "processes": {},
  "image_metadata": {},
  "commit_sha": {},
  "created": {},
  "updated": {}
}
```


## Get Image Log

```http
GET /images/{image_id}/log
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `image_id` | string |  |

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


## Trigger Build

```http
POST /images/{application_id}/build
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `application_id` | string |  |

**Response:**

```json
{
  "ok": true,
  "image_id": "string",
  "version": 0
}
```

