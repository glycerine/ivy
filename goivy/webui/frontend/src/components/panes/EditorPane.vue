<script setup>
import { useEditorStore } from '../../stores/editorStore.js';
import { useLayoutStore } from '../../stores/layoutStore.js';
import { callApp, runCommand } from '../legacyCommand.js';

const editorStore = useEditorStore();
const layoutStore = useLayoutStore();

function setKeymap(keymap) {
  editorStore.setKeymap(keymap);
  callApp('setEditorKeymap', keymap);
}
</script>

<template>
  <div id="editor-panel" :style="layoutStore.editorPanelStyle">
    <div class="panel-header">
      <div class="editor-title-row">
        <strong class="column-title">Editing:</strong>
        <span id="model-editor-label" class="editor-path" :title="`Editing: ${editorStore.label}`">{{ editorStore.label }}</span>
      </div>
      <div class="panel-header-actions">
        <button id="file-close-current" class="editor-close-btn" title="Close current file" @click="runCommand($event, () => callApp('closeCurrentFile'))">x</button>
        <button
          id="file-reopen-last"
          class="reopen-last-btn"
          :style="{ display: editorStore.reopenLastVisible ? '' : 'none' }"
          @click="runCommand($event, () => callApp('reopenLastFile'))"
        >
          {{ editorStore.reopenLastLabel }}
        </button>
        <span class="keymap-radios">
          <label><input type="radio" name="keymap" value="sublime" :checked="editorStore.keymap === 'sublime'" @change="setKeymap('sublime')"> Sublime</label>
          <label><input type="radio" name="keymap" value="emacs" :checked="editorStore.keymap === 'emacs'" @change="setKeymap('emacs')"> Emacs-ish</label>
          <label><input type="radio" name="keymap" value="vim" :checked="editorStore.keymap === 'vim'" @change="setKeymap('vim')"> Vim</label>
          <a class="keymap-docs-link" href="https://codemirror.net/5/doc/manual.html#keymaps" target="_blank" rel="noopener noreferrer">keymap docs</a>
        </span>
      </div>
    </div>
    <div class="editor-text-shell">
      <textarea id="model-editor" class="model-editor-text" spellcheck="false" placeholder="Load an .ivy file to see its source here..."></textarea>
      <div v-if="editorStore.saveState === 'saving'" class="ivy-save-editor-sheen" aria-hidden="true"></div>
    </div>
  </div>
</template>
