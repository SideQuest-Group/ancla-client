/** Configuration options for the AnclaClient. */
export interface AnclaClientOptions {
  /** API key for authentication. Falls back to ANCLA_API_KEY env var. */
  apiKey?: string;
  /** Ancla server URL. Defaults to https://ancla.dev */
  server?: string;
}

/** Workspace resource (formerly Organization). */
export interface Workspace {
  id: string;
  name: string;
  slug: string;
  member_count: number;
  project_count: number;
  service_count: number;
}

/** Detailed workspace with members. */
export interface WorkspaceDetail extends Workspace {
  members: WorkspaceMember[];
}

/** Workspace member. */
export interface WorkspaceMember {
  username: string;
  email: string;
  admin: boolean;
}

/** Project resource. */
export interface Project {
  id: string;
  name: string;
  slug: string;
  workspace_slug: string;
  service_count: number;
}

/** Detailed project resource. */
export interface ProjectDetail extends Project {
  workspace_name: string;
  created: string;
  updated: string;
}

/** Environment resource. */
export interface Environment {
  id: string;
  name: string;
  slug: string;
  service_count: number;
  created: string;
}

/** Service resource (formerly Application). */
export interface Service {
  id: string;
  name: string;
  slug: string;
  platform: string;
}

/** Detailed service resource. */
export interface ServiceDetail extends Service {
  github_repository: string;
  auto_deploy_branch: string;
  process_counts: Record<string, number>;
}

/** Options for updating a service. */
export interface UpdateServiceOptions {
  name?: string;
  github_repository?: string;
  auto_deploy_branch?: string;
}

/** Deploy result returned after triggering a deploy. */
export interface DeployResult {
  build_id: string;
}

/** Scale result is empty on success. */
export type ScaleResult = void;

/** Configuration variable. */
export interface ConfigVar {
  id: string;
  name: string;
  value: string;
  secret: boolean;
  buildtime: boolean;
}

/** Options for setting a config variable. */
export interface SetConfigOptions {
  name: string;
  value: string;
  secret?: boolean;
}

/** Build resource (formerly Image). */
export interface Build {
  id: string;
  version: number;
  built: boolean;
  error: boolean;
  created: string;
}

/** Paginated build list response. */
export interface BuildList {
  items: Build[];
}

/** Build creation result. */
export interface BuildResult {
  build_id: string;
  version: number;
}

/** Deploy resource (collapsed from Release + Deployment). */
export interface Deploy {
  id: string;
  complete: boolean;
  error: boolean;
  error_detail: string;
  job_id: string;
  created: string;
  updated: string;
}

/** Paginated deploy list response. */
export interface DeployList {
  items: Deploy[];
}

/** Deploy log output. */
export interface DeployLog {
  status: string;
  log_text: string;
}

/** Pipeline status for a service. */
export interface PipelineStatus {
  build: { status: string } | null;
  deploy: { status: string } | null;
}

/** API error response body shape. */
export interface ApiErrorBody {
  status: number;
  message?: string;
  detail?: string;
}
