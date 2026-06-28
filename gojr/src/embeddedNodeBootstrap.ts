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
        return `globalThis.Object.assign(exports, { ${names.map(([sourceName, exportName]) => `${JSON.stringify(exportName)}: ${sourceName}`).join(", ")} });`;
      });
    source = source.replace(/^export\s+default\s+(.+);\s*$/gm,
      (_match: string, expression: string) => `exports.default = ${expression};`);
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
      source += `\nglobalThis.Object.assign(exports, { ${[...new Set(exportedNames)].join(", ")} });\n`;
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
        const hostRequire = typeof root.require === "function"
          ? root.require as (specifier: string) => AnyRecord
          : typeof require === "function"
            ? require as (specifier: string) => AnyRecord
            : undefined;
        if (hostRequire === undefined) {
          throw new Error(`host module require is not available for ${specifier}`);
        }
        return hostRequire(specifier);
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

  function processEnvironment(): { GOJR_RANDOM_SEED?: string | undefined } | undefined {
    const processEnv = (root.process as { env?: Record<string, string | undefined> } | undefined)?.env;
    return processEnv === undefined ? undefined : { GOJR_RANDOM_SEED: processEnv.GOJR_RANDOM_SEED };
  }

  function installEmbeddedRuntime(moduleBundleJson: string): void {
    const embeddedRequire = createEmbeddedModuleLoader(moduleBundleJson);
    installNodeStdioHooks(embeddedRequire);
    const gojrModule = embeddedRequire("/src/index.js");
    const runtimeOptions = (extra: AnyRecord = {}) => gojrModule.runtimeOptionsFromEnvironment(processEnvironment(), extra);
    const gojrSession = new gojrModule.GoJuniorSession(runtimeOptions({ sheet: {} }));

    function evaluationJSON(result: any, extraOutput: string[] = []): string {
      return gojrModule.evaluationResultToHostJSON(result, extraOutput);
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

    root.__gojrRunMainFilesWithPackages = async function(json: string): Promise<string> {
      const request = JSON.parse(json);
      const result = await gojrModule.runMainSourceFilesWithPackagesOnNode(request);
      const streamed = request && request.streamOutput === true;
      return evaluationJSON(
        { ...result, output: streamed ? [] : result.output, incomplete: false },
        streamed ? [] : result.packageOutput || []
      );
    };

    root.__gojrTestFilesWithPackages = async function(json: string): Promise<string> {
      const result = await gojrModule.testSourceFilesWithPackagesOnNode(JSON.parse(json));
      return evaluationJSON(result, result.packageOutput || []);
    };

    root.__gojrCompile = function(json: string): string {
      const result = gojrModule.compileSourceFilesWithPackagesOnNode(JSON.parse(json));
      return gojrModule.compileResultToHostJSON(result);
    };

    root.__gojrTest = async function(source: string): Promise<string> {
      const result = await gojrModule.testSource(source, runtimeOptions());
      return evaluationJSON({ ...result, incomplete: false });
    };

    root.__gojrTestFiles = async function(json: string): Promise<string> {
      const result = await gojrModule.testSourceFiles(JSON.parse(json), runtimeOptions());
      return evaluationJSON({ ...result, incomplete: false });
    };

    root.__gojrSetSheet = function(json: string): string {
      gojrSession.setSheet(gojrModule.parseSheetJson(json));
      return JSON.stringify({ ok: true, value: "sheet loaded" });
    };

    root.__gojrBuild = function(json: string): string {
      const request = JSON.parse(json);
      const result = gojrModule.buildPackagesOnNode(request);
      return gojrModule.buildReportToHostJSON(result);
    };

    root.__gojrInspectJS = function(json: string): string {
      const request = JSON.parse(json);
      const result = gojrModule.inspectPackageJavaScriptOnNode(request);
      return gojrModule.inspectPackageJavaScriptReportToHostJSON(result);
    };

    root.__gojrCache = function(json: string): string {
      const result = gojrModule.packageCacheOnNode(JSON.parse(json));
      return JSON.stringify(result);
    };

    root.__gojrBench = async function(json: string): Promise<string> {
      const result = await gojrModule.benchmarkGoJuniorOnNode(JSON.parse(json));
      return gojrModule.benchmarkGoJuniorOnNodeHostJSON(result);
    };

    root.__gojrRunFixture = async function(json: string): Promise<string> {
      const result = await gojrModule.runSpreadsheetFixtureJson(json, runtimeOptions());
      return gojrModule.spreadsheetFixtureResultToHostJSON(result);
    };

    root.__gojrRunFixtureWithPackages = async function(json: string): Promise<string> {
      const result = await gojrModule.runSpreadsheetFixtureWithPackagesOnNode(JSON.parse(json));
      return gojrModule.spreadsheetFixtureResultToHostJSON(result);
    };
  }

  function installNodeStdioHooks(embeddedRequire: EmbeddedRequire): void {
    if (
      typeof root.__gojrReadSync === "function" &&
      typeof root.__gojrWriteSync === "function" &&
      typeof root.__gojrStatSync === "function"
    ) return;
    let fs: AnyRecord;
    try {
      fs = embeddedRequire("node:fs");
    } catch {
      try {
        fs = embeddedRequire("fs");
      } catch {
        return;
      }
    }
    if (typeof root.__gojrReadSync !== "function" && typeof fs.readSync === "function") {
      let stdinIsTTY = false;
      try {
        const tty = embeddedRequire("node:tty");
        stdinIsTTY = typeof tty.isatty === "function" && tty.isatty(0) === true;
      } catch {
        stdinIsTTY = false;
      }
      root.__gojrReadSync = (fd: number, buffer: Uint8Array, offset: number, length: number, position: number | null) => {
        const count = fs.readSync(fd, buffer, offset, length, position);
        if (fd === 0 && stdinIsTTY && count === 1 && buffer[offset] === 4) return 0;
        return count;
      };
    }
    if (typeof root.__gojrWriteSync !== "function" && typeof fs.writeSync === "function") {
      root.__gojrWriteSync = (fd: number, buffer: Uint8Array, offset: number, length: number, position: number | null) =>
        fs.writeSync(fd, buffer, offset, length, position);
    }
    if (typeof root.__gojrStatSync !== "function" && typeof fs.statSync === "function") {
      root.__gojrStatSync = (path: string) => fs.statSync(String(path ?? ""));
    }
  }

  root.__gojrCreateEmbeddedModuleLoader = createEmbeddedModuleLoader;
  root.__gojrInstallEmbeddedRuntime = installEmbeddedRuntime;
})();
