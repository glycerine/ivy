import { describe, expect, test } from "vitest";
import { parseRuntimeJson, parseSheetJson } from "../src/index.js";

describe("CLI JSON runtime input", () => {
  test("keeps quoted values as strings", () => {
    const sheet = parseSheetJson('{"A1":"40n","A2":"40","A3":"40.0"}');

    expect(sheet.A1).toBe("40n");
    expect(sheet.A2).toBe("40");
    expect(sheet.A3).toBe("40.0");
  });

  test("parses integer number tokens as exact integers", () => {
    const sheet = parseSheetJson('{"A1":40,"A2":9223372036854775808}');

    expect(sheet.A1).toBe(40n);
    expect(sheet.A2).toBe(9223372036854775808n);
  });

  test("parses decimal and exponent number tokens as float64", () => {
    const sheet = parseSheetJson('{"A1":40.0,"A2":4e1,"A3":-2.5}');

    expect(sheet.A1).toBe(40);
    expect(sheet.A2).toBe(40);
    expect(sheet.A3).toBe(-2.5);
  });

  test("rejects invalid JSON numbers", () => {
    expect(() => parseRuntimeJson('{"A1":040}')).toThrow(/leading zero/);
  });
});
