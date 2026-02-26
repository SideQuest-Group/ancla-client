---
title: Services
description: API reference for managing services within environments.
---

Services are the deployable units within an environment. A service maps to a single container image, with its own scaling, health checks, and deployment lifecycle.

## List services

```http
GET /orgs/{org}/projects/{project}/envs/{env}/services/
```

### Response

```json
[
  {
    "id": "uuid",
    "name": "API",
    "slug": "api",
    "platform": "wind",
    "github_repository": "acme/api",
    "process_counts": {"web": 2, "worker": 1}
  }
]
```

## Create a service

```http
POST /orgs/{org}/projects/{project}/envs/{env}/services/
```

### Request body

```json
{
  "name": "API",
  "platform": "wind",
  "github_repository": "acme/api",
  "health_check_path": "/_health/"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | Service name |
| `platform` | string | yes | One of: `wind`, `steam`, `diesel`, `stirling`, `nuclear`, `electric` |
| `github_repository` | string | no | GitHub repo (owner/name) |
| `health_check_path` | string | no | Path for health checks. Defaults to `/_health/`. |
| `auto_deploy_branch` | string | no | Branch for auto-deploy |
| `deployment_timeout` | int | no | Timeout in seconds. Defaults to 180. |

## Get a service

```http
GET /orgs/{org}/projects/{project}/envs/{env}/services/{svc}
```

## Update a service

```http
PATCH /orgs/{org}/projects/{project}/envs/{env}/services/{svc}
```

Accepts the same fields as create (all optional).

## Service configuration

### List config

```http
GET /orgs/{org}/projects/{project}/envs/{env}/services/{svc}/config
```

### Get resolved config

Returns the fully merged configuration, combining org, project, environment, and service-level variables:

```http
GET /orgs/{org}/projects/{project}/envs/{env}/services/{svc}/config/resolved
```

### Set a config variable

```http
POST /orgs/{org}/projects/{project}/envs/{env}/services/{svc}/config
```

## Images

### List images

```http
GET /orgs/{org}/projects/{project}/envs/{env}/services/{svc}/images
```

Supports pagination via `page` and `per_page` query parameters.

### Get an image by version

```http
GET /orgs/{org}/projects/{project}/envs/{env}/services/{svc}/images/{version}
```

The `{version}` parameter is the image version number (integer), not a UUID.

### Get build log

```http
GET /orgs/{org}/projects/{project}/envs/{env}/services/{svc}/images/{version}/log
```

### Trigger a build

```http
POST /orgs/{org}/projects/{project}/envs/{env}/services/{svc}/images/build
```

Enqueues a new image build. Returns the image ID and version.

## Releases

### List releases

```http
GET /orgs/{org}/projects/{project}/envs/{env}/services/{svc}/releases
```

Returns the 50 most recent releases.

### Get a release by version

```http
GET /orgs/{org}/projects/{project}/envs/{env}/services/{svc}/releases/{version}
```

### Get release build log

```http
GET /orgs/{org}/projects/{project}/envs/{env}/services/{svc}/releases/{version}/log
```

## Deploy a service

```http
POST /orgs/{org}/projects/{project}/envs/{env}/services/{svc}/deploy
```

Triggers a full deploy: creates a new release from the latest image and deploys it. Returns the release and deployment IDs.
