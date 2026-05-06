<script setup>
import { callApp, ivyApp } from './legacyCommand.js';

async function handleModelFileChange(event) {
  const input = event.target;
  const file = input.files && input.files[0];
  if (file) {
    await callApp('loadFile', file);
  }
  input.value = '';
}

async function handleEventTraceFileChange(event) {
  const input = event.target;
  const file = input.files && input.files[0];
  if (file) {
    await callApp('loadEventTraceFile', file);
  }
  input.value = '';
}

async function handleAnalysisStateFileChange(event) {
  const input = event.target;
  const file = input.files && input.files[0];
  if (file) {
    const app = ivyApp();
    try {
      await callApp('loadAnalysisStateFile', file);
    } catch (ex) {
      if (app && app.controls && typeof app.controls.setStatus === 'function') {
        app.controls.setStatus(`Load analysis state failed: ${ex.message}`, 'error');
      }
    }
  }
  input.value = '';
}
</script>

<template>
  <input type="file" id="file-input" class="hidden-file-input" accept=".ivy" @change="handleModelFileChange">
  <input type="file" id="event-file-input" class="hidden-file-input" accept=".iev,.pats,.txt" @change="handleEventTraceFileChange">
  <input type="file" id="analysis-state-file-input" class="hidden-file-input" accept=".json,.ivyweb.json" @change="handleAnalysisStateFileChange">
</template>
