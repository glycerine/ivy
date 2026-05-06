import { mount } from '@vue/test-utils';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import TabBar from './TabBar.vue';
import { useSheetStore } from '../../stores/sheetStore.js';
import { registerCommand, resetCommandRegistry } from '../../services/commandRegistry.js';

describe('TabBar', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    resetCommandRegistry();
  });

  it('routes tab activation and close clicks through registered commands', async () => {
    const switchSheet = vi.fn();
    const removeSheet = vi.fn();
    registerCommand('switchSheet', switchSheet);
    registerCommand('removeSheet', removeSheet);
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
