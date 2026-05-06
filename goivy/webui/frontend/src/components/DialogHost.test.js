import { beforeEach, describe, expect, it } from 'vitest';
import DialogHost from './DialogHost.vue';
import { useDialogStore } from '../stores/dialogStore.js';
import { createFreshPinia } from '../test/pinia.js';
import { mountWithPinia } from '../test/mount.js';

describe('DialogHost', () => {
  beforeEach(() => {
    createFreshPinia();
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

    const wrapper = mountWithPinia(DialogHost);

    expect(wrapper.text()).toContain('Confirm');
    expect(wrapper.text()).toContain('Pick one');
    await wrapper.findAll('[data-ivy-dialog-button]')[1].trigger('click');

    await expect(promise).resolves.toBe('use');
  });

  it('resolves ok and ok-cancel dialogs through rendered buttons', async () => {
    const dialogStore = useDialogStore();
    const wrapper = mountWithPinia(DialogHost);

    const ok = dialogStore.open({ type: 'ok', title: 'ivyweb', message: 'hello' });
    await wrapper.vm.$nextTick();
    expect(wrapper.find('[data-ivy-dialog]').exists()).toBe(true);
    expect(wrapper.find('.dialog-title').text()).toBe('ivyweb');
    await wrapper.find('[data-ivy-dialog-button]').trigger('click');
    await expect(ok).resolves.toBe(true);
    expect(wrapper.find('[data-ivy-dialog]').exists()).toBe(false);

    const cancel = dialogStore.open({ type: 'okCancel', title: 'Confirm', message: 'continue?' });
    await wrapper.vm.$nextTick();
    await wrapper.findAll('[data-ivy-dialog-button]')[0].trigger('click');
    await expect(cancel).resolves.toBe(false);
  });

  it('returns edited text, entry text, integers, and list selections from Vue controls', async () => {
    const dialogStore = useDialogStore();
    const wrapper = mountWithPinia(DialogHost);

    const text = dialogStore.open({
      type: 'text',
      title: 'Text',
      message: 'edit it',
      text: 'old',
      options: { okLabel: 'Use', cancel: true },
    });
    await wrapper.vm.$nextTick();
    await wrapper.find('[data-ivy-dialog-text]').setValue('new');
    await wrapper.findAll('[data-ivy-dialog-button]')[1].trigger('click');
    await expect(text).resolves.toBe('new');

    const entry = dialogStore.open({ type: 'entry', title: 'Name', message: 'graph name', initialValue: 'old' });
    await wrapper.vm.$nextTick();
    await wrapper.find('[data-ivy-dialog-entry]').setValue('new name');
    await wrapper.findAll('[data-ivy-dialog-button]')[1].trigger('click');
    await expect(entry).resolves.toBe('new name');

    const integer = dialogStore.open({
      type: 'integer',
      title: 'Bound',
      message: 'choose bound',
      initialValue: 2,
      options: { min: 1, max: 5 },
    });
    await wrapper.vm.$nextTick();
    await wrapper.find('[data-ivy-dialog-int]').setValue('10');
    await wrapper.findAll('[data-ivy-dialog-button]')[1].trigger('click');
    expect(wrapper.find('[data-ivy-dialog-error]').text()).toContain('at most 5');
    await wrapper.find('[data-ivy-dialog-int]').setValue('4');
    await wrapper.findAll('[data-ivy-dialog-button]')[1].trigger('click');
    await expect(integer).resolves.toBe(4);

    const single = dialogStore.open({ type: 'listbox', title: 'Pick', message: 'one', items: ['a', 'b'] });
    await wrapper.vm.$nextTick();
    const singleSelect = wrapper.find('[data-ivy-dialog-list]');
    singleSelect.element.options[1].selected = true;
    await singleSelect.trigger('change');
    await wrapper.findAll('[data-ivy-dialog-button]')[1].trigger('click');
    await expect(single).resolves.toBe('b');
  });

  it('supports multi-select listboxes through the dialog store', async () => {
    const dialogStore = useDialogStore();
    const wrapper = mountWithPinia(DialogHost);
    const multi = dialogStore.open({
      type: 'listbox',
      title: 'Pick',
      message: 'many',
      items: ['a', 'b', 'c'],
      options: { multiple: true },
    });
    await wrapper.vm.$nextTick();

    const select = wrapper.find('[data-ivy-dialog-list]');
    select.element.options[0].selected = true;
    select.element.options[2].selected = true;
    await select.trigger('change');
    await wrapper.findAll('[data-ivy-dialog-button]')[1].trigger('click');

    await expect(multi).resolves.toEqual(['a', 'c']);
  });
});
