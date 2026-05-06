<script setup>
import { computed } from 'vue';
import { useLayoutStore } from '../../stores/layoutStore.js';

defineProps({
  counter: {
    type: Number,
    required: true,
  },
});

const layoutStore = useLayoutStore();
const argPanelStyle = computed(() => (
  layoutStore.argPanelWidth ? layoutStore.argPanelStyle : { flex: '0 0 30%' }
));
const statePanelStyle = computed(() => ({
  flex: layoutStore.statePanelWidth ? `0 0 ${layoutStore.statePanelWidth}px` : '0 0 220px',
  minWidth: '180px',
  overflowY: 'auto',
}));
</script>

<template>
  <div class="sheet-columns">
    <div class="sheet-left">
      <div class="sheet-main">
        <div class="panel" :style="argPanelStyle">
          <div class="panel-header">
            <strong class="column-title">ARG (Abstract Reachability Graph)</strong>
          </div>
          <div :id="`arg-graph-${counter}`" class="graph-container" @contextmenu.prevent></div>
        </div>
        <div class="divider"></div>
        <div class="panel concept-sheet-panel">
          <div class="panel-header">
            <strong class="column-title">Concept graph</strong>
          </div>
          <div :id="`concept-graph-${counter}`" class="graph-container" @contextmenu.prevent></div>
        </div>
      </div>

      <div class="info-panel" :style="layoutStore.detailsPanelStyle">
        <div :id="`info-header-${counter}`" class="info-header" title="Drag to resize Details">Details</div>
        <div :id="`info-content-${counter}`">Select a node or edge to see details</div>
      </div>
    </div>

    <div class="divider"></div>
    <div class="panel" :style="statePanelStyle">
      <div class="panel-header">
        <strong class="column-title">State/relations</strong>
      </div>
      <div class="state-controls-placeholder"></div>
    </div>
  </div>
</template>
