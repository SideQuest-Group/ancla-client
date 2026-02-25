---
page_title: "ancla_app Resource - Ancla"
subcategory: ""
description: |-
  Manages an Ancla application within a project.
---

# ancla_app (Resource)

Manages an Ancla application within a project. Applications represent deployable services on the Ancla platform.

## Example Usage

### Basic Application

```terraform
resource "ancla_app" "api" {
  name              = "API Service"
  organization_slug = ancla_org.example.slug
  project_slug      = ancla_project.web.slug
  platform          = "docker"
}
```

### Application with GitHub Integration and Scaling

```terraform
resource "ancla_app" "api" {
  name              = "API Service"
  organization_slug = ancla_org.example.slug
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

## Schema

### Required

- `name` (String) The display name of the application.
- `organization_slug` (String) The slug of the organization this application belongs to. Changing this forces a new resource to be created.
- `project_slug` (String) The slug of the project this application belongs to. Changing this forces a new resource to be created.
- `platform` (String) The platform type of the application (e.g., `docker`). Changing this forces a new resource to be created.

### Optional

- `github_repository` (String) The GitHub repository linked to this application, in `owner/repo` format.
- `auto_deploy_branch` (String) The branch that triggers automatic deployments.
- `process_counts` (Map of Number) Map of process type to replica count (e.g., `web = 2`, `worker = 1`).

### Read-Only

- `id` (String) The unique identifier of the application.
- `slug` (String) The URL-friendly slug of the application. Derived from the name.

## Import

Applications can be imported using the format `<organization_slug>/<project_slug>/<app_slug>`.

```shell
terraform import ancla_app.api my-organization/web-platform/api-service
```
