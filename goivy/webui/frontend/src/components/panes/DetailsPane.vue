<script setup>
import { useDetailsStore } from '../../stores/detailsStore.js';

const detailsStore = useDetailsStore();
</script>

<template>
  <div id="info-panel" class="info-panel">
    <div id="info-header" class="info-header" title="Drag to resize Details">Details</div>
    <div id="info-content">
      {{ detailsStore.text }}
      <template v-if="detailsStore.traceActionVisible">
        <br>
        <button type="button" class="btn small" data-check-view-trace="true" @click="detailsStore.runTraceAction()">View error trace</button>
      </template>
    </div>
  </div>
</template>
