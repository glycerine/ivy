<template>
  <div id="tutorial-container" :style="layoutStore.tutorialPanelStyle">
    <div class="panel-header tutorial-url-bar">
      <button id="tutorial-back" class="tutorial-nav-btn" title="Back" :disabled="!layoutStore.canGoBack" @click="runCommand($event, () => layoutStore.goTutorialBack())">&#9664;</button>
      <button id="tutorial-fwd" class="tutorial-nav-btn" title="Forward" :disabled="!layoutStore.canGoForward" @click="runCommand($event, () => layoutStore.goTutorialForward())">&#9654;</button>
      <button id="tutorial-reload" class="tutorial-nav-btn" title="Reload" @click="runCommand($event, () => layoutStore.reloadTutorial())">&#8635;</button>
      <input
        id="tutorial-url"
        type="text"
        class="tutorial-url-input"
        :value="layoutStore.tutorialInput"
        spellcheck="false"
        @input="layoutStore.setTutorialInput($event.target.value)"
        @keydown="handleUrlKeydown"
      >
      <button id="tutorial-close" class="tutorial-close-btn" title="Close tutorial" @click="runCommand($event, closeTutorial)">&#215;</button>
    </div>
    <iframe
      id="tutorial-iframe"
      :key="layoutStore.tutorialFrameKey"
      class="tutorial-iframe"
      :src="layoutStore.tutorialUrl"
      @load="recordFrameLoad"
    ></iframe>
  </div>
</template>

<script setup>
import { useLayoutStore } from '../../stores/layoutStore.js';
import { ivyApp, runCommand } from '../legacyCommand.js';

const layoutStore = useLayoutStore();

function handleUrlKeydown(event) {
  if (event.key !== 'Enter') return;
  runCommand(event, () => layoutStore.navigateTutorial(layoutStore.tutorialInput));
}

function closeTutorial() {
  const app = ivyApp();
  if (app && typeof app.toggleTutorial === 'function') {
    app.toggleTutorial(true);
    return;
  }
  layoutStore.setTutorialVisible(false);
}

function recordFrameLoad(event) {
  try {
    const href = event.target.contentWindow.location.href;
    layoutStore.recordTutorialLoad(href);
  } catch (e) {
    // Cross-origin tutorial pages cannot be inspected by the parent frame.
  }
}
</script>
