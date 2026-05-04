# Plan: Fix File Dropdown Close + macOS Flash Effect
Created: 2026-05-04 UTC

## Context

Two related problems with the File dropdown menu in ivyweb:

1. **Bug**: After clicking a File menu item (Load, Save as, Download, etc.), the dropdown stays visible. It should close.
2. **Enhancement**: macOS Cocoa style — briefly invert the menu item's colors (~50 ms) before closing, giving tactile feedback that the item was selected.

---

## Root Cause of the Bug

**File:** `/Users/jaten/ivy/goivy/webui/static/js/ivyweb_app.js`, lines 2412–2421

`closeAllDropdowns(e)` has a guard clause that aborts early if the click target is inside `.dropdown-content`:

```javascript
closeAllDropdowns(e) {
    if (e && e.target && e.target.closest && e.target.closest('.dropdown-content')) {
        return;  // <-- BUG: all menu items ARE inside .dropdown-content
    }
    ...
}
```

Every File menu item (`<a id="file-load">`, `<a id="file-save-as">`, etc.) lives inside `<div class="dropdown-content">`, so this guard always fires and the dropdown never closes, even when a handler calls `closeAllDropdowns(e)` explicitly.

The guard was intended to protect the global `document` click handler (line 282) from closing the dropdown when the user clicks inside it — but it also blocks the explicit calls from menu item handlers.

---

## Fix

### 1. Remove the guard from `closeAllDropdowns`; fix the global handler instead

**`ivyweb_app.js` line 2412–2421** — replace the guard-ed version:

```javascript
closeAllDropdowns(e) {
    // Don't close if clicking inside a dropdown
    if (e && e.target && e.target.closest && e.target.closest('.dropdown-content')) {
        return;
    }
    var all = document.querySelectorAll('.dropdown.open');
    for (var j = 0; j < all.length; j++) {
        all[j].classList.remove('open');
    }
}
```

With a clean, no-guard version:

```javascript
closeAllDropdowns() {
    var all = document.querySelectorAll('.dropdown.open');
    for (var j = 0; j < all.length; j++) {
        all[j].classList.remove('open');
    }
}
```

**`ivyweb_app.js` global click handler** (around line 282) — add the "outside dropdown" check there instead:

```javascript
document.addEventListener('click', function (e) {
    self.controls.hideContextMenu();
    if (!e.target.closest('.dropdown')) {
        self.closeAllDropdowns();
    }
});
```

### 2. Add `flashAndClose(el, callback)` helper

Add this method to `IvyApp` (near `closeAllDropdowns`):

```javascript
flashAndClose(el, callback) {
    var self = this;
    el.classList.add('menu-flash');
    setTimeout(function () {
        el.classList.remove('menu-flash');
        self.closeAllDropdowns();
        if (callback) callback();
    }, 50);
}
```

### 3. Update all File menu item handlers to use `flashAndClose`

In `setupEventHandlers()` (lines 188–226), replace each handler's `closeAllDropdowns(e)` + direct action call with `flashAndClose(this, fn)`:

```javascript
// File > Load...
document.getElementById('file-load').addEventListener('click', function (e) {
    e.preventDefault();
    self.flashAndClose(this, function () { fileInput.click(); });
});

// File > Save as...
document.getElementById('file-save-as').addEventListener('click', function (e) {
    e.preventDefault();
    self.flashAndClose(this, function () { self.saveAs(); });
});

// File > Download current model
document.getElementById('file-download').addEventListener('click', function (e) {
    e.preventDefault();
    self.flashAndClose(this, function () { self.downloadModel(); });
});

// File > New Model
document.getElementById('file-new').addEventListener('click', function (e) {
    e.preventDefault();
    self.flashAndClose(this, function () { self.newModel(); });
});

// File > Save Invariant
document.getElementById('file-save-invariant').addEventListener('click', function (e) {
    e.preventDefault();
    self.flashAndClose(this, function () { self.saveInvariant(); });
});
```

Also update `bindMenuAction` (lines 2426–2438) to use `flashAndClose`:

```javascript
bindMenuAction(id, callback) {
    var self = this;
    var el = document.getElementById(id);
    if (!el) return;
    el.addEventListener('click', function (e) {
        e.preventDefault();
        e.stopPropagation();
        self.flashAndClose(this, callback);
    });
}
```

Also update recent-file link handlers (around line 2056):
```javascript
link.addEventListener('click', function (e) {
    e.preventDefault();
    self.flashAndClose(this, function () { self.loadRecentSession(s.id); });
});
```

### 4. Add `.menu-flash` CSS class

**`/Users/jaten/ivy/goivy/webui/static/css/ivy.css`** — append after `.dropdown.open .dropdown-content` (line 557):

```css
.dropdown-content a.menu-flash {
    background-color: #c8c8c8;
    color: #1e1e1e;
}
```

This inverts the item: light background, dark text — matching macOS Cocoa highlight-then-dismiss behavior.

---

## Critical Files

- `/Users/jaten/ivy/goivy/webui/static/js/ivyweb_app.js`
  - `closeAllDropdowns()` — line 2412 (remove guard)
  - global `document` click handler — ~line 282 (add `.closest('.dropdown')` guard there)
  - `setupEventHandlers()` File menu items — lines 188–226 (use `flashAndClose`)
  - `bindMenuAction()` — line 2426 (use `flashAndClose`)
  - recent-file links — ~line 2056 (use `flashAndClose`)
  - new `flashAndClose()` method — add near `closeAllDropdowns`
- `/Users/jaten/ivy/goivy/webui/static/css/ivy.css`
  - add `.menu-flash` rule after line 557

---

## Verification

1. Start ivyweb, open the browser
2. Click "File" — dropdown appears and stays open (existing behavior preserved)
3. Click "Load..." — item briefly flashes inverted (~50 ms), then dropdown closes, file picker opens
4. Click "File" again, click "Save as..." — same flash + close behavior
5. Click anywhere outside the File menu — dropdown closes (existing behavior preserved)
6. Click inside the dropdown on a non-link area (separator) — dropdown stays open (preserved)
