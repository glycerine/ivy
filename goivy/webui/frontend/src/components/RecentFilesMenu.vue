<script setup>
import { useRecentFilesStore } from '../stores/recentFilesStore.js';

const recentFilesStore = useRecentFilesStore();

function closeDropdowns() {
  document.querySelectorAll('.dropdown.open').forEach((dropdown) => dropdown.classList.remove('open'));
}

function choose(item, event) {
  const link = event.currentTarget;
  link.classList.add('menu-flash');
  setTimeout(() => {
    link.classList.remove('menu-flash');
    closeDropdowns();
    recentFilesStore.load(item);
  }, 50);
}
</script>

<template>
  <div id="file-recent-list">
    <a
      v-if="recentFilesStore.items.length === 0"
      href="#"
      style="color:#666;pointer-events:none;"
    >(no recent files)</a>
    <a
      v-for="item in recentFilesStore.items"
      v-else
      :key="item.id"
      href="#"
      :title="item.title || ''"
      @click.stop.prevent="choose(item, $event)"
    >
      {{ item.label }}
    </a>
  </div>
</template>
