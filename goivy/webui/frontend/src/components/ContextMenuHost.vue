<script setup>
import { nextTick, ref, watch } from 'vue';
import { useContextMenuStore } from '../stores/contextMenuStore.js';

const contextMenuStore = useContextMenuStore();
const menuEl = ref(null);

function fitMenuToViewport() {
  if (!contextMenuStore.visible || !menuEl.value) return;
  const rect = menuEl.value.getBoundingClientRect();
  const viewWidth = window.innerWidth || 0;
  const viewHeight = window.innerHeight || 0;
  const maxX = Math.max(0, viewWidth - rect.width);
  const maxY = Math.max(0, viewHeight - rect.height);
  const nextX = Math.min(contextMenuStore.x, maxX);
  const nextY = Math.min(contextMenuStore.y, maxY);
  if (nextX !== contextMenuStore.x || nextY !== contextMenuStore.y) {
    contextMenuStore.setPosition(nextX, nextY);
  }
}

watch(
  () => [contextMenuStore.visible, contextMenuStore.x, contextMenuStore.y, contextMenuStore.items.length],
  async () => {
    await nextTick();
    fitMenuToViewport();
  },
  { flush: 'post', immediate: true },
);
</script>

<template>
  <div id="context-menu" ref="menuEl" class="context-menu" :style="contextMenuStore.style">
    <template v-for="item in contextMenuStore.items" :key="item.key">
      <div v-if="item.kind === 'separator'" class="context-menu-separator"></div>
      <div v-else-if="item.kind === 'header'" class="context-menu-header">{{ item.header }}</div>
      <div
        v-else
        class="context-menu-item"
        :data-action-id="item.id"
        @click.stop="contextMenuStore.runItem(item)"
      >
        {{ item.name }}
      </div>
    </template>
  </div>
</template>
