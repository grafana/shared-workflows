const {
  describe,
  test,
  expect,
  spyOn,
  beforeEach,
  afterEach,
} = require("bun:test");
const crypto = require("node:crypto");
const fs = require("node:fs");

const {
  sha256Hex,
  getState,
  saveState,
  setOutput,
  retry,
  setSecret,
  fetchIdToken,
} = require("./lib.js");

describe("lib.js", () => {
  describe("sha256Hex", () => {
    test("computes correct sha256 hash matching bash echo -n | sha256sum", () => {
      expect(sha256Hex("hello")).toBe(
        "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
      );
    });
  });

  describe("getState", () => {
    afterEach(() => {
      delete process.env.STATE_MY_KEY;
    });

    test("reads from STATE_ prefix", () => {
      process.env.STATE_MY_KEY = "my_value";
      expect(getState("MY_KEY")).toBe("my_value");
    });

    test("returns empty string if unset", () => {
      expect(getState("MY_KEY")).toBe("");
    });
  });

  describe("retry", () => {
    test("succeeds on first try", async () => {
      let calls = 0;
      const result = await retry(
        { label: "test", attempts: 3, baseDelayMs: 1 },
        async () => {
          calls++;
          return "success";
        },
      );
      expect(result).toBe("success");
      expect(calls).toBe(1);
    });

    test("fails twice then succeeds", async () => {
      let calls = 0;
      const result = await retry(
        { label: "test", attempts: 3, baseDelayMs: 1 },
        async () => {
          calls++;
          if (calls < 3) throw new Error("fail");
          return "success";
        },
      );
      expect(result).toBe("success");
      expect(calls).toBe(3);
    });

    test("throws on all attempts failing", async () => {
      expect(
        retry({ label: "test", attempts: 3, baseDelayMs: 1 }, async () => {
          throw new Error("fatal failure");
        }),
      ).rejects.toThrow("fatal failure");
    });
  });

  describe("setSecret", () => {
    let logSpy;
    beforeEach(() => {
      logSpy = spyOn(console, "log").mockImplementation(() => {});
    });
    afterEach(() => {
      logSpy.mockRestore();
    });

    test("logs ::add-mask:: if value is truthy", () => {
      setSecret("top-secret");
      expect(logSpy).toHaveBeenCalledWith("::add-mask::top-secret");
    });

    test("does not log if value is falsy", () => {
      setSecret("");
      expect(logSpy).not.toHaveBeenCalled();
    });
  });
});
