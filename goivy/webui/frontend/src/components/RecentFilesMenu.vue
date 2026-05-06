<script setup>
import { useRecentFilesStore } from '../stores/recentFilesStore.js';
import { useDropdownStore } from '../stores/dropdownStore.js';

const recentFilesStore = useRecentFilesStore();
const dropdownStore = useDropdownStore();

function choose(item) {
  recentFilesStore.flash(item.id);
  window.setTimeout(() => {
    recentFilesStore.clearFlash();
    dropdownStore.closeAll();
    recentFilesStore.load(item);
  }, 50);
}
</script>

<template>
  <div id="file-recent-list">
    <a
      v-if="recentFilesStore.items.length === 0"
      href="#"
      class="recent-files-empty"
    >(no recent files)</a>
    <a
      v-for="item in recentFilesStore.items"
      v-else
      :key="item.id"
      href="#"
      :title="item.title || ''"
      :class="{ 'menu-flash': recentFilesStore.flashingId === String(item.id) }"
      @click.stop.prevent="choose(item)"
    >
      {{ item.label }}
    </a>
  </div>
</template>
