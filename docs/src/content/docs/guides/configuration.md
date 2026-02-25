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
