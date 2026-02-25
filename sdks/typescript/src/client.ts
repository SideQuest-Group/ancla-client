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
  App,
  AppDetail,
  ConfigVar,
  CreateReleaseResult,
  Deployment,
  DeployResult,
  Image,
  ImageList,
  Org,
  OrgDetail,
  PipelineStatus,
  Project,
  ProjectDetail,
  Release,
  ReleaseList,
  SetConfigOptions,
  UpdateAppOptions,
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
  // Organizations
  // ---------------------------------------------------------------------------

  /** List all organizations the authenticated user belongs to. */
  async listOrgs(): Promise<Org[]> {
    return this.request<Org[]>("GET", "/organizations/");
  }

  /** Get detailed information about an organization. */
  async getOrg(slug: string): Promise<OrgDetail> {
    return this.request<OrgDetail>("GET", `/organizations/${slug}`);
  }

  /** Create a new organization. */
  async createOrg(name: string): Promise<Org> {
    return this.request<Org>("POST", "/organizations/", { name });
  }

  /** Update an organization's name. */
  async updateOrg(slug: string, name: string): Promise<Org> {
    return this.request<Org>("PATCH", `/organizations/${slug}`, { name });
  }

  /** Delete an organization. */
  async deleteOrg(slug: string): Promise<void> {
    await this.request("DELETE", `/organizations/${slug}`);
  }

  // ---------------------------------------------------------------------------
  // Projects
  // ---------------------------------------------------------------------------

  /** List all projects, optionally filtered by organization slug. */
  async listProjects(org?: string): Promise<Project[]> {
    const path = org ? `/projects/${org}` : "/projects/";
    return this.request<Project[]>("GET", path);
  }

  /** Get detailed information about a project. */
  async getProject(org: string, slug: string): Promise<ProjectDetail> {
    return this.request<ProjectDetail>("GET", `/projects/${org}/${slug}`);
  }

  /** Create a new project within an organization. */
  async createProject(org: string, name: string): Promise<Project> {
    return this.request<Project>("POST", `/projects/${org}`, { name });
  }

  /** Update a project's name. */
  async updateProject(
    org: string,
    slug: string,
    name: string,
  ): Promise<Project> {
    return this.request<Project>("PATCH", `/projects/${org}/${slug}`, { name });
  }

  /** Delete a project. */
  async deleteProject(org: string, slug: string): Promise<void> {
    await this.request("DELETE", `/projects/${org}/${slug}`);
  }

  // ---------------------------------------------------------------------------
  // Applications
  // ---------------------------------------------------------------------------

  /** List applications in a project. */
  async listApps(org: string, project: string): Promise<App[]> {
    return this.request<App[]>("GET", `/applications/${org}/${project}`);
  }

  /** Get detailed application information. */
  async getApp(
    org: string,
    project: string,
    slug: string,
  ): Promise<AppDetail> {
    return this.request<AppDetail>(
      "GET",
      `/applications/${org}/${project}/${slug}`,
    );
  }

  /** Create a new application. */
  async createApp(
    org: string,
    project: string,
    name: string,
    platform: string,
  ): Promise<App> {
    return this.request<App>("POST", `/applications/${org}/${project}`, {
      name,
      platform,
    });
  }

  /** Update an application. */
  async updateApp(
    org: string,
    project: string,
    slug: string,
    opts: UpdateAppOptions,
  ): Promise<AppDetail> {
    return this.request<AppDetail>(
      "PATCH",
      `/applications/${org}/${project}/${slug}`,
      opts,
    );
  }

  /** Delete an application. */
  async deleteApp(
    org: string,
    project: string,
    slug: string,
  ): Promise<void> {
    await this.request("DELETE", `/applications/${org}/${project}/${slug}`);
  }

  /** Trigger a full deploy for an application (by app ID). */
  async deployApp(appId: string): Promise<DeployResult> {
    return this.request<DeployResult>(
      "POST",
      `/applications/${appId}/deploy`,
    );
  }

  /** Scale application processes (by app ID). */
  async scaleApp(
    appId: string,
    counts: Record<string, number>,
  ): Promise<void> {
    await this.request("POST", `/applications/${appId}/scale`, {
      process_counts: counts,
    });
  }

  /** Get pipeline status for an application (by app ID). */
  async getAppStatus(appId: string): Promise<PipelineStatus> {
    return this.request<PipelineStatus>(
      "GET",
      `/applications/${appId}/pipeline-status`,
    );
  }

  // ---------------------------------------------------------------------------
  // Configuration
  // ---------------------------------------------------------------------------

  /** List configuration variables for an application (by app ID). */
  async listConfig(appId: string): Promise<ConfigVar[]> {
    return this.request<ConfigVar[]>("GET", `/configurations/${appId}`);
  }

  /** Get a single configuration variable by ID. */
  async getConfig(appId: string, configId: string): Promise<ConfigVar> {
    return this.request<ConfigVar>(
      "GET",
      `/configurations/${appId}/${configId}`,
    );
  }

  /** Set (create or update) a configuration variable. */
  async setConfig(
    appId: string,
    key: string,
    value: string,
    opts?: SetConfigOptions,
  ): Promise<void> {
    await this.request("POST", `/configurations/${appId}`, {
      name: key,
      value,
      ...opts,
    });
  }

  /** Delete a configuration variable. */
  async deleteConfig(appId: string, configId: string): Promise<void> {
    await this.request("DELETE", `/configurations/${appId}/${configId}`);
  }

  // ---------------------------------------------------------------------------
  // Images
  // ---------------------------------------------------------------------------

  /** List images for an application (by app ID). */
  async listImages(appId: string): Promise<ImageList> {
    return this.request<ImageList>("GET", `/images/${appId}`);
  }

  /** Get a single image by ID. */
  async getImage(imageId: string): Promise<Image> {
    return this.request<Image>("GET", `/images/${imageId}/log`);
  }

  // ---------------------------------------------------------------------------
  // Releases
  // ---------------------------------------------------------------------------

  /** List releases for an application (by app ID). */
  async listReleases(appId: string): Promise<ReleaseList> {
    return this.request<ReleaseList>("GET", `/releases/${appId}`);
  }

  /** Get a single release by ID. */
  async getRelease(releaseId: string): Promise<Release> {
    return this.request<Release>("GET", `/releases/${releaseId}/detail`);
  }

  /** Create a new release for an application (by app ID). */
  async createRelease(appId: string): Promise<CreateReleaseResult> {
    return this.request<CreateReleaseResult>(
      "POST",
      `/releases/${appId}/create`,
    );
  }

  // ---------------------------------------------------------------------------
  // Deployments
  // ---------------------------------------------------------------------------

  /** Get deployment details by ID. */
  async getDeployment(deploymentId: string): Promise<Deployment> {
    return this.request<Deployment>(
      "GET",
      `/deployments/${deploymentId}/detail`,
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
        throw new AuthenticationError(
          message ?? "Not authenticated",
          body,
        );
      case 404:
        throw new NotFoundError(message ?? "Not found", body);
      case 422:
        throw new ValidationError(message ?? "Validation error", body);
      default:
        if (status >= 500) {
          throw new ServerError(
            message ?? "Server error",
            body,
          );
        }
        throw new AnclaError(
          message ?? `Request failed (${status})`,
          status,
          body,
        );
    }
  }
}
