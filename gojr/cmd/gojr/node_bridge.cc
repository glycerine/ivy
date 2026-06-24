#include "node_bridge.h"

#include <node.h>
#include <v8.h>

#include <cstdlib>
#include <cstdio>
#include <cstring>
#include <memory>
#include <sstream>
#include <string>
#include <vector>

struct gojr_node_runtime {
  std::shared_ptr<node::InitializationResult> initialization;
  std::unique_ptr<node::CommonEnvironmentSetup> setup;
  node::MultiIsolatePlatform* platform = nullptr;
};

namespace {

char* copy_string(const std::string& text) {
  char* out = static_cast<char*>(std::malloc(text.size() + 1));
  if (out == nullptr) return nullptr;
  std::memcpy(out, text.c_str(), text.size() + 1);
  return out;
}

void set_error(char** error_out, const std::string& message) {
  if (error_out != nullptr) *error_out = copy_string(message);
}

std::string v8_to_string(v8::Isolate* isolate, v8::Local<v8::Value> value) {
  v8::String::Utf8Value utf8(isolate, value);
  if (*utf8 == nullptr) return "";
  return std::string(*utf8, utf8.length());
}

std::string exception_to_string(v8::Isolate* isolate, v8::TryCatch& try_catch) {
  std::ostringstream out;
  if (!try_catch.Exception().IsEmpty()) {
    out << v8_to_string(isolate, try_catch.Exception());
  } else {
    out << "unknown JavaScript exception";
  }

  v8::Local<v8::Message> message = try_catch.Message();
  if (!message.IsEmpty()) {
    v8::Local<v8::Context> context = isolate->GetCurrentContext();
    v8::MaybeLocal<v8::String> maybe_resource = message->GetScriptOrigin().ResourceName()->ToString(context);
    v8::Local<v8::String> resource;
    if (maybe_resource.ToLocal(&resource)) {
      out << " at " << v8_to_string(isolate, resource);
    }
    int line = message->GetLineNumber(context).FromMaybe(0);
    int column = message->GetStartColumn(context).FromMaybe(0) + 1;
    if (line > 0) out << ":" << line << ":" << column;
  }
  return out.str();
}

std::string value_exception_to_string(v8::Isolate* isolate, v8::Local<v8::Value> value) {
  if (value->IsNativeError()) return v8_to_string(isolate, value);
  if (value->IsObject()) {
    v8::Local<v8::Context> context = isolate->GetCurrentContext();
    v8::Local<v8::String> stack_name = v8::String::NewFromUtf8Literal(isolate, "stack");
    v8::Local<v8::Value> stack;
    if (value.As<v8::Object>()->Get(context, stack_name).ToLocal(&stack) && !stack->IsUndefined()) {
      return v8_to_string(isolate, stack);
    }
  }
  return v8_to_string(isolate, value);
}

std::string js_string_literal(const std::string& text) {
  std::ostringstream out;
  out << '"';
  for (unsigned char ch : text) {
    switch (ch) {
      case '\\': out << "\\\\"; break;
      case '"': out << "\\\""; break;
      case '\b': out << "\\b"; break;
      case '\f': out << "\\f"; break;
      case '\n': out << "\\n"; break;
      case '\r': out << "\\r"; break;
      case '\t': out << "\\t"; break;
      default:
        if (ch < 0x20) {
          const char* hex = "0123456789abcdef";
          out << "\\u00" << hex[(ch >> 4) & 0xf] << hex[ch & 0xf];
        } else {
          out << static_cast<char>(ch);
        }
    }
  }
  out << '"';
  return out.str();
}

std::string bootstrap_source(const std::string& module_bundle_json) {
  return R"JS(
const __gojrEmbeddedModules = new Map(
  JSON.parse()JS" + js_string_literal(module_bundle_json) + R"JS().modules.map((module) => [module.path, module.source])
);
const __gojrEmbeddedCache = new Map();

function __gojrDirname(path) {
  const index = path.lastIndexOf("/");
  return index <= 0 ? "/" : path.slice(0, index);
}

function __gojrNormalizePath(path) {
  const parts = [];
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

function __gojrResolveModule(specifier, parent) {
  if (!specifier.startsWith(".") && !specifier.startsWith("/")) return specifier;
  const base = specifier.startsWith("/") ? specifier : `${__gojrDirname(parent)}/${specifier}`;
  return __gojrNormalizePath(base);
}

function __gojrImportListToDestructure(list) {
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

function __gojrExportList(list) {
  return list.split(",")
    .map((item) => item.trim())
    .filter(Boolean)
    .map((item) => {
      const match = /^([A-Za-z_$][\w$]*)(?:\s+as\s+([A-Za-z_$][\w$]*))?$/.exec(item);
      if (!match) throw new Error(`unsupported embedded export specifier: ${item}`);
      return [match[1], match[2] || match[1]];
    });
}

function __gojrTransformModule(source, filename) {
  const exportedNames = [];
  const earlyFunctionExports = [];
  source.replace(/^export\s+(async\s+)?function(\*)?\s+([A-Za-z_$][\w$]*)/gm,
    (_match, _asyncPrefix, _generatorMarker, name) => {
      earlyFunctionExports.push(name);
      return _match;
    });
  source = source.replace(/^export\s+\{\s*\};\s*$/gm, "");
  source = source.replace(/^import\s+\{([^}]+)\}\s+from\s+["']([^"']+)["'];\s*$/gm,
    (_match, imports, specifier) => `const { ${__gojrImportListToDestructure(imports)} } = __gojrRequire(${JSON.stringify(specifier)}, __filename);`);
  source = source.replace(/^import\s+\*\s+as\s+([A-Za-z_$][\w$]*)\s+from\s+["']([^"']+)["'];\s*$/gm,
    (_match, name, specifier) => `const ${name} = __gojrRequire(${JSON.stringify(specifier)}, __filename);`);
  source = source.replace(/^import\s+["']([^"']+)["'];\s*$/gm,
    (_match, specifier) => `__gojrRequire(${JSON.stringify(specifier)}, __filename);`);
  source = source.replace(/^export\s+\{([^}]+)\}\s+from\s+["']([^"']+)["'];\s*$/gm,
    (_match, exportsList, specifier) => `__gojrReExport(${JSON.stringify(specifier)}, ${JSON.stringify(__gojrExportList(exportsList))}, exports, __filename);`);
  source = source.replace(/^export\s+\*\s+as\s+([A-Za-z_$][\w$]*)\s+from\s+["']([^"']+)["'];\s*$/gm,
    (_match, name, specifier) => `exports[${JSON.stringify(name)}] = __gojrRequire(${JSON.stringify(specifier)}, __filename);`);
  source = source.replace(/^export\s+\*\s+from\s+["']([^"']+)["'];\s*$/gm,
    (_match, specifier) => `__gojrExportAll(${JSON.stringify(specifier)}, exports, __filename);`);
  source = source.replace(/^export\s+\{([^}]*)\};\s*$/gm,
    (_match, exportsList) => {
      const names = __gojrExportList(exportsList);
      if (names.length === 0) return "";
      return `globalThis.Object.assign(exports, { ${names.map(([sourceName, exportName]) => `${JSON.stringify(exportName)}: ${sourceName}`).join(", ")} });`;
    });
  source = source.replace(/^export\s+(async\s+)?function(\*)?\s+([A-Za-z_$][\w$]*)/gm,
    (_match, asyncPrefix, generatorMarker, name) => {
      exportedNames.push(name);
      return `exports[${JSON.stringify(name)}] = ${name};\n${asyncPrefix || ""}function${generatorMarker || ""} ${name}`;
    });
  source = source.replace(/^export\s+(function|class)\s+([A-Za-z_$][\w$]*)/gm,
    (_match, kind, name) => {
      exportedNames.push(name);
      return `${kind} ${name}`;
    });
  source = source.replace(/^export\s+(const|let|var)\s+([A-Za-z_$][\w$]*)/gm,
    (_match, kind, name) => {
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

function __gojrReExport(specifier, names, target, parent) {
  const module = __gojrRequire(specifier, parent);
  for (const [sourceName, exportName] of names) {
    globalThis.Object.defineProperty(target, exportName, {
      enumerable: true,
      get() { return module[sourceName]; }
    });
  }
}

function __gojrExportAll(specifier, target, parent) {
  const module = __gojrRequire(specifier, parent);
  for (const key of globalThis.Object.keys(module)) {
    if (key === "default" || globalThis.Object.prototype.hasOwnProperty.call(target, key)) continue;
    globalThis.Object.defineProperty(target, key, {
      enumerable: true,
      get() { return module[key]; }
    });
  }
}

function __gojrRequire(specifier, parent = "/src/index.js") {
  if (!specifier.startsWith(".") && !specifier.startsWith("/")) return require(specifier);
  const filename = __gojrResolveModule(specifier, parent);
  const cached = __gojrEmbeddedCache.get(filename);
  if (cached) return cached.exports;
  const source = __gojrEmbeddedModules.get(filename);
  if (source === undefined) throw new Error(`embedded Go-junior module not found: ${filename}`);
  const module = { exports: {} };
  __gojrEmbeddedCache.set(filename, module);
  const transformed = __gojrTransformModule(source, filename);
  const fn = new Function("exports", "module", "__gojrRequire", "__gojrReExport", "__gojrExportAll", "__filename", "__dirname", transformed);
  fn(module.exports, module, __gojrRequire, __gojrReExport, __gojrExportAll, filename, __gojrDirname(filename));
  return module.exports;
}

const gojrModule = __gojrRequire("/src/index.js");

function gojrRuntimeOptions(extra = {}) {
  const seed = globalThis.process?.env?.GOJR_RANDOM_SEED;
  return seed === undefined || seed === ""
    ? { ...extra }
    : { ...extra, randomSeed: seed };
}

const gojrSession = new gojrModule.GoJuniorSession(gojrRuntimeOptions({ sheet: {} }));

function gojrDiagnosticString(diagnostic) {
  const filename = diagnostic?.span?.filename || diagnostic?.filename || "gojr-repl.go";
  const location = diagnostic && diagnostic.span
    ? `${filename}:${diagnostic.span.line}:${diagnostic.span.column}: `
    : `${filename}: `;
  const stack = typeof diagnostic?.stack === "string" && diagnostic.stack !== ""
    ? `\n${diagnostic.stack}`
    : "";
  return `${location}${diagnostic.severity} ${diagnostic.code}: ${diagnostic.message}${stack}`;
}

function gojrFormatResult(result) {
  if (result.values) return result.values.map((value) => gojrFormatValue(value)).join(", ");
  if (Object.prototype.hasOwnProperty.call(result, "value")) return gojrFormatValue(result.value);
  return "";
}

function gojrFormatValue(value) {
  if (typeof value === "function") {
    const name = value.name ? value.name.replace(/^_fn_/, "") : "";
    return name ? `<func ${name}>` : "<func>";
  }
  return gojrModule.formatReplValue(value);
}

function gojrResultValueIsNil(result) {
  if (Array.isArray(result.values)) return result.values.length === 1 && result.values[0] === null;
  if (Object.prototype.hasOwnProperty.call(result, "value")) return result.value === null;
  return false;
}

function gojrObservedDeps(result) {
  return Array.isArray(result.observedDeps) ? result.observedDeps : [];
}

function gojrEvaluationJSON(result, extraOutput = []) {
  const diagnostics = result.diagnostics || [];
  return JSON.stringify({
    ok: !diagnostics.some((diagnostic) => diagnostic.severity === "error"),
    incomplete: result.incomplete === true,
    diagnostics: diagnostics.map(gojrDiagnosticString),
    output: [...extraOutput, ...(result.output || [])].join(""),
    value: gojrFormatResult(result),
    valueIsNil: gojrResultValueIsNil(result),
    observedDeps: gojrObservedDeps(result)
  });
}

function gojrSpreadsheetDiagnosticString(diagnostic) {
  const cell = diagnostic && diagnostic.cell
    ? `${diagnostic.cell.sheet}!${diagnostic.cell.cell}`
    : "spreadsheet";
  return `${cell}: error ${diagnostic.code}: ${diagnostic.message}`;
}

function gojrFormatFixtureSheets(sheets) {
  const formatted = {};
  for (const [sheetName, cells] of Object.entries(sheets || {})) {
    formatted[sheetName] = {};
    for (const [cell, value] of Object.entries(cells || {})) {
      formatted[sheetName][cell] = gojrFormatValue(value);
    }
  }
  return formatted;
}

globalThis.__gojrEval = async function(source) {
  const result = await gojrSession.evaluate(source);
  return gojrEvaluationJSON(result);
};

globalThis.__gojrEvalFiles = async function(json) {
  const result = await gojrModule.evaluateSourceFiles(JSON.parse(json), gojrRuntimeOptions());
  return gojrEvaluationJSON(result);
};

globalThis.__gojrEvalWithPackages = async function(json) {
  const result = await gojrModule.evaluateSourceWithPackagesOnNode(JSON.parse(json));
  return gojrEvaluationJSON(result, result.packageOutput || []);
};

globalThis.__gojrEvalFilesWithPackages = async function(json) {
  const result = await gojrModule.evaluateSourceFilesWithPackagesOnNode(JSON.parse(json));
  return gojrEvaluationJSON(result, result.packageOutput || []);
};

globalThis.__gojrTestFilesWithPackages = async function(json) {
  const result = await gojrModule.testSourceFilesWithPackagesOnNode(JSON.parse(json));
  return gojrEvaluationJSON(result, result.packageOutput || []);
};

globalThis.__gojrCompile = function(json) {
  const result = gojrModule.compileSourceFilesWithPackagesOnNode(JSON.parse(json));
  const diagnostics = result.diagnostics || [];
  return JSON.stringify({
    ok: !diagnostics.some((diagnostic) => diagnostic.severity === "error"),
    diagnostics: diagnostics.map(gojrDiagnosticString),
    output: ""
  });
};

globalThis.__gojrTest = async function(source) {
  const result = await gojrModule.testSource(source, gojrRuntimeOptions());
  const diagnostics = result.diagnostics || [];
  return JSON.stringify({
    ok: !diagnostics.some((diagnostic) => diagnostic.severity === "error"),
    incomplete: false,
    diagnostics: diagnostics.map(gojrDiagnosticString),
    output: (result.output || []).join(""),
    value: gojrFormatResult(result),
    valueIsNil: gojrResultValueIsNil(result),
    observedDeps: gojrObservedDeps(result)
  });
};

globalThis.__gojrTestFiles = async function(json) {
  const result = await gojrModule.testSourceFiles(JSON.parse(json), gojrRuntimeOptions());
  const diagnostics = result.diagnostics || [];
  return JSON.stringify({
    ok: !diagnostics.some((diagnostic) => diagnostic.severity === "error"),
    incomplete: false,
    diagnostics: diagnostics.map(gojrDiagnosticString),
    output: (result.output || []).join(""),
    value: gojrFormatResult(result),
    valueIsNil: gojrResultValueIsNil(result),
    observedDeps: gojrObservedDeps(result)
  });
};

globalThis.__gojrSetSheet = function(json) {
  gojrSession.setSheet(gojrModule.parseSheetJson(json));
  return JSON.stringify({ ok: true, value: "sheet loaded" });
};

function gojrDefaultPackageCacheParent() {
  return gojrModule.defaultPackageCacheParent();
}

globalThis.__gojrBuild = function(json) {
  const request = JSON.parse(json);
  const result = gojrModule.buildPackagesOnNode(request);
  const diagnostics = result.diagnostics || [];
  return JSON.stringify({
    ok: !diagnostics.some((diagnostic) => diagnostic.severity === "error"),
    incomplete: false,
    diagnostics: diagnostics.map(gojrDiagnosticString),
    output: "",
    artifacts: result.artifacts || [],
    built: result.built || [],
    skipped: result.skipped || []
  });
};

globalThis.__gojrInspectJS = function(json) {
  const request = JSON.parse(json);
  const result = gojrModule.inspectPackageJavaScriptOnNode(request);
  const diagnostics = result.diagnostics || [];
  return JSON.stringify({
    ok: !diagnostics.some((diagnostic) => diagnostic.severity === "error"),
    incomplete: false,
    diagnostics: diagnostics.map(gojrDiagnosticString),
    output: "",
    artifacts: result.artifacts || [],
    built: result.built || [],
    skipped: result.skipped || [],
    source: result.source || ""
  });
};

globalThis.__gojrCache = function(json) {
  const result = gojrModule.packageCacheOnNode(JSON.parse(json));
  return JSON.stringify(result);
};

globalThis.__gojrRunFixture = async function(json) {
  const result = await gojrModule.runSpreadsheetFixtureJson(json, gojrRuntimeOptions());
  const diagnostics = result.diagnostics || [];
  return JSON.stringify({
    ok: result.ok === true && !diagnostics.length,
    unstable: result.unstable === true,
    diagnostics: diagnostics.map(gojrSpreadsheetDiagnosticString),
    evaluated: result.evaluated || [],
    sheets: gojrFormatFixtureSheets(result.sheets || {}),
    observedDeps: result.observedDeps || {}
  });
};

globalThis.__gojrRunFixtureWithPackages = async function(json) {
  const result = await gojrModule.runSpreadsheetFixtureWithPackagesOnNode(JSON.parse(json));
  const packageDiagnostics = result.packageDiagnostics || [];
  if (packageDiagnostics.some((diagnostic) => diagnostic.severity === "error")) {
    return JSON.stringify({
      ok: false,
      unstable: false,
      diagnostics: packageDiagnostics.map(gojrDiagnosticString),
      evaluated: [],
      sheets: {},
      observedDeps: {}
    });
  }
  const diagnostics = result.diagnostics || [];
  return JSON.stringify({
    ok: result.ok === true && !diagnostics.length,
    unstable: result.unstable === true,
    diagnostics: diagnostics.map(gojrSpreadsheetDiagnosticString),
    evaluated: result.evaluated || [],
    sheets: gojrFormatFixtureSheets(result.sheets || {}),
    observedDeps: result.observedDeps || {}
  });
};
)JS";
}

bool run_bootstrap(gojr_node_runtime* runtime, const std::string& module_bundle_json, std::string* error) {
  v8::Isolate* isolate = runtime->setup->isolate();
  v8::Locker locker(isolate);
  v8::Isolate::Scope isolate_scope(isolate);
  v8::HandleScope handle_scope(isolate);
  v8::Local<v8::Context> context = runtime->setup->context();
  v8::Context::Scope context_scope(context);
  v8::TryCatch try_catch(isolate);

  std::string source = bootstrap_source(module_bundle_json);
  if (std::getenv("GOJR_DEBUG_BOOTSTRAP") != nullptr) {
    std::fprintf(stderr, "%s\n", source.c_str());
  }
  v8::MaybeLocal<v8::Value> loaded = node::LoadEnvironment(runtime->setup->env(), source);
  if (loaded.IsEmpty()) {
    *error = exception_to_string(isolate, try_catch);
    return false;
  }
  return true;
}

char* call_global_string_function(
    gojr_node_runtime* runtime,
    const char* function_name,
    const char* argument,
    char** error_out) {
  if (runtime == nullptr) {
    set_error(error_out, "Node runtime is not initialized");
    return nullptr;
  }

  v8::Isolate* isolate = runtime->setup->isolate();
  v8::Locker locker(isolate);
  v8::Isolate::Scope isolate_scope(isolate);
  v8::HandleScope handle_scope(isolate);
  v8::Local<v8::Context> context = runtime->setup->context();
  v8::Context::Scope context_scope(context);
  v8::TryCatch try_catch(isolate);

  v8::Local<v8::String> name = v8::String::NewFromUtf8(isolate, function_name).ToLocalChecked();
  v8::Local<v8::Value> value;
  if (!context->Global()->Get(context, name).ToLocal(&value) || !value->IsFunction()) {
    set_error(error_out, std::string(function_name) + " is not installed in the embedded Node runtime");
    return nullptr;
  }

  v8::Local<v8::Function> function = value.As<v8::Function>();
  v8::Local<v8::Value> argv[1] = {
      v8::String::NewFromUtf8(isolate, argument == nullptr ? "" : argument).ToLocalChecked()};
  v8::Local<v8::Value> result;
  if (!function->Call(context, context->Global(), 1, argv).ToLocal(&result)) {
    set_error(error_out, exception_to_string(isolate, try_catch));
    return nullptr;
  }

  if (result->IsPromise()) {
    v8::Local<v8::Promise> promise = result.As<v8::Promise>();
    while (promise->State() == v8::Promise::kPending) {
      isolate->PerformMicrotaskCheckpoint();
      if (promise->State() != v8::Promise::kPending) break;
      v8::Maybe<int> spun = node::SpinEventLoop(runtime->setup->env());
      if (spun.IsNothing()) {
        set_error(error_out, exception_to_string(isolate, try_catch));
        return nullptr;
      }
      isolate->PerformMicrotaskCheckpoint();
      if (promise->State() == v8::Promise::kPending) {
        set_error(error_out, std::string(function_name) + " returned a Promise that did not settle");
        return nullptr;
      }
    }
    if (promise->State() == v8::Promise::kRejected) {
      set_error(error_out, value_exception_to_string(isolate, promise->Result()));
      return nullptr;
    }
    result = promise->Result();
  }

  v8::Local<v8::String> result_string;
  if (!result->ToString(context).ToLocal(&result_string)) {
    set_error(error_out, exception_to_string(isolate, try_catch));
    return nullptr;
  }
  return copy_string(v8_to_string(isolate, result_string));
}

}  // namespace

extern "C" gojr_node_runtime* gojr_node_new(const char* module_bundle_json, char** error_out) {
  auto* runtime = new gojr_node_runtime();

  std::vector<std::string> args = {"gojr-embedded-node",
                                   "--disable-warning=ExperimentalWarning"};
  runtime->initialization = node::InitializeOncePerProcess(
      args,
      {node::ProcessInitializationFlags::kNoStdioInitialization,
       node::ProcessInitializationFlags::kNoDefaultSignalHandling,
       node::ProcessInitializationFlags::kDisableNodeOptionsEnv});

  if (!runtime->initialization || runtime->initialization->early_return()) {
    int exit_code = runtime->initialization ? runtime->initialization->exit_code() : -1;
    delete runtime;
    set_error(error_out, "Node initialization failed with exit code " + std::to_string(exit_code));
    return nullptr;
  }

  runtime->platform = runtime->initialization->platform();
  std::vector<std::string> errors;
  runtime->setup = node::CommonEnvironmentSetup::Create(
      runtime->platform,
      &errors,
      runtime->initialization->args(),
      runtime->initialization->exec_args(),
      node::EnvironmentFlags::Flags(
          node::EnvironmentFlags::kDefaultFlags |
          node::EnvironmentFlags::kNoGlobalSearchPaths |
          node::EnvironmentFlags::kNoNativeAddons));

  if (!runtime->setup) {
    std::ostringstream out;
    out << "Node environment setup failed";
    for (const auto& error : errors) out << "\n" << error;
    delete runtime;
    set_error(error_out, out.str());
    return nullptr;
  }

  std::string error;
  if (!run_bootstrap(runtime, module_bundle_json == nullptr ? "" : module_bundle_json, &error)) {
    delete runtime;
    set_error(error_out, "Go-junior bootstrap failed: " + error);
    return nullptr;
  }

  return runtime;
}

extern "C" char* gojr_node_eval(gojr_node_runtime* runtime, const char* source, char** error_out) {
  return call_global_string_function(runtime, "__gojrEval", source, error_out);
}

extern "C" char* gojr_node_eval_files(gojr_node_runtime* runtime, const char* json, char** error_out) {
  return call_global_string_function(runtime, "__gojrEvalFiles", json, error_out);
}

extern "C" char* gojr_node_eval_with_packages(gojr_node_runtime* runtime, const char* json, char** error_out) {
  return call_global_string_function(runtime, "__gojrEvalWithPackages", json, error_out);
}

extern "C" char* gojr_node_eval_files_with_packages(gojr_node_runtime* runtime, const char* json, char** error_out) {
  return call_global_string_function(runtime, "__gojrEvalFilesWithPackages", json, error_out);
}

extern "C" char* gojr_node_compile(gojr_node_runtime* runtime, const char* json, char** error_out) {
  return call_global_string_function(runtime, "__gojrCompile", json, error_out);
}

extern "C" char* gojr_node_test(gojr_node_runtime* runtime, const char* source, char** error_out) {
  return call_global_string_function(runtime, "__gojrTest", source, error_out);
}

extern "C" char* gojr_node_test_files(gojr_node_runtime* runtime, const char* json, char** error_out) {
  return call_global_string_function(runtime, "__gojrTestFiles", json, error_out);
}

extern "C" char* gojr_node_test_files_with_packages(gojr_node_runtime* runtime, const char* json, char** error_out) {
  return call_global_string_function(runtime, "__gojrTestFilesWithPackages", json, error_out);
}

extern "C" char* gojr_node_set_sheet(gojr_node_runtime* runtime, const char* json, char** error_out) {
  return call_global_string_function(runtime, "__gojrSetSheet", json, error_out);
}

extern "C" char* gojr_node_build(gojr_node_runtime* runtime, const char* json, char** error_out) {
  return call_global_string_function(runtime, "__gojrBuild", json, error_out);
}

extern "C" char* gojr_node_inspect_js(gojr_node_runtime* runtime, const char* json, char** error_out) {
  return call_global_string_function(runtime, "__gojrInspectJS", json, error_out);
}

extern "C" char* gojr_node_cache(gojr_node_runtime* runtime, const char* json, char** error_out) {
  return call_global_string_function(runtime, "__gojrCache", json, error_out);
}

extern "C" char* gojr_node_run_fixture(gojr_node_runtime* runtime, const char* json, char** error_out) {
  return call_global_string_function(runtime, "__gojrRunFixture", json, error_out);
}

extern "C" char* gojr_node_run_fixture_with_packages(gojr_node_runtime* runtime, const char* json, char** error_out) {
  return call_global_string_function(runtime, "__gojrRunFixtureWithPackages", json, error_out);
}

extern "C" void gojr_node_free(gojr_node_runtime* runtime) {
  delete runtime;
}

extern "C" void gojr_string_free(char* value) {
  std::free(value);
}
