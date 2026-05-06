import { createPinia, setActivePinia } from 'pinia';

export function createFreshPinia() {
  const pinia = createPinia();
  setActivePinia(pinia);
  return pinia;
}

export function resetPinia() {
  return createFreshPinia();
}
