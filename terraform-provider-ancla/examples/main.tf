terraform {
  required_providers {
    ancla = {
      source = "registry.terraform.io/sidequest-labs/ancla"
    }
  }
}

provider "ancla" {
  # server  = "https://ancla.dev"   # Optional, defaults to https://ancla.dev
  # api_key = "your-api-key"        # Or set ANCLA_API_KEY env var
}

# --- Organizations ---

resource "ancla_org" "example" {
  name = "My Organization"
}

data "ancla_org" "existing" {
  slug = "existing-org"
}

# --- Projects ---

resource "ancla_project" "web" {
  name              = "Web Platform"
  organization_slug = ancla_org.example.slug
}

data "ancla_project" "existing" {
  organization_slug = "existing-org"
  slug              = "existing-project"
}

# --- Applications ---

resource "ancla_app" "api" {
  name              = "API Service"
  organization_slug = ancla_org.example.slug
  project_slug      = ancla_project.web.slug
  platform          = "docker"

  github_repository = "sidequest-labs/api-service"
  auto_deploy_branch = "main"

  process_counts = {
    web    = 2
    worker = 1
  }
}

data "ancla_app" "existing" {
  organization_slug = "existing-org"
  project_slug      = "existing-project"
  slug              = "existing-app"
}

# --- Configuration ---

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

# --- Outputs ---

output "org_slug" {
  value = ancla_org.example.slug
}

output "app_id" {
  value = ancla_app.api.id
}
