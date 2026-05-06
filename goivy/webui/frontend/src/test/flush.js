import { nextTick } from 'vue';

export async function flushPromises() {
  await Promise.resolve();
  await nextTick();
}

export { nextTick };
