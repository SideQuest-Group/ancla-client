/** Configuration options for the AnclaClient. */
export interface AnclaClientOptions {
  /** API key for authentication. Falls back to ANCLA_API_KEY env var. */
  apiKey?: string;
  /** Ancla server URL. Defaults to https://ancla.dev */
  server?: string;
}

/** Organization resource. */
export interface Org {
  id: string;
  name: string;
  slug: string;
  member_count: number;
  project_count: number;
}

/** Detailed organization with members. */
export interface OrgDetail {
  name: string;
  slug: string;
  project_count: number;
  application_count: number;
  members: OrgMember[];
}

/** Organization member. */
export interface OrgMember {
  username: string;
  email: string;
  admin: boolean;
}

/** Project resource. */
export interface Project {
  id: string;
  name: string;
  slug: string;
  organization_slug: string;
  application_count: number;
}

/** Detailed project resource. */
export interface ProjectDetail {
  name: string;
  slug: string;
  organization_slug: string;
  organization_name: string;
  application_count: number;
  created: string;
  updated: string;
}

/** Application summary in list responses. */
export interface App {
  name: string;
  slug: string;
  platform: string;
}

/** Detailed application resource. */
export interface AppDetail {
  name: string;
  slug: string;
  platform: string;
  github_repository: string;
  auto_deploy_branch: string;
  process_counts: Record<string, number>;
}

/** Options for updating an application. */
export interface UpdateAppOptions {
  name?: string;
  platform?: string;
  github_repository?: string;
  auto_deploy_branch?: string;
}

/** Deploy result returned after triggering a deploy. */
export interface DeployResult {
  image_id: string;
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
  secret?: boolean;
  buildtime?: boolean;
}

/** Container image resource. */
export interface Image {
  id: string;
  version: number;
  built: boolean;
  error: boolean;
  created: string;
}

/** Paginated image list response. */
export interface ImageList {
  items: Image[];
}

/** Release resource. */
export interface Release {
  id: string;
  version: number;
  platform: string;
  built: boolean;
  error: boolean;
  created: string;
}

/** Paginated release list response. */
export interface ReleaseList {
  items: Release[];
}

/** Release creation result. */
export interface CreateReleaseResult {
  release_id: string;
  version: number;
}

/** Deployment resource. */
export interface Deployment {
  id: string;
  complete: boolean;
  error: boolean;
  error_detail: string;
  job_id: string;
  created: string;
  updated: string;
}

/** Pipeline status for an application. */
export interface PipelineStatus {
  build: { status: string } | null;
  release: { status: string } | null;
  deploy: { status: string } | null;
}

/** API error response body shape. */
export interface ApiErrorBody {
  status?: number;
  message?: string;
  detail?: string;
}
