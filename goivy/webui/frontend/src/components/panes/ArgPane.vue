<template>
  <div id="arg-panel" class="panel">
    <div class="panel-header">
      <strong class="column-title">ARG (Abstract Reachability Graph)</strong>
      <div class="panel-header-actions">
        <div class="dropdown" :class="{ open: dropdownStore.isOpen('arg-inv-menu') }">
          <span class="panel-menu" data-dropdown="arg-inv-menu" @click="toggleDropdownCommand($event, dropdownStore, 'arg-inv-menu')">Invariant</span>
          <div id="arg-inv-menu" class="dropdown-content">
            <a href="#" id="arg-check-induction" @click="runMenuCommand($event, () => callApp('checkInduction'))">Check induction</a>
            <a href="#" id="arg-bounded-check" @click="runMenuCommand($event, () => callApp('boundedCheck'))">Bounded check</a>
            <a href="#" id="arg-diagram" @click="runMenuCommand($event, () => callApp('diagramDomain'))">Diagram</a>
            <a href="#" id="arg-weaken" @click="runMenuCommand($event, () => callApp('weakenInvariant'))">Weaken</a>
            <div class="dropdown-sep"></div>
            <a href="#" id="arg-save-invariant" @click="runMenuCommand($event, () => callApp('saveInvariant'))">Save Invariant...</a>
            <a href="#" id="arg-save-abs" @click="runMenuCommand($event, () => callApp('saveAbstraction'))">Save Abstraction...</a>
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
import { callApp, runMenuCommand, toggleDropdownCommand } from '../legacyCommand.js';
import { useDropdownStore } from '../../stores/dropdownStore.js';

const dropdownStore = useDropdownStore();
</script>
