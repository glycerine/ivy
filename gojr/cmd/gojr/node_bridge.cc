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

std::string bootstrap_source(const std::string& module_path) {
  return R"JS(
const { createRequire } = require('node:module');
const gojrRequire = createRequire(process.cwd() + '/gojr-embedded-repl.js');
const gojrModule = gojrRequire()JS" + js_string_literal(module_path) + R"JS();
let gojrSheet = {};

function gojrDiagnosticString(diagnostic) {
  const location = diagnostic && diagnostic.span
    ? `${diagnostic.span.line}:${diagnostic.span.column}: `
    : "";
  return `${location}${diagnostic.severity} ${diagnostic.code}: ${diagnostic.message}`;
}

function gojrFormatResult(result) {
  if (result.values) return result.values.map((value) => gojrModule.formatValue(value)).join(", ");
  if (Object.prototype.hasOwnProperty.call(result, "value")) return gojrModule.formatValue(result.value);
  return "";
}

globalThis.__gojrEval = function(source) {
  const result = gojrModule.evaluateSource(source, { sheet: gojrSheet });
  const diagnostics = result.diagnostics || [];
  return JSON.stringify({
    ok: !diagnostics.some((diagnostic) => diagnostic.severity === "error"),
    diagnostics: diagnostics.map(gojrDiagnosticString),
    output: (result.output || []).join(""),
    value: gojrFormatResult(result)
  });
};

globalThis.__gojrSetSheet = function(json) {
  gojrSheet = gojrModule.parseSheetJson(json);
  return JSON.stringify({ ok: true, value: "sheet loaded" });
};
)JS";
}

bool run_bootstrap(gojr_node_runtime* runtime, const std::string& module_path, std::string* error) {
  v8::Isolate* isolate = runtime->setup->isolate();
  v8::Locker locker(isolate);
  v8::Isolate::Scope isolate_scope(isolate);
  v8::HandleScope handle_scope(isolate);
  v8::Local<v8::Context> context = runtime->setup->context();
  v8::Context::Scope context_scope(context);
  v8::TryCatch try_catch(isolate);

  std::string source = bootstrap_source(module_path);
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

  v8::Local<v8::String> result_string;
  if (!result->ToString(context).ToLocal(&result_string)) {
    set_error(error_out, exception_to_string(isolate, try_catch));
    return nullptr;
  }
  return copy_string(v8_to_string(isolate, result_string));
}

}  // namespace

extern "C" gojr_node_runtime* gojr_node_new(const char* module_path, char** error_out) {
  auto* runtime = new gojr_node_runtime();

  std::vector<std::string> args = {"gojr-embedded-node"};
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
  if (!run_bootstrap(runtime, module_path == nullptr ? "" : module_path, &error)) {
    delete runtime;
    set_error(error_out, "Go-junior bootstrap failed: " + error);
    return nullptr;
  }

  return runtime;
}

extern "C" char* gojr_node_eval(gojr_node_runtime* runtime, const char* source, char** error_out) {
  return call_global_string_function(runtime, "__gojrEval", source, error_out);
}

extern "C" char* gojr_node_set_sheet(gojr_node_runtime* runtime, const char* json, char** error_out) {
  return call_global_string_function(runtime, "__gojrSetSheet", json, error_out);
}

extern "C" void gojr_node_free(gojr_node_runtime* runtime) {
  delete runtime;
}

extern "C" void gojr_string_free(char* value) {
  std::free(value);
}
