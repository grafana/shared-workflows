import { Octokit } from "@octokit/rest";
import { graphql } from "@octokit/graphql";
import { minimatch } from "minimatch";
import { RequestError } from "@octokit/request-error";

// Define a simplified type for DependabotAlert with used properties
export interface DependabotAlert {
  number: number;
  dependency?: {
    package?: {
      name?: string;
    };
    manifest_path?: string;
  };
  security_advisory?: {
    severity?: string;
  };
  security_vulnerability?: {
    severity?: string;
  };
}

// GraphQL types for vulnerability alerts and associated PRs
export interface VulnerabilityAlert {
  number: number;
  dependabotUpdate?: {
    pullRequest?: {
      number: number;
    };
  };
}


export async function run() {
  try {
    const token = process.env.GITHUB_TOKEN;
    const alertTypes = (process.env.INPUT_ALERT_TYPES || "dependency")
      .split(",")
      .map((t) => t.trim());
    // Parse path patterns from multi-line format, filtering out empty lines
    const pathPatterns = (process.env.INPUT_PATHS || "")
      .split("\n")
      .map((p) => p.trim())
      .filter((p) => p !== "");

    const dismissalComment = process.env.INPUT_DISMISSAL_COMMENT;
    const dismissalReason = process.env.INPUT_DISMISSAL_REASON;
    const closePRs = process.env.INPUT_CLOSE_PRS === "true";

    const allowedDismissalReasons = [
      "fix_started",
      "inaccurate",
      "no_bandwidth",
      "not_used",
      "tolerable_risk",
    ];

    if (dismissalReason && !allowedDismissalReasons.includes(dismissalReason)) {
      throw new Error(
        `Invalid dismissal reason: ${dismissalReason}. Must be one of ${allowedDismissalReasons.join(", ")}`,
      );
    }

    if (!token) {
      throw new Error("Missing required env var GITHUB_TOKEN");
    }

    if (pathPatterns.length === 0) {
      throw new Error(
        "No path patterns provided. Please specify paths to match.",
      );
    }

    const octokit = new Octokit({ auth: token });
    const [owner, repo] = (process.env.GITHUB_REPOSITORY || "/").split("/");

    if (!owner || !repo) {
      throw new Error(
        "Could not determine repository owner and name from GITHUB_REPOSITORY",
      );
    }

    // Check token permissions first
    console.log("Checking GitHub token permissions...");
    try {
      const { data: tokenData } = await octokit.request(
        "GET /repos/{owner}/{repo}",
        {
          owner,
          repo,
        },
      );
      console.log(
        `Successfully authenticated with GitHub as ${tokenData.owner.login || "unknown user"}`,
      );
    } catch (error) {
      console.error(
        "Error authenticating with GitHub. Please check your token.",
      );
      throw error;
    }

    console.log(`Fetching dependabot alerts for ${owner}/${repo}...`);
    console.log(`Using path patterns:`);
    pathPatterns.forEach((pattern) => {
      console.log(`- ${pattern}`);
    });

    try {
      // Get all open alerts using the correct endpoint
      const alerts = await fetchAllAlerts(octokit, owner, repo, alertTypes);
      console.log(`Found ${alerts.length} open alerts`);

      if (alerts.length === 0) {
        console.log("No alerts found to process.");
        return;
      }

      // Process matching alerts
      const alertsToProcess: number[] = [];
      for (const alert of alerts) {
        const manifestPath = alert.dependency?.manifest_path;
        if (manifestPath && matchesAnyPattern(manifestPath, pathPatterns)) {
          console.log(
            `Alert #${alert.number} for ${alert.dependency?.package?.name ?? "unknown package"} in ${manifestPath} matches patterns`,
          );
          alertsToProcess.push(alert.number);
        } else {
          console.log(
            `Skipping alert #${alert.number} for ${manifestPath || "unknown path"} (does not match any pattern)`,
          );
        }
      }

      if (alertsToProcess.length === 0) {
        console.log("No alerts matched the provided path patterns.");
        return;
      }

      // Fetch alert-PR mappings only if we need to close PRs
      let alertPRMappings = new Map<number, number>();
      if (closePRs) {
        console.log("Fetching PR mappings for alerts before dismissal...");
        try {
          alertPRMappings = await fetchSpecificAlertsWithPRs(
            token,
            owner,
            repo,
            alertsToProcess,
          );
          console.log(`Found ${alertPRMappings.size} alerts with associated PRs`);
        } catch (error) {
          console.error("Error fetching alert-PR mappings. Cannot proceed with PR closure.", error);
          process.exit(1);
        }
      }

      console.log(`Dismissing ${alertsToProcess.length} alerts...`);

      try {
        for (const alertNumber of alertsToProcess) {
          // Find and close associated PR before dismissing alert
          const prNumber = alertPRMappings.get(alertNumber);
          if (prNumber) {
            console.log(`Closing PR #${prNumber} for alert #${alertNumber}...`);
            try {
              await octokit.rest.pulls.update({
                owner,
                repo,
                pull_number: prNumber,
                state: "closed",
              });
            } catch (error) {
              console.error(`Error closing PR #${prNumber} for alert #${alertNumber}:`, error);
              process.exit(1);
            }
          }

          // Now dismiss the alert
          await octokit.request(
            "PATCH /repos/{owner}/{repo}/dependabot/alerts/{alert_number}",
            {
              owner,
              repo,
              alert_number: alertNumber,
              state: "dismissed",
              dismissed_reason: dismissalReason as
                | "fix_started"
                | "inaccurate"
                | "no_bandwidth"
                | "not_used"
                | "tolerable_risk"
                | undefined,
              dismissed_comment: dismissalComment,
            },
          );
          console.log(`Alert #${alertNumber} dismissed successfully`);
        }
        console.log(`Successfully processed ${alertsToProcess.length} alerts.`);
      } catch (error) {
        if (error instanceof RequestError) {
          if (error.status === 403) {
            console.error(
              `Error: Permission denied when trying to dismiss alerts.`,
            );
            console.error(
              `Make sure the GITHUB_TOKEN has 'security-events: write' permission.`,
            );
            process.exit(1);
          } else {
            console.error(`API Error ${error.status}: ${error.message}`);
            process.exit(1);
          }
        } else {
          console.error(
            "Error dismissing alerts:",
            error instanceof Error ? error.message : String(error),
          );
          process.exit(1);
        }
      }
    } catch (error) {
      if (error instanceof RequestError) {
        if (error.status === 404) {
          console.error(
            `Error: Repository ${owner}/${repo} not found or Dependabot alerts are not enabled.`,
          );
          console.error(
            "Make sure Dependabot alerts are enabled in the repository's Security settings.",
          );
          process.exit(1);
        } else if (
          error.status === 403 ||
          error.message.includes("Resource not accessible by integration")
        ) {
          console.error(`
ERROR: Cannot access Dependabot alerts.

This error may occur for several reasons:
1. Dependabot alerts might not be enabled for this repository
   - Enable them at: https://github.com/${owner}/${repo}/settings/security_analysis

2. The token being used might not have sufficient permissions
   - For GITHUB_TOKEN: Make sure the workflow has 'security-events: write' permission
   - For fine-grained tokens: Make sure the 'Dependabot alerts: write' permission is enabled

3. Organization policies might be restricting access to security features
   - Check with your organization administrator

If you're sure the token has the right permissions, try accessing the alerts manually at:
https://github.com/${owner}/${repo}/security/dependabot
          `);
          process.exit(1);
        } else {
          console.error(`API Error ${error.status}: ${error.message}`);
          process.exit(1);
        }
      } else {
        console.error(
          "Error:",
          error instanceof Error ? error.message : String(error),
        );
        process.exit(1);
      }
    }
  } catch (error) {
    console.error(
      "Error:",
      error instanceof Error ? error.message : String(error),
    );
    process.exit(1);
  }
}

export async function fetchAllAlerts(
  octokit: Octokit,
  owner: string,
  repo: string,
  alertTypes: string[],
): Promise<DependabotAlert[]> {
  const allAlerts: DependabotAlert[] = await octokit.paginate(
    "GET /repos/{owner}/{repo}/dependabot/alerts",
    {
      owner,
      repo,
      state: "open",
      per_page: 100,
    },
  );

  // Filter alerts based on severity or 'dependency' type
  const filteredAlerts = allAlerts.filter(
    (alert) =>
      alertTypes.includes(alert.security_advisory?.severity as string) ||
      alertTypes.includes(alert.security_vulnerability?.severity as string) ||
      alertTypes.includes("dependency"),
  );

  return filteredAlerts;
}

export async function fetchSpecificAlertsWithPRs(
  token: string,
  owner: string,
  repo: string,
  alertNumbers: number[],
): Promise<Map<number, number>> {
  if (alertNumbers.length === 0) return new Map();

  const graphqlWithAuth = graphql.defaults({
    headers: {
      authorization: `token ${token}`,
    },
  });

  // Build a query to fetch each alert by its specific number
  const alertQueries = alertNumbers
    .map(
      (num, index) => `
    alert${index}: vulnerabilityAlert(number: ${num}) {
      number
      dependabotUpdate {
        pullRequest {
          number
        }
      }
    }
  `,
    )
    .join("");

  const query = `
    query GetSpecificVulnerabilityAlerts($owner: String!, $repo: String!) {
      repository(owner: $owner, name: $repo) {
        ${alertQueries}
      }
    }
  `;

  const result = await graphqlWithAuth<{
    repository: Record<string, VulnerabilityAlert | null>;
  }>(query, {
    owner,
    repo,
  });

  const mappings = new Map<number, number>();
  for (const [key, alert] of Object.entries(result.repository)) {
    if (alert && key.startsWith("alert") && alert.dependabotUpdate?.pullRequest?.number) {
      mappings.set(alert.number, alert.dependabotUpdate.pullRequest.number);
    }
  }

  return mappings;
}


export function matchesAnyPattern(
  manifestPath: string | undefined,
  patterns: string[],
): boolean {
  if (!manifestPath) return false;

  return patterns.some((pattern) => {
    try {
      return minimatch(manifestPath, pattern, {
        matchBase: true,
      });
    } catch (error) {
      console.error(
        `Error matching pattern ${pattern}:`,
        error instanceof Error ? error.message : String(error),
      );
      return false;
    }
  });
}

if (import.meta.main) {
  await run();
}
