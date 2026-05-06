import { beforeEach, describe, expect, it } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { useLayoutStore } from './layoutStore.js';

describe('layoutStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('tracks tutorial visibility and constrained pane sizes', () => {
    const layout = useLayoutStore();

    layout.setTutorialVisible(false);
    layout.setDetailsHeight(40);
    layout.setEditorWidth(100);

    expect(layout.tutorialVisible).toBe(false);
    expect(layout.detailsHeight).toBe(80);
    expect(layout.editorWidth).toBe(200);
  });
});
