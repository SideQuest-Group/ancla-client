---
title: API Reference
description: REST API reference for the Ancla deployment platform.
---

The Ancla API is a REST API at `https://ancla.dev/api/v1`. All endpoints return JSON and require authentication via API key or session.

## Base URL

```
https://ancla.dev/api/v1
```

## Authentication

Include your API key in the `X-API-Key` header:

```bash
curl -H "X-API-Key: ancla_your_key_here" https://ancla.dev/api/v1/auth/session
```

Or use session-based auth via the login endpoint.

## Resources

| Resource | Endpoints | Description |
|----------|-----------|-------------|
| [Authentication](/api/authentication/) | 6 | Session, login, logout, CLI auth, OAuth |
| [Organizations](/api/organizations/) | 6 | Manage orgs and members |
| [Projects](/api/projects/) | 3 | Create and list projects |
| [Applications](/api/applications/) | 10 | CRUD, scaling, deploy, pipeline status |
| [Images](/api/images/) | 4 | Build and inspect container images |
| [Releases](/api/releases/) | 5 | Create releases and trigger deploys |
| [Deployments](/api/deployments/) | 2 | Inspect deployment status and logs |
| [Configuration](/api/configuration/) | 6 | Manage environment variables |

## Error responses

All errors return a JSON body:

```json
{
  "status": 404,
  "message": "not found"
}
```

| Status | Meaning |
|--------|---------|
| 401 | Not authenticated — run `ancla login` or provide an API key |
| 403 | Permission denied — you don't have access to this resource |
| 404 | Resource not found |
| 422 | Validation error — check the request body |
| 500 | Server error |
