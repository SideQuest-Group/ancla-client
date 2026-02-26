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

## Workspaces

```go
ctx := context.Background()

workspaces, err := client.ListWorkspaces(ctx)

ws, err := client.GetWorkspace(ctx, "my-ws")

newWS, err := client.CreateWorkspace(ctx, "New Workspace")

updated, err := client.UpdateWorkspace(ctx, "my-ws", "Renamed Workspace")

err = client.DeleteWorkspace(ctx, "old-ws")
```

## Projects

```go
projects, err := client.ListProjects(ctx, "my-ws")

project, err := client.GetProject(ctx, "my-ws", "my-project")

newProject, err := client.CreateProject(ctx, "my-ws", "New Project")

updated, err := client.UpdateProject(ctx, "my-ws", "my-project", "Renamed")

err = client.DeleteProject(ctx, "my-ws", "old-project")
```

## Environments

```go
envs, err := client.ListEnvironments(ctx, "my-ws", "my-project")

env, err := client.GetEnvironment(ctx, "my-ws", "my-project", "production")

newEnv, err := client.CreateEnvironment(ctx, "my-ws", "my-project", "staging")

err = client.DeleteEnvironment(ctx, "my-ws", "my-project", "old-env")
```

## Services

```go
services, err := client.ListServices(ctx, "my-ws", "my-project", "production")

svc, err := client.GetService(ctx, "my-ws", "my-project", "production", "api")

newSvc, err := client.CreateService(ctx, "my-ws", "my-project", "production", "Worker", "docker")

updated, err := client.UpdateService(ctx, "my-ws", "my-project", "production", "api", ancla.UpdateServiceOptions{
    Name: ancla.StringPtr("Renamed"),
})

err = client.DeleteService(ctx, "my-ws", "my-project", "production", "old-svc")
```

### Deploy and scale

```go
result, err := client.DeployService(ctx, "svc-uuid")
fmt.Println(result.BuildID)

err = client.ScaleService(ctx, "svc-uuid", map[string]int{
    "web":    2,
    "worker": 1,
})

status, err := client.GetServiceStatus(ctx, "svc-uuid")
fmt.Println(status.Build.Status)  // "complete"
```

## Config vars

```go
vars, err := client.ListConfigVars(ctx, "svc-uuid")

v, err := client.GetConfigVar(ctx, "svc-uuid", "config-uuid")

err = client.SetConfigVar(ctx, "svc-uuid", ancla.SetConfigVarRequest{
    Name:   "DATABASE_URL",
    Value:  "postgres://localhost/mydb",
    Secret: true,
})

err = client.DeleteConfigVar(ctx, "svc-uuid", "config-uuid")
```

## Builds

```go
builds, err := client.ListBuilds(ctx, "svc-uuid")
// builds.Items is []Build

build, err := client.GetBuild(ctx, "build-uuid")

buildResult, err := client.CreateBuild(ctx, "svc-uuid")
fmt.Println(buildResult.BuildID, buildResult.Version)
```

## Deploys

```go
deploys, err := client.ListDeploys(ctx, "svc-uuid")
// deploys.Items is []Deploy

deploy, err := client.GetDeploy(ctx, "deploy-uuid")
fmt.Println(deploy.Complete, deploy.Error)

log, err := client.GetDeployLog(ctx, "deploy-uuid")
fmt.Println(log.LogText)
```

## Error handling

API errors are returned as `*ancla.APIError`:

```go
ws, err := client.GetWorkspace(ctx, "nonexistent")
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

**Resources:** `Workspace`, `WorkspaceMember`, `Project`, `Environment`, `Service`, `ConfigVar`, `Build`, `BuildList`, `BuildLog`, `Deploy`, `DeployList`, `DeployLog`, `PipelineStatus`, `StageStatus`

**Requests:** `CreateWorkspaceRequest`, `UpdateWorkspaceRequest`, `CreateProjectRequest`, `UpdateProjectRequest`, `CreateEnvironmentRequest`, `CreateServiceRequest`, `UpdateServiceOptions`, `ScaleRequest`, `SetConfigVarRequest`

**Responses:** `DeployResult`, `BuildResult`
