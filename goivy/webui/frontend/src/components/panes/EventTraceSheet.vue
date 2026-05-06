<script setup>
import { computed } from 'vue';
import EventTraceNode from './EventTraceNode.vue';
import { callApp, hasAppMethod } from '../../services/uiCommandService.js';
import { useEventTraceStore } from '../../stores/eventTraceStore.js';

const props = defineProps({
  sheetId: {
    type: String,
    required: true,
  },
});

const eventTraceStore = useEventTraceStore();
const sheet = computed(() => eventTraceStore.sheetById(props.sheetId));

async function askEntry(title, message, okLabel, callback) {
  if (!hasAppMethod('entryDialog')) return;
  const pattern = await callApp('entryDialog', title, message, '', { okLabel });
  if (pattern !== null && pattern !== '') {
    await callback(pattern);
  }
}

function selectedPattern() {
  return eventTraceStore.selectedPattern(props.sheetId);
}

function selectedPatternIndex(event) {
  return event.target.selectedIndex;
}

async function loadPatterns() {
  if (!hasAppMethod('textDialog')) return;
  const text = await callApp('textDialog', 'Load patterns', 'Paste patterns:', '', { okLabel: 'Load' });
  if (text !== null) await callApp('loadEventPatterns', props.sheetId, text);
}
</script>

<template>
  <div class="event-viewer">
    <div class="event-tree-panel panel">
      <div class="panel-header">
        <span class="panel-title">Events</span>
        <button
          class="menu-btn event-filter-btn"
          type="button"
          @click="askEntry('Filter events', 'Pattern:', 'Filter', (pattern) => callApp('filterEventTrace', pattern))"
        >Filter...</button>
        <button
          class="menu-btn event-find-fwd-btn"
          type="button"
          @click="askEntry('Find event', 'Pattern:', 'Find', (pattern) => callApp('findEventTrace', pattern, false))"
        >&gt;&gt;</button>
        <button
          class="menu-btn event-find-rev-btn"
          type="button"
          @click="askEntry('Find event', 'Pattern:', 'Find', (pattern) => callApp('findEventTrace', pattern, true))"
        >&lt;&lt;</button>
      </div>
      <div class="event-tree" :data-event-tree="sheetId">
        <ul class="event-tree-list">
          <EventTraceNode
            v-for="event in (sheet && sheet.events) || []"
            :key="event.address"
            :sheet-id="sheetId"
            :event="event"
          />
        </ul>
      </div>
    </div>
    <div class="event-pattern-panel panel">
      <div class="panel-header"><span class="panel-title">Patterns</span></div>
      <select
        class="event-pattern-list"
        size="8"
        :value="selectedPattern()"
        @change="eventTraceStore.setSelectedPatternIndex(sheetId, selectedPatternIndex($event))"
      >
        <option
          v-for="pattern in (sheet && sheet.patterns) || []"
          :key="pattern"
          :value="pattern"
        >{{ pattern }}</option>
      </select>
      <div class="event-pattern-buttons">
        <button class="menu-btn event-pattern-rev" type="button" @click="selectedPattern() && callApp('findEventTrace', selectedPattern(), true)">&lt;&lt;</button>
        <button class="menu-btn event-pattern-fwd" type="button" @click="selectedPattern() && callApp('findEventTrace', selectedPattern(), false)">&gt;&gt;</button>
        <button class="menu-btn event-pattern-add" type="button" @click="askEntry('Add pattern', 'Pattern:', 'Add', (pattern) => callApp('addEventPattern', sheetId, pattern))">+</button>
        <button class="menu-btn event-pattern-remove" type="button" @click="callApp('removeSelectedEventPattern', sheetId)">-</button>
      </div>
      <div class="event-pattern-buttons">
        <button class="menu-btn event-pattern-save" type="button" @click="callApp('saveEventPatterns', sheetId)">Save</button>
        <button
          class="menu-btn event-pattern-load"
          type="button"
          @click="loadPatterns"
        >Load</button>
        <button class="menu-btn event-pattern-clear" type="button" @click="callApp('clearEventPatterns', sheetId)">Clear</button>
      </div>
    </div>
  </div>
</template>
