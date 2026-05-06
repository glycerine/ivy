<template>
  <div id="outer-container">
    <div id="top-row">
      <SheetArea />
      <div id="divider3" class="divider" @mousedown="startEditorResize"></div>
      <EditorPane />
    </div>
    <div id="divider-h" class="divider-horizontal" v-show="layoutStore.tutorialVisible" @mousedown="startTutorialResize"></div>
    <TutorialPane />
  </div>
</template>

<script setup>
import EditorPane from './panes/EditorPane.vue';
import SheetArea from './panes/SheetArea.vue';
import TutorialPane from './panes/TutorialPane.vue';
import { useLayoutStore } from '../stores/layoutStore.js';
import { scheduleLayoutRefresh, startMouseDrag } from '../resizeDrag.js';

const layoutStore = useLayoutStore();

function startEditorResize(event) {
  const divider = event.currentTarget;
  const sheetArea = document.getElementById('sheet-area');
  const editorPanel = document.getElementById('editor-panel');
  const topRow = document.getElementById('top-row');
  if (!divider || !sheetArea || !editorPanel || !topRow) return;
  const startX = event.clientX;
  const startWidth = editorPanel.offsetWidth;
  const minEditorWidth = 200;
  const minSheetAreaWidth = 400;
  sheetArea.style.flex = '1 1 auto';
  divider.classList.add('active');
  startMouseDrag(event, {
    cursor: 'col-resize',
    onMove(moveEvent) {
      const dividerWidth = divider.offsetWidth || 4;
      const maxWidth = Math.max(minEditorWidth, topRow.offsetWidth - dividerWidth - minSheetAreaWidth);
      const newWidth = Math.max(minEditorWidth, Math.min(startWidth + startX - moveEvent.clientX, maxWidth));
      layoutStore.setEditorWidth(newWidth);
      scheduleLayoutRefresh();
    },
    onEnd() {
      divider.classList.remove('active');
      scheduleLayoutRefresh();
    },
  });
}

function startTutorialResize(event) {
  const divider = event.currentTarget;
  const tutorial = document.getElementById('tutorial-container');
  const outerContainer = document.getElementById('outer-container');
  const iframe = document.getElementById('tutorial-iframe');
  if (!divider || !tutorial) return;
  const startY = event.clientY;
  const startHeight = tutorial.offsetHeight;
  divider.classList.add('active');
  if (iframe) iframe.style.pointerEvents = 'none';
  startMouseDrag(event, {
    cursor: 'row-resize',
    onMove(moveEvent) {
      const maxHeight = outerContainer ? outerContainer.offsetHeight - 100 : 600;
      const newHeight = Math.max(80, Math.min(startHeight + startY - moveEvent.clientY, maxHeight));
      layoutStore.setTutorialHeight(newHeight);
      scheduleLayoutRefresh();
    },
    onEnd() {
      divider.classList.remove('active');
      if (iframe) iframe.style.pointerEvents = '';
      scheduleLayoutRefresh();
    },
  });
}
</script>
