<template>
  <div id="tutorial-container" :style="{ display: layoutStore.tutorialVisible ? '' : 'none' }">
    <div class="panel-header tutorial-url-bar">
      <button id="tutorial-back" class="tutorial-nav-btn" title="Back">&#9664;</button>
      <button id="tutorial-fwd" class="tutorial-nav-btn" title="Forward">&#9654;</button>
      <button id="tutorial-reload" class="tutorial-nav-btn" title="Reload">&#8635;</button>
      <input id="tutorial-url" type="text" class="tutorial-url-input" value="/static/tutorial/kenmcmil.github.io/ivy/language.html" spellcheck="false">
      <button id="tutorial-close" class="tutorial-close-btn" title="Close tutorial">&#215;</button>
    </div>
    <iframe id="tutorial-iframe" class="tutorial-iframe" src="/static/tutorial/kenmcmil.github.io/ivy/language.html"></iframe>
  </div>
</template>

<script setup>
import { useLayoutStore } from '../../stores/layoutStore.js';

const layoutStore = useLayoutStore();
</script>
