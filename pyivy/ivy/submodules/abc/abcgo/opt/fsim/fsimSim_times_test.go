package fsim

import (
	"io"
	"os"
	"regexp"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

// Verbose simulation ends with Fsim_ManSimulate's "Maxcut = ..." line and
// ABC_PRT( "Sim time", ... ); the port printed "FSIM: ... elapsed=1.2ms".
func TestRunSimulationVerboseLikeABC(t *testing.T) {
	g := aig.NewGraph()
	a, q := g.NewInput(), g.NewInput()
	g.AddOutput(q)
	g.AddOutput(a)
	aig.Aig_ManSetRegNum(g, 1)
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout := os.Stdout
	os.Stdout = w
	done := make(chan string)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()
	_, err = RunSimulation(g, &Fsim_ParSim_t{Words: 1, Iterations: 2, Verbose: true})
	os.Stdout = stdout
	w.Close()
	out := <-done
	if err != nil {
		t.Fatal(err)
	}
	want := regexp.MustCompile(`^Maxcut = +[0-9]+\.  AigMem = +[0-9]+\.[0-9][0-9] MB\.  SimMem = +[0-9]+\.[0-9][0-9] MB\.  Sim time = +[0-9]+\.[0-9][0-9] sec\n$`)
	if !want.MatchString(out) {
		t.Fatalf("got %q", out)
	}
}
