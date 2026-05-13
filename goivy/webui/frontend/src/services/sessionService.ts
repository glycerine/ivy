export function createIvyApi({
  fallbackApiFactory = null,
}: { fallbackApiFactory?: null | (() => any) } = {}) {
  const win = globalThis.window;
  if (win && win.__IVY_ENGINE__) {
    return win.__IVY_ENGINE__;
  }
  if (typeof fallbackApiFactory === 'function') {
    return fallbackApiFactory();
  }
  throw new Error('No Ivy API factory is available');
}

export async function createSession(api: any, {
  controls = null,
}: { controls?: any } = {}) {
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
  doc = globalThis.document,
} = {}) {
  const displaySessionId = sessionId || '';
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
