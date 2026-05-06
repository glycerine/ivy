<template>
  <div id="concept-panel" class="panel">
    <div class="panel-header">
      <strong class="column-title">Concept graph</strong>
      <div class="panel-header-actions">
        <div class="dropdown" :class="{ open: dropdownStore.isOpen('conj-menu') }">
          <span class="panel-menu" data-dropdown="conj-menu" @click="toggleDropdownCommand($event, dropdownStore, 'conj-menu')">Conjecture</span>
          <div id="conj-menu" class="dropdown-content">
            <a href="#" id="conj-undo" :class="menuFlashClass('conj-undo')" @click="runMenuCommand($event, () => runUiCommand('doUndo'))">Undo</a>
            <a href="#" id="conj-redo" :class="menuFlashClass('conj-redo')" @click="runMenuCommand($event, () => runUiCommand('doRedo'))">Redo</a>
            <div class="dropdown-sep"></div>
            <a href="#" id="conj-pdr-step" :class="menuFlashClass('conj-pdr-step')" @click="runMenuCommand($event, () => runUiCommand('pdrStep'))">PDR step</a>
            <a href="#" id="conj-concrete" :class="menuFlashClass('conj-concrete')" @click="runMenuCommand($event, () => runUiCommand('concreteStep'))">Concrete</a>
            <a href="#" id="conj-gather" :class="menuFlashClass('conj-gather')" @click="runMenuCommand($event, () => runUiCommand('gatherFacts'))">Gather</a>
            <a href="#" id="conj-cti-gather" :class="menuFlashClass('conj-cti-gather')" @click="runMenuCommand($event, () => runUiCommand('ctiConceptAction', 'cti_gather'))">CTI Gather</a>
            <a href="#" id="conj-cti-minimize" :class="menuFlashClass('conj-cti-minimize')" @click="runMenuCommand($event, () => runUiCommand('ctiConceptAction', 'cti_minimize'))">Minimize</a>
            <a href="#" id="conj-cti-check-sufficient" :class="menuFlashClass('conj-cti-check-sufficient')" @click="runMenuCommand($event, () => runUiCommand('ctiConceptAction', 'cti_check_sufficient'))">Check sufficient</a>
            <a href="#" id="conj-cti-check-inductive" :class="menuFlashClass('conj-cti-check-inductive')" @click="runMenuCommand($event, () => runUiCommand('ctiConceptAction', 'cti_check_inductive'))">Check relative induction</a>
            <a href="#" id="conj-cti-strengthen" :class="menuFlashClass('conj-cti-strengthen')" @click="runMenuCommand($event, () => runUiCommand('ctiConceptAction', 'cti_strengthen'))">Strengthen</a>
            <a href="#" id="conj-reverse" :class="menuFlashClass('conj-reverse')" @click="runMenuCommand($event, () => runUiCommand('reverseStep'))">Reverse</a>
            <a href="#" id="conj-path-reach" :class="menuFlashClass('conj-path-reach')" @click="runMenuCommand($event, () => runUiCommand('pathReach'))">Path reach</a>
            <a href="#" id="conj-reach" :class="menuFlashClass('conj-reach')" @click="runMenuCommand($event, () => runUiCommand('reachStep'))">Reach</a>
            <a href="#" id="conj-conjecture" :class="menuFlashClass('conj-conjecture')" @click="runMenuCommand($event, () => runUiCommand('makeConjecture'))">Conjecture</a>
            <a href="#" id="conj-backtrack" :class="menuFlashClass('conj-backtrack')" @click="runMenuCommand($event, () => runUiCommand('backtrack'))">Backtrack</a>
            <div class="dropdown-sep"></div>
            <a href="#" id="conj-recalculate" :class="menuFlashClass('conj-recalculate')" @click="runMenuCommand($event, () => runUiCommand('recalculateGraph'))">Recalculate</a>
            <a href="#" id="conj-diagram" :class="menuFlashClass('conj-diagram')" @click="runMenuCommand($event, () => runUiCommand('diagramDomain'))">Diagram</a>
            <a href="#" id="conj-remember" :class="menuFlashClass('conj-remember')" @click="runMenuCommand($event, () => runUiCommand('rememberGraph'))">Remember</a>
            <a href="#" id="conj-export" :class="menuFlashClass('conj-export')" @click="runMenuCommand($event, () => runUiCommand('exportConjecture'))">Export</a>
          </div>
        </div>
        <div class="dropdown" :class="{ open: dropdownStore.isOpen('view-menu') }">
          <span class="panel-menu" data-dropdown="view-menu" @click="toggleDropdownCommand($event, dropdownStore, 'view-menu')">View</span>
          <div id="view-menu" class="dropdown-content">
            <a href="#" id="view-add-relation" :class="menuFlashClass('view-add-relation')" @click="runMenuCommand($event, () => runUiCommand('addRelationFromString'))">Add relation</a>
          </div>
        </div>
        <DynamicMenuRegion region="concept" />
      </div>
    </div>
    <div id="concept-graph" class="graph-container" @contextmenu.prevent></div>
  </div>
</template>

<script setup>
import DynamicMenuRegion from '../DynamicMenuRegion.vue';
import { runMenuCommand, runUiCommand, toggleDropdownCommand } from '../../services/uiCommandService.js';
import { useDropdownStore } from '../../stores/dropdownStore.js';

const dropdownStore = useDropdownStore();

function menuFlashClass(id) {
  return { 'menu-flash': dropdownStore.flashingItemId === id };
}
</script>
