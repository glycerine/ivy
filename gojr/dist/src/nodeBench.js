import { existsSync, mkdirSync, readFileSync, readdirSync, statSync, writeFileSync } from "node:fs";
import { basename, dirname, join, resolve } from "node:path";
import * as inspector from "node:inspector";
import { formatDiagnostic } from "./diagnostics.js";
import { defaultGoJuniorBenchmarkCases, formatGoJuniorBenchmarkReport, runGoJuniorBenchmark } from "./bench.js";
import { createNodeSourcePackageProvider } from "./nodeHost.js";
export async function benchmarkGoJuniorOnNode(request = {}) {
    const cases = benchmarkCasesFromNodeRequest(request);
    const run = async () => {
        const report = await runGoJuniorBenchmark(cases, request);
        return request.cpuProfilePath ? { ...report, profilePath: request.cpuProfilePath } : report;
    };
    if (!request.cpuProfilePath)
        return run();
    return withV8CpuProfile(request.cpuProfilePath, run);
}
export function benchmarkGoJuniorOnNodeHostPayload(report) {
    return {
        ok: report.ok,
        diagnostics: report.diagnostics.map(formatDiagnostic),
        output: formatGoJuniorBenchmarkReport(report),
        ...(report.profilePath ? { profilePath: report.profilePath } : {}),
        report
    };
}
export function benchmarkGoJuniorOnNodeHostJSON(report) {
    return JSON.stringify(benchmarkGoJuniorOnNodeHostPayload(report));
}
function benchmarkCasesFromNodeRequest(request) {
    if (request.files && request.files.length > 0) {
        return [nodeRequestFilesCase(request.files, request)];
    }
    if (request.target && request.target.trim() !== "") {
        return [nodeRequestFilesCase(readBenchmarkTarget(request.target), request)];
    }
    return defaultGoJuniorBenchmarkCases();
}
function nodeRequestFilesCase(files, request) {
    const packageName = request.packageName ?? packageNameFromSource(files);
    const sourcePackageProvider = createNodeSourcePackageProvider(request.sourceRoots ?? []);
    return {
        name: request.name ?? request.importPath ?? packageName ?? "target",
        files,
        ...(request.importPath ? { importPath: request.importPath } : {}),
        ...(packageName ? { packageName } : {}),
        ...(request.phaseSet ? { phaseSet: request.phaseSet } : {}),
        ...(sourcePackageProvider ? { sourcePackageProvider } : {}),
        ...(request.artifactRoot ? { artifactRoot: request.artifactRoot } : {}),
        ...(request.packageCacheParent ? { packageCacheParent: request.packageCacheParent } : {})
    };
}
function readBenchmarkTarget(target) {
    const path = resolve(target);
    const stat = statSync(path);
    if (!stat.isDirectory()) {
        return [{
                filename: path,
                source: readFileSync(path, "utf8")
            }];
    }
    return readdirSync(path)
        .filter((name) => !name.startsWith(".") && !name.startsWith("_") && name.endsWith(".go") && !name.endsWith("_test.go"))
        .sort()
        .map((name) => {
        const filename = join(path, name);
        return {
            filename,
            source: readFileSync(filename, "utf8")
        };
    });
}
function packageNameFromSource(files) {
    for (const file of files) {
        const match = /^\s*package\s+([A-Za-z_]\w*)/m.exec(file.source);
        if (match?.[1])
            return match[1];
    }
    return undefined;
}
async function withV8CpuProfile(profilePath, body) {
    const resolved = resolve(profilePath);
    const session = new inspector.Session();
    let result;
    let bodyError;
    session.connect();
    try {
        await inspectorPost(session, "Profiler.enable");
        await inspectorPost(session, "Profiler.start");
        try {
            result = await body();
        }
        catch (error) {
            bodyError = error;
        }
        const stopped = await inspectorPost(session, "Profiler.stop");
        mkdirSync(dirname(resolved), { recursive: true });
        writeFileSync(resolved, JSON.stringify(stopped.profile), "utf8");
    }
    finally {
        session.disconnect();
    }
    if (bodyError !== undefined)
        throw bodyError;
    return result;
}
function inspectorPost(session, method, params) {
    return new Promise((resolvePost, rejectPost) => {
        session.post(method, params ?? {}, (error, value) => {
            if (error) {
                rejectPost(error);
                return;
            }
            resolvePost(value);
        });
    });
}
export function normalizeBenchmarkCacheMode(value) {
    if (value === undefined || value === "")
        return undefined;
    if (value === "cold" || value === "warm")
        return value;
    throw new Error(`unknown benchmark cache mode ${value}; expected cold or warm`);
}
export function normalizeBenchmarkPhaseSet(value) {
    if (value === undefined || value === "")
        return undefined;
    if (value === "front-end" || value === "package-build" || value === "front-end-and-build")
        return value;
    throw new Error(`unknown benchmark phase set ${value}; expected front-end, package-build, or front-end-and-build`);
}
export function defaultCpuProfilePath(target) {
    const stem = target && target.trim() !== "" ? basename(target).replace(/[^A-Za-z0-9_.-]/g, "_") : "builtin";
    const dir = existsSync("/tmp") ? "/tmp" : ".";
    return join(dir, `gojr-${stem}.cpuprofile`);
}
