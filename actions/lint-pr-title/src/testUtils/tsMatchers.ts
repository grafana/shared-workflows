import { expect } from "bun:test";

/**
 * Jest-style matchers don't narrow types, so you can't take advantage of what
 * you know after an `expect()`. This function narrows the type of a potentially
 * `undefined` value.
 */
export function expect_toBeDefined<T>(arg: T): asserts arg is NonNullable<T> {
  expect(arg).toBeDefined();
}
