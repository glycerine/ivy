<template>
  <div id="context-menu" class="context-menu" style="display:none;"></div>

  <input type="file" id="file-input" accept=".ivy" style="display:none;">
  <input type="file" id="event-file-input" accept=".iev,.pats,.txt" style="display:none;">
  <input type="file" id="analysis-state-file-input" accept=".json,.ivyweb.json" style="display:none;">

  <div id="text-dialog-overlay" class="dialog-overlay" style="display:none;">
    <div class="dialog-box">
      <div class="dialog-title" id="text-dialog-title">ivyweb</div>
      <div class="dialog-message" id="text-dialog-message"></div>
      <textarea id="text-dialog-text" class="dialog-text" rows="4" cols="80"></textarea>
      <div class="dialog-buttons">
        <button id="text-dialog-ok" class="dialog-btn">OK</button>
      </div>
    </div>
  </div>

  <div id="external-change-dialog-overlay" class="dialog-overlay" style="display:none;">
    <div class="dialog-box external-change-dialog-box">
      <div class="dialog-title">File changed on disk</div>
      <div class="dialog-message" id="external-change-dialog-message"></div>
      <div class="dialog-buttons">
        <button id="external-change-do-nothing" class="dialog-btn">Do nothing</button>
        <button id="external-change-reload" class="dialog-btn">Revert to on-disk version</button>
        <button id="external-change-merge" class="dialog-btn">Merge disk version into edit buffer</button>
        <button id="external-change-overwrite" class="dialog-btn dialog-btn-danger">Overwrite on-disk with edited buffer</button>
      </div>
    </div>
  </div>

  <div id="dirty-close-dialog-overlay" class="dialog-overlay" style="display:none;">
    <div class="dialog-box dirty-close-dialog-box">
      <div class="dialog-title">File has changed</div>
      <div class="dialog-message" id="dirty-close-dialog-message"></div>
      <div class="dialog-buttons">
        <button id="dirty-close-cancel" class="dialog-btn">Cancel the close</button>
        <button id="dirty-close-save" class="dialog-btn">Save</button>
        <button id="dirty-close-discard" class="dialog-btn dialog-btn-danger">Discard Edits</button>
      </div>
    </div>
  </div>

  <div id="save-as-explain-notice" class="save-as-explain-notice" style="display:none;">
    The browser security model requires re-confirmation of the save path on disk when IvyWeb cannot locate an IndexedDB cached file handle.
  </div>

  <div id="loading-overlay" style="display:none;">
    <div class="spinner"></div>
    <div id="loading-message">Loading...</div>
  </div>
</template>
