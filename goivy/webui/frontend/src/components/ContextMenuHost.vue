<script setup>
import { useContextMenuStore } from '../stores/contextMenuStore.js';

const contextMenuStore = useContextMenuStore();
</script>

<template>
  <div id="context-menu" class="context-menu" :style="contextMenuStore.style">
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
