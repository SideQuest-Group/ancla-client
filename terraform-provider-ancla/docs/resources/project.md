---
page_title: "ancla_project Resource - Ancla"
subcategory: ""
description: |-
  Manages an Ancla project within an organization.
---

# ancla_project (Resource)

Manages an Ancla project within an organization. Projects group related applications together under a single organization.

## Example Usage

```terraform
resource "ancla_org" "example" {
  name = "My Organization"
}

resource "ancla_project" "web" {
  name              = "Web Platform"
  organization_slug = ancla_org.example.slug
}
```

## Schema

### Required

- `name` (String) The display name of the project.
- `organization_slug` (String) The slug of the organization this project belongs to. Changing this forces a new resource to be created.

### Read-Only

- `id` (String) The unique identifier of the project.
- `slug` (String) The URL-friendly slug of the project. Derived from the name.
- `application_count` (Number) The number of applications in the project.

## Import

Projects can be imported using the format `<organization_slug>/<project_slug>`.

```shell
terraform import ancla_project.web my-organization/web-platform
```
