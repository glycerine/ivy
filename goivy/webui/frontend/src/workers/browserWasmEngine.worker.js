const DEFAULT_ASSET_BASE_URL = '/static/wasm/';
const DEFAULT_INCLUDE_ROOT = '/include';

let assetBaseUrl = DEFAULT_ASSET_BASE_URL;
let includeRoot = DEFAULT_INCLUDE_ROOT;
let initialized = false;
let wasmReady = null;
let sequence = 0;
const sessions = new Map();

self.onmessage = (event) => {
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
  wasmReady = wasmReady || (async () => {
    const [{ createSmtZ3Imports }, { installGoIvyNodeFS }] = await Promise.all([
      import('./smtZ3Imports.js'),
      import('./goivyNodeFS.js'),
    ]);
    const includeTree = await loadIncludeTree();
    const z3 = await loadZ3();
    const fsHost = installGoIvyNodeFS({
      includeRoot,
      includeTree,
      stdout: (bytes) => post({ type: 'stdout', value: fsHost.decode(bytes) }),
      stderr: (bytes) => post({ type: 'stderr', value: fsHost.decode(bytes) }),
    });
    await import(/* @vite-ignore */ `${assetBaseUrl}wasm_exec-go1.25.6.js`);
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
    void go.run(result.instance);
    await waitForDispatch();
  })();
  return wasmReady;
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
      post({ type: 'stdout', value: String(text) });
    },
    printErr(text) {
      post({ type: 'stderr', value: String(text) });
    },
    locateFile(path) {
      if (String(path).endsWith('.wasm')) return `${assetBaseUrl}z3-471-api.wasm`;
      return `${assetBaseUrl}${path}`;
    },
  });
}

async function waitForDispatch() {
  for (let i = 0; i < 2000; i += 1) {
    if (typeof globalThis.goivyWebEngineDispatch === 'function') return;
    await new Promise((resolve) => setTimeout(resolve, 5));
  }
  throw new Error('goivy webengine wasm did not register its dispatcher');
}

async function callWasm(request) {
  await loadWebEngineWasm();
  const dispatch = globalThis.goivyWebEngineDispatch;
  if (typeof dispatch !== 'function') {
    throw new Error('goivy webengine wasm dispatcher is unavailable');
  }
  const response = JSON.parse(dispatch(JSON.stringify(request)));
  if (response.type === 'error') {
    throw new Error(response.error || 'goivy webengine wasm failed');
  }
  return response.value;
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

function post(response) {
  self.postMessage(response);
}
