---
page_title: "ancla_org Data Source - Ancla"
subcategory: ""
description: |-
  Reads an Ancla organization by slug.
---

# ancla_org (Data Source)

Use this data source to read information about an existing Ancla organization.

## Example Usage

```terraform
data "ancla_org" "existing" {
  slug = "existing-org"
}

output "org_name" {
  value = data.ancla_org.existing.name
}
```

## Schema

### Required

- `slug` (String) The URL-friendly slug of the organization.

### Read-Only

- `id` (String) The unique identifier of the organization.
- `name` (String) The display name of the organization.
- `member_count` (Number) The number of members in the organization.
- `project_count` (Number) The number of projects in the organization.
- `application_count` (Number) The total number of applications across all projects.
