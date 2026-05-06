<template>
  <div id="outer-container">
    <div id="top-row">
      <SheetArea />
      <div id="divider3" class="divider"></div>
      <EditorPane />
    </div>
    <div id="divider-h" class="divider-horizontal"></div>
    <TutorialPane />
  </div>
</template>

<script setup>
import EditorPane from './panes/EditorPane.vue';
import SheetArea from './panes/SheetArea.vue';
import TutorialPane from './panes/TutorialPane.vue';
</script>
