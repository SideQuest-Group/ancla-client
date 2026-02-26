---
title: Configuration
description: How the Ancla CLI resolves settings from flags, env vars, and config files.
---

## Config precedence

The CLI resolves settings in this order (highest priority first):

1. **CLI flags** — `--api-key`
2. **Environment variables** — `ANCLA_API_KEY`
3. **Local config** — `.ancla/config.yaml` in the current directory or any parent
4. **Global config** — `~/.ancla/config.yaml`

## Config file format

```yaml
# ~/.ancla/config.yaml
api_key: ancla_your_key_here
```

## Per-project config

Create a `.ancla/config.yaml` in your project root to override global settings for that project. The CLI walks up from the current working directory looking for a `.ancla/` directory.

```bash
mkdir .ancla
echo 'api_key: ancla_project_specific_key' > .ancla/config.yaml
```

This is useful for using different API keys per project or workspace.

## Config var scopes

Config variables can be set at different scopes in the resource hierarchy. Use the `--scope` flag to target a specific level:

| Scope | Flag | Description |
|-------|------|-------------|
| Workspace | `--scope workspace` | Inherited by all projects, environments, and services in the workspace |
| Project | `--scope project` | Inherited by all environments and services in the project |
| Environment | `--scope env` | Inherited by all services in the environment |
| Service | `--scope service` (default) | Applies only to the specific service |

### Examples

```bash
# Set a config var at the workspace level (inherited by everything)
ancla config set my-ws KEY=val --scope workspace

# Set at the environment level
ancla config set my-ws/my-project/production KEY=val --scope env

# Set at the service level (default)
ancla config set my-ws/my-project/production/my-service KEY=val

# List config vars at the project scope
ancla config list my-ws/my-project --scope project
```

Lower scopes override higher scopes. A service-level config var with the same name as a workspace-level one takes precedence for that service.

## Managing settings

### View current settings

```bash
ancla settings show
```

API keys are masked in output.

### Set a value

```bash
ancla settings set api_key ancla_your_key_here
```

### Open in editor

```bash
ancla settings edit
```

Opens the config file in `$EDITOR` (defaults to `vi`).

### Show config file paths

```bash
ancla settings path
```

Shows both the global and local (if found) config file locations.
