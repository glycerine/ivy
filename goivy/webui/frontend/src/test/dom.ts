export function installEditorDom() {
  document.body.innerHTML = [
    '<div id="save-as-explain-notice" style="display: none"></div>',
    '<div id="model-editor-label"></div>',
    '<button id="file-reopen-last"></button>',
    '<span id="loaded-file"></span>',
    '<select id="mode-select"><option value="concrete" selected>concrete</option></select>',
    '<tbody id="state-checkbox-body"></tbody>',
    '<div id="file-recent-list"></div>',
  ].join('');
}

export function clearDom() {
  document.body.innerHTML = '';
}
