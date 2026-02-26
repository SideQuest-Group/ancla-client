---
title: Terraform / OpenTofu Provider
description: Manage Ancla resources with infrastructure-as-code.
---

The `sidequest-labs/ancla` provider lets you manage Ancla workspaces, projects, environments, services, and config variables as Terraform or OpenTofu resources.

## Install

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    ancla = {
      source = "registry.terraform.io/sidequest-labs/ancla"
    }
  }
}
```

Run `terraform init` to download the provider.

## Configure the provider

```hcl
provider "ancla" {
  api_key = var.ancla_api_key
}
```

| Attribute | Environment Variable | Default | Description |
|-----------|---------------------|---------|-------------|
| `api_key` | `ANCLA_API_KEY` | | API key for authentication |
| `server` | `ANCLA_SERVER` | `https://ancla.dev` | Ancla server URL |

Both attributes can be set via environment variables instead. `ANCLA_API_KEY` is the recommended approach for CI.

## Resources

### ancla_workspace

```hcl
resource "ancla_workspace" "main" {
  name = "My Workspace"
}
```

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | yes | Workspace display name |

**Read-only:** `id`, `slug`

Supports `terraform import ancla_workspace.main <slug>`.

### ancla_project

```hcl
resource "ancla_project" "web" {
  name           = "Web Platform"
  workspace_slug = ancla_workspace.main.slug
}
```

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | yes | Project name |
| `workspace_slug` | string | yes | Slug of the parent workspace |

**Read-only:** `id`, `slug`

Supports `terraform import ancla_project.web <workspace-slug>/<project-slug>`.

### ancla_environment

```hcl
resource "ancla_environment" "production" {
  name           = "production"
  workspace_slug = ancla_workspace.main.slug
  project_slug   = ancla_project.web.slug
}

resource "ancla_environment" "staging" {
  name           = "staging"
  workspace_slug = ancla_workspace.main.slug
  project_slug   = ancla_project.web.slug
}
```

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | yes | Environment name (e.g. production, staging) |
| `workspace_slug` | string | yes | Parent workspace slug |
| `project_slug` | string | yes | Parent project slug |

**Read-only:** `id`, `slug`

Supports `terraform import ancla_environment.production <workspace-slug>/<project-slug>/<env-slug>`.

### ancla_service

```hcl
resource "ancla_service" "api" {
  name           = "API Service"
  workspace_slug = ancla_workspace.main.slug
  project_slug   = ancla_project.web.slug
  env_slug       = ancla_environment.production.slug
  platform       = "docker"

  github_repository  = "sidequest-labs/api-service"
  auto_deploy_branch = "main"

  process_counts = {
    web    = 2
    worker = 1
  }
}
```

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | yes | Service name |
| `workspace_slug` | string | yes | Parent workspace slug |
| `project_slug` | string | yes | Parent project slug |
| `env_slug` | string | yes | Parent environment slug |
| `platform` | string | yes | Platform type (e.g. `docker`) |
| `github_repository` | string | no | GitHub repo (owner/name) |
| `auto_deploy_branch` | string | no | Branch that triggers auto-deploy |
| `process_counts` | map(number) | no | Process scaling counts |

**Read-only:** `id`, `slug`

Supports `terraform import ancla_service.api <workspace-slug>/<project-slug>/<env-slug>/<service-slug>`.

### ancla_config_var

```hcl
resource "ancla_config_var" "database_url" {
  service_id = ancla_service.api.id
  name       = "DATABASE_URL"
  value      = "postgres://localhost:5432/mydb"
}

resource "ancla_config_var" "secret_key" {
  service_id = ancla_service.api.id
  name       = "SECRET_KEY"
  value      = "super-secret-value"
  secret     = true
  buildtime  = false
}
```

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `service_id` | string | yes | ID of the parent service |
| `name` | string | yes | Variable name |
| `value` | string | yes | Variable value |
| `secret` | bool | no | Mark as secret (hidden in UI/API) |
| `buildtime` | bool | no | Available at build time |
| `scope` | string | no | Scope level: `workspace`, `project`, `env`, or `service` (default) |

**Read-only:** `id`

Supports `terraform import ancla_config_var.database_url <config-id>`.

## Data sources

### data.ancla_workspace

Look up an existing workspace by slug:

```hcl
data "ancla_workspace" "existing" {
  slug = "my-ws"
}
```

Returns `name`, `slug`, `member_count`, `project_count`.

### data.ancla_project

```hcl
data "ancla_project" "existing" {
  workspace_slug = "my-ws"
  slug           = "my-project"
}
```

Returns `name`, `slug`, `workspace_slug`, `environment_count`.

### data.ancla_environment

```hcl
data "ancla_environment" "existing" {
  workspace_slug = "my-ws"
  project_slug   = "my-project"
  slug           = "production"
}
```

Returns `name`, `slug`, `workspace_slug`, `project_slug`, `service_count`.

### data.ancla_service

```hcl
data "ancla_service" "existing" {
  workspace_slug = "my-ws"
  project_slug   = "my-project"
  env_slug       = "production"
  slug           = "my-service"
}
```

Returns `id`, `name`, `slug`, `platform`, `github_repository`, `auto_deploy_branch`, `process_counts`.

## Full example

```hcl
terraform {
  required_providers {
    ancla = {
      source = "registry.terraform.io/sidequest-labs/ancla"
    }
  }
}

provider "ancla" {}

resource "ancla_workspace" "acme" {
  name = "Acme Corp"
}

resource "ancla_project" "backend" {
  name           = "Backend Services"
  workspace_slug = ancla_workspace.acme.slug
}

resource "ancla_environment" "production" {
  name           = "production"
  workspace_slug = ancla_workspace.acme.slug
  project_slug   = ancla_project.backend.slug
}

resource "ancla_environment" "staging" {
  name           = "staging"
  workspace_slug = ancla_workspace.acme.slug
  project_slug   = ancla_project.backend.slug
}

resource "ancla_service" "api" {
  name           = "REST API"
  workspace_slug = ancla_workspace.acme.slug
  project_slug   = ancla_project.backend.slug
  env_slug       = ancla_environment.production.slug
  platform       = "docker"
  github_repository  = "acme/rest-api"
  auto_deploy_branch = "main"

  process_counts = {
    web = 2
  }
}

resource "ancla_config_var" "db" {
  service_id = ancla_service.api.id
  name       = "DATABASE_URL"
  value      = "postgres://db.internal:5432/api"
  secret     = true
}

output "api_service_id" {
  value = ancla_service.api.id
}
```
