<script setup>
import { useEventTraceStore } from '../../stores/eventTraceStore.js';
import { useSheetStore } from '../../stores/sheetStore.js';
import EventTraceSheet from './EventTraceSheet.vue';

const eventTraceStore = useEventTraceStore();
const sheetStore = useSheetStore();
</script>

<template>
  <div
    v-for="sheet in eventTraceStore.sheetList"
    :id="sheet.id"
    :key="sheet.id"
    class="sheet-content event-sheet"
    :class="{ active: sheetStore.activeSheetId === sheet.id }"
  >
    <EventTraceSheet :sheet-id="sheet.id" />
  </div>
</template>
