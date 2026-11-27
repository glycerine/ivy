// Command abc is the Go port of ABC's interactive shell (src/base/main/main.c).
// It registers command modules in the order of Abc_FrameInit.
package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/base/abci"
	"github.com/glycerine/abcgo/base/bac"
	"github.com/glycerine/abcgo/base/cba"
	"github.com/glycerine/abcgo/base/cmd"
	abcio "github.com/glycerine/abcgo/base/io"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/base/pla"
	"github.com/glycerine/abcgo/base/test"
	_ "github.com/glycerine/abcgo/base/ver" // registers the Verilog parser for read
	"github.com/glycerine/abcgo/base/wlc"
	abcif "github.com/glycerine/abcgo/map/if"
	"github.com/glycerine/abcgo/map/mapper"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/map/scl"
	"github.com/glycerine/abcgo/map/super"
	"github.com/glycerine/abcgo/sat/glucose"
	"github.com/glycerine/abcgo/sat/glucose2"
)

func init() {
	for _, f := range []func(*abcmain.Frame){
		abcio.Io_Init, abci.Abc_Init, abcif.If_Init, mapper.Map_Init, mio.Mio_Init,
		super.Super_Init, cmd.Load_Init, scl.Scl_Init, wlc.Wlc_Init, bac.Bac_Init,
		cba.Cba_Init, pla.Pla_Init, gia.Sim_Init, test.Test_Init,
		glucose.Glucose_Init, glucose2.Glucose2_Init,
	} {
		abcmain.Abc_FrameAddInitializer(&abcmain.Abc_FrameInitializer_t{Init: f})
	}
}

// Setting ABCGO_CPUPROFILE=file writes a Go CPU profile of the whole run;
// with it, ABCGO_MEMPROFILE=file also writes an allocation profile.
func main() {
	// ABC is sequential; see misc/util/procs.go.
	if os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(1)
	}
	setGCDefaults()
	if path := os.Getenv("ABCGO_CPUPROFILE"); path != "" {
		f, err := os.Create(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		status := abcmain.Run(os.Args)
		pprof.StopCPUProfile()
		f.Close()
		if mem := os.Getenv("ABCGO_MEMPROFILE"); mem != "" {
			if m, err := os.Create(mem); err == nil {
				pprof.Lookup("allocs").WriteTo(m, 0)
				m.Close()
			}
		}
		os.Exit(status)
	}
	os.Exit(abcmain.Run(os.Args))
}
