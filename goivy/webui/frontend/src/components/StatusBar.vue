<script setup>
import { useSessionStore } from '../stores/sessionStore.js';

const sessionStore = useSessionStore();
</script>

<template>
  <div id="statusbar" :class="sessionStore.statusLevel">
    <span class="status-message">{{ sessionStore.status }}</span>
    <span id="session-id" class="session-id">{{ sessionStore.sessionDisplay }}</span>
  </div>
</template>
