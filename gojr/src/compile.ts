import { REPL_FILENAME, SourceFile } from "./diagnostics.js";
import { checkFrontSourceFiles } from "./front/checker.js";
import type { CheckResult } from "./front/checker.js";
import type { PackageInfo } from "./front/types.js";
import { EvaluationOptions, typeCheckConfig } from "./runtime.js";

export interface CompileResult {
  readonly diagnostics: CheckResult["diagnostics"];
  readonly checked: CheckResult;
}

export interface CompilePackageOptions extends EvaluationOptions {
  importPath?: string;
  packageName?: string;
}

export interface CompilePackageResult extends CompileResult {
  readonly packageInfo: PackageInfo;
}

export function compileSource(source: string, options: EvaluationOptions = {}): CompileResult {
  return compileSourceFiles([{
    filename: options.filename ?? REPL_FILENAME,
    source
  }], options);
}

export function compileSourceFiles(files: SourceFile[], options: EvaluationOptions = {}): CompileResult {
  const checked = checkFrontSourceFiles(files, typeCheckConfig(options));
  return {
    diagnostics: checked.diagnostics,
    checked
  };
}

export function compilePackageSourceFiles(files: SourceFile[], options: CompilePackageOptions = {}): CompilePackageResult {
  const packageName = options.packageName ?? packageNameFromSourceFiles(files) ?? "main";
  const checked = checkFrontSourceFiles(files, {
    ...typeCheckConfig(options),
    packageName,
    packagePath: options.importPath ?? packageName
  });
  return {
    diagnostics: checked.diagnostics,
    checked,
    packageInfo: checked.pkg
  };
}

function packageNameFromSourceFiles(files: SourceFile[]): string | undefined {
  for (const file of files) {
    const match = /^\s*package\s+([A-Za-z_]\w*)/m.exec(file.source);
    if (match?.[1]) return match[1];
  }
  return undefined;
}
