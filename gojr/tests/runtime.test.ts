import { describe, expect, test } from "vitest";
import { evaluateSource } from "../src/index.js";

function expectRuns(source: string, options = {}) {
  const result = evaluateSource(source, options);
  expect(result.diagnostics).toEqual([]);
  return result;
}

describe("Go-junior runtime slice", () => {
  test("evaluates sheet arithmetic without number casts", () => {
    const result = expectRuns(`
var x = sheet.A1 + sheet.B1 * 2
if x > 10 {
  return x
}
return x * 2
`, {
      sheet: {
        A1: 4n,
        B1: 3n
      }
    });

    expect(result.value).toBe(20n);
    expect(result.values).toEqual([20n]);
  });

  test("resolves absolute refs, cross-sheet namespaces, and fmt aliases", () => {
    const result = expectRuns(`
import f "fmt"

f.Printf("A1=%v\\n", sheet.$A$1)
return Budget.B2 + sheet.$A$1
`, {
      sheet: {
        A1: 2n
      },
      sheets: {
        Budget: {
          B2: 5n
        }
      }
    });

    expect(result.output).toEqual(["A1=2\n"]);
    expect(result.value).toBe(7n);
  });

  test("evaluates spreadsheet ranges as row-major two-dimensional arrays", () => {
    const result = expectRuns(`
return sheet.A1:B2
`, {
      sheet: {
        A1: 1n,
        B1: 2n,
        A2: 3n,
        B2: 4n
      }
    });

    expect(result.value).toEqual([
      [1n, 2n],
      [3n, 4n]
    ]);
  });

  test("runs defers after the script body in reverse order", () => {
    const result = expectRuns(`
import "fmt"

defer fmt.Printf("third")
defer fmt.Printf(" second ")
fmt.Printf("first")
`);

    expect(result.output).toEqual(["first", " second ", "third"]);
  });

  test("reports panicOn failures as runtime diagnostics", () => {
    const result = evaluateSource(`
panicOn("bad")
`);

    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.code).toBe("GJPANIC001");
  });
});
