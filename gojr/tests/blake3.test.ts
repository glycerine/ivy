import { describe, expect, test } from "./testHarness.js";
import { blake3HashBytes, blake3HashString } from "../src/index.js";

describe("Go-junior BLAKE3 cache hashing", () => {
  test("matches github.com/glycerine/blake3 33-byte URL-base64 vectors", () => {
    expect(blake3HashString("")).toBe("blake3.33B-rxNJufX5oaagQE3qNtzJSZvLJcmtwRK3zJqTyuQfMmLg");
    expect(blake3HashString("abc")).toBe("blake3.33B-ZDezrDhGUTP_tjt1JzqNtUjFWEZdedsD_TWcbNW9nYUf");
    expect(blake3HashString("hello")).toBe("blake3.33B-6o8WPbOGgpJeRJHF5Y1Ls1Bu-MFOt4qG6QjFYkpnIA_p");
    expect(blake3HashString("The quick brown fox jumps over the lazy dog")).toBe("blake3.33B-LxUUGBqtzNkTq9lM-lknAaVoarI_jfHf8bdHEP68bUrA");
    expect(blake3HashBytes(new Uint8Array(2048))).toBe("blake3.33B-viqN49z0bJTOhc3I4HrDCPTYqVSQ2VbDjXgP1hDbCBNs");
  });
});
