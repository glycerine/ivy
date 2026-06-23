import { REPL_FILENAME } from "./diagnostics.js";
import { checkFrontSourceFiles } from "./front/checker.js";
import { typeCheckConfig } from "./runtime.js";
export function compileSource(source, options = {}) {
    return compileSourceFiles([{
            filename: options.filename ?? REPL_FILENAME,
            source
        }], options);
}
export function compileSourceFiles(files, options = {}) {
    const checked = checkFrontSourceFiles(files, typeCheckConfig(options));
    return {
        diagnostics: checked.diagnostics,
        checked
    };
}
export function compilePackageSourceFiles(files, options = {}) {
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
function packageNameFromSourceFiles(files) {
    for (const file of files) {
        const match = /^\s*package\s+([A-Za-z_]\w*)/m.exec(file.source);
        if (match?.[1])
            return match[1];
    }
    return undefined;
}
