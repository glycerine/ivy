<script setup>
import { useStateRelationsStore } from '../../stores/stateRelationsStore.js';
import { useLayoutStore } from '../../stores/layoutStore.js';

const stateRelationsStore = useStateRelationsStore();
const layoutStore = useLayoutStore();

const columns = [
  { key: 'all_to_all', label: '+', title: (name) => `Show definite edges (${name})` },
  { key: 'edge_unknown', label: '?', title: (name) => `Show unknown edges (${name})` },
  { key: 'none_to_none', label: '-', title: (name) => `Show absent edges (${name})` },
  { key: 'transitive', label: 'T', title: (name) => `Transitive reduction (${name})` },
];
</script>

<template>
  <div id="state-panel" class="panel" :style="layoutStore.statePanelStyle">
    <div class="panel-header">
      <strong class="column-title">State/relations</strong>
      <div class="panel-header-actions">
        <span id="state-label">State: {{ stateRelationsStore.stateLabel }}</span>
      </div>
    </div>
    <div id="state-controls">
      <table id="state-checkbox-table">
        <thead>
          <tr>
            <th class="chk-col">+</th>
            <th class="chk-col">?</th>
            <th class="chk-col">-</th>
            <th class="chk-col">T</th>
            <th class="name-col"></th>
          </tr>
        </thead>
        <tbody id="state-checkbox-body">
          <tr v-for="row in stateRelationsStore.rows" :key="row.name">
            <td v-for="column in columns" :key="column.key">
              <input
                type="checkbox"
                :name="row.name"
                :value="column.key"
                :title="column.title(row.name)"
                :checked="row.checked[column.key]"
                @change="stateRelationsStore.toggle(row.name, column.key, $event.target.checked)"
              >
            </td>
            <td class="name-col">
              <a href="#" @click.prevent>{{ row.name }}</a>
            </td>
          </tr>
          <tr v-if="stateRelationsStore.showPlaceholder">
            <td colspan="5" style="color:#666;font-style:italic;">No relations loaded</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
