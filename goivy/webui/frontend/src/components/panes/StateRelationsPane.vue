<template>
  <div id="state-panel" class="panel">
    <div class="panel-header">
      <strong class="column-title">State/relations</strong>
      <div class="panel-header-actions">
        <span id="state-label">State: 0</span>
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
        <tbody id="state-checkbox-body"></tbody>
      </table>
    </div>
  </div>
</template>
