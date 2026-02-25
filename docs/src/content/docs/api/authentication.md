---
title: "Authentication"
description: API reference for authentication endpoints.
---

<!-- Auto-generated from openapi.json — do not edit manually -->

## Get Session

```http
GET /auth/session
```

**Response:**

```json
{
  "authenticated": true,
  "user": {}
}
```


## Login

```http
POST /auth/login
```

**Request body:**

```json
{
  "email": "string",
  "password": "string"
}
```

**Response:**

```json
{
  "authenticated": true,
  "user": {}
}
```


## Logout

```http
POST /auth/logout
```

**Response:**

```json
{
  "ok": true
}
```

