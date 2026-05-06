<script setup>
import { useDetailsStore } from '../../stores/detailsStore.js';
import { useLayoutStore } from '../../stores/layoutStore.js';

const detailsStore = useDetailsStore();
const layoutStore = useLayoutStore();
</script>

<template>
  <div id="info-panel" class="info-panel" :style="layoutStore.detailsPanelStyle">
    <div id="info-header" class="info-header" title="Drag to resize Details">Details</div>
    <div id="info-content">
      <template v-if="detailsStore.facts.length > 0">
        <div class="constraint-facts-title">Constraints:</div>
        <button
          v-for="fact in detailsStore.facts"
          :key="fact.index"
          type="button"
          class="constraint-fact"
          :class="{ inactive: !fact.selected }"
          :data-constraint-fact="String(fact.index)"
          :aria-pressed="fact.selected ? 'true' : 'false'"
          @click="detailsStore.toggleFact(fact.index)"
        >
          {{ fact.text }}
        </button>
      </template>
      <template v-else>
        {{ detailsStore.text }}
      </template>
      <template v-if="detailsStore.traceActionVisible">
        <br>
        <button type="button" class="btn small" data-check-view-trace="true" @click="detailsStore.runTraceAction()">View error trace</button>
      </template>
    </div>
  </div>
</template>
