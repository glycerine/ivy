<template>
  <div id="menubar">
    <div class="menu-group">
      <div class="dropdown">
        <span class="panel-menu" data-dropdown="file-menu">File</span>
        <div id="file-menu" class="dropdown-content">
          <a href="#" id="file-load">Load...</a>
          <a href="#" id="file-open-event-trace">Open Event Trace...</a>
          <a href="#" id="file-save-as">Save as...</a>
          <a href="#" id="file-download">Download current model</a>
          <a href="#" id="file-save-analysis-state">Save Analysis State...</a>
          <a href="#" id="file-load-analysis-state">Load Analysis State...</a>
          <a href="#" id="file-save-invariant">Save Invariant...</a>
          <div class="dropdown-sep"></div>
          <div id="file-recent-list"></div>
          <div class="dropdown-sep"></div>
          <div style="height:4px"></div>
          <a href="#" id="file-new">New Model</a>
        </div>
      </div>
    </div>
    <div class="menu-group">
      <span class="menu-label">Mode</span>
      <select id="mode-select">
        <option value="induction">Induction</option>
        <option value="pdr" selected>PDR</option>
        <option value="concrete">Concrete</option>
        <option value="abstract">Abstract</option>
        <option value="bounded">Bounded</option>
      </select>
    </div>
    <div class="menu-group">
      <button id="btn-check" class="menu-btn action-btn">Check</button>
      <button id="btn-show-reachable" class="menu-btn">Show Reachable</button>
      <button id="btn-undo" class="menu-btn">Undo</button>
      <button id="btn-reset-domain" class="menu-btn">Reset Domain</button>
      <button id="btn-diagram-domain" class="menu-btn">Diagram Domain</button>
      <span id="loaded-file" class="loaded-file"></span>
    </div>
    <div class="menu-group menu-right">
      <button id="btn-toggle-tutorial" class="menu-btn">Hide Tutorial</button>
    </div>
  </div>
</template>
