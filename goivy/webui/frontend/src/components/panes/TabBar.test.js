import { mount } from '@vue/test-utils';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import TabBar from './TabBar.vue';
import { useSheetStore } from '../../stores/sheetStore.js';

describe('TabBar', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    window.ivyApp = undefined;
  });

  it('routes tab activation and close clicks through the app bridge', async () => {
    const switchSheet = vi.fn();
    const removeSheet = vi.fn();
    window.ivyApp = { switchSheet, removeSheet };
    const sheets = useSheetStore();
    sheets.upsertTab({ id: 'sheet-2', label: 'Trace', closable: true });

    const wrapper = mount(TabBar);

    await wrapper.find('[data-sheet="sheet-2"]').trigger('click');
    expect(switchSheet).toHaveBeenCalledWith('sheet-2');

    await wrapper.find('[data-sheet="sheet-2"] .tab-close').trigger('click');
    expect(removeSheet).toHaveBeenCalledWith('sheet-2');
    expect(switchSheet).toHaveBeenCalledTimes(1);
  });
});
