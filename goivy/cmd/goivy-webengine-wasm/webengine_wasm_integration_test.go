//go:build !js || !wasm

package webenginewasm_test

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBrowserWasmClientServerInductionKeepsGoRuntimeAlive(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser wasm integration test in short mode")
	}
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is required for browser wasm integration test")
	}

	root := findGoivyRoot(t)
	specPath := filepath.Join(root, "..", "ivy-lang-examples", "doc", "examples", "client_server_example.ivy")
	requireFile(t, specPath)

	tmp := t.TempDir()
	assetDir := filepath.Join(tmp, "assets")
	if err := os.Mkdir(assetDir, 0o755); err != nil {
		t.Fatalf("create temp wasm asset dir: %v", err)
	}
	for _, name := range []string{
		"goivy-webengine.wasm",
		"wasm_exec-go1.25.6.js",
		"z3-471-api.js",
		"z3-471-api.wasm",
		"include-tree.json",
	} {
		copyFile(t, filepath.Join(root, "webui", "static", "wasm", name), filepath.Join(assetDir, name))
	}

	driverPath := filepath.Join(tmp, "drive-webengine-wasm.mjs")
	if err := os.WriteFile(driverPath, []byte(webengineWASMDriver), 0o644); err != nil {
		t.Fatalf("write wasm driver: %v", err)
	}
	cfgPath := filepath.Join(tmp, "driver-config.json")
	cfg := map[string]string{
		"assetDir":    assetDir,
		"worker":      filepath.Join(root, "webui", "frontend", "src", "workers", "browserWasmEngine.worker.js"),
		"includeRoot": "/include",
		"specPath":    specPath,
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal wasm driver config: %v", err)
	}
	if err := os.WriteFile(cfgPath, data, 0o644); err != nil {
		t.Fatalf("write wasm driver config: %v", err)
	}

	run := exec.Command(node, driverPath, cfgPath)
	run.Dir = root
	out, err := run.CombinedOutput()
	if err != nil {
		t.Fatalf("run browser wasm client_server induction: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), `"result":"pass"`) && !strings.Contains(string(out), `"result":"fail"`) {
		t.Fatalf("browser wasm induction did not report a completed check result:\n%s", out)
	}
}

func findGoivyRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find goivy root from %s", dir)
		}
		dir = parent
	}
}

func requireFile(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("required file %s: %v", path, err)
	}
}

func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	in, err := os.Open(src)
	if err != nil {
		t.Fatalf("open %s: %v", src, err)
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		t.Fatalf("create %s: %v", dst, err)
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		t.Fatalf("copy %s to %s: %v", src, dst, err)
	}
	if err := out.Close(); err != nil {
		t.Fatalf("close %s: %v", dst, err)
	}
}

const webengineWASMDriver = `
import fs from 'node:fs';
import path from 'node:path';
import { webcrypto } from 'node:crypto';
import { performance } from 'node:perf_hooks';
import { fileURLToPath, pathToFileURL } from 'node:url';

const nodeProcess = globalThis.process;
const startedAt = performance.now();

function diag(message) {
  const elapsedMs = Math.round(performance.now() - startedAt);
  fs.writeSync(2, '[webengine-wasm-test +' + elapsedMs + 'ms] ' + message + '\n');
}

function readJSON(file) {
  return JSON.parse(fs.readFileSync(file, 'utf8'));
}

function fileURL(file) {
  return pathToFileURL(path.resolve(file)).href;
}

function installHostGlobals() {
  if (!globalThis.crypto) globalThis.crypto = webcrypto;
  if (!globalThis.performance) globalThis.performance = performance;
  if (!globalThis.TextEncoder) globalThis.TextEncoder = TextEncoder;
  if (!globalThis.TextDecoder) globalThis.TextDecoder = TextDecoder;
  globalThis.process = undefined;
  globalThis.XMLHttpRequest = class FileXMLHttpRequest {
    constructor() {
      this.responseType = '';
      this.response = null;
      this.responseText = '';
      this.status = 0;
      this.readyState = 0;
      this.onload = null;
      this.onerror = null;
      this.onreadystatechange = null;
    }
    open(method, url, async = true) {
      this.method = method;
      this.url = String(url);
      this.async = async;
      this.readyState = 1;
    }
    setRequestHeader() {}
    overrideMimeType() {}
    send() {
      try {
        const file = this.url.startsWith('file:') ? fileURLToPath(this.url) : this.url;
        const bytes = fs.readFileSync(file);
        this.status = 200;
        this.readyState = 4;
        if (this.responseType === 'arraybuffer') {
          this.response = bytes.buffer.slice(bytes.byteOffset, bytes.byteOffset + bytes.byteLength);
        } else {
          this.responseText = bytes.toString('utf8');
          this.response = this.responseText;
        }
        if (typeof this.onreadystatechange === 'function') this.onreadystatechange();
        if (typeof this.onload === 'function') this.onload();
      } catch (error) {
        this.status = 404;
        this.readyState = 4;
        if (typeof this.onreadystatechange === 'function') this.onreadystatechange();
        if (typeof this.onerror === 'function') {
          this.onerror(error);
          return;
        }
        throw error;
      }
    }
  };
}

function installFileFetch() {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async (input) => {
    const url = typeof input === 'string' ? input : String(input && input.url || input);
    if (!url.startsWith('file:')) {
      if (typeof originalFetch === 'function') return originalFetch(input);
      throw new Error('unsupported fetch URL in wasm worker test: ' + url);
    }
    const file = fileURLToPath(url);
    const bytes = fs.readFileSync(file);
    const headers = new Headers();
    if (file.endsWith('.wasm')) headers.set('Content-Type', 'application/wasm');
    if (file.endsWith('.js')) headers.set('Content-Type', 'text/javascript');
    if (file.endsWith('.json')) headers.set('Content-Type', 'application/json');
    return new Response(bytes, { status: 200, statusText: 'OK', headers });
  };
}

function installWorkerHost() {
  const pending = new Map();
  const events = [];
  globalThis.self = {
    onmessage: null,
    postMessage(message) {
      if (message && message.type === 'event') {
        events.push(message.event || message);
        return;
      }
      if (message && (message.type === 'stdout' || message.type === 'stderr')) {
        return;
      }
      const requestId = message && message.requestId;
      const waiter = requestId ? pending.get(requestId) : null;
      if (!waiter) {
        events.push(message);
        return;
      }
      pending.delete(requestId);
      if (message.type === 'error') {
        waiter.reject(new Error(message.error || 'browser wasm worker failed'));
      } else {
        waiter.resolve(message.value);
      }
    },
  };
  globalThis.importScripts = () => {};
  return {
    events,
    send(request) {
      if (typeof globalThis.self.onmessage !== 'function') {
        return Promise.reject(new Error('worker did not install onmessage'));
      }
      return new Promise((resolve, reject) => {
        pending.set(request.requestId, { resolve, reject });
        globalThis.self.onmessage({ data: request });
      });
    },
  };
}

function assetBaseURL(dir) {
  const text = fileURL(dir);
  return text.endsWith('/') ? text : text + '/';
}

async function sendWithTimeout(host, request, timeoutMs = 60000) {
  let timeout;
  try {
    return await Promise.race([
      host.send(request),
      new Promise((_, reject) => {
        timeout = setTimeout(() => reject(new Error('timed out waiting for ' + request.type)), timeoutMs);
      }),
    ]);
  } finally {
    clearTimeout(timeout);
  }
}

async function main() {
  if (nodeProcess.argv.length !== 3) {
    throw new Error('usage: node drive-webengine-wasm.mjs config.json');
  }
  const cfg = readJSON(nodeProcess.argv[2]);
  installHostGlobals();
  installFileFetch();
  const host = installWorkerHost();
  globalThis.location = { href: fileURL(cfg.worker) };
  globalThis.self.location = globalThis.location;
  const assets = assetBaseURL(cfg.assetDir);

  diag('importing browser wasm worker');
  await import(fileURL(cfg.worker));
  await sendWithTimeout(host, {
    type: 'init',
    requestId: 'init-1',
    assetBaseUrl: assets,
    includeRoot: cfg.includeRoot || '/include',
  });

  diag('loading client_server_example.ivy');
  const session = await sendWithTimeout(host, {
    type: 'new-session',
    requestId: 'new-session-1',
    projectId: 'webui',
  });
  const sessionId = session && (session.id || session.session_id || session.sessionId);
  if (!sessionId) throw new Error('wasm webengine returned no session id');

  const source = fs.readFileSync(cfg.specPath, 'utf8');
  const load = await sendWithTimeout(host, {
    type: 'load-model',
    requestId: 'load-model-1',
    sessionId,
    model: {
      id: 'client-server',
      projectId: 'webui',
      filename: 'client_server_example.ivy',
      text: source,
      isolate: '',
      engineRevision: 1,
    },
  });
  if (!load || load.status !== 'ok') {
    throw new Error('load-model returned ' + JSON.stringify(load));
  }

  diag('running induction');
  const check = await sendWithTimeout(host, {
    type: 'run-command',
    requestId: 'check-induction-1',
    sessionId,
    intent: {
      commandId: 'check.induction',
      args: { mode: 'induction' },
    },
  });
  if (!check || (check.result !== 'pass' && check.result !== 'fail')) {
    throw new Error('check.induction returned ' + JSON.stringify(check));
  }

  diag('checking runtime survived induction');
  const snapshot = await sendWithTimeout(host, {
    type: 'get-snapshot',
    requestId: 'snapshot-1',
    sessionId,
    snapshot: { arg: true },
  });
  if (!snapshot || !snapshot.arg || !Array.isArray(snapshot.arg.elements)) {
    throw new Error('snapshot after induction returned ' + JSON.stringify(snapshot));
  }

  diag('checking concept snapshot after induction');
  const conceptSnapshot = await sendWithTimeout(host, {
    type: 'get-snapshot',
    requestId: 'snapshot-concept-1',
    sessionId,
    snapshot: { concept: {} },
  });
  if (!conceptSnapshot || !conceptSnapshot.concept || !Array.isArray(conceptSnapshot.concept.elements)) {
    throw new Error('concept snapshot after induction returned ' + JSON.stringify(conceptSnapshot));
  }

  nodeProcess.stdout.write(JSON.stringify({
    result: check.result,
    argElements: snapshot.arg.elements.length,
    conceptElements: conceptSnapshot.concept.elements.length,
  }) + '\n', () => {
    nodeProcess.exit(0);
  });
}

main().catch((error) => {
  fs.writeSync(2, (error && error.stack ? error.stack : String(error)) + '\n');
  nodeProcess.exit(1);
});
`
