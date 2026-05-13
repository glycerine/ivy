<script setup>
import { useLayoutStore } from '../../stores/layoutStore.js';
import { runCommand, runUiCommand } from '../../services/uiCommandService.js';

const layoutStore = useLayoutStore();

function setKeymap(keymap) {
  runUiCommand('setEditorKeymap', keymap);
}
</script>

<template>
  <div id="editor-panel" :style="layoutStore.editorPanelStyle">
    <div class="panel-header">
      <div class="editor-title-row">
        <strong class="column-title">Editing:</strong>
        <span id="model-editor-label" class="editor-path" title="Editing: (unsaved file)">(unsaved file)</span>
      </div>
      <div class="panel-header-actions">
        <button id="file-close-current" class="editor-close-btn" title="Close current file" @click="runCommand($event, () => runUiCommand('file.close'))">x</button>
        <button
          id="file-reopen-last"
          class="reopen-last-btn"
          style="display: none;"
          @click="runCommand($event, () => runUiCommand('file.reopenLast'))"
        >
          Re-open last file
        </button>
        <span class="keymap-radios">
          <label><input type="radio" name="keymap" value="sublime" checked @change="setKeymap('sublime')"> Sublime</label>
          <label><input type="radio" name="keymap" value="emacs" @change="setKeymap('emacs')"> Emacs-ish</label>
          <label><input type="radio" name="keymap" value="vim" @change="setKeymap('vim')"> Vim</label>
          <a class="keymap-docs-link" href="https://codemirror.net/5/doc/manual.html#keymaps" target="_blank" rel="noopener noreferrer">keymap docs</a>
        </span>
      </div>
    </div>
    <div class="editor-text-shell">
      <textarea id="model-editor" class="model-editor-text" spellcheck="false" placeholder="Load an .ivy file to see its source here..."></textarea>
    </div>
  </div>
</template>
