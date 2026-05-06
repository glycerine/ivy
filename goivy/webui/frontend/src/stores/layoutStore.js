import { defineStore } from 'pinia';

export const DEFAULT_TUTORIAL_URL = '/static/tutorial/kenmcmil.github.io/ivy/language.html';

function normalizeTutorialUrl(rawUrl) {
  const trimmed = String(rawUrl || '').trim();
  if (!trimmed) return '';
  if (/^https?:\/\//i.test(trimmed) || trimmed.startsWith('/')) return trimmed;
  return `https://${trimmed}`;
}

function cleanTutorialFrameUrl(rawUrl) {
  const value = String(rawUrl || '').trim();
  if (!value || value === 'about:blank') return '';
  try {
    const parsed = new URL(value);
    if (typeof window !== 'undefined' && parsed.origin === window.location.origin) {
      return `${parsed.pathname}${parsed.search}${parsed.hash}`;
    }
    return value;
  } catch (e) {
    return value;
  }
}

export const useLayoutStore = defineStore('layout', {
  state: () => ({
    tutorialVisible: true,
    tutorialUrl: DEFAULT_TUTORIAL_URL,
    tutorialInput: DEFAULT_TUTORIAL_URL,
    tutorialHistory: [DEFAULT_TUTORIAL_URL],
    tutorialHistoryIndex: 0,
    tutorialFrameKey: 0,
    tutorialButtonFlashing: false,
    argPanelWidth: null,
    statePanelWidth: null,
    detailsHeight: null,
    editorWidth: null,
    tutorialHeight: null,
  }),
  getters: {
    canGoBack: (state) => state.tutorialHistoryIndex > 0,
    canGoForward: (state) => state.tutorialHistoryIndex < state.tutorialHistory.length - 1,
    argPanelStyle: (state) => (
      state.argPanelWidth ? { flex: `0 0 ${state.argPanelWidth}px` } : {}
    ),
    statePanelStyle: (state) => (
      state.statePanelWidth ? { flex: `0 0 ${state.statePanelWidth}px` } : {}
    ),
    detailsPanelStyle: (state) => (
      state.detailsHeight ? { flex: `0 0 ${state.detailsHeight}px`, height: `${state.detailsHeight}px` } : {}
    ),
    editorPanelStyle: (state) => (
      state.editorWidth ? { flex: `0 0 ${state.editorWidth}px` } : {}
    ),
    tutorialPanelStyle: (state) => ({
      display: state.tutorialVisible ? '' : 'none',
      ...(state.tutorialHeight ? { flex: `0 0 ${state.tutorialHeight}px` } : {}),
    }),
  },
  actions: {
    setTutorialVisible(visible) {
      this.tutorialVisible = Boolean(visible);
    },
    flashTutorialButton(durationMs = 1200) {
      this.tutorialButtonFlashing = true;
      window.setTimeout(() => {
        this.tutorialButtonFlashing = false;
      }, Number(durationMs) || 1200);
    },
    setTutorialInput(value) {
      this.tutorialInput = String(value || '');
    },
    navigateTutorial(rawUrl) {
      const url = normalizeTutorialUrl(rawUrl);
      if (!url) return;
      if (this.tutorialHistoryIndex < this.tutorialHistory.length - 1) {
        this.tutorialHistory = this.tutorialHistory.slice(0, this.tutorialHistoryIndex + 1);
      }
      if (this.tutorialHistory[this.tutorialHistoryIndex] !== url) {
        this.tutorialHistory.push(url);
        this.tutorialHistoryIndex = this.tutorialHistory.length - 1;
      }
      this.tutorialInput = url;
      this.tutorialUrl = url;
    },
    goTutorialBack() {
      if (!this.canGoBack) return;
      this.tutorialHistoryIndex -= 1;
      const url = this.tutorialHistory[this.tutorialHistoryIndex];
      this.tutorialInput = url;
      this.tutorialUrl = url;
      this.tutorialFrameKey += 1;
    },
    goTutorialForward() {
      if (!this.canGoForward) return;
      this.tutorialHistoryIndex += 1;
      const url = this.tutorialHistory[this.tutorialHistoryIndex];
      this.tutorialInput = url;
      this.tutorialUrl = url;
      this.tutorialFrameKey += 1;
    },
    reloadTutorial() {
      const url = this.tutorialHistory[this.tutorialHistoryIndex] || this.tutorialUrl;
      if (!url) return;
      this.tutorialInput = url;
      this.tutorialUrl = url;
      this.tutorialFrameKey += 1;
    },
    recordTutorialLoad(rawUrl) {
      const url = cleanTutorialFrameUrl(rawUrl);
      if (!url) return;
      if (this.tutorialHistory[this.tutorialHistoryIndex] !== url) {
        if (this.tutorialHistoryIndex < this.tutorialHistory.length - 1) {
          this.tutorialHistory = this.tutorialHistory.slice(0, this.tutorialHistoryIndex + 1);
        }
        this.tutorialHistory.push(url);
        this.tutorialHistoryIndex = this.tutorialHistory.length - 1;
      }
      this.tutorialInput = url;
      this.tutorialUrl = url;
    },
    setArgPanelWidth(width) {
      this.argPanelWidth = Math.max(150, Number(width) || 150);
    },
    setStatePanelWidth(width) {
      this.statePanelWidth = Math.max(200, Number(width) || 200);
    },
    setDetailsHeight(height) {
      this.detailsHeight = Math.max(72, Number(height) || 72);
    },
    setEditorWidth(width) {
      this.editorWidth = Math.max(200, Number(width) || 200);
    },
    setTutorialHeight(height) {
      this.tutorialHeight = Math.max(80, Number(height) || 80);
    },
  },
});
