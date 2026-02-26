# ancla-client

CLI client for the [Ancla](https://ancla.dev) deployment platform.

## Installation

### Quick install (recommended)

```bash
curl -LsSf https://ancla.dev/install.sh | sh
```

Pin a specific version:

```bash
curl -LsSf https://ancla.dev/install.sh | sh -s -- --version v0.5.0
```

### Go

```bash
go install github.com/SideQuest-Group/ancla-client/cmd/ancla@latest
```

### Python / uv

```bash
pip install ancla-cli
# or
uv tool install ancla-cli
```

### GitHub Releases

Download a pre-built binary from the [Releases](https://github.com/SideQuest-Group/ancla-client/releases) page.

## Quick Start

```bash
# Authenticate (stores API key in ~/.ancla/config.yaml)
ancla login

# List organizations
ancla orgs list

# List projects
ancla projects list

# Deploy an application
ancla apps deploy <app-id>

# Check deployment status
ancla apps status <app-id>
```

## Configuration

The CLI reads configuration from (in order of precedence):

1. CLI flags (`--server`, `--api-key`)
2. Environment variables (`ANCLA_SERVER`, `ANCLA_API_KEY`)
3. Config file (`~/.ancla/config.yaml`)

## Commands

| Command | Description |
|---------|-------------|
| `ancla login` | Authenticate interactively |
| `ancla whoami` | Show current session |
| `ancla orgs list` | List organizations |
| `ancla orgs get <slug>` | Get organization details |
| `ancla projects list` | List projects |
| `ancla projects get <org>/<project>` | Get project details |
| `ancla apps list <org>/<project>` | List applications |
| `ancla apps get <org>/<project>/<app>` | Get application details |
| `ancla apps deploy <app-id>` | Trigger a full deploy |
| `ancla apps scale <app-id> web=3` | Scale processes |
| `ancla apps status <app-id>` | Pipeline status |
| `ancla images list <app-id>` | List images |
| `ancla images build <app-id>` | Trigger an image build |
| `ancla images log <image-id>` | Show build log |
| `ancla releases list <app-id>` | List releases |
| `ancla releases create <app-id>` | Create a release |
| `ancla releases deploy <release-id>` | Deploy a release |
| `ancla deployments get <id>` | Get deployment details |
| `ancla deployments log <id>` | Show deployment log |
| `ancla config list <app-id>` | List config vars |
| `ancla config set <app-id> KEY=val` | Set a config var |
| `ancla config delete <app-id> <id>` | Delete a config var |
| `ancla config import <app-id> -f .env` | Bulk import from .env |
| `ancla version` | Show CLI version |

## Development

```bash
make build    # Build binary to dist/ancla
make test     # Run tests
make vet      # Run go vet
make fmt      # Format code
make lint     # Run all linting checks
```

### Pre-commit hooks

This project uses [prek](https://prek.j178.dev/) for pre-commit hooks.

```bash
prek install          # Set up git hooks
prek run --all-files  # Run all checks manually
```

## License

Apache License 2.0 — see [LICENSE](LICENSE).
