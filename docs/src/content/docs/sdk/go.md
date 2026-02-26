---
title: Go SDK
description: The ancla-go package — install, configure, and use from Go programs.
---

## Install

Requires Go 1.21+.

```bash
go get github.com/sidequest-labs/ancla-go
```

## Create a client

```go
import ancla "github.com/sidequest-labs/ancla-go"

client := ancla.New("ancla_your_key_here")
```

To use a custom server or HTTP client:

```go
client := ancla.New(
    "ancla_your_key_here",
    ancla.WithServer("https://ancla.example.com"),
    ancla.WithHTTPClient(&http.Client{Timeout: 60 * time.Second}),
)
```

The client wraps the provided `http.Client`'s transport to inject the `X-API-Key` header on every request.

All methods take a `context.Context` as their first argument.

## Organizations

```go
ctx := context.Background()

orgs, err := client.ListOrgs(ctx)

org, err := client.GetOrg(ctx, "my-org")

newOrg, err := client.CreateOrg(ctx, "New Org")

updated, err := client.UpdateOrg(ctx, "my-org", "Renamed Org")

err = client.DeleteOrg(ctx, "old-org")
```

## Projects

```go
projects, err := client.ListProjects(ctx, "my-org")

project, err := client.GetProject(ctx, "my-org", "my-project")

newProject, err := client.CreateProject(ctx, "my-org", "New Project")

updated, err := client.UpdateProject(ctx, "my-org", "my-project", "Renamed")

err = client.DeleteProject(ctx, "my-org", "old-project")
```

## Applications

```go
apps, err := client.ListApps(ctx, "my-org", "my-project")

app, err := client.GetApp(ctx, "my-org", "my-project", "api-service")

newApp, err := client.CreateApp(ctx, "my-org", "my-project", "Worker", "docker")

updated, err := client.UpdateApp(ctx, "my-org", "my-project", "api-service", ancla.UpdateAppOptions{
    Name: ancla.StringPtr("Renamed"),
})

err = client.DeleteApp(ctx, "my-org", "my-project", "old-app")
```

### Deploy and scale

```go
result, err := client.DeployApp(ctx, "app-uuid")
fmt.Println(result.ImageID)

err = client.ScaleApp(ctx, "app-uuid", map[string]int{
    "web":    2,
    "worker": 1,
})

status, err := client.GetAppStatus(ctx, "app-uuid")
fmt.Println(status.Build.Status)  // "complete"
```

## Configuration

```go
vars, err := client.ListConfig(ctx, "app-uuid")

v, err := client.GetConfig(ctx, "app-uuid", "config-uuid")

err = client.SetConfig(ctx, "app-uuid", ancla.SetConfigRequest{
    Name:   "DATABASE_URL",
    Value:  "postgres://localhost/mydb",
    Secret: true,
})

err = client.DeleteConfig(ctx, "app-uuid", "config-uuid")
```

## Images

```go
images, err := client.ListImages(ctx, "app-uuid")
// images.Items is []Image

image, err := client.GetImage(ctx, "image-uuid")

buildResult, err := client.BuildImage(ctx, "app-uuid")
fmt.Println(buildResult.ImageID, buildResult.Version)
```

## Releases

```go
releases, err := client.ListReleases(ctx, "app-uuid")
// releases.Items is []Release

release, err := client.GetRelease(ctx, "release-uuid")

result, err := client.CreateRelease(ctx, "app-uuid")
fmt.Println(result.ReleaseID, result.Version)

deployResult, err := client.DeployRelease(ctx, "release-uuid")
fmt.Println(deployResult.DeploymentID)
```

## Deployments

```go
deployment, err := client.GetDeployment(ctx, "deployment-uuid")
fmt.Println(deployment.Complete, deployment.Error)

log, err := client.GetDeploymentLog(ctx, "deployment-uuid")
fmt.Println(log.LogText)
```

## Error handling

API errors are returned as `*ancla.APIError`:

```go
org, err := client.GetOrg(ctx, "nonexistent")
if err != nil {
    var apiErr *ancla.APIError
    if errors.As(err, &apiErr) {
        fmt.Println(apiErr.StatusCode, apiErr.Message)
    }
}
```

Helper functions for common checks:

```go
if ancla.IsNotFound(err) {
    // 404
}
if ancla.IsUnauthorized(err) {
    // 401
}
if ancla.IsForbidden(err) {
    // 403
}
```

## Types

All request/response types are exported from the package root:

**Resources:** `Org`, `OrgMember`, `Project`, `App`, `ConfigVar`, `Image`, `ImageList`, `ImageLog`, `Release`, `ReleaseList`, `Deployment`, `DeploymentLog`, `PipelineStatus`, `StageStatus`

**Requests:** `CreateOrgRequest`, `UpdateOrgRequest`, `CreateProjectRequest`, `UpdateProjectRequest`, `CreateAppRequest`, `UpdateAppOptions`, `ScaleRequest`, `SetConfigRequest`

**Responses:** `DeployResult`, `BuildResult`, `CreateReleaseResult`, `DeployReleaseResult`
