//go:build web

package webui

import (
	goivy "github.com/glycerine/ivy/goivy"
	"math/rand"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-rod/rod"
)

// setupBrowserB returns the shared browser for benchmarks.
func setupBrowserB(b *testing.B) *rod.Browser {
	b.Helper()
	sharedBrowserOnce.Do(initSharedBrowser)
	if sharedBrowserErr != "" {
		b.Skipf("skipping: %s", sharedBrowserErr)
	}
	return sharedBrowser
}

// startTestServerB creates an httptest.Server for benchmarks.
func startTestServerB(b *testing.B) *httptest.Server {
	b.Helper()
	cfg := goivy.NewConfig()
	srv := NewServer(cfg, ":0")
	ts := httptest.NewServer(srv)
	b.Cleanup(ts.Close)
	return ts
}

// BenchmarkRodPageLifecycle measures the per-test overhead:
// incognito context creation, page load, MustWaitStable, and teardown.
func BenchmarkRodPageLifecycle(b *testing.B) {
	browser := setupBrowserB(b)
	ts := startTestServerB(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		incognito := browser.MustIncognito()
		page := incognito.MustPage(ts.URL)
		page.MustWaitLoad()
		page.MustWaitStable()
		page.MustClose()
		incognito.MustClose()
	}
}

// BenchmarkRodOperations loads the app and runs random UI operations,
// simulating a typical interactive session. Use with:
//
//	go test -tags web -bench BenchmarkRodOperations -benchtime 30s -count 1 ./webui/
//	go test -tags web -bench BenchmarkRodOperations -cpuprofile cpu.prof -memprofile mem.prof ./webui/
func BenchmarkRodOperations(b *testing.B) {
	browser := setupBrowserB(b)
	ts := startTestServerB(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		incognito := browser.MustIncognito()
		page := incognito.MustPage(ts.URL)
		page.MustWaitLoad()
		page.MustWaitStable()

		runRandomOps(b, page, 10)

		page.MustClose()
		incognito.MustClose()
	}
}

// BenchmarkRodOpsOnLivePage measures throughput of operations on
// an already-loaded page (no page creation overhead per iteration).
func BenchmarkRodOpsOnLivePage(b *testing.B) {
	browser := setupBrowserB(b)
	ts := startTestServerB(b)

	incognito := browser.MustIncognito()
	page := incognito.MustPage(ts.URL)
	page.MustWaitLoad()
	page.MustWaitStable()
	b.Cleanup(func() {
		if err := page.Close(); err != nil {
			b.Logf("cleanup: page.Close: %v", err)
		}
		if err := incognito.Close(); err != nil {
			b.Logf("cleanup: incognito.Close: %v", err)
		}
	})

	// Create a session so the app is fully initialized.
	_ = page.MustEval(`() =>
		fetch('/api/session/new', {method:'POST'})
			.then(r => r.json())
			.then(d => d.session_id || '')
	`).String()
	page.MustWaitStable()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runRandomOps(b, page, 1)
	}
}

// BenchmarkRodSubOperations runs each operation type as a sub-benchmark
// so you can see exactly which operations are slow.
func BenchmarkRodSubOperations(b *testing.B) {
	browser := setupBrowserB(b)
	ts := startTestServerB(b)

	incognito := browser.MustIncognito()
	page := incognito.MustPage(ts.URL)
	page.MustWaitLoad()
	page.MustWaitStable()
	b.Cleanup(func() {
		if err := page.Close(); err != nil {
			b.Logf("cleanup: page.Close: %v", err)
		}
		if err := incognito.Close(); err != nil {
			b.Logf("cleanup: incognito.Close: %v", err)
		}
	})

	// Create session.
	_ = page.MustEval(`() =>
		fetch('/api/session/new', {method:'POST'})
			.then(r => r.json())
			.then(d => d.session_id || '')
	`).String()
	page.MustWaitStable()

	for _, o := range benchOps {
		b.Run(o.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				o.fn(page)
			}
		})
	}
}

// ---------- operation table ----------

type benchOp struct {
	name string
	fn   func(page *rod.Page)
}

var benchOps = []benchOp{
	{"eval_title", func(p *rod.Page) {
		p.MustEval(`() => document.title`)
	}},
	{"typeof_cytoscape", func(p *rod.Page) {
		p.MustEval(`() => typeof cytoscape`)
	}},
	{"typeof_IvyApp", func(p *rod.Page) {
		p.MustEval(`() => typeof IvyApp`)
	}},
	{"query_menubar", func(p *rod.Page) {
		p.MustEval(`() => document.getElementById('menubar') !== null`)
	}},
	{"query_statusbar", func(p *rod.Page) {
		p.MustEval(`() => document.getElementById('statusbar') !== null`)
	}},
	{"query_arg_panel", func(p *rod.Page) {
		p.MustEval(`() => (document.getElementById('arg-panel') || document.getElementById('arg-graph')) !== null`)
	}},
	{"query_concept_panel", func(p *rod.Page) {
		p.MustEval(`() => (document.getElementById('concept-panel') || document.getElementById('concept-graph')) !== null`)
	}},
	{"query_context_menu", func(p *rod.Page) {
		p.MustEval(`() => {
			var cm = document.getElementById('context-menu');
			return cm ? cm.style.display : 'missing';
		}`)
	}},
	{"query_divider", func(p *rod.Page) {
		p.MustEval(`() => document.getElementById('divider') !== null`)
	}},
	{"query_info_panel", func(p *rod.Page) {
		p.MustEval(`() => document.getElementById('info-panel') !== null`)
	}},
	{"computed_style_menubar", func(p *rod.Page) {
		p.MustEval(`() => {
			var el = document.getElementById('menubar');
			return el ? getComputedStyle(el).backgroundColor : '';
		}`)
	}},
	{"fetch_session_new", func(p *rod.Page) {
		p.MustEval(`() => fetch('/api/session/new', {method:'POST'}).then(r => r.json()).then(d => d.session_id || '')`)
	}},
	{"fetch_arg", func(p *rod.Page) {
		p.MustEval(`() => {
			if (!window.app || !window.app.sessionId) return 'no-session';
			return fetch('/api/session/' + window.app.sessionId + '/arg').then(r => r.text());
		}`)
	}},
	{"fetch_concept", func(p *rod.Page) {
		p.MustEval(`() => {
			if (!window.app || !window.app.sessionId) return 'no-session';
			return fetch('/api/session/' + window.app.sessionId + '/concept').then(r => r.text());
		}`)
	}},
	{"fetch_toggles", func(p *rod.Page) {
		p.MustEval(`() => {
			if (!window.app || !window.app.sessionId) return 'no-session';
			return fetch('/api/session/' + window.app.sessionId + '/toggles').then(r => r.text());
		}`)
	}},
	{"click_mode_select", func(p *rod.Page) {
		p.MustEval(`() => {
			var s = document.getElementById('mode-select');
			if (!s) return 'missing';
			var modes = ['induction', 'abstract', 'bounded'];
			s.value = modes[Math.floor(Math.random() * modes.length)];
			s.dispatchEvent(new Event('change'));
			return s.value;
		}`)
	}},
	{"dispatch_contextmenu", func(p *rod.Page) {
		p.MustEval(`() => {
			var el = document.getElementById('arg-graph');
			if (!el) return 'missing';
			el.dispatchEvent(new MouseEvent('contextmenu', {bubbles:true, clientX:50, clientY:50}));
			return 'ok';
		}`)
	}},
	{"dispatch_resize", func(p *rod.Page) {
		p.MustEval(`() => {
			var div = document.getElementById('divider');
			if (!div) return 'missing';
			var rect = div.getBoundingClientRect();
			var cx = rect.left + rect.width/2, cy = rect.top + rect.height/2;
			var dx = (Math.random() - 0.5) * 200;
			div.dispatchEvent(new MouseEvent('mousedown', {bubbles:true, clientX:cx, clientY:cy}));
			document.dispatchEvent(new MouseEvent('mousemove', {bubbles:true, clientX:cx+dx, clientY:cy}));
			document.dispatchEvent(new MouseEvent('mouseup', {bubbles:true, clientX:cx+dx, clientY:cy}));
			return 'ok';
		}`)
	}},
	{"wait_stable", func(p *rod.Page) {
		p.MustWaitStable()
	}},
}

func runRandomOps(b *testing.B, page *rod.Page, n int) {
	b.Helper()
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for j := 0; j < n; j++ {
		o := benchOps[rng.Intn(len(benchOps))]
		o.fn(page)
	}
}
