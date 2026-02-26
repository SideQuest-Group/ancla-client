import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { AnclaClient } from "../src/client.js";
import {
  AuthenticationError,
  NotFoundError,
  ServerError,
  AnclaError,
} from "../src/errors.js";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** Create a mock Response from a status and JSON body. */
function mockResponse(status: number, body: unknown): Response {
  const text = typeof body === "string" ? body : JSON.stringify(body);
  return {
    ok: status >= 200 && status < 300,
    status,
    text: () => Promise.resolve(text),
  } as Response;
}

// ---------------------------------------------------------------------------
// Setup
// ---------------------------------------------------------------------------

let client: AnclaClient;
let fetchMock: ReturnType<typeof vi.fn>;

beforeEach(() => {
  fetchMock = vi.fn();
  vi.stubGlobal("fetch", fetchMock);
  client = new AnclaClient({
    apiKey: "test-key",
    server: "https://test.ancla.dev",
  });
});

afterEach(() => {
  vi.restoreAllMocks();
});

// ---------------------------------------------------------------------------
// Client initialization
// ---------------------------------------------------------------------------

describe("AnclaClient initialization", () => {
  it("uses the provided server and api key", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, []));
    await client.listWorkspaces();

    const [url, opts] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe("https://test.ancla.dev/api/v1/workspaces/");
    expect((opts.headers as Record<string, string>)["X-API-Key"]).toBe(
      "test-key",
    );
  });

  it("defaults to https://ancla.dev when no server is provided", async () => {
    const defaultClient = new AnclaClient({ apiKey: "k" });
    fetchMock.mockResolvedValueOnce(mockResponse(200, []));
    await defaultClient.listWorkspaces();

    const [url] = fetchMock.mock.calls[0] as [string];
    expect(url).toBe("https://ancla.dev/api/v1/workspaces/");
  });

  it("strips trailing slashes from server URL", async () => {
    const slashClient = new AnclaClient({
      apiKey: "k",
      server: "https://ancla.dev///",
    });
    fetchMock.mockResolvedValueOnce(mockResponse(200, []));
    await slashClient.listWorkspaces();

    const [url] = fetchMock.mock.calls[0] as [string];
    expect(url).toBe("https://ancla.dev/api/v1/workspaces/");
  });

  it("reads ANCLA_API_KEY from environment when no key is given", async () => {
    vi.stubEnv("ANCLA_API_KEY", "env-key");
    const envClient = new AnclaClient({ server: "https://test.ancla.dev" });
    fetchMock.mockResolvedValueOnce(mockResponse(200, []));
    await envClient.listWorkspaces();

    const [, opts] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect((opts.headers as Record<string, string>)["X-API-Key"]).toBe(
      "env-key",
    );
    vi.unstubAllEnvs();
  });
});

// ---------------------------------------------------------------------------
// Workspaces CRUD
// ---------------------------------------------------------------------------

describe("Workspaces", () => {
  it("listWorkspaces returns an array of workspaces", async () => {
    const payload = [
      {
        id: "1",
        name: "Acme",
        slug: "acme",
        member_count: 3,
        project_count: 2,
        service_count: 5,
      },
    ];
    fetchMock.mockResolvedValueOnce(mockResponse(200, payload));

    const workspaces = await client.listWorkspaces();
    expect(workspaces).toEqual(payload);
  });

  it("getWorkspace returns workspace details with members", async () => {
    const payload = {
      id: "1",
      name: "Acme",
      slug: "acme",
      member_count: 3,
      project_count: 2,
      service_count: 5,
      members: [{ username: "alice", email: "alice@acme.co", admin: true }],
    };
    fetchMock.mockResolvedValueOnce(mockResponse(200, payload));

    const ws = await client.getWorkspace("acme");
    expect(ws.slug).toBe("acme");
    expect(ws.members).toHaveLength(1);
    expect(ws.members[0].admin).toBe(true);

    const [url] = fetchMock.mock.calls[0] as [string];
    expect(url).toBe("https://test.ancla.dev/api/v1/workspaces/acme");
  });

  it("createWorkspace sends a POST with the name", async () => {
    const payload = {
      id: "2",
      name: "NewWs",
      slug: "newws",
      member_count: 1,
      project_count: 0,
      service_count: 0,
    };
    fetchMock.mockResolvedValueOnce(mockResponse(201, payload));

    const ws = await client.createWorkspace("NewWs");
    expect(ws.slug).toBe("newws");

    const [url, opts] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe("https://test.ancla.dev/api/v1/workspaces/");
    expect(opts.method).toBe("POST");
    expect(JSON.parse(opts.body as string)).toEqual({ name: "NewWs" });
  });

  it("updateWorkspace sends a PATCH with the new name", async () => {
    const payload = {
      id: "2",
      name: "Renamed",
      slug: "newws",
      member_count: 1,
      project_count: 0,
      service_count: 0,
    };
    fetchMock.mockResolvedValueOnce(mockResponse(200, payload));

    const ws = await client.updateWorkspace("newws", "Renamed");
    expect(ws.name).toBe("Renamed");

    const [, opts] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(opts.method).toBe("PATCH");
  });

  it("deleteWorkspace sends a DELETE request", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(204, ""));

    await client.deleteWorkspace("newws");

    const [url, opts] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe("https://test.ancla.dev/api/v1/workspaces/newws");
    expect(opts.method).toBe("DELETE");
  });
});

// ---------------------------------------------------------------------------
// Projects
// ---------------------------------------------------------------------------

describe("Projects", () => {
  it("listProjects fetches projects within a workspace", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, []));
    await client.listProjects("acme");

    const [url] = fetchMock.mock.calls[0] as [string];
    expect(url).toBe("https://test.ancla.dev/api/v1/workspaces/acme/projects/");
  });

  it("getProject builds correct URL path", async () => {
    const payload = {
      id: "p1",
      name: "MyProject",
      slug: "myproj",
      workspace_slug: "acme",
      workspace_name: "Acme",
      service_count: 3,
      created: "2025-01-01",
      updated: "2025-06-01",
    };
    fetchMock.mockResolvedValueOnce(mockResponse(200, payload));

    const project = await client.getProject("acme", "myproj");
    expect(project.slug).toBe("myproj");

    const [url] = fetchMock.mock.calls[0] as [string];
    expect(url).toBe(
      "https://test.ancla.dev/api/v1/workspaces/acme/projects/myproj",
    );
  });
});

// ---------------------------------------------------------------------------
// Services
// ---------------------------------------------------------------------------

describe("Services", () => {
  const ws = "acme";
  const proj = "myproj";
  const env = "production";
  const basePath = `https://test.ancla.dev/api/v1/workspaces/${ws}/projects/${proj}/envs/${env}/services`;

  it("listServices builds the correct path", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, []));
    await client.listServices(ws, proj, env);

    const [url] = fetchMock.mock.calls[0] as [string];
    expect(url).toBe(`${basePath}/`);
  });

  it("deployService sends a POST to the deploy endpoint", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, { build_id: "b-123" }));

    const result = await client.deployService(ws, proj, env, "web");
    expect(result.build_id).toBe("b-123");

    const [url, opts] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe(`${basePath}/web/deploy`);
    expect(opts.method).toBe("POST");
  });

  it("scaleService sends process counts in the body", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, ""));

    await client.scaleService(ws, proj, env, "web", { web: 2, worker: 1 });

    const [, opts] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(JSON.parse(opts.body as string)).toEqual({
      process_counts: { web: 2, worker: 1 },
    });
  });
});

// ---------------------------------------------------------------------------
// Config
// ---------------------------------------------------------------------------

describe("Config", () => {
  const ws = "acme";
  const proj = "myproj";
  const env = "production";
  const svc = "web";
  const basePath = `https://test.ancla.dev/api/v1/workspaces/${ws}/projects/${proj}/envs/${env}/services/${svc}/config`;

  it("listConfig fetches configs for a service", async () => {
    const payload = [
      {
        id: "c1",
        name: "DB_URL",
        value: "postgres://...",
        secret: false,
        buildtime: false,
      },
    ];
    fetchMock.mockResolvedValueOnce(mockResponse(200, payload));

    const configs = await client.listConfig(ws, proj, env, svc);
    expect(configs).toHaveLength(1);
    expect(configs[0].name).toBe("DB_URL");
  });

  it("setConfig sends name, value, and options", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, ""));

    await client.setConfig(ws, proj, env, svc, {
      name: "SECRET_KEY",
      value: "s3cret",
      secret: true,
    });

    const [, opts] = fetchMock.mock.calls[0] as [string, RequestInit];
    const body = JSON.parse(opts.body as string);
    expect(body).toEqual({
      name: "SECRET_KEY",
      value: "s3cret",
      secret: true,
    });
  });

  it("deleteConfig sends a DELETE with correct path", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(204, ""));

    await client.deleteConfig(ws, proj, env, svc, "config-id");

    const [url, opts] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe(`${basePath}/config-id`);
    expect(opts.method).toBe("DELETE");
  });
});

// ---------------------------------------------------------------------------
// Builds
// ---------------------------------------------------------------------------

describe("Builds", () => {
  it("listBuilds returns paginated build list", async () => {
    const payload = {
      items: [
        {
          id: "b1",
          version: 1,
          built: true,
          error: false,
          created: "2025-01-01",
        },
      ],
    };
    fetchMock.mockResolvedValueOnce(mockResponse(200, payload));

    const result = await client.listBuilds(
      "acme",
      "myproj",
      "production",
      "web",
    );
    expect(result.items).toHaveLength(1);
    expect(result.items[0].built).toBe(true);
  });
});

// ---------------------------------------------------------------------------
// Deploys
// ---------------------------------------------------------------------------

describe("Deploys", () => {
  it("getDeploy returns deploy details", async () => {
    const payload = {
      id: "d1",
      complete: true,
      error: false,
      error_detail: "",
      job_id: "j1",
      created: "2025-01-01",
      updated: "2025-01-01",
    };
    fetchMock.mockResolvedValueOnce(mockResponse(200, payload));

    const dep = await client.getDeploy("d1");
    expect(dep.complete).toBe(true);

    const [url] = fetchMock.mock.calls[0] as [string];
    expect(url).toBe("https://test.ancla.dev/api/v1/deploys/d1/detail");
  });
});

// ---------------------------------------------------------------------------
// Error handling
// ---------------------------------------------------------------------------

describe("Error handling", () => {
  it("throws AuthenticationError on 401", async () => {
    fetchMock.mockResolvedValueOnce(
      mockResponse(401, { message: "Invalid key" }),
    );

    await expect(client.listWorkspaces()).rejects.toThrow(AuthenticationError);
    await expect(
      (async () => {
        fetchMock.mockResolvedValueOnce(
          mockResponse(401, { message: "Invalid key" }),
        );
        return client.listWorkspaces();
      })(),
    ).rejects.toThrow("Invalid key");
  });

  it("throws NotFoundError on 404", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(404, {}));
    await expect(client.getWorkspace("nope")).rejects.toThrow(NotFoundError);
  });

  it("throws ServerError on 500", async () => {
    fetchMock.mockResolvedValueOnce(
      mockResponse(500, { message: "Internal server error" }),
    );
    await expect(client.listWorkspaces()).rejects.toThrow(ServerError);
  });

  it("throws AnclaError on other 4xx codes with custom message", async () => {
    fetchMock.mockResolvedValueOnce(
      mockResponse(403, { message: "Forbidden" }),
    );
    await expect(client.listWorkspaces()).rejects.toThrow(AnclaError);
  });

  it("includes the response body in the error", async () => {
    fetchMock.mockResolvedValueOnce(
      mockResponse(404, { message: "Workspace not found" }),
    );

    try {
      await client.getWorkspace("missing");
      expect.fail("Should have thrown");
    } catch (err) {
      expect(err).toBeInstanceOf(NotFoundError);
      expect((err as NotFoundError).body).toContain("Workspace not found");
      expect((err as NotFoundError).status).toBe(404);
    }
  });

  it("handles non-JSON error bodies gracefully", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(502, "Bad Gateway"));
    await expect(client.listWorkspaces()).rejects.toThrow(ServerError);
  });
});
