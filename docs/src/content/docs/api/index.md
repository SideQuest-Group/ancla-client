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

| Resource | Description |
|----------|-------------|
| [Authentication](/api/authentication/) | Session, login, logout, CLI auth, OAuth |
| [Organizations](/api/organizations/) | Manage orgs, members, and org-level config |
| [Projects](/api/projects/) | Create and manage projects within orgs |
| [Applications](/api/applications/) | CRUD, scaling, deploy, pipeline status |
| [Environments](/api/environments/) | Deployment targets (production, staging, preview) |
| [Services](/api/services/) | Deployable units within environments |
| [Images](/api/images/) | Build and inspect container images |
| [Releases](/api/releases/) | Create releases and trigger deploys |
| [Deployments](/api/deployments/) | Inspect deployment status and logs |
| [Configuration](/api/configuration/) | Manage environment variables at every scope |
| [Pipeline](/api/pipeline/) | Pipeline status, metrics, deploy, rollback |
| [Promotions](/api/promotions/) | Promote releases between environments |
| [Observability](/api/observability/) | CPU, memory, and pod metrics |
| [Teams](/api/teams/) | Team management and team-scoped config |
| [Integrations](/api/integrations/) | Docker auth, GitHub webhooks, signing certs |

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
