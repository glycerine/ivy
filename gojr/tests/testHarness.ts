import { deepStrictEqual, fail, notDeepStrictEqual } from "node:assert/strict";
import { describe, test } from "node:test";

export { describe, test };

type Constructor = Function & { readonly name: string };
type ExpectedAny = { readonly kind: "ExpectedAny"; readonly constructor: Constructor };

interface Matchers<T> {
  readonly not: Matchers<T>;
  toBe(expected: unknown): void;
  toEqual(expected: unknown): void;
  toContain(expected: unknown): void;
  toHaveLength(expected: number): void;
  toMatchObject(expected: unknown): void;
  toBeDefined(): void;
  toBeUndefined(): void;
  toBeNull(): void;
  toBeInstanceOf(expected: Constructor): void;
  toBeGreaterThan(expected: number): void;
  toBeNaN(): void;
  toThrow(expected?: RegExp): void;
}

interface Expect {
  <T>(actual: T): Matchers<T>;
  any(expected: Constructor): ExpectedAny;
}

export const expect: Expect = Object.assign(
  <T>(actual: T): Matchers<T> => makeMatchers(actual, false),
  {
    any(expected: Constructor): ExpectedAny {
      return { kind: "ExpectedAny", constructor: expected };
    }
  }
);

function makeMatchers<T>(actual: T, negate: boolean): Matchers<T> {
  const check = (passed: boolean, message: string): void => {
    if (negate ? passed : !passed) fail(message);
  };

  return {
    get not() {
      return makeMatchers(actual, !negate);
    },
    toBe(expected) {
      check(Object.is(actual, expected), `expected ${String(actual)} ${negate ? "not " : ""}to be ${String(expected)}`);
    },
    toEqual(expected) {
      if (containsExpectedAny(expected)) {
        check(matchesExact(actual, expected), "expected values to be deeply equal");
        return;
      }
      if (negate) {
        notDeepStrictEqual(actual, expected);
        return;
      }
      deepStrictEqual(actual, expected);
    },
    toContain(expected) {
      const passed = typeof actual === "string"
        ? actual.includes(String(expected))
        : Array.isArray(actual) && actual.includes(expected);
      check(passed, `expected ${String(actual)} ${negate ? "not " : ""}to contain ${String(expected)}`);
    },
    toHaveLength(expected) {
      const length = (actual as { length?: unknown }).length;
      check(length === expected, `expected length ${String(length)} ${negate ? "not " : ""}to be ${expected}`);
    },
    toMatchObject(expected) {
      check(matchesPartial(actual, expected), "expected object to match");
    },
    toBeDefined() {
      check(actual !== undefined, "expected value to be defined");
    },
    toBeUndefined() {
      check(actual === undefined, "expected value to be undefined");
    },
    toBeNull() {
      check(actual === null, "expected value to be null");
    },
    toBeInstanceOf(expected) {
      check(actual instanceof expected, `expected value ${negate ? "not " : ""}to be instance of ${expected.name}`);
    },
    toBeGreaterThan(expected) {
      check(typeof actual === "number" && actual > expected, `expected ${String(actual)} ${negate ? "not " : ""}to be greater than ${expected}`);
    },
    toBeNaN() {
      check(typeof actual === "number" && Number.isNaN(actual), `expected ${String(actual)} ${negate ? "not " : ""}to be NaN`);
    },
    toThrow(expected) {
      if (typeof actual !== "function") fail("expected value to be a function");
      let thrown: unknown;
      try {
        actual();
      } catch (error) {
        thrown = error;
      }
      const passed = thrown !== undefined && (!expected || expected.test(thrown instanceof Error ? thrown.message : String(thrown)));
      check(passed, `expected function ${negate ? "not " : ""}to throw`);
    }
  };
}

function containsExpectedAny(value: unknown): boolean {
  if (isExpectedAny(value)) return true;
  if (Array.isArray(value)) return value.some(containsExpectedAny);
  if (value && typeof value === "object") return Object.values(value).some(containsExpectedAny);
  return false;
}

function matchesExact(actual: unknown, expected: unknown): boolean {
  if (isExpectedAny(expected)) return matchesAny(actual, expected.constructor);
  if (Array.isArray(expected)) {
    return Array.isArray(actual) &&
      actual.length === expected.length &&
      expected.every((item, index) => matchesExact(actual[index], item));
  }
  if (expected && typeof expected === "object") {
    if (!actual || typeof actual !== "object" || Array.isArray(actual)) return false;
    const actualObject = actual as Record<string, unknown>;
    const expectedObject = expected as Record<string, unknown>;
    const actualKeys = Object.keys(actualObject);
    const expectedKeys = Object.keys(expectedObject);
    return actualKeys.length === expectedKeys.length &&
      expectedKeys.every((key) => matchesExact(actualObject[key], expectedObject[key]));
  }
  return Object.is(actual, expected);
}

function matchesPartial(actual: unknown, expected: unknown): boolean {
  if (isExpectedAny(expected)) return matchesAny(actual, expected.constructor);
  if (Array.isArray(expected)) {
    return Array.isArray(actual) &&
      expected.length <= actual.length &&
      expected.every((item, index) => matchesPartial(actual[index], item));
  }
  if (expected && typeof expected === "object") {
    if (!actual || typeof actual !== "object") return false;
    const actualObject = actual as Record<string, unknown>;
    const expectedObject = expected as Record<string, unknown>;
    return Object.keys(expectedObject).every((key) => matchesPartial(actualObject[key], expectedObject[key]));
  }
  return Object.is(actual, expected);
}

function isExpectedAny(value: unknown): value is ExpectedAny {
  return Boolean(value && typeof value === "object" && (value as ExpectedAny).kind === "ExpectedAny");
}

function matchesAny(actual: unknown, expected: Constructor): boolean {
  if (expected === Object) return actual !== null && typeof actual === "object";
  if (expected === String) return typeof actual === "string" || actual instanceof String;
  if (expected === Number) return typeof actual === "number" || actual instanceof Number;
  if (expected === Boolean) return typeof actual === "boolean" || actual instanceof Boolean;
  return actual instanceof (expected as new (...args: never[]) => unknown);
}
