---
title: Remote Access
description: SSH into containers, open database shells, and manage cache services.
---

These commands give you direct access to running containers and their attached services. All of them use the linked app context or accept an explicit `org/project/app` argument.

## SSH

Open an interactive SSH session to a running container:

```bash
ancla ssh
```

Connect to a specific process type (defaults to `web`):

```bash
ancla ssh --process worker
```

With an explicit app path:

```bash
ancla ssh my-org/my-project/my-app
```

The CLI requests ephemeral credentials from the Ancla API and launches `ssh` with them. You need OpenSSH installed locally. No SSH keys to manage; the platform handles authentication.

## Shell

Similar to `ssh`, but uses the platform's exec API directly. No SSH client needed.

```bash
ancla shell
```

Choose the process type and shell command:

```bash
ancla shell -p worker
ancla shell -c /bin/bash
```

Default process type is `web`. Default command is `/bin/sh`.

## Database shell

Connect to your app's primary database:

```bash
ancla dbshell
```

The CLI detects the database engine and launches the right client:

- **PostgreSQL** &rarr; `psql`
- **MySQL** &rarr; `mysql`
- **Other** &rarr; prints the connection URL for manual use

Credentials come from the platform. You need `psql` or `mysql` installed locally.

```bash
ancla dbshell my-org/my-project/my-app
```

With `--json`, the command prints connection details without opening a shell (passwords omitted):

```bash
ancla dbshell --json
```

```json
{
  "engine": "postgresql",
  "host": "db-abc123.ancla.internal",
  "port": 5432,
  "name": "myapp_production",
  "user": "myapp"
}
```

## Cache management

View cache service details:

```bash
ancla cache info
```

```
Engine: redis
Host:   cache-abc123.ancla.internal
Port:   6379
```

Open an interactive Redis CLI session:

```bash
ancla cache cli
```

For Redis, this launches `redis-cli` with the right host, port, and auth. For other cache engines, it prints the connection URL.

Flush the cache:

```bash
ancla cache flush
```

This prompts for confirmation. Skip it with `--yes`:

```bash
ancla cache flush --yes
```

## Prerequisites

These commands launch local client binaries. Make sure you have the right ones installed:

| Command | Requires |
|---------|----------|
| `ancla ssh` | `ssh` (OpenSSH) |
| `ancla dbshell` | `psql` or `mysql` |
| `ancla cache cli` | `redis-cli` (for Redis) |

All other commands (`shell`, `cache info`, `cache flush`) work without extra dependencies.
