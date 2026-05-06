import { mount } from '@vue/test-utils';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import TutorialPane from './TutorialPane.vue';
import { DEFAULT_TUTORIAL_URL, useLayoutStore } from '../../stores/layoutStore.js';

describe('TutorialPane', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    window.ivyApp = undefined;
  });

  it('renders tutorial navigation from the layout store', async () => {
    const wrapper = mount(TutorialPane);
    const layout = useLayoutStore();
    const input = wrapper.find('#tutorial-url');

    expect(input.element.value).toBe(DEFAULT_TUTORIAL_URL);
    expect(wrapper.find('#tutorial-iframe').attributes('src')).toBe(DEFAULT_TUTORIAL_URL);
    expect(wrapper.find('#tutorial-back').attributes('disabled')).toBeDefined();

    await input.setValue('/static/tutorial/page-2.html');
    await input.trigger('keydown', { key: 'Enter' });

    expect(layout.tutorialUrl).toBe('/static/tutorial/page-2.html');
    expect(wrapper.find('#tutorial-iframe').attributes('src')).toBe('/static/tutorial/page-2.html');
    expect(wrapper.find('#tutorial-back').attributes('disabled')).toBeUndefined();

    await wrapper.find('#tutorial-back').trigger('click');
    expect(layout.tutorialUrl).toBe(DEFAULT_TUTORIAL_URL);
  });

  it('routes the close button through the legacy app when available', async () => {
    const toggleTutorial = vi.fn();
    window.ivyApp = { toggleTutorial };

    const wrapper = mount(TutorialPane);
    await wrapper.find('#tutorial-close').trigger('click');

    expect(toggleTutorial).toHaveBeenCalledWith(true);
  });
});
