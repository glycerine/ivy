<template>
  <div id="concept-panel" class="panel">
    <div class="panel-header">
      <strong class="column-title">Concept graph</strong>
      <div class="panel-header-actions">
        <div class="dropdown">
          <span class="panel-menu" data-dropdown="conj-menu">Conjecture</span>
          <div id="conj-menu" class="dropdown-content">
            <a href="#" id="conj-undo">Undo</a>
            <a href="#" id="conj-redo">Redo</a>
            <div class="dropdown-sep"></div>
            <a href="#" id="conj-pdr-step">PDR step</a>
            <a href="#" id="conj-concrete">Concrete</a>
            <a href="#" id="conj-gather">Gather</a>
            <a href="#" id="conj-cti-gather">CTI Gather</a>
            <a href="#" id="conj-cti-minimize">Minimize</a>
            <a href="#" id="conj-cti-check-sufficient">Check sufficient</a>
            <a href="#" id="conj-cti-check-inductive">Check relative induction</a>
            <a href="#" id="conj-cti-strengthen">Strengthen</a>
            <a href="#" id="conj-reverse">Reverse</a>
            <a href="#" id="conj-path-reach">Path reach</a>
            <a href="#" id="conj-reach">Reach</a>
            <a href="#" id="conj-conjecture">Conjecture</a>
            <a href="#" id="conj-backtrack">Backtrack</a>
            <div class="dropdown-sep"></div>
            <a href="#" id="conj-recalculate">Recalculate</a>
            <a href="#" id="conj-diagram">Diagram</a>
            <a href="#" id="conj-remember">Remember</a>
            <a href="#" id="conj-export">Export</a>
          </div>
        </div>
        <div class="dropdown">
          <span class="panel-menu" data-dropdown="view-menu">View</span>
          <div id="view-menu" class="dropdown-content">
            <a href="#" id="view-add-relation">Add relation</a>
          </div>
        </div>
        <DynamicMenuRegion region="concept" />
      </div>
    </div>
    <div id="concept-graph" class="graph-container"></div>
  </div>
</template>

<script setup>
import DynamicMenuRegion from '../DynamicMenuRegion.vue';
</script>
