---
title: Environments
description: API reference for managing project environments.
---

Environments represent deployment targets within a project (production, staging, development, preview). Each environment has its own config variables, services, and deployment history.

## List environments

```http
GET /orgs/{org}/projects/{project}/envs/
```

### Response

```json
[
  {
    "id": "uuid",
    "name": "Production",
    "slug": "production",
    "env_type": "production",
    "protected": true,
    "auto_deploy_branch": "main",
    "ephemeral": false
  }
]
```

## Create an environment

```http
POST /orgs/{org}/projects/{project}/envs/
```

### Request body

```json
{
  "name": "Preview PR-42",
  "env_type": "preview",
  "auto_deploy_branch": "feature/login",
  "protected": false
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | Display name |
| `env_type` | string | no | One of: `production`, `staging`, `development`, `preview`. Defaults to `development`. |
| `auto_deploy_branch` | string | no | Git branch that triggers auto-deploy |
| `protected` | bool | no | Whether deployments require approval |

## Get an environment

```http
GET /orgs/{org}/projects/{project}/envs/{env}
```

## Update an environment

```http
PATCH /orgs/{org}/projects/{project}/envs/{env}
```

Accepts the same fields as create (all optional).

## Environment configuration

### List config

```http
GET /orgs/{org}/projects/{project}/envs/{env}/config
```

### Set a config variable

```http
POST /orgs/{org}/projects/{project}/envs/{env}/config
```

```json
{
  "name": "DATABASE_URL",
  "value": "postgres://...",
  "secret": true
}
```

### Delete a config variable

```http
DELETE /orgs/{org}/projects/{project}/envs/{env}/config/{config_id}
```

### Bulk set config

```http
POST /orgs/{org}/projects/{project}/envs/{env}/config/bulk
```

```json
[
  {"name": "KEY_1", "value": "val1"},
  {"name": "KEY_2", "value": "val2", "secret": true}
]
```

## Environment deployments

### List deployments

```http
GET /orgs/{org}/projects/{project}/envs/{env}/deployments
```

Returns the most recent 50 deployments.

### Get a deployment

```http
GET /orgs/{org}/projects/{project}/envs/{env}/deployments/{deployment_id}
```

### Get deployment log

```http
GET /orgs/{org}/projects/{project}/envs/{env}/deployments/{deployment_id}/log
```

### Trigger a deploy

```http
POST /orgs/{org}/projects/{project}/envs/{env}/deploy?service={slug}
```

Deploys a specific service within the environment. The `service` query parameter is the service slug.
