---
title: Development Workflow
description: Day-to-day commands for deploying, monitoring, and debugging your apps.
---

Once you've [linked your directory](/guides/project-linking/) to an app, these commands make up the core development loop.

## Deploy

Trigger a full build-release-deploy pipeline:

```bash
ancla apps deploy my-org/my-project/my-app
```

Add `--follow` to stream build and deploy logs in real time instead of returning immediately:

```bash
ancla apps deploy my-org/my-project/my-app --follow
```

## Check status

```bash
ancla status
```

Shows the linked app's current pipeline state:

```
Org:     my-org
Project: my-project
App:     my-app

STAGE    STATUS
Build    complete
Release  complete
Deploy   running
```

## View logs

Pull the latest deployment's logs:

```bash
ancla logs
```

Stream logs as they come in:

```bash
ancla logs -f
```

For a specific deployment (not the latest), use the lower-level command:

```bash
ancla deployments log <deployment-id>
ancla deployments log <deployment-id> --follow
```

## Run local commands with app config

`ancla run` fetches your app's config variables from the API and injects them as environment variables into a local command. This is how you run your app locally with production (or staging) config without copying `.env` files around.

```bash
ancla run -- python manage.py migrate
ancla run -- npm start
ancla run -- env | grep DATABASE
```

The `--` separator is required. Everything after it becomes the command and its arguments.

Non-secret config variables are injected. Secret values are skipped for safety. Your local environment variables are preserved; app config overlays on top of them.

## Scale processes

```bash
ancla apps scale my-org/my-project/my-app web=2 worker=1
```

## Take an app down

Scale all processes to zero in one shot:

```bash
ancla down
```

This prompts for confirmation. Skip the prompt in CI with `--yes`:

```bash
ancla down --yes
```

You can also target a specific app without a link:

```bash
ancla down my-org/my-project/my-app --yes
```

## Open in browser

Open the Ancla dashboard for your linked app:

```bash
ancla open
```

Open the documentation site:

```bash
ancla docs
```

## List everything

Quick overview of all your projects grouped by org:

```bash
ancla list
```

Or drill into specific resources:

```bash
ancla orgs list
ancla projects list --org my-org
ancla apps list my-org/my-project
```
