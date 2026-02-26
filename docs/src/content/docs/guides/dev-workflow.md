---
title: Development Workflow
description: Day-to-day commands for deploying, monitoring, and debugging your services.
---

Once you've [linked your directory](/guides/project-linking/) to a service, these commands make up the core development loop.

## Deploy

Trigger a full build-deploy pipeline:

```bash
ancla services deploy my-ws/my-project/production/my-service
```

Add `--follow` to stream build and deploy logs in real time instead of returning immediately:

```bash
ancla services deploy my-ws/my-project/production/my-service --follow
```

## Check status

```bash
ancla status
```

Shows the linked service's current pipeline state:

```
Workspace: my-ws
Project:   my-project
Env:       production
Service:   my-service

STAGE    STATUS
Build    complete
Deploy   running
```

## View logs

Pull the latest deploy's logs:

```bash
ancla logs
```

Stream logs as they come in:

```bash
ancla logs -f
```

For a specific deploy (not the latest), use the lower-level command:

```bash
ancla deploys log <deploy-id>
ancla deploys log <deploy-id> --follow
```

## Run local commands with service config

`ancla run` fetches your service's config variables from the API and injects them as environment variables into a local command. This is how you run your service locally with production (or staging) config without copying `.env` files around.

```bash
ancla run -- python manage.py migrate
ancla run -- npm start
ancla run -- env | grep DATABASE
```

The `--` separator is required. Everything after it becomes the command and its arguments.

Non-secret config variables are injected. Secret values are skipped for safety. Your local environment variables are preserved; service config overlays on top of them.

## Scale processes

```bash
ancla services scale my-ws/my-project/production/my-service web=2 worker=1
```

## Take a service down

Scale all processes to zero in one shot:

```bash
ancla down
```

This prompts for confirmation. Skip the prompt in CI with `--yes`:

```bash
ancla down --yes
```

You can also target a specific service without a link:

```bash
ancla down my-ws/my-project/production/my-service --yes
```

## Open in browser

Open the Ancla dashboard for your linked service:

```bash
ancla open
```

Open the documentation site:

```bash
ancla docs
```

## List everything

Quick overview of all your projects grouped by workspace:

```bash
ancla list
```

Or drill into specific resources:

```bash
ancla workspaces list
ancla projects list --workspace my-ws
ancla envs list my-ws/my-project
ancla services list my-ws/my-project/production
```
