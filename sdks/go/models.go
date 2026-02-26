package ancla

// Workspace represents a workspace on the Ancla platform.
type Workspace struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Slug         string            `json:"slug"`
	MemberCount  int               `json:"member_count"`
	ProjectCount int               `json:"project_count"`
	ServiceCount int               `json:"service_count"`
	Members      []WorkspaceMember `json:"members,omitempty"`
}

// WorkspaceMember represents a member within a workspace.
type WorkspaceMember struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Admin    bool   `json:"admin"`
}

// Project represents a project within a workspace.
type Project struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	WorkspaceSlug string `json:"workspace_slug"`
	WorkspaceName string `json:"workspace_name"`
	ServiceCount  int    `json:"service_count"`
	Created       string `json:"created"`
	Updated       string `json:"updated"`
}

// Environment represents an environment within a project.
type Environment struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	ServiceCount int    `json:"service_count"`
	Created      string `json:"created"`
}

// Service represents a service within an environment.
type Service struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	Slug             string         `json:"slug"`
	Platform         string         `json:"platform"`
	GithubRepository string         `json:"github_repository,omitempty"`
	AutoDeployBranch string         `json:"auto_deploy_branch,omitempty"`
	ProcessCounts    map[string]int `json:"process_counts,omitempty"`
}

// Build represents a container build for a service.
type Build struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
	Built   bool   `json:"built"`
	Error   bool   `json:"error"`
	Created string `json:"created"`
}

// BuildList wraps the paginated build response.
type BuildList struct {
	Items []Build `json:"items"`
}

// BuildResult is the response from triggering a build.
type BuildResult struct {
	BuildID string `json:"build_id"`
	Version int    `json:"version"`
}

// BuildLog contains build log information.
type BuildLog struct {
	Status  string `json:"status"`
	Version int    `json:"version"`
	LogText string `json:"log_text"`
}

// Deploy represents a deploy for a service.
type Deploy struct {
	ID          string `json:"id"`
	Complete    bool   `json:"complete"`
	Error       bool   `json:"error"`
	ErrorDetail string `json:"error_detail"`
	JobID       string `json:"job_id"`
	Created     string `json:"created"`
	Updated     string `json:"updated"`
}

// DeployLog contains log information for a deploy.
type DeployLog struct {
	Status  string `json:"status"`
	LogText string `json:"log_text"`
}

// DeployList wraps the paginated deploy response.
type DeployList struct {
	Items []Deploy `json:"items"`
}

// ConfigVar represents a configuration variable with scope.
type ConfigVar struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Value     string `json:"value"`
	Secret    bool   `json:"secret"`
	Buildtime bool   `json:"buildtime"`
	Scope     string `json:"scope"`
}

// SetConfigRequest is the payload for setting a configuration variable.
type SetConfigRequest struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Secret bool   `json:"secret,omitempty"`
}

// PipelineStatus represents the pipeline status for a service.
type PipelineStatus struct {
	Build  *StageStatus `json:"build"`
	Deploy *StageStatus `json:"deploy"`
}

// StageStatus represents the status of a single pipeline stage.
type StageStatus struct {
	Status string `json:"status"`
}

// ScaleRequest is the payload for scaling service processes.
type ScaleRequest struct {
	ProcessCounts map[string]int `json:"process_counts"`
}

// CreateWorkspaceRequest is the payload for creating a workspace.
type CreateWorkspaceRequest struct {
	Name string `json:"name"`
}

// UpdateWorkspaceRequest is the payload for updating a workspace.
type UpdateWorkspaceRequest struct {
	Name string `json:"name"`
}

// CreateProjectRequest is the payload for creating a project.
type CreateProjectRequest struct {
	Name string `json:"name"`
}

// UpdateProjectRequest is the payload for updating a project.
type UpdateProjectRequest struct {
	Name string `json:"name"`
}

// CreateEnvironmentRequest is the payload for creating an environment.
type CreateEnvironmentRequest struct {
	Name string `json:"name"`
}

// CreateServiceRequest is the payload for creating a service.
type CreateServiceRequest struct {
	Name     string `json:"name"`
	Platform string `json:"platform"`
}

// UpdateServiceOptions holds optional fields for updating a service.
type UpdateServiceOptions struct {
	Name             *string `json:"name,omitempty"`
	GithubRepository *string `json:"github_repository,omitempty"`
	AutoDeployBranch *string `json:"auto_deploy_branch,omitempty"`
}
