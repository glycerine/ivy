package cmd

import (
	"bytes"
	"fmt"
	"github.com/glycerine/abcgo/aig/ioa"
	"github.com/glycerine/abcgo/sat/satoko"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAutoParameterCartesianProduct(t *testing.T) {
	rows, err := ParseAutoParameters(strings.NewReader("# comment\nC10 S0.95\nR0.99 R0.9\nv\n"))
	if err != nil {
		t.Fatal(err)
	}
	settings, err := CreateAutoSettings(rows)
	if err != nil {
		t.Fatal(err)
	}
	if len(settings) != 4 {
		t.Fatal(len(settings))
	}
	for i, setting := range settings {
		if setting.Options.ConflictLimit != 10 || setting.Options.VariableDecay != .95 {
			t.Fatal(setting)
		}
		if setting.Options.Verbose != (i%2 == 1) {
			t.Fatal(i, setting)
		}
	}
	for _, text := range []string{"", "C 10\n", "C1..2\n", "!\n"} {
		if _, err := ParseAutoParameters(strings.NewReader(text)); err == nil {
			t.Fatalf("accepted %q", text)
		}
	}
}
func TestAutoOptionsMapping(t *testing.T) {
	options, err := deriveAutoOptions(strings.Fields("autotuner -C 11 -P 12 -D 0.7 -E 1.3 -F 5 -G 6 -H 7 -I 8 -J 9 -K 10 -L 11 -M 60 -N 20 -O 12 -Q 13 -R 0.8 -S 0.9 -v"))
	if err != nil {
		t.Fatal(err)
	}
	if options.ConflictLimit != 11 || options.PropagationLimit != 12 || options.FirstReduce != 8 || options.FreezeLBD != 11 || options.LearntRatio != .6 || options.GarbageRatio != .2 || options.VariableDecay != .9 || !options.Verbose {
		t.Fatal(options)
	}
}
func TestAutoParallelCostsAndErrors(t *testing.T) {
	models := []*ioa.Model{{MaxVar: 1}, {MaxVar: 2}, {MaxVar: 3}}
	cost, err := evaluateAuto(models, satoko.DefaultOptions(), 3, func(model *ioa.Model, _ satoko.Options) (int64, error) { return int64(model.MaxVar), nil })
	if err != nil || cost != 6 {
		t.Fatal(cost, err)
	}
	if _, err := evaluateAuto(models, satoko.DefaultOptions(), 2, nil); err == nil {
		t.Fatal("missing backend silently accepted")
	}
	if _, err := evaluateAuto(models, satoko.DefaultOptions(), 2, func(*ioa.Model, satoko.Options) (int64, error) { return math.MaxInt64, nil }); err == nil {
		t.Fatal("overflow accepted")
	}
	if _, err := evaluateAuto(models, satoko.DefaultOptions(), 2, func(*ioa.Model, satoko.Options) (int64, error) { return 0, fmt.Errorf("backend error") }); err == nil {
		t.Fatal("lost backend error")
	}
}
func TestAutoRunSelectsFirstMinimum(t *testing.T) {
	directory := t.TempDir()
	config, list, aig := filepath.Join(directory, "config"), filepath.Join(directory, "list"), filepath.Join(directory, "test.aig")
	if err := os.WriteFile(config, []byte("C1\nS0.7 S0.8\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := ioa.Ioa_WriteAiger(&ioa.Model{MaxVar: 1, Inputs: []uint32{2}, Outputs: []uint32{2}}, aig, 0, 0); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(list, []byte(aig+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	best, err := RunAutoTuner(config, list, 3, &out, func(_ *ioa.Model, options satoko.Options) (int64, error) {
		if options.VariableDecay == .7 {
			return 5, nil
		}
		return 2, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if best.Cost != 2 || best.Setting.Options.VariableDecay != .8 {
		t.Fatal(best, out.String())
	}
}

func TestAutoCommandDefaultMFBackend(t *testing.T) {
	f, out := testFrame(t)
	directory := t.TempDir()
	config, list, aig := filepath.Join(directory, "config"), filepath.Join(directory, "list"), filepath.Join(directory, "benchmark.aig")
	if err := os.WriteFile(config, []byte("C100\nS0.8 S0.9\n"), 0600); err != nil {
		t.Fatal(err)
	}
	// A real mapped cone, with an AND and a complemented output. This runs the
	// production MF coarsening/cut/CNF backend and native Satoko in the command.
	model := &ioa.Model{MaxVar: 3, Inputs: []uint32{2, 4}, ANDs: []ioa.And{{Literal: 6, Left: 4, Right: 2}}, Outputs: []uint32{7}}
	if err := ioa.Ioa_WriteAiger(model, aig, 0, 0); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(list, []byte(aig+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if currentAutoEvaluator() == nil {
		t.Fatal("default MF backend missing")
	}
	command := "autotuner -N 3 -C " + shellQuote(config) + " -F " + shellQuote(list)
	if status := Cmd_CommandExecute(f, command); status != 0 {
		t.Fatal(status, out.String())
	}
	if !strings.Contains(out.String(), "The best settings are:") || strings.Count(out.String(), "Evaluating settings:") != 2 {
		t.Fatal(out.String())
	}
}
