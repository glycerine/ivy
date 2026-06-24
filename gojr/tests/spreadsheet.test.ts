import { describe, expect, test } from "./testHarness.js";
import {
  SpreadsheetEngine,
  SpreadsheetFormulaCompilerCache,
  cellDependency,
  rangeDependency,
  spreadsheetFormulaCacheKey
} from "../src/spreadsheet.js";
import type { SpreadsheetCellRef, SpreadsheetFormulaCompileInput, SpreadsheetFormulaContext } from "../src/spreadsheet.js";

const S = "sheet";

function ref(cell: string, sheet = S): SpreadsheetCellRef {
  return { sheet, cell };
}

describe("Go-junior spreadsheet dependency engine", () => {
  test("recalculates formulas from literal dependencies", async () => {
    const engine = new SpreadsheetEngine();
    let evaluations = 0;

    engine.setLiteral(ref("A1"), 2);
    engine.setFormula(ref("B1"), (ctx) => {
      evaluations++;
      return Number(ctx.cell("A1")) + 1;
    }, { declaredDeps: [cellDependency(ref("A1"))] });

    expect(await engine.recalculate()).toMatchObject({ unstable: false });
    expect(engine.value(ref("B1"))).toBe(3);
    expect(evaluations).toBe(1);
    expect(engine.observedDeps(ref("B1"))).toEqual([cellDependency(ref("A1"))]);
    expect(engine.reverseDependentsOf(cellDependency(ref("A1")))).toEqual([ref("B1")]);

    engine.setLiteral(ref("A1"), 4);
    await engine.recalculate();

    expect(engine.value(ref("B1"))).toBe(5);
    expect(evaluations).toBe(2);
  });

  test("evaluates dirty diamond graphs once in dependency order", async () => {
    const engine = new SpreadsheetEngine();
    const order: string[] = [];

    engine.setLiteral(ref("A1"), 10);
    engine.setFormula(ref("B1"), (ctx) => {
      order.push("B1");
      return Number(ctx.cell("A1")) + 1;
    }, { declaredDeps: [cellDependency(ref("A1"))] });
    engine.setFormula(ref("C1"), (ctx) => {
      order.push("C1");
      return Number(ctx.cell("A1")) + 2;
    }, { declaredDeps: [cellDependency(ref("A1"))] });
    engine.setFormula(ref("D1"), (ctx) => {
      order.push("D1");
      return Number(ctx.cell("B1")) + Number(ctx.cell("C1"));
    }, { declaredDeps: [cellDependency(ref("B1")), cellDependency(ref("C1"))] });

    await engine.recalculate();
    expect(engine.value(ref("D1"))).toBe(23);
    expect(order).toEqual(["B1", "C1", "D1"]);

    order.length = 0;
    engine.setLiteral(ref("A1"), 20);
    await engine.recalculate();

    expect(engine.value(ref("D1"))).toBe(43);
    expect(order).toEqual(["B1", "C1", "D1"]);
  });

  test("moves dynamic observed dependency edges when a branch changes", async () => {
    const engine = new SpreadsheetEngine();
    let evaluations = 0;

    engine.setLiteral(ref("A1"), true);
    engine.setLiteral(ref("B1"), 10);
    engine.setLiteral(ref("C1"), 100);
    engine.setFormula(ref("D1"), (ctx) => {
      evaluations++;
      return ctx.cell("A1") ? ctx.cell("B1") : ctx.cell("C1");
    }, { declaredDeps: [cellDependency(ref("A1")), cellDependency(ref("B1")), cellDependency(ref("C1"))] });

    await engine.recalculate();
    expect(engine.value(ref("D1"))).toBe(10);
    expect(evaluations).toBe(2);
    expect(engine.observedDeps(ref("D1"))).toEqual([cellDependency(ref("A1")), cellDependency(ref("B1"))]);

    engine.setLiteral(ref("C1"), 101);
    await engine.recalculate();
    expect(engine.value(ref("D1"))).toBe(10);
    expect(evaluations).toBe(2);

    engine.setLiteral(ref("A1"), false);
    await engine.recalculate();
    expect(engine.value(ref("D1"))).toBe(101);
    expect(engine.observedDeps(ref("D1"))).toEqual([cellDependency(ref("A1")), cellDependency(ref("C1"))]);
    expect(engine.reverseDependentsOf(cellDependency(ref("B1")))).toEqual([]);
    expect(engine.reverseDependentsOf(cellDependency(ref("C1")))).toEqual([ref("D1")]);

    engine.setLiteral(ref("B1"), 11);
    await engine.recalculate();
    expect(engine.value(ref("D1"))).toBe(101);
    expect(evaluations).toBe(4);

    engine.setLiteral(ref("C1"), 102);
    await engine.recalculate();
    expect(engine.value(ref("D1"))).toBe(102);
    expect(evaluations).toBe(5);
  });

  test("tracks cross-sheet cell dependencies", async () => {
    const engine = new SpreadsheetEngine();

    engine.setLiteral(ref("A1", "Data"), 5);
    engine.setFormula(ref("A1"), (ctx) => Number(ctx.cell("Data", "A1")) * 2, {
      declaredDeps: [cellDependency(ref("A1", "Data"))]
    });

    await engine.recalculate();
    expect(engine.value(ref("A1"))).toBe(10);
    expect(engine.observedDeps(ref("A1"))).toEqual([cellDependency(ref("A1", "Data"))]);

    engine.setLiteral(ref("A1", "Data"), 7);
    await engine.recalculate();
    expect(engine.value(ref("A1"))).toBe(14);
  });

  test("tracks range dependencies and dirties formulas when an included cell changes", async () => {
    const engine = new SpreadsheetEngine();
    let evaluations = 0;

    engine.setLiteral(ref("A1"), 1);
    engine.setLiteral(ref("B1"), 2);
    engine.setLiteral(ref("A2"), 3);
    engine.setLiteral(ref("B2"), 4);
    engine.setFormula(ref("C1"), (ctx) => {
      evaluations++;
      return ctx.range("A1", "B2").flat().reduce<number>((sum, value) => sum + Number(value), 0);
    }, { declaredDeps: [rangeDependency(S, "A1", "B2")] });

    await engine.recalculate();
    expect(engine.value(ref("C1"))).toBe(10);
    expect(evaluations).toBe(1);

    engine.setLiteral(ref("B2"), 10);
    await engine.recalculate();
    expect(engine.value(ref("C1"))).toBe(16);
    expect(evaluations).toBe(2);
  });

  test("reports dependency cycles when iterative calculation is disabled", async () => {
    const engine = new SpreadsheetEngine();

    engine.setFormula(ref("A1"), (ctx) => Number(ctx.cell("B1")) + 1, {
      declaredDeps: [cellDependency(ref("B1"))]
    });
    engine.setFormula(ref("B1"), (ctx) => Number(ctx.cell("A1")) + 1, {
      declaredDeps: [cellDependency(ref("A1"))]
    });

    const result = await engine.recalculate();
    expect(result.diagnostics.map((diagnostic) => diagnostic.code)).toEqual(["#CYCLE!", "#CYCLE!"]);
    expect((engine.value(ref("A1")) as { code: string }).code).toBe("#CYCLE!");
    expect((engine.value(ref("B1")) as { code: string }).code).toBe("#CYCLE!");
  });

  test("reports self-reference cycles when iterative calculation is disabled", async () => {
    const engine = new SpreadsheetEngine();

    engine.setFormula(ref("A1"), (ctx) => Number(ctx.cell("A1")) + 1, {
      declaredDeps: [cellDependency(ref("A1"))]
    });

    await engine.recalculate();
    expect(engine.diagnostics(ref("A1"))[0]?.code).toBe("#CYCLE!");
  });

  test("converges opted-in iterative cycles with numeric tolerance", async () => {
    const engine = new SpreadsheetEngine({
      iterativeCalculation: {
        enabled: true,
        maxIterations: 20,
        numericTolerance: 0.001
      }
    });

    engine.setFormula(ref("A1"), (ctx) => (Number(ctx.cell("A1") ?? 0) + 1) / 2, {
      declaredDeps: [cellDependency(ref("A1"))]
    });

    const result = await engine.recalculate();
    expect(result.diagnostics).toEqual([]);
    expect(Math.abs(Number(engine.value(ref("A1"))) - 1) <= 0.001).toBe(true);
  });

  test("reports iterative divergence and oscillation separately", async () => {
    const diverging = new SpreadsheetEngine({
      iterativeCalculation: { enabled: true, maxIterations: 3, detectOscillation: true }
    });
    diverging.setFormula(ref("A1"), (ctx) => Number(ctx.cell("A1") ?? 0) + 1, {
      declaredDeps: [cellDependency(ref("A1"))]
    });

    await diverging.recalculate();
    expect(diverging.diagnostics(ref("A1"))[0]?.code).toBe("#DIVERGE!");

    const oscillating = new SpreadsheetEngine({
      iterativeCalculation: { enabled: true, maxIterations: 10, detectOscillation: true }
    });
    oscillating.setFormula(ref("A1"), (ctx) => ctx.cell("A1") === 1 ? 0 : 1, {
      declaredDeps: [cellDependency(ref("A1"))]
    });

    await oscillating.recalculate();
    expect(oscillating.diagnostics(ref("A1"))[0]?.code).toBe("#OSCILLATE!");
  });

  test("reports unstable acyclic dynamic dependency sets", async () => {
    const engine = new SpreadsheetEngine({ dependencyStabilizationLimit: 3 });
    let flip = false;

    engine.setLiteral(ref("B1"), 1);
    engine.setLiteral(ref("C1"), 2);
    engine.setFormula(ref("D1"), (ctx) => {
      flip = !flip;
      return flip ? ctx.cell("B1") : ctx.cell("C1");
    });

    const result = await engine.recalculate();
    expect(result.unstable).toBe(true);
    expect(engine.diagnostics(ref("D1"))[0]?.code).toBe("#UNSTABLE_DEPS!");
  });

  test("removing a formula removes old reverse dependency edges", async () => {
    const engine = new SpreadsheetEngine();

    engine.setLiteral(ref("A1"), 1);
    engine.setFormula(ref("B1"), (ctx) => Number(ctx.cell("A1")) + 1, {
      declaredDeps: [cellDependency(ref("A1"))]
    });
    await engine.recalculate();
    expect(engine.reverseDependentsOf(cellDependency(ref("A1")))).toEqual([ref("B1")]);

    engine.removeCell(ref("B1"));
    expect(engine.reverseDependentsOf(cellDependency(ref("A1")))).toEqual([]);
  });

  test("skips unchanged formula installs by cache key and keeps dependency edges", async () => {
    const engine = new SpreadsheetEngine();
    const deps = [cellDependency(ref("A1"))];
    const cacheKey = spreadsheetFormulaCacheKey({
      filename: "sheet!B1.gojr",
      source: "sheet.A1.(int64) + 1",
      declaredDeps: deps
    });
    let evaluations = 0;

    engine.setLiteral(ref("A1"), 1);
    engine.setFormula(ref("B1"), (ctx) => {
      evaluations++;
      return Number(ctx.cell("A1")) + 1;
    }, { declaredDeps: deps, cacheKey });

    await engine.recalculate();
    expect(engine.value(ref("B1"))).toBe(2);
    expect(evaluations).toBe(1);

    const skipped = engine.setFormula(ref("B1"), () => {
      evaluations += 1000;
      return -1;
    }, { declaredDeps: deps, cacheKey });

    expect(skipped.action).toBe("skipped");
    expect(engine.formulaCacheKey(ref("B1"))).toBe(cacheKey);
    await engine.recalculate();
    expect(engine.value(ref("B1"))).toBe(2);
    expect(evaluations).toBe(1);

    engine.setLiteral(ref("A1"), 2);
    await engine.recalculate();
    expect(engine.value(ref("B1"))).toBe(3);
    expect(evaluations).toBe(2);
    expect(engine.reverseDependentsOf(cellDependency(ref("A1")))).toEqual([ref("B1")]);
  });

  test("reinstalls changed formula cache keys and dirties downstream formulas", async () => {
    const engine = new SpreadsheetEngine();
    const deps = [cellDependency(ref("A1"))];
    const firstKey = spreadsheetFormulaCacheKey({
      filename: "sheet!B1.gojr",
      source: "sheet.A1.(int64) + 1",
      declaredDeps: deps
    });
    const secondKey = spreadsheetFormulaCacheKey({
      filename: "sheet!B1.gojr",
      source: "sheet.A1.(int64) + 10",
      declaredDeps: deps
    });
    const order: string[] = [];

    engine.setLiteral(ref("A1"), 1);
    engine.setFormula(ref("B1"), (ctx) => {
      order.push("B1:first");
      return Number(ctx.cell("A1")) + 1;
    }, { declaredDeps: deps, cacheKey: firstKey });
    engine.setFormula(ref("C1"), (ctx) => {
      order.push("C1");
      return Number(ctx.cell("B1")) * 2;
    }, { declaredDeps: [cellDependency(ref("B1"))] });

    await engine.recalculate();
    expect(engine.value(ref("C1"))).toBe(4);
    order.length = 0;

    const installed = engine.setFormula(ref("B1"), (ctx) => {
      order.push("B1:second");
      return Number(ctx.cell("A1")) + 10;
    }, { declaredDeps: deps, cacheKey: secondKey });

    expect(installed.action).toBe("installed");
    await engine.recalculate();
    expect(engine.value(ref("B1"))).toBe(11);
    expect(engine.value(ref("C1"))).toBe(22);
    expect(order).toEqual(["B1:second", "C1"]);
  });

  test("compiler cache separates source-shape recompilation from dependency value recalculation", async () => {
    const engine = new SpreadsheetEngine();
    const cache = new SpreadsheetFormulaCompilerCache();
    const deps = [cellDependency(ref("A1"))];
    const compiled: string[] = [];
    const evaluated: string[] = [];
    const compiler = async (input: SpreadsheetFormulaCompileInput) => {
      compiled.push(input.source);
      const increment = input.source.includes("+ 10") ? 10 : 1;
      return (ctx: SpreadsheetFormulaContext) => {
        evaluated.push(input.source);
        return Number(ctx.cell("A1")) + increment;
      };
    };

    engine.setLiteral(ref("A1"), 1);
    const first = await cache.install(engine, ref("B1"), {
      filename: "sheet!B1.gojr",
      source: "sheet.A1.(int64) + 1",
      declaredDeps: deps,
      compiler
    });
    expect(first).toMatchObject({ action: "installed", compileAction: "compiled" });

    await engine.recalculate();
    expect(engine.value(ref("B1"))).toBe(2);
    expect(compiled).toEqual(["sheet.A1.(int64) + 1"]);
    expect(evaluated).toEqual(["sheet.A1.(int64) + 1"]);

    engine.setLiteral(ref("A1"), 2);
    const unchanged = await cache.install(engine, ref("B1"), {
      filename: "sheet!B1.gojr",
      source: "sheet.A1.(int64) + 1",
      declaredDeps: deps,
      compiler: async () => {
        throw new Error("unchanged formula should not compile");
      }
    });
    expect(unchanged).toMatchObject({ action: "skipped", compileAction: "skipped" });

    await engine.recalculate();
    expect(engine.value(ref("B1"))).toBe(3);
    expect(compiled).toEqual(["sheet.A1.(int64) + 1"]);
    expect(evaluated).toEqual(["sheet.A1.(int64) + 1", "sheet.A1.(int64) + 1"]);

    const changed = await cache.install(engine, ref("B1"), {
      filename: "sheet!B1.gojr",
      source: "sheet.A1.(int64) + 10",
      declaredDeps: deps,
      compiler
    });
    expect(changed).toMatchObject({ action: "installed", compileAction: "compiled" });
    await engine.recalculate();
    expect(engine.value(ref("B1"))).toBe(12);
    expect(compiled).toEqual(["sheet.A1.(int64) + 1", "sheet.A1.(int64) + 10"]);

    const reverted = await cache.install(engine, ref("B1"), {
      filename: "sheet!B1.gojr",
      source: "sheet.A1.(int64) + 1",
      declaredDeps: deps,
      compiler: async () => {
        throw new Error("previously compiled formula should be reused");
      }
    });
    expect(reverted).toMatchObject({ action: "installed", compileAction: "reused" });
    await engine.recalculate();
    expect(engine.value(ref("B1"))).toBe(3);
    expect(compiled).toEqual(["sheet.A1.(int64) + 1", "sheet.A1.(int64) + 10"]);
  });

  test("compiler cache includes package cache keys in formula recompilation decisions", async () => {
    const engine = new SpreadsheetEngine();
    const cache = new SpreadsheetFormulaCompilerCache();
    const deps = [cellDependency(ref("A1"))];
    const compiledKeys: string[] = [];
    const compiler = async (input: SpreadsheetFormulaCompileInput) => {
      compiledKeys.push(input.cacheKey);
      return (ctx: SpreadsheetFormulaContext) => Number(ctx.cell("A1")) + compiledKeys.length;
    };

    engine.setLiteral(ref("A1"), 10);
    const first = await cache.install(engine, ref("B1"), {
      filename: "sheet!B1.gojr",
      source: "return lib.Add(sheet.A1.(int64))",
      declaredDeps: deps,
      packageCacheKeys: ["example.com/lib:v1"],
      compiler
    });
    await engine.recalculate();
    expect(engine.value(ref("B1"))).toBe(11);

    const samePackage = await cache.install(engine, ref("B1"), {
      filename: "sheet!B1.gojr",
      source: "return lib.Add(sheet.A1.(int64))",
      declaredDeps: deps,
      packageCacheKeys: ["example.com/lib:v1"],
      compiler: async () => {
        throw new Error("same package cache key should not compile");
      }
    });
    expect(samePackage).toMatchObject({ action: "skipped", compileAction: "skipped", cacheKey: first.cacheKey });

    const changedPackage = await cache.install(engine, ref("B1"), {
      filename: "sheet!B1.gojr",
      source: "return lib.Add(sheet.A1.(int64))",
      declaredDeps: deps,
      packageCacheKeys: ["example.com/lib:v2"],
      compiler
    });
    await engine.recalculate();
    expect(engine.value(ref("B1"))).toBe(12);
    expect(changedPackage.cacheKey).not.toBe(first.cacheKey);
    expect(compiledKeys).toEqual([first.cacheKey, changedPackage.cacheKey]);
  });
});
