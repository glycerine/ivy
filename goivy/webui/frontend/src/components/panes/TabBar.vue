<script setup>
import { useSheetStore } from '../../stores/sheetStore.js';
import { runCommand, runUiCommand } from '../../services/uiCommandService.js';

const sheetStore = useSheetStore();
</script>

<template>
  <div id="tab-bar">
    <button
      v-for="tab in sheetStore.tabs"
      :key="tab.id"
      class="sheet-tab"
      :class="{ active: sheetStore.activeSheetId === tab.id }"
      :data-sheet="tab.id"
      @click="runCommand($event, () => runUiCommand('switchSheet', tab.id))"
    >
      <span>{{ tab.label }}</span>
      <span
        v-if="tab.closable"
        class="tab-close"
        title="Close tab"
        @click="runCommand($event, () => runUiCommand('removeSheet', tab.id))"
      >&times;</span>
    </button>
  </div>
</template>
