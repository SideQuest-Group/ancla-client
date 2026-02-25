---
title: "Projects"
description: API reference for projects endpoints.
---

<!-- Auto-generated from openapi.json — do not edit manually -->

## List Projects

```http
GET /projects
```

**Response:**

```json
[
  {
    "id": "string",
    "name": "string",
    "slug": "string",
    "organization_id": "string",
    "organization_slug": "string",
    "application_count": 0,
    "created": {}
  }
]
```


## Get Project

```http
GET /projects/{org_slug}/{project_slug}
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `org_slug` | string |  |
| `project_slug` | string |  |

**Response:**

```json
{
  "id": "string",
  "name": "string",
  "slug": "string",
  "organization_id": "string",
  "organization_slug": "string",
  "organization_name": "string",
  "application_count": 0,
  "created": {},
  "updated": {}
}
```


## Create Project

```http
POST /projects
```

**Request body:**

```json
{
  "name": "string",
  "slug": "string",
  "organization_id": "string"
}
```

**Response:**

```json
{
  "id": "string",
  "name": "string",
  "slug": "string",
  "organization_id": "string",
  "organization_slug": "string",
  "organization_name": "string",
  "application_count": 0,
  "created": {},
  "updated": {}
}
```

