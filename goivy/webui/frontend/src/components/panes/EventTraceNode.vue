<script setup>
import { computed } from 'vue';
import { useEventTraceStore } from '../../stores/eventTraceStore.js';
import { hasUiCommand, runUiCommand } from '../../services/uiCommandService.js';

const props = defineProps({
  sheetId: {
    type: String,
    required: true,
  },
  event: {
    type: Object,
    required: true,
  },
});

const eventTraceStore = useEventTraceStore();
const hasSubs = computed(() => props.event.subs && props.event.subs.length > 0);
const expanded = computed(() => eventTraceStore.isExpanded(props.sheetId, props.event.address));
const selected = computed(() => {
  const sheet = eventTraceStore.sheetById(props.sheetId);
  return !!sheet && sheet.selectedEventAddress === props.event.address;
});

function selectRow() {
  if (hasUiCommand('selectEventTraceRow')) {
    runUiCommand('selectEventTraceRow', props.sheetId, props.event.address);
    return;
  }
  eventTraceStore.selectEvent(props.sheetId, props.event.address);
}

function toggle(event) {
  event.stopPropagation();
  if (!hasSubs.value) return;
  if (hasUiCommand('toggleEventTraceNode')) {
    runUiCommand('toggleEventTraceNode', props.sheetId, props.event.address);
    return;
  }
  eventTraceStore.setExpanded(props.sheetId, props.event.address, !expanded.value);
}
</script>

<template>
  <li class="event-tree-node" :data-event-node="event.address">
    <div
      class="event-row"
      :class="{ selected }"
      :data-event-address="event.address"
      @click="selectRow"
    >
      <button
        class="event-toggle"
        type="button"
        :disabled="!hasSubs"
        :data-event-toggle="hasSubs ? event.address : undefined"
        @click="toggle"
      >{{ hasSubs ? (expanded ? '-' : '+') : '' }}</button>
      <span class="event-text">{{ event.text }}</span>
    </div>
    <ul v-if="hasSubs && expanded" class="event-tree-list">
      <EventTraceNode
        v-for="child in event.subs"
        :key="child.address"
        :sheet-id="sheetId"
        :event="child"
      />
    </ul>
  </li>
</template>
