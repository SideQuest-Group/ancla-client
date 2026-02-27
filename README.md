# ancla-client

CLI client for the [Ancla](https://ancla.dev) deployment platform.

## Installation

### Go

```bash
go install ancla.dev/cli@latest
```

### Python

```bash
uv tool install ancla
# or
pip install ancla
```

### Homebrew

```bash
brew install SideQuest-Group/ancla-client/ancla
```

### npm / Bun

```bash
npx ancla
# or
bunx ancla
```

### Cargo

```bash
cargo binstall ancla
```

### GitHub Releases

Download a pre-built binary from the [Releases](https://github.com/SideQuest-Group/ancla-client/releases) page.

## Quick Start

```bash
# Authenticate (stores API key in ~/.ancla/config.yaml)
ancla login

# List workspaces
ancla workspaces list

# List projects
ancla projects list

# Deploy a service
ancla services deploy my-ws/my-project/production/my-service

# Check deploy status
ancla services status my-ws/my-project/production/my-service
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
| `ancla workspaces list` | List workspaces |
| `ancla workspaces get <slug>` | Get workspace details |
| `ancla projects list` | List projects |
| `ancla projects get <ws>/<project>` | Get project details |
| `ancla envs list <ws>/<project>` | List environments |
| `ancla envs get <ws>/<project>/<env>` | Get environment details |
| `ancla services list <ws>/<project>/<env>` | List services |
| `ancla services get <ws>/<project>/<env>/<svc>` | Get service details |
| `ancla services deploy <ws>/<project>/<env>/<svc>` | Trigger a full deploy |
| `ancla services scale <ws>/<project>/<env>/<svc> web=3` | Scale processes |
| `ancla services status <ws>/<project>/<env>/<svc>` | Pipeline status |
| `ancla builds list <svc-id>` | List builds |
| `ancla builds create <svc-id>` | Trigger a build |
| `ancla builds log <build-id>` | Show build log |
| `ancla deploys list <svc-id>` | List deploys |
| `ancla deploys get <id>` | Get deploy details |
| `ancla deploys log <id>` | Show deploy log |
| `ancla config list <svc-id>` | List config vars |
| `ancla config set <svc-id> KEY=val` | Set a config var |
| `ancla config delete <svc-id> <id>` | Delete a config var |
| `ancla config import <svc-id> -f .env` | Bulk import from .env |
| `ancla config list --scope workspace` | List config vars at workspace scope |
| `ancla version` | Show CLI version |

Full documentation at [docs.ancla.dev](https://docs.ancla.dev).

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
