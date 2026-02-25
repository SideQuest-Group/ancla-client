---
title: OpenTofu Provider
description: Manage Ancla resources with infrastructure-as-code.
---

:::note
The Ancla OpenTofu provider is under development. This section will be populated when the provider is available.
:::

## Planned resources

The provider will support managing:

- **Organizations** — create and configure organizations
- **Projects** — create projects within organizations
- **Applications** — define applications with build and deploy settings
- **Configuration** — manage application environment variables
- **Hooks** — configure deployment webhooks

## Registry

The provider will be published to the [OpenTofu Registry](https://github.com/opentofu/registry) once available.

## Example (preview)

```hcl
terraform {
  required_providers {
    ancla = {
      source = "sidequest-labs/ancla"
    }
  }
}

provider "ancla" {
  server  = "https://ancla.dev"
  api_key = var.ancla_api_key
}

resource "ancla_project" "myapp" {
  organization = "my-org"
  name         = "my-project"
}
```
