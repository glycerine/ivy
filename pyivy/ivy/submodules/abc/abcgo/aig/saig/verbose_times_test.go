package saig

import (
	"bytes"
	"io"
	"os"
	"regexp"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func captureSaigStdout(t *testing.T, action func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout := os.Stdout
	os.Stdout = w
	done := make(chan []byte)
	go func() {
		var b bytes.Buffer
		io.Copy(&b, r)
		done <- b.Bytes()
	}()
	func() {
		defer func() { os.Stdout = stdout }()
		action()
	}()
	w.Close()
	return string(<-done)
}

// Raig_ManSimulate -v ends with the memory line and ABC_PRT( "Total time" ).
func TestRaigSimulateVerboseTimeLikeABC(t *testing.T) {
	graph := aig.NewGraph()
	x, unused, state := graph.NewInput(), graph.NewInput(), graph.NewInput()
	_ = unused
	graph.AddOutput(graph.NewXor(state, x))
	graph.AddOutput(state.Not())
	aig.Aig_ManSetRegNum(graph, 1)
	out := captureSaigStdout(t, func() { Raig_ManSimulate(graph, 1, 8, 0, 0, 1) })
	if !regexp.MustCompile(`Maxcut = +[0-9]+\.  AigMem = +[0-9.]+ MB\.  SimMem = +[0-9.]+ MB\.  Total time = +[0-9]+\.[0-9]{2} sec\n$`).MatchString(out) {
		t.Errorf("Raig_ManSimulate -v:\n%q", out)
	}
}
