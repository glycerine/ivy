import { connectSessionEvents, createSession } from './sessionService.ts';

export function recentFileItems(sessions, truncatePath) {
  const seen = {};
  const unique = [];
  for (const sess of sessions || []) {
    if (!sess.fileName || sess.fileName === '(unnamed)') continue;
    const dedupKey = sess.filePath || sess.fileName;
    if (seen[dedupKey]) continue;
    seen[dedupKey] = true;
    unique.push(sess);
  }

  const baseNameCount = {};
  for (const session of unique) {
    baseNameCount[session.fileName] = (baseNameCount[session.fileName] || 0) + 1;
  }

  return unique.slice(0, 6).map((session) => {
    let label = session.fileName;
    if (baseNameCount[session.fileName] > 1 && session.filePath) {
      label = `${session.fileName}  ${truncatePath(session.filePath, 15)}`;
    }
    let title = '';
    if (session.timestamp) {
      title = `${session.filePath || session.fileName}\nLast used: ${new Date(session.timestamp).toLocaleString()}`;
    }
    return { id: session.id, label, title, session };
  });
}

export function populateRecentFiles(app, persist, {
  doc = globalThis.document,
} = {}) {
  const container = doc && doc.getElementById('file-recent-list');
  if (!container) return;
  container.innerHTML = '';

  const sessions = persist.listSessions();
  if (sessions.length === 0) {
    const empty = doc.createElement('a');
    empty.href = '#';
    empty.textContent = '(no recent files)';
    empty.style.color = '#666';
    empty.style.pointerEvents = 'none';
    container.appendChild(empty);
    return;
  }

  const items = recentFileItems(sessions, persist.truncatePath);
  for (const item of items) {
    const link = doc.createElement('a');
    link.href = '#';
    link.textContent = item.label;
    link.setAttribute('data-session-id', item.id);
    if (item.title) link.title = item.title;
    link.addEventListener('click', function (event) {
      event.preventDefault();
      if (typeof app.flashAndClose === 'function') {
        app.flashAndClose(this, () => app.loadRecentSession(item.id));
      } else {
        app.loadRecentSession(item.id);
      }
    });
    container.appendChild(link);
  }
}

export async function loadRecentSession(app, persist, savedSessionId) {
  if (app._persistedFileContent) {
    persist.save(app);
  }

  const state = persist.loadSession(savedSessionId);
  if (!state || !state.fileContent) {
    app.controls.setStatus('Could not load session: no saved data', 'error');
    return;
  }

  try {
    await createSession(app.api, { controls: app.controls });
  } catch (_err) {
    return;
  }

  connectSessionEvents(app.api, app.handleEvent.bind(app), app.handleConnectionLost && app.handleConnectionLost.bind(app));

  const restored = await persist.restore(app, state);
  if (restored) {
    persist.setFileName(
      app._persistedFileName || state.fileName,
      app._persistedFilePath || state.filePath || state.fileName,
    );
    app.updateSessionDisplay(persist.getSessionIdFromURL() || app.api.sessionId);
    app.controls.setStatus(`Loaded: ${state.fileName}`, 'success');
  } else {
    app.controls.setStatus('Restore failed', 'error');
  }
}
