export function currentIvyBridge(win = globalThis.window) {
  return win && win.__ivyVueBridge;
}

export function createIvyApi({
  bridge = currentIvyBridge(),
  fallbackApiFactory,
} = {}) {
  if (bridge && typeof bridge.createIvyApi === 'function') {
    try {
      return bridge.createIvyApi();
    } catch (err) {
      console.warn('Vue engine bridge unavailable, falling back to IvyAPI:', err);
    }
  }
  if (typeof fallbackApiFactory === 'function') {
    return fallbackApiFactory();
  }
  if (bridge && typeof bridge.getEngine === 'function') {
    return bridge.getEngine();
  }
  throw new Error('No Ivy API factory is available');
}

export async function createSession(api, { controls } = {}) {
  try {
    await api.createSession();
    return api.sessionId;
  } catch (err) {
    if (controls && typeof controls.setStatus === 'function') {
      controls.setStatus(`Failed to create session: ${err.message}`, 'error');
    }
    console.error('Session creation failed:', err);
    throw err;
  }
}

export function updateSessionDisplay(sessionId, {
  bridge = currentIvyBridge(),
  doc = globalThis.document,
} = {}) {
  const displaySessionId = sessionId || '';
  if (bridge && typeof bridge.setSessionId === 'function') {
    bridge.setSessionId(displaySessionId);
    return true;
  }
  const sessionEl = doc && doc.getElementById('session-id');
  if (sessionEl) {
    sessionEl.textContent = displaySessionId ? `Session: ${displaySessionId}` : '';
    return true;
  }
  return false;
}

export function connectSessionEvents(api, onEvent, onConnectionLost) {
  if (!api || !api.sessionId || typeof api.connectEvents !== 'function') {
    return false;
  }
  api.connectEvents(onEvent);
  api.onConnectionLost = onConnectionLost;
  return true;
}
