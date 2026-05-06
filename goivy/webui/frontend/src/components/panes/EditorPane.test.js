import { mount } from '@vue/test-utils';
import { beforeEach, describe, expect, it } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import EditorPane from './EditorPane.vue';
import { useEditorStore } from '../../stores/editorStore.js';

describe('EditorPane', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('renders the save sheen from editor save state', () => {
    const editor = useEditorStore();
    editor.load({ path: 'client.ivy', content: 'old' });
    editor.edit('new');
    editor.markSaving();

    const wrapper = mount(EditorPane);

    expect(wrapper.text()).toContain('client.ivy [saving...]');
    expect(wrapper.find('.ivy-save-editor-sheen').exists()).toBe(true);
  });
});
