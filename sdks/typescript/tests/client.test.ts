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
    await client.listOrgs();

    const [url, opts] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe("https://test.ancla.dev/api/v1/organizations/");
    expect((opts.headers as Record<string, string>)["X-API-Key"]).toBe(
      "test-key",
    );
  });

  it("defaults to https://ancla.dev when no server is provided", async () => {
    const defaultClient = new AnclaClient({ apiKey: "k" });
    fetchMock.mockResolvedValueOnce(mockResponse(200, []));
    await defaultClient.listOrgs();

    const [url] = fetchMock.mock.calls[0] as [string];
    expect(url).toBe("https://ancla.dev/api/v1/organizations/");
  });

  it("strips trailing slashes from server URL", async () => {
    const slashClient = new AnclaClient({
      apiKey: "k",
      server: "https://ancla.dev///",
    });
    fetchMock.mockResolvedValueOnce(mockResponse(200, []));
    await slashClient.listOrgs();

    const [url] = fetchMock.mock.calls[0] as [string];
    expect(url).toBe("https://ancla.dev/api/v1/organizations/");
  });

  it("reads ANCLA_API_KEY from environment when no key is given", async () => {
    vi.stubEnv("ANCLA_API_KEY", "env-key");
    const envClient = new AnclaClient({ server: "https://test.ancla.dev" });
    fetchMock.mockResolvedValueOnce(mockResponse(200, []));
    await envClient.listOrgs();

    const [, opts] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect((opts.headers as Record<string, string>)["X-API-Key"]).toBe(
      "env-key",
    );
    vi.unstubAllEnvs();
  });
});

// ---------------------------------------------------------------------------
// Orgs CRUD
// ---------------------------------------------------------------------------

describe("Orgs", () => {
  it("listOrgs returns an array of organizations", async () => {
    const payload = [
      {
        id: "1",
        name: "Acme",
        slug: "acme",
        member_count: 3,
        project_count: 2,
      },
    ];
    fetchMock.mockResolvedValueOnce(mockResponse(200, payload));

    const orgs = await client.listOrgs();
    expect(orgs).toEqual(payload);
  });

  it("getOrg returns organization details with members", async () => {
    const payload = {
      name: "Acme",
      slug: "acme",
      project_count: 2,
      application_count: 5,
      members: [{ username: "alice", email: "alice@acme.co", admin: true }],
    };
    fetchMock.mockResolvedValueOnce(mockResponse(200, payload));

    const org = await client.getOrg("acme");
    expect(org.slug).toBe("acme");
    expect(org.members).toHaveLength(1);
    expect(org.members[0].admin).toBe(true);

    const [url] = fetchMock.mock.calls[0] as [string];
    expect(url).toBe("https://test.ancla.dev/api/v1/organizations/acme");
  });

  it("createOrg sends a POST with the name", async () => {
    const payload = {
      id: "2",
      name: "NewOrg",
      slug: "neworg",
      member_count: 1,
      project_count: 0,
    };
    fetchMock.mockResolvedValueOnce(mockResponse(201, payload));

    const org = await client.createOrg("NewOrg");
    expect(org.slug).toBe("neworg");

    const [url, opts] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe("https://test.ancla.dev/api/v1/organizations/");
    expect(opts.method).toBe("POST");
    expect(JSON.parse(opts.body as string)).toEqual({ name: "NewOrg" });
  });

  it("updateOrg sends a PATCH with the new name", async () => {
    const payload = {
      id: "2",
      name: "Renamed",
      slug: "neworg",
      member_count: 1,
      project_count: 0,
    };
    fetchMock.mockResolvedValueOnce(mockResponse(200, payload));

    const org = await client.updateOrg("neworg", "Renamed");
    expect(org.name).toBe("Renamed");

    const [, opts] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(opts.method).toBe("PATCH");
  });

  it("deleteOrg sends a DELETE request", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(204, ""));

    await client.deleteOrg("neworg");

    const [url, opts] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe("https://test.ancla.dev/api/v1/organizations/neworg");
    expect(opts.method).toBe("DELETE");
  });
});

// ---------------------------------------------------------------------------
// Projects
// ---------------------------------------------------------------------------

describe("Projects", () => {
  it("listProjects fetches all projects", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, []));
    await client.listProjects();

    const [url] = fetchMock.mock.calls[0] as [string];
    expect(url).toBe("https://test.ancla.dev/api/v1/projects/");
  });

  it("listProjects with org filters by org slug", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, []));
    await client.listProjects("acme");

    const [url] = fetchMock.mock.calls[0] as [string];
    expect(url).toBe("https://test.ancla.dev/api/v1/projects/acme");
  });

  it("getProject builds correct URL path", async () => {
    const payload = {
      name: "MyProject",
      slug: "myproj",
      organization_slug: "acme",
      organization_name: "Acme",
      application_count: 3,
      created: "2025-01-01",
      updated: "2025-06-01",
    };
    fetchMock.mockResolvedValueOnce(mockResponse(200, payload));

    const project = await client.getProject("acme", "myproj");
    expect(project.slug).toBe("myproj");

    const [url] = fetchMock.mock.calls[0] as [string];
    expect(url).toBe("https://test.ancla.dev/api/v1/projects/acme/myproj");
  });
});

// ---------------------------------------------------------------------------
// Apps
// ---------------------------------------------------------------------------

describe("Apps", () => {
  it("listApps builds the correct path", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, []));
    await client.listApps("acme", "myproj");

    const [url] = fetchMock.mock.calls[0] as [string];
    expect(url).toBe(
      "https://test.ancla.dev/api/v1/applications/acme/myproj",
    );
  });

  it("deployApp sends a POST to the deploy endpoint", async () => {
    fetchMock.mockResolvedValueOnce(
      mockResponse(200, { image_id: "img-123" }),
    );

    const result = await client.deployApp("app-id");
    expect(result.image_id).toBe("img-123");

    const [url, opts] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe(
      "https://test.ancla.dev/api/v1/applications/app-id/deploy",
    );
    expect(opts.method).toBe("POST");
  });

  it("scaleApp sends process counts in the body", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, ""));

    await client.scaleApp("app-id", { web: 2, worker: 1 });

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
  it("listConfig fetches configs for an app", async () => {
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

    const configs = await client.listConfig("app-id");
    expect(configs).toHaveLength(1);
    expect(configs[0].name).toBe("DB_URL");
  });

  it("setConfig sends name, value, and options", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, ""));

    await client.setConfig("app-id", "SECRET_KEY", "s3cret", {
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

    await client.deleteConfig("app-id", "config-id");

    const [url, opts] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe(
      "https://test.ancla.dev/api/v1/configurations/app-id/config-id",
    );
    expect(opts.method).toBe("DELETE");
  });
});

// ---------------------------------------------------------------------------
// Images
// ---------------------------------------------------------------------------

describe("Images", () => {
  it("listImages returns paginated image list", async () => {
    const payload = {
      items: [
        { id: "i1", version: 1, built: true, error: false, created: "2025-01-01" },
      ],
    };
    fetchMock.mockResolvedValueOnce(mockResponse(200, payload));

    const result = await client.listImages("app-id");
    expect(result.items).toHaveLength(1);
    expect(result.items[0].built).toBe(true);
  });
});

// ---------------------------------------------------------------------------
// Releases
// ---------------------------------------------------------------------------

describe("Releases", () => {
  it("createRelease sends a POST and returns the result", async () => {
    fetchMock.mockResolvedValueOnce(
      mockResponse(200, { release_id: "r1", version: 3 }),
    );

    const result = await client.createRelease("app-id");
    expect(result.release_id).toBe("r1");
    expect(result.version).toBe(3);
  });
});

// ---------------------------------------------------------------------------
// Deployments
// ---------------------------------------------------------------------------

describe("Deployments", () => {
  it("getDeployment returns deployment details", async () => {
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

    const dep = await client.getDeployment("d1");
    expect(dep.complete).toBe(true);

    const [url] = fetchMock.mock.calls[0] as [string];
    expect(url).toBe(
      "https://test.ancla.dev/api/v1/deployments/d1/detail",
    );
  });
});

// ---------------------------------------------------------------------------
// Error handling
// ---------------------------------------------------------------------------

describe("Error handling", () => {
  it("throws AuthenticationError on 401", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(401, { message: "Invalid key" }));

    await expect(client.listOrgs()).rejects.toThrow(AuthenticationError);
    await expect(
      (async () => {
        fetchMock.mockResolvedValueOnce(
          mockResponse(401, { message: "Invalid key" }),
        );
        return client.listOrgs();
      })(),
    ).rejects.toThrow("Invalid key");
  });

  it("throws NotFoundError on 404", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(404, {}));
    await expect(client.getOrg("nope")).rejects.toThrow(NotFoundError);
  });

  it("throws ServerError on 500", async () => {
    fetchMock.mockResolvedValueOnce(
      mockResponse(500, { message: "Internal server error" }),
    );
    await expect(client.listOrgs()).rejects.toThrow(ServerError);
  });

  it("throws AnclaError on other 4xx codes", async () => {
    fetchMock.mockResolvedValueOnce(
      mockResponse(403, { message: "Forbidden" }),
    );
    await expect(client.listOrgs()).rejects.toThrow(AnclaError);
  });

  it("includes the response body in the error", async () => {
    fetchMock.mockResolvedValueOnce(
      mockResponse(404, { message: "Org not found" }),
    );

    try {
      await client.getOrg("missing");
      expect.fail("Should have thrown");
    } catch (err) {
      expect(err).toBeInstanceOf(NotFoundError);
      expect((err as NotFoundError).body).toContain("Org not found");
      expect((err as NotFoundError).status).toBe(404);
    }
  });

  it("handles non-JSON error bodies gracefully", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(502, "Bad Gateway"));
    await expect(client.listOrgs()).rejects.toThrow(ServerError);
  });
});
