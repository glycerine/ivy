// Command ivyweb starts the Ivy Interactive Verification web UI.
//
// Usage:
//
//	ivyweb                    # listen on :8080, Go backend
//	ivyweb -addr :9090        # listen on port 9090
//	ivyweb -addr 0.0.0.0:80  # listen on all interfaces, port 80
//	ivyweb -open              # open browser automatically
//	ivyweb -py                # use Python Ivy backend only
//	ivyweb -conform           # use both backends, check conformance
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"

	iu "github.com/glycerine/goivy/ivyutils"
	"github.com/glycerine/goivy/webui"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address (host:port)")
	open := flag.Bool("open", false, "open browser automatically")
	usePy := flag.Bool("py", false, "use Python Ivy backend only")
	conform := flag.Bool("conform", false, "send to both Go and Python backends, check conformance")
	flag.Parse()

	if *usePy && *conform {
		log.Fatal("-py and -conform are mutually exclusive")
	}

	cfg := iu.NewConfig()

	var backend webui.Backend
	switch {
	case *conform:
		goBE := webui.NewGoBackend(cfg)
		defer goBE.Close()
		pyBE, err := webui.NewPyBackend(cfg)
		if err != nil {
			log.Fatalf("failed to start Python backend: %v", err)
		}
		defer pyBE.Close()
		backend = webui.NewConformBackend(goBE, pyBE)
		fmt.Printf("IVy: Conformance mode (Go + Python)\n")
	case *usePy:
		pyBE, err := webui.NewPyBackend(cfg)
		if err != nil {
			log.Fatalf("failed to start Python backend: %v", err)
		}
		defer pyBE.Close()
		backend = pyBE
		fmt.Printf("IVy: Python backend\n")
	default:
		backend = webui.NewGoBackend(cfg)
		defer backend.Close()
		fmt.Printf("IVy: Go backend\n")
	}

	url := fmt.Sprintf("http://localhost%s", *addr)
	if (*addr)[0] != ':' {
		url = fmt.Sprintf("http://%s", *addr)
	}

	fmt.Printf("Listening on %s\n", url)

	if *open {
		go openBrowser(url)
	}

	srv := webui.NewServer(cfg, *addr, backend)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
}

// openBrowser tries to open the URL in the default browser.
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		fmt.Fprintf(os.Stderr, "Open %s in your browser\n", url)
		return
	}
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open browser: %v\nOpen %s manually\n", err, url)
	}
}
