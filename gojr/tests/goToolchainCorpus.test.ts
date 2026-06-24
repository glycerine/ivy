import { readFile } from "node:fs/promises";
import { join } from "node:path";
import { describe, expect, test } from "./testHarness.js";
import { evaluateSourceFiles } from "../src/index.js";
import type { EvaluationResult } from "../src/index.js";

const corpusRoot = join(process.cwd(), "_test", "go-toolchain", "test");

describe("Go toolchain corpus smoke tests", () => {
  test("executes selected Go distribution // run fixtures", async () => {
    const fixtures: Array<{ path: string; output?: string[] }> = [
      { path: "alias1.go", output: [] },
      { path: "bigmap.go", output: [] },
      { path: "align.go", output: [] },
      { path: "char_lit.go", output: [] },
      { path: "clear.go", output: [] },
      { path: "decl.go", output: [] },
      { path: "defer.go", output: [] },
      { path: "ddd.go", output: [] },
      { path: "divide.go", output: [] },
      { path: "floatcmp.go", output: [] },
      { path: "helloworld.go", output: ["hello, world\n"] },
      { path: "for.go", output: [] },
      { path: "closure1.go", output: [] },
      { path: "closure2.go", output: [] },
      { path: "compos.go", output: [] },
      { path: "const.go", output: [] },
      { path: "const3.go", output: [] },
      { path: "const8.go", output: [] },
      { path: "func.go", output: [] },
      { path: "func4.go", output: [] },
      { path: "func6.go", output: [] },
      { path: "func7.go", output: [] },
      { path: "func8.go", output: [] },
      { path: "if.go", output: [] },
      { path: "intcvt.go", output: [] },
      { path: "initcomma.go", output: [] },
      { path: "range3.go", output: [] },
      { path: "range4.go", output: [] },
      { path: "typeswitch1.go", output: [] },
      { path: "varinit.go", output: [] },
      { path: "mapclear.go", output: [] },
      { path: "map.go", output: [] },
      { path: "method3.go", output: [] },
      { path: "method7.go", output: [] },
      { path: "newexpr.go", output: [] },
      { path: "print.go" },
      { path: "string_lit.go", output: [] },
      { path: "abi/convF_criteria.go" },
      { path: "abi/convT64_criteria.go" },
      { path: "abi/defer_aggregate.go", output: [] },
      { path: "abi/double_nested_addressed_struct.go", output: [] },
      { path: "abi/double_nested_struct.go", output: [] },
      { path: "abi/f_ret_z_not.go" },
      { path: "abi/named_results.go" },
      { path: "iota.go", output: [] },
      { path: "literal.go", output: [] },
      { path: "ken/intervar.go", output: [] },
      { path: "ken/interbasic.go", output: [] },
      { path: "ken/interfun.go", output: [] },
      { path: "ken/array.go", output: [] },
      { path: "ken/complit.go", output: [] },
      { path: "ken/for.go", output: [] },
      { path: "ken/litfun.go", output: [] },
      { path: "ken/range.go", output: [] },
      { path: "ken/ptrvar.go", output: [] },
      { path: "ken/robfor.go", output: [] },
      { path: "ken/sliceslice.go", output: [] },
      { path: "ken/simparray.go", output: [] },
      { path: "ken/simpbool.go", output: [] },
      { path: "ken/simpconv.go", output: [] },
      { path: "ken/simpfun.go", output: [] },
      { path: "ken/strvar.go", output: [] },
      { path: "ken/string.go", output: ["abc", "xyz", "-", "abcxyz", "-", "abcxyz", "-", "abcxyz", "-", "abcxyz", "-", "abcxyz", "-", "abc", "xyz", "\n"] }
    ];

    for (const fixture of fixtures) {
      const result = await runFixtureMain(fixture.path);
      expect(nonExitSuccessDiagnostics(result)).toEqual([]);
      if (fixture.output) expect(result.output).toEqual(fixture.output);
    }
  });
});

function nonExitSuccessDiagnostics(result: EvaluationResult): EvaluationResult["diagnostics"] {
  return result.diagnostics.filter((diagnostic) => (
    diagnostic.code !== "GOJR_RUNTIME001" || diagnostic.message !== "os.Exit(0)"
  ));
}

async function runFixtureMain(relativePath: string): Promise<EvaluationResult> {
  const filename = join(corpusRoot, relativePath);
  const source = await readFile(filename, "utf8");
  return evaluateSourceFiles([{
    filename,
    source: `${source}\nmain()\n`
  }], {
    randomSeed: "gojr-go-toolchain-corpus"
  });
}
