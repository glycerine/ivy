import { beforeEach, describe, expect, it } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { useEditorStore } from './editorStore.js';

describe('editorStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('shows dirty, saving, and saved labels without mixing states', () => {
    const editor = useEditorStore();
    editor.load({ path: 'client_server_example.ivy', content: 'saved' });

    expect(editor.label).toBe('client_server_example.ivy [saved]');

    editor.edit('changed');
    expect(editor.label).toBe('** client_server_example.ivy');

    editor.markSaving();
    expect(editor.label).toBe('client_server_example.ivy [saving...]');

    editor.markSaved('changed');
    expect(editor.label).toBe('client_server_example.ivy [saved]');
  });

  it('matches the legacy unsaved-file label rules', () => {
    const editor = useEditorStore();

    editor.applyLegacySnapshot({ path: '', content: '', savedContent: '' });
    expect(editor.label).toBe('(unsaved file)');

    editor.applyLegacySnapshot({ path: '', content: 'x', savedContent: '' });
    expect(editor.label).toBe('** (unsaved file)');

    editor.applyLegacySnapshot({ path: '', content: 'x', savedContent: '', saveInProgress: true });
    expect(editor.label).toBe('(unsaved file) [saving...]');
  });
});
