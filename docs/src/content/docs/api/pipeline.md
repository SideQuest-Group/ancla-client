---
title: Pipeline
description: API reference for pipeline status, metrics, and deployment operations.
---

The pipeline API provides a unified view of the build-release-deploy lifecycle for a service within a project.

All endpoints accept `service` and `env` as query parameters to scope the pipeline to a specific service and environment.

## Get pipeline status

```http
GET /orgs/{org}/projects/{project}/pipeline/status?service={svc}&env={env}
```

### Response

```json
{
  "build": {"status": "complete"},
  "release": {"status": "complete"},
  "deploy": {"status": "running"}
}
```

Each stage is `null` if no activity has occurred.

## Get pipeline metrics

```http
GET /orgs/{org}/projects/{project}/pipeline/metrics?service={svc}&env={env}
```

Returns success rates and counts for the last 30 days.

## Trigger a deploy

```http
POST /orgs/{org}/projects/{project}/pipeline/deploy?service={svc}&env={env}
```

Runs the full pipeline: build image, create release, deploy.

## Get deployment history

```http
GET /orgs/{org}/projects/{project}/pipeline/history?service={svc}&env={env}&limit=10
```

Returns recent deployments. The `limit` parameter defaults to 10.

## Rollback

```http
POST /orgs/{org}/projects/{project}/pipeline/rollback?service={svc}&env={env}
```

Rolls back to the previous successful deployment. Creates a new deployment using the last known-good release.
