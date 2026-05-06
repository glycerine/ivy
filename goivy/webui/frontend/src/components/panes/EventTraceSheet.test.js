import { mount } from '@vue/test-utils';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import EventTraceSheet from './EventTraceSheet.vue';
import { useEventTraceStore } from '../../stores/eventTraceStore.js';
import { registerCommand, resetCommandRegistry } from '../../services/commandRegistry.js';

describe('EventTraceSheet', () => {
  let pinia;

  beforeEach(() => {
    pinia = createPinia();
    setActivePinia(pinia);
    resetCommandRegistry();
  });

  afterEach(() => {
    resetCommandRegistry();
  });

  it('reflects the selected pattern index from Pinia in the listbox', async () => {
    const traces = useEventTraceStore();
    traces.upsertSheet({
      id: 'events-1',
      events: [],
      patterns: ['root(a)', 'child(a)'],
    });
    traces.setSelectedPatternIndex('events-1', 1);

    const wrapper = mount(EventTraceSheet, {
      props: { sheetId: 'events-1' },
      global: { plugins: [pinia] },
    });
    const select = wrapper.find('.event-pattern-list');

    expect(select.element.value).toBe('child(a)');

    await select.setValue('root(a)');

    expect(traces.selectedPatternIndex('events-1')).toBe(0);
    expect(traces.selectedPattern('events-1')).toBe('root(a)');
  });

  it('routes event row selection and expansion through registered commands when present', async () => {
    const traces = useEventTraceStore();
    traces.upsertSheet({
      id: 'events-1',
      events: [{ address: '0', text: 'root(a)', subs: [{ address: '0/0', text: 'child(a)' }] }],
      patterns: [],
    });
    const selectEventTraceRow = vi.fn();
    const toggleEventTraceNode = vi.fn();
    registerCommand('selectEventTraceRow', selectEventTraceRow);
    registerCommand('toggleEventTraceNode', toggleEventTraceNode);

    const wrapper = mount(EventTraceSheet, {
      props: { sheetId: 'events-1' },
      global: { plugins: [pinia] },
    });

    await wrapper.find('[data-event-address="0"]').trigger('click');
    await wrapper.find('[data-event-toggle="0"]').trigger('click');

    expect(selectEventTraceRow).toHaveBeenCalledWith('events-1', '0');
    expect(toggleEventTraceNode).toHaveBeenCalledWith('events-1', '0');
    expect(traces.sheetById('events-1').selectedEventAddress).toBe('');
  });

  it('selects rows and lazily expands children through Pinia without the app bridge', async () => {
    const traces = useEventTraceStore();
    traces.upsertSheet({
      id: 'events-1',
      events: [{ address: '0', text: 'root(a)', subs: [{ address: '0/0', text: 'child(a)' }] }],
      patterns: [],
    });

    const wrapper = mount(EventTraceSheet, {
      props: { sheetId: 'events-1' },
      global: { plugins: [pinia] },
    });

    expect(wrapper.find('[data-event-address="0/0"]').exists()).toBe(false);

    await wrapper.find('[data-event-toggle="0"]').trigger('click');
    expect(wrapper.find('[data-event-address="0/0"]').text()).toContain('child(a)');

    await wrapper.find('[data-event-address="0/0"]').trigger('click');
    expect(traces.sheetById('events-1').selectedEventAddress).toBe('0/0');
    expect(wrapper.find('[data-event-address="0/0"]').classes()).toContain('selected');
  });
});
