---
title: Scripting & Automation
description: Use JSON output, quiet mode, and follow flags for CI/CD and scripts.
---

The CLI is built for both interactive use and automation. Three flags control output behavior across all commands.

## JSON output

Every command that prints a table also supports JSON:

```bash
ancla workspaces list --json
```

```json
[
  {
    "name": "My Workspace",
    "slug": "my-ws",
    "member_count": 3,
    "project_count": 2
  }
]
```

The long form works too:

```bash
ancla workspaces list --output json
```

Pipe JSON into `jq` for extraction:

```bash
ancla services list my-ws/my-project/production --json | jq '.[].slug'
```

## Quiet mode

Suppress spinners, progress messages, and confirmations:

```bash
ancla services deploy my-ws/my-project/production/my-service --quiet
```

Short form:

```bash
ancla services deploy my-ws/my-project/production/my-service -q
```

In quiet mode, the CLI prints only essential output: IDs, final status, and errors. Combine with `--json` for fully machine-parseable output:

```bash
ancla services deploy my-ws/my-project/production/my-service -q --json
```

## Following long-running operations

Build and deploy commands accept `--follow` (or `-f`) to stream output until the operation completes:

```bash
ancla services deploy my-ws/my-project/production/my-service --follow
ancla builds create my-ws/my-project/production/my-service --follow
ancla deploys log <deploy-id> --follow
ancla logs -f
```

Without `--follow`, these commands print the current state and exit.

## Skipping confirmation prompts

Destructive commands (`down`, `cache flush`, `config delete`) prompt for confirmation in interactive use. Skip the prompt with `--yes`:

```bash
ancla down --yes
ancla cache flush --yes
```

## CI/CD example

A GitHub Actions step that deploys and waits for completion:

```yaml
- name: Deploy to Ancla
  env:
    ANCLA_API_KEY: ${{ secrets.ANCLA_API_KEY }}
  run: |
    ancla services deploy my-ws/my-project/production/my-service --follow --quiet
```

The `ANCLA_API_KEY` env var is picked up automatically. No config file or login step needed.

## Exit codes

The CLI exits with code 0 on success and non-zero on failure. Deploy failures, auth errors, and invalid arguments all produce non-zero exits, so `set -e` in shell scripts works as expected.
