---
title: Terraform / OpenTofu Provider
description: Manage Ancla resources with infrastructure-as-code.
---

The `sidequest-labs/ancla` provider lets you manage Ancla orgs, projects, apps, and config variables as Terraform or OpenTofu resources.

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

### ancla_org

```hcl
resource "ancla_org" "main" {
  name = "My Organization"
}
```

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | yes | Organization display name |

**Read-only:** `id`, `slug`

Supports `terraform import ancla_org.main <slug>`.

### ancla_project

```hcl
resource "ancla_project" "web" {
  name              = "Web Platform"
  organization_slug = ancla_org.main.slug
}
```

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | yes | Project name |
| `organization_slug` | string | yes | Slug of the parent organization |

**Read-only:** `id`, `slug`

Supports `terraform import ancla_project.web <org-slug>/<project-slug>`.

### ancla_app

```hcl
resource "ancla_app" "api" {
  name              = "API Service"
  organization_slug = ancla_org.main.slug
  project_slug      = ancla_project.web.slug
  platform          = "docker"

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
| `name` | string | yes | App name |
| `organization_slug` | string | yes | Parent org slug |
| `project_slug` | string | yes | Parent project slug |
| `platform` | string | yes | Platform type (e.g. `docker`) |
| `github_repository` | string | no | GitHub repo (owner/name) |
| `auto_deploy_branch` | string | no | Branch that triggers auto-deploy |
| `process_counts` | map(number) | no | Process scaling counts |

**Read-only:** `id`, `slug`

Supports `terraform import ancla_app.api <org-slug>/<project-slug>/<app-slug>`.

### ancla_config

```hcl
resource "ancla_config" "database_url" {
  app_id = ancla_app.api.id
  name   = "DATABASE_URL"
  value  = "postgres://localhost:5432/mydb"
}

resource "ancla_config" "secret_key" {
  app_id    = ancla_app.api.id
  name      = "SECRET_KEY"
  value     = "super-secret-value"
  secret    = true
  buildtime = false
}
```

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `app_id` | string | yes | ID of the parent application |
| `name` | string | yes | Variable name |
| `value` | string | yes | Variable value |
| `secret` | bool | no | Mark as secret (hidden in UI/API) |
| `buildtime` | bool | no | Available at build time |

**Read-only:** `id`

Supports `terraform import ancla_config.database_url <config-id>`.

## Data sources

### data.ancla_org

Look up an existing organization by slug:

```hcl
data "ancla_org" "existing" {
  slug = "my-org"
}
```

Returns `name`, `slug`, `member_count`, `project_count`.

### data.ancla_project

```hcl
data "ancla_project" "existing" {
  organization_slug = "my-org"
  slug              = "my-project"
}
```

Returns `name`, `slug`, `organization_slug`, `application_count`.

### data.ancla_app

```hcl
data "ancla_app" "existing" {
  organization_slug = "my-org"
  project_slug      = "my-project"
  slug              = "my-app"
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

resource "ancla_org" "acme" {
  name = "Acme Corp"
}

resource "ancla_project" "backend" {
  name              = "Backend Services"
  organization_slug = ancla_org.acme.slug
}

resource "ancla_app" "api" {
  name              = "REST API"
  organization_slug = ancla_org.acme.slug
  project_slug      = ancla_project.backend.slug
  platform          = "docker"
  github_repository = "acme/rest-api"
  auto_deploy_branch = "main"

  process_counts = {
    web = 2
  }
}

resource "ancla_config" "db" {
  app_id = ancla_app.api.id
  name   = "DATABASE_URL"
  value  = "postgres://db.internal:5432/api"
  secret = true
}

output "api_app_id" {
  value = ancla_app.api.id
}
```
