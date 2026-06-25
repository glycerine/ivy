import { REPL_FILENAME, SourceFile } from "./diagnostics.js";
import { checkGoJuniorSourceFiles, type GoJuniorCheckResult } from "./typecheck.js";
import type { Package as GoTypesPackage } from "./go/types/index.js";
import { EvaluationOptions, typeCheckConfig } from "./runtime.js";
import { parseFrontSourceFiles } from "./front/parser.js";

export interface CompileResult {
  readonly diagnostics: GoJuniorCheckResult["diagnostics"];
  readonly checked: GoJuniorCheckResult;
}

export interface CompilePackageOptions extends EvaluationOptions {
  importPath?: string;
  packageName?: string;
}

export interface CompilePackageResult extends CompileResult {
  readonly packageInfo: GoTypesPackage;
}

export function compileSource(source: string, options: EvaluationOptions = {}): CompileResult {
  return compileSourceFiles([{
    filename: options.filename ?? REPL_FILENAME,
    source
  }], options);
}

export function compileSourceFiles(files: SourceFile[], options: EvaluationOptions = {}): CompileResult {
  const checked = checkGoJuniorSourceFiles(files, typeCheckConfig(options));
  return {
    diagnostics: checked.diagnostics,
    checked
  };
}

export function compilePackageSourceFiles(files: SourceFile[], options: CompilePackageOptions = {}): CompilePackageResult {
  const packageName = options.packageName ?? packageNameFromSourceFiles(files) ?? "main";
  const checked = checkGoJuniorSourceFiles(files, {
    ...typeCheckConfig(options),
    packageName,
    packagePath: options.importPath ?? packageName,
    autoImportFmt: false
  });
  return {
    diagnostics: checked.diagnostics,
    checked,
    packageInfo: checked.pkg
  };
}

function packageNameFromSourceFiles(files: SourceFile[]): string | undefined {
  return parseFrontSourceFiles(files).files.find((file) => file.name)?.name?.name;
}
