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

bool settle_promise(
    gojr_node_runtime* runtime,
    const char* function_name,
    v8::Local<v8::Value>* result,
    v8::TryCatch& try_catch,
    std::string* error) {
  if (!(*result)->IsPromise()) return true;

  v8::Isolate* isolate = runtime->setup->isolate();
  v8::Local<v8::Promise> promise = (*result).As<v8::Promise>();
  while (promise->State() == v8::Promise::kPending) {
    isolate->PerformMicrotaskCheckpoint();
    if (promise->State() != v8::Promise::kPending) break;
    v8::Maybe<int> spun = node::SpinEventLoop(runtime->setup->env());
    if (spun.IsNothing()) {
      *error = exception_to_string(isolate, try_catch);
      return false;
    }
    isolate->PerformMicrotaskCheckpoint();
    if (promise->State() == v8::Promise::kPending) {
      *error = "Go-junior operation did not complete";
      return false;
    }
  }
  if (promise->State() == v8::Promise::kRejected) {
    *error = value_exception_to_string(isolate, promise->Result());
    return false;
  }
  *result = promise->Result();
  return true;
}

bool run_bootstrap(
    gojr_node_runtime* runtime,
    const std::string& bootstrap_source,
    const std::string& module_bundle_json,
    std::string* error) {
  if (bootstrap_source.empty()) {
    *error = "embedded Go-junior bootstrap source is empty";
    return false;
  }

  v8::Isolate* isolate = runtime->setup->isolate();
  v8::Locker locker(isolate);
  v8::Isolate::Scope isolate_scope(isolate);
  v8::HandleScope handle_scope(isolate);
  v8::Local<v8::Context> context = runtime->setup->context();
  v8::Context::Scope context_scope(context);
  v8::TryCatch try_catch(isolate);

  if (std::getenv("GOJR_DEBUG_BOOTSTRAP") != nullptr) {
    std::fprintf(stderr, "%s\n", bootstrap_source.c_str());
  }
  v8::MaybeLocal<v8::Value> loaded = node::LoadEnvironment(runtime->setup->env(), bootstrap_source);
  if (loaded.IsEmpty()) {
    *error = exception_to_string(isolate, try_catch);
    return false;
  }

  v8::Local<v8::String> installer_name = v8::String::NewFromUtf8Literal(isolate, "__gojrInstallEmbeddedRuntime");
  v8::Local<v8::Value> installer_value;
  if (!context->Global()->Get(context, installer_name).ToLocal(&installer_value) || !installer_value->IsFunction()) {
    *error = "__gojrInstallEmbeddedRuntime is not installed by the embedded Go-junior bootstrap";
    return false;
  }

  v8::Local<v8::Function> installer = installer_value.As<v8::Function>();
  v8::Local<v8::Value> argv[1] = {
      v8::String::NewFromUtf8(isolate, module_bundle_json.c_str()).ToLocalChecked()};
  v8::Local<v8::Value> result;
  if (!installer->Call(context, context->Global(), 1, argv).ToLocal(&result)) {
    *error = exception_to_string(isolate, try_catch);
    return false;
  }
  return settle_promise(runtime, "__gojrInstallEmbeddedRuntime", &result, try_catch, error);
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

  std::string error;
  if (!settle_promise(runtime, function_name, &result, try_catch, &error)) {
    set_error(error_out, error);
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

extern "C" gojr_node_runtime* gojr_node_new(
    const char* bootstrap_source,
    const char* module_bundle_json,
    char** error_out) {
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
  if (!run_bootstrap(
          runtime,
          bootstrap_source == nullptr ? "" : bootstrap_source,
          module_bundle_json == nullptr ? "" : module_bundle_json,
          &error)) {
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

extern "C" char* gojr_node_run_main_files_with_packages(gojr_node_runtime* runtime, const char* json, char** error_out) {
  return call_global_string_function(runtime, "__gojrRunMainFilesWithPackages", json, error_out);
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
