---
page_title: "ancla_project Data Source - Ancla"
subcategory: ""
description: |-
  Reads an Ancla project by organization slug and project slug.
---

# ancla_project (Data Source)

Use this data source to read information about an existing Ancla project.

## Example Usage

```terraform
data "ancla_project" "existing" {
  organization_slug = "existing-org"
  slug              = "existing-project"
}

output "project_name" {
  value = data.ancla_project.existing.name
}
```

## Schema

### Required

- `organization_slug` (String) The slug of the organization this project belongs to.
- `slug` (String) The URL-friendly slug of the project.

### Read-Only

- `id` (String) The unique identifier of the project.
- `name` (String) The display name of the project.
- `application_count` (Number) The number of applications in the project.
- `created` (String) The creation timestamp of the project.
- `updated` (String) The last update timestamp of the project.
