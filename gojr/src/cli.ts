#!/usr/bin/env node
import { readFile } from "node:fs/promises";
import { parseProgram } from "./index.js";
import { formatDiagnostic, REPL_FILENAME, type Diagnostic } from "./diagnostics.js";
import { parseSheetJson, parseSheetsJson } from "./jsonInput.js";
import {
  benchmarkGoJuniorOnNode,
  formatGoJuniorBenchmarkReport,
  normalizeBenchmarkCacheMode,
  normalizeBenchmarkPhaseSet
} from "./index.js";
import { evaluateSource, formatReplValue, SheetData } from "./runtime.js";

interface CliOptions {
  command: "parse" | "run" | "eval" | "bench" | "help";
  file?: string;
  expression?: string;
  sheet?: SheetData;
  sheets?: Record<string, SheetData>;
  randomSeed?: string;
  json?: boolean;
  iterations?: number;
  warmupIterations?: number;
  caseNames?: string[];
  cacheMode?: "cold" | "warm";
  phaseSet?: "front-end" | "package-build" | "front-end-and-build";
  cpuProfilePath?: string;
}

async function main(argv: string[]): Promise<number> {
  const options = parseArgs(argv);
  if (options.command === "help") {
    printUsage();
    return 0;
  }

  if (options.command === "bench") {
    const report = await benchmarkGoJuniorOnNode({
      ...(options.file ? { target: options.file } : {}),
      ...(options.iterations !== undefined ? { iterations: options.iterations } : {}),
      ...(options.warmupIterations !== undefined ? { warmupIterations: options.warmupIterations } : {}),
      ...(options.caseNames ? { caseNames: options.caseNames } : {}),
      ...(options.cacheMode ? { cacheMode: options.cacheMode } : {}),
      ...(options.phaseSet ? { phaseSet: options.phaseSet } : {}),
      ...(options.cpuProfilePath ? { cpuProfilePath: options.cpuProfilePath } : {})
    });
    if (options.json) {
      console.log(JSON.stringify(report, jsonReplacer, 2));
    } else {
      process.stdout.write(formatGoJuniorBenchmarkReport(report));
    }
    return report.ok ? 0 : 1;
  }

  const source = options.command === "eval"
    ? options.expression ?? ""
    : await readSource(options.file);
  const filename = options.command === "eval" ? REPL_FILENAME : options.file && options.file !== "-" ? options.file : REPL_FILENAME;

  if (options.command === "parse") {
    const result = parseProgram(source, filename);
    printDiagnostics(result.diagnostics);
    console.log(JSON.stringify(result.ast, jsonReplacer, 2));
    return result.diagnostics.some((diagnostic) => diagnostic.severity === "error") ? 1 : 0;
  }

  const result = await evaluateSource(source, {
    filename,
    ...(options.sheet ? { sheet: options.sheet } : {}),
    ...(options.sheets ? { sheets: options.sheets } : {}),
    ...(options.randomSeed !== undefined ? { randomSeed: options.randomSeed } : {}),
    stdout: (text) => {
      process.stdout.write(text);
    }
  });
  printDiagnostics(result.diagnostics);
  if (result.diagnostics.some((diagnostic) => diagnostic.severity === "error")) {
    return 1;
  }
  if (result.exitCode !== undefined) {
    return result.exitCode;
  }

  if (options.command === "eval") {
    if (result.values) {
      console.log(result.values.map(formatReplValue).join(", "));
    } else if (result.value !== undefined) {
      console.log(formatReplValue(result.value));
    }
  } else if (result.values) {
    console.log(result.values.map(formatReplValue).join(", "));
  }

  return 0;
}

function parseArgs(argv: string[]): CliOptions {
  const args = [...argv];
  const command = args.shift();
  if (!command || command === "help" || command === "--help" || command === "-h") {
    return { command: "help" };
  }
  if (command !== "parse" && command !== "run" && command !== "eval" && command !== "bench") {
    throw new Error(`unknown command: ${command}`);
  }

  const options: CliOptions = { command };
  while (args.length > 0) {
    const arg = args.shift();
    if (!arg) break;
    if (arg === "--sheet-json") {
      options.sheet = parseSheetJson(args.shift() ?? "{}");
      continue;
    }
    if (arg === "--sheets-json") {
      options.sheets = parseSheetsJson(args.shift() ?? "{}");
      continue;
    }
    if (arg === "--seed" || arg === "--random-seed") {
      const seed = args.shift();
      if (seed === undefined) throw new Error(`${arg} expects a seed value`);
      options.randomSeed = seed;
      continue;
    }
    if (command === "bench" && (arg === "--json" || arg === "-json")) {
      options.json = true;
      continue;
    }
    if (command === "bench" && (arg === "-n" || arg === "--iterations")) {
      options.iterations = parsePositiveInt(args.shift(), arg);
      continue;
    }
    if (command === "bench" && (arg === "--warmup" || arg === "-warmup")) {
      options.warmupIterations = parseNonNegativeInt(args.shift(), arg);
      continue;
    }
    if (command === "bench" && (arg === "--case" || arg === "-case")) {
      const name = args.shift();
      if (!name) throw new Error(`${arg} expects a benchmark case name`);
      options.caseNames = [...(options.caseNames ?? []), name];
      continue;
    }
    if (command === "bench" && (arg === "--cache" || arg === "-cache")) {
      const cacheMode = normalizeBenchmarkCacheMode(args.shift());
      if (cacheMode) options.cacheMode = cacheMode;
      continue;
    }
    if (command === "bench" && (arg === "--phase" || arg === "-phase")) {
      const phaseSet = normalizeBenchmarkPhaseSet(args.shift());
      if (phaseSet) options.phaseSet = phaseSet;
      continue;
    }
    if (command === "bench" && (arg === "--cpuprofile" || arg === "-cpuprofile")) {
      const path = args.shift();
      if (!path) throw new Error(`${arg} expects a profile output path`);
      options.cpuProfilePath = path;
      continue;
    }
    if (arg.startsWith("--")) {
      throw new Error(`unknown option: ${arg}`);
    }
    if (command === "eval") {
      options.expression = [arg, ...args].join(" ");
      break;
    }
    options.file = arg;
  }
  return options;
}

function parsePositiveInt(value: string | undefined, option: string): number {
  const number = Number.parseInt(value ?? "", 10);
  if (!Number.isFinite(number) || number <= 0) throw new Error(`${option} expects a positive integer`);
  return number;
}

function parseNonNegativeInt(value: string | undefined, option: string): number {
  const number = Number.parseInt(value ?? "", 10);
  if (!Number.isFinite(number) || number < 0) throw new Error(`${option} expects a non-negative integer`);
  return number;
}

async function readSource(file: string | undefined): Promise<string> {
  if (file && file !== "-") {
    return readFile(file, "utf8");
  }
  const chunks: Buffer[] = [];
  for await (const chunk of process.stdin) {
    chunks.push(Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk));
  }
  return Buffer.concat(chunks).toString("utf8");
}

function printDiagnostics(diagnostics: Diagnostic[]): void {
  for (const diagnostic of diagnostics) {
    const line = formatDiagnostic(diagnostic);
    if (diagnostic.severity === "error") {
      console.error(line);
    } else {
      console.warn(line);
    }
  }
}

function jsonReplacer(_key: string, value: unknown): unknown {
  return typeof value === "bigint" ? `${value}n` : value;
}

function printUsage(): void {
  console.log(`gojr parse [file]
gojr run [file] [--sheet-json '{"A1":1}'] [--seed replay-seed]
gojr eval <source> [--sheet-json '{"A1":1}'] [--seed replay-seed]
gojr bench [-n N] [--warmup N] [--case NAME] [--cache cold|warm] [--phase front-end|package-build|front-end-and-build] [--cpuprofile PATH] [file|dir]

Use "-" or omit file to read from stdin. JSON strings are always strings.
JSON integers become exact integer values. JSON numbers with a decimal point
or exponent become float64 values.`);
}

main(process.argv.slice(2)).then((code) => {
  process.exitCode = code;
}).catch((error: unknown) => {
  console.error(error instanceof Error ? error.message : String(error));
  process.exitCode = 1;
});
