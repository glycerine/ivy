(() => {
  type AnyRecord = Record<string, any>;
  type ExportBinding = readonly [string, string];
  type EmbeddedRequire = (specifier: string, parent?: string) => AnyRecord;

  interface EmbeddedModule {
    readonly path: string;
    readonly source: string;
  }

  interface EmbeddedModuleBundle {
    readonly modules: readonly EmbeddedModule[];
  }

  interface EmbeddedModuleInstance {
    readonly exports: AnyRecord;
  }

  const root = globalThis as AnyRecord;

  function dirname(path: string): string {
    const index = path.lastIndexOf("/");
    return index <= 0 ? "/" : path.slice(0, index);
  }

  function normalizePath(path: string): string {
    const parts: string[] = [];
    for (const part of path.split("/")) {
      if (!part || part === ".") continue;
      if (part === "..") {
        parts.pop();
        continue;
      }
      parts.push(part);
    }
    return "/" + parts.join("/");
  }

  function resolveModule(specifier: string, parent: string): string {
    if (!specifier.startsWith(".") && !specifier.startsWith("/")) return specifier;
    const base = specifier.startsWith("/") ? specifier : `${dirname(parent)}/${specifier}`;
    return normalizePath(base);
  }

  function importListToDestructure(list: string): string {
    return list.split(",")
      .map((item) => item.trim())
      .filter(Boolean)
      .map((item) => {
        const match = /^([A-Za-z_$][\w$]*)(?:\s+as\s+([A-Za-z_$][\w$]*))?$/.exec(item);
        if (!match) throw new Error(`unsupported embedded import specifier: ${item}`);
        return match[2] ? `${match[1]}: ${match[2]}` : match[1];
      })
      .join(", ");
  }

  function exportList(list: string): ExportBinding[] {
    return list.split(",")
      .map((item) => item.trim())
      .filter(Boolean)
      .map((item) => {
        const match = /^([A-Za-z_$][\w$]*)(?:\s+as\s+([A-Za-z_$][\w$]*))?$/.exec(item);
        if (!match) throw new Error(`unsupported embedded export specifier: ${item}`);
        return [match[1]!, match[2] || match[1]!];
      });
  }

  function transformModule(source: string, filename: string): string {
    const exportedNames: string[] = [];
    const earlyFunctionExports: string[] = [];
    source.replace(/^export\s+(async\s+)?function(\*)?\s+([A-Za-z_$][\w$]*)/gm,
      (_match: string, _asyncPrefix: string | undefined, _generatorMarker: string | undefined, name: string) => {
        earlyFunctionExports.push(name);
        return _match;
      });
    source = source.replace(/^export\s+\{\s*\};\s*$/gm, "");
    source = source.replace(/^import\s+\{([^}]+)\}\s+from\s+["']([^"']+)["'];\s*$/gm,
      (_match: string, imports: string, specifier: string) => `const { ${importListToDestructure(imports)} } = __gojrRequire(${JSON.stringify(specifier)}, __filename);`);
    source = source.replace(/^import\s+\*\s+as\s+([A-Za-z_$][\w$]*)\s+from\s+["']([^"']+)["'];\s*$/gm,
      (_match: string, name: string, specifier: string) => `const ${name} = __gojrRequire(${JSON.stringify(specifier)}, __filename);`);
    source = source.replace(/^import\s+["']([^"']+)["'];\s*$/gm,
      (_match: string, specifier: string) => `__gojrRequire(${JSON.stringify(specifier)}, __filename);`);
    source = source.replace(/^export\s+\{([^}]+)\}\s+from\s+["']([^"']+)["'];\s*$/gm,
      (_match: string, exportsList: string, specifier: string) => `__gojrReExport(${JSON.stringify(specifier)}, ${JSON.stringify(exportList(exportsList))}, exports, __filename);`);
    source = source.replace(/^export\s+\*\s+as\s+([A-Za-z_$][\w$]*)\s+from\s+["']([^"']+)["'];\s*$/gm,
      (_match: string, name: string, specifier: string) => `exports[${JSON.stringify(name)}] = __gojrRequire(${JSON.stringify(specifier)}, __filename);`);
    source = source.replace(/^export\s+\*\s+from\s+["']([^"']+)["'];\s*$/gm,
      (_match: string, specifier: string) => `__gojrExportAll(${JSON.stringify(specifier)}, exports, __filename);`);
    source = source.replace(/^export\s+\{([^}]*)\};\s*$/gm,
      (_match: string, exportsList: string) => {
        const names = exportList(exportsList);
        if (names.length === 0) return "";
        return `Object.assign(exports, { ${names.map(([sourceName, exportName]) => `${JSON.stringify(exportName)}: ${sourceName}`).join(", ")} });`;
      });
    source = source.replace(/^export\s+(async\s+)?function(\*)?\s+([A-Za-z_$][\w$]*)/gm,
      (_match: string, asyncPrefix: string | undefined, generatorMarker: string | undefined, name: string) => {
        exportedNames.push(name);
        return `exports[${JSON.stringify(name)}] = ${name};\n${asyncPrefix || ""}function${generatorMarker || ""} ${name}`;
      });
    source = source.replace(/^export\s+(function|class)\s+([A-Za-z_$][\w$]*)/gm,
      (_match: string, kind: string, name: string) => {
        exportedNames.push(name);
        return `${kind} ${name}`;
      });
    source = source.replace(/^export\s+(const|let|var)\s+([A-Za-z_$][\w$]*)/gm,
      (_match: string, kind: string, name: string) => {
        exportedNames.push(name);
        return `${kind} ${name}`;
      });
    if (exportedNames.length > 0) {
      source += `\nObject.assign(exports, { ${[...new Set(exportedNames)].join(", ")} });\n`;
    }
    if (earlyFunctionExports.length > 0) {
      source = `${[...new Set(earlyFunctionExports)].map((name) => `exports[${JSON.stringify(name)}] = ${name};`).join("\n")}\n${source}`;
    }
    return source + `\n//# sourceURL=embedded-gojr:${filename}\n`;
  }

  function createEmbeddedModuleLoader(moduleBundleJson: string): EmbeddedRequire {
    const parsed = JSON.parse(moduleBundleJson || "{\"modules\":[]}") as EmbeddedModuleBundle;
    const embeddedModules = new Map(parsed.modules.map((module) => [module.path, module.source]));
    const embeddedCache = new Map<string, EmbeddedModuleInstance>();

    function reExport(specifier: string, names: readonly ExportBinding[], target: AnyRecord, parent: string): void {
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

    function exportAll(specifier: string, target: AnyRecord, parent: string): void {
      const module = embeddedRequire(specifier, parent);
      for (const key of Object.keys(module)) {
        if (key === "default" || Object.prototype.hasOwnProperty.call(target, key)) continue;
        Object.defineProperty(target, key, {
          enumerable: true,
          get() {
            return module[key];
          }
        });
      }
    }

    function embeddedRequire(specifier: string, parent = "/src/index.js"): AnyRecord {
      if (!specifier.startsWith(".") && !specifier.startsWith("/")) {
        if (typeof root.require !== "function") {
          throw new Error(`host module require is not available for ${specifier}`);
        }
        return root.require(specifier) as AnyRecord;
      }
      const filename = resolveModule(specifier, parent);
      const cached = embeddedCache.get(filename);
      if (cached) return cached.exports;
      const source = embeddedModules.get(filename);
      if (source === undefined) throw new Error(`embedded Go-junior module not found: ${filename}`);
      const module: EmbeddedModuleInstance = { exports: {} };
      embeddedCache.set(filename, module);
      const transformed = transformModule(source, filename);
      const fn = new Function("exports", "module", "__gojrRequire", "__gojrReExport", "__gojrExportAll", "__filename", "__dirname", transformed);
      fn(module.exports, module, embeddedRequire, reExport, exportAll, filename, dirname(filename));
      return module.exports;
    }

    return embeddedRequire;
  }

  function runtimeOptions(extra: AnyRecord = {}): AnyRecord {
    const processEnv = (root.process as { env?: Record<string, string | undefined> } | undefined)?.env;
    const seed = processEnv?.GOJR_RANDOM_SEED;
    return seed === undefined || seed === ""
      ? { ...extra }
      : { ...extra, randomSeed: seed };
  }

  function diagnosticString(diagnostic: any): string {
    const filename = diagnostic?.span?.filename || diagnostic?.filename || "gojr-repl.go";
    const location = diagnostic && diagnostic.span
      ? `${filename}:${diagnostic.span.line}:${diagnostic.span.column}: `
      : `${filename}: `;
    const stack = typeof diagnostic?.stack === "string" && diagnostic.stack !== ""
      ? `\n${diagnostic.stack}`
      : "";
    return `${location}${diagnostic.severity} ${diagnostic.code}: ${diagnostic.message}${stack}`;
  }

  function spreadsheetDiagnosticString(diagnostic: any): string {
    const cell = diagnostic && diagnostic.cell
      ? `${diagnostic.cell.sheet}!${diagnostic.cell.cell}`
      : "spreadsheet";
    return `${cell}: error ${diagnostic.code}: ${diagnostic.message}`;
  }

  function installEmbeddedRuntime(moduleBundleJson: string): void {
    const embeddedRequire = createEmbeddedModuleLoader(moduleBundleJson);
    const gojrModule = embeddedRequire("/src/index.js");
    const gojrSession = new gojrModule.GoJuniorSession(runtimeOptions({ sheet: {} }));

    function formatValue(value: any): string {
      if (typeof value === "function") {
        const name = value.name ? value.name.replace(/^_fn_/, "") : "";
        return name ? `<func ${name}>` : "<func>";
      }
      return gojrModule.formatReplValue(value);
    }

    function formatResult(result: any): string {
      if (result.values) return result.values.map((value: any) => formatValue(value)).join(", ");
      if (Object.prototype.hasOwnProperty.call(result, "value")) return formatValue(result.value);
      return "";
    }

    function resultValueIsNil(result: any): boolean {
      if (Array.isArray(result.values)) return result.values.length === 1 && result.values[0] === null;
      if (Object.prototype.hasOwnProperty.call(result, "value")) return result.value === null;
      return false;
    }

    function observedDeps(result: any): any[] {
      return Array.isArray(result.observedDeps) ? result.observedDeps : [];
    }

    function evaluationJSON(result: any, extraOutput: string[] = []): string {
      const diagnostics = result.diagnostics || [];
      return JSON.stringify({
        ok: !diagnostics.some((diagnostic: any) => diagnostic.severity === "error"),
        incomplete: result.incomplete === true,
        diagnostics: diagnostics.map(diagnosticString),
        output: [...extraOutput, ...(result.output || [])].join(""),
        value: formatResult(result),
        valueIsNil: resultValueIsNil(result),
        observedDeps: observedDeps(result)
      });
    }

    function formatFixtureSheets(sheets: any): AnyRecord {
      const formatted: AnyRecord = {};
      for (const [sheetName, cells] of Object.entries((sheets || {}) as AnyRecord)) {
        formatted[sheetName] = {};
        for (const [cell, value] of Object.entries((cells || {}) as AnyRecord)) {
          formatted[sheetName][cell] = formatValue(value);
        }
      }
      return formatted;
    }

    root.__gojrEval = async function(source: string): Promise<string> {
      const result = await gojrSession.evaluate(source);
      return evaluationJSON(result);
    };

    root.__gojrEvalFiles = async function(json: string): Promise<string> {
      const result = await gojrModule.evaluateSourceFiles(JSON.parse(json), runtimeOptions());
      return evaluationJSON(result);
    };

    root.__gojrEvalWithPackages = async function(json: string): Promise<string> {
      const result = await gojrModule.evaluateSourceWithPackagesOnNode(JSON.parse(json));
      return evaluationJSON(result, result.packageOutput || []);
    };

    root.__gojrEvalFilesWithPackages = async function(json: string): Promise<string> {
      const result = await gojrModule.evaluateSourceFilesWithPackagesOnNode(JSON.parse(json));
      return evaluationJSON(result, result.packageOutput || []);
    };

    root.__gojrTestFilesWithPackages = async function(json: string): Promise<string> {
      const result = await gojrModule.testSourceFilesWithPackagesOnNode(JSON.parse(json));
      return evaluationJSON(result, result.packageOutput || []);
    };

    root.__gojrCompile = function(json: string): string {
      const result = gojrModule.compileSourceFilesWithPackagesOnNode(JSON.parse(json));
      const diagnostics = result.diagnostics || [];
      return JSON.stringify({
        ok: !diagnostics.some((diagnostic: any) => diagnostic.severity === "error"),
        diagnostics: diagnostics.map(diagnosticString),
        output: ""
      });
    };

    root.__gojrTest = async function(source: string): Promise<string> {
      const result = await gojrModule.testSource(source, runtimeOptions());
      const diagnostics = result.diagnostics || [];
      return JSON.stringify({
        ok: !diagnostics.some((diagnostic: any) => diagnostic.severity === "error"),
        incomplete: false,
        diagnostics: diagnostics.map(diagnosticString),
        output: (result.output || []).join(""),
        value: formatResult(result),
        valueIsNil: resultValueIsNil(result),
        observedDeps: observedDeps(result)
      });
    };

    root.__gojrTestFiles = async function(json: string): Promise<string> {
      const result = await gojrModule.testSourceFiles(JSON.parse(json), runtimeOptions());
      const diagnostics = result.diagnostics || [];
      return JSON.stringify({
        ok: !diagnostics.some((diagnostic: any) => diagnostic.severity === "error"),
        incomplete: false,
        diagnostics: diagnostics.map(diagnosticString),
        output: (result.output || []).join(""),
        value: formatResult(result),
        valueIsNil: resultValueIsNil(result),
        observedDeps: observedDeps(result)
      });
    };

    root.__gojrSetSheet = function(json: string): string {
      gojrSession.setSheet(gojrModule.parseSheetJson(json));
      return JSON.stringify({ ok: true, value: "sheet loaded" });
    };

    root.__gojrBuild = function(json: string): string {
      const request = JSON.parse(json);
      const result = gojrModule.buildPackagesOnNode(request);
      const diagnostics = result.diagnostics || [];
      return JSON.stringify({
        ok: !diagnostics.some((diagnostic: any) => diagnostic.severity === "error"),
        incomplete: false,
        diagnostics: diagnostics.map(diagnosticString),
        output: "",
        artifacts: result.artifacts || [],
        built: result.built || [],
        skipped: result.skipped || []
      });
    };

    root.__gojrInspectJS = function(json: string): string {
      const request = JSON.parse(json);
      const result = gojrModule.inspectPackageJavaScriptOnNode(request);
      const diagnostics = result.diagnostics || [];
      return JSON.stringify({
        ok: !diagnostics.some((diagnostic: any) => diagnostic.severity === "error"),
        incomplete: false,
        diagnostics: diagnostics.map(diagnosticString),
        output: "",
        artifacts: result.artifacts || [],
        built: result.built || [],
        skipped: result.skipped || [],
        source: result.source || ""
      });
    };

    root.__gojrCache = function(json: string): string {
      const result = gojrModule.packageCacheOnNode(JSON.parse(json));
      return JSON.stringify(result);
    };

    root.__gojrRunFixture = async function(json: string): Promise<string> {
      const result = await gojrModule.runSpreadsheetFixtureJson(json, runtimeOptions());
      const diagnostics = result.diagnostics || [];
      return JSON.stringify({
        ok: result.ok === true && !diagnostics.length,
        unstable: result.unstable === true,
        diagnostics: diagnostics.map(spreadsheetDiagnosticString),
        evaluated: result.evaluated || [],
        sheets: formatFixtureSheets(result.sheets || {}),
        observedDeps: result.observedDeps || {}
      });
    };

    root.__gojrRunFixtureWithPackages = async function(json: string): Promise<string> {
      const result = await gojrModule.runSpreadsheetFixtureWithPackagesOnNode(JSON.parse(json));
      const packageDiagnostics = result.packageDiagnostics || [];
      if (packageDiagnostics.some((diagnostic: any) => diagnostic.severity === "error")) {
        return JSON.stringify({
          ok: false,
          unstable: false,
          diagnostics: packageDiagnostics.map(diagnosticString),
          evaluated: [],
          sheets: {},
          observedDeps: {}
        });
      }
      const diagnostics = result.diagnostics || [];
      return JSON.stringify({
        ok: result.ok === true && !diagnostics.length,
        unstable: result.unstable === true,
        diagnostics: diagnostics.map(spreadsheetDiagnosticString),
        evaluated: result.evaluated || [],
        sheets: formatFixtureSheets(result.sheets || {}),
        observedDeps: result.observedDeps || {}
      });
    };
  }

  root.__gojrCreateEmbeddedModuleLoader = createEmbeddedModuleLoader;
  root.__gojrInstallEmbeddedRuntime = installEmbeddedRuntime;
})();
