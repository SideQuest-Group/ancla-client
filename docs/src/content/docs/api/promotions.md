---
title: Promotions
description: API reference for promoting releases between environments.
---

Promotions move a release from one environment to another (e.g., staging to production). The API previews config differences and handles the deployment in the target environment.

## Preview a promotion

```http
GET /orgs/{org}/projects/{project}/promote/preview?service={svc}&source_env={env}&target_env={env}
```

Returns the config diff between source and target so you can review before promoting.

### Response

```json
{
  "source_env": "staging",
  "target_env": "production",
  "config_diff": {
    "added": ["NEW_VAR"],
    "removed": [],
    "changed": ["DATABASE_URL"]
  },
  "source_release_version": 42
}
```

## Execute a promotion

```http
POST /orgs/{org}/projects/{project}/promote
```

### Request body

```json
{
  "service": "api",
  "source_env": "staging",
  "target_env": "production"
}
```

Creates a new release in the target environment based on the source release and enqueues a deployment.

### Response

```json
{
  "id": "uuid",
  "status": "deploying",
  "source_release_id": "uuid",
  "target_release_id": "uuid",
  "target_deployment_id": "uuid"
}
```

### Promotion statuses

| Status | Meaning |
|--------|---------|
| `pending` | Created, not yet started |
| `deploying` | Target deployment in progress |
| `complete` | Successfully deployed in target environment |
| `error` | Deployment failed |
