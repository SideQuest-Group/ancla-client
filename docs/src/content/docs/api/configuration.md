---
title: "Configuration"
description: API reference for configuration endpoints.
---

<!-- Auto-generated from openapi.json — do not edit manually -->

## List Configurations

```http
GET /configurations/{application_id}
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `application_id` | string |  |

**Response:**

```json
[
  {
    "id": "string",
    "name": "string",
    "value": "string",
    "secret": true,
    "buildtime": true,
    "created": {},
    "updated": {}
  }
]
```


## Get Configuration

```http
GET /configurations/{application_id}/{config_id}
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `application_id` | string |  |
| `config_id` | string |  |

**Response:**

```json
{
  "id": "string",
  "name": "string",
  "value": "string",
  "secret": true,
  "buildtime": true,
  "created": {},
  "updated": {}
}
```


## Create Configuration

```http
POST /configurations/{application_id}
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `application_id` | string |  |

**Request body:**

```json
{
  "name": "string",
  "value": "string",
  "secret": true,
  "buildtime": true
}
```

**Response:**

```json
{
  "id": "string",
  "name": "string",
  "value": "string",
  "secret": true,
  "buildtime": true,
  "created": {},
  "updated": {}
}
```


## Bulk Import

```http
POST /configurations/{application_id}/bulk
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `application_id` | string |  |

**Request body:**

```json
{
  "variables": {},
  "raw": {},
  "secret": true,
  "buildtime": true
}
```

**Response:**

```json
{
  "created": [
    "string"
  ],
  "skipped": [
    "string"
  ],
  "errors": [
    {}
  ]
}
```


## Update Configuration

```http
PATCH /configurations/{application_id}/{config_id}
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `application_id` | string |  |
| `config_id` | string |  |

**Request body:**

```json
{
  "value": "string",
  "secret": {},
  "buildtime": {}
}
```

**Response:**

```json
{
  "id": "string",
  "name": "string",
  "value": "string",
  "secret": true,
  "buildtime": true,
  "created": {},
  "updated": {}
}
```


## Delete Configuration

```http
DELETE /configurations/{application_id}/{config_id}
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `application_id` | string |  |
| `config_id` | string |  |

**Response:**

```json
{
  "ok": true
}
```

