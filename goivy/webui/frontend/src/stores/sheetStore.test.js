import { beforeEach, describe, expect, it } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { useSheetStore } from './sheetStore.js';

describe('sheetStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('tracks tabs, active sheet, and root sheet protection', () => {
    const sheets = useSheetStore();

    sheets.upsertTab({ id: 'sheet-2', label: 'Trace', type: 'analysis' });
    sheets.activateTab('sheet-2');

    expect(sheets.tabs.map((tab) => tab.id)).toEqual(['sheet-1', 'sheet-2']);
    expect(sheets.tabs[1]).toMatchObject({ label: 'Trace', closable: true });
    expect(sheets.activeSheetId).toBe('sheet-2');

    sheets.removeTab('sheet-1');
    expect(sheets.tabs.map((tab) => tab.id)).toEqual(['sheet-1', 'sheet-2']);

    sheets.removeTab('sheet-2');
    expect(sheets.tabs.map((tab) => tab.id)).toEqual(['sheet-1']);
    expect(sheets.activeSheetId).toBe('sheet-1');
  });
});
