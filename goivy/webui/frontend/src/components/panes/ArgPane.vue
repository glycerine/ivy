<template>
  <div id="arg-panel" class="panel" :style="layoutStore.argPanelStyle">
    <div class="panel-header">
      <strong class="column-title">ARG (Abstract Reachability Graph)</strong>
      <div class="panel-header-actions">
        <div class="dropdown" :class="{ open: dropdownStore.isOpen('arg-inv-menu') }">
          <span class="panel-menu" data-dropdown="arg-inv-menu" @click="toggleDropdownCommand($event, dropdownStore, 'arg-inv-menu')">Invariant</span>
          <div id="arg-inv-menu" class="dropdown-content">
            <a href="#" id="arg-check-induction" :class="menuFlashClass('arg-check-induction')" @click="runMenuCommand($event, () => callApp('checkInduction'))">Check induction</a>
            <a href="#" id="arg-bounded-check" :class="menuFlashClass('arg-bounded-check')" @click="runMenuCommand($event, () => callApp('boundedCheck'))">Bounded check</a>
            <a href="#" id="arg-diagram" :class="menuFlashClass('arg-diagram')" @click="runMenuCommand($event, () => callApp('diagramDomain'))">Diagram</a>
            <a href="#" id="arg-weaken" :class="menuFlashClass('arg-weaken')" @click="runMenuCommand($event, () => callApp('weakenInvariant'))">Weaken</a>
            <div class="dropdown-sep"></div>
            <a href="#" id="arg-save-invariant" :class="menuFlashClass('arg-save-invariant')" @click="runMenuCommand($event, () => callApp('saveInvariant'))">Save Invariant...</a>
            <a href="#" id="arg-save-abs" :class="menuFlashClass('arg-save-abs')" @click="runMenuCommand($event, () => callApp('saveAbstraction'))">Save Abstraction...</a>
          </div>
        </div>
        <DynamicMenuRegion region="arg" />
      </div>
    </div>
    <div id="arg-graph" class="graph-container" @contextmenu.prevent></div>
  </div>
</template>

<script setup>
import DynamicMenuRegion from '../DynamicMenuRegion.vue';
import { callApp, runMenuCommand, toggleDropdownCommand } from '../legacyCommand.js';
import { useDropdownStore } from '../../stores/dropdownStore.js';
import { useLayoutStore } from '../../stores/layoutStore.js';

const dropdownStore = useDropdownStore();
const layoutStore = useLayoutStore();

function menuFlashClass(id) {
  return { 'menu-flash': dropdownStore.flashingItemId === id };
}
</script>
