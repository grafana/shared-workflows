import {
  afterEach,
  beforeEach,
  describe,
  expect,
  it,
  Mock,
  mock,
  spyOn,
} from "bun:test";
import {
  matchesAnyPattern,
  fetchAllAlerts,
  fetchSpecificAlertsWithPRs,
  run,
  DependabotAlert,
} from "./index";
import { Octokit } from "@octokit/rest";
import { RequestError } from "@octokit/request-error";
import * as minimatchModule from "minimatch"; // Import for spying

// Mock Octokit globally for all tests in this file
const mockOctokit = {
  request: mock(),
  paginate: mock(),
  rest: {
    pulls: {
      update: mock(),
    },
  },
};

// Mock the graphql module
const mockGraphql = mock();
await mock.module("@octokit/graphql", () => ({
  graphql: {
    defaults: mock(() => mockGraphql),
  },
}));

await mock.module("@octokit/rest", () => ({
  Octokit: mock(() => mockOctokit),
}));

describe("Dependabot Auto Triage Action", () => {
  let consoleLogSpy: Mock<typeof console.log>;
  let consoleErrorSpy: Mock<typeof console.error>;
  let processExitSpy: Mock<typeof process.exit>;

  beforeEach(() => {
    // Reset mocks and environment variables
    mockOctokit.request.mockClear();
    mockOctokit.paginate.mockClear();
    mockOctokit.rest.pulls.update.mockClear();
    mockGraphql.mockClear();
    process.env.GITHUB_TOKEN = "test-token";
    process.env.GITHUB_REPOSITORY = "owner/repo";
    process.env.INPUT_ALERT_TYPES = "dependency";
    process.env.INPUT_PATHS = "**/package-lock.json\n**/yarn.lock";
    process.env.INPUT_DISMISSAL_COMMENT = "Test dismissal comment";
    process.env.INPUT_DISMISSAL_REASON = "tolerable_risk";
    process.env.INPUT_CLOSE_PRS = "false";

    // Spy on console messages and process.exit
    consoleLogSpy = spyOn(console, "log").mockImplementation(() => {});
    consoleErrorSpy = spyOn(console, "error").mockImplementation(() => {});
    processExitSpy = spyOn(process, "exit").mockImplementation((() => {}) as (
      code?: number,
    ) => never);
  });

  afterEach(() => {
    mock.restore(); // Restores all mocks
    // Clear environment variables
    delete process.env.GITHUB_TOKEN;
    delete process.env.GITHUB_REPOSITORY;
    delete process.env.INPUT_ALERT_TYPES;
    delete process.env.INPUT_PATHS;
    delete process.env.INPUT_DISMISSAL_COMMENT;
    delete process.env.INPUT_DISMISSAL_REASON;
    delete process.env.INPUT_CLOSE_PRS;
  });

  describe("matchesAnyPattern", () => {
    let minimatchSpy: Mock<typeof minimatchModule.minimatch>;

    beforeEach(() => {
      minimatchSpy = spyOn(minimatchModule, "minimatch");
    });

    afterEach(() => {
      minimatchSpy.mockRestore();
    });

    it("should return true if manifestPath matches any pattern", () => {
      const manifestPath = "src/package-lock.json";
      const patterns = ["**/package-lock.json", "**/gemfile.lock"];
      expect(matchesAnyPattern(manifestPath, patterns)).toBe(true);
      expect(minimatchSpy).toHaveBeenCalledTimes(1); // Matches first pattern
    });

    it("should return false if manifestPath does not match any pattern", () => {
      const manifestPath = "src/some/other/file.txt";
      const patterns = ["**/package-lock.json", "**/gemfile.lock"];
      expect(matchesAnyPattern(manifestPath, patterns)).toBe(false);
      expect(minimatchSpy).toHaveBeenCalledTimes(2);
    });

    it("should return false if manifestPath is undefined", () => {
      const patterns = ["**/package-lock.json"];
      // matchesAnyPattern returns early
      expect(matchesAnyPattern(undefined, patterns)).toBe(false);
      expect(minimatchSpy).not.toHaveBeenCalled();
    });

    it("should return false if patterns array is empty", () => {
      const manifestPath = "src/package-lock.json";
      const patterns: string[] = [];
      // matchesAnyPattern returns early
      expect(matchesAnyPattern(manifestPath, patterns)).toBe(false);
      expect(minimatchSpy).not.toHaveBeenCalled();
    });

    it("should handle glob patterns correctly", () => {
      expect(matchesAnyPattern("frontend/package.json", ["frontend/**"])).toBe(
        true,
      );
      expect(matchesAnyPattern("backend/package.json", ["frontend/**"])).toBe(
        false,
      );
      expect(minimatchSpy).toHaveBeenCalledTimes(2);
    });

    it("should use matchBase option for minimatch", () => {
      expect(
        matchesAnyPattern("package-lock.json", ["package-lock.json"]),
      ).toBe(true);
      expect(
        matchesAnyPattern("sub/package-lock.json", ["package-lock.json"]),
      ).toBe(true);
      expect(minimatchSpy).toHaveBeenCalledTimes(2);
    });

    it("should log an error, handle it, and continue matching with other patterns", () => {
      const manifestPath = "src/package-lock.json";
      const badPattern = "force-error";
      const goodPattern = "**/package-lock.json";
      const patterns = [badPattern, goodPattern];

      expect(matchesAnyPattern(manifestPath, patterns)).toBe(true);
      expect(minimatchSpy).toHaveBeenCalledTimes(2);
      expect(minimatchSpy.mock.calls[0][1]).toBe(badPattern);
      expect(minimatchSpy.mock.calls[1][1]).toBe(goodPattern);
    });
  });

  describe("fetchAllAlerts", () => {
    const mockOwner = "owner";
    const mockRepo = "repo";

    const createMockAlert = (
      number: number,
      severity: string,
      packageName: string,
      manifestPath: string,
    ): DependabotAlert => ({
      number,
      dependency: {
        package: { name: packageName },
        manifest_path: manifestPath,
      },
      security_advisory: { severity },
      security_vulnerability: { severity }, // Assuming advisory and vulnerability severities are the same for simplicity
    });

    it("should fetch and filter alerts based on severity", async () => {
      const mockAlertsResponse: DependabotAlert[] = [
        createMockAlert(1, "critical", "pkg-a", "path/to/manifest1"),
        createMockAlert(2, "high", "pkg-b", "path/to/manifest2"),
        createMockAlert(3, "low", "pkg-c", "path/to/manifest3"),
      ];
      mockOctokit.paginate.mockResolvedValue(mockAlertsResponse);

      const octokitInstance = new Octokit(); // Real instance not used due to global mock
      const filteredAlerts = await fetchAllAlerts(
        octokitInstance,
        mockOwner,
        mockRepo,
        ["critical", "high"],
      );

      expect(mockOctokit.paginate).toHaveBeenCalledWith(
        "GET /repos/{owner}/{repo}/dependabot/alerts",
        {
          owner: mockOwner,
          repo: mockRepo,
          state: "open",
          per_page: 100,
        },
      );
      expect(filteredAlerts).toHaveLength(2);
      expect(filteredAlerts.find((a) => a.number === 1)).toBeDefined();
      expect(filteredAlerts.find((a) => a.number === 2)).toBeDefined();
      expect(filteredAlerts.find((a) => a.number === 3)).toBeUndefined();
    });

    it("should fetch and filter alerts for 'dependency' type if no severity matches", async () => {
      const mockAlertsResponse: DependabotAlert[] = [
        createMockAlert(1, "critical", "pkg-a", "path/to/manifest1"),
        createMockAlert(2, "high", "pkg-b", "path/to/manifest2"),
      ];
      mockOctokit.paginate.mockResolvedValue(mockAlertsResponse);
      const octokitInstance = new Octokit();
      const filteredAlerts = await fetchAllAlerts(
        octokitInstance,
        mockOwner,
        mockRepo,
        ["dependency"],
      );

      expect(filteredAlerts).toHaveLength(2);
    });

    it("should return all alerts if 'dependency' is in alertTypes along with severities", async () => {
      const mockAlertsResponse: DependabotAlert[] = [
        createMockAlert(1, "critical", "pkg-a", "path/to/manifest1"),
        createMockAlert(2, "low", "pkg-b", "path/to/manifest2"),
      ];
      mockOctokit.paginate.mockResolvedValue(mockAlertsResponse);
      const octokitInstance = new Octokit();
      const filteredAlerts = await fetchAllAlerts(
        octokitInstance,
        mockOwner,
        mockRepo,
        ["critical", "dependency"],
      );
      // The filter is OR based, so both alerts should be returned because 'dependency' matches all, and critical matches one.
      // The current implementation will include all if 'dependency' is present.
      expect(filteredAlerts).toHaveLength(2);
    });

    it("should return empty array if no alerts match severity and 'dependency' type is not included", async () => {
      const mockAlertsResponse: DependabotAlert[] = [
        createMockAlert(1, "critical", "pkg-a", "path/to/manifest1"),
      ];
      mockOctokit.paginate.mockResolvedValue(mockAlertsResponse);
      const octokitInstance = new Octokit();
      const filteredAlerts = await fetchAllAlerts(
        octokitInstance,
        mockOwner,
        mockRepo,
        ["low"],
      );

      expect(filteredAlerts).toHaveLength(0);
    });

    it("should handle empty list of alerts from paginate", async () => {
      mockOctokit.paginate.mockResolvedValue([]);
      const octokitInstance = new Octokit();
      const filteredAlerts = await fetchAllAlerts(
        octokitInstance,
        mockOwner,
        mockRepo,
        ["critical"],
      );
      expect(filteredAlerts).toHaveLength(0);
    });
  });

  describe("fetchSpecificAlertsWithPRs", () => {
    it("should fetch alerts with associated PRs using GraphQL", async () => {
      const alertNumbers = [1, 2, 3];
      const mockGraphqlResponse = {
        repository: {
          alert0: {
            number: 1,
            dependabotUpdate: {
              pullRequest: { number: 101 },
            },
          },
          alert1: {
            number: 2,
            dependabotUpdate: null,
          },
          alert2: {
            number: 3,
            dependabotUpdate: {
              pullRequest: { number: 103 },
            },
          },
        },
      };

      mockGraphql.mockResolvedValue(mockGraphqlResponse);

      const result = await fetchSpecificAlertsWithPRs(
        "test-token",
        "owner",
        "repo",
        alertNumbers,
      );

      expect(result.size).toBe(2);
      expect(result.get(1)).toBe(101);
      expect(result.get(2)).toBeUndefined();
      expect(result.get(3)).toBe(103);
      expect(mockGraphql).toHaveBeenCalledWith(
        expect.stringContaining("query GetSpecificVulnerabilityAlerts"),
        { owner: "owner", repo: "repo" },
      );
    });

    it("should return empty map for empty alert numbers", async () => {
      const result = await fetchSpecificAlertsWithPRs(
        "test-token",
        "owner",
        "repo",
        [],
      );

      expect(result.size).toBe(0);
      expect(mockGraphql).not.toHaveBeenCalled();
    });

    it("should handle null alerts in GraphQL response", async () => {
      const alertNumbers = [1, 2];
      const mockGraphqlResponse = {
        repository: {
          alert0: null,
          alert1: {
            number: 2,
            dependabotUpdate: {
              pullRequest: { number: 102 },
            },
          },
        },
      };

      mockGraphql.mockResolvedValue(mockGraphqlResponse);

      const result = await fetchSpecificAlertsWithPRs(
        "test-token",
        "owner",
        "repo",
        alertNumbers,
      );

      expect(result.size).toBe(1);
      expect(result.get(1)).toBeUndefined();
      expect(result.get(2)).toBe(102);
    });

    it("should handle alerts without pull requests", async () => {
      const alertNumbers = [1];
      const mockGraphqlResponse = {
        repository: {
          alert0: {
            number: 1,
            dependabotUpdate: null,
          },
        },
      };

      mockGraphql.mockResolvedValue(mockGraphqlResponse);

      const result = await fetchSpecificAlertsWithPRs(
        "test-token",
        "owner",
        "repo",
        alertNumbers,
      );

      expect(result.size).toBe(0);
    });
  });

  // More tests for fetchAllAlerts and run will be added here
  describe("run", () => {
    const createMockAlert = (
      number: number,
      severity: string,
      packageName: string,
      manifestPath: string,
    ): DependabotAlert => ({
      number,
      dependency: {
        package: { name: packageName },
        manifest_path: manifestPath,
      },
      security_advisory: { severity },
      security_vulnerability: { severity },
    });

    it("should successfully fetch, filter, and dismiss alerts", async () => {
      const mockAlerts: DependabotAlert[] = [
        createMockAlert(1, "high", "pkg-a", "src/package-lock.json"),
        createMockAlert(2, "critical", "pkg-b", "other/yarn.lock"),
        createMockAlert(3, "low", "pkg-c", "nomatch/package.json"),
      ];
      mockOctokit.paginate.mockResolvedValue(mockAlerts);
      mockOctokit.request.mockResolvedValue({ status: 200 }); // For dismissal

      // Mock successful repo check
      mockOctokit.request.mockImplementation((route: string) => {
        if (route === "GET /repos/{owner}/{repo}") {
          return { data: { owner: { login: "test-user" } } };
        }
        if (
          route ===
          "PATCH /repos/{owner}/{repo}/dependabot/alerts/{alert_number}"
        ) {
          return { status: 200 };
        }
        return {};
      });

      process.env.INPUT_PATHS = "src/package-lock.json\nother/yarn.lock";

      await run();

      expect(consoleLogSpy).toHaveBeenCalledWith(
        "Fetching dependabot alerts for owner/repo...",
      );
      expect(consoleLogSpy).toHaveBeenCalledWith("Using path patterns:");
      expect(consoleLogSpy).toHaveBeenCalledWith("- src/package-lock.json");
      expect(consoleLogSpy).toHaveBeenCalledWith("- other/yarn.lock");
      expect(consoleLogSpy).toHaveBeenCalledWith(
        `Found ${mockAlerts.length} open alerts`,
      );
      expect(consoleLogSpy).toHaveBeenCalledWith(
        "Alert #1 for pkg-a in src/package-lock.json matches patterns",
      );
      expect(consoleLogSpy).toHaveBeenCalledWith(
        "Alert #2 for pkg-b in other/yarn.lock matches patterns",
      );
      expect(consoleLogSpy).toHaveBeenCalledWith(
        "Skipping alert #3 for nomatch/package.json (does not match any pattern)",
      );
      expect(consoleLogSpy).toHaveBeenCalledWith("Dismissing 2 alerts...");
      expect(mockOctokit.request).toHaveBeenCalledWith(
        "PATCH /repos/{owner}/{repo}/dependabot/alerts/{alert_number}",
        expect.objectContaining({ alert_number: 1, state: "dismissed" }),
      );
      expect(mockOctokit.request).toHaveBeenCalledWith(
        "PATCH /repos/{owner}/{repo}/dependabot/alerts/{alert_number}",
        expect.objectContaining({ alert_number: 2, state: "dismissed" }),
      );
      expect(consoleLogSpy).toHaveBeenCalledWith(
        "Alert #1 dismissed successfully",
      );
      expect(consoleLogSpy).toHaveBeenCalledWith(
        "Alert #2 dismissed successfully",
      );
      expect(consoleLogSpy).toHaveBeenCalledWith(
        "Successfully processed 2 alerts.",
      );
      expect(processExitSpy).not.toHaveBeenCalled();
    });

    it("should exit if GITHUB_TOKEN is missing", async () => {
      delete process.env.GITHUB_TOKEN;
      await run();
      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "Error:",
        "Missing required env var GITHUB_TOKEN",
      );
      expect(processExitSpy).toHaveBeenCalledWith(1);
    });

    it("should exit if INPUT_PATHS is missing", async () => {
      process.env.INPUT_PATHS = "";
      await run();
      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "Error:",
        "No path patterns provided. Please specify paths to match.",
      );
      expect(processExitSpy).toHaveBeenCalledWith(1);
    });

    it("should exit if INPUT_DISMISSAL_REASON is invalid", async () => {
      process.env.INPUT_DISMISSAL_REASON = "wrong_reason";
      await run();
      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "Error:",
        "Invalid dismissal reason: wrong_reason. Must be one of fix_started, inaccurate, no_bandwidth, not_used, tolerable_risk",
      );
      expect(processExitSpy).toHaveBeenCalledWith(1);
    });

    it("should handle no alerts found", async () => {
      mockOctokit.paginate.mockResolvedValue([]);
      // Mock successful repo check
      mockOctokit.request.mockImplementation((route: string) => {
        if (route === "GET /repos/{owner}/{repo}") {
          return { data: { owner: { login: "test-user" } } };
        }
        return {};
      });

      await run();
      expect(consoleLogSpy).toHaveBeenCalledWith("Found 0 open alerts");
      expect(consoleLogSpy).toHaveBeenCalledWith("No alerts found to process.");
      expect(processExitSpy).not.toHaveBeenCalled();
    });

    it("should handle no alerts matching path patterns", async () => {
      const mockAlerts: DependabotAlert[] = [
        createMockAlert(1, "high", "pkg-a", "nomatch/package.json"),
      ];
      mockOctokit.paginate.mockResolvedValue(mockAlerts);
      // Mock successful repo check
      mockOctokit.request.mockImplementation((route: string) => {
        if (route === "GET /repos/{owner}/{repo}") {
          return { data: { owner: { login: "test-user" } } };
        }
        return {};
      });

      process.env.INPUT_PATHS = "src/package-lock.json";
      await run();
      expect(consoleLogSpy).toHaveBeenCalledWith(
        "No alerts matched the provided path patterns.",
      );
      expect(processExitSpy).not.toHaveBeenCalled();
    });

    it("should handle error during token authentication check", async () => {
      mockOctokit.request.mockImplementation((route: string) => {
        if (route === "GET /repos/{owner}/{repo}") {
          throw new Error("Auth failed");
        }
        return {};
      });

      await run();
      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "Error authenticating with GitHub. Please check your token.",
      );
      expect(consoleErrorSpy).toHaveBeenCalledWith("Error:", "Auth failed");
      expect(processExitSpy).toHaveBeenCalledWith(1);
    });

    it("should handle 404 error when fetching alerts (repo not found or alerts disabled)", async () => {
      mockOctokit.request.mockImplementation((route: string) => {
        if (route === "GET /repos/{owner}/{repo}") {
          return { data: { owner: { login: "test-user" } } };
        }
        return {};
      });
      mockOctokit.paginate.mockRejectedValue(
        new RequestError("Not Found", 404, {
          request: { method: "GET", headers: {}, url: "http://dummy.url/api" },
          response: { headers: {}, status: 404, url: "", data: {} },
        }),
      );

      await run();
      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "Error: Repository owner/repo not found or Dependabot alerts are not enabled.",
      );
      expect(processExitSpy).toHaveBeenCalledWith(1);
    });

    it("should handle 403 error when fetching alerts (permission issue)", async () => {
      mockOctokit.request.mockImplementation((route: string) => {
        if (route === "GET /repos/{owner}/{repo}") {
          return { data: { owner: { login: "test-user" } } };
        }
        return {};
      });
      mockOctokit.paginate.mockRejectedValue(
        new RequestError("Forbidden", 403, {
          request: { method: "GET", headers: {}, url: "http://dummy.url/api" },
          response: { headers: {}, status: 403, url: "", data: {} },
        }),
      );

      await run();
      expect(consoleErrorSpy).toHaveBeenCalledWith(
        expect.stringContaining("ERROR: Cannot access Dependabot alerts."),
      );
      expect(processExitSpy).toHaveBeenCalledWith(1);
    });

    it("should handle general API error when fetching alerts", async () => {
      mockOctokit.request.mockImplementation((route: string) => {
        if (route === "GET /repos/{owner}/{repo}") {
          return { data: { owner: { login: "test-user" } } };
        }
        return {};
      });

      mockOctokit.paginate.mockRejectedValue(
        new RequestError("Server Error", 500, {
          request: { method: "GET", headers: {}, url: "http://dummy.url/api" },
          response: { headers: {}, status: 500, url: "", data: {} },
        }),
      );

      await run();
      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "API Error 500: Server Error",
      );
      expect(processExitSpy).toHaveBeenCalledWith(1);
    });

    it("should handle 403 error when dismissing alerts", async () => {
      const mockAlerts: DependabotAlert[] = [
        createMockAlert(1, "high", "pkg-a", "src/package-lock.json"),
      ];
      mockOctokit.paginate.mockResolvedValue(mockAlerts);
      mockOctokit.request.mockImplementation((route: string) => {
        if (route === "GET /repos/{owner}/{repo}") {
          return { data: { owner: { login: "test-user" } } };
        }
        if (
          route ===
          "PATCH /repos/{owner}/{repo}/dependabot/alerts/{alert_number}"
        ) {
          throw new RequestError("Forbidden for dismiss", 403, {
            request: {
              method: "PATCH",
              headers: {},
              url: "http://dummy.url/api",
            },
            response: { headers: {}, status: 403, url: "", data: {} },
          });
        }
        return {};
      });

      process.env.INPUT_PATHS = "src/package-lock.json";
      await run();

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "Error: Permission denied when trying to dismiss alerts.",
      );
      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "Make sure the GITHUB_TOKEN has 'security-events: write' permission.",
      );
      expect(processExitSpy).toHaveBeenCalledWith(1);
    });

    it("should handle non-RequestError when dismissing alerts", async () => {
      const mockAlerts: DependabotAlert[] = [
        createMockAlert(1, "high", "pkg-a", "src/package-lock.json"),
      ];
      mockOctokit.paginate.mockResolvedValue(mockAlerts);
      mockOctokit.request.mockImplementation((route: string) => {
        if (route === "GET /repos/{owner}/{repo}") {
          return { data: { owner: { login: "test-user" } } };
        }
        if (
          route ===
          "PATCH /repos/{owner}/{repo}/dependabot/alerts/{alert_number}"
        ) {
          throw new Error("Some other dismiss error");
        }
        return {};
      });

      process.env.INPUT_PATHS = "src/package-lock.json";
      await run();
      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "Error dismissing alerts:",
        "Some other dismiss error",
      );
      expect(processExitSpy).toHaveBeenCalledWith(1);
    });

    it("should close PRs when INPUT_CLOSE_PRS is true", async () => {
      const mockAlerts: DependabotAlert[] = [
        createMockAlert(1, "high", "pkg-a", "src/package-lock.json"),
        createMockAlert(2, "critical", "pkg-b", "other/yarn.lock"),
      ];
      mockOctokit.paginate.mockResolvedValue(mockAlerts);
      mockOctokit.request.mockResolvedValue({ status: 200 }); // For dismissal
      mockOctokit.rest.pulls.update.mockResolvedValue({ status: 200 }); // For PR closure

      // Mock successful repo check
      mockOctokit.request.mockImplementation((route: string) => {
        if (route === "GET /repos/{owner}/{repo}") {
          return { data: { owner: { login: "test-user" } } };
        }
        if (
          route ===
          "PATCH /repos/{owner}/{repo}/dependabot/alerts/{alert_number}"
        ) {
          return { status: 200 };
        }
        return {};
      });

      // Mock GraphQL response for PR mappings
      const mockGraphqlResponse = {
        repository: {
          alert0: {
            number: 1,
            dependabotUpdate: {
              pullRequest: { number: 101 },
            },
          },
          alert1: {
            number: 2,
            dependabotUpdate: {
              pullRequest: { number: 102 },
            },
          },
        },
      };
      mockGraphql.mockResolvedValue(mockGraphqlResponse);

      process.env.INPUT_PATHS = "src/package-lock.json\nother/yarn.lock";
      process.env.INPUT_CLOSE_PRS = "true";

      await run();

      expect(consoleLogSpy).toHaveBeenCalledWith(
        "Fetching PR mappings for all alerts to ensure safe closure...",
      );
      expect(consoleLogSpy).toHaveBeenCalledWith(
        "Found 2 alerts with associated PRs",
      );
      expect(consoleLogSpy).toHaveBeenCalledWith(
        "PR #101 will be closed (all 1 associated alerts are being dismissed: 1)"
      );
      expect(consoleLogSpy).toHaveBeenCalledWith(
        "PR #102 will be closed (all 1 associated alerts are being dismissed: 2)"
      );
      expect(consoleLogSpy).toHaveBeenCalledWith(
        "Closing PR #101 for alert #1...",
      );
      expect(consoleLogSpy).toHaveBeenCalledWith(
        "Closing PR #102 for alert #2...",
      );
      expect(mockOctokit.rest.pulls.update).toHaveBeenCalledWith({
        owner: "owner",
        repo: "repo",
        pull_number: 101,
        state: "closed",
      });
      expect(mockOctokit.rest.pulls.update).toHaveBeenCalledWith({
        owner: "owner",
        repo: "repo",
        pull_number: 102,
        state: "closed",
      });
      expect(consoleLogSpy).toHaveBeenCalledWith(
        "Successfully processed 2 alerts.",
      );
      expect(processExitSpy).not.toHaveBeenCalled();
    });

    it("should handle error when fetching PR mappings", async () => {
      const mockAlerts: DependabotAlert[] = [
        createMockAlert(1, "high", "pkg-a", "src/package-lock.json"),
      ];
      mockOctokit.paginate.mockResolvedValue(mockAlerts);

      // Mock successful repo check
      mockOctokit.request.mockImplementation((route: string) => {
        if (route === "GET /repos/{owner}/{repo}") {
          return { data: { owner: { login: "test-user" } } };
        }
        return {};
      });

      // Mock GraphQL error
      mockGraphql.mockRejectedValue(new Error("GraphQL API error"));

      process.env.INPUT_PATHS = "src/package-lock.json";
      process.env.INPUT_CLOSE_PRS = "true";

      await run();

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "Error fetching alert-PR mappings. Cannot proceed with PR closure.",
        expect.any(Error),
      );
      expect(processExitSpy).toHaveBeenCalledWith(1);
    });

    it("should handle error when closing PRs", async () => {
      const mockAlerts: DependabotAlert[] = [
        createMockAlert(1, "high", "pkg-a", "src/package-lock.json"),
      ];
      mockOctokit.paginate.mockResolvedValue(mockAlerts);
      mockOctokit.request.mockResolvedValue({ status: 200 }); // For repo check
      mockOctokit.rest.pulls.update.mockRejectedValue(
        new Error("PR close error"),
      );

      // Mock successful repo check
      mockOctokit.request.mockImplementation((route: string) => {
        if (route === "GET /repos/{owner}/{repo}") {
          return { data: { owner: { login: "test-user" } } };
        }
        return {};
      });

      // Mock GraphQL response for PR mappings
      const mockGraphqlResponse = {
        repository: {
          alert0: {
            number: 1,
            dependabotUpdate: {
              pullRequest: { number: 101 },
            },
          },
        },
      };
      mockGraphql.mockResolvedValue(mockGraphqlResponse);

      process.env.INPUT_PATHS = "src/package-lock.json";
      process.env.INPUT_CLOSE_PRS = "true";

      await run();

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "Error closing PR #101 for alert #1:",
        expect.any(Error),
      );
      expect(processExitSpy).toHaveBeenCalledWith(1);
    });

    it("should skip closing PRs when no PR mappings are found", async () => {
      const mockAlerts: DependabotAlert[] = [
        createMockAlert(1, "high", "pkg-a", "src/package-lock.json"),
      ];
      mockOctokit.paginate.mockResolvedValue(mockAlerts);
      mockOctokit.request.mockResolvedValue({ status: 200 }); // For dismissal

      // Mock successful repo check
      mockOctokit.request.mockImplementation((route: string) => {
        if (route === "GET /repos/{owner}/{repo}") {
          return { data: { owner: { login: "test-user" } } };
        }
        if (
          route ===
          "PATCH /repos/{owner}/{repo}/dependabot/alerts/{alert_number}"
        ) {
          return { status: 200 };
        }
        return {};
      });

      // Mock GraphQL response with no PR mappings
      const mockGraphqlResponse = {
        repository: {
          alert0: {
            number: 1,
            dependabotUpdate: null,
          },
        },
      };
      mockGraphql.mockResolvedValue(mockGraphqlResponse);

      process.env.INPUT_PATHS = "src/package-lock.json";
      process.env.INPUT_CLOSE_PRS = "true";

      await run();

      expect(consoleLogSpy).toHaveBeenCalledWith(
        "Found 0 alerts with associated PRs",
      );
      expect(mockOctokit.rest.pulls.update).not.toHaveBeenCalled();
      expect(consoleLogSpy).toHaveBeenCalledWith(
        "Alert #1 dismissed successfully",
      );
      expect(processExitSpy).not.toHaveBeenCalled();
    });

    it("should not fetch PR mappings when INPUT_CLOSE_PRS is false", async () => {
      const mockAlerts: DependabotAlert[] = [
        createMockAlert(1, "high", "pkg-a", "src/package-lock.json"),
      ];
      mockOctokit.paginate.mockResolvedValue(mockAlerts);
      mockOctokit.request.mockResolvedValue({ status: 200 }); // For dismissal

      // Mock successful repo check
      mockOctokit.request.mockImplementation((route: string) => {
        if (route === "GET /repos/{owner}/{repo}") {
          return { data: { owner: { login: "test-user" } } };
        }
        if (
          route ===
          "PATCH /repos/{owner}/{repo}/dependabot/alerts/{alert_number}"
        ) {
          return { status: 200 };
        }
        return {};
      });

      process.env.INPUT_PATHS = "src/package-lock.json";
      process.env.INPUT_CLOSE_PRS = "false";

      await run();

      expect(consoleLogSpy).not.toHaveBeenCalledWith(
        "Fetching PR mappings for all alerts to ensure safe closure...",
      );
      expect(mockGraphql).not.toHaveBeenCalled();
      expect(mockOctokit.rest.pulls.update).not.toHaveBeenCalled();
      expect(consoleLogSpy).toHaveBeenCalledWith(
        "Alert #1 dismissed successfully",
      );
      expect(processExitSpy).not.toHaveBeenCalled();
    });

    it("should only close PRs when all associated alerts are being dismissed", async () => {
      // Create 3 alerts: 2 will be dismissed (match patterns), 1 will not (doesn't match pattern)
      const mockAlerts: DependabotAlert[] = [
        createMockAlert(1, "high", "pkg-a", "src/package-lock.json"), // Will be dismissed
        createMockAlert(2, "critical", "pkg-b", "src/package-lock.json"), // Will be dismissed
        createMockAlert(3, "medium", "pkg-c", "other/package.json"), // Will NOT be dismissed (pattern mismatch)
      ];
      mockOctokit.paginate.mockResolvedValue(mockAlerts);
      mockOctokit.request.mockResolvedValue({ status: 200 }); // For dismissal
      mockOctokit.rest.pulls.update.mockResolvedValue({ status: 200 }); // For PR closure

      // Mock successful repo check
      mockOctokit.request.mockImplementation((route: string) => {
        if (route === "GET /repos/{owner}/{repo}") {
          return { data: { owner: { login: "test-user" } } };
        }
        if (
          route ===
          "PATCH /repos/{owner}/{repo}/dependabot/alerts/{alert_number}"
        ) {
          return { status: 200 };
        }
        return {};
      });

      // Mock GraphQL response: PR 101 is only for alert 1 (safe to close), 
      // PR 102 is for both alert 2 and alert 3 (NOT safe to close)
      const mockGraphqlResponse = {
        repository: {
          alert0: {
            number: 1,
            dependabotUpdate: {
              pullRequest: { number: 101 },
            },
          },
          alert1: {
            number: 2,
            dependabotUpdate: {
              pullRequest: { number: 102 },
            },
          },
          alert2: {
            number: 3,
            dependabotUpdate: {
              pullRequest: { number: 102 }, // Same PR as alert 2
            },
          },
        },
      };
      mockGraphql.mockResolvedValue(mockGraphqlResponse);

      process.env.INPUT_PATHS = "src/package-lock.json"; // Only matches alerts 1 and 2
      process.env.INPUT_CLOSE_PRS = "true";

      await run();

      // Should fetch PR mappings for all alerts
      expect(consoleLogSpy).toHaveBeenCalledWith(
        "Fetching PR mappings for all alerts to ensure safe closure...",
      );
      expect(consoleLogSpy).toHaveBeenCalledWith(
        "Found 3 alerts with associated PRs",
      );
      
      // Should indicate PR 101 will be closed (only associated with alert 1 which is being dismissed)
      expect(consoleLogSpy).toHaveBeenCalledWith(
        "PR #101 will be closed (all 1 associated alerts are being dismissed: 1)"
      );
      
      // Should indicate PR 102 will NOT be closed (associated with alert 2 being dismissed AND alert 3 not being dismissed)
      expect(consoleLogSpy).toHaveBeenCalledWith(
        "PR #102 will NOT be closed (1 alerts retained: 3)"
      );

      // Should only close PR 101
      expect(consoleLogSpy).toHaveBeenCalledWith(
        "Closing PR #101 for alert #1...",
      );
      expect(mockOctokit.rest.pulls.update).toHaveBeenCalledWith({
        owner: "owner",
        repo: "repo",
        pull_number: 101,
        state: "closed",
      });

      // Should skip closing PR 102 and log the reason
      expect(consoleLogSpy).toHaveBeenCalledWith(
        "Skipping closure of PR #102 for alert #2 (PR has other alerts that are not being dismissed)"
      );

      // Should NOT call update for PR 102
      expect(mockOctokit.rest.pulls.update).not.toHaveBeenCalledWith({
        owner: "owner",
        repo: "repo",
        pull_number: 102,
        state: "closed",
      });

      // Should dismiss both matching alerts
      expect(consoleLogSpy).toHaveBeenCalledWith(
        "Alert #1 dismissed successfully",
      );
      expect(consoleLogSpy).toHaveBeenCalledWith(
        "Alert #2 dismissed successfully",
      );
      expect(consoleLogSpy).toHaveBeenCalledWith(
        "Successfully processed 2 alerts.",
      );
      expect(processExitSpy).not.toHaveBeenCalled();
    });
  });
});
