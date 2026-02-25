package ancla

// Org represents an organization on the Ancla platform.
type Org struct {
	ID               string      `json:"id"`
	Name             string      `json:"name"`
	Slug             string      `json:"slug"`
	MemberCount      int         `json:"member_count"`
	ProjectCount     int         `json:"project_count"`
	ApplicationCount int         `json:"application_count"`
	Members          []OrgMember `json:"members,omitempty"`
}

// OrgMember represents a member within an organization.
type OrgMember struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Admin    bool   `json:"admin"`
}

// Project represents a project within an organization.
type Project struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Slug             string `json:"slug"`
	OrganizationSlug string `json:"organization_slug"`
	OrganizationName string `json:"organization_name"`
	ApplicationCount int    `json:"application_count"`
	Created          string `json:"created"`
	Updated          string `json:"updated"`
}

// App represents an application within a project.
type App struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	Slug             string         `json:"slug"`
	Platform         string         `json:"platform"`
	GithubRepository string         `json:"github_repository,omitempty"`
	AutoDeployBranch string         `json:"auto_deploy_branch,omitempty"`
	ProcessCounts    map[string]int `json:"process_counts,omitempty"`
}

// DeployResult is the response from triggering a deploy.
type DeployResult struct {
	ImageID string `json:"image_id"`
}

// ScaleRequest is the payload for scaling application processes.
type ScaleRequest struct {
	ProcessCounts map[string]int `json:"process_counts"`
}

// PipelineStatus represents the pipeline status for an application.
type PipelineStatus struct {
	Build   *StageStatus `json:"build"`
	Release *StageStatus `json:"release"`
	Deploy  *StageStatus `json:"deploy"`
}

// StageStatus represents the status of a single pipeline stage.
type StageStatus struct {
	Status string `json:"status"`
}

// ConfigVar represents a configuration variable for an application.
type ConfigVar struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Value     string `json:"value"`
	Secret    bool   `json:"secret"`
	Buildtime bool   `json:"buildtime"`
}

// SetConfigRequest is the payload for setting a configuration variable.
type SetConfigRequest struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Secret bool   `json:"secret,omitempty"`
}

// Image represents a container image for an application.
type Image struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
	Built   bool   `json:"built"`
	Error   bool   `json:"error"`
	Created string `json:"created"`
}

// ImageList wraps the paginated image response.
type ImageList struct {
	Items []Image `json:"items"`
}

// BuildResult is the response from triggering an image build.
type BuildResult struct {
	ImageID string `json:"image_id"`
	Version int    `json:"version"`
}

// ImageLog contains build log information for an image.
type ImageLog struct {
	Status  string `json:"status"`
	Version int    `json:"version"`
	LogText string `json:"log_text"`
}

// Release represents a release for an application.
type Release struct {
	ID       string `json:"id"`
	Version  int    `json:"version"`
	Platform string `json:"platform"`
	Built    bool   `json:"built"`
	Error    bool   `json:"error"`
	Created  string `json:"created"`
}

// ReleaseList wraps the paginated release response.
type ReleaseList struct {
	Items []Release `json:"items"`
}

// CreateReleaseResult is the response from creating a release.
type CreateReleaseResult struct {
	ReleaseID string `json:"release_id"`
	Version   int    `json:"version"`
}

// DeployReleaseResult is the response from deploying a release.
type DeployReleaseResult struct {
	DeploymentID string `json:"deployment_id"`
}

// Deployment represents a deployment.
type Deployment struct {
	ID          string `json:"id"`
	Complete    bool   `json:"complete"`
	Error       bool   `json:"error"`
	ErrorDetail string `json:"error_detail"`
	JobID       string `json:"job_id"`
	Created     string `json:"created"`
	Updated     string `json:"updated"`
}

// DeploymentLog contains log information for a deployment.
type DeploymentLog struct {
	Status  string `json:"status"`
	LogText string `json:"log_text"`
}

// UpdateAppOptions holds optional fields for updating an application.
type UpdateAppOptions struct {
	Name             *string `json:"name,omitempty"`
	GithubRepository *string `json:"github_repository,omitempty"`
	AutoDeployBranch *string `json:"auto_deploy_branch,omitempty"`
}

// CreateOrgRequest is the payload for creating an organization.
type CreateOrgRequest struct {
	Name string `json:"name"`
}

// UpdateOrgRequest is the payload for updating an organization.
type UpdateOrgRequest struct {
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

// CreateAppRequest is the payload for creating an application.
type CreateAppRequest struct {
	Name     string `json:"name"`
	Platform string `json:"platform"`
}
