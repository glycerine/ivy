package scl

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func writeBufferBLIF(t *testing.T, net TimingNetwork, path string) {
	t.Helper()
	var text strings.Builder
	names := make(map[int]string)
	for _, id := range net.SclObjectIDs() {
		names[id] = fmt.Sprintf("n%d", id)
	}
	for i, id := range net.SclInputs() {
		names[id] = fmt.Sprintf("i%d", i)
	}
	for i, output := range net.SclOutputs() {
		names[net.SclFanins(output)[0]] = fmt.Sprintf("o%d", i)
	}
	fmt.Fprintln(&text, ".model buffering")
	fmt.Fprint(&text, ".inputs")
	for _, id := range net.SclInputs() {
		fmt.Fprintf(&text, " %s", names[id])
	}
	fmt.Fprint(&text, "\n.outputs")
	for i := range net.SclOutputs() {
		fmt.Fprintf(&text, " o%d", i)
	}
	fmt.Fprintln(&text)
	order, err := bufferOrder(net)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range order {
		if !net.SclIsNode(id) {
			continue
		}
		cell := bufferCell(net, id)
		if cell == nil {
			t.Fatalf("missing cell %d", id)
		}
		fmt.Fprintf(&text, ".gate %s", cell.Name)
		for pin, fanin := range net.SclFanins(id) {
			if bufferPhase(net, id, pin) {
				t.Fatal("reference BLIF fixture must use explicit phases")
			}
			fmt.Fprintf(&text, " %s=%s", cell.Pins[pin].Name, names[fanin])
		}
		fmt.Fprintf(&text, " %s=%s\n", cell.Pins[cell.Inputs].Name, names[id])
	}
	fmt.Fprintln(&text, ".end")
	if err = os.WriteFile(path, []byte(text.String()), 0600); err != nil {
		t.Fatal(err)
	}
}
func TestLoadGainBufferingAgainstOriginalABC(t *testing.T) {
	binary, err := filepath.Abs("../../../abc")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(binary); err != nil {
		t.Skip("original ABC executable unavailable")
	}
	before := bufferTestFixture(15)
	// Real mapped BLIF has a distinct output wire for each port. Keep those
	// explicit unary gates in both networks so the C and Go inputs match.
	for _, output := range before.outputs {
		buffer, e := before.SclCreateNode([]int{5}, 0)
		if e != nil {
			t.Fatal(e)
		}
		if e = before.SclPatchFanin(output, 0, buffer); e != nil {
			t.Fatal(e)
		}
	}
	params := SC_BusPars{GainRatio: 300, Slew: 10, Degree: 3, AddBuffers: true}
	native, err := Abc_SclBufferingPerform(before, before.mapping.Lib, &params)
	if err != nil {
		t.Fatal(err)
	}
	if len(native.SclObjectIDs()) <= len(before.SclObjectIDs()) {
		t.Fatal("native buffering added no stages")
	}
	dir := t.TempDir()
	lib, input, output, nativePath := filepath.Join(dir, "test.lib"), filepath.Join(dir, "input.blif"), filepath.Join(dir, "reference.blif"), filepath.Join(dir, "native.blif")
	if err = Abc_SclWriteLiberty(lib, before.mapping.Lib); err != nil {
		t.Fatal(err)
	}
	writeBufferBLIF(t, before, input)
	writeBufferBLIF(t, native, nativePath)
	command := "read_lib " + lib + "; read_blif " + input + "; buffer -N 3 -G 300 -S 10; write_blif " + output + "; cec " + nativePath + " " + output
	log, err := exec.Command(binary, "-q", command).CombinedOutput()
	if err != nil {
		t.Fatalf("original ABC buffer: %v\n%s", err, log)
	}
	if _, err = os.Stat(output); err != nil {
		t.Fatalf("original buffer produced no result: %v\n%s", err, log)
	}
	originalData, err := os.ReadFile(input)
	if err != nil {
		t.Fatal(err)
	}
	referenceData, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Count(referenceData, []byte(".gate ")) <= bytes.Count(originalData, []byte(".gate ")) {
		t.Fatal("original C buffer inserted no real stages")
	}
	if got, want := len(sclMappedNodes(native)), bytes.Count(referenceData, []byte(".gate ")); got != want {
		t.Fatalf("buffer stage count native=%d original C=%d", got, want)
	}
	if !bytes.Contains(log, []byte("Networks are equivalent")) {
		t.Fatalf("original/native buffering mismatch:\n%s", log)
	}
}
