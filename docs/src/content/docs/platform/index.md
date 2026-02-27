---
title: Platform Overview
description: How Ancla organizes your infrastructure — workspaces, projects, environments, and services.
---

Ancla is a deployment platform. You give it a container (or source code and a Dockerfile), tell it where to run, and it handles the rest: builds, TLS, secrets, rollouts, rollbacks, scaling, logs, shell access.

The platform runs on Kubernetes, but you don't need to know that. The CLI, API, and dashboard abstract the infrastructure away. If you *do* want to know, the stack is K8s + Consul + Vault + BuildKit + a private container registry.

## Resource hierarchy

Everything in Ancla is organized into four levels:

```
Workspace
└── Project
    └── Environment
        └── Service
```

**Workspaces** are the top level. A workspace maps to a team, an org, or whatever boundary makes sense for you. Each workspace gets its own Kubernetes namespace, its own member list, and its own config scope.

**Projects** live inside workspaces. A project groups related services — your API, your frontend, your worker, whatever runs together. Projects can have their own config variables that all services within them inherit.

**Environments** are deployment targets within a project. The common setup is `production`, `staging`, and `development`, but you can name them whatever you want. Environments can also be ephemeral — Ancla creates preview environments from pull requests and tears them down when the PR closes.

**Services** are the actual running things. One service = one container. Each service within an environment has its own build history, deploy history, config variables, process scaling, health checks, and logs.

The addressing scheme for all of this is slash-separated:

```
my-workspace/my-project/production/api
```

That string identifies a specific service (`api`) in a specific environment (`production`) of a specific project, inside a specific workspace. You'll see this pattern everywhere — in the CLI, the API paths, the dashboard URLs.

## What a service contains

A service tracks:

- **Builds** — versioned container images, auto-incrementing (`v1`, `v2`, `v3`, ...). Each build records its Dockerfile, Procfile, build log, commit SHA, and whether it succeeded.
- **Deploys** — versioned snapshots of an image + config rolled out to the cluster. A deploy is immutable once created. If a deploy fails, the previous one stays active.
- **Config variables** — key/value pairs injected as environment variables at runtime. Can be plaintext or encrypted secrets stored in Vault.
- **Process types** — defined by your Procfile (`web`, `worker`, `beat`, etc.). Each type is scaled independently.

## How things connect

```
GitHub push → Build (BuildKit) → Deploy (K8s rollout)
     │                                    │
     └── Webhook                          ├── Consul (service registration)
                                          ├── Vault (secret injection)
                                          └── TLS (auto-provisioned)
```

When you push to a branch with auto-deploy enabled, or run `ancla services deploy`, the platform:

1. Clones your repo (or pulls from the registry if you pushed a pre-built image)
2. Builds a container image with BuildKit
3. Stores the image in the private registry
4. Creates a new deploy record — a frozen snapshot of the image + your current config
5. Rolls out the deploy to Kubernetes with zero downtime
6. Registers the service in Consul for discovery
7. Injects secrets from Vault at runtime via envconsul
8. Reports health checks and marks the deploy active (or rolls back if health checks fail)

See [Deploy Pipeline](/platform/deploy-pipeline/) for the full breakdown.

## Teams and access control

Workspaces have members. Members have roles. Roles control what you can see and do.

Teams within a workspace let you organize members into groups and assign config variables at the team scope. A team doesn't change what you can access — that's still role-based — but it does let you share config across services owned by a particular group.

## Config inheritance

Config variables cascade down the hierarchy:

```
Workspace config
  └── Team config
        └── Project config
              └── Environment config
                    └── Service config (highest priority)
```

A `DATABASE_URL` set at the project level applies to every service in every environment of that project — unless an environment or service overrides it. The most specific scope wins.

See [Secrets & Config](/platform/secrets-and-config/) for how this works in practice.

## Promotions

Ancla supports promoting a deploy from one environment to another. When you promote `staging` → `production`, the platform takes the exact build and config snapshot from staging and deploys it to production. Same image, same config resolution, different target.

Promotions have a preview step — you can see what will change before executing.

## Tools

You interact with the platform through:

- **CLI** (`ancla`) — the primary interface. Deploy, scale, configure, debug, all from your terminal.
- **REST API** — every CLI command maps to an API call. Build your own integrations.
- **SDKs** — Python, Go, and TypeScript clients generated from the OpenAPI spec.
- **Terraform provider** — manage Ancla resources as infrastructure-as-code.
- **Dashboard** — web UI at ancla.dev for when you want a visual overview.

The CLI and SDKs talk to the same API. The dashboard is a separate Astro frontend backed by the same API. There's no hidden admin API — everything the dashboard does, the CLI can do.
