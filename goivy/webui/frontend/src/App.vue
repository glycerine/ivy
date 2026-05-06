<script setup>
import { nextTick, onMounted } from 'vue';
import Menubar from './components/Menubar.vue';
import WorkspaceShell from './components/WorkspaceShell.vue';
import DialogHost from './components/DialogHost.vue';
import StatusBar from './components/StatusBar.vue';

onMounted(async () => {
  await nextTick();
  if (typeof window.startIvyApp === 'function') {
    window.startIvyApp();
  }
});
</script>

<template>
  <Menubar />
  <WorkspaceShell />
  <DialogHost />
  <StatusBar />
</template>
