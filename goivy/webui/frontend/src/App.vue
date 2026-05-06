<script setup>
import { onMounted, onUnmounted } from 'vue';
import Menubar from './components/Menubar.vue';
import WorkspaceShell from './components/WorkspaceShell.vue';
import ContextMenuHost from './components/ContextMenuHost.vue';
import DialogHost from './components/DialogHost.vue';
import FileInputHost from './components/FileInputHost.vue';
import SessionOverlayHost from './components/SessionOverlayHost.vue';
import StatusBar from './components/StatusBar.vue';
import ToastHost from './components/ToastHost.vue';
import { installGlobalInteractions } from './globalInteractions.js';
import { currentAppServices } from './services/appServices.js';
import { useContextMenuStore, useDropdownStore, useMenuDescriptorStore } from './stores/index.js';

const contextMenuStore = useContextMenuStore();
const dropdownStore = useDropdownStore();
const menuDescriptorStore = useMenuDescriptorStore();
let cleanupGlobalInteractions = null;

onMounted(() => {
  cleanupGlobalInteractions = installGlobalInteractions({
    contextMenuStore,
    dropdownStore,
    menuDescriptorStore,
  });
  currentAppServices().start();
});

onUnmounted(() => {
  if (cleanupGlobalInteractions) cleanupGlobalInteractions();
  cleanupGlobalInteractions = null;
});
</script>

<template>
  <Menubar />
  <WorkspaceShell />
  <ContextMenuHost />
  <DialogHost />
  <FileInputHost />
  <SessionOverlayHost />
  <ToastHost />
  <StatusBar />
</template>
