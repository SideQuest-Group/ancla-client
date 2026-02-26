---
title: Teams
description: API reference for managing teams within organizations.
---

Teams are groups of users within an organization. Config variables can be scoped to a team, sitting between org-level and project-level in the configuration hierarchy.

## List teams

```http
GET /orgs/{org}/teams/
```

### Response

```json
[
  {
    "id": "uuid",
    "name": "Backend",
    "slug": "backend"
  }
]
```

## Create a team

```http
POST /orgs/{org}/teams/
```

### Request body

```json
{
  "name": "Backend"
}
```

## Team configuration

### List team config

```http
GET /orgs/{org}/teams/{team}/config
```

### Set a team config variable

```http
POST /orgs/{org}/teams/{team}/config
```

```json
{
  "name": "SHARED_SECRET",
  "value": "team-shared-value",
  "secret": true
}
```

Team-level config variables are inherited by all projects and services that belong to the team, and can be overridden at the project, environment, or service level.
