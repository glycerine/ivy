import { REPL_FILENAME } from "./diagnostics.js";
import { checkGoJuniorSourceFiles } from "./typecheck.js";
import { typeCheckConfig } from "./runtime.js";
import { parseFrontSourceFiles } from "./front/parser.js";
export function compileSource(source, options = {}) {
    return compileSourceFiles([{
            filename: options.filename ?? REPL_FILENAME,
            source
        }], options);
}
export function compileSourceFiles(files, options = {}) {
    const checked = checkGoJuniorSourceFiles(files, typeCheckConfig(options));
    return {
        diagnostics: checked.diagnostics,
        checked
    };
}
export function compilePackageSourceFiles(files, options = {}) {
    const packageName = options.packageName ?? packageNameFromSourceFiles(files) ?? "main";
    const checked = checkGoJuniorSourceFiles(files, {
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
    return parseFrontSourceFiles(files).files.find((file) => file.name)?.name?.name;
}
