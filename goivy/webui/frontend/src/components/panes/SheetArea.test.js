import { mount } from '@vue/test-utils';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import SheetArea from './SheetArea.vue';
import { useLayoutStore } from '../../stores/layoutStore.js';
import { installAppServices, resetAppServicesForTests } from '../../services/appServices.js';

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

describe('SheetArea resizers', () => {
  let refreshLayout;

  beforeEach(() => {
    setActivePinia(createPinia());
    refreshLayout = vi.fn();
    installAppServices({ refreshLayout });
  });

  afterEach(() => {
    document.body.innerHTML = '';
    resetAppServicesForTests();
  });

  it('owns the ARG/concept divider through the layout store', async () => {
    const wrapper = mount(SheetArea, { attachTo: document.body });
    const layout = useLayoutStore();
    const divider = wrapper.find('#divider').element;
    const argPanel = wrapper.find('#arg-panel').element;
    setSize(divider.parentElement, { width: 900 });
    setSize(argPanel, { width: 300 });

    divider.dispatchEvent(mouse('mousedown', { x: 300 }));
    document.dispatchEvent(mouse('mousemove', { x: 380 }));
    document.dispatchEvent(mouse('mouseup', { x: 380 }));

    expect(layout.argPanelWidth).toBe(380);
    expect(argPanel.style.flex).toBe('');
    expect(refreshLayout).toHaveBeenCalled();
  });

  it('owns the state/relation divider through the layout store', () => {
    const wrapper = mount(SheetArea, { attachTo: document.body });
    const layout = useLayoutStore();
    const divider = wrapper.find('#divider2').element;
    const statePanel = wrapper.find('#state-panel').element;
    setSize(divider.parentElement, { width: 1000 });
    setSize(statePanel, { width: 240 });

    divider.dispatchEvent(mouse('mousedown', { x: 700 }));
    document.dispatchEvent(mouse('mousemove', { x: 620 }));
    document.dispatchEvent(mouse('mouseup', { x: 620 }));

    expect(layout.statePanelWidth).toBe(320);
    expect(statePanel.style.flex).toBe('');
  });

  it('owns the root Details splitter through the layout store', () => {
    const wrapper = mount(SheetArea, { attachTo: document.body });
    const layout = useLayoutStore();
    const header = wrapper.find('#info-header').element;
    const panel = wrapper.find('#info-panel').element;
    const sheetLeft = wrapper.find('.sheet-left').element;
    setSize(panel, { height: 120 });
    setSize(sheetLeft, { height: 600 });

    header.dispatchEvent(mouse('mousedown', { y: 500 }));
    document.dispatchEvent(mouse('mousemove', { y: 440 }));
    document.dispatchEvent(mouse('mouseup', { y: 440 }));

    expect(layout.detailsHeight).toBe(180);
    expect(panel.style.flex).toBe('');
  });
});
