const DEFAULT_SHEET = "sheet";
const DEFAULT_DEPENDENCY_STABILIZATION_LIMIT = 20;
const DEFAULT_MAX_ITERATIONS = 100;
const DEFAULT_NUMERIC_TOLERANCE = 0.001;
const DEFAULT_MAX_OBSERVED_STATES = 16;
export class SpreadsheetEngine {
    options;
    cells = new Map();
    reverseDependents = new Map();
    dirtyFormulas = new Set();
    constructor(options = {}) {
        this.options = options;
    }
    setLiteral(ref, value, options = {}) {
        const normalized = normalizeCellRef(ref);
        const id = cellId(normalized);
        const previous = this.cells.get(id);
        const replacedFormula = previous?.kind === "formula";
        const valueChanged = !previous || previous.kind !== "literal" || !spreadsheetValueEqual(previous.value, value);
        const typeChanged = typeTextOf(previous) !== options.typeText;
        if (replacedFormula)
            this.removeFormulaEdges(id, previous);
        this.cells.set(id, {
            kind: "literal",
            value,
            ...(options.typeText ? { typeText: options.typeText } : {})
        });
        this.dirtyFormulas.delete(id);
        if (valueChanged || typeChanged || replacedFormula)
            this.markDependentsDirty(cellDependency(normalized));
    }
    setFormula(ref, evaluator, options = {}) {
        const normalized = normalizeCellRef(ref);
        const id = cellId(normalized);
        const previous = this.cells.get(id);
        const declaredDeps = dependencyMap(options.declaredDeps ?? []);
        if (this.formulaInstallMatches(normalized, options)) {
            return {
                action: "skipped",
                cacheKey: options.cacheKey
            };
        }
        if (previous?.kind === "formula")
            this.removeFormulaEdges(id, previous);
        const record = {
            kind: "formula",
            evaluator,
            value: previous?.value,
            ...(options.typeText ? { typeText: options.typeText } : {}),
            ...(options.cacheKey ? { cacheKey: options.cacheKey } : {}),
            declaredDeps,
            observedDeps: new Map(),
            activeDeps: new Map(declaredDeps),
            diagnostics: [],
            hasEvaluated: false
        };
        this.cells.set(id, record);
        this.installEdges(id, record.activeDeps);
        this.dirtyFormulas.add(id);
        this.markDependentsDirty(cellDependency(normalized));
        return {
            action: "installed",
            ...(options.cacheKey ? { cacheKey: options.cacheKey } : {})
        };
    }
    formulaInstallMatches(ref, options = {}) {
        const record = this.cells.get(cellId(normalizeCellRef(ref)));
        if (record?.kind !== "formula" || options.cacheKey === undefined)
            return false;
        return record.cacheKey === options.cacheKey &&
            typeTextOf(record) === options.typeText &&
            dependencyMapsEqual(record.declaredDeps, dependencyMap(options.declaredDeps ?? []));
    }
    removeCell(ref) {
        const normalized = normalizeCellRef(ref);
        const id = cellId(normalized);
        const previous = this.cells.get(id);
        if (!previous)
            return;
        if (previous.kind === "formula")
            this.removeFormulaEdges(id, previous);
        this.cells.delete(id);
        this.dirtyFormulas.delete(id);
        this.markDependentsDirty(cellDependency(normalized));
    }
    value(ref) {
        return this.cells.get(cellId(normalizeCellRef(ref)))?.value;
    }
    diagnostics(ref) {
        const record = this.cells.get(cellId(normalizeCellRef(ref)));
        return record?.kind === "formula" ? record.diagnostics : [];
    }
    declaredDeps(ref) {
        const record = this.cells.get(cellId(normalizeCellRef(ref)));
        return record?.kind === "formula" ? [...record.declaredDeps.values()] : [];
    }
    observedDeps(ref) {
        const record = this.cells.get(cellId(normalizeCellRef(ref)));
        return record?.kind === "formula" ? [...record.observedDeps.values()] : [];
    }
    formulaCacheKey(ref) {
        const record = this.cells.get(cellId(normalizeCellRef(ref)));
        return record?.kind === "formula" ? record.cacheKey : undefined;
    }
    reverseDependentsOf(dependency) {
        return [...this.directDependents(normalizeDependency(dependency))].map(cellRefFromId).sort(compareCellRefs);
    }
    async recalculate() {
        const evaluated = [];
        const diagnostics = [];
        const dependencyLimit = this.options.dependencyStabilizationLimit ?? DEFAULT_DEPENDENCY_STABILIZATION_LIMIT;
        let pass = 0;
        let unstable = false;
        while (this.dirtyFormulas.size > 0) {
            if (pass++ >= dependencyLimit) {
                unstable = true;
                for (const id of [...this.dirtyFormulas].sort()) {
                    const diagnostic = this.setFormulaError(id, "#UNSTABLE_DEPS!", "dynamic dependency set did not stabilize", [cellRefFromId(id)]);
                    diagnostics.push(diagnostic);
                }
                this.dirtyFormulas.clear();
                break;
            }
            const dirtyNow = new Set([...this.dirtyFormulas].filter((id) => this.cells.get(id)?.kind === "formula"));
            this.dirtyFormulas.clear();
            const components = stronglyConnectedComponents(dirtyNow, (id) => this.formulaDependencies(id, dirtyNow));
            const cyclic = components.filter((component) => this.isCyclicComponent(component));
            const cyclicIds = new Set(cyclic.flat());
            if (cyclic.length > 0) {
                if (!this.iterativeOptions().enabled) {
                    for (const component of cyclic) {
                        const related = component.map(cellRefFromId).sort(compareCellRefs);
                        for (const id of component) {
                            const diagnostic = this.setFormulaError(id, "#CYCLE!", "circular reference", related);
                            diagnostics.push(diagnostic);
                            evaluated.push(cellRefFromId(id));
                            this.markDependentsDirty(cellDependency(cellRefFromId(id)), new Set(component));
                        }
                    }
                }
                else {
                    for (const component of cyclic) {
                        const outcome = await this.evaluateIterativeComponent(component, evaluated);
                        diagnostics.push(...outcome);
                    }
                }
            }
            const acyclicDirty = new Set([...dirtyNow].filter((id) => !cyclicIds.has(id)));
            const order = topologicalOrder(acyclicDirty, (id) => this.formulaDependencies(id, acyclicDirty));
            for (const id of order) {
                const record = this.cells.get(id);
                if (record?.kind !== "formula")
                    continue;
                const outcome = await this.evaluateFormula(id, record);
                evaluated.push(cellRefFromId(id));
                diagnostics.push(...record.diagnostics);
                if (outcome.valueChanged)
                    this.markDependentsDirty(cellDependency(cellRefFromId(id)), dirtyNow);
                if (outcome.depsChanged)
                    this.dirtyFormulas.add(id);
            }
        }
        return {
            evaluated,
            diagnostics,
            unstable
        };
    }
    async evaluateFormula(id, record) {
        const previousValue = record.value;
        const observed = new Map();
        const context = new SpreadsheetFormulaEvaluationContext(this, cellRefFromId(id), observed);
        try {
            const raw = await record.evaluator(context);
            if (isSpreadsheetFormulaEvaluation(raw)) {
                record.value = raw.value;
                if (raw.observedDeps) {
                    observed.clear();
                    for (const [key, dependency] of dependencyMap(raw.observedDeps))
                        observed.set(key, dependency);
                }
            }
            else {
                record.value = raw;
            }
            record.diagnostics = [];
        }
        catch (error) {
            const message = error instanceof Error ? error.message : String(error);
            const diagnostic = makeDiagnostic("#ERROR!", message, cellRefFromId(id), []);
            record.value = spreadsheetError("#ERROR!", message, [cellRefFromId(id)]);
            record.diagnostics = [diagnostic];
        }
        record.hasEvaluated = true;
        const depsChanged = !dependencyMapsEqual(record.activeDeps, observed);
        if (depsChanged) {
            this.removeEdges(id, record.activeDeps);
            record.activeDeps = new Map(observed);
            this.installEdges(id, record.activeDeps);
        }
        record.observedDeps = observed;
        return {
            valueChanged: !spreadsheetValueEqual(previousValue, record.value),
            depsChanged,
            previousValue
        };
    }
    async evaluateIterativeComponent(component, evaluated) {
        const ids = [...component].sort();
        const diagnostics = [];
        const options = this.iterativeOptions();
        const maxIterations = options.maxIterations ?? DEFAULT_MAX_ITERATIONS;
        const tolerance = options.numericTolerance ?? DEFAULT_NUMERIC_TOLERANCE;
        const detectOscillation = options.detectOscillation ?? true;
        const maxObservedStates = options.maxObservedStates ?? DEFAULT_MAX_OBSERVED_STATES;
        const initialValues = new Map(ids.map((id) => [id, this.cells.get(id)?.value]));
        const seen = new Set();
        let lastDepsChanged = false;
        for (let iteration = 0; iteration < maxIterations; iteration++) {
            const before = new Map(ids.map((id) => [id, this.cells.get(id)?.value]));
            lastDepsChanged = false;
            for (const id of ids) {
                const record = this.cells.get(id);
                if (record?.kind !== "formula")
                    continue;
                const outcome = await this.evaluateFormula(id, record);
                lastDepsChanged ||= outcome.depsChanged;
                evaluated.push(cellRefFromId(id));
                diagnostics.push(...record.diagnostics);
            }
            const after = new Map(ids.map((id) => [id, this.cells.get(id)?.value]));
            if (!lastDepsChanged && valueVectorsEqual(before, after, tolerance)) {
                this.clearDirty(ids);
                this.markChangedComponentDependents(ids, initialValues);
                return diagnostics;
            }
            const signature = valueVectorSignature(after);
            if (detectOscillation && seen.has(signature)) {
                diagnostics.push(...this.setComponentError(ids, "#OSCILLATE!", "iterative calculation repeated a previous value vector"));
                this.clearDirty(ids);
                this.markChangedComponentDependents(ids, initialValues);
                return diagnostics;
            }
            seen.add(signature);
            if (seen.size > maxObservedStates)
                seen.delete(seen.values().next().value);
        }
        diagnostics.push(...this.setComponentError(ids, lastDepsChanged ? "#UNSTABLE_DEPS!" : "#DIVERGE!", lastDepsChanged ? "dynamic dependency set did not stabilize" : "iterative calculation did not converge"));
        this.clearDirty(ids);
        this.markChangedComponentDependents(ids, initialValues);
        return diagnostics;
    }
    setComponentError(ids, code, message) {
        const related = ids.map(cellRefFromId).sort(compareCellRefs);
        return ids.map((id) => this.setFormulaError(id, code, message, related));
    }
    setFormulaError(id, code, message, relatedCells) {
        const record = this.cells.get(id);
        const cell = cellRefFromId(id);
        const diagnostic = makeDiagnostic(code, message, cell, relatedCells);
        if (record?.kind === "formula") {
            record.value = spreadsheetError(code, message, relatedCells.length > 0 ? relatedCells : [cell]);
            record.diagnostics = [diagnostic];
            record.hasEvaluated = true;
        }
        return diagnostic;
    }
    clearDirty(ids) {
        for (const id of ids)
            this.dirtyFormulas.delete(id);
    }
    markChangedComponentDependents(ids, initialValues) {
        for (const id of ids) {
            const current = this.cells.get(id)?.value;
            if (!spreadsheetValueEqual(initialValues.get(id), current)) {
                this.markDependentsDirty(cellDependency(cellRefFromId(id)), new Set(ids));
            }
        }
    }
    removeFormulaEdges(id, record) {
        this.removeEdges(id, record.activeDeps);
    }
    installEdges(id, dependencies) {
        for (const key of dependencies.keys()) {
            const dependents = this.reverseDependents.get(key) ?? new Set();
            dependents.add(id);
            this.reverseDependents.set(key, dependents);
        }
    }
    removeEdges(id, dependencies) {
        for (const key of dependencies.keys()) {
            const dependents = this.reverseDependents.get(key);
            if (!dependents)
                continue;
            dependents.delete(id);
            if (dependents.size === 0)
                this.reverseDependents.delete(key);
        }
    }
    markDependentsDirty(dependency, skip = new Set()) {
        const queue = [...this.directDependents(normalizeDependency(dependency))];
        const seen = new Set();
        while (queue.length > 0) {
            const id = queue.shift();
            if (!id || seen.has(id))
                continue;
            seen.add(id);
            if (skip.has(id))
                continue;
            const record = this.cells.get(id);
            if (record?.kind !== "formula")
                continue;
            this.dirtyFormulas.add(id);
            queue.push(...this.directDependents(cellDependency(cellRefFromId(id))));
        }
    }
    directDependents(dependency) {
        const normalized = normalizeDependency(dependency);
        const dependents = new Set(this.reverseDependents.get(dependencyKey(normalized)) ?? []);
        if (normalized.kind === "cell") {
            for (const [key, formulas] of this.reverseDependents.entries()) {
                const parsed = dependencyFromKey(key);
                if (parsed.kind === "range" && parsed.sheet === normalized.sheet && rangeContains(parsed, normalized.cell)) {
                    for (const formula of formulas)
                        dependents.add(formula);
                }
            }
        }
        return dependents;
    }
    formulaDependencies(id, within) {
        const record = this.cells.get(id);
        const dependencies = new Set();
        if (record?.kind !== "formula")
            return dependencies;
        for (const dependency of record.activeDeps.values()) {
            if (dependency.kind === "cell") {
                const dependencyId = cellId(dependency);
                if (dependencyId !== id && this.cells.get(dependencyId)?.kind === "formula" && (!within || within.has(dependencyId))) {
                    dependencies.add(dependencyId);
                }
            }
            else {
                for (const dependencyId of this.formulaIdsInRange(dependency)) {
                    if (dependencyId !== id && (!within || within.has(dependencyId)))
                        dependencies.add(dependencyId);
                }
            }
        }
        return dependencies;
    }
    formulaIdsInRange(range) {
        const ids = [];
        const bounds = normalizedRangeBounds(range.start, range.end);
        for (const [id, record] of this.cells.entries()) {
            if (record.kind !== "formula")
                continue;
            const ref = cellRefFromId(id);
            if (ref.sheet !== range.sheet)
                continue;
            const address = parseCellAddress(ref.cell);
            if (address.row >= bounds.top && address.row <= bounds.bottom && address.col >= bounds.left && address.col <= bounds.right) {
                ids.push(id);
            }
        }
        return ids;
    }
    isCyclicComponent(component) {
        if (component.length > 1)
            return true;
        const id = component[0];
        if (!id)
            return false;
        const record = this.cells.get(id);
        if (record?.kind !== "formula")
            return false;
        return [...record.activeDeps.values()].some((dependency) => {
            if (dependency.kind === "cell")
                return cellId(dependency) === id;
            const ref = cellRefFromId(id);
            return dependency.sheet === ref.sheet && rangeContains(dependency, ref.cell);
        });
    }
    readCellValue(ref) {
        return this.cells.get(cellId(normalizeCellRef(ref)))?.value;
    }
    readRangeValues(range) {
        const normalized = normalizeRangeDependency(range);
        const bounds = normalizedRangeBounds(normalized.start, normalized.end);
        const rows = [];
        for (let row = bounds.top; row <= bounds.bottom; row++) {
            const values = [];
            for (let col = bounds.left; col <= bounds.right; col++) {
                values.push(this.readCellValue({ sheet: normalized.sheet, cell: formatCellAddress(col, row) }));
            }
            rows.push(values);
        }
        return rows;
    }
    iterativeOptions() {
        const configured = this.options.iterativeCalculation ?? {};
        return {
            enabled: configured.enabled ?? false,
            maxIterations: configured.maxIterations ?? DEFAULT_MAX_ITERATIONS,
            numericTolerance: configured.numericTolerance ?? DEFAULT_NUMERIC_TOLERANCE,
            detectOscillation: configured.detectOscillation ?? true,
            maxObservedStates: configured.maxObservedStates ?? DEFAULT_MAX_OBSERVED_STATES
        };
    }
}
export class SpreadsheetFormulaCompilerCache {
    evaluators = new Map();
    has(cacheKey) {
        return this.evaluators.has(cacheKey);
    }
    delete(cacheKey) {
        return this.evaluators.delete(cacheKey);
    }
    clear() {
        this.evaluators.clear();
    }
    async install(engine, ref, options) {
        const declaredDeps = [...dependencyMap(options.declaredDeps ?? []).values()];
        const cacheKey = spreadsheetFormulaCacheKey({
            source: options.source,
            ...(options.filename ? { filename: options.filename } : {}),
            ...(options.typeText ? { typeText: options.typeText } : {}),
            declaredDeps,
            ...(options.packageCacheKeys ? { packageCacheKeys: options.packageCacheKeys } : {}),
            ...(options.compilerVersion ? { compilerVersion: options.compilerVersion } : {}),
            ...(options.hostSpecVersion ? { hostSpecVersion: options.hostSpecVersion } : {}),
            ...(options.capabilityPolicy ? { capabilityPolicy: options.capabilityPolicy } : {})
        });
        const installOptions = {
            declaredDeps,
            ...(options.typeText ? { typeText: options.typeText } : {}),
            cacheKey
        };
        if (engine.formulaInstallMatches(ref, installOptions)) {
            return {
                action: "skipped",
                cacheKey,
                compileAction: "skipped"
            };
        }
        let compileAction = "reused";
        let evaluator = this.evaluators.get(cacheKey);
        if (!evaluator) {
            evaluator = await options.compiler({
                ref: normalizeCellRef(ref),
                cacheKey,
                source: options.source,
                ...(options.filename ? { filename: options.filename } : {}),
                ...(options.typeText ? { typeText: options.typeText } : {}),
                declaredDeps,
                ...(options.packageCacheKeys ? { packageCacheKeys: options.packageCacheKeys } : {}),
                ...(options.compilerVersion ? { compilerVersion: options.compilerVersion } : {}),
                ...(options.hostSpecVersion ? { hostSpecVersion: options.hostSpecVersion } : {}),
                ...(options.capabilityPolicy ? { capabilityPolicy: options.capabilityPolicy } : {})
            });
            this.evaluators.set(cacheKey, evaluator);
            compileAction = "compiled";
        }
        const result = engine.setFormula(ref, evaluator, installOptions);
        return {
            action: result.action,
            cacheKey,
            compileAction: result.action === "skipped" ? "skipped" : compileAction
        };
    }
}
class SpreadsheetFormulaEvaluationContext {
    engine;
    currentCell;
    observed;
    constructor(engine, currentCell, observed) {
        this.engine = engine;
        this.currentCell = currentCell;
        this.observed = observed;
    }
    cell(cellOrSheet, maybeCell) {
        const ref = maybeCell === undefined
            ? normalizeCellRef({ sheet: this.currentCell.sheet, cell: cellOrSheet })
            : normalizeCellRef({ sheet: cellOrSheet, cell: maybeCell });
        const dependency = cellDependency(ref);
        this.observed.set(dependencyKey(dependency), dependency);
        return this.engine.readCellValue(ref);
    }
    range(startOrSheet, endOrStart, maybeEnd) {
        const dependency = maybeEnd === undefined
            ? normalizeRangeDependency({ kind: "range", sheet: this.currentCell.sheet, start: startOrSheet, end: endOrStart })
            : normalizeRangeDependency({ kind: "range", sheet: startOrSheet, start: endOrStart, end: maybeEnd });
        this.observed.set(dependencyKey(dependency), dependency);
        return this.engine.readRangeValues(dependency);
    }
}
export function cellDependency(ref) {
    const normalized = normalizeCellRef(ref);
    return { kind: "cell", sheet: normalized.sheet, cell: normalized.cell };
}
export function rangeDependency(sheet, start, end) {
    return normalizeRangeDependency({ kind: "range", sheet, start, end });
}
export function spreadsheetError(code, message, cells) {
    return {
        kind: "SpreadsheetError",
        code,
        message,
        cells: cells.map(normalizeCellRef)
    };
}
export function spreadsheetFormulaEvaluation(value, observedDeps) {
    return {
        kind: "SpreadsheetFormulaEvaluation",
        value,
        ...(observedDeps ? { observedDeps } : {})
    };
}
export function spreadsheetFormulaCacheKey(input) {
    return stableSpreadsheetHash([
        "gojr-formula-v1",
        input.compilerVersion ?? "gojr-dev",
        input.hostSpecVersion ?? "host-v0",
        input.capabilityPolicy ?? "default",
        input.filename ?? "",
        input.typeText ?? "",
        input.source,
        [...(input.packageCacheKeys ?? [])].sort().join("\n"),
        [...dependencyMap(input.declaredDeps ?? []).keys()].sort().join("\n")
    ].join("\0"));
}
function makeDiagnostic(code, message, cell, relatedCells) {
    return {
        code,
        message,
        cell: normalizeCellRef(cell),
        relatedCells: relatedCells.map(normalizeCellRef)
    };
}
function typeTextOf(record) {
    return record?.typeText;
}
function dependencyMap(dependencies) {
    const map = new Map();
    for (const dependency of dependencies) {
        const normalized = normalizeDependency(dependency);
        map.set(dependencyKey(normalized), normalized);
    }
    return map;
}
function isSpreadsheetFormulaEvaluation(value) {
    return Boolean(value && typeof value === "object" && value.kind === "SpreadsheetFormulaEvaluation");
}
function normalizeDependency(dependency) {
    return dependency.kind === "cell" ? cellDependency(dependency) : normalizeRangeDependency(dependency);
}
function normalizeCellRef(ref) {
    return {
        sheet: ref.sheet || DEFAULT_SHEET,
        cell: normalizeCellAddress(ref.cell)
    };
}
function normalizeRangeDependency(dependency) {
    const start = parseCellAddress(dependency.start);
    const end = parseCellAddress(dependency.end);
    const left = Math.min(start.col, end.col);
    const right = Math.max(start.col, end.col);
    const top = Math.min(start.row, end.row);
    const bottom = Math.max(start.row, end.row);
    return {
        kind: "range",
        sheet: dependency.sheet || DEFAULT_SHEET,
        start: formatCellAddress(left, top),
        end: formatCellAddress(right, bottom)
    };
}
function normalizeCellAddress(cell) {
    return cell.replace(/\$/g, "").toUpperCase();
}
function cellId(ref) {
    const normalized = normalizeCellRef(ref);
    return `${normalized.sheet}!${normalized.cell}`;
}
function cellRefFromId(id) {
    const separator = id.indexOf("!");
    return {
        sheet: separator >= 0 ? id.slice(0, separator) : DEFAULT_SHEET,
        cell: separator >= 0 ? id.slice(separator + 1) : id
    };
}
function dependencyKey(dependency) {
    const normalized = normalizeDependency(dependency);
    if (normalized.kind === "cell")
        return `cell:${normalized.sheet}:${normalized.cell}`;
    return `range:${normalized.sheet}:${normalized.start}:${normalized.end}`;
}
function dependencyFromKey(key) {
    const [kind, sheet, start, end] = key.split(":");
    if (kind === "cell")
        return cellDependency({ sheet: sheet ?? DEFAULT_SHEET, cell: start ?? "A1" });
    return rangeDependency(sheet ?? DEFAULT_SHEET, start ?? "A1", end ?? start ?? "A1");
}
function rangeContains(range, cell) {
    const bounds = normalizedRangeBounds(range.start, range.end);
    const address = parseCellAddress(cell);
    return address.row >= bounds.top && address.row <= bounds.bottom && address.col >= bounds.left && address.col <= bounds.right;
}
function normalizedRangeBounds(startCell, endCell) {
    const start = parseCellAddress(startCell);
    const end = parseCellAddress(endCell);
    return {
        left: Math.min(start.col, end.col),
        right: Math.max(start.col, end.col),
        top: Math.min(start.row, end.row),
        bottom: Math.max(start.row, end.row)
    };
}
function parseCellAddress(cell) {
    const match = /^([A-Z]+)([1-9][0-9]*)$/.exec(normalizeCellAddress(cell));
    if (!match)
        throw new Error(`${cell} is not a spreadsheet cell address`);
    return {
        col: columnNameToNumber(match[1] ?? "A"),
        row: Number(match[2] ?? "1")
    };
}
function columnNameToNumber(name) {
    let value = 0;
    for (const char of name)
        value = value * 26 + (char.charCodeAt(0) - 64);
    return value;
}
function formatCellAddress(col, row) {
    let name = "";
    let value = col;
    while (value > 0) {
        const remainder = (value - 1) % 26;
        name = String.fromCharCode(65 + remainder) + name;
        value = Math.floor((value - 1) / 26);
    }
    return `${name}${row}`;
}
function compareCellRefs(left, right) {
    return left.sheet.localeCompare(right.sheet) || left.cell.localeCompare(right.cell);
}
function topologicalOrder(nodes, dependenciesOf) {
    const order = [];
    const temporary = new Set();
    const permanent = new Set();
    const visit = (id) => {
        if (permanent.has(id) || temporary.has(id))
            return;
        temporary.add(id);
        for (const dependency of [...dependenciesOf(id)].sort())
            visit(dependency);
        temporary.delete(id);
        permanent.add(id);
        order.push(id);
    };
    for (const id of [...nodes].sort())
        visit(id);
    return order;
}
function stronglyConnectedComponents(nodes, dependenciesOf) {
    let index = 0;
    const stack = [];
    const indexes = new Map();
    const lowlinks = new Map();
    const onStack = new Set();
    const components = [];
    const strongConnect = (id) => {
        indexes.set(id, index);
        lowlinks.set(id, index);
        index++;
        stack.push(id);
        onStack.add(id);
        for (const dependency of dependenciesOf(id)) {
            if (!nodes.has(dependency))
                continue;
            if (!indexes.has(dependency)) {
                strongConnect(dependency);
                lowlinks.set(id, Math.min(lowlinks.get(id) ?? 0, lowlinks.get(dependency) ?? 0));
            }
            else if (onStack.has(dependency)) {
                lowlinks.set(id, Math.min(lowlinks.get(id) ?? 0, indexes.get(dependency) ?? 0));
            }
        }
        if (lowlinks.get(id) === indexes.get(id)) {
            const component = [];
            let current;
            do {
                current = stack.pop();
                if (!current)
                    break;
                onStack.delete(current);
                component.push(current);
            } while (current !== id);
            components.push(component.sort());
        }
    };
    for (const id of [...nodes].sort()) {
        if (!indexes.has(id))
            strongConnect(id);
    }
    return components;
}
function dependencyMapsEqual(left, right) {
    if (left.size !== right.size)
        return false;
    for (const key of left.keys()) {
        if (!right.has(key))
            return false;
    }
    return true;
}
function valueVectorsEqual(left, right, tolerance) {
    if (left.size !== right.size)
        return false;
    for (const [key, value] of left.entries()) {
        if (!spreadsheetValueEqual(value, right.get(key), tolerance))
            return false;
    }
    return true;
}
function valueVectorSignature(values) {
    return [...values.entries()]
        .sort(([left], [right]) => left.localeCompare(right))
        .map(([key, value]) => `${key}=${stableStringify(value)}`)
        .join("|");
}
function spreadsheetValueEqual(left, right, numericTolerance = 0) {
    if (typeof left === "number" && typeof right === "number") {
        if (Number.isNaN(left) && Number.isNaN(right))
            return true;
        return Math.abs(left - right) <= numericTolerance;
    }
    if (typeof left === "bigint" || typeof right === "bigint")
        return left === right;
    if (Object.is(left, right))
        return true;
    if (Array.isArray(left) || Array.isArray(right)) {
        return Array.isArray(left) &&
            Array.isArray(right) &&
            left.length === right.length &&
            left.every((value, index) => spreadsheetValueEqual(value, right[index], numericTolerance));
    }
    if (left && right && typeof left === "object" && typeof right === "object") {
        const leftObject = left;
        const rightObject = right;
        const leftKeys = Object.keys(leftObject).sort();
        const rightKeys = Object.keys(rightObject).sort();
        return leftKeys.length === rightKeys.length &&
            leftKeys.every((key, index) => key === rightKeys[index] && spreadsheetValueEqual(leftObject[key], rightObject[key], numericTolerance));
    }
    return false;
}
function stableStringify(value) {
    if (typeof value === "bigint")
        return `${value}n`;
    if (typeof value === "number") {
        if (Number.isNaN(value))
            return "NaN";
        if (value === Infinity)
            return "+Inf";
        if (value === -Infinity)
            return "-Inf";
        return String(value);
    }
    if (typeof value === "string")
        return JSON.stringify(value);
    if (value === null || value === undefined || typeof value === "boolean")
        return String(value);
    if (Array.isArray(value))
        return `[${value.map(stableStringify).join(",")}]`;
    if (typeof value === "object") {
        const object = value;
        return `{${Object.keys(object).sort().map((key) => `${JSON.stringify(key)}:${stableStringify(object[key])}`).join(",")}}`;
    }
    return String(value);
}
function stableSpreadsheetHash(text) {
    let hash = 0xcbf29ce484222325n;
    const prime = 0x100000001b3n;
    const mask = 0xffffffffffffffffn;
    for (let index = 0; index < text.length; index++) {
        hash ^= BigInt(text.charCodeAt(index));
        hash = (hash * prime) & mask;
    }
    return hash.toString(16).padStart(16, "0");
}
