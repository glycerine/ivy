import { buildPackages, parseGoJuniorPackageArchive } from "./build.js";
import { formatDiagnostic, hasErrorDiagnostics, REPL_FILENAME } from "./diagnostics.js";
import { parseFrontSourceFiles } from "./front/parser.js";
import { frontFilesToProgramAst } from "./frontToAst.js";
import { checkGoJuniorFiles } from "./typecheck.js";
class BenchmarkMemoryArtifactStore {
    writes = new Map();
    mtimes = new Map();
    clock = 1000;
    read(path) {
        return this.writes.get(path);
    }
    mtimeMs(path) {
        return this.mtimes.get(path);
    }
    writeAtomic(path, source) {
        this.writes.set(path, source);
        this.clock += 1000;
        this.mtimes.set(path, this.clock);
    }
}
class PhaseAccumulator {
    samples = new Map();
    add(name, durationMs) {
        const list = this.samples.get(name);
        if (list) {
            list.push(durationMs);
            return;
        }
        this.samples.set(name, [durationMs]);
    }
    summaries() {
        return [...this.samples.entries()].map(([name, values]) => {
            const totalMs = values.reduce((sum, value) => sum + value, 0);
            return {
                name,
                iterations: values.length,
                totalMs,
                meanMs: totalMs / values.length,
                minMs: Math.min(...values),
                maxMs: Math.max(...values)
            };
        });
    }
}
export async function runGoJuniorBenchmark(inputCases, options = {}) {
    const selectedCases = benchmarkCases(inputCases, options.caseNames);
    const iterations = positiveInteger(options.iterations, 1);
    const warmupIterations = nonNegativeInteger(options.warmupIterations, 0);
    const cacheMode = options.cacheMode ?? "cold";
    const phaseSet = options.phaseSet ?? "front-end-and-build";
    const diagnostics = [];
    const cases = [];
    for (const benchmarkCase of selectedCases) {
        const report = await runOneBenchmarkCase(benchmarkCase, {
            ...options,
            iterations,
            warmupIterations,
            cacheMode,
            phaseSet
        });
        diagnostics.push(...report.diagnostics);
        cases.push(report);
    }
    return {
        ok: cases.every((item) => item.ok) && !hasErrorDiagnostics(diagnostics),
        iterations,
        warmupIterations,
        cacheMode,
        phaseSet,
        diagnostics,
        cases
    };
}
function benchmarkCases(inputCases, caseNames) {
    const cases = inputCases && inputCases.length > 0 ? inputCases : defaultGoJuniorBenchmarkCases();
    const names = new Set((caseNames ?? []).map((name) => name.trim()).filter(Boolean));
    return names.size === 0 ? cases : cases.filter((benchmarkCase) => names.has(benchmarkCase.name));
}
async function runOneBenchmarkCase(benchmarkCase, options) {
    const files = benchmarkCase.files.map(ensureSourceFile);
    const phaseSet = benchmarkCase.phaseSet ?? options.phaseSet;
    const accumulator = new PhaseAccumulator();
    const diagnostics = [];
    const metrics = {
        parsedFilesMax: 0,
        astJsonBytesMax: 0,
        artifactCountMax: 0,
        artifactBytesMax: 0,
        pkgdefBytesMax: 0,
        javascriptBytesMax: 0,
        runtimeAstJsonBytesMax: 0
    };
    const warmStore = options.cacheMode === "warm" ? new BenchmarkMemoryArtifactStore() : undefined;
    const totalIterations = options.warmupIterations + options.iterations;
    for (let index = 0; index < totalIterations; index += 1) {
        const collect = index >= options.warmupIterations;
        const iterationDiagnostics = [];
        const parsedResult = includesFrontEnd(phaseSet)
            ? await timed(accumulator, "parse", collect, () => parseFrontSourceFiles(files))
            : undefined;
        if (parsedResult) {
            iterationDiagnostics.push(...parsedResult.diagnostics);
            metrics.parsedFilesMax = Math.max(metrics.parsedFilesMax, parsedResult.files.length);
            if (!hasErrorDiagnostics(parsedResult.diagnostics)) {
                const ast = await timed(accumulator, "ast-lower", collect, () => frontFilesToProgramAst(parsedResult.files, [], []));
                metrics.astJsonBytesMax = Math.max(metrics.astJsonBytesMax, utf8ByteLength(stableStringify(ast)));
                const packageName = benchmarkCase.packageName ?? parsedResult.files.find((file) => file.name)?.name?.name ?? "main";
                const config = {
                    packageName,
                    packagePath: benchmarkCase.importPath ?? packageName,
                    autoImportFmt: false
                };
                const checked = await timed(accumulator, "typecheck", collect, () => checkGoJuniorFiles(parsedResult.files, parsedResult.statements, [], config));
                iterationDiagnostics.push(...checked.diagnostics);
            }
        }
        if (includesPackageBuild(phaseSet) && !hasErrorDiagnostics(iterationDiagnostics)) {
            const store = warmStore ?? new BenchmarkMemoryArtifactStore();
            const build = await timed(accumulator, "package-build", collect, () => buildPackages({
                importPath: benchmarkCase.importPath ?? benchmarkCase.packageName ?? "main",
                files,
                ...(benchmarkCase.packageSources ? { packageSources: benchmarkCase.packageSources } : {}),
                ...(benchmarkCase.sourcePackageProvider ? { sourcePackageProvider: benchmarkCase.sourcePackageProvider } : {}),
                artifactRoot: benchmarkCase.artifactRoot ?? options.artifactRoot ?? "gojr:bench",
                ...(benchmarkCase.packageCacheParent ?? options.packageCacheParent
                    ? { packageCacheParent: benchmarkCase.packageCacheParent ?? options.packageCacheParent }
                    : {}),
                ...(options.compilerVersion ? { compilerVersion: options.compilerVersion } : {})
            }, store));
            iterationDiagnostics.push(...build.diagnostics);
            collectArtifactMetrics(store, metrics);
        }
        if (collect)
            diagnostics.push(...iterationDiagnostics);
    }
    return {
        name: benchmarkCase.name,
        ok: !hasErrorDiagnostics(diagnostics),
        iterations: options.iterations,
        sourceBytes: files.reduce((sum, file) => sum + utf8ByteLength(file.source), 0),
        fileCount: files.length,
        ...(benchmarkCase.importPath ? { importPath: benchmarkCase.importPath } : {}),
        ...(benchmarkCase.packageName ? { packageName: benchmarkCase.packageName } : {}),
        diagnostics,
        phases: accumulator.summaries(),
        metrics
    };
}
function includesFrontEnd(phaseSet) {
    return phaseSet === "front-end" || phaseSet === "front-end-and-build";
}
function includesPackageBuild(phaseSet) {
    return phaseSet === "package-build" || phaseSet === "front-end-and-build";
}
async function timed(accumulator, name, collect, fn) {
    const start = nowMs();
    try {
        return await fn();
    }
    finally {
        const duration = nowMs() - start;
        if (collect)
            accumulator.add(name, duration);
    }
}
function collectArtifactMetrics(store, metrics) {
    metrics.artifactCountMax = Math.max(metrics.artifactCountMax, store.writes.size);
    let artifactBytes = 0;
    let pkgdefBytes = 0;
    let javascriptBytes = 0;
    let runtimeAstJsonBytes = 0;
    for (const source of store.writes.values()) {
        artifactBytes += utf8ByteLength(source);
        const archive = parseGoJuniorPackageArchive(source);
        const pkgdef = archive?.members.find((member) => member.name === "__.PKGDEF")?.data ?? "";
        pkgdefBytes += utf8ByteLength(pkgdef);
        javascriptBytes += utf8ByteLength(archive?.javascript ?? "");
        if (archive?.pkgdef.runtime?.ast) {
            runtimeAstJsonBytes += utf8ByteLength(stableStringify(archive.pkgdef.runtime.ast));
        }
    }
    metrics.artifactBytesMax = Math.max(metrics.artifactBytesMax, artifactBytes);
    metrics.pkgdefBytesMax = Math.max(metrics.pkgdefBytesMax, pkgdefBytes);
    metrics.javascriptBytesMax = Math.max(metrics.javascriptBytesMax, javascriptBytes);
    metrics.runtimeAstJsonBytesMax = Math.max(metrics.runtimeAstJsonBytesMax, runtimeAstJsonBytes);
}
export function defaultGoJuniorBenchmarkCases() {
    return [
        {
            name: "tiny-function",
            importPath: "bench/tiny",
            packageName: "tiny",
            files: [{
                    filename: "bench/tiny/tiny.go",
                    source: `package tiny

func Add(a, b int) int {
	return a + b
}
`
                }]
        },
        {
            name: "loop-and-branch",
            importPath: "bench/loop",
            packageName: "loop",
            files: [{
                    filename: "bench/loop/loop.go",
                    source: `package loop

func SumEven(n int) int {
	total := 0
	for i := 0; i < n; i++ {
		if i%2 == 0 {
			total += i
		}
	}
	return total
}
`
                }]
        },
        {
            name: "byte-literal-4k",
            importPath: "bench/bytes4k",
            packageName: "bytes4k",
            files: [{
                    filename: "bench/bytes4k/bytes.go",
                    source: byteLiteralSource(4096)
                }]
        }
    ];
}
function byteLiteralSource(length) {
    const values = [];
    for (let index = 0; index < length; index += 1)
        values.push(String(index % 256));
    const lines = [];
    for (let index = 0; index < values.length; index += 32) {
        lines.push(`\t${values.slice(index, index + 32).join(", ")},`);
    }
    return `package bytes4k

var Payload = []byte{
${lines.join("\n")}
}

func PayloadLen() int {
	return len(Payload)
}
`;
}
export function formatGoJuniorBenchmarkReport(report) {
    const lines = [];
    lines.push(`gojr bench: iterations=${report.iterations} warmup=${report.warmupIterations} cache=${report.cacheMode} phase=${report.phaseSet} ok=${report.ok}`);
    if (report.profilePath)
        lines.push(`gojr bench: cpu profile ${report.profilePath}`);
    for (const benchmarkCase of report.cases) {
        lines.push(`case ${benchmarkCase.name}: files=${benchmarkCase.fileCount} source_bytes=${benchmarkCase.sourceBytes} ok=${benchmarkCase.ok}`);
        if (benchmarkCase.metrics.artifactBytesMax > 0) {
            lines.push(`  artifact_bytes_max=${benchmarkCase.metrics.artifactBytesMax} pkgdef_bytes_max=${benchmarkCase.metrics.pkgdefBytesMax} js_bytes_max=${benchmarkCase.metrics.javascriptBytesMax} runtime_ast_json_bytes_max=${benchmarkCase.metrics.runtimeAstJsonBytesMax}`);
        }
        for (const phase of benchmarkCase.phases) {
            lines.push(`  ${phase.name}: total_ms=${formatMs(phase.totalMs)} mean_ms=${formatMs(phase.meanMs)} min_ms=${formatMs(phase.minMs)} max_ms=${formatMs(phase.maxMs)} n=${phase.iterations}`);
        }
        for (const diagnostic of benchmarkCase.diagnostics) {
            lines.push(`  ${formatDiagnostic(diagnostic)}`);
        }
    }
    for (const diagnostic of report.diagnostics.filter((diagnostic) => !report.cases.some((item) => item.diagnostics.includes(diagnostic)))) {
        lines.push(formatDiagnostic(diagnostic));
    }
    return `${lines.join("\n")}\n`;
}
function formatMs(value) {
    if (!Number.isFinite(value))
        return "NaN";
    return value.toFixed(value >= 10 ? 2 : 3);
}
function ensureSourceFile(file) {
    return {
        filename: file.filename || REPL_FILENAME,
        source: file.source ?? ""
    };
}
function positiveInteger(value, fallback) {
    const number = Math.trunc(Number(value));
    return Number.isFinite(number) && number > 0 ? number : fallback;
}
function nonNegativeInteger(value, fallback) {
    const number = Math.trunc(Number(value));
    return Number.isFinite(number) && number >= 0 ? number : fallback;
}
function nowMs() {
    return typeof performance !== "undefined" && typeof performance.now === "function"
        ? performance.now()
        : Date.now();
}
function utf8ByteLength(source) {
    if (typeof TextEncoder !== "undefined")
        return new TextEncoder().encode(source).byteLength;
    return source.length;
}
function stableStringify(value) {
    return JSON.stringify(value, (_key, item) => typeof item === "bigint" ? `${item}n` : item) ?? "";
}
