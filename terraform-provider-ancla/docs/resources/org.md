---
page_title: "ancla_org Resource - Ancla"
subcategory: ""
description: |-
  Manages an Ancla organization.
---

# ancla_org (Resource)

Manages an Ancla organization. Organizations are the top-level grouping for projects and applications on the Ancla platform.

## Example Usage

```terraform
resource "ancla_org" "example" {
  name = "My Organization"
}
```

## Schema

### Required

- `name` (String) The display name of the organization.

### Read-Only

- `id` (String) The unique identifier of the organization.
- `slug` (String) The URL-friendly slug of the organization. Derived from the name.
- `member_count` (Number) The number of members in the organization.
- `project_count` (Number) The number of projects in the organization.

## Import

Organizations can be imported using the organization slug.

```shell
terraform import ancla_org.example my-organization
```
