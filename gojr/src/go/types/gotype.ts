// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore

// Build this command explicitly: go build gotype.go

import { existsSync, readFileSync, readdirSync, statSync } from "node:fs";
import { join } from "node:path";
import { Print } from "../../front/ast.js";
import type { File as AstFile } from "../../front/ast.js";
import { AllErrors, ParseComments, ParseFile, SkipObjectResolution, Trace, type Mode } from "../../front/parser.js";
import { PrintError } from "../../front/scanner.js";
import { Config } from "./api.js";
import { SizesFor } from "./sizes.js";

export let testFiles = false;
export let xtestFiles = false;
export let allErrors = false;
export let verbose = false;
export let compiler = "source";

export let printAST = false;
export let printTrace = false;
export let parseComments = false;
export let panicOnError = false;

export const fset = {
  files: [] as { filename: string; lineCount: number }[],
  Iterate(fn: (file: { filename: string; LineCount: () => number }) => boolean): void {
    for (const file of this.files) {
      if (!fn({ filename: file.filename, LineCount: () => file.lineCount })) {
        return;
      }
    }
  }
};
export let errorCount = 0;
export let sequential = false;
export let parserMode: Mode = 0;

export function initParserMode(): void {
  parserMode = SkipObjectResolution;
  if (allErrors) {
    parserMode |= AllErrors;
  }
  if (printAST) {
    sequential = true;
  }
  if (printTrace) {
    parserMode |= Trace;
    sequential = true;
  }
  if (parseComments && (printAST || printTrace)) {
    parserMode |= ParseComments;
  }
}

export const usageString = `usage: gotype [flags] [path ...]

The gotype command, like the front-end of a Go compiler, parses and
type-checks a single Go package. Errors are reported if the analysis
fails; otherwise gotype is quiet (unless -v is set).

Without a list of paths, gotype reads from standard input, which
must provide a single Go source file defining a complete package.

With a single directory argument, gotype checks the Go files in
that directory, comprising a single package. Use -t to include the
(in-package) _test.go files. Use -x to type check only external
test files.

Otherwise, each path must be the filename of a Go file belonging
to the same package.

Imports are processed by importing directly from the source of
imported packages (default), or by importing from compiled and
installed packages (by setting -c to the respective compiler).

The -c flag must be set to a compiler ("gc", "gccgo") when type-
checking packages containing imports with relative import paths
(import "./mypkg") because the source importer cannot know which
files to include for such packages.
`;

export function usage(): void {
  process.stderr.write(usageString + "\n");
  printDefaults();
  process.exit(2);
}

export function report(err: unknown): void {
  if (panicOnError) {
    throw err;
  }
  PrintError((text: string) => process.stderr.write(text), err);
  const list = err as { length?: number; Len?: () => number } | null;
  if (list !== null && typeof list === "object" && Array.isArray(list)) {
    errorCount += list.length;
    return;
  }
  errorCount++;
}

// parse may be called concurrently.
export function parse(filename: string, src: unknown): [AstFile | undefined, unknown] {
  if (verbose) {
    console.log(filename);
  }
  const text = src === null || src === undefined ? readFileSync(filename, "utf8") : src;
  recordFile(filename, text);
  const [file, err] = ParseFile(fset, filename, text, parserMode); // ok to access fset concurrently
  if (printAST) {
    Print(fset, file);
  }
  return [file, err];
}

export function parseStdin(): [AstFile | undefined, unknown] {
  let src: string;
  try {
    src = readFileSync(0, "utf8");
  } catch (err) {
    return [undefined, err];
  }
  return parse("<standard input>", src);
}

export function parseFiles(dir: string, filenames: string[]): [AstFile[], unknown] {
  let files = new Array<AstFile | undefined>(filenames.length);
  const errors = new Array<unknown>(filenames.length);

  for (let i = 0; i < filenames.length; i++) {
    const filepath = join(dir, filenames[i]!);
    [files[i], errors[i]] = parse(filepath, null);
    if (sequential) {
      // synchronous TypeScript execution has already waited
    }
  }

  // If there are errors, return the first one for deterministic results.
  let first: unknown;
  for (const err of errors) {
    if (err !== undefined && err !== null) {
      first = err;
      // If we have an error, some files may be nil.
      // Remove them. (The go/parser always returns
      // a possibly partial AST even in the presence
      // of errors, except if the file doesn't exist
      // in the first place, in which case it cannot
      // matter.)
      let i = 0;
      for (const f of files) {
        if (f !== undefined) {
          files[i] = f;
          i++;
        }
      }
      files = files.slice(0, i);
      break;
    }
  }

  return [files.filter((file): file is AstFile => file !== undefined), first];
}

export function parseDir(dir: string): [AstFile[] | undefined, unknown] {
  let pkginfo: pkgInfo;
  try {
    pkginfo = importDir(dir);
  } catch (err) {
    if (!isNoGoError(err)) {
      return [undefined, err];
    }
    pkginfo = new pkgInfo();
  }

  if (xtestFiles) {
    return parseFiles(dir, pkginfo.XTestGoFiles);
  }

  let filenames = [...pkginfo.GoFiles, ...pkginfo.CgoFiles];
  if (testFiles) {
    filenames = [...filenames, ...pkginfo.TestGoFiles];
  }
  return parseFiles(dir, filenames);
}

export function getPkgFiles(args: string[]): [AstFile[] | undefined, unknown] {
  if (args.length === 0) {
    // stdin
    const [file, err] = parseStdin();
    if (err !== undefined && err !== null) {
      return [undefined, err];
    }
    return [[file!], undefined];
  }

  if (args.length === 1) {
    // possibly a directory
    const path = args[0]!;
    let info: ReturnType<typeof statSync>;
    try {
      info = statSync(path);
    } catch (err) {
      return [undefined, err];
    }
    if (info.isDirectory()) {
      return parseDir(path);
    }
  }

  // list of files
  return parseFiles("", args);
}

export function checkPkgFiles(files: AstFile[] | undefined): void {
  class bailout {}

  // if checkPkgFiles is called multiple times, set up conf only once
  const conf = new Config();
  conf.FakeImportC = true;
  conf.Error = (err: unknown) => {
    if (!allErrors && errorCount >= 10) {
      throw new bailout();
    }
    report(err);
  };
  conf.Importer = null;
  conf.Sizes = SizesFor("gc", archName());

  try {
    const path = "pkg"; // any non-empty string will do for now
    conf.Check(path, fset, files ?? [], null);
  } catch (p) {
    if (p instanceof bailout) {
      // normal return or early exit
    } else {
      // re-panic
      throw p;
    }
  }
}

export function printStats(d: number): void {
  let fileCount = 0;
  let lineCount = 0;
  fset.Iterate((f) => {
    fileCount++;
    lineCount += f.LineCount();
    return true;
  });

  console.log(
    `${durationString(d)} (${fileCount} files, ${lineCount} lines, ${Math.trunc(lineCount / (d / 1000))} lines/s)`
  );
}

export function main(argv = process.argv.slice(2)): void {
  const args = parseFlags(argv);
  initParserMode();

  const start = Date.now();

  const [files, err] = getPkgFiles(args);
  if (err !== undefined && err !== null) {
    report(err);
    // ok to continue (files may be empty, but not nil)
  }

  checkPkgFiles(files);
  if (errorCount > 0) {
    process.exit(2);
  }

  if (verbose) {
    printStats(Date.now() - start);
  }
}

class pkgInfo {
  public GoFiles: string[] = [];
  public CgoFiles: string[] = [];
  public TestGoFiles: string[] = [];
  public XTestGoFiles: string[] = [];
}

class NoGoError extends Error {}

function importDir(dir: string): pkgInfo {
  const info = new pkgInfo();
  if (!existsSync(dir)) {
    throw new Error(`directory not found: ${dir}`);
  }
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    if (!entry.isFile() || !entry.name.endsWith(".go")) {
      continue;
    }
    if (entry.name.endsWith("_test.go")) {
      const src = readFileSync(join(dir, entry.name), "utf8");
      if (packageName(src).endsWith("_test")) {
        info.XTestGoFiles.push(entry.name);
      } else {
        info.TestGoFiles.push(entry.name);
      }
      continue;
    }
    const src = readFileSync(join(dir, entry.name), "utf8");
    if (src.includes('import "C"') || src.includes("import `C`")) {
      info.CgoFiles.push(entry.name);
    } else {
      info.GoFiles.push(entry.name);
    }
  }
  if (info.GoFiles.length === 0 && info.CgoFiles.length === 0 && info.TestGoFiles.length === 0 && info.XTestGoFiles.length === 0) {
    throw new NoGoError(`no Go files in ${dir}`);
  }
  return info;
}

function isNoGoError(err: unknown): boolean {
  return err instanceof NoGoError;
}

function packageName(src: string): string {
  return src.match(/^\s*package\s+([A-Za-z_][A-Za-z0-9_]*)/m)?.[1] ?? "";
}

function parseFlags(argv: string[]): string[] {
  const args: string[] = [];
  for (let i = 0; i < argv.length; i++) {
    const arg = argv[i]!;
    if (arg === "-h" || arg === "-help" || arg === "--help") {
      usage();
    } else if (arg === "-t") {
      testFiles = true;
    } else if (arg === "-x") {
      xtestFiles = true;
    } else if (arg === "-e") {
      allErrors = true;
    } else if (arg === "-v") {
      verbose = true;
    } else if (arg === "-ast") {
      printAST = true;
    } else if (arg === "-trace") {
      printTrace = true;
    } else if (arg === "-comments") {
      parseComments = true;
    } else if (arg === "-panic") {
      panicOnError = true;
    } else if (arg === "-c") {
      i++;
      compiler = argv[i] ?? compiler;
    } else if (arg.startsWith("-c=")) {
      compiler = arg.slice("-c=".length);
    } else if (arg.startsWith("-")) {
      usage();
    } else {
      args.push(arg);
    }
  }
  return args;
}

function printDefaults(): void {
  process.stderr.write(`  -t
    include in-package test files in a directory
  -x
    consider only external test files in a directory
  -e
    report all errors, not just the first 10
  -v
    verbose mode
  -c string
    compiler used for installed packages (gc, gccgo, or source) (default "source")
  -ast
    print AST
  -trace
    print parse trace
  -comments
    parse comments
  -panic
    panic on first error
`);
}

function recordFile(filename: string, src: unknown): void {
  if (typeof src !== "string") {
    return;
  }
  fset.files.push({ filename, lineCount: src.split(/\r\n|\r|\n/u).length });
}

function archName(): string {
  switch (process.arch) {
    case "x64":
      return "amd64";
    case "ia32":
      return "386";
    case "arm64":
      return "arm64";
    default:
      return process.arch;
  }
}

function durationString(ms: number): string {
  if (ms < 1000) {
    return `${ms}ms`;
  }
  return `${(ms / 1000).toFixed(3)}s`;
}
