import { mount } from '@vue/test-utils';
import { beforeEach, describe, expect, it } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import DialogHost from './DialogHost.vue';
import { useDialogStore } from '../stores/dialogStore.js';

describe('DialogHost', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('renders an active dialog and resolves button choices through the store', async () => {
    const dialogStore = useDialogStore();
    const promise = dialogStore.open({
      type: 'buttons',
      title: 'Confirm',
      message: 'Pick one',
      buttons: [
        { label: 'Cancel', value: 'cancel' },
        { label: 'Use it', value: 'use' },
      ],
    });

    const wrapper = mount(DialogHost);

    expect(wrapper.text()).toContain('Confirm');
    expect(wrapper.text()).toContain('Pick one');
    await wrapper.findAll('[data-ivy-dialog-button]')[1].trigger('click');

    await expect(promise).resolves.toBe('use');
  });
});
