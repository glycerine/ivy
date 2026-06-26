import { existsSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { basename, isAbsolute, join } from "node:path";
import { spawnSync } from "node:child_process";

export interface ClangDiscoveryOptions {
  env?: Record<string, string | undefined>;
  candidates?: string[];
  probe?: (clangPath: string) => ClangProbeResult;
}

export interface ClangProbeResult {
  ok: boolean;
  message?: string;
}

export interface ClangDiscoveryResult {
  ok: boolean;
  clangPath?: string;
  attempted: Array<{ path: string; ok: boolean; message?: string }>;
}

export interface CompileCToWasmOptions {
  clangPath: string;
  source: string;
  exports: string[];
  importMemory?: boolean;
  optimize?: string;
}

export function discoverClangForWasm(options: ClangDiscoveryOptions = {}): ClangDiscoveryResult {
  const env = options.env ?? process.env;
  const probe = options.probe ?? probeClangSupportsWasm;
  const candidates = clangCandidates(env, options.candidates);
  const attempted: Array<{ path: string; ok: boolean; message?: string }> = [];
  for (const candidate of candidates) {
    if (isAbsolute(candidate) && !existsSync(candidate)) {
      attempted.push({ path: candidate, ok: false, message: "not found" });
      continue;
    }
    const result = probe(candidate);
    attempted.push({
      path: candidate,
      ok: result.ok,
      ...(result.message ? { message: result.message } : {})
    });
    if (result.ok) {
      return {
        ok: true,
        clangPath: candidate,
        attempted
      };
    }
  }
  return {
    ok: false,
    attempted
  };
}

export function clangCandidates(env: Record<string, string | undefined> = process.env, extra: string[] | undefined = undefined): string[] {
  const out: string[] = [];
  const add = (path: string | undefined): void => {
    const text = String(path ?? "").trim();
    if (text !== "" && !out.includes(text)) out.push(text);
  };
  add(env.GOJR_CLANG);
  for (const item of extra ?? []) add(item);
  add("/usr/local/opt/llvm/bin/clang");
  add("/opt/homebrew/opt/llvm/bin/clang");
  add("clang");
  return out;
}

export function probeClangSupportsWasm(clangPath: string): ClangProbeResult {
  const dir = mkdtempSync(join(tmpdir(), "gojr-clang-probe-"));
  try {
    const out = join(dir, "probe.o");
    const result = spawnSync(clangPath, [
      "--target=wasm32",
      "-O2",
      "-ffreestanding",
      "-nostdlib",
      "-x",
      "c",
      "-c",
      "-o",
      out,
      "-"
    ], {
      input: "int gojr_probe(void) { return 0; }\n",
      encoding: "utf8"
    });
    if (result.error) {
      return { ok: false, message: result.error.message };
    }
    if (result.status !== 0) {
      const message = [result.stderr, result.stdout].filter(Boolean).join("\n").trim();
      return { ok: false, message: message || `exit status ${result.status}` };
    }
    return { ok: true };
  } finally {
    rmSync(dir, { recursive: true, force: true });
  }
}

export function compileCToWasm(options: CompileCToWasmOptions): Uint8Array {
  if (options.exports.length === 0) throw new Error("compileCToWasm requires at least one export");
  const dir = mkdtempSync(join(tmpdir(), "gojr-clang-stencil-"));
  try {
    const stem = safeStem(options.exports[0] ?? "stencil");
    const sourcePath = join(dir, `${stem}.c`);
    const objectPath = join(dir, `${stem}.o`);
    const wasmPath = join(dir, `${stem}.wasm`);
    writeFileSync(sourcePath, options.source, "utf8");
    runClang(options.clangPath, [
      "--target=wasm32",
      options.optimize ?? "-O2",
      "-ffreestanding",
      "-nostdlib",
      "-c",
      sourcePath,
      "-o",
      objectPath
    ]);
    const linkArgs = [
      "--target=wasm32",
      options.optimize ?? "-O2",
      "-ffreestanding",
      "-nostdlib",
      "-Wl,--no-entry",
      ...(options.importMemory ? ["-Wl,--import-memory"] : []),
      ...options.exports.map((name) => `-Wl,--export=${name}`),
      objectPath,
      "-o",
      wasmPath
    ];
    runClang(options.clangPath, linkArgs);
    return new Uint8Array(readFileSync(wasmPath));
  } finally {
    rmSync(dir, { recursive: true, force: true });
  }
}

function runClang(clangPath: string, args: string[]): void {
  const result = spawnSync(clangPath, args, { encoding: "utf8" });
  if (result.error) throw result.error;
  if (result.status !== 0) {
    const message = [result.stderr, result.stdout].filter(Boolean).join("\n").trim();
    throw new Error(`${basename(clangPath)} failed: ${message || `exit status ${result.status}`}`);
  }
}

function safeStem(name: string): string {
  const text = name.replace(/[^A-Za-z0-9_]/g, "_").replace(/^[^A-Za-z_]/, "_");
  return text || "stencil";
}
