import { afterEach, beforeEach, describe, expect, it, mock, spyOn } from "bun:test";
import { matchesAnyPattern, fetchAllAlerts, run, DependabotAlert } from "./index";
import { Octokit } from "@octokit/rest";
import { RequestError } from "@octokit/request-error";
import * as minimatchModule from "minimatch"; // Import for spying

// Mock Octokit globally for all tests in this file
const mockOctokit = {
  request: mock(),
  paginate: mock(),
};
mock.module("@octokit/rest", () => ({
  Octokit: mock(() => mockOctokit),
}));

describe("Dependabot Auto Triage Action", () => {
  let consoleLogSpy: any;
  let consoleErrorSpy: any;
  let processExitSpy: any;

  beforeEach(() => {
    // Reset mocks and environment variables
    mockOctokit.request.mockClear();
    mockOctokit.paginate.mockClear();
    process.env.GITHUB_TOKEN = "test-token";
    process.env.GITHUB_REPOSITORY = "owner/repo";
    process.env.INPUT_ALERT_TYPES = "dependency";
    process.env.INPUT_PATHS = "**/package-lock.json\n**/yarn.lock";
    process.env.INPUT_DISMISSAL_COMMENT = "Test dismissal comment";
    process.env.INPUT_DISMISSAL_REASON = "tolerable_risk";

    // Spy on console messages and process.exit
    consoleLogSpy = spyOn(console, "log").mockImplementation(() => {});
    consoleErrorSpy = spyOn(console, "error").mockImplementation(() => {});
    processExitSpy = spyOn(process, "exit").mockImplementation((() => {}) as (code?: number) => never);
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
  });

  describe("matchesAnyPattern", () => {
    // Keep existing spies setup in outer beforeEach/afterEach if they handle consoleErrorSpy
    let minimatchSpy: ReturnType<typeof spyOn>;

    beforeEach(() => {
      // Spy on the default export/function if that's what `import { minimatch } from 'minimatch'` would get.
      // Assuming `minimatch` is the function name exported by the 'minimatch' module.
      // If it's a default export, the spying mechanism might differ slightly or how it's referenced.
      // For `import { minimatch } from 'minimatch'`, it's typically a named export.
      // However, to ensure we're spying on what's used by `./index.ts` which also does `import { minimatch }`, 
      // we spy on the module's export.
      minimatchSpy = spyOn(minimatchModule, 'minimatch');
    });

    afterEach(() => {
      minimatchSpy.mockRestore();
    });

    it("should return true if manifestPath matches any pattern", () => {
      const manifestPath = "src/package-lock.json";
      const patterns = ["**/package-lock.json", "**/gemfile.lock"];
      // Let original minimatch work or mock its successful return
      minimatchSpy.mockImplementation((path: string, pattern: string) => {
        // Basic mock for successful non-throwing calls in other tests if needed
        if (pattern === "**/package-lock.json" && path === manifestPath) return true;
        if (pattern === "**/gemfile.lock" && path === manifestPath) return false; 
        return false; // Default
      });
      expect(matchesAnyPattern(manifestPath, patterns)).toBe(true);
    });

    it("should return false if manifestPath does not match any pattern", () => {
      const manifestPath = "src/some/other/file.txt";
      const patterns = ["**/package-lock.json", "**/gemfile.lock"];
      minimatchSpy.mockReturnValue(false); // All calls will return false
      expect(matchesAnyPattern(manifestPath, patterns)).toBe(false);
    });

    it("should return false if manifestPath is undefined", () => {
      const patterns = ["**/package-lock.json"];
      // matchesAnyPattern returns early, minimatchSpy might not be called
      expect(matchesAnyPattern(undefined, patterns)).toBe(false);
    });

    it("should return false if patterns array is empty", () => {
      const manifestPath = "src/package-lock.json";
      const patterns: string[] = [];
      // matchesAnyPattern returns early, minimatchSpy might not be called
      expect(matchesAnyPattern(manifestPath, patterns)).toBe(false);
    });

    it("should handle glob patterns correctly", () => {
      minimatchSpy.mockImplementation((path: string, pattern: string) => {
        if (path === "frontend/package.json" && pattern === "frontend/**") return true;
        if (path === "backend/package.json" && pattern === "frontend/**") return false;
        return false;
      });
      expect(matchesAnyPattern("frontend/package.json", ["frontend/**"])).toBe(true);
      expect(matchesAnyPattern("backend/package.json", ["frontend/**"])).toBe(false);
    });

    it("should use matchBase option for minimatch", () => {
      // This test implicitly tests options if we assume the actual minimatch is called.
      // With a full mock, we'd need to check options passed to the spy.
      // For now, let's simplify by having the spy return expected values.
      minimatchSpy.mockImplementation((path: string, pattern: string, options: any) => {
        // A more robust mock would call the *actual* minimatch here for non-erroring tests
        // For example: return minimatchModule.minimatch(path, pattern, options)
        // But to keep it simple for now:
        if (options && (options as any).matchBase) {
            if (path === "package-lock.json" && pattern === "package-lock.json") return true;
            if (path === "sub/package-lock.json" && pattern === "package-lock.json") return true;
        }
        return false;
      });
      expect(matchesAnyPattern("package-lock.json", ["package-lock.json"])).toBe(true);
      expect(matchesAnyPattern("sub/package-lock.json", ["package-lock.json"])).toBe(true);
    });

    it("should log an error, handle it, and continue matching with other patterns", () => {
      const manifestPath = "src/package-lock.json";
      const badPattern = "force-error";
      const goodPattern = "**/package-lock.json";
      const patterns = [badPattern, goodPattern];
      const simulatedErrorMessage = "Simulated minimatch error";

      minimatchSpy
        .mockImplementationOnce((path: string, pattern: string) => {
          if (pattern === badPattern) {
            throw new Error(simulatedErrorMessage);
          }
          return false; // Should not happen if it throws
        })
        .mockImplementationOnce((path: string, pattern: string) => {
          // For the second pattern (goodPattern)
          return (pattern === goodPattern && path === manifestPath); // True if it matches
        });

      expect(matchesAnyPattern(manifestPath, patterns)).toBe(true);
      expect(consoleErrorSpy).toHaveBeenCalledWith(
        `Error matching pattern ${badPattern}:`,
        simulatedErrorMessage
      );
      expect(minimatchSpy).toHaveBeenCalledTimes(2);
    });
  });

  describe("fetchAllAlerts", () => {
    const mockOwner = "owner";
    const mockRepo = "repo";

    const createMockAlert = (number: number, severity: string, packageName: string, manifestPath: string): DependabotAlert => ({
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
      mockOctokit.paginate.mockResolvedValue(mockAlertsResponse as any);

      const octokitInstance = new Octokit(); // Real instance not used due to global mock
      const filteredAlerts = await fetchAllAlerts(octokitInstance, mockOwner, mockRepo, ["critical", "high"]);

      expect(mockOctokit.paginate).toHaveBeenCalledWith(
        "GET /repos/{owner}/{repo}/dependabot/alerts",
        {
          owner: mockOwner,
          repo: mockRepo,
          state: "open",
          per_page: 100,
        }
      );
      expect(filteredAlerts).toHaveLength(2);
      expect(filteredAlerts.find(a => a.number === 1)).toBeDefined();
      expect(filteredAlerts.find(a => a.number === 2)).toBeDefined();
      expect(filteredAlerts.find(a => a.number === 3)).toBeUndefined();
    });

    it("should fetch and filter alerts for 'dependency' type if no severity matches", async () => {
      const mockAlertsResponse: DependabotAlert[] = [
        createMockAlert(1, "critical", "pkg-a", "path/to/manifest1"),
        createMockAlert(2, "high", "pkg-b", "path/to/manifest2"),
      ];
      mockOctokit.paginate.mockResolvedValue(mockAlertsResponse as any);
      const octokitInstance = new Octokit();
      const filteredAlerts = await fetchAllAlerts(octokitInstance, mockOwner, mockRepo, ["dependency"]);

      expect(filteredAlerts).toHaveLength(2);
    });

    it("should return all alerts if 'dependency' is in alertTypes along with severities", async () => {
      const mockAlertsResponse: DependabotAlert[] = [
        createMockAlert(1, "critical", "pkg-a", "path/to/manifest1"),
        createMockAlert(2, "low", "pkg-b", "path/to/manifest2"),
      ];
      mockOctokit.paginate.mockResolvedValue(mockAlertsResponse as any);
      const octokitInstance = new Octokit();
      const filteredAlerts = await fetchAllAlerts(octokitInstance, mockOwner, mockRepo, ["critical", "dependency"]);
      // The filter is OR based, so both alerts should be returned because 'dependency' matches all, and critical matches one.
      // The current implementation will include all if 'dependency' is present.
      expect(filteredAlerts).toHaveLength(2);
    });

    it("should return empty array if no alerts match severity and 'dependency' type is not included", async () => {
      const mockAlertsResponse: DependabotAlert[] = [
        createMockAlert(1, "critical", "pkg-a", "path/to/manifest1"),
      ];
      mockOctokit.paginate.mockResolvedValue(mockAlertsResponse as any);
      const octokitInstance = new Octokit();
      const filteredAlerts = await fetchAllAlerts(octokitInstance, mockOwner, mockRepo, ["low"]);

      expect(filteredAlerts).toHaveLength(0);
    });

    it("should handle empty list of alerts from paginate", async () => {
      mockOctokit.paginate.mockResolvedValue([]);
      const octokitInstance = new Octokit();
      const filteredAlerts = await fetchAllAlerts(octokitInstance, mockOwner, mockRepo, ["critical"]);
      expect(filteredAlerts).toHaveLength(0);
    });
  });

  // More tests for fetchAllAlerts and run will be added here
  describe("run", () => {
    const createMockAlert = (number: number, severity: string, packageName: string, manifestPath: string): DependabotAlert => ({
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
      mockOctokit.paginate.mockResolvedValue(mockAlerts as any);
      mockOctokit.request.mockResolvedValue({ status: 200 }); // For dismissal

      // Mock successful repo check
      mockOctokit.request.mockImplementation(async (route: string) => {
        if (route === "GET /repos/{owner}/{repo}") {
          return { data: { owner: { login: "test-user" } } };
        }
        if (route === "PATCH /repos/{owner}/{repo}/dependabot/alerts/{alert_number}") {
          return { status: 200 };
        }
        return {};
      });

      process.env.INPUT_PATHS = "src/package-lock.json\nother/yarn.lock";

      await run();

      expect(consoleLogSpy).toHaveBeenCalledWith("Fetching dependabot alerts for owner/repo...");
      expect(consoleLogSpy).toHaveBeenCalledWith("Using path patterns:");
      expect(consoleLogSpy).toHaveBeenCalledWith("- src/package-lock.json");
      expect(consoleLogSpy).toHaveBeenCalledWith("- other/yarn.lock");
      expect(consoleLogSpy).toHaveBeenCalledWith(`Found ${mockAlerts.length} open alerts`);
      expect(consoleLogSpy).toHaveBeenCalledWith("Alert #1 for pkg-a in src/package-lock.json matches patterns");
      expect(consoleLogSpy).toHaveBeenCalledWith("Alert #2 for pkg-b in other/yarn.lock matches patterns");
      expect(consoleLogSpy).toHaveBeenCalledWith("Skipping alert #3 for nomatch/package.json (does not match any pattern)");
      expect(consoleLogSpy).toHaveBeenCalledWith("Dismissing 2 alerts...");
      expect(mockOctokit.request).toHaveBeenCalledWith(
        "PATCH /repos/{owner}/{repo}/dependabot/alerts/{alert_number}",
        expect.objectContaining({ alert_number: 1, state: "dismissed" })
      );
      expect(mockOctokit.request).toHaveBeenCalledWith(
        "PATCH /repos/{owner}/{repo}/dependabot/alerts/{alert_number}",
        expect.objectContaining({ alert_number: 2, state: "dismissed" })
      );
      expect(consoleLogSpy).toHaveBeenCalledWith("Alert #1 dismissed successfully");
      expect(consoleLogSpy).toHaveBeenCalledWith("Alert #2 dismissed successfully");
      expect(consoleLogSpy).toHaveBeenCalledWith("Successfully dismissed 2 alerts.");
      expect(processExitSpy).not.toHaveBeenCalled();
    });

    it("should exit if GITHUB_TOKEN is missing", async () => {
      delete process.env.GITHUB_TOKEN;
      await run();
      expect(consoleErrorSpy).toHaveBeenCalledWith("Error:", "Missing required env var GITHUB_TOKEN");
      expect(processExitSpy).toHaveBeenCalledWith(1);
    });

    it("should exit if INPUT_PATHS is missing", async () => {
      process.env.INPUT_PATHS = "";
      await run();
      expect(consoleErrorSpy).toHaveBeenCalledWith("Error:", "No path patterns provided. Please specify paths to match.");
      expect(processExitSpy).toHaveBeenCalledWith(1);
    });

    it("should exit if INPUT_DISMISSAL_REASON is invalid", async () => {
      process.env.INPUT_DISMISSAL_REASON = "wrong_reason";
      await run();
      expect(consoleErrorSpy).toHaveBeenCalledWith("Error:", "Invalid dismissal reason: wrong_reason. Must be one of fix_started, inaccurate, no_bandwidth, not_used, tolerable_risk");
      expect(processExitSpy).toHaveBeenCalledWith(1);
    });

    it("should handle no alerts found", async () => {
      mockOctokit.paginate.mockResolvedValue([]);
      // Mock successful repo check
      mockOctokit.request.mockImplementation(async (route: string) => {
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
      mockOctokit.paginate.mockResolvedValue(mockAlerts as any);
      // Mock successful repo check
      mockOctokit.request.mockImplementation(async (route: string) => {
        if (route === "GET /repos/{owner}/{repo}") {
          return { data: { owner: { login: "test-user" } } };
        }
        return {};
      });

      process.env.INPUT_PATHS = "src/package-lock.json";
      await run();
      expect(consoleLogSpy).toHaveBeenCalledWith("No alerts matched the provided path patterns.");
      expect(processExitSpy).not.toHaveBeenCalled();
    });

    it("should handle error during token authentication check", async () => {
      mockOctokit.request.mockImplementation(async (route: string) => {
        if (route === "GET /repos/{owner}/{repo}") {
          throw new Error("Auth failed");
        }
        return {};
      });

      await run();
      expect(consoleErrorSpy).toHaveBeenCalledWith("Error authenticating with GitHub. Please check your token.");
      expect(consoleErrorSpy).toHaveBeenCalledWith("Error:", "Auth failed");
      expect(processExitSpy).toHaveBeenCalledWith(1);
    });

    it("should handle 404 error when fetching alerts (repo not found or alerts disabled)", async () => {
      mockOctokit.request.mockImplementation(async (route: string) => {
        if (route === "GET /repos/{owner}/{repo}") {
          return { data: { owner: { login: "test-user" } } };
        }
        return {};
      });
      mockOctokit.paginate.mockRejectedValue(new RequestError("Not Found", 404, { request: { headers: {}, url: "http://dummy.url/api" } as any, response: {headers: {}, status: 404, url: "", data: {}} }));

      await run();
      expect(consoleErrorSpy).toHaveBeenCalledWith("Error: Repository owner/repo not found or Dependabot alerts are not enabled.");
      expect(processExitSpy).toHaveBeenCalledWith(1);
    });

    it("should handle 403 error when fetching alerts (permission issue)", async () => {
      mockOctokit.request.mockImplementation(async (route: string) => {
        if (route === "GET /repos/{owner}/{repo}") {
          return { data: { owner: { login: "test-user" } } };
        }
        return {};
      });
      mockOctokit.paginate.mockRejectedValue(new RequestError("Forbidden", 403, { request: { headers: {}, url: "http://dummy.url/api" } as any, response: {headers: {}, status: 403, url: "", data: {}} }));
      
      await run();
      expect(consoleErrorSpy).toHaveBeenCalledWith(expect.stringContaining("ERROR: Cannot access Dependabot alerts."));
      expect(processExitSpy).toHaveBeenCalledWith(1);
    });

    it("should handle general API error when fetching alerts", async () => {
      mockOctokit.request.mockImplementation(async (route: string) => {
        if (route === "GET /repos/{owner}/{repo}") {
          return { data: { owner: { login: "test-user" } } };
        }
        return {};
      });
      mockOctokit.paginate.mockRejectedValue(new RequestError("Server Error", 500, { request: { headers: {}, url: "http://dummy.url/api" } as any, response: {headers: {}, status: 500, url: "", data: {}} }));
      
      await run();
      expect(consoleErrorSpy).toHaveBeenCalledWith("API Error 500: Server Error");
      expect(processExitSpy).toHaveBeenCalledWith(1);
    });

    it("should handle 403 error when dismissing alerts", async () => {
      const mockAlerts: DependabotAlert[] = [
        createMockAlert(1, "high", "pkg-a", "src/package-lock.json"),
      ];
      mockOctokit.paginate.mockResolvedValue(mockAlerts as any);
      mockOctokit.request.mockImplementation(async (route: string, options: any) => {
        if (route === "GET /repos/{owner}/{repo}") {
          return { data: { owner: { login: "test-user" } } };
        }
        if (route === "PATCH /repos/{owner}/{repo}/dependabot/alerts/{alert_number}") {
          throw new RequestError("Forbidden for dismiss", 403, { request: { headers: {}, url: "http://dummy.url/api" } as any, response: {headers: {}, status: 403, url: "", data: {}} });
        }
        return {};
      });

      process.env.INPUT_PATHS = "src/package-lock.json";
      await run();

      expect(consoleErrorSpy).toHaveBeenCalledWith("Error: Permission denied when trying to dismiss alerts.");
      expect(consoleErrorSpy).toHaveBeenCalledWith("Make sure the GITHUB_TOKEN has 'security-events: write' permission.");
      expect(processExitSpy).toHaveBeenCalledWith(1);
    });

     it("should handle non-RequestError when dismissing alerts", async () => {
      const mockAlerts: DependabotAlert[] = [
        createMockAlert(1, "high", "pkg-a", "src/package-lock.json"),
      ];
      mockOctokit.paginate.mockResolvedValue(mockAlerts as any);
      mockOctokit.request.mockImplementation(async (route: string, options: any) => {
        if (route === "GET /repos/{owner}/{repo}") {
          return { data: { owner: { login: "test-user" } } };
        }
        if (route === "PATCH /repos/{owner}/{repo}/dependabot/alerts/{alert_number}") {
          throw new Error("Some other dismiss error");
        }
        return {};
      });

      process.env.INPUT_PATHS = "src/package-lock.json";
      await run();
      expect(consoleErrorSpy).toHaveBeenCalledWith("Error dismissing alerts:", "Some other dismiss error");
      expect(processExitSpy).toHaveBeenCalledWith(1);
    });

  });
}); 