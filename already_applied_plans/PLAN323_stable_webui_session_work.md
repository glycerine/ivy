# PLAN: Fix intermittent webui browser test hang — SSE reconnect storm

Created: 2026-04-17, 18:30

## Context

Browser tests intermittently hang forever at `page.MustWaitStable()`. The root cause is the `EventSource` SSE auto-reconnect cycle destabilizing the DOM. When the SSE endpoint returns an error or the connection drops, the browser's `EventSource` silently reconnects every ~3 seconds, causing continuous network activity and subtle DOM changes that prevent rod's `WaitStable` from ever seeing two identical DOM snapshots.

The user wants: (1) deterministic tests that never hang, (2) a foundation for future user-facing connection error surfacing (5-second threshold), (3) no over-engineering.

## Root cause trace

1. `DOMContentLoaded` → `new IvyApp().init()` (ivyweb_app.js:2803-2804)
2. `init()` calls `await this.api.createSession()` → sets `this.api.sessionId` (line 42)
3. Then `this.api.connectEvents(this.handleEvent.bind(this))` (line 93)
4. `connectEvents()` opens `new EventSource('/api/session/{id}/events')` (ivyweb_api.js:304)
5. Server's `apiEvents` (handlers.go:419) blocks in `for/select` waiting for events
6. If the SSE connection drops or errors, `EventSource.onerror` fires (line 315) — browser auto-reconnects
7. Each reconnect cycle causes network activity → DOM snapshot changes → `WaitStable` never returns

The test sequence: `newPage()` → `MustWaitLoad()` → return page → test calls `MustWaitStable()`. Between `MustWaitLoad` (HTML loaded) and `MustWaitStable`, the async `init()` is running — it creates the session and opens the SSE connection. The SSE connection is now churning in the background.

## Fix — two layers

### Layer 1: Page timeout safety net (ALREADY DONE)

`.Timeout(30 * time.Second)` on the page in `newPage()` — converts infinite hangs to bounded test failures.

### Layer 2: SSE reconnect control with backoff and max retries

Replace the raw `EventSource` with a managed wrapper that has:
- **Max retries** (e.g., 5) before giving up
- **Exponential backoff** (1s, 2s, 4s, 8s, 16s)
- **`onConnectionLost` callback** — paves the way for future UI error surfacing
- **Clean close** — no reconnect after intentional `disconnectEvents()`

This eliminates the infinite reconnect storm that destabilizes `WaitStable`.

### Changes

#### `webui/static/js/ivyweb_api.js` — replace `connectEvents`

Replace the current `connectEvents` method (lines 299-318) with:

```javascript
/**
 * Connect to Server-Sent Events for real-time updates.
 * Uses manual reconnect with exponential backoff instead of
 * EventSource's default auto-reconnect, which can storm
 * indefinitely and destabilize browser tests.
 *
 * @param {function} onEvent - Callback receiving parsed event objects
 */
connectEvents(onEvent) {
    this.disconnectEvents();
    this._sseRetries = 0;
    this._sseMaxRetries = 5;
    this._sseClosed = false;
    this._sseOnEvent = onEvent;
    this._sseConnect();
}

/**
 * Internal: open one EventSource connection.
 * On error, reconnects with exponential backoff up to _sseMaxRetries.
 */
_sseConnect() {
    if (this._sseClosed || !this.sessionId) return;

    var url = this.baseURL + '/api/session/' + this.sessionId + '/events';
    var es = new EventSource(url);
    this.eventSource = es;
    var self = this;

    es.onopen = function () {
        self._sseRetries = 0;  // reset on successful connect
    };

    es.onmessage = function (e) {
        try {
            var event = JSON.parse(e.data);
            self._sseOnEvent(event);
        } catch (err) {
            console.error('Failed to parse SSE event:', err, e.data);
        }
    };

    es.onerror = function () {
        es.close();  // stop browser auto-reconnect
        if (self._sseClosed) return;
        self._sseRetries++;
        if (self._sseRetries > self._sseMaxRetries) {
            console.error('SSE: max retries exceeded, giving up');
            if (self.onConnectionLost) {
                self.onConnectionLost();
            }
            return;
        }
        var delay = Math.min(1000 * Math.pow(2, self._sseRetries - 1), 16000);
        console.warn('SSE: reconnect attempt ' + self._sseRetries +
                     '/' + self._sseMaxRetries + ' in ' + delay + 'ms');
        self._sseTimer = setTimeout(function () {
            self._sseConnect();
        }, delay);
    };
}

/**
 * Disconnect from Server-Sent Events.
 */
disconnectEvents() {
    this._sseClosed = true;
    if (this._sseTimer) {
        clearTimeout(this._sseTimer);
        this._sseTimer = null;
    }
    if (this.eventSource) {
        this.eventSource.close();
        this.eventSource = null;
    }
}
```

Key design points:
- `es.close()` in `onerror` prevents the browser's built-in auto-reconnect
- Manual reconnect with exponential backoff (1s → 2s → 4s → 8s → 16s)
- Max 5 retries then stops — no infinite storm
- `onopen` resets retry count — a successful connection means the server is up
- `this._sseClosed` flag prevents reconnect after intentional `disconnectEvents()`
- `this.onConnectionLost` callback — future hook for UI error message
- Guard `if (!this.sessionId) return` — don't even try to connect without a session

#### `webui/static/js/ivyweb_app.js` — wire up `onConnectionLost` to visible UI

In `init()`, after `connectEvents` (line 93), add the callback that surfaces the error in two places: the status bar AND a non-modal toast notification that the user can dismiss:

```javascript
this.api.onConnectionLost = function () {
    self.controls.setStatus('Server connection lost', 'error');
    self._showToast('Connection to server lost. Check that the server is running and reload the page.', 'error');
};
```

Add the `_showToast` method to `IvyApp`:

```javascript
/**
 * Show a non-modal toast notification. Auto-dismisses after 10s
 * or on click. Appends to document body so it floats over everything.
 */
_showToast(message, level) {
    var toast = document.createElement('div');
    toast.className = 'ivy-toast ivy-toast-' + (level || 'info');
    toast.textContent = message;
    toast.style.cssText = 'position:fixed;top:20px;right:20px;z-index:10000;' +
        'padding:12px 20px;border-radius:6px;max-width:400px;cursor:pointer;' +
        'font-size:14px;box-shadow:0 4px 12px rgba(0,0,0,0.3);';
    if (level === 'error') {
        toast.style.background = '#d32f2f';
        toast.style.color = '#fff';
    } else {
        toast.style.background = '#333';
        toast.style.color = '#fff';
    }
    toast.onclick = function () { toast.remove(); };
    document.body.appendChild(toast);
    setTimeout(function () { toast.remove(); }, 10000);
}
```

This gives the user immediate, visible feedback when the SSE connection dies — not hidden in the console. The toast auto-dismisses after 10 seconds and can be clicked away. The status bar also shows the error persistently.

## Files to modify

1. `/Users/jaten/ivy/goivy/webui/static/js/ivyweb_api.js` — replace `connectEvents`/`disconnectEvents` with backoff logic
2. `/Users/jaten/ivy/goivy/webui/static/js/ivyweb_app.js` — wire `onConnectionLost` to status bar
3. `/Users/jaten/ivy/goivy/webui/browser_test.go` — already has `.Timeout(30*time.Second)` (Layer 1)

## Verification

```bash
# Run browser tests multiple times to check for flakiness
cd ~/ivy/goivy && for i in 1 2 3; do echo "=== Run $i ==="; make test-web; done
```

Tests should either pass or fail with a timeout error within 30 seconds — never hang.
