<script setup>
import { useToastStore } from '../stores/toastStore.js';

const toastStore = useToastStore();
</script>

<template>
  <div class="ivy-toast-stack" aria-live="polite">
    <button
      v-for="toast in toastStore.items"
      :key="toast.id"
      type="button"
      class="ivy-toast"
      :class="[`ivy-toast-${toast.level}`, toast.className]"
      @click="toastStore.remove(toast.id)"
    >
      {{ toast.message }}
    </button>
  </div>
</template>
