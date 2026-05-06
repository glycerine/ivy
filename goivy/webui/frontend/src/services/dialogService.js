export function bridgeDialog(config, bridge = globalThis.window && globalThis.window.__ivyVueBridge) {
  if (bridge && typeof bridge.showDialog === 'function') {
    return bridge.showDialog(config);
  }
  return undefined;
}

export function okDialog(title, message, bridge) {
  return bridgeDialog({ type: 'ok', title, message }, bridge);
}

export function okCancelDialog(title, message, bridge) {
  return bridgeDialog({ type: 'okCancel', title, message }, bridge);
}

export function textDialog(title, message, text, options, bridge) {
  return bridgeDialog({ type: 'text', title, message, text, options: options || {} }, bridge);
}

export function entryDialog(title, message, initialValue, options, bridge) {
  return bridgeDialog({ type: 'entry', title, message, initialValue, options: options || {} }, bridge);
}

export function integerDialog(title, message, initialValue, options, bridge) {
  return bridgeDialog({ type: 'integer', title, message, initialValue, options: options || {} }, bridge);
}

export function listboxDialog(title, message, items, options, bridge) {
  return bridgeDialog({ type: 'listbox', title, message, items: items || [], options: options || {} }, bridge);
}

export function buttonListDialog(title, message, buttons, bridge) {
  return bridgeDialog({ type: 'buttons', title, message, buttons: buttons || [] }, bridge);
}
