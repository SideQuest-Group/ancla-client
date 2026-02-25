---
title: "Organizations"
description: API reference for organizations endpoints.
---

<!-- Auto-generated from openapi.json — do not edit manually -->

## List Organizations

```http
GET /organizations
```

**Response:**

```json
[
  {
    "id": "string",
    "name": "string",
    "slug": "string",
    "member_count": 0,
    "project_count": 0
  }
]
```


## Get Organization

```http
GET /organizations/{org_slug}
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `org_slug` | string |  |

**Response:**

```json
{
  "id": "string",
  "name": "string",
  "slug": "string",
  "members": [
    {
      "user_id": "string",
      "username": "string",
      "email": {},
      "admin": true
    }
  ],
  "project_count": 0,
  "application_count": 0
}
```


## Create Organization

```http
POST /organizations
```

**Request body:**

```json
{
  "name": "string",
  "slug": "string"
}
```

**Response:**

```json
{
  "id": "string",
  "name": "string",
  "slug": "string",
  "members": [
    {
      "user_id": "string",
      "username": "string",
      "email": {},
      "admin": true
    }
  ],
  "project_count": 0,
  "application_count": 0
}
```


## Add Member

```http
POST /organizations/{org_slug}/members
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `org_slug` | string |  |

**Request body:**

```json
{
  "username": "string",
  "admin": true
}
```

**Response:**

```json
{
  "id": "string",
  "name": "string",
  "slug": "string",
  "members": [
    {
      "user_id": "string",
      "username": "string",
      "email": {},
      "admin": true
    }
  ],
  "project_count": 0,
  "application_count": 0
}
```


## Toggle Member Admin

```http
POST /organizations/{org_slug}/members/{user_id}/toggle-admin
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `org_slug` | string |  |
| `user_id` | string |  |

**Response:**

```json
{
  "ok": true,
  "admin": true
}
```


## Remove Member

```http
DELETE /organizations/{org_slug}/members/{user_id}
```

**Path parameters:**

| Name | Type | Description |
|------|------|-------------|
| `org_slug` | string |  |
| `user_id` | string |  |

**Response:**

```json
{
  "ok": true
}
```

