---
title: Secrets & Config
description: How config variables work — scopes, inheritance, secrets, and build-time injection.
---

Config variables are key/value pairs injected into your containers as environment variables. They're the primary way to configure services on Ancla.

## Setting config

```bash
# Set a plaintext variable on a service
ancla config set my-ws/my-project/production/api DATABASE_URL=postgresql://...

# Set a secret (encrypted in Vault, never exposed in logs or API responses)
ancla config set my-ws/my-project/production/api SECRET_KEY=supersecret --secret

# Import from a .env file
ancla config import my-ws/my-project/production/api -f .env
```

## Scopes

Config can be set at five levels:

| Scope | What it affects | Flag |
|-------|----------------|------|
| Workspace | Every service in the workspace | `--scope workspace` |
| Team | Services owned by that team | `--scope team` |
| Project | Every service in the project | `--scope project` |
| Environment | Every service in the environment | `--scope env` |
| Service | That service only | `--scope service` (default) |

```bash
# Share a monitoring key across the whole workspace
ancla config set my-ws DATADOG_API_KEY=abc123 --scope workspace

# Override the database URL for production only
ancla config set my-ws/my-project/production DATABASE_URL=postgresql://prod-db/... --scope env
```

Lower scopes override higher scopes. If `DATABASE_URL` is set at both the project and service level, the service-level value wins for that service.

## Resolved config

To see the final merged config for a service (all scopes resolved):

```bash
ancla config list my-ws/my-project/production/api
```

Or through the API:

```http
GET /workspaces/{ws}/projects/{project}/envs/{env}/services/{svc}/config/resolved
```

This shows exactly what the container will receive. Secret values show as `**secret**` — they're only decrypted at runtime inside the container.

## How secrets work

Secrets are stored in HashiCorp Vault, not in the database. When you set a config variable with `--secret`:

1. The value is written to Vault at a path scoped to the service
2. The database stores the key name and a reference to the Vault path, but not the value
3. At deploy time, Ancla generates an [envconsul](https://github.com/hashicorp/envconsul) sidecar config that maps the Vault path to the environment variable name
4. When the container starts, envconsul authenticates with Vault and injects the secret as an environment variable

Secrets never appear in build logs, API responses, deploy records, or the database. The only place a secret value exists in plaintext is inside the running container's process environment.

## Build-time variables

Some variables need to be available during the Docker build — things like private package registry tokens or build flags. Mark them as build-time:

```bash
ancla config set my-ws/my-project/production/api PIP_INDEX_URL=https://private.pypi.org/simple --buildtime
```

Build-time variables are injected as Docker `ARG`s. They appear in the build log (except secrets, which are injected via BuildKit secret mounts).

A variable can be both `--secret` and `--buildtime`. In that case, it's injected as a BuildKit secret during builds and as a Vault-backed env var at runtime.

## Config and deploys

Config changes don't take effect immediately. They're captured in the next deploy.

When you trigger a deploy, the platform snapshots all resolved config for the service at that moment and freezes it into the release record. Changing config after the deploy doesn't affect running containers — they have the config from when the deploy was created.

This means:

- You can update config variables, review the changes, and then deploy when ready
- Rolling back a deploy restores the config from that release, not the current config
- Two deploys of the same build with different config produce different releases

## Running locally with service config

The CLI can inject a service's config into a local command:

```bash
ancla run -- python manage.py migrate
```

This pulls the resolved config from the API and sets it as environment variables for the subprocess. Secret values are skipped for safety. Your local environment variables are preserved; service config overlays on top.
