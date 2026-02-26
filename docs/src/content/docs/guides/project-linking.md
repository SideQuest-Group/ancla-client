---
title: Project Linking
description: Link your working directory to an Ancla app so commands know what to target.
---

Most workflow commands (`status`, `logs`, `run`, `down`, `ssh`, `shell`, `dbshell`) need to know which application you're working on. Instead of passing `org/project/app` to every command, link your directory once and forget about it.

## Quick link

If you already know your org, project, and app slugs:

```bash
ancla link my-org/my-project/my-app
```

This writes a `.ancla/config.yaml` in your current directory. Every `ancla` command run from this directory (or any subdirectory) will pick it up.

You can also link at a higher level:

```bash
ancla link my-org                       # org only
ancla link my-org/my-project            # org + project
```

## Interactive setup with `init`

If you'd rather pick from a menu:

```bash
ancla init
```

This fetches your orgs, projects, and apps from the API and walks you through selecting each one. The result is the same `.ancla/config.yaml` file.

If the directory is already linked, `init` asks before overwriting.

## What gets stored

The link creates `.ancla/config.yaml` in your working directory:

```yaml
org: my-org
project: my-project
app: my-app
```

This is separate from your global config at `~/.ancla/config.yaml` (which holds your API key and server URL). The local file only stores the link context.

## Checking the current link

```bash
ancla status
```

Shows the linked org, project, and app along with the current pipeline status (build/release/deploy).

```
Org:     my-org
Project: my-project
App:     my-app

STAGE    STATUS
Build    complete
Release  complete
Deploy   running
```

## Unlinking

```bash
ancla unlink
```

Removes the local `.ancla/config.yaml`. Commands that depend on a link will prompt you to re-link.

## How link resolution works

When you run a command that needs an app context, the CLI looks for `.ancla/config.yaml` starting from your current directory and walking up toward the filesystem root. The first one it finds wins.

This means you can have a monorepo with per-service links:

```
my-monorepo/
  .ancla/config.yaml          # org + project level
  services/
    api/
      .ancla/config.yaml      # linked to the api app
    frontend/
      .ancla/config.yaml      # linked to the frontend app
```

## Link vs. explicit arguments

Every command that uses the link context also accepts an explicit argument. The argument always wins:

```bash
ancla ssh                              # uses linked app
ancla ssh other-org/other-proj/other   # ignores link, uses argument
```
