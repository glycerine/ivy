<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useContextMenuStore } from '../stores/contextMenuStore.js';
import { useDialogStore } from '../stores/dialogStore.js';
import { useSessionStore } from '../stores/sessionStore.js';

const contextMenuStore = useContextMenuStore();
const dialogStore = useDialogStore();
const sessionStore = useSessionStore();
const textInput = ref(null);
const scalarInput = ref(null);
const listInput = ref(null);

const activeDialog = computed(() => dialogStore.active);
const dialogButtons = computed(() => {
  const active = activeDialog.value;
  if (!active) return [];
  const opts = active.options || {};
  if (active.type === 'ok') {
    return [{ label: 'OK', action: 'submit' }];
  }
  if (active.type === 'okCancel') {
    return [{ label: 'Cancel', action: 'cancel' }, { label: 'OK', action: 'submit' }];
  }
  if (active.type === 'buttons') {
    return (active.buttons || []).map((button) => ({
      label: button.label || String(button.value),
      action: 'button',
      value: button.value,
      danger: !!button.danger,
    }));
  }
  const buttons = [];
  if (opts.cancel || (active.type !== 'text' && opts.cancel !== false)) {
    buttons.push({ label: 'Cancel', action: 'cancel' });
  }
  buttons.push({ label: opts.okLabel || 'OK', action: 'submit' });
  return buttons;
});

function handleDialogButton(button) {
  if (button.action === 'cancel') {
    dialogStore.cancel();
    return;
  }
  if (button.action === 'button') {
    dialogStore.chooseButton(button);
    return;
  }
  dialogStore.submit();
}

function handleKeydown(event) {
  if (event.key === 'Escape' && dialogStore.active) {
    dialogStore.escape();
  }
}

watch(activeDialog, async (active) => {
  if (!active) return;
  await nextTick();
  const target = textInput.value || scalarInput.value || listInput.value;
  if (target && typeof target.focus === 'function') {
    target.focus();
    if (typeof target.select === 'function' && active.type !== 'listbox') {
      target.select();
    }
  }
});

onMounted(() => {
  document.addEventListener('keydown', handleKeydown);
});

onBeforeUnmount(() => {
  document.removeEventListener('keydown', handleKeydown);
});
</script>

<template>
  <div id="context-menu" class="context-menu" :style="contextMenuStore.style">
    <template v-for="item in contextMenuStore.items" :key="item.key">
      <div v-if="item.kind === 'separator'" class="context-menu-separator"></div>
      <div v-else-if="item.kind === 'header'" class="context-menu-header">{{ item.header }}</div>
      <div
        v-else
        class="context-menu-item"
        :data-action-id="item.id"
        @click.stop="contextMenuStore.runItem(item)"
      >
        {{ item.name }}
      </div>
    </template>
  </div>

  <div
    v-if="activeDialog"
    class="dialog-overlay"
    data-ivy-dialog="true"
    style="display:flex;"
  >
    <div class="dialog-box">
      <div class="dialog-title">{{ activeDialog.title }}</div>
      <div class="dialog-message">{{ activeDialog.message }}</div>
      <div class="dialog-body">
        <textarea
          v-if="activeDialog.type === 'text'"
          ref="textInput"
          class="dialog-text"
          :rows="activeDialog.options.rows || 4"
          :cols="activeDialog.options.cols || 80"
          :readonly="!!activeDialog.options.readOnly"
          :value="activeDialog.inputValue"
          data-ivy-dialog-text="true"
          @input="dialogStore.setInputValue($event.target.value)"
        ></textarea>
        <input
          v-else-if="activeDialog.type === 'entry'"
          ref="scalarInput"
          type="text"
          class="dialog-input"
          :value="activeDialog.inputValue"
          data-ivy-dialog-entry="true"
          @input="dialogStore.setInputValue($event.target.value)"
        >
        <input
          v-else-if="activeDialog.type === 'integer'"
          ref="scalarInput"
          type="number"
          class="dialog-input"
          :value="activeDialog.inputValue"
          :min="activeDialog.options.min"
          :max="activeDialog.options.max"
          data-ivy-dialog-int="true"
          @input="dialogStore.setInputValue($event.target.value)"
        >
        <select
          v-else-if="activeDialog.type === 'listbox'"
          ref="listInput"
          class="dialog-input dialog-listbox"
          :size="activeDialog.options.size || Math.min(Math.max(activeDialog.entries.length, 2), 12)"
          :multiple="!!activeDialog.options.multiple"
          data-ivy-dialog-list="true"
          @change="activeDialog.options.multiple
            ? dialogStore.setSelectedIndices(Array.from($event.target.selectedOptions).map((option) => option.getAttribute('data-ivy-dialog-index')))
            : dialogStore.setSelectedIndex($event.target.selectedOptions[0] ? $event.target.selectedOptions[0].getAttribute('data-ivy-dialog-index') : '')"
        >
          <option
            v-for="(entry, index) in activeDialog.entries"
            :key="index"
            :value="String(entry.value)"
            :data-ivy-dialog-index="String(index)"
          >
            {{ entry.label }}
          </option>
        </select>
      </div>
      <div
        class="dialog-message dialog-error"
        data-ivy-dialog-error="true"
        :style="{ display: activeDialog.error ? 'block' : 'none' }"
      >
        {{ activeDialog.error }}
      </div>
      <div class="dialog-buttons">
        <button
          v-for="button in dialogButtons"
          :key="button.label"
          type="button"
          class="dialog-btn"
          :class="{ 'dialog-btn-danger': button.danger }"
          data-ivy-dialog-button="true"
          @click="handleDialogButton(button)"
        >
          {{ button.label }}
        </button>
      </div>
    </div>
  </div>

  <input type="file" id="file-input" accept=".ivy" style="display:none;">
  <input type="file" id="event-file-input" accept=".iev,.pats,.txt" style="display:none;">
  <input type="file" id="analysis-state-file-input" accept=".json,.ivyweb.json" style="display:none;">

  <div id="text-dialog-overlay" class="dialog-overlay" style="display:none;">
    <div class="dialog-box">
      <div class="dialog-title" id="text-dialog-title">ivyweb</div>
      <div class="dialog-message" id="text-dialog-message"></div>
      <textarea id="text-dialog-text" class="dialog-text" rows="4" cols="80"></textarea>
      <div class="dialog-buttons">
        <button id="text-dialog-ok" class="dialog-btn">OK</button>
      </div>
    </div>
  </div>

  <div id="external-change-dialog-overlay" class="dialog-overlay" style="display:none;">
    <div class="dialog-box external-change-dialog-box">
      <div class="dialog-title">File changed on disk</div>
      <div class="dialog-message" id="external-change-dialog-message"></div>
      <div class="dialog-buttons">
        <button id="external-change-do-nothing" class="dialog-btn">Do nothing</button>
        <button id="external-change-reload" class="dialog-btn">Revert to on-disk version</button>
        <button id="external-change-merge" class="dialog-btn">Merge disk version into edit buffer</button>
        <button id="external-change-overwrite" class="dialog-btn dialog-btn-danger">Overwrite on-disk with edited buffer</button>
      </div>
    </div>
  </div>

  <div id="dirty-close-dialog-overlay" class="dialog-overlay" style="display:none;">
    <div class="dialog-box dirty-close-dialog-box">
      <div class="dialog-title">File has changed</div>
      <div class="dialog-message" id="dirty-close-dialog-message"></div>
      <div class="dialog-buttons">
        <button id="dirty-close-cancel" class="dialog-btn">Cancel the close</button>
        <button id="dirty-close-save" class="dialog-btn">Save</button>
        <button id="dirty-close-discard" class="dialog-btn dialog-btn-danger">Discard Edits</button>
      </div>
    </div>
  </div>

  <div id="save-as-explain-notice" class="save-as-explain-notice" :style="{ display: sessionStore.saveAsNoticeVisible ? 'block' : 'none' }">
    The browser security model requires re-confirmation of the save path on disk when IvyWeb cannot locate an IndexedDB cached file handle.
  </div>

  <div id="loading-overlay" :style="{ display: sessionStore.loading ? 'flex' : 'none' }">
    <div class="spinner"></div>
    <div id="loading-message">{{ sessionStore.loadingMessage }}</div>
  </div>
</template>
