import { formatDiagnostic, type SourceFile } from "./diagnostics.js";
import type { Package as GoTypesPackage } from "./go/types/index.js";
import { parseRuntimeJson } from "./jsonInput.js";
import {
  GoJuniorSession,
  type EvaluationContext,
  RuntimeObject,
  RuntimeValue,
  SheetData
} from "./runtime.js";
import {
  IterativeCalculationOptions,
  SpreadsheetCellRef,
  SpreadsheetDependency,
  SpreadsheetDiagnostic,
  SpreadsheetEngine,
  SpreadsheetEngineOptions,
  SpreadsheetValue,
  spreadsheetFormulaCacheKey,
  spreadsheetFormulaEvaluation
} from "./spreadsheet.js";

export type SpreadsheetFixtureCellInput = RuntimeValue | SpreadsheetFixtureCellSpec;

export interface SpreadsheetFixtureCellSpec {
  readonly value?: RuntimeValue;
  readonly formula?: string;
  readonly typeText?: string;
  readonly declaredDeps?: readonly SpreadsheetDependency[];
}

export interface SpreadsheetFixture {
  readonly currentSheetName?: string;
  readonly sheet?: Record<string, SpreadsheetFixtureCellInput>;
  readonly cells?: Record<string, SpreadsheetFixtureCellInput>;
  readonly sheets?: Record<string, Record<string, SpreadsheetFixtureCellInput>>;
  readonly iterativeCalculation?: IterativeCalculationOptions;
  readonly dependencyStabilizationLimit?: number;
  readonly randomSeed?: number | string | bigint;
}

export interface SpreadsheetFixtureRunOptions {
  readonly randomSeed?: number | string | bigint;
  readonly packages?: Record<string, RuntimeObject>;
  readonly packageInfos?: Record<string, GoTypesPackage>;
  readonly packageContexts?: Record<string, EvaluationContext>;
}

export interface SpreadsheetFixtureRunResult {
  readonly ok: boolean;
  readonly unstable: boolean;
  readonly evaluated: readonly SpreadsheetCellRef[];
  readonly diagnostics: readonly SpreadsheetDiagnostic[];
  readonly sheets: Record<string, Record<string, SpreadsheetValue>>;
  readonly observedDeps: Record<string, readonly SpreadsheetDependency[]>;
}

const DEFAULT_SHEET = "sheet";

export function parseSpreadsheetFixtureJson(json: string): SpreadsheetFixture {
  const value = parseRuntimeJson(json);
  if (!isRuntimeObject(value)) {
    throw new Error("spreadsheet fixture JSON must be an object");
  }
  return fixtureFromRuntimeObject(value);
}

export async function runSpreadsheetFixtureJson(
  json: string,
  options: SpreadsheetFixtureRunOptions = {}
): Promise<SpreadsheetFixtureRunResult> {
  return runSpreadsheetFixture(parseSpreadsheetFixtureJson(json), options);
}

export async function runSpreadsheetFixture(
  fixture: SpreadsheetFixture,
  options: SpreadsheetFixtureRunOptions = {}
): Promise<SpreadsheetFixtureRunResult> {
  const currentSheetName = fixture.currentSheetName ?? DEFAULT_SHEET;
  const sheets = normalizeFixtureSheets(fixture, currentSheetName);
  const engine = new SpreadsheetEngine(spreadsheetOptions(fixture));
  const knownRefs = new Map<string, SpreadsheetCellRef>();

  for (const [sheetName, cells] of Object.entries(sheets)) {
    for (const [cell, input] of Object.entries(cells)) {
      const ref = normalizeRef({ sheet: sheetName, cell });
      knownRefs.set(refId(ref), ref);
      const spec = cellInputSpec(input);
      if (spec.formula !== undefined) {
        const formulaFilename = `${ref.sheet}!${ref.cell}.gojr`;
        engine.setFormula(ref, async () => {
          const result = await evaluateFormulaCell(spec.formula ?? "", ref, engine, knownRefs, currentSheetName, fixture, options);
          return spreadsheetFormulaEvaluation(result.value, result.observedDeps);
        }, {
          ...(spec.declaredDeps ? { declaredDeps: spec.declaredDeps } : {}),
          ...(spec.typeText ? { typeText: spec.typeText } : {}),
          cacheKey: spreadsheetFormulaCacheKey({
            source: spec.formula,
            filename: formulaFilename,
            ...(spec.typeText ? { typeText: spec.typeText } : {}),
            ...(spec.declaredDeps ? { declaredDeps: spec.declaredDeps } : {}),
            packageCacheKeys: packageInfoKeys(options.packageInfos)
          })
        });
      } else {
        engine.setLiteral(ref, spec.value ?? null, {
          ...(spec.typeText ? { typeText: spec.typeText } : {})
        });
      }
    }
  }

  const recalculation = await engine.recalculate();
  return {
    ok: recalculation.diagnostics.length === 0,
    unstable: recalculation.unstable,
    evaluated: recalculation.evaluated,
    diagnostics: recalculation.diagnostics,
    sheets: snapshotKnownSheets(engine, knownRefs),
    observedDeps: observedDepsByFormula(engine, knownRefs)
  };
}

export function collectSpreadsheetFixtureFormulaSourceFiles(fixture: SpreadsheetFixture): SourceFile[] {
  const currentSheetName = fixture.currentSheetName ?? DEFAULT_SHEET;
  const sheets = normalizeFixtureSheets(fixture, currentSheetName);
  const files: SourceFile[] = [];
  for (const [sheetName, cells] of Object.entries(sheets)) {
    for (const [cell, input] of Object.entries(cells)) {
      const spec = cellInputSpec(input);
      if (spec.formula === undefined) continue;
      const ref = normalizeRef({ sheet: sheetName, cell });
      files.push({
        filename: `${ref.sheet}!${ref.cell}.gojr`,
        source: spec.formula
      });
    }
  }
  return files.sort((left, right) => left.filename.localeCompare(right.filename));
}

async function evaluateFormulaCell(
  source: string,
  ref: SpreadsheetCellRef,
  engine: SpreadsheetEngine,
  knownRefs: ReadonlyMap<string, SpreadsheetCellRef>,
  currentSheetName: string,
  fixture: SpreadsheetFixture,
  options: SpreadsheetFixtureRunOptions
): Promise<{ value: RuntimeValue; observedDeps: readonly SpreadsheetDependency[] }> {
  const snapshots = snapshotKnownSheets(engine, knownRefs) as Record<string, SheetData>;
  const session = new GoJuniorSession({
    sheet: snapshots[currentSheetName] ?? {},
    sheets: snapshots,
    currentSheetName,
    filename: `${ref.sheet}!${ref.cell}.gojr`,
    ...(options.packages ? { packages: options.packages } : {}),
    ...(options.packageInfos ? { packageInfos: options.packageInfos } : {}),
    ...(options.packageContexts ? { packageContexts: options.packageContexts } : {}),
    ...(fixture.randomSeed !== undefined ? { randomSeed: fixture.randomSeed } : {}),
    ...(fixture.randomSeed === undefined && options.randomSeed !== undefined ? { randomSeed: options.randomSeed } : {})
  });
  const result = await session.evaluate(source);
  const errors = result.diagnostics.filter((diagnostic) => diagnostic.severity === "error");
  if (errors.length > 0) {
    throw new Error(errors.map(formatDiagnostic).join("\n"));
  }
  return {
    value: runtimeValueFromResult(result),
    observedDeps: result.observedDeps ?? []
  };
}

function fixtureFromRuntimeObject(value: RuntimeObject): SpreadsheetFixture {
  const fixture: {
    currentSheetName?: string;
    sheet?: Record<string, SpreadsheetFixtureCellInput>;
    cells?: Record<string, SpreadsheetFixtureCellInput>;
    sheets?: Record<string, Record<string, SpreadsheetFixtureCellInput>>;
    iterativeCalculation?: IterativeCalculationOptions;
    dependencyStabilizationLimit?: number;
    randomSeed?: number | string | bigint;
  } = {};

  if (typeof value.currentSheetName === "string") fixture.currentSheetName = value.currentSheetName;
  if (isRuntimeObject(value.sheet)) fixture.sheet = fixtureCellsFromObject(value.sheet);
  if (isRuntimeObject(value.cells)) fixture.cells = fixtureCellsFromObject(value.cells);
  if (isRuntimeObject(value.sheets)) fixture.sheets = fixtureSheetsFromObject(value.sheets);
  if (isRuntimeObject(value.iterativeCalculation)) fixture.iterativeCalculation = iterativeOptionsFromObject(value.iterativeCalculation);
  const dependencyLimit = numberOption(value.dependencyStabilizationLimit, "dependencyStabilizationLimit");
  if (dependencyLimit !== undefined) fixture.dependencyStabilizationLimit = dependencyLimit;
  if (isRandomSeed(value.randomSeed)) fixture.randomSeed = value.randomSeed;
  return fixture;
}

function fixtureCellsFromObject(value: RuntimeObject): Record<string, SpreadsheetFixtureCellInput> {
  const cells: Record<string, SpreadsheetFixtureCellInput> = {};
  for (const [cell, input] of Object.entries(value)) {
    cells[cell] = fixtureCellInputFromRuntimeValue(input);
  }
  return cells;
}

function fixtureSheetsFromObject(value: RuntimeObject): Record<string, Record<string, SpreadsheetFixtureCellInput>> {
  const sheets: Record<string, Record<string, SpreadsheetFixtureCellInput>> = {};
  for (const [sheetName, cells] of Object.entries(value)) {
    if (!isRuntimeObject(cells)) {
      throw new Error(`spreadsheet fixture sheet ${sheetName} must be an object`);
    }
    sheets[sheetName] = fixtureCellsFromObject(cells);
  }
  return sheets;
}

function fixtureCellInputFromRuntimeValue(value: RuntimeValue): SpreadsheetFixtureCellInput {
  if (!isRuntimeObject(value) || !looksLikeCellSpec(value)) return value;
  const spec: {
    value?: RuntimeValue;
    formula?: string;
    typeText?: string;
  } = {};
  if (Object.prototype.hasOwnProperty.call(value, "value")) spec.value = value.value ?? null;
  if (typeof value.formula === "string") spec.formula = value.formula;
  if (typeof value.typeText === "string") spec.typeText = value.typeText;
  return spec;
}

function iterativeOptionsFromObject(value: RuntimeObject): IterativeCalculationOptions {
  const options: {
    enabled?: boolean;
    maxIterations?: number;
    numericTolerance?: number;
    detectOscillation?: boolean;
    maxObservedStates?: number;
  } = {};
  if (typeof value.enabled === "boolean") options.enabled = value.enabled;
  if (typeof value.detectOscillation === "boolean") options.detectOscillation = value.detectOscillation;
  const maxIterations = numberOption(value.maxIterations, "iterativeCalculation.maxIterations");
  const numericTolerance = numberOption(value.numericTolerance, "iterativeCalculation.numericTolerance");
  const maxObservedStates = numberOption(value.maxObservedStates, "iterativeCalculation.maxObservedStates");
  if (maxIterations !== undefined) options.maxIterations = maxIterations;
  if (numericTolerance !== undefined) options.numericTolerance = numericTolerance;
  if (maxObservedStates !== undefined) options.maxObservedStates = maxObservedStates;
  return options;
}

function spreadsheetOptions(fixture: SpreadsheetFixture): SpreadsheetEngineOptions {
  return {
    ...(fixture.dependencyStabilizationLimit !== undefined ? { dependencyStabilizationLimit: fixture.dependencyStabilizationLimit } : {}),
    ...(fixture.iterativeCalculation ? { iterativeCalculation: fixture.iterativeCalculation } : {})
  };
}

function packageInfoKeys(packageInfos: Record<string, GoTypesPackage> | undefined): string[] {
  if (!packageInfos) return [];
  return Object.entries(packageInfos)
    .map(([path, info]) => {
      const exports = info.Scope().Names()
        .map((name) => info.Scope().Lookup(name))
        .filter((object) => object !== null)
        .map((object) => `${object.constructor.name}:${object.Name()}:${object.Type()?.String() ?? "<nil>"}`)
        .join(",");
      return `${path}:${info.Name()}:${exports}`;
    });
}

function normalizeFixtureSheets(
  fixture: SpreadsheetFixture,
  currentSheetName: string
): Record<string, Record<string, SpreadsheetFixtureCellInput>> {
  const sheets: Record<string, Record<string, SpreadsheetFixtureCellInput>> = {};
  for (const [sheetName, cells] of Object.entries(fixture.sheets ?? {})) {
    sheets[sheetName] = { ...cells };
  }
  const currentCells = {
    ...(sheets[currentSheetName] ?? {}),
    ...(fixture.sheet ?? {}),
    ...(fixture.cells ?? {})
  };
  if (Object.keys(currentCells).length > 0 || Object.keys(sheets).length === 0) {
    sheets[currentSheetName] = currentCells;
  }
  return sheets;
}

function cellInputSpec(input: SpreadsheetFixtureCellInput): SpreadsheetFixtureCellSpec {
  if (isRuntimeObject(input) && looksLikeCellSpec(input)) {
    return input as unknown as SpreadsheetFixtureCellSpec;
  }
  return { value: input as RuntimeValue };
}

function snapshotKnownSheets(
  engine: SpreadsheetEngine,
  knownRefs: ReadonlyMap<string, SpreadsheetCellRef>
): Record<string, Record<string, SpreadsheetValue>> {
  const sheets: Record<string, Record<string, SpreadsheetValue>> = {};
  for (const ref of [...knownRefs.values()].sort(compareRefs)) {
    const sheet = sheets[ref.sheet] ?? {};
    sheet[ref.cell] = engine.readCellValue(ref) ?? null;
    sheets[ref.sheet] = sheet;
  }
  return sheets;
}

function observedDepsByFormula(
  engine: SpreadsheetEngine,
  knownRefs: ReadonlyMap<string, SpreadsheetCellRef>
): Record<string, readonly SpreadsheetDependency[]> {
  const observed: Record<string, readonly SpreadsheetDependency[]> = {};
  for (const ref of [...knownRefs.values()].sort(compareRefs)) {
    const deps = engine.observedDeps(ref);
    if (deps.length > 0) observed[refId(ref)] = deps;
  }
  return observed;
}

function runtimeValueFromResult(result: { value?: RuntimeValue; values?: RuntimeValue[] }): RuntimeValue {
  if (result.values) {
    if (result.values.length === 0) return null;
    if (result.values.length === 1) return result.values[0] ?? null;
    return result.values;
  }
  return result.value ?? null;
}

function looksLikeCellSpec(value: RuntimeObject): boolean {
  return Object.prototype.hasOwnProperty.call(value, "formula") ||
    Object.prototype.hasOwnProperty.call(value, "value") ||
    Object.prototype.hasOwnProperty.call(value, "typeText");
}

function numberOption(value: RuntimeValue | undefined, name: string): number | undefined {
  if (value === undefined) return undefined;
  if (typeof value === "number") return value;
  if (typeof value === "bigint") return Number(value);
  throw new Error(`${name} must be numeric`);
}

function isRandomSeed(value: RuntimeValue | undefined): value is number | string | bigint {
  return typeof value === "number" || typeof value === "string" || typeof value === "bigint";
}

function isRuntimeObject(value: unknown): value is RuntimeObject {
  return Boolean(value && typeof value === "object" && !Array.isArray(value));
}

function normalizeRef(ref: SpreadsheetCellRef): SpreadsheetCellRef {
  return {
    sheet: ref.sheet || DEFAULT_SHEET,
    cell: ref.cell.replace(/\$/g, "").toUpperCase()
  };
}

function refId(ref: SpreadsheetCellRef): string {
  const normalized = normalizeRef(ref);
  return `${normalized.sheet}!${normalized.cell}`;
}

function compareRefs(left: SpreadsheetCellRef, right: SpreadsheetCellRef): number {
  return left.sheet.localeCompare(right.sheet) || left.cell.localeCompare(right.cell);
}
