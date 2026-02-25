---
title: Getting Started
description: Install the Ancla CLI and deploy your first application.
---

## Installation

### Pre-built binaries

Download the latest release from [GitHub Releases](https://github.com/SideQuest-Group/ancla-client/releases) for your platform.

```bash
# macOS (Apple Silicon)
curl -Lo ancla https://github.com/SideQuest-Group/ancla-client/releases/latest/download/ancla_darwin_arm64
chmod +x ancla
sudo mv ancla /usr/local/bin/

# macOS (Intel)
curl -Lo ancla https://github.com/SideQuest-Group/ancla-client/releases/latest/download/ancla_darwin_amd64
chmod +x ancla
sudo mv ancla /usr/local/bin/

# Linux (amd64)
curl -Lo ancla https://github.com/SideQuest-Group/ancla-client/releases/latest/download/ancla_linux_amd64
chmod +x ancla
sudo mv ancla /usr/local/bin/
```

### From source

Requires Go 1.24+.

```bash
go install github.com/SideQuest-Group/ancla-client/cmd/ancla@latest
```

### Using uv (Python tool runner)

If you have [uv](https://docs.astral.sh/uv/) and the ancla package published:

```bash
uvx ancla
```

## Quick start

### 1. Log in

```bash
ancla login
```

This opens your browser for authentication. See the [Authentication guide](/guides/authentication/) for CI and manual options.

### 2. List your organizations

```bash
ancla orgs list
```

### 3. List projects

```bash
ancla projects list --org <org-slug>
```

### 4. Deploy an application

```bash
ancla apps deploy --app <app-slug>
```

### 5. Check deployment status

```bash
ancla apps status --app <app-slug>
```

## Next steps

- [Authentication](/guides/authentication/) — browser login, manual keys, CI setup
- [Configuration](/guides/configuration/) — config files, env vars, precedence
- [Shell Completion](/guides/shell-completion/) — tab completion for your shell
- [CLI Reference](/cli/ancla/) — full command documentation
