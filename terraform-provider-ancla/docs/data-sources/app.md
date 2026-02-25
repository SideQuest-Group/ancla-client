---
page_title: "ancla_app Data Source - Ancla"
subcategory: ""
description: |-
  Reads an Ancla application by organization, project, and app slug.
---

# ancla_app (Data Source)

Use this data source to read information about an existing Ancla application.

## Example Usage

```terraform
data "ancla_app" "existing" {
  organization_slug = "existing-org"
  project_slug      = "existing-project"
  slug              = "existing-app"
}

output "app_platform" {
  value = data.ancla_app.existing.platform
}
```

## Schema

### Required

- `organization_slug` (String) The slug of the organization.
- `project_slug` (String) The slug of the project.
- `slug` (String) The URL-friendly slug of the application.

### Read-Only

- `id` (String) The unique identifier of the application.
- `name` (String) The display name of the application.
- `platform` (String) The platform type of the application.
- `github_repository` (String) The GitHub repository linked to this application.
- `auto_deploy_branch` (String) The branch that triggers automatic deployments.
- `process_counts` (Map of Number) Map of process type to replica count.
