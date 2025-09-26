import fs from "fs";
import yaml from "js-yaml";
import * as path from "path";

const HG_TOKEN = process.env.HG_TOKEN;
const APPS_YAML_FILE = path.join(
  process.cwd(),
  "./provisioning/plugins/apps.yaml",
);
const DATASOURCES_YAML_FILE = path.join(
  process.cwd(),
  "./provisioning/datasources/default.yaml",
);
// TODO check how to get the uid from the provisioning api or call /api/datasources instead
function getUid(dataSource, stackSlug) {
  let uid = "";
  switch (dataSource.type) {
    case "prometheus":
      if (dataSource.name === `grafanacloud-${stackSlug}-prom`) {
        uid = "grafanacloud-prom";
      }
      break;

    case "loki":
      if (dataSource.name === `grafanacloud-${stackSlug}-logs`) {
        uid = "grafanacloud-logs";
      }
      break;

    case "tempo":
      if (dataSource.name === `grafanacloud-${stackSlug}-traces`) {
        uid = "grafanacloud-traces";
      }
      break;

    case "alertmanager":
      if (dataSource.name === `grafanacloud-${stackSlug}-ngalertmanager`) {
        uid = "grafanacloud-ngalertmanager";
      }
      break;
  }
  return uid;
}

function formatDataSource(dataSource, stackSlug) {
  if (dataSource) {
    const uid = !dataSource.uid
      ? getUid(dataSource, stackSlug)
      : dataSource.uid;
    return {
      name: dataSource.name,
      type: dataSource.type,
      ...(uid && { uid }),
      url: dataSource.url,
      basicAuth: dataSource.basicAuth === 1 || dataSource.basicAuth === true,
      basicAuthUser: dataSource.basicAuthUser
        ? Number(dataSource.basicAuthUser)
        : undefined,
      isDefault: dataSource.isDefault === 1 || dataSource.isDefault === true,
      jsonData: dataSource.jsonData,
      secureJsonData: {
        basicAuthPassword: dataSource.basicAuthPassword,
      },
    };
  }
  return dataSource;
}

function removeEmptyProperties(obj) {
  // Check if the input is an object or an array
  if (Array.isArray(obj)) {
    // If it's an array, recursively clean each element
    return obj
      .map((item) => removeEmptyProperties(item))
      .filter((item) => item !== null && typeof item !== "undefined");
  }

  // Check if the input is a plain object
  if (typeof obj === "object" && obj !== null) {
    const newObj = {};
    for (const key in obj) {
      if (Object.prototype.hasOwnProperty.call(obj, key)) {
        const value = obj[key];

        // Recursively clean nested objects/arrays
        const cleanedValue = removeEmptyProperties(value);

        // Check for empty values and skip them
        if (
          cleanedValue !== "" &&
          cleanedValue !== null &&
          cleanedValue !== undefined &&
          !(Array.isArray(cleanedValue) && cleanedValue.length === 0) &&
          !(
            typeof cleanedValue === "object" &&
            Object.keys(cleanedValue).length === 0
          )
        ) {
          newObj[key] = cleanedValue;
        }
      }
    }
    return newObj;
  }

  // If the value is not an object or array, return it as is
  return obj;
}

function getBaseUrlByEnv(env) {
  switch (env) {
    case "prod-us-east":
      return "https://hg-api-prod-us-east-0.grafana.net";
    case "prod":
      return "https://hg-api-prod-us-central-0.grafana.net";
    case "ops":
      return "https://hg-api-ops-eu-south-0.grafana-ops.net";
    case "dev-east":
      return "https://hg-api-dev-us-east-0.grafana-dev.net";
    case "dev-central":
    default:
      return "https://hg-api-dev-us-central-0.grafana-dev.net";
  }
}

async function fetchMultipleAppConfigs(stackSlug, env, pluginIds) {
  try {
    const fetchPromises = pluginIds.map((pluginId) =>
      fetchAppConfig(stackSlug, env, pluginId),
    );
    return await Promise.all(fetchPromises);
  } catch (error) {
    console.error("Error fetching multiple app configs:", error.message);
    throw error;
  }
}

async function fetchAppConfig(stackSlug, env, pluginId) {
  try {
    const baseUrl = getBaseUrlByEnv(env);
    const url = `${baseUrl}/instances/${stackSlug}/provisioned-plugins/${pluginId}`;

    const response = await fetch(url, {
      headers: {
        "User-Agent": `plop/${pluginId}-provisioning`,
        Authorization: `Bearer ${HG_TOKEN}`,
      },
    });
    return response.json();
  } catch (error) {
    console.error("Error fetching app config", pluginId, ":", error.message);
    throw error;
  }
}

async function fetchMultipleDatasources(stackSlug, env, datasourceNames) {
  try {
    const fetchPromises = datasourceNames.map((dsName) =>
      fetchDataSource(stackSlug, env, dsName),
    );
    if (fetchPromises.length > 0) {
      return Promise.all(fetchPromises);
    }
    return Promise.all([]);
  } catch (error) {
    console.error("Error fetching multiple data sources:", error.message);
    throw error;
  }
}

async function fetchDataSource(stackSlug, env, datasourceName) {
  try {
    const baseUrl = getBaseUrlByEnv(env);
    const url = `${baseUrl}/instances/${stackSlug}/datasources/${datasourceName}`;
    const response = await fetch(url, {
      headers: {
        "User-Agent": `plop/${datasourceName}-provisioning`,
        Authorization: `Bearer ${HG_TOKEN}`,
      },
    });
    const dataSourceWithToken = await response.json();
    const dataSourceWithNoEmptyField =
      removeEmptyProperties(dataSourceWithToken);
    return formatDataSource(dataSourceWithNoEmptyField, stackSlug);
  } catch (error) {
    console.error(
      "Error fetching datasource",
      datasourceName,
      ":",
      error.message,
    );
    throw error;
  }
}

function createDataSourcesYamlFile() {
  const dir = path.dirname(DATASOURCES_YAML_FILE);

  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true });
  }

  const initialContent = {
    apiVersion: 1,
    prune: true,

    datasources: [],
  };
  fs.writeFileSync(DATASOURCES_YAML_FILE, yaml.dump(initialContent));
  return initialContent;
}

async function fetchGrafanaConfig(stackSlug, env, pluginId) {
  try {
    const baseUrl = getBaseUrlByEnv(env);
    const url = `${baseUrl}/instances/${stackSlug}/config`;
    const response = await fetch(url, {
      headers: {
        "User-Agent": `plop/${pluginId}-provisioning`,
        Authorization: `Bearer ${HG_TOKEN}`,
      },
    });
    return response.json();
  } catch (error) {
    console.error("Error fetching gcom token:", error.message);
    throw error;
  }
}

function createAppsYamlFile() {
  const dir = path.dirname(APPS_YAML_FILE);

  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true });
  }

  const initialContent = {
    apiVersion: 1,

    apps: [],
  };
  fs.writeFileSync(APPS_YAML_FILE, yaml.dump(initialContent));
  return initialContent;
}

function addAppConfigs(yamlData, appConfigs) {
  appConfigs.forEach((appConfig) => {
    if (appConfig.type === "grafana-asserts-app") {
      appConfig.jsonData.instanceUrl = "http://localhost:3000";
    }
    yamlData.apps.push(appConfig);
    console.log(`App with type '${appConfig.type}' has been added`);
  });
}

function addDataSourceConfigs(yamlData, dataSourceConfigs = []) {
  dataSourceConfigs.forEach((dsConfig, i) => {
    yamlData.datasources.push(dsConfig);
    console.log(
      `Data source with type '${dsConfig.type}' and name '${dsConfig.name}' has been added`,
    );
  });
}

function writeDataSourcesYamlFile(yamlData) {
  const yamlString = yaml.dump(yamlData);
  fs.writeFileSync(DATASOURCES_YAML_FILE, yamlString);
  console.log("default.yaml data source file has been updated.");
}

function writeAppsYamlFile(yamlData) {
  const yamlString = yaml.dump(yamlData);

  // just for asserts
  const fixed = yamlString.replace(
    "enableGrafanaManagedLLM: true",
    "enableGrafanaManagedLLM: false",
  );
  fs.writeFileSync(APPS_YAML_FILE, fixed);
  console.log(
    "apps.yaml plugins file has been updated. Asserts prop enableGrafanaManagedLLM was disabled",
  );
}

async function fillAnswers(answers) {
  const appConfigs = await fetchMultipleAppConfigs(
    answers.STACK_SLUG,
    answers.ENV,
    answers.PLUGIN_IDS,
  );
  const yamlAppsData = createAppsYamlFile();
  addAppConfigs(yamlAppsData, appConfigs);
  writeAppsYamlFile(yamlAppsData);

  const dataSourceConfigs = await fetchMultipleDatasources(
    answers.STACK_SLUG,
    answers.ENV,
    answers.DATASOURCE_IDS,
  );
  const yamlDataSourcesData = createDataSourcesYamlFile();
  addDataSourceConfigs(yamlDataSourcesData, dataSourceConfigs);
  writeDataSourcesYamlFile(yamlDataSourcesData);

  const grafanaConfig = await fetchGrafanaConfig(
    answers.STACK_SLUG,
    answers.ENV,
    answers.GF_PLUGIN_ID,
  );
  answers.GF_GRAFANA_COM_SSO_API_TOKEN =
    grafanaConfig.hosted_grafana.hg_auth_token;

  // Use hardcoded URL for ops stack when grafana_net.url is missing
  const grafanaNetUrl =
    answers.STACK_SLUG === "ops" && !grafanaConfig.grafana_net?.url
      ? "https://grafana-ops.com"
      : grafanaConfig.grafana_net.url;

  answers.GF_GRAFANA_COM_URL = grafanaNetUrl;
  answers.GF_GRAFANA_COM_API_URL = `${grafanaNetUrl}/api`;
  answers.GF_PLUGINS_PREINSTALL_SYNC = answers.PLUGIN_IDS.filter(
    (p) => p !== answers.GF_PLUGIN_ID,
  ).join(",");
}

export default function (plop) {
  plop.setHelper("env", (text) => process.env[text]);

  plop.setGenerator("e2e-testing-provisioning", {
    prompts: [],
    actions: [
      async function loadRemoteProvisioning(answers) {
        try {
          if (!HG_TOKEN) {
            console.error("HG_TOKEN environment variable is not set.");
            process.exit(1);
          }

          if (!process.env.E2E_STACK_SLUG) {
            console.error("E2E_STACK_SLUG environment variable is not set.");
            process.exit(1);
          }

          if (!process.env.E2E_PLUGIN_ID) {
            console.error("E2E_PLUGIN_ID environment variable is not set.");
            process.exit(1);
          }
          answers.STACK_SLUG = process.env.E2E_STACK_SLUG;
          answers.ENV = process.env.E2E_ENV;

          answers.GF_PLUGIN_ID = process.env.E2E_PLUGIN_ID;
          const otherPlugins = process.env.E2E_OTHER_PLUGINS
            ? process.env.E2E_OTHER_PLUGINS.split(",").map((i) => i.trim())
            : [];
          answers.PLUGIN_IDS = [process.env.E2E_PLUGIN_ID].concat(otherPlugins);

          answers.DATASOURCE_IDS = process.env.E2E_DATASOURCE_IDS
            ? process.env.E2E_DATASOURCE_IDS.split(",").map((i) => i.trim())
            : [];
          answers.GF_PLUGINS_PREINSTALL_SYNC = otherPlugins.join(",");

          await fillAnswers(answers);

          return "Remote Provisioning data loaded successfully for e2e tests.";
        } catch (error) {
          console.error("Failed to load Remote Provisioning:", error.message);
          return "Failed to load Remote Provisioning data";
        }
      },
      {
        type: "add",
        path: "./docker-compose.yaml",
        templateFile: "plop-templates/docker-compose.hbs.yaml",
        force: true,
      },
    ],
  });
}
