"use strict";
(() => {
    const root = globalThis;
    function dirname(path) {
        const index = path.lastIndexOf("/");
        return index <= 0 ? "/" : path.slice(0, index);
    }
    function normalizePath(path) {
        const parts = [];
        for (const part of path.split("/")) {
            if (!part || part === ".")
                continue;
            if (part === "..") {
                parts.pop();
                continue;
            }
            parts.push(part);
        }
        return "/" + parts.join("/");
    }
    function resolveModule(specifier, parent) {
        if (!specifier.startsWith(".") && !specifier.startsWith("/"))
            return specifier;
        const base = specifier.startsWith("/") ? specifier : `${dirname(parent)}/${specifier}`;
        return normalizePath(base);
    }
    function importListToDestructure(list) {
        return list.split(",")
            .map((item) => item.trim())
            .filter(Boolean)
            .map((item) => {
            const match = /^([A-Za-z_$][\w$]*)(?:\s+as\s+([A-Za-z_$][\w$]*))?$/.exec(item);
            if (!match)
                throw new Error(`unsupported embedded import specifier: ${item}`);
            return match[2] ? `${match[1]}: ${match[2]}` : match[1];
        })
            .join(", ");
    }
    function exportList(list) {
        return list.split(",")
            .map((item) => item.trim())
            .filter(Boolean)
            .map((item) => {
            const match = /^([A-Za-z_$][\w$]*)(?:\s+as\s+([A-Za-z_$][\w$]*))?$/.exec(item);
            if (!match)
                throw new Error(`unsupported embedded export specifier: ${item}`);
            return [match[1], match[2] || match[1]];
        });
    }
    function transformModule(source, filename) {
        const exportedNames = [];
        const earlyFunctionExports = [];
        source.replace(/^export\s+(async\s+)?function(\*)?\s+([A-Za-z_$][\w$]*)/gm, (_match, _asyncPrefix, _generatorMarker, name) => {
            earlyFunctionExports.push(name);
            return _match;
        });
        source = source.replace(/^export\s+\{\s*\};\s*$/gm, "");
        source = source.replace(/^import\s+\{([^}]+)\}\s+from\s+["']([^"']+)["'];\s*$/gm, (_match, imports, specifier) => `const { ${importListToDestructure(imports)} } = __gojrRequire(${JSON.stringify(specifier)}, __filename);`);
        source = source.replace(/^import\s+\*\s+as\s+([A-Za-z_$][\w$]*)\s+from\s+["']([^"']+)["'];\s*$/gm, (_match, name, specifier) => `const ${name} = __gojrRequire(${JSON.stringify(specifier)}, __filename);`);
        source = source.replace(/^import\s+["']([^"']+)["'];\s*$/gm, (_match, specifier) => `__gojrRequire(${JSON.stringify(specifier)}, __filename);`);
        source = source.replace(/^export\s+\{([^}]+)\}\s+from\s+["']([^"']+)["'];\s*$/gm, (_match, exportsList, specifier) => `__gojrReExport(${JSON.stringify(specifier)}, ${JSON.stringify(exportList(exportsList))}, exports, __filename);`);
        source = source.replace(/^export\s+\*\s+as\s+([A-Za-z_$][\w$]*)\s+from\s+["']([^"']+)["'];\s*$/gm, (_match, name, specifier) => `exports[${JSON.stringify(name)}] = __gojrRequire(${JSON.stringify(specifier)}, __filename);`);
        source = source.replace(/^export\s+\*\s+from\s+["']([^"']+)["'];\s*$/gm, (_match, specifier) => `__gojrExportAll(${JSON.stringify(specifier)}, exports, __filename);`);
        source = source.replace(/^export\s+\{([^}]*)\};\s*$/gm, (_match, exportsList) => {
            const names = exportList(exportsList);
            if (names.length === 0)
                return "";
            return `globalThis.Object.assign(exports, { ${names.map(([sourceName, exportName]) => `${JSON.stringify(exportName)}: ${sourceName}`).join(", ")} });`;
        });
        source = source.replace(/^export\s+(async\s+)?function(\*)?\s+([A-Za-z_$][\w$]*)/gm, (_match, asyncPrefix, generatorMarker, name) => {
            exportedNames.push(name);
            return `exports[${JSON.stringify(name)}] = ${name};\n${asyncPrefix || ""}function${generatorMarker || ""} ${name}`;
        });
        source = source.replace(/^export\s+(function|class)\s+([A-Za-z_$][\w$]*)/gm, (_match, kind, name) => {
            exportedNames.push(name);
            return `${kind} ${name}`;
        });
        source = source.replace(/^export\s+(const|let|var)\s+([A-Za-z_$][\w$]*)/gm, (_match, kind, name) => {
            exportedNames.push(name);
            return `${kind} ${name}`;
        });
        if (exportedNames.length > 0) {
            source += `\nglobalThis.Object.assign(exports, { ${[...new Set(exportedNames)].join(", ")} });\n`;
        }
        if (earlyFunctionExports.length > 0) {
            source = `${[...new Set(earlyFunctionExports)].map((name) => `exports[${JSON.stringify(name)}] = ${name};`).join("\n")}\n${source}`;
        }
        return source + `\n//# sourceURL=embedded-gojr:${filename}\n`;
    }
    function createEmbeddedModuleLoader(moduleBundleJson) {
        const parsed = JSON.parse(moduleBundleJson || "{\"modules\":[]}");
        const embeddedModules = new Map(parsed.modules.map((module) => [module.path, module.source]));
        const embeddedCache = new Map();
        function reExport(specifier, names, target, parent) {
            const module = embeddedRequire(specifier, parent);
            for (const [sourceName, exportedName] of names) {
                Object.defineProperty(target, exportedName, {
                    enumerable: true,
                    get() {
                        return module[sourceName];
                    }
                });
            }
        }
        function exportAll(specifier, target, parent) {
            const module = embeddedRequire(specifier, parent);
            for (const key of Object.keys(module)) {
                if (key === "default" || Object.prototype.hasOwnProperty.call(target, key))
                    continue;
                Object.defineProperty(target, key, {
                    enumerable: true,
                    get() {
                        return module[key];
                    }
                });
            }
        }
        function embeddedRequire(specifier, parent = "/src/index.js") {
            if (!specifier.startsWith(".") && !specifier.startsWith("/")) {
                const hostRequire = typeof root.require === "function"
                    ? root.require
                    : typeof require === "function"
                        ? require
                        : undefined;
                if (hostRequire === undefined) {
                    throw new Error(`host module require is not available for ${specifier}`);
                }
                return hostRequire(specifier);
            }
            const filename = resolveModule(specifier, parent);
            const cached = embeddedCache.get(filename);
            if (cached)
                return cached.exports;
            const source = embeddedModules.get(filename);
            if (source === undefined)
                throw new Error(`embedded Go-junior module not found: ${filename}`);
            const module = { exports: {} };
            embeddedCache.set(filename, module);
            const transformed = transformModule(source, filename);
            const fn = new Function("exports", "module", "__gojrRequire", "__gojrReExport", "__gojrExportAll", "__filename", "__dirname", transformed);
            fn(module.exports, module, embeddedRequire, reExport, exportAll, filename, dirname(filename));
            return module.exports;
        }
        return embeddedRequire;
    }
    function processEnvironment() {
        const processEnv = root.process?.env;
        return processEnv === undefined ? undefined : { GOJR_RANDOM_SEED: processEnv.GOJR_RANDOM_SEED };
    }
    function installEmbeddedRuntime(moduleBundleJson) {
        const embeddedRequire = createEmbeddedModuleLoader(moduleBundleJson);
        const gojrModule = embeddedRequire("/src/index.js");
        const runtimeOptions = (extra = {}) => gojrModule.runtimeOptionsFromEnvironment(processEnvironment(), extra);
        const gojrSession = new gojrModule.GoJuniorSession(runtimeOptions({ sheet: {} }));
        function evaluationJSON(result, extraOutput = []) {
            return gojrModule.evaluationResultToHostJSON(result, extraOutput);
        }
        root.__gojrEval = async function (source) {
            const result = await gojrSession.evaluate(source);
            return evaluationJSON(result);
        };
        root.__gojrEvalFiles = async function (json) {
            const result = await gojrModule.evaluateSourceFiles(JSON.parse(json), runtimeOptions());
            return evaluationJSON(result);
        };
        root.__gojrEvalWithPackages = async function (json) {
            const result = await gojrModule.evaluateSourceWithPackagesOnNode(JSON.parse(json));
            return evaluationJSON(result, result.packageOutput || []);
        };
        root.__gojrEvalFilesWithPackages = async function (json) {
            const result = await gojrModule.evaluateSourceFilesWithPackagesOnNode(JSON.parse(json));
            return evaluationJSON(result, result.packageOutput || []);
        };
        root.__gojrTestFilesWithPackages = async function (json) {
            const result = await gojrModule.testSourceFilesWithPackagesOnNode(JSON.parse(json));
            return evaluationJSON(result, result.packageOutput || []);
        };
        root.__gojrCompile = function (json) {
            const result = gojrModule.compileSourceFilesWithPackagesOnNode(JSON.parse(json));
            return gojrModule.compileResultToHostJSON(result);
        };
        root.__gojrTest = async function (source) {
            const result = await gojrModule.testSource(source, runtimeOptions());
            return evaluationJSON({ ...result, incomplete: false });
        };
        root.__gojrTestFiles = async function (json) {
            const result = await gojrModule.testSourceFiles(JSON.parse(json), runtimeOptions());
            return evaluationJSON({ ...result, incomplete: false });
        };
        root.__gojrSetSheet = function (json) {
            gojrSession.setSheet(gojrModule.parseSheetJson(json));
            return JSON.stringify({ ok: true, value: "sheet loaded" });
        };
        root.__gojrBuild = function (json) {
            const request = JSON.parse(json);
            const result = gojrModule.buildPackagesOnNode(request);
            return gojrModule.buildReportToHostJSON(result);
        };
        root.__gojrInspectJS = function (json) {
            const request = JSON.parse(json);
            const result = gojrModule.inspectPackageJavaScriptOnNode(request);
            return gojrModule.inspectPackageJavaScriptReportToHostJSON(result);
        };
        root.__gojrCache = function (json) {
            const result = gojrModule.packageCacheOnNode(JSON.parse(json));
            return JSON.stringify(result);
        };
        root.__gojrRunFixture = async function (json) {
            const result = await gojrModule.runSpreadsheetFixtureJson(json, runtimeOptions());
            return gojrModule.spreadsheetFixtureResultToHostJSON(result);
        };
        root.__gojrRunFixtureWithPackages = async function (json) {
            const result = await gojrModule.runSpreadsheetFixtureWithPackagesOnNode(JSON.parse(json));
            return gojrModule.spreadsheetFixtureResultToHostJSON(result);
        };
    }
    root.__gojrCreateEmbeddedModuleLoader = createEmbeddedModuleLoader;
    root.__gojrInstallEmbeddedRuntime = installEmbeddedRuntime;
})();
