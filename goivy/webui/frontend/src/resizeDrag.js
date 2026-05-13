import { currentAppServices } from './services/appServices.js';

function refreshNow() {
  currentAppServices().refreshLayout();
}

export function scheduleLayoutRefresh() {
  refreshNow();
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
