import { Octokit } from "@octokit/rest";
import { minimatch } from "minimatch";
import { RequestError } from "@octokit/request-error";

async function run() {
  try {
    const token = process.env.GITHUB_TOKEN;
    const alertTypes = (process.env.INPUT_ALERT_TYPES || "dependency").split(",").map(t => t.trim());
    // Parse path patterns from multi-line format, filtering out empty lines
    const pathPatterns = (process.env.INPUT_PATHS || "").split("\n")
      .map(p => p.trim())
      .filter(p => p !== "");
      
    const dismissalComment = process.env.INPUT_DISMISSAL_COMMENT;
    const dismissalReason = process.env.INPUT_DISMISSAL_REASON;
    
    if (!token) {
      throw new Error("Missing required env var GITHUB_TOKEN");
    }

    if (pathPatterns.length === 0) {
      throw new Error("No path patterns provided. Please specify paths to match.");
    }
    
    const octokit = new Octokit({ auth: token });
    const [owner, repo] = (process.env.GITHUB_REPOSITORY || "/").split("/");
    
    if (!owner || !repo) {
      throw new Error("Could not determine repository owner and name from GITHUB_REPOSITORY");
    }

    // Check token permissions first
    console.log("Checking GitHub token permissions...");
    try {
      const { data: tokenData } = await octokit.request("GET /repos/{owner}/{repo}", {
        owner,
        repo,
      });
      console.log(`Successfully authenticated with GitHub as ${tokenData.owner?.login || "unknown user"}`);
    } catch (error) {
      console.error("Error authenticating with GitHub. Please check your token.");
      throw error;
    }
    
    console.log(`Fetching dependabot alerts for ${owner}/${repo}...`);
    console.log(`Using path patterns:`);
    pathPatterns.forEach(pattern => console.log(`- ${pattern}`));
    
    try {
      // Get all open alerts using the correct endpoint
      const alerts = await fetchAllAlerts(octokit, owner, repo, alertTypes);
      console.log(`Found ${alerts.length} open alerts`);
      
      if (alerts.length === 0) {
        console.log("No alerts found to process.");
        return;
      }
      
      // Process matching alerts
      const alertsToProcess = [];
      for (const alert of alerts) {
        const manifestPath = alert.dependency?.manifest_path;
        if (manifestPath && matchesAnyPattern(manifestPath, pathPatterns)) {
          console.log(`Alert #${alert.number} for ${alert.dependency.package.name} in ${manifestPath} matches patterns`);
          alertsToProcess.push(alert.number);
        } else {
          console.log(`Skipping alert #${alert.number} for ${manifestPath || "unknown path"} (does not match any pattern)`);
        }
      }
      
      if (alertsToProcess.length === 0) {
        console.log("No alerts matched the provided path patterns.");
        return;
      }
      
      console.log(`Dismissing ${alertsToProcess.length} alerts...`);
      
      // Use the correct endpoint to dismiss multiple alerts at once
      try {
        for (const alertNumber of alertsToProcess) {
          await octokit.request('PATCH /repos/{owner}/{repo}/dependabot/alerts/{alert_number}', {
            owner,
            repo,
            alert_number: alertNumber,
            state: 'dismissed',
            dismissed_reason: dismissalReason as "fix_started" | "inaccurate" | "no_bandwidth" | "not_used" | "tolerable_risk" | undefined,
            dismissed_comment: dismissalComment
          });
          console.log(`Alert #${alertNumber} dismissed successfully`);
        }
        console.log(`Successfully dismissed ${alertsToProcess.length} alerts.`);
      } catch (error) {
        if (error instanceof RequestError) {
          if (error.status === 403) {
            console.error(`Error: Permission denied when trying to dismiss alerts.`);
            console.error(`Make sure the GITHUB_TOKEN has 'security-events: write' permission.`);
            process.exit(1);
          } else {
            console.error(`API Error ${error.status}: ${error.message}`);
            process.exit(1);
          }
        } else {
          console.error("Error dismissing alerts:", error instanceof Error ? error.message : String(error));
          process.exit(1);
        }
      }
    } catch (error) {
      if (error instanceof RequestError) {
        if (error.status === 404) {
          console.error(`Error: Repository ${owner}/${repo} not found or Dependabot alerts are not enabled.`);
          console.error("Make sure Dependabot alerts are enabled in the repository's Security settings.");
          process.exit(1);
        } else if (error.status === 403 || error.message.includes("Resource not accessible by integration")) {
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
        console.error("Error:", error instanceof Error ? error.message : String(error));
        process.exit(1);
      }
    }
  } catch (error) {
    console.error("Error:", error instanceof Error ? error.message : String(error));
    process.exit(1);
  }
}

async function fetchAllAlerts(octokit: Octokit, owner: string, repo: string, alertTypes: string[]) {
  const alerts = [];
  let page = 1;
  let hasMorePages = true;
  
  try {
    while (hasMorePages) {
      // Using the dependabot alerts endpoint directly
      const response = await octokit.request('GET /repos/{owner}/{repo}/dependabot/alerts', {
        owner,
        repo,
        state: 'open',
        per_page: 100,
        page
      });
      
      if (response.data.length === 0) {
        hasMorePages = false;
      } else {
        const filteredAlerts = response.data.filter(alert => 
          alertTypes.includes(alert.security_advisory?.severity as string) || 
          alertTypes.includes(alert.security_vulnerability?.severity as string) ||
          alertTypes.includes("dependency")
        );
        alerts.push(...filteredAlerts);
        page++;
      }
    }
    
    return alerts;
  } catch (error) {
    // Re-throw to be handled by the main function
    throw error;
  }
}

function matchesAnyPattern(manifestPath: string | undefined, patterns: string[]): boolean {
  if (!manifestPath) return false;
  
  return patterns.some(pattern => {
    try {
      return minimatch(manifestPath, pattern, { matchBase: true });
    } catch (error) {
      console.error(`Error matching pattern ${pattern}:`, error instanceof Error ? error.message : String(error));
      return false;
    }
  });
}

run();