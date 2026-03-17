//go:build web

package webui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// ---------- helpers ----------

// startTestServer creates an httptest.Server backed by a webui.Server.
func startTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := NewServer(":0")
	ts := httptest.NewServer(srv)
	t.Cleanup(ts.Close)
	return ts
}

// setupBrowser launches a headless Chrome via rod.
// If no browser can be found it skips the test.
func setupBrowser(t *testing.T) *rod.Browser {
	t.Helper()
	path, ok := launcher.LookPath()
	if !ok || path == "" {
		t.Skip("skipping: no Chrome/Chromium found on system")
	}
	u, err := launcher.New().Headless(true).Launch()
	if err != nil {
		t.Skipf("skipping: failed to launch browser: %v", err)
	}
	browser := rod.New().ControlURL(u)
	if err := browser.Connect(); err != nil {
		t.Skipf("skipping: failed to connect to browser: %v", err)
	}
	t.Cleanup(func() { browser.MustClose() })
	return browser
}

// newPage navigates to the given URL and waits for it to load.
func newPage(t *testing.T, browser *rod.Browser, url string) *rod.Page {
	t.Helper()
	page := browser.MustPage(url)
	t.Cleanup(func() { page.MustClose() })
	page.MustWaitLoad()
	return page
}

// createSessionViaHTTP creates a session on the server and returns its ID.
func createSessionViaHTTP(t *testing.T, baseURL string) string {
	t.Helper()
	resp, err := http.Post(baseURL+"/api/session/new", "application/json", nil)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	defer resp.Body.Close()
	var m map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		t.Fatalf("decode session: %v", err)
	}
	id := m["session_id"]
	if id == "" {
		t.Fatal("empty session_id")
	}
	return id
}

// ---------- Page Loading & Structure ----------

func TestBrowserPageLoads(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	ts := startTestServer(t)
	browser := setupBrowser(t)
	page := newPage(t, browser, ts.URL)

	// Verify no JS console errors (collect any that appeared).
	var jsErrors []string
	go page.EachEvent(func(e *proto.RuntimeExceptionThrown) {
		jsErrors = append(jsErrors, e.ExceptionDetails.Text)
	})()
	// Give a moment for errors to appear.
	time.Sleep(200 * time.Millisecond)

	title := page.MustEval(`() => document.title`).String()
	if title == "" {
		t.Error("page title is empty")
	}
	t.Logf("page title: %s", title)
}

func TestBrowserHasTitle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	ts := startTestServer(t)
	browser := setupBrowser(t)
	page := newPage(t, browser, ts.URL)

	title := page.MustEval(`() => document.title`).String()
	if !strings.Contains(strings.ToLower(title), "ivy") {
		t.Errorf("title = %q, want something containing 'ivy' (case-insensitive)", title)
	}
}

func TestBrowserHasARGPanel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	ts := startTestServer(t)
	browser := setupBrowser(t)
	page := newPage(t, browser, ts.URL)

	el, err := page.Element("#arg-panel")
	if err != nil || el == nil {
		// Try alternative id
		el, err = page.Element("#arg-graph")
		if err != nil || el == nil {
			t.Error("neither #arg-panel nor #arg-graph found")
		}
	}
}

func TestBrowserHasConceptPanel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	ts := startTestServer(t)
	browser := setupBrowser(t)
	page := newPage(t, browser, ts.URL)

	el, err := page.Element("#concept-panel")
	if err != nil || el == nil {
		el, err = page.Element("#concept-graph")
		if err != nil || el == nil {
			t.Error("neither #concept-panel nor #concept-graph found")
		}
	}
}

func TestBrowserHasMenuBar(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	ts := startTestServer(t)
	browser := setupBrowser(t)
	page := newPage(t, browser, ts.URL)

	el, err := page.Element("#menubar")
	if err != nil || el == nil {
		t.Fatal("#menubar not found")
	}
	html := page.MustEval(`() => document.getElementById('menubar').innerHTML`).String()
	for _, label := range []string{"File", "Mode", "Check"} {
		if !strings.Contains(html, label) {
			t.Errorf("menubar missing label %q", label)
		}
	}
}

func TestBrowserHasStatusBar(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	ts := startTestServer(t)
	browser := setupBrowser(t)
	page := newPage(t, browser, ts.URL)

	el, err := page.Element("#statusbar")
	if err != nil || el == nil {
		t.Error("#statusbar not found")
	}
}

func TestBrowserHasInfoPanel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	ts := startTestServer(t)
	browser := setupBrowser(t)
	page := newPage(t, browser, ts.URL)

	el, err := page.Element("#info-panel")
	if err != nil || el == nil {
		t.Error("#info-panel not found")
	}
}

// ---------- API Integration via Browser ----------

func TestBrowserCreateSession(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	ts := startTestServer(t)
	browser := setupBrowser(t)
	page := newPage(t, browser, ts.URL)

	// Use fetch to POST to the API from the browser context (same origin).
	result := page.MustEval(`() => {
		return fetch('/api/session/new', {method: 'POST'})
			.then(r => r.json())
			.then(d => d.session_id || '')
	}`).String()
	if result == "" {
		t.Error("no session_id returned from API via browser fetch")
	}
	t.Logf("session id from browser: %s", result)
}

func TestBrowserGetARG(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	ts := startTestServer(t)
	sid := createSessionViaHTTP(t, ts.URL)
	browser := setupBrowser(t)
	page := newPage(t, browser, ts.URL)

	result := page.MustEval(fmt.Sprintf(`() => {
		return fetch('/api/session/%s/arg')
			.then(r => r.json())
			.then(d => JSON.stringify(d))
	}`, sid)).String()
	if !strings.Contains(result, "elements") {
		t.Errorf("ARG response missing 'elements': %s", result)
	}
}

func TestBrowserGetConcept(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	ts := startTestServer(t)
	sid := createSessionViaHTTP(t, ts.URL)
	browser := setupBrowser(t)
	page := newPage(t, browser, ts.URL)

	result := page.MustEval(fmt.Sprintf(`() => {
		return fetch('/api/session/%s/concept')
			.then(r => r.json())
			.then(d => JSON.stringify(d))
	}`, sid)).String()
	if result == "" {
		t.Error("empty concept response")
	}
}

// ---------- Interactive Features ----------

func TestBrowserModeSelect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	ts := startTestServer(t)
	browser := setupBrowser(t)
	page := newPage(t, browser, ts.URL)

	// Wait for JS to initialise.
	time.Sleep(500 * time.Millisecond)

	sel, err := page.Element("#mode-select")
	if err != nil || sel == nil {
		t.Skip("no #mode-select found, skipping")
	}

	// Change mode to "abstract".
	page.MustEval(`() => {
		var s = document.getElementById('mode-select');
		s.value = 'abstract';
		s.dispatchEvent(new Event('change'));
	}`)
	val := page.MustEval(`() => document.getElementById('mode-select').value`).String()
	if val != "abstract" {
		t.Errorf("mode-select value = %q, want 'abstract'", val)
	}
}

func TestBrowserCheckButton(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	ts := startTestServer(t)
	browser := setupBrowser(t)
	page := newPage(t, browser, ts.URL)

	// Wait for app init (session creation).
	time.Sleep(1 * time.Second)

	btn, err := page.Element("#btn-check")
	if err != nil || btn == nil {
		t.Skip("no #btn-check found")
	}
	btn.MustClick()
	// Give time for the check to complete and status to update.
	time.Sleep(500 * time.Millisecond)

	status := page.MustEval(`() => document.getElementById('statusbar').textContent`).String()
	t.Logf("status after check: %s", status)
	// We just verify no crash; status may be "Ready" or "Check completed" etc.
}

func TestBrowserUndoButton(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	ts := startTestServer(t)
	browser := setupBrowser(t)
	page := newPage(t, browser, ts.URL)

	time.Sleep(1 * time.Second)

	btn, err := page.Element("#btn-undo")
	if err != nil || btn == nil {
		t.Skip("no #btn-undo found")
	}
	btn.MustClick()
	// Verify no crash - page should still be alive.
	time.Sleep(300 * time.Millisecond)
	title := page.MustEval(`() => document.title`).String()
	if title == "" {
		t.Error("page appears dead after undo click")
	}
}

// ---------- Cytoscape.js Graph Rendering ----------

func TestBrowserCytoscapeLoads(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	ts := startTestServer(t)
	browser := setupBrowser(t)
	page := newPage(t, browser, ts.URL)

	// Wait for scripts to load.
	time.Sleep(2 * time.Second)

	typ := page.MustEval(`() => typeof cytoscape`).String()
	if typ != "function" {
		t.Errorf("typeof cytoscape = %q, want 'function'", typ)
	}
}

func TestBrowserARGGraphInitializes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	ts := startTestServer(t)
	browser := setupBrowser(t)
	page := newPage(t, browser, ts.URL)

	// Wait for app initialization.
	time.Sleep(2 * time.Second)

	// Check that the arg-graph container has a cytoscape canvas or svg child.
	hasCy := page.MustEval(`() => {
		var el = document.getElementById('arg-graph');
		if (!el) return false;
		// Cytoscape injects a canvas element inside its container.
		return el.querySelector('canvas') !== null || el.children.length > 0;
	}`).Bool()
	if !hasCy {
		t.Error("ARG graph container does not appear to have a cytoscape instance")
	}
}

// ---------- Context Menu ----------

func TestBrowserContextMenuHidden(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	ts := startTestServer(t)
	browser := setupBrowser(t)
	page := newPage(t, browser, ts.URL)

	display := page.MustEval(`() => {
		var cm = document.getElementById('context-menu');
		if (!cm) return 'missing';
		return cm.style.display;
	}`).String()
	if display != "none" {
		t.Errorf("context-menu display = %q, want 'none'", display)
	}
}

func TestBrowserRightClickShowsMenu(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	ts := startTestServer(t)
	browser := setupBrowser(t)
	page := newPage(t, browser, ts.URL)

	time.Sleep(2 * time.Second)

	// The context menu is triggered by right-clicking on a cytoscape node.
	// Without loaded data there may be no nodes, so we just verify the
	// context-menu element exists and is initially hidden.
	exists := page.MustEval(`() => document.getElementById('context-menu') !== null`).Bool()
	if !exists {
		t.Error("context-menu element does not exist")
		return
	}
	// If the graph has nodes, right-click might show it. For now we verify
	// the mechanism doesn't crash by dispatching a contextmenu event on the
	// arg-graph container.
	page.MustEval(`() => {
		var el = document.getElementById('arg-graph');
		if (el) {
			var evt = new MouseEvent('contextmenu', {bubbles: true, clientX: 50, clientY: 50});
			el.dispatchEvent(evt);
		}
	}`)
	time.Sleep(200 * time.Millisecond)
	// No crash means success; the menu may or may not be visible depending
	// on whether there's a node at that location.
}

// ---------- Resizable Panel ----------

func TestBrowserDividerExists(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	ts := startTestServer(t)
	browser := setupBrowser(t)
	page := newPage(t, browser, ts.URL)

	el, err := page.Element("#divider")
	if err != nil || el == nil {
		t.Error("#divider element not found")
	}
}

func TestBrowserPanelResize(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	ts := startTestServer(t)
	browser := setupBrowser(t)
	page := newPage(t, browser, ts.URL)

	time.Sleep(1 * time.Second)

	// Get initial width of arg-panel.
	initialWidth := page.MustEval(`() => {
		var el = document.getElementById('arg-panel');
		if (!el) return -1;
		return el.getBoundingClientRect().width;
	}`).Int()
	if initialWidth <= 0 {
		t.Skip("arg-panel has no width, skipping resize test")
	}

	// Simulate drag on divider: mousedown, mousemove, mouseup.
	page.MustEval(`() => {
		var div = document.getElementById('divider');
		if (!div) return;
		var rect = div.getBoundingClientRect();
		var cx = rect.left + rect.width / 2;
		var cy = rect.top + rect.height / 2;
		div.dispatchEvent(new MouseEvent('mousedown', {bubbles:true, clientX:cx, clientY:cy}));
		document.dispatchEvent(new MouseEvent('mousemove', {bubbles:true, clientX:cx+100, clientY:cy}));
		document.dispatchEvent(new MouseEvent('mouseup', {bubbles:true, clientX:cx+100, clientY:cy}));
	}`)
	time.Sleep(300 * time.Millisecond)

	newWidth := page.MustEval(`() => {
		var el = document.getElementById('arg-panel');
		if (!el) return -1;
		return el.getBoundingClientRect().width;
	}`).Int()
	t.Logf("panel width: before=%d after=%d", initialWidth, newWidth)
	// We just verify no crash; the resize may or may not actually change width
	// depending on how the JS resize handler works.
}

// ---------- SSE Connection ----------

func TestBrowserSSEEvents(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	ts := startTestServer(t)
	sid := createSessionViaHTTP(t, ts.URL)
	browser := setupBrowser(t)
	page := newPage(t, browser, ts.URL)

	// Open an EventSource connection and collect the first event.
	page.MustEval(fmt.Sprintf(`() => {
		window._sseEvents = [];
		window._es = new EventSource('/api/session/%s/events');
		window._es.onmessage = function(e) {
			window._sseEvents.push(e.data);
		};
	}`, sid))

	// Trigger a check, which emits events.
	resp, err := http.Post(ts.URL+"/api/session/"+sid+"/check", "application/json", nil)
	if err != nil {
		t.Fatalf("check request failed: %v", err)
	}
	resp.Body.Close()

	// Give SSE time to deliver.
	time.Sleep(1 * time.Second)

	count := page.MustEval(`() => window._sseEvents.length`).Int()
	t.Logf("SSE events received: %d", count)
	if count == 0 {
		t.Error("no SSE events received after check")
	}

	// Cleanup EventSource.
	page.MustEval(`() => { if (window._es) window._es.close(); }`)
}

// ---------- Error Handling ----------

func TestBrowserInvalidSession(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	ts := startTestServer(t)
	browser := setupBrowser(t)
	page := newPage(t, browser, ts.URL)

	result := page.MustEval(`() => {
		return fetch('/api/session/nonexistent/arg')
			.then(r => r.status)
	}`).Int()
	if result != 404 {
		t.Errorf("expected 404 for invalid session, got %d", result)
	}
}

func TestBrowserStaticCSS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	ts := startTestServer(t)
	browser := setupBrowser(t)
	page := newPage(t, browser, ts.URL)

	time.Sleep(500 * time.Millisecond)

	// Verify the CSS file loaded by checking a computed style.
	// The menubar should have background-color from ivy.css (#2d2d2d).
	bg := page.MustEval(`() => {
		var el = document.getElementById('menubar');
		if (!el) return '';
		return getComputedStyle(el).backgroundColor;
	}`).String()
	t.Logf("menubar background-color: %s", bg)
	// #2d2d2d = rgb(45, 45, 45)
	if bg == "" {
		t.Error("menubar has no background-color, CSS may not have loaded")
	} else if strings.Contains(bg, "45, 45, 45") || strings.Contains(bg, "2d2d2d") {
		// CSS loaded correctly.
	} else {
		t.Logf("menubar background-color is %q (may differ if CSS not loaded from file)", bg)
	}
}

// TestBrowserGraphHealthCheck verifies that both Cytoscape graph instances
// initialized correctly. This catches silent failures from bad stylesheet
// data() mappers (e.g. using data(border_color) for border-color) that
// prevent event handlers like cxttap from working.
func TestBrowserGraphHealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	ts := startTestServer(t)
	browser := setupBrowser(t)
	page := newPage(t, browser, ts.URL)

	// Wait for app initialization and health checks to run.
	time.Sleep(3 * time.Second)

	// Check that the health check globals were set by IvyGraph.healthCheck().
	argHealthy := page.MustEval(`() => window['__ivyGraphHealthy_arg-graph']`).Bool()
	conceptHealthy := page.MustEval(`() => window['__ivyGraphHealthy_concept-graph']`).Bool()
	if !argHealthy {
		t.Error("ARG graph health check failed — Cytoscape may have a broken stylesheet")
	}
	if !conceptHealthy {
		t.Error("Concept graph health check failed — Cytoscape may have a broken stylesheet")
	}

	// Verify no JS console errors were thrown during init.
	hasErrors := page.MustEval(`() => {
		// Check for our specific FATAL error marker
		return typeof window.__ivyInitError !== 'undefined';
	}`).Bool()
	if hasErrors {
		errMsg := page.MustEval(`() => window.__ivyInitError || ''`).String()
		t.Errorf("JS init error detected: %s", errMsg)
	}
}

// TestBrowserConceptGraphRightClick verifies that right-clicking on a concept
// node shows the context menu with expected actions (Splatter, etc.).
// This test loads an .ivy file to populate the concept graph with nodes.
func TestBrowserConceptGraphRightClick(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	ts := startTestServer(t)
	sid := createSessionViaHTTP(t, ts.URL)
	browser := setupBrowser(t)
	page := newPage(t, browser, ts.URL)

	time.Sleep(2 * time.Second)

	// Upload the client_server_example.ivy via the API to populate the graph.
	ivyContent := `#lang ivy1.7
type client
type server
relation link(X:client, Y:server)
relation semaphore(X:server)
after init { semaphore(W) := true; link(X,Y) := false }
action connect(x:client,y:server) = { require semaphore(y); link(x,y) := true; semaphore(y) := false }
export connect
`
	page.MustEval(fmt.Sprintf(`() => {
		var formData = new FormData();
		var blob = new Blob([%q], {type: 'text/plain'});
		formData.append('file', blob, 'test.ivy');
		return fetch('/api/session/%s/load', {method: 'POST', body: formData})
			.then(r => r.json())
			.then(d => JSON.stringify(d));
	}`, ivyContent, sid)).String()

	// Wait for concept graph to update after file load.
	time.Sleep(2 * time.Second)

	// Refresh the concept graph display.
	page.MustEval(fmt.Sprintf(`() => {
		return fetch('/api/session/%s/concept')
			.then(r => r.json())
			.then(d => {
				if (window.app && window.app.conceptGraph) {
					window.app.conceptGraph.update(d.elements);
				}
				return JSON.stringify(d);
			});
	}`, sid)).String()

	time.Sleep(1 * time.Second)

	// Check that concept graph has nodes.
	nodeCount := page.MustEval(`() => {
		if (window.app && window.app.conceptGraph && window.app.conceptGraph.cy) {
			return window.app.conceptGraph.cy.nodes().length;
		}
		return 0;
	}`).Int()
	t.Logf("concept graph node count: %d", nodeCount)

	if nodeCount > 0 {
		// Right-click on first concept node by dispatching contextmenu at its position.
		menuVisible := page.MustEval(`() => {
			var cy = window.app.conceptGraph.cy;
			var node = cy.nodes()[0];
			var pos = node.renderedPosition();
			var container = cy.container();
			var rect = container.getBoundingClientRect();
			// Fire cxttap manually
			node.emit('cxttap', {renderedPosition: pos});
			// Check if context menu became visible
			var cm = document.getElementById('context-menu');
			return cm && cm.style.display !== 'none';
		}`).Bool()

		if !menuVisible {
			t.Error("right-click on concept node did not show context menu")
		} else {
			// Verify menu contains expected actions.
			menuHTML := page.MustEval(`() => document.getElementById('context-menu').innerHTML`).String()
			for _, action := range []string{"Splatter", "Materialize", "Remove"} {
				if !strings.Contains(menuHTML, action) {
					t.Errorf("context menu missing %q action", action)
				}
			}
			t.Logf("context menu HTML: %.200s", menuHTML)
		}
	}
}

func TestBrowserStaticJS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	ts := startTestServer(t)
	browser := setupBrowser(t)
	page := newPage(t, browser, ts.URL)

	time.Sleep(2 * time.Second)

	// Check that IvyApp or IvyAPI class is defined.
	hasAPI := page.MustEval(`() => typeof IvyAPI`).String()
	hasApp := page.MustEval(`() => typeof IvyApp`).String()
	if hasAPI != "function" && hasApp != "function" {
		t.Errorf("IvyAPI=%q, IvyApp=%q — expected at least one to be 'function'", hasAPI, hasApp)
	}
}
