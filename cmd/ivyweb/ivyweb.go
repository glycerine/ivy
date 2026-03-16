// Command ivyweb starts the Ivy Interactive Verification web UI.
//
// Usage:
//
//	ivyweb                    # listen on :8080
//	ivyweb -addr :9090        # listen on port 9090
//	ivyweb -addr 0.0.0.0:80  # listen on all interfaces, port 80
//	ivyweb -open              # open browser automatically
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/glycerine/goivy/webui"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address (host:port)")
	open := flag.Bool("open", false, "open browser automatically")
	flag.Parse()

	url := fmt.Sprintf("http://localhost%s", *addr)
	if (*addr)[0] != ':' {
		url = fmt.Sprintf("http://%s", *addr)
	}

	fmt.Printf("IVy: Interactive Verification\n")
	fmt.Printf("Listening on %s\n", url)

	if *open {
		go openBrowser(url)
	}

	srv := webui.NewServer(*addr)
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
