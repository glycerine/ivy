import { mount } from '@vue/test-utils';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import EditorPane from './EditorPane.vue';
import { useEditorStore } from '../../stores/editorStore.js';

describe('EditorPane', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    window.ivyApp = undefined;
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

  it('updates the keymap store and CodeMirror bridge from radios', async () => {
    const setEditorKeymap = vi.fn();
    window.ivyApp = { setEditorKeymap };
    const editor = useEditorStore();
    const wrapper = mount(EditorPane);

    await wrapper.find('input[value="emacs"]').setValue();

    expect(editor.keymap).toBe('emacs');
    expect(setEditorKeymap).toHaveBeenCalledWith('emacs');
  });

  it('renders the reopen-last-file button from editor state', () => {
    const editor = useEditorStore();
    editor.setReopenLastFileButton(true, 'Re-open last file client.ivy');

    const wrapper = mount(EditorPane);
    const button = wrapper.find('#file-reopen-last');

    expect(button.isVisible()).toBe(true);
    expect(button.text()).toBe('Re-open last file client.ivy');
  });
});
