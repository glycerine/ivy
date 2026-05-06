<script setup>
import { useSessionStore } from '../stores/sessionStore.js';

const sessionStore = useSessionStore();
</script>

<template>
  <div id="save-as-explain-notice" v-show="sessionStore.saveAsNoticeVisible" class="save-as-explain-notice">
    The browser security model requires re-confirmation of the save path on disk when IvyWeb cannot locate an IndexedDB cached file handle.
  </div>

  <div id="loading-overlay" v-show="sessionStore.loading">
    <div class="spinner"></div>
    <div id="loading-message">{{ sessionStore.loadingMessage }}</div>
  </div>
</template>
