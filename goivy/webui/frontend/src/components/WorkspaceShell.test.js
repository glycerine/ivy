import { mount } from '@vue/test-utils';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import WorkspaceShell from './WorkspaceShell.vue';
import { useLayoutStore } from '../stores/layoutStore.js';

function setSize(el, { width, height } = {}) {
  if (width != null) {
    Object.defineProperty(el, 'offsetWidth', { configurable: true, value: width });
  }
  if (height != null) {
    Object.defineProperty(el, 'offsetHeight', { configurable: true, value: height });
  }
}

function mouse(type, { x = 0, y = 0 } = {}) {
  return new MouseEvent(type, {
    clientX: x,
    clientY: y,
    bubbles: true,
    cancelable: true,
  });
}

describe('WorkspaceShell resizers', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    window.ivyApp = {
      _refreshGraphsAndEditorLayout: vi.fn(),
    };
  });

  afterEach(() => {
    document.body.innerHTML = '';
    window.ivyApp = undefined;
  });

  it('owns the editor divider through the layout store', () => {
    const wrapper = mount(WorkspaceShell, { attachTo: document.body });
    const layout = useLayoutStore();
    const topRow = wrapper.find('#top-row').element;
    const sheetArea = wrapper.find('#sheet-area').element;
    const divider = wrapper.find('#divider3').element;
    const editorPanel = wrapper.find('#editor-panel').element;
    setSize(topRow, { width: 1200 });
    setSize(divider, { width: 4 });
    setSize(editorPanel, { width: 440 });

    divider.dispatchEvent(mouse('mousedown', { x: 760 }));
    document.dispatchEvent(mouse('mousemove', { x: 620 }));
    document.dispatchEvent(mouse('mouseup', { x: 620 }));

    expect(sheetArea.style.flex).toBe('1 1 auto');
    expect(layout.editorWidth).toBe(580);
    expect(editorPanel.style.flex).toBe('');
    expect(window.ivyApp._refreshGraphsAndEditorLayout).toHaveBeenCalled();
  });

  it('owns the tutorial divider through the layout store', () => {
    const wrapper = mount(WorkspaceShell, { attachTo: document.body });
    const layout = useLayoutStore();
    const outer = wrapper.find('#outer-container').element;
    const tutorial = wrapper.find('#tutorial-container').element;
    const divider = wrapper.find('#divider-h').element;
    setSize(outer, { height: 700 });
    setSize(tutorial, { height: 180 });

    divider.dispatchEvent(mouse('mousedown', { y: 500 }));
    document.dispatchEvent(mouse('mousemove', { y: 460 }));
    document.dispatchEvent(mouse('mouseup', { y: 460 }));

    expect(layout.tutorialHeight).toBe(220);
  });
});
