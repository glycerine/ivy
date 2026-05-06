<template>
  <div id="arg-panel" class="panel">
    <div class="panel-header">
      <strong class="column-title">ARG (Abstract Reachability Graph)</strong>
      <div class="panel-header-actions">
        <div class="dropdown">
          <span class="panel-menu" data-dropdown="arg-inv-menu">Invariant</span>
          <div id="arg-inv-menu" class="dropdown-content">
            <a href="#" id="arg-check-induction">Check induction</a>
            <a href="#" id="arg-bounded-check">Bounded check</a>
            <a href="#" id="arg-diagram">Diagram</a>
            <a href="#" id="arg-weaken">Weaken</a>
            <div class="dropdown-sep"></div>
            <a href="#" id="arg-save-invariant">Save Invariant...</a>
            <a href="#" id="arg-save-abs">Save Abstraction...</a>
          </div>
        </div>
        <DynamicMenuRegion region="arg" />
      </div>
    </div>
    <div id="arg-graph" class="graph-container"></div>
  </div>
</template>

<script setup>
import DynamicMenuRegion from '../DynamicMenuRegion.vue';
</script>
