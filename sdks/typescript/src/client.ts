import {
  AnclaError,
  AuthenticationError,
  NotFoundError,
  ValidationError,
  ServerError,
} from "./errors.js";
import type {
  AnclaClientOptions,
  ApiErrorBody,
  Build,
  BuildList,
  ConfigVar,
  Deploy,
  DeployList,
  DeployLog,
  DeployResult,
  Environment,
  PipelineStatus,
  Project,
  ProjectDetail,
  Service,
  ServiceDetail,
  SetConfigOptions,
  UpdateServiceOptions,
  Workspace,
  WorkspaceDetail,
} from "./types.js";

const DEFAULT_SERVER = "https://ancla.dev";

/**
 * Client for the Ancla PaaS REST API.
 *
 * Uses native `fetch` with no external HTTP dependencies.
 * All requests are authenticated via the `X-API-Key` header.
 */
export class AnclaClient {
  private readonly server: string;
  private readonly apiKey: string;

  constructor(options: AnclaClientOptions = {}) {
    this.server = (options.server ?? DEFAULT_SERVER).replace(/\/+$/, "");
    this.apiKey = options.apiKey ?? this.readEnvKey();
  }

  // ---------------------------------------------------------------------------
  // Workspaces
  // ---------------------------------------------------------------------------

  /** List all workspaces the authenticated user belongs to. */
  async listWorkspaces(): Promise<Workspace[]> {
    return this.request<Workspace[]>("GET", "/workspaces/");
  }

  /** Get detailed information about a workspace. */
  async getWorkspace(slug: string): Promise<WorkspaceDetail> {
    return this.request<WorkspaceDetail>("GET", `/workspaces/${slug}`);
  }

  /** Create a new workspace. */
  async createWorkspace(name: string): Promise<Workspace> {
    return this.request<Workspace>("POST", "/workspaces/", { name });
  }

  /** Update a workspace's name. */
  async updateWorkspace(slug: string, name: string): Promise<Workspace> {
    return this.request<Workspace>("PATCH", `/workspaces/${slug}`, { name });
  }

  /** Delete a workspace. */
  async deleteWorkspace(slug: string): Promise<void> {
    await this.request("DELETE", `/workspaces/${slug}`);
  }

  // ---------------------------------------------------------------------------
  // Projects
  // ---------------------------------------------------------------------------

  /** List all projects within a workspace. */
  async listProjects(ws: string): Promise<Project[]> {
    return this.request<Project[]>("GET", `/workspaces/${ws}/projects/`);
  }

  /** Get detailed information about a project. */
  async getProject(ws: string, slug: string): Promise<ProjectDetail> {
    return this.request<ProjectDetail>(
      "GET",
      `/workspaces/${ws}/projects/${slug}`,
    );
  }

  /** Create a new project within a workspace. */
  async createProject(ws: string, name: string): Promise<Project> {
    return this.request<Project>("POST", `/workspaces/${ws}/projects/`, {
      name,
    });
  }

  /** Update a project's name. */
  async updateProject(
    ws: string,
    slug: string,
    name: string,
  ): Promise<Project> {
    return this.request<Project>(
      "PATCH",
      `/workspaces/${ws}/projects/${slug}`,
      { name },
    );
  }

  /** Delete a project. */
  async deleteProject(ws: string, slug: string): Promise<void> {
    await this.request("DELETE", `/workspaces/${ws}/projects/${slug}`);
  }

  // ---------------------------------------------------------------------------
  // Environments
  // ---------------------------------------------------------------------------

  /** List environments within a project. */
  async listEnvironments(ws: string, proj: string): Promise<Environment[]> {
    return this.request<Environment[]>(
      "GET",
      `/workspaces/${ws}/projects/${proj}/envs/`,
    );
  }

  /** Get a single environment. */
  async getEnvironment(
    ws: string,
    proj: string,
    slug: string,
  ): Promise<Environment> {
    return this.request<Environment>(
      "GET",
      `/workspaces/${ws}/projects/${proj}/envs/${slug}`,
    );
  }

  /** Create a new environment within a project. */
  async createEnvironment(
    ws: string,
    proj: string,
    name: string,
  ): Promise<Environment> {
    return this.request<Environment>(
      "POST",
      `/workspaces/${ws}/projects/${proj}/envs/`,
      { name },
    );
  }

  // ---------------------------------------------------------------------------
  // Services
  // ---------------------------------------------------------------------------

  /** Build the service base path. */
  private servicePath(ws: string, proj: string, env: string): string {
    return `/workspaces/${ws}/projects/${proj}/envs/${env}/services`;
  }

  /** List services in an environment. */
  async listServices(
    ws: string,
    proj: string,
    env: string,
  ): Promise<Service[]> {
    return this.request<Service[]>(
      "GET",
      `${this.servicePath(ws, proj, env)}/`,
    );
  }

  /** Get detailed service information. */
  async getService(
    ws: string,
    proj: string,
    env: string,
    slug: string,
  ): Promise<ServiceDetail> {
    return this.request<ServiceDetail>(
      "GET",
      `${this.servicePath(ws, proj, env)}/${slug}`,
    );
  }

  /** Create a new service. */
  async createService(
    ws: string,
    proj: string,
    env: string,
    name: string,
    platform: string,
  ): Promise<Service> {
    return this.request<Service>(
      "POST",
      `${this.servicePath(ws, proj, env)}/`,
      { name, platform },
    );
  }

  /** Update a service. */
  async updateService(
    ws: string,
    proj: string,
    env: string,
    slug: string,
    opts: UpdateServiceOptions,
  ): Promise<ServiceDetail> {
    return this.request<ServiceDetail>(
      "PATCH",
      `${this.servicePath(ws, proj, env)}/${slug}`,
      opts,
    );
  }

  /** Delete a service. */
  async deleteService(
    ws: string,
    proj: string,
    env: string,
    slug: string,
  ): Promise<void> {
    await this.request("DELETE", `${this.servicePath(ws, proj, env)}/${slug}`);
  }

  /** Trigger a full deploy for a service. */
  async deployService(
    ws: string,
    proj: string,
    env: string,
    slug: string,
  ): Promise<DeployResult> {
    return this.request<DeployResult>(
      "POST",
      `${this.servicePath(ws, proj, env)}/${slug}/deploy`,
    );
  }

  /** Scale service processes. */
  async scaleService(
    ws: string,
    proj: string,
    env: string,
    slug: string,
    counts: Record<string, number>,
  ): Promise<void> {
    await this.request(
      "POST",
      `${this.servicePath(ws, proj, env)}/${slug}/scale`,
      {
        process_counts: counts,
      },
    );
  }

  /** Get pipeline status for a service. */
  async getServiceStatus(
    ws: string,
    proj: string,
    env: string,
    slug: string,
  ): Promise<PipelineStatus> {
    return this.request<PipelineStatus>(
      "GET",
      `${this.servicePath(ws, proj, env)}/${slug}/pipeline-status`,
    );
  }

  // ---------------------------------------------------------------------------
  // Builds
  // ---------------------------------------------------------------------------

  /** List builds for a service. */
  async listBuilds(
    ws: string,
    proj: string,
    env: string,
    svc: string,
  ): Promise<BuildList> {
    return this.request<BuildList>(
      "GET",
      `${this.servicePath(ws, proj, env)}/${svc}/builds/`,
    );
  }

  /** Get a single build's log by build ID. */
  async getBuild(buildId: string): Promise<Build> {
    return this.request<Build>("GET", `/builds/${buildId}/log`);
  }

  // ---------------------------------------------------------------------------
  // Deploys
  // ---------------------------------------------------------------------------

  /** List deploys for a service. */
  async listDeploys(
    ws: string,
    proj: string,
    env: string,
    svc: string,
  ): Promise<DeployList> {
    return this.request<DeployList>(
      "GET",
      `${this.servicePath(ws, proj, env)}/${svc}/deploys/`,
    );
  }

  /** Get deploy details by ID. */
  async getDeploy(deployId: string): Promise<Deploy> {
    return this.request<Deploy>("GET", `/deploys/${deployId}/detail`);
  }

  /** Get deploy log output by ID. */
  async getDeployLog(deployId: string): Promise<DeployLog> {
    return this.request<DeployLog>("GET", `/deploys/${deployId}/log`);
  }

  // ---------------------------------------------------------------------------
  // Configuration
  // ---------------------------------------------------------------------------

  /** List configuration variables for a service. */
  async listConfig(
    ws: string,
    proj: string,
    env: string,
    svc: string,
  ): Promise<ConfigVar[]> {
    return this.request<ConfigVar[]>(
      "GET",
      `${this.servicePath(ws, proj, env)}/${svc}/config/`,
    );
  }

  /** Set (create or update) a configuration variable. */
  async setConfig(
    ws: string,
    proj: string,
    env: string,
    svc: string,
    opts: SetConfigOptions,
  ): Promise<void> {
    await this.request(
      "POST",
      `${this.servicePath(ws, proj, env)}/${svc}/config/`,
      opts,
    );
  }

  /** Delete a configuration variable by ID. */
  async deleteConfig(
    ws: string,
    proj: string,
    env: string,
    svc: string,
    configId: string,
  ): Promise<void> {
    await this.request(
      "DELETE",
      `${this.servicePath(ws, proj, env)}/${svc}/config/${configId}`,
    );
  }

  // ---------------------------------------------------------------------------
  // Internal helpers
  // ---------------------------------------------------------------------------

  /** Build the full API URL for a given path. */
  private url(path: string): string {
    return `${this.server}/api/v1${path}`;
  }

  /** Read the API key from the environment, if available. */
  private readEnvKey(): string {
    try {
      // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
      const env = (globalThis as Record<string, unknown>).process as
        | { env?: Record<string, string | undefined> }
        | undefined;
      return env?.env?.ANCLA_API_KEY ?? "";
    } catch {
      return "";
    }
  }

  /**
   * Execute an HTTP request against the Ancla API.
   *
   * Automatically sets the `X-API-Key` and `Content-Type` headers,
   * parses JSON responses, and maps error status codes to typed errors.
   */
  private async request<T = unknown>(
    method: string,
    path: string,
    body?: unknown,
  ): Promise<T> {
    const headers: Record<string, string> = {};
    if (this.apiKey) {
      headers["X-API-Key"] = this.apiKey;
    }

    let requestBody: string | undefined;
    if (body !== undefined) {
      headers["Content-Type"] = "application/json";
      requestBody = JSON.stringify(body);
    }

    const response = await fetch(this.url(path), {
      method,
      headers,
      body: requestBody,
    });

    const responseText = await response.text();

    if (!response.ok) {
      this.throwForStatus(response.status, responseText);
    }

    if (!responseText) {
      return undefined as T;
    }

    return JSON.parse(responseText) as T;
  }

  /**
   * Map an HTTP error status code to a typed SDK error.
   * Attempts to extract a human-readable message from the API error body.
   */
  private throwForStatus(status: number, body: string): never {
    let message: string | undefined;
    try {
      const parsed = JSON.parse(body) as ApiErrorBody;
      message = parsed.message || parsed.detail;
    } catch {
      // body is not JSON -- use a default message
    }

    switch (status) {
      case 401:
        throw new AuthenticationError(message ?? "Not authenticated", body);
      case 404:
        throw new NotFoundError(message ?? "Not found", body);
      case 422:
        throw new ValidationError(message ?? "Validation error", body);
      default:
        if (status >= 500) {
          throw new ServerError(message ?? "Server error", body);
        }
        throw new AnclaError(
          message ?? `Request failed (${status})`,
          status,
          body,
        );
    }
  }
}
