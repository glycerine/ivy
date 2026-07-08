const DEFAULT_ASSET_BASE_URL = '/static/wasm/';
const DEFAULT_INCLUDE_ROOT = '/include';
const MAX_RUNTIME_LOG_CHARS = 20000;

let assetBaseUrl = DEFAULT_ASSET_BASE_URL;
let includeRoot = DEFAULT_INCLUDE_ROOT;
let initialized = false;
let wasmReady = null;
let wasmGeneration = 0;
let wasmRuntimeError = null;
let wasmRuntimeCrashReported = false;
let runtimeLogEntries = [];
let sequence = 0;
const sessions = new Map();
const workerScope = globalThis.self || globalThis;

workerScope.onmessage = (event) => {
  void handleMessage(parseMessage(event.data));
};

function parseMessage(data) {
  if (typeof data === 'string') return JSON.parse(data);
  return data || {};
}

async function handleMessage(request) {
  try {
    switch (request.type) {
      case 'init':
        assetBaseUrl = normalizeAssetBaseUrl(request.assetBaseUrl || DEFAULT_ASSET_BASE_URL);
        includeRoot = request.includeRoot || DEFAULT_INCLUDE_ROOT;
        await loadWebEngineWasm();
        initialized = true;
        post({ type: 'ready', requestId: request.requestId, value: { ok: true } });
        break;

      case 'new-session': {
        requireInit();
        const session = await callWasm({
          type: 'new-session',
          requestId: request.requestId,
          projectId: request.projectId || 'webui',
        });
        const sessionId = session && (session.id || session.session_id);
        if (!sessionId) throw new Error('browser wasm engine returned no session id');
        sessions.set(sessionId, {
          id: sessionId,
          projectId: request.projectId || 'webui',
          currentModel: null,
        });
        post({ type: 'session', requestId: request.requestId, value: session });
        postEvent(sessionId, { type: 'session-ready', data: session });
        break;
      }

      case 'load-model': {
        requireInit();
        const session = requireSession(request.sessionId);
        session.currentModel = request.model || null;
        postEvent(session.id, {
          type: 'job-progress',
          data: {
            phase: 'browser-wasm-load',
            message: 'Loading model in browser engine...',
          },
        });
        const value = await callWasm({
          type: 'load-model',
          requestId: request.requestId,
          sessionId: session.id,
          model: request.model || {},
        });
        postJobComplete(session.id, {
          phase: 'browser-wasm-load',
          message: 'Model loaded in browser engine.',
          value,
        });
        post({ type: 'load-result', requestId: request.requestId, value });
        break;
      }

      case 'run-command': {
        requireInit();
        const session = requireSession(request.sessionId);
        postEvent(session.id, {
          type: 'job-progress',
          data: {
            phase: 'browser-wasm-command',
            commandId: request.intent && request.intent.commandId,
            message: 'Running command in browser engine...',
          },
        });
        const value = await callWasm({
          type: 'run-command',
          requestId: request.requestId,
          sessionId: session.id,
          intent: request.intent || {},
        });
        postJobComplete(session.id, {
          phase: 'browser-wasm-command',
          commandId: request.intent && request.intent.commandId,
          message: commandCompleteMessage(value),
          value,
        });
        post({ type: 'command-result', requestId: request.requestId, value });
        break;
      }

      case 'get-snapshot': {
        requireInit();
        const session = requireSession(request.sessionId);
        const value = await callWasm({
          type: 'get-snapshot',
          requestId: request.requestId,
          sessionId: session.id,
          snapshot: request.snapshot || {},
        });
        post({ type: 'snapshot', requestId: request.requestId, value });
        break;
      }

      case 'cancel-job':
        post({ type: 'cancelled', requestId: request.requestId, value: { status: 'cancelled' } });
        break;

      default:
        throw new Error('unknown browser wasm worker request: ' + request.type);
    }
  } catch (error) {
    post({
      type: 'error',
      requestId: request && request.requestId,
      error: error && error.message ? error.message : String(error),
    });
  }
}

async function loadWebEngineWasm() {
  if (wasmReady) return wasmReady;
  const generation = wasmGeneration + 1;
  wasmGeneration = generation;
  wasmRuntimeError = null;
  wasmRuntimeCrashReported = false;
  runtimeLogEntries = [];
  wasmReady = (async () => {
    const [{ createSmtZ3Imports }, { installGoIvyNodeFS }] = await Promise.all([
      import('./smtZ3Imports.js'),
      import('./goivyNodeFS.js'),
    ]);
    const includeTree = await loadIncludeTree();
    const z3 = await loadZ3();
    const fsHost = installGoIvyNodeFS({
      includeRoot,
      includeTree,
      stdout: (bytes) => {
        const text = fsHost.decode(bytes);
        appendRuntimeLog('stdout', text);
        post({ type: 'stdout', value: text });
      },
      stderr: (bytes) => {
        const text = fsHost.decode(bytes);
        appendRuntimeLog('stderr', text);
        post({ type: 'stderr', value: text });
      },
    });
    if (typeof globalThis.Go !== 'function') {
      await import(/* @vite-ignore */ `${assetBaseUrl}wasm_exec-go1.25.6.js`);
    }
    if (typeof globalThis.Go !== 'function') {
      throw new Error('wasm_exec-go1.25.6.js did not expose Go');
    }
    let wasmMemory = null;
    const go = new globalThis.Go();
    go.argv = ['goivy-webengine-wasm'];
    go.env = {
      GOIVY_INCLUDE: includeRoot,
      GOIVY_WASM_MEMORY_LIMIT: '3GiB',
      GOIVY_WASM_GOGC: '50',
    };
    go.importObject.smt_z3 = createSmtZ3Imports({ z3, getGoMemory: () => wasmMemory });
    const result = await instantiateWasm(`${assetBaseUrl}goivy-webengine.wasm`, go.importObject);
    wasmMemory = result.instance.exports.mem || null;
    let runPromise;
    try {
      runPromise = go.run(result.instance);
    } catch (error) {
      resetWasmRuntime(generation, error, 'go.run threw before startup completed');
      throw error;
    }
    Promise.resolve(runPromise).then(
      () => resetWasmRuntime(generation, new Error('goivy webengine wasm exited'), 'go.run resolved'),
      (error) => resetWasmRuntime(generation, error, 'go.run rejected'),
    );
    await waitForDispatch(() => (generation === wasmGeneration ? wasmRuntimeError : null));
  })();
  return wasmReady;
}

function discardWasmRuntime(error) {
  wasmReady = null;
  wasmRuntimeError = error || null;
  globalThis.goivyWebEngineDispatch = undefined;
}

function resetWasmRuntime(generation, error, reason = 'Go runtime exited') {
  if (generation !== wasmGeneration) return;
  reportWasmRuntimeCrash(error, reason);
  discardWasmRuntime(error);
}

function appendRuntimeLog(stream, text) {
  const value = String(text || '');
  if (!value) return;
  runtimeLogEntries.push({ stream, text: value });
  let total = runtimeLogEntries.reduce((sum, entry) => sum + entry.text.length, 0);
  while (total > MAX_RUNTIME_LOG_CHARS && runtimeLogEntries.length > 0) {
    const first = runtimeLogEntries[0];
    const excess = total - MAX_RUNTIME_LOG_CHARS;
    if (first.text.length <= excess) {
      runtimeLogEntries.shift();
      total -= first.text.length;
    } else {
      first.text = first.text.slice(excess);
      total -= excess;
    }
  }
}

function runtimeLogText() {
  return runtimeLogEntries.map((entry) => `[${entry.stream}] ${entry.text}`).join('');
}

function errorMessage(error) {
  return String((error && error.message) || error || 'Go WASM runtime exited');
}

function errorStack(error) {
  return String((error && error.stack) || '');
}

function reportWasmRuntimeCrash(error, reason) {
  if (wasmRuntimeCrashReported) return;
  wasmRuntimeCrashReported = true;
  post({
    type: 'event',
    requestId: nextId('event'),
    event: {
      type: 'browser_wasm_runtime_crash',
      data: {
        reason: reason || 'Go WASM runtime exited',
        message: errorMessage(error),
        stack: errorStack(error),
        recent_output: runtimeLogText(),
        asset_base_url: assetBaseUrl,
        generation: wasmGeneration,
        timestamp: new Date().toISOString(),
      },
    },
  });
}

async function instantiateWasm(url, imports) {
  try {
    return await WebAssembly.instantiateStreaming(fetch(url), imports);
  } catch (_err) {
    const response = await fetch(url);
    if (!response.ok) {
      throw new Error(`failed to load Go Ivy wasm: ${response.status} ${response.statusText}`);
    }
    return WebAssembly.instantiate(await response.arrayBuffer(), imports);
  }
}

async function loadIncludeTree() {
  const response = await fetch(`${assetBaseUrl}include-tree.json`);
  if (!response.ok) {
    throw new Error(`failed to load Ivy include tree: ${response.status} ${response.statusText}`);
  }
  const tree = await response.json();
  return { root: includeRoot, files: Array.isArray(tree.files) ? tree.files : [] };
}

async function loadZ3() {
  const source = await (await fetch(`${assetBaseUrl}z3-471-api.js`)).text();
  const initZ3 = new Function(`${source}; return initZ3;`)();
  return initZ3({
    print(text) {
      const value = String(text);
      appendRuntimeLog('stdout', value);
      post({ type: 'stdout', value });
    },
    printErr(text) {
      const value = String(text);
      appendRuntimeLog('stderr', value);
      post({ type: 'stderr', value });
    },
    locateFile(path) {
      if (String(path).endsWith('.wasm')) return `${assetBaseUrl}z3-471-api.wasm`;
      return `${assetBaseUrl}${path}`;
    },
  });
}

async function waitForDispatch(runtimeError) {
  for (let i = 0; i < 2000; i += 1) {
    const error = runtimeError && runtimeError();
    if (error) throw error;
    if (typeof globalThis.goivyWebEngineDispatch === 'function') return;
    await new Promise((resolve) => setTimeout(resolve, 5));
  }
  throw new Error('goivy webengine wasm did not register its dispatcher');
}

async function callWasm(request, retryOnExited = true) {
  await loadWebEngineWasm();
  const dispatch = globalThis.goivyWebEngineDispatch;
  if (typeof dispatch !== 'function') {
    if (retryOnExited) {
      discardWasmRuntime(new Error('goivy webengine wasm dispatcher is unavailable'));
      return callWasm(request, false);
    }
    throw new Error('goivy webengine wasm dispatcher is unavailable');
  }
  let rawResponse;
  try {
    rawResponse = dispatch(JSON.stringify(request));
  } catch (error) {
    if (retryOnExited && isGoProgramExited(error)) {
      reportWasmRuntimeCrash(error, `stale dispatcher while handling ${request.type || 'request'}`);
      discardWasmRuntime(error);
      return callWasm(request, false);
    }
    throw error;
  }
  let response;
  try {
    response = parseWasmResponse(rawResponse);
  } catch (error) {
    if (retryOnExited && isMissingWasmResponse(rawResponse)) {
      const runtimeError = wasmRuntimeError || error;
      reportWasmRuntimeCrash(runtimeError, `missing dispatcher response while handling ${request.type || 'request'}`);
      discardWasmRuntime(runtimeError);
      return callWasm(request, false);
    }
    throw error;
  }
  if (response.type === 'error') {
    throw new Error(response.error || 'goivy webengine wasm failed');
  }
  return response.value;
}

function parseWasmResponse(rawResponse) {
  if (rawResponse === undefined) {
    throw new Error('goivy webengine wasm returned no response');
  }
  if (rawResponse === null) {
    throw new Error('goivy webengine wasm returned null response');
  }
  if (typeof rawResponse !== 'string') {
    throw new Error(`goivy webengine wasm returned a non-string response: ${typeof rawResponse}`);
  }
  if (rawResponse.trim() === '') {
    throw new Error('goivy webengine wasm returned an empty response');
  }
  try {
    return JSON.parse(rawResponse);
  } catch (error) {
    throw new Error(`goivy webengine wasm returned invalid JSON: ${errorMessage(error)}`);
  }
}

function isMissingWasmResponse(rawResponse) {
  return rawResponse === undefined || rawResponse === null || rawResponse === '';
}

function isGoProgramExited(error) {
  return String((error && error.message) || error || '').includes('Go program has already exited');
}

function requireInit() {
  if (!initialized) throw new Error('Browser WASM engine is not initialized');
}

function requireSession(sessionId) {
  const session = sessions.get(sessionId);
  if (!session) throw new Error('Unknown browser WASM session: ' + sessionId);
  return session;
}

function normalizeAssetBaseUrl(value) {
  const text = String(value || DEFAULT_ASSET_BASE_URL);
  return text.endsWith('/') ? text : `${text}/`;
}

function nextId(prefix) {
  sequence += 1;
  return `${prefix}-${sequence}`;
}

function postEvent(sessionId, event) {
  post({ type: 'event', requestId: nextId('event'), sessionId, event });
}

function postJobComplete(sessionId, { phase, commandId, message, value }) {
  postEvent(sessionId, {
    type: 'job-progress',
    data: {
      phase,
      commandId,
      status: 'complete',
      level: resultLevel(value),
      message,
    },
  });
}

function commandCompleteMessage(value) {
  if (value && typeof value.message === 'string' && value.message.trim()) {
    return value.message.trim();
  }
  const verdict = value && (value.result || value.status);
  if (verdict && verdict !== 'ok') {
    return `Browser command complete: ${verdict}`;
  }
  return 'Browser command complete.';
}

function resultLevel(value) {
  const verdict = String(value && (value.result || value.status) || '').toLowerCase();
  if (verdict === 'error' || verdict === 'fail' || verdict === 'failed') return 'error';
  if (verdict === 'warning' || verdict === 'cancelled' || verdict === 'canceled') return 'warning';
  return 'success';
}

function post(response) {
  workerScope.postMessage(response);
}
