//go:build web

package webui

import (
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

func TestRodTimingBreakdown(t *testing.T) {
	t0 := time.Now()
	path, ok := launcher.LookPath()
	t1 := time.Now()
	vv("LookPath: %v (path=%s, ok=%v)", t1.Sub(t0), path, ok)
	if !ok {
		t.Skip("no browser")
	}

	l := launcher.New().Headless(true)
	t2 := time.Now()
	vv("launcher.New: %v", t2.Sub(t1))

	u, err := l.Launch()
	t3 := time.Now()
	vv("Launch: %v (err=%v)", t3.Sub(t2), err)
	if err != nil {
		t.Skip(err)
	}

	browser := rod.New().ControlURL(u)
	t4 := time.Now()
	vv("rod.New+ControlURL: %v", t4.Sub(t3))

	err = browser.Connect()
	t5 := time.Now()
	vv("Connect: %v (err=%v)", t5.Sub(t4), err)
	if err != nil {
		t.Skip(err)
	}

	// Now test page creation
	ts := startTestServer(t)
	t6 := time.Now()
	vv("startTestServer: %v", t6.Sub(t5))

	page := browser.MustPage(ts.URL)
	t7 := time.Now()
	vv("MustPage: %v", t7.Sub(t6))

	page.MustWaitLoad()
	t8 := time.Now()
	vv("MustWaitLoad: %v", t8.Sub(t7))

	// Second page to see if reuse is faster
	page2 := browser.MustPage(ts.URL)
	t9 := time.Now()
	vv("MustPage(2nd): %v", t9.Sub(t8))

	page2.MustWaitLoad()
	t10 := time.Now()
	vv("MustWaitLoad(2nd): %v", t10.Sub(t9))

	page2.MustClose()
	page.MustClose()
	browser.MustClose()
	t11 := time.Now()
	vv("Cleanup: %v", t11.Sub(t10))
	vv("TOTAL: %v", t11.Sub(t0))
}
