<script setup>
import { useSessionStore } from '../stores/sessionStore.js';

const sessionStore = useSessionStore();
</script>

<template>
  <div id="statusbar" :class="sessionStore.statusLevel">{{ sessionStore.status }}</div>
</template>
