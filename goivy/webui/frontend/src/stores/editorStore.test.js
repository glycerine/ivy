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

  it('validates editor keymap choices', () => {
    const editor = useEditorStore();

    editor.setKeymap('vim');
    expect(editor.keymap).toBe('vim');

    editor.setKeymap('unknown');
    expect(editor.keymap).toBe('sublime');
  });

  it('tracks the reopen-last-file affordance', () => {
    const editor = useEditorStore();

    editor.setReopenLastFileButton(true, 'Re-open last file client.ivy');
    expect(editor.reopenLastVisible).toBe(true);
    expect(editor.reopenLastLabel).toBe('Re-open last file client.ivy');

    editor.setReopenLastFileButton(false, 'ignored');
    expect(editor.reopenLastVisible).toBe(false);
    expect(editor.reopenLastLabel).toBe('');
  });
});
