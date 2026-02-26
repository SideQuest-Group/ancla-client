---
title: Project Linking
description: Link your working directory to an Ancla service so commands know what to target.
---

Most workflow commands (`status`, `logs`, `run`, `down`, `ssh`, `shell`, `dbshell`) need to know which service you're working on. Instead of passing `workspace/project/env/service` to every command, link your directory once and forget about it.

## Quick link

If you already know your workspace, project, environment, and service slugs:

```bash
ancla link my-ws/my-project/production/my-service
```

This writes a `.ancla/config.yaml` in your current directory. Every `ancla` command run from this directory (or any subdirectory) will pick it up.

You can also link at a higher level:

```bash
ancla link my-ws                                  # workspace only
ancla link my-ws/my-project                       # workspace + project
ancla link my-ws/my-project/production            # workspace + project + env
```

## Interactive setup with `init`

If you'd rather pick from a menu:

```bash
ancla init
```

This fetches your workspaces, projects, environments, and services from the API and walks you through selecting each one. The result is the same `.ancla/config.yaml` file.

If the directory is already linked, `init` asks before overwriting.

## What gets stored

The link creates `.ancla/config.yaml` in your working directory:

```yaml
workspace: my-ws
project: my-project
env: production
service: my-service
```

This is separate from your global config at `~/.ancla/config.yaml` (which holds your API key and server URL). The local file only stores the link context.

## Checking the current link

```bash
ancla status
```

Shows the linked workspace, project, environment, and service along with the current pipeline status (build/deploy).

```
Workspace: my-ws
Project:   my-project
Env:       production
Service:   my-service

STAGE    STATUS
Build    complete
Deploy   running
```

## Unlinking

```bash
ancla unlink
```

Removes the local `.ancla/config.yaml`. Commands that depend on a link will prompt you to re-link.

## How link resolution works

When you run a command that needs a service context, the CLI looks for `.ancla/config.yaml` starting from your current directory and walking up toward the filesystem root. The first one it finds wins.

This means you can have a monorepo with per-service links:

```
my-monorepo/
  .ancla/config.yaml          # workspace + project level
  services/
    api/
      .ancla/config.yaml      # linked to the api service in production
    frontend/
      .ancla/config.yaml      # linked to the frontend service in production
```

## Link vs. explicit arguments

Every command that uses the link context also accepts an explicit argument. The argument always wins:

```bash
ancla ssh                                                  # uses linked service
ancla ssh other-ws/other-proj/staging/other-svc            # ignores link, uses argument
```
