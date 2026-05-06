import { nextTick } from 'vue';

function legacyApp() {
  return globalThis.window && globalThis.window.ivyApp;
}

function refreshNow() {
  const app = legacyApp();
  if (!app) return;
  if (typeof app._refreshGraphsAndEditorLayout === 'function') {
    app._refreshGraphsAndEditorLayout();
    return;
  }
  if (app.argGraph && typeof app.argGraph.resize === 'function') app.argGraph.resize();
  if (app.conceptGraph && typeof app.conceptGraph.resize === 'function') app.conceptGraph.resize();
}

export function scheduleLegacyLayoutRefresh() {
  refreshNow();
  nextTick(refreshNow);
  const win = globalThis.window;
  if (win && typeof win.requestAnimationFrame === 'function') {
    win.requestAnimationFrame(refreshNow);
  }
}

export function setGraphPointerEvents(root, value) {
  if (!root || typeof root.querySelectorAll !== 'function') return;
  root.querySelectorAll('.graph-container').forEach((graph) => {
    graph.style.pointerEvents = value;
  });
}

export function startMouseDrag(event, {
  cursor,
  onMove,
  onEnd,
  onStart,
} = {}) {
  const doc = event.currentTarget ? event.currentTarget.ownerDocument : globalThis.document;
  if (!doc) return;
  if (typeof event.preventDefault === 'function') event.preventDefault();
  if (typeof onStart === 'function') onStart();
  doc.body.style.cursor = cursor || '';
  doc.body.style.userSelect = 'none';

  const move = (moveEvent) => {
    if (typeof onMove === 'function') onMove(moveEvent);
  };
  const up = (upEvent) => {
    doc.removeEventListener('mousemove', move);
    doc.removeEventListener('mouseup', up);
    doc.body.style.cursor = '';
    doc.body.style.userSelect = '';
    if (typeof onEnd === 'function') onEnd(upEvent);
  };

  doc.addEventListener('mousemove', move);
  doc.addEventListener('mouseup', up);
}
