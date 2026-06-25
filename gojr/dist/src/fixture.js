import { formatDiagnostic } from "./diagnostics.js";
import { parseRuntimeJson } from "./jsonInput.js";
import { GoJuniorSession } from "./runtime.js";
import { SpreadsheetEngine, spreadsheetFormulaCacheKey, spreadsheetFormulaEvaluation } from "./spreadsheet.js";
const DEFAULT_SHEET = "sheet";
export function parseSpreadsheetFixtureJson(json) {
    const value = parseRuntimeJson(json);
    if (!isRuntimeObject(value)) {
        throw new Error("spreadsheet fixture JSON must be an object");
    }
    return fixtureFromRuntimeObject(value);
}
export async function runSpreadsheetFixtureJson(json, options = {}) {
    return runSpreadsheetFixture(parseSpreadsheetFixtureJson(json), options);
}
export async function runSpreadsheetFixture(fixture, options = {}) {
    const currentSheetName = fixture.currentSheetName ?? DEFAULT_SHEET;
    const sheets = normalizeFixtureSheets(fixture, currentSheetName);
    const engine = new SpreadsheetEngine(spreadsheetOptions(fixture));
    const knownRefs = new Map();
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
            }
            else {
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
export function collectSpreadsheetFixtureFormulaSourceFiles(fixture) {
    const currentSheetName = fixture.currentSheetName ?? DEFAULT_SHEET;
    const sheets = normalizeFixtureSheets(fixture, currentSheetName);
    const files = [];
    for (const [sheetName, cells] of Object.entries(sheets)) {
        for (const [cell, input] of Object.entries(cells)) {
            const spec = cellInputSpec(input);
            if (spec.formula === undefined)
                continue;
            const ref = normalizeRef({ sheet: sheetName, cell });
            files.push({
                filename: `${ref.sheet}!${ref.cell}.gojr`,
                source: spec.formula
            });
        }
    }
    return files.sort((left, right) => left.filename.localeCompare(right.filename));
}
async function evaluateFormulaCell(source, ref, engine, knownRefs, currentSheetName, fixture, options) {
    const snapshots = snapshotKnownSheets(engine, knownRefs);
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
function fixtureFromRuntimeObject(value) {
    const fixture = {};
    if (typeof value.currentSheetName === "string")
        fixture.currentSheetName = value.currentSheetName;
    if (isRuntimeObject(value.sheet))
        fixture.sheet = fixtureCellsFromObject(value.sheet);
    if (isRuntimeObject(value.cells))
        fixture.cells = fixtureCellsFromObject(value.cells);
    if (isRuntimeObject(value.sheets))
        fixture.sheets = fixtureSheetsFromObject(value.sheets);
    if (isRuntimeObject(value.iterativeCalculation))
        fixture.iterativeCalculation = iterativeOptionsFromObject(value.iterativeCalculation);
    const dependencyLimit = numberOption(value.dependencyStabilizationLimit, "dependencyStabilizationLimit");
    if (dependencyLimit !== undefined)
        fixture.dependencyStabilizationLimit = dependencyLimit;
    if (isRandomSeed(value.randomSeed))
        fixture.randomSeed = value.randomSeed;
    return fixture;
}
function fixtureCellsFromObject(value) {
    const cells = {};
    for (const [cell, input] of Object.entries(value)) {
        cells[cell] = fixtureCellInputFromRuntimeValue(input);
    }
    return cells;
}
function fixtureSheetsFromObject(value) {
    const sheets = {};
    for (const [sheetName, cells] of Object.entries(value)) {
        if (!isRuntimeObject(cells)) {
            throw new Error(`spreadsheet fixture sheet ${sheetName} must be an object`);
        }
        sheets[sheetName] = fixtureCellsFromObject(cells);
    }
    return sheets;
}
function fixtureCellInputFromRuntimeValue(value) {
    if (!isRuntimeObject(value) || !looksLikeCellSpec(value))
        return value;
    const spec = {};
    if (Object.prototype.hasOwnProperty.call(value, "value"))
        spec.value = value.value ?? null;
    if (typeof value.formula === "string")
        spec.formula = value.formula;
    if (typeof value.typeText === "string")
        spec.typeText = value.typeText;
    return spec;
}
function iterativeOptionsFromObject(value) {
    const options = {};
    if (typeof value.enabled === "boolean")
        options.enabled = value.enabled;
    if (typeof value.detectOscillation === "boolean")
        options.detectOscillation = value.detectOscillation;
    const maxIterations = numberOption(value.maxIterations, "iterativeCalculation.maxIterations");
    const numericTolerance = numberOption(value.numericTolerance, "iterativeCalculation.numericTolerance");
    const maxObservedStates = numberOption(value.maxObservedStates, "iterativeCalculation.maxObservedStates");
    if (maxIterations !== undefined)
        options.maxIterations = maxIterations;
    if (numericTolerance !== undefined)
        options.numericTolerance = numericTolerance;
    if (maxObservedStates !== undefined)
        options.maxObservedStates = maxObservedStates;
    return options;
}
function spreadsheetOptions(fixture) {
    return {
        ...(fixture.dependencyStabilizationLimit !== undefined ? { dependencyStabilizationLimit: fixture.dependencyStabilizationLimit } : {}),
        ...(fixture.iterativeCalculation ? { iterativeCalculation: fixture.iterativeCalculation } : {})
    };
}
function packageInfoKeys(packageInfos) {
    if (!packageInfos)
        return [];
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
function normalizeFixtureSheets(fixture, currentSheetName) {
    const sheets = {};
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
function cellInputSpec(input) {
    if (isRuntimeObject(input) && looksLikeCellSpec(input)) {
        return input;
    }
    return { value: input };
}
function snapshotKnownSheets(engine, knownRefs) {
    const sheets = {};
    for (const ref of [...knownRefs.values()].sort(compareRefs)) {
        const sheet = sheets[ref.sheet] ?? {};
        sheet[ref.cell] = engine.readCellValue(ref) ?? null;
        sheets[ref.sheet] = sheet;
    }
    return sheets;
}
function observedDepsByFormula(engine, knownRefs) {
    const observed = {};
    for (const ref of [...knownRefs.values()].sort(compareRefs)) {
        const deps = engine.observedDeps(ref);
        if (deps.length > 0)
            observed[refId(ref)] = deps;
    }
    return observed;
}
function runtimeValueFromResult(result) {
    if (result.values) {
        if (result.values.length === 0)
            return null;
        if (result.values.length === 1)
            return result.values[0] ?? null;
        return result.values;
    }
    return result.value ?? null;
}
function looksLikeCellSpec(value) {
    return Object.prototype.hasOwnProperty.call(value, "formula") ||
        Object.prototype.hasOwnProperty.call(value, "value") ||
        Object.prototype.hasOwnProperty.call(value, "typeText");
}
function numberOption(value, name) {
    if (value === undefined)
        return undefined;
    if (typeof value === "number")
        return value;
    if (typeof value === "bigint")
        return Number(value);
    throw new Error(`${name} must be numeric`);
}
function isRandomSeed(value) {
    return typeof value === "number" || typeof value === "string" || typeof value === "bigint";
}
function isRuntimeObject(value) {
    return Boolean(value && typeof value === "object" && !Array.isArray(value));
}
function normalizeRef(ref) {
    return {
        sheet: ref.sheet || DEFAULT_SHEET,
        cell: ref.cell.replace(/\$/g, "").toUpperCase()
    };
}
function refId(ref) {
    const normalized = normalizeRef(ref);
    return `${normalized.sheet}!${normalized.cell}`;
}
function compareRefs(left, right) {
    return left.sheet.localeCompare(right.sheet) || left.cell.localeCompare(right.cell);
}
