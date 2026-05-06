import { defineStore } from 'pinia';

export const useEditorStore = defineStore('editor', {
  state: () => ({
    path: '',
    content: '',
    savedContent: '',
    saveState: 'idle',
    hasFile: false,
    keymap: 'sublime',
  }),
  getters: {
    dirty: (state) => state.content !== state.savedContent,
    displayName: (state) => state.path || '(unsaved file)',
    label(state) {
      const base = state.path || '(unsaved file)';
      if (!state.path) {
        if (state.saveState === 'saving') return `${base} [saving...]`;
        return state.content !== state.savedContent ? `** ${base}` : base;
      }
      if (state.saveState === 'saving') return `${base} [saving...]`;
      if (state.content !== state.savedContent) return `** ${base}`;
      return `${base} [saved]`;
    },
  },
  actions: {
    load({ path = '', content = '' } = {}) {
      this.path = path;
      this.content = content;
      this.savedContent = content;
      this.hasFile = Boolean(path);
      this.saveState = 'idle';
    },
    edit(content) {
      this.content = content;
      if (this.saveState === 'saved') {
        this.saveState = 'idle';
      }
    },
    markSaving() {
      this.saveState = 'saving';
    },
    markSaved(content = this.content) {
      this.savedContent = content;
      this.saveState = 'saved';
    },
    markSaveFailed() {
      this.saveState = 'idle';
    },
    applyLegacySnapshot({
      path = '',
      content = '',
      savedContent = '',
      saveInProgress = false,
    } = {}) {
      this.path = path;
      this.content = content;
      this.savedContent = savedContent;
      this.hasFile = Boolean(path);
      this.saveState = saveInProgress ? 'saving' : 'idle';
    },
  },
});
