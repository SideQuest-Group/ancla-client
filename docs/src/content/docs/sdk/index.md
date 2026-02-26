---
title: SDKs
description: Programmatic access to the Ancla platform from Python, Go, and TypeScript.
---

Three official SDKs wrap the Ancla REST API. Pick the one that fits your stack.

| SDK | Package | Install |
|-----|---------|---------|
| [Python](/sdk/python/) | `ancla-sdk` | `pip install ancla-sdk` |
| [Go](/sdk/go/) | `github.com/sidequest-labs/ancla-go` | `go get github.com/sidequest-labs/ancla-go` |
| [TypeScript](/sdk/typescript/) | `@ancla/sdk` | `npm install @ancla/sdk` |

All three SDKs:

- Authenticate via API key (passed directly or read from `ANCLA_API_KEY` env var)
- Default to `https://ancla.dev` as the server URL
- Map HTTP errors to typed exceptions/errors
- Cover the same resources: orgs, projects, apps, config, images, releases, deployments

## Authentication

Every SDK reads the `ANCLA_API_KEY` environment variable automatically. You can also pass the key explicitly at client creation time. The key is sent as an `X-API-Key` header on every request.

Generate an API key by running `ancla login` in the CLI, or create one in the Ancla web UI under account settings.

## Which SDK to use

**Python** if you're building deployment scripts, CI integrations, or internal tooling in Python. Uses `httpx` and `pydantic` under the hood. Synchronous by default; works as a context manager.

**Go** if you're building Go services that need to interact with Ancla, or extending the CLI itself. Zero external dependencies beyond the standard library.

**TypeScript** if you're building Node.js tooling, serverless functions, or browser-based dashboards. Uses native `fetch` with no runtime dependencies.
