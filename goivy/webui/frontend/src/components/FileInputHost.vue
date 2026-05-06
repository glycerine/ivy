<script setup>
import { callApp, setAppStatus } from '../services/uiCommandService.js';

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
    try {
      await callApp('loadAnalysisStateFile', file);
    } catch (ex) {
      setAppStatus(`Load analysis state failed: ${ex.message}`, 'error');
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
