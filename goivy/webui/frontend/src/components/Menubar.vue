<template>
  <div id="menubar">
    <div class="menu-group">
      <div class="dropdown" :class="{ open: dropdownStore.isOpen('file-menu') }">
        <span class="panel-menu" data-dropdown="file-menu" @click="toggleDropdownCommand($event, dropdownStore, 'file-menu', () => runUiCommand('populateRecentFiles'))">File</span>
        <div id="file-menu" class="dropdown-content">
          <a href="#" id="file-load" :class="menuFlashClass('file-load')" @click="runMenuCommand($event, () => runUiCommand('file.load'))">Load...</a>
          <a href="#" id="file-open-event-trace" :class="menuFlashClass('file-open-event-trace')" @click="runMenuCommand($event, () => runUiCommand('chooseAndLoadEventTraceFile'))">Open Event Trace...</a>
          <a href="#" id="file-save-as" :class="menuFlashClass('file-save-as')" @click="runMenuCommand($event, () => runUiCommand('file.saveAs'))">Save as...</a>
          <a href="#" id="file-download" :class="menuFlashClass('file-download')" @click="runMenuCommand($event, () => runUiCommand('file.download'))">Download current model</a>
          <a href="#" id="file-save-analysis-state" :class="menuFlashClass('file-save-analysis-state')" @click="runMenuCommand($event, () => runUiCommand('saveAnalysisState'))">Save Analysis State...</a>
          <a href="#" id="file-load-analysis-state" :class="menuFlashClass('file-load-analysis-state')" @click="runMenuCommand($event, () => runUiCommand('chooseAndLoadAnalysisStateFile'))">Load Analysis State...</a>
          <a href="#" id="file-save-invariant" :class="menuFlashClass('file-save-invariant')" @click="runMenuCommand($event, () => runUiCommand('saveInvariant'))">Save Invariant...</a>
          <div class="dropdown-sep"></div>
          <RecentFilesMenu />
          <div class="dropdown-sep"></div>
          <div class="menu-spacer"></div>
          <a href="#" id="file-new" :class="menuFlashClass('file-new')" @click="runMenuCommand($event, () => runUiCommand('file.new'))">New Model</a>
        </div>
      </div>
    </div>
    <div class="menu-group">
      <span class="menu-label">Mode</span>
      <select id="mode-select" :value="sessionStore.mode" @change="sessionStore.setMode($event.target.value)">
        <option value="induction">Induction</option>
        <option value="pdr">PDR</option>
        <option value="concrete">Concrete</option>
        <option value="abstract">Abstract</option>
        <option value="bounded">Bounded</option>
      </select>
    </div>
    <div class="menu-group">
      <button id="btn-check" class="menu-btn action-btn" @click="runCommand($event, () => runUiCommand('runCheck'))">Check</button>
      <button id="btn-show-reachable" class="menu-btn" @click="runCommand($event, () => runUiCommand('showReachableStates'))">Show Reachable</button>
      <button id="btn-undo" class="menu-btn" @click="runCommand($event, () => runUiCommand('doUndo'))">Undo</button>
      <button id="btn-reset-domain" class="menu-btn" @click="runCommand($event, () => runUiCommand('resetDomain'))">Reset Domain</button>
      <button id="btn-diagram-domain" class="menu-btn" @click="runCommand($event, () => runUiCommand('diagramDomain'))">Diagram Domain</button>
      <span id="loaded-file" class="loaded-file" :title="sessionStore.loadedFileTitle">{{ sessionStore.loadedFileDisplay }}</span>
    </div>
    <div class="menu-group menu-right">
      <button
        id="btn-toggle-tutorial"
        class="menu-btn"
        :class="{ 'btn-flash': layoutStore.tutorialButtonFlashing }"
        @click="runCommand($event, () => runUiCommand('toggleTutorial'))"
      >{{ layoutStore.tutorialVisible ? 'Hide Tutorial' : 'Show Tutorial' }}</button>
    </div>
  </div>
</template>

<script setup>
import RecentFilesMenu from './RecentFilesMenu.vue';
import { runCommand, runMenuCommand, runUiCommand, toggleDropdownCommand } from '../services/uiCommandService.js';
import { useDropdownStore } from '../stores/dropdownStore.js';
import { useLayoutStore } from '../stores/layoutStore.js';
import { useSessionStore } from '../stores/sessionStore.js';

const layoutStore = useLayoutStore();
const dropdownStore = useDropdownStore();
const sessionStore = useSessionStore();

function menuFlashClass(id) {
  return { 'menu-flash': dropdownStore.flashingItemId === id };
}
</script>
