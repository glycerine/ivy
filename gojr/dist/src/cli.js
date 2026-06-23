#!/usr/bin/env node
import { readFile } from "node:fs/promises";
import { parseProgram } from "./index.js";
import { parseSheetJson, parseSheetsJson } from "./jsonInput.js";
import { evaluateSource, formatReplValue } from "./runtime.js";
async function main(argv) {
    const options = parseArgs(argv);
    if (options.command === "help") {
        printUsage();
        return 0;
    }
    const source = options.command === "eval"
        ? options.expression ?? ""
        : await readSource(options.file);
    if (options.command === "parse") {
        const result = parseProgram(source);
        printDiagnostics(result.diagnostics);
        console.log(JSON.stringify(result.ast, jsonReplacer, 2));
        return result.diagnostics.some((diagnostic) => diagnostic.severity === "error") ? 1 : 0;
    }
    const result = evaluateSource(source, {
        ...(options.sheet ? { sheet: options.sheet } : {}),
        ...(options.sheets ? { sheets: options.sheets } : {}),
        stdout: (text) => {
            process.stdout.write(text);
        }
    });
    printDiagnostics(result.diagnostics);
    if (result.diagnostics.some((diagnostic) => diagnostic.severity === "error")) {
        return 1;
    }
    if (options.command === "eval") {
        if (result.values) {
            console.log(result.values.map(formatReplValue).join(", "));
        }
        else if (result.value !== undefined) {
            console.log(formatReplValue(result.value));
        }
    }
    else if (result.values) {
        console.log(result.values.map(formatReplValue).join(", "));
    }
    return 0;
}
function parseArgs(argv) {
    const args = [...argv];
    const command = args.shift();
    if (!command || command === "help" || command === "--help" || command === "-h") {
        return { command: "help" };
    }
    if (command !== "parse" && command !== "run" && command !== "eval") {
        throw new Error(`unknown command: ${command}`);
    }
    const options = { command };
    while (args.length > 0) {
        const arg = args.shift();
        if (!arg)
            break;
        if (arg === "--sheet-json") {
            options.sheet = parseSheetJson(args.shift() ?? "{}");
            continue;
        }
        if (arg === "--sheets-json") {
            options.sheets = parseSheetsJson(args.shift() ?? "{}");
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
async function readSource(file) {
    if (file && file !== "-") {
        return readFile(file, "utf8");
    }
    const chunks = [];
    for await (const chunk of process.stdin) {
        chunks.push(Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk));
    }
    return Buffer.concat(chunks).toString("utf8");
}
function printDiagnostics(diagnostics) {
    for (const diagnostic of diagnostics) {
        const location = diagnostic.span ? `${diagnostic.span.line}:${diagnostic.span.column}: ` : "";
        const line = `${location}${diagnostic.severity} ${diagnostic.code}: ${diagnostic.message}`;
        if (diagnostic.severity === "error") {
            console.error(line);
        }
        else {
            console.warn(line);
        }
    }
}
function jsonReplacer(_key, value) {
    return typeof value === "bigint" ? `${value}n` : value;
}
function printUsage() {
    console.log(`gojr parse [file]
gojr run [file] [--sheet-json '{"A1":1}']
gojr eval <source> [--sheet-json '{"A1":1}']

Use "-" or omit file to read from stdin. JSON strings are always strings.
JSON integers become exact integer values. JSON numbers with a decimal point
or exponent become float64 values.`);
}
main(process.argv.slice(2)).then((code) => {
    process.exitCode = code;
}).catch((error) => {
    console.error(error instanceof Error ? error.message : String(error));
    process.exitCode = 1;
});
