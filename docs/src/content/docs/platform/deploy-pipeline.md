---
title: Deploy Pipeline
description: What happens between "ancla deploy" and your app being live.
---

## The short version

```
Source code → Build (image) → Release (image + config) → Rollout (K8s) → Health check → Active
```

Each step is visible. Builds have logs. Deploys have status. If something fails, the pipeline stops and the previous deploy stays active.

## Triggering a deploy

Three ways to kick off a deploy:

**CLI:**
```bash
ancla services deploy my-ws/my-project/production/api
```

**GitHub auto-deploy:** Push to a branch configured as the auto-deploy branch for a service. The GitHub App webhook triggers the pipeline automatically.

**API:**
```http
POST /workspaces/{ws}/projects/{project}/pipeline/deploy?service=api&env=production
```

All three start the same pipeline.

## Step 1: Build

The platform clones your repository (via GitHub App credentials or from the build ref you specified) and builds a container image using BuildKit.

What happens during the build:

- Ancla reads your `Procfile` to discover process types (`web`, `worker`, etc.)
- Build-time config variables (marked as `buildtime: true`) are injected as Docker build args
- The image is tagged with an auto-incrementing version number (`v1`, `v2`, `v3`, ...)
- The finished image is pushed to the private registry
- The build log is captured and stored

If the build fails, the pipeline stops. No deploy happens. The error and log are saved on the build record.

You can also skip the build step entirely by pointing a service at a pre-built image in an external registry.

## Step 2: Release

A release is a frozen snapshot: the build's image tag + the service's current config variables at that moment in time. Once created, a release is immutable.

This is why you can roll back reliably. Release v12 will always produce the same container with the same environment, regardless of what you've changed in config since then.

The release step also generates envconsul configuration files — one per process type — that tell each container how to pull its secrets from Vault at runtime.

## Step 3: Rollout

The platform applies the release to Kubernetes:

1. **Namespace** — created if it doesn't exist (one per workspace)
2. **Image pull secret** — a short-lived JWT scoped to pull this specific image from the registry
3. **Deployment** — a K8s Deployment per process type (`web`, `worker`, etc.), each with the replica count you've configured
4. **Service & Ingress** — for the `web` process type, the platform creates a K8s Service and configures TLS via Let's Encrypt
5. **Consul registration** — the service is registered in Consul for internal discovery

Rollouts use Kubernetes' rolling update strategy. New pods start before old pods stop, so there's no downtime window.

## Step 4: Health check

The platform watches the new pods. Each service has a configurable health check path (default: `/_health/`) and a deployment timeout (default: 180 seconds).

If the new pods pass health checks within the timeout, the deploy is marked `active`.

If they don't, the deploy is marked `error` and the previous deploy's pods stay running. The failed pods are cleaned up.

## Rollback

```bash
ancla services deploy my-ws/my-project/production/api --version 11
```

Or through the API:

```http
POST /workspaces/{ws}/projects/{project}/pipeline/rollback?service=api&env=production
```

A rollback re-deploys a previous release. Same image, same config snapshot, same behavior as when it was first deployed.

## Pipeline status

Check where a deploy is in the pipeline:

```bash
ancla status
```

```
Workspace: my-ws
Project:   my-project
Env:       production
Service:   api

STAGE    STATUS
Build    complete
Deploy   running
```

The API equivalent:

```http
GET /workspaces/{ws}/projects/{project}/pipeline/status?service=api&env=production
```

## Pipeline metrics

The platform tracks build duration, deploy duration, and success/failure rates per service. View them in the dashboard or pull them from the observability API:

```http
GET /workspaces/{ws}/projects/{project}/pipeline/metrics?service=api&env=production
```

## Auto-deploy branches

Each service can have an `auto_deploy_branch` — a Git branch name. When the GitHub App receives a push event for that branch, it triggers the full pipeline for that service.

Set it in the dashboard or via the API when creating/updating a service. Common patterns:

- `main` → production
- `develop` → staging
- PR branches → ephemeral preview environments (handled automatically when configured)

## Process types and the Procfile

Ancla reads a `Procfile` at the root of your repo to determine what processes to run:

```
web: gunicorn app:application --bind 0.0.0.0:$PORT
worker: celery -A tasks worker
beat: celery -A tasks beat
```

Each line becomes a separate Kubernetes Deployment. Scale them independently:

```bash
ancla services scale my-ws/my-project/production/api web=3 worker=2
```

If there's no Procfile, Ancla assumes a single `web` process using the Dockerfile's `CMD`.
