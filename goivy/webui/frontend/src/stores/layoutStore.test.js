import { beforeEach, describe, expect, it } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { DEFAULT_TUTORIAL_URL, useLayoutStore } from './layoutStore.js';

describe('layoutStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('tracks tutorial visibility and constrained pane sizes', () => {
    const layout = useLayoutStore();

    layout.setTutorialVisible(false);
    layout.setArgPanelWidth(100);
    layout.setStatePanelWidth(100);
    layout.setDetailsHeight(40);
    layout.setEditorWidth(100);
    layout.setTutorialHeight(40);

    expect(layout.tutorialVisible).toBe(false);
    expect(layout.argPanelWidth).toBe(150);
    expect(layout.statePanelWidth).toBe(200);
    expect(layout.detailsHeight).toBe(72);
    expect(layout.editorWidth).toBe(200);
    expect(layout.tutorialHeight).toBe(80);
    expect(layout.argPanelStyle).toEqual({ flex: '0 0 150px' });
    expect(layout.statePanelStyle).toEqual({ flex: '0 0 200px' });
    expect(layout.detailsPanelStyle).toEqual({ flex: '0 0 72px', height: '72px' });
    expect(layout.editorPanelStyle).toEqual({ flex: '0 0 200px' });
    expect(layout.tutorialPanelStyle).toEqual({ display: 'none', flex: '0 0 80px' });
  });

  it('keeps tutorial navigation history in Pinia', () => {
    const layout = useLayoutStore();

    layout.navigateTutorial('/static/tutorial/page-2.html');
    layout.navigateTutorial('example.com/docs');

    expect(layout.tutorialUrl).toBe('https://example.com/docs');
    expect(layout.tutorialInput).toBe('https://example.com/docs');
    expect(layout.canGoBack).toBe(true);
    expect(layout.canGoForward).toBe(false);

    layout.goTutorialBack();
    expect(layout.tutorialUrl).toBe('/static/tutorial/page-2.html');
    expect(layout.canGoForward).toBe(true);

    layout.navigateTutorial('/static/tutorial/page-3.html');
    expect(layout.tutorialHistory).toEqual([
      DEFAULT_TUTORIAL_URL,
      '/static/tutorial/page-2.html',
      '/static/tutorial/page-3.html',
    ]);
    expect(layout.canGoForward).toBe(false);
  });

  it('records same-origin iframe loads as clean paths', () => {
    const layout = useLayoutStore();

    layout.recordTutorialLoad(`${window.location.origin}/static/tutorial/linked.html#section`);

    expect(layout.tutorialUrl).toBe('/static/tutorial/linked.html#section');
    expect(layout.tutorialInput).toBe('/static/tutorial/linked.html#section');
  });
});
