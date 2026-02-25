---
page_title: "ancla_config Resource - Ancla"
subcategory: ""
description: |-
  Manages a configuration variable for an Ancla application.
---

# ancla_config (Resource)

Manages a configuration variable (environment variable) for an Ancla application. Configuration variables are key-value pairs injected into the application runtime. Variables can optionally be marked as secrets or made available at build time.

## Example Usage

### Basic Configuration Variable

```terraform
resource "ancla_config" "database_url" {
  app_id = ancla_app.api.id
  name   = "DATABASE_URL"
  value  = "postgres://localhost:5432/mydb"
}
```

### Secret Configuration Variable

```terraform
resource "ancla_config" "secret_key" {
  app_id    = ancla_app.api.id
  name      = "SECRET_KEY"
  value     = "super-secret-value"
  secret    = true
  buildtime = false
}
```

## Schema

### Required

- `app_id` (String) The application ID this configuration variable belongs to. Changing this forces a new resource to be created.
- `name` (String) The name (key) of the configuration variable. Changing this forces a new resource to be created.
- `value` (String, Sensitive) The value of the configuration variable.

### Optional

- `secret` (Boolean) Whether this variable is a secret. Secret values are hidden by default in API responses. Defaults to `false`.
- `buildtime` (Boolean) Whether this variable is available at build time. Defaults to `false`.

### Read-Only

- `id` (String) The unique identifier of the configuration variable.

~> **Note:** When `secret` is set to `true`, the API returns a masked value on subsequent reads. Terraform will retain the value from the original configuration and will not detect external changes to the secret value.

## Import

Configuration variables can be imported using the format `<app_id>/<config_id>`.

```shell
terraform import ancla_config.database_url 01234567-abcd-efgh-ijkl-0123456789ab/98765432-dcba-hgfe-lkji-ba9876543210
```
