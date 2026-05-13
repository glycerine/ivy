<template>
  <div id="sheet-area" @mousedown="handleSheetResizeMouseDown">
    <TabBar />

    <div id="sheet-1" class="sheet-content active">
      <div class="sheet-columns">
        <div class="sheet-left">
          <div class="sheet-main">
            <ArgPane />
            <div id="divider" class="divider"></div>
            <ConceptPane />
          </div>

          <DetailsPane />
        </div>

        <div id="divider2" class="divider"></div>
        <StateRelationsPane />
      </div>
    </div>
  </div>
</template>

<script setup>
import ArgPane from './ArgPane.vue';
import ConceptPane from './ConceptPane.vue';
import DetailsPane from './DetailsPane.vue';
import StateRelationsPane from './StateRelationsPane.vue';
import TabBar from './TabBar.vue';
import { useLayoutStore } from '../../stores/layoutStore.js';
import { scheduleLayoutRefresh, setGraphPointerEvents, startMouseDrag } from '../../resizeDrag.js';

const layoutStore = useLayoutStore();

function detailsMinimumMainHeight(sheetLeft) {
  const fallback = 44;
  if (!sheetLeft) return fallback;
  const sheetMain = sheetLeft.querySelector('.sheet-main');
  if (!sheetMain) return fallback;
  let height = fallback;
  sheetMain.querySelectorAll('.panel-header').forEach((header) => {
    height = Math.max(height, header.offsetHeight || 0);
  });
  return height;
}

function startGraphDividerResize(event, divider) {
  const container = divider.parentElement;
  const panel = divider.previousElementSibling;
  if (!container || !panel) return;
  const startX = event.clientX;
  const startWidth = panel.offsetWidth;
  divider.classList.add('active');
  setGraphPointerEvents(event.currentTarget, 'none');
  startMouseDrag(event, {
    cursor: 'col-resize',
    onMove(moveEvent) {
      const containerWidth = container.offsetWidth || 800;
      const newWidth = Math.max(150, Math.min(startWidth + moveEvent.clientX - startX, containerWidth - 200));
      if (panel.id === 'arg-panel') {
        layoutStore.setArgPanelWidth(newWidth);
      } else {
        panel.style.flex = `0 0 ${newWidth}px`;
      }
      scheduleLayoutRefresh();
    },
    onEnd() {
      divider.classList.remove('active');
      setGraphPointerEvents(event.currentTarget, '');
      scheduleLayoutRefresh();
    },
  });
}

function startStateDividerResize(event, divider) {
  const rightSection = document.getElementById('state-panel');
  const topRow = divider.parentElement;
  if (!rightSection) return;
  const startX = event.clientX;
  const startWidth = rightSection.offsetWidth;
  divider.classList.add('active');
  startMouseDrag(event, {
    cursor: 'col-resize',
    onMove(moveEvent) {
      const maxWidth = topRow ? topRow.offsetWidth - 300 : 800;
      const newWidth = Math.max(200, Math.min(startWidth + startX - moveEvent.clientX, maxWidth));
      layoutStore.setStatePanelWidth(newWidth);
      scheduleLayoutRefresh();
    },
    onEnd() {
      divider.classList.remove('active');
      scheduleLayoutRefresh();
    },
  });
}

function startDetailsResize(event, header) {
  const panel = header.closest('.info-panel') || header.parentElement;
  const sheetLeft = panel ? panel.closest('.sheet-left') : null;
  if (!panel || !sheetLeft) return;
  const sheetMain = sheetLeft.querySelector('.sheet-main');
  const startY = event.clientY;
  const startHeight = panel.offsetHeight;
  header.classList.add('active');
  setGraphPointerEvents(sheetLeft, 'none');
  startMouseDrag(event, {
    cursor: 'row-resize',
    onMove(moveEvent) {
      const minDetailsHeight = 72;
      const minMainHeight = detailsMinimumMainHeight(sheetLeft);
      const maxHeight = Math.max(minDetailsHeight, sheetLeft.offsetHeight - minMainHeight);
      const newHeight = Math.max(minDetailsHeight, Math.min(startHeight + startY - moveEvent.clientY, maxHeight));
      if (sheetMain) sheetMain.style.minHeight = `${minMainHeight}px`;
      if (panel.id === 'info-panel') {
        layoutStore.setDetailsHeight(newHeight);
      } else {
        panel.style.flex = `0 0 ${newHeight}px`;
        panel.style.height = `${newHeight}px`;
      }
      scheduleLayoutRefresh();
    },
    onEnd() {
      header.classList.remove('active');
      setGraphPointerEvents(sheetLeft, '');
      scheduleLayoutRefresh();
    },
  });
}

function handleSheetResizeMouseDown(event) {
  const target = event.target;
  if (!(target instanceof Element)) return;
  if (target.id === 'divider2') {
    startStateDividerResize(event, target);
    return;
  }
  const header = target.closest('.info-header, #info-header');
  if (header && event.currentTarget.contains(header)) {
    startDetailsResize(event, header);
    return;
  }
  if (target.classList.contains('divider') && target.parentElement && target.parentElement.classList.contains('sheet-main')) {
    startGraphDividerResize(event, target);
  }
}
</script>
