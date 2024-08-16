import { Context } from "@actions/github/lib/context";
import { WebhookPayload } from "@actions/github/lib/interfaces";
import { Schema as WebhookSchema } from "@octokit/webhooks-types";

/**
 * Replace all `null` values in a type with `unknown`. The schema has specified
 * `null` for many fields where a value may be present but shouldn't be used.
 * Since the value is present, that means that real payloads don't satisfy the
 * interface. We can replace `null` with `unknown` to make the type more
 * permissive and allow the object to be typed.
 */
type DeepReplaceNullWithUnknown<T> = T extends null
  ? unknown
  : T extends Array<infer U>
    ? Array<DeepReplaceNullWithUnknown<U>>
    : T extends object
      ? { [K in keyof T]: DeepReplaceNullWithUnknown<T[K]> }
      : T;

/**
 * Allow extra fields in an object. This is useful for payloads where some extra
 * fields may be present that aren't in the schema. In that case the type
 * checker reports an `Object literal may only specify known properties` error.
 * This type allows such types to have extra fields.
 *
 */
type DeepAllowExtraFields<T> =
  T extends Array<infer U>
    ? Array<DeepAllowExtraFields<U>>
    : T extends object
      ? { [K in keyof T]: DeepAllowExtraFields<T[K]> } & {
          [key: string]: unknown;
        }
      : T;

/**
 * A combination of `DeepReplaceNullWithUnknown` and `DeepAllowExtraFields` to
 * allow real event payloads to be type checked. Use `satisfies` rather than a
 * type annotation so that the rest of the program uses the actual type, not
 * this one, which should allow the variable to be passed to other APIs.
 */
type RealWorldEventPayload<T extends WebhookSchema> = DeepAllowExtraFields<
  DeepReplaceNullWithUnknown<T>
>;

export interface GitHubPayload<T extends WebhookSchema> {
  [key: string]: unknown;
  action: string;
  api_url: string;
  event: RealWorldEventPayload<T>;
  event_name: string;
  graphql_url: string;
  job: string;
  repository: string;
  run_id: string;
  run_number: string;
  server_url: string;
  workflow: string;
}

/**
 * A temporary environment for use in tests. This class sets environment
 * variables in the constructor and unsets them with `dispose`. If the object is
 * instansiated with `using`, the dispose function will be called when the
 * object goes out of scope.
 */
class TempEnvironment implements Disposable {
  private readonly oldEnv: Record<string, string | undefined>;

  constructor(readonly env: Record<string, string>) {
    this.oldEnv = process.env;
    for (const key in env) {
      this.oldEnv[key] = process.env[key];
      process.env[key] = env[key];
    }
  }

  [Symbol.dispose]() {
    for (const [key, value] of Object.entries(this.oldEnv)) {
      if (value === undefined) {
        //eslint-disable-next-line @typescript-eslint/no-dynamic-delete
        delete process.env[key];
        return;
      }
      process.env[key] = value;
    }
  }
}

export function newContext<T extends WebhookSchema>(
  payload: GitHubPayload<T>,
  sha: string,
  ref: string,
): Context {
  // To create a new `Context()`, we need to set and then unset the various
  // `GITHUB_` environment variables. That makes this function non-thread-safe.
  using _env = new TempEnvironment({
    GITHUB_EVENT_NAME: payload.event_name,
    GITHUB_SHA: sha,
    GITHUB_REF: ref,
    GITHUB_WORKFLOW: payload.workflow,
    GITHUB_ACTION: payload.action,
    GITHUB_JOB: payload.job,
    GITHUB_RUN_NUMBER: payload.run_number,
    GITHUB_RUN_ID: payload.run_id,
    GITHUB_API_URL: payload.api_url,
    GITHUB_GRAPHQL_URL: payload.graphql_url,
    GITHUB_SERVER_URL: payload.server_url,
  });

  const ctx = new Context();
  ctx.payload = payload.event as WebhookPayload;

  return ctx;
}
