import { mount } from '@vue/test-utils';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import FileInputHost from './FileInputHost.vue';

function makeFile(name) {
  return new File(['content'], name, { type: 'text/plain' });
}

describe('FileInputHost', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    window.ivyApp = undefined;
  });

  it('routes hidden file input changes through the app bridge', async () => {
    const loadFile = vi.fn(async () => undefined);
    const loadEventTraceFile = vi.fn(async () => undefined);
    const loadAnalysisStateFile = vi.fn(async () => undefined);
    window.ivyApp = { loadFile, loadEventTraceFile, loadAnalysisStateFile };

    const wrapper = mount(FileInputHost);

    const modelInput = wrapper.find('#file-input').element;
    Object.defineProperty(modelInput, 'files', { configurable: true, value: [makeFile('model.ivy')] });
    await wrapper.find('#file-input').trigger('change');
    expect(loadFile).toHaveBeenCalledWith(modelInput.files[0]);
    expect(modelInput.value).toBe('');

    const eventInput = wrapper.find('#event-file-input').element;
    Object.defineProperty(eventInput, 'files', { configurable: true, value: [makeFile('trace.iev')] });
    await wrapper.find('#event-file-input').trigger('change');
    expect(loadEventTraceFile).toHaveBeenCalledWith(eventInput.files[0]);

    const stateInput = wrapper.find('#analysis-state-file-input').element;
    Object.defineProperty(stateInput, 'files', { configurable: true, value: [makeFile('state.ivyweb.json')] });
    await wrapper.find('#analysis-state-file-input').trigger('change');
    expect(loadAnalysisStateFile).toHaveBeenCalledWith(stateInput.files[0]);
  });
});
