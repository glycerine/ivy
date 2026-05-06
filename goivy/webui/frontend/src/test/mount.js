import { mount } from '@vue/test-utils';
import { getActivePinia } from 'pinia';
import { createFreshPinia } from './pinia.js';

export function mountWithPinia(component, options = {}) {
  const pinia = options.pinia || getActivePinia() || createFreshPinia();
  return mount(component, {
    ...options,
    global: {
      ...(options.global || {}),
      plugins: [...((options.global && options.global.plugins) || []), pinia],
    },
  });
}
