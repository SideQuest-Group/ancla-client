---
page_title: "Ancla Provider"
subcategory: ""
description: |-
  The Ancla provider is used to manage resources on the Ancla PaaS platform.
---

# Ancla Provider

The Ancla provider allows you to manage infrastructure on the [Ancla](https://ancla.dev) Platform-as-a-Service (PaaS). Use this provider to create and manage organizations, projects, applications, and application configuration variables.

## Authentication

The provider requires an API key for authentication. You can obtain an API key from the [Ancla dashboard](https://ancla.dev).

The API key can be provided in two ways:

1. **Provider configuration** (not recommended for version-controlled files):

```terraform
provider "ancla" {
  api_key = "your-api-key"
}
```

2. **Environment variable** (recommended):

```shell
export ANCLA_API_KEY="your-api-key"
```

## Example Usage

```terraform
terraform {
  required_providers {
    ancla = {
      source = "sidequest-labs/ancla"
    }
  }
}

provider "ancla" {
  # Configuration options
}

resource "ancla_org" "example" {
  name = "My Organization"
}

resource "ancla_project" "web" {
  name              = "Web Platform"
  organization_slug = ancla_org.example.slug
}

resource "ancla_app" "api" {
  name              = "API Service"
  organization_slug = ancla_org.example.slug
  project_slug      = ancla_project.web.slug
  platform          = "docker"
}

resource "ancla_config" "database_url" {
  app_id = ancla_app.api.id
  name   = "DATABASE_URL"
  value  = "postgres://localhost:5432/mydb"
}
```

## Schema

### Optional

- `server` (String) The Ancla server URL. Defaults to `https://ancla.dev`. Can also be set with the `ANCLA_SERVER` environment variable.
- `api_key` (String, Sensitive) The API key for authentication. Can also be set with the `ANCLA_API_KEY` environment variable.
