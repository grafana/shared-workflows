import * as mainModule from "./main";

import {
  afterEach,
  beforeAll,
  describe,
  expect,
  it,
  mock,
  spyOn,
} from "bun:test";

import { WrongEventTypeError } from "./main";
import { run } from ".";

async function setEventType(eventType: string) {
  await mock.module("@actions/github", () => ({
    context: {
      eventName: eventType,
    },
  }));
}

/**
 * Tests the "external" interface, the parts which rely on the environment - by
 * setting env vars and inputs.
 */
describe("lint", () => {
  beforeAll(() => {
    process.env.GITHUB_TOKEN = "token";
  });

  afterEach(() => {
    mock.restore();
  });

  it("rejects unknown event types", async () => {
    await setEventType("unknown");

    expect(run()).rejects.toThrow(WrongEventTypeError);
  });

  it("loads the specified config file", async () => {
    await mock.module("@actions/core", () => ({
      getInput: () => "/my-config.js",
    }));

    await mock.module("fs/promises", () => ({
      access: () => {},
    }));

    const spy = spyOn(mainModule, "loadConfig");

    expect(run()).rejects.toThrow();

    expect(spy).toHaveBeenCalledWith("/my-config.js");
  });
});
