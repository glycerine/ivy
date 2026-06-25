#!/usr/bin/env node
import { readFile } from "node:fs/promises";
import { parseProgram } from "./index.js";
import { formatDiagnostic, REPL_FILENAME, type Diagnostic } from "./diagnostics.js";
import { parseSheetJson, parseSheetsJson } from "./jsonInput.js";
import { evaluateSource, formatReplValue, SheetData } from "./runtime.js";

interface CliOptions {
  command: "parse" | "run" | "eval" | "help";
  file?: string;
  expression?: string;
  sheet?: SheetData;
  sheets?: Record<string, SheetData>;
  randomSeed?: string;
}

async function main(argv: string[]): Promise<number> {
  const options = parseArgs(argv);
  if (options.command === "help") {
    printUsage();
    return 0;
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
  if (command !== "parse" && command !== "run" && command !== "eval") {
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
