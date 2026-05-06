<script setup>
defineProps({
  counter: {
    type: Number,
    required: true,
  },
});
</script>

<template>
  <div class="sheet-columns">
    <div class="sheet-left">
      <div class="sheet-main">
        <div class="panel" style="flex:0 0 30%;">
          <div class="panel-header">
            <strong class="column-title">ARG (Abstract Reachability Graph)</strong>
          </div>
          <div :id="`arg-graph-${counter}`" class="graph-container"></div>
        </div>
        <div class="divider"></div>
        <div class="panel" style="flex:1;">
          <div class="panel-header">
            <strong class="column-title">Concept graph</strong>
          </div>
          <div :id="`concept-graph-${counter}`" class="graph-container"></div>
        </div>
      </div>

      <div class="info-panel">
        <div :id="`info-header-${counter}`" class="info-header" title="Drag to resize Details">Details</div>
        <div :id="`info-content-${counter}`">Select a node or edge to see details</div>
      </div>
    </div>

    <div class="divider"></div>
    <div class="panel" style="flex:0 0 220px;min-width:180px;overflow-y:auto;">
      <div class="panel-header">
        <strong class="column-title">State/relations</strong>
      </div>
      <div class="state-controls-placeholder"></div>
    </div>
  </div>
</template>
