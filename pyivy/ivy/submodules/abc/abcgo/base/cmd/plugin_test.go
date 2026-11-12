package cmd

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/ioa"
	abcmain "github.com/glycerine/abcgo/base/main"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPluginProtocolParsers(t *testing.T) {
	file := filepath.Join(t.TempDir(), "result")
	if err := os.WriteFile(file, []byte("result:\tproved\nbug-free-depth: 12\nabstraction: 01?10\ncounter-example: 10\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if got := Abc_ManReadStatus(file, "result:"); got != 1 {
		t.Fatal(got)
	}
	if got := Abc_ManReadInteger(file, "bug-free-depth:"); got != 12 {
		t.Fatal(got)
	}
	if got := fmt.Sprint(Abc_ManReadBinary(file, "abstraction:")); got != "[0 1 0 1 0]" {
		t.Fatal(got)
	}
}
func TestPluginAigEncoding(t *testing.T) {
	model := &ioa.Model{MaxVar: 1, Inputs: []uint32{2}, Outputs: []uint32{2}}
	data, err := ioa.Marshal(model, ioa.WriteOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var encoded strings.Builder
	fmt.Fprintf(&encoded, "%d ", len(data))
	for i := 0; i < len(data); i += 3 {
		var b [3]byte
		copy(b[:], data[i:])
		for _, v := range []byte{b[0] & 63, b[0]>>6 | (b[1]&15)<<2, b[1]>>4 | (b[2]&3)<<4, b[2] >> 2} {
			encoded.WriteByte(v + '0')
		}
	}
	got, err := DecodePluginAIG(encoded.String())
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Inputs) != 1 || got.Outputs[0] != 2 {
		t.Fatal(got)
	}
	if _, err := DecodePluginAIG("999 0000"); err == nil {
		t.Fatal("accepted truncated payload")
	}
}
func TestPluginWitnessExpansionAndVerification(t *testing.T) {
	model := &ioa.Model{MaxVar: 2, Inputs: []uint32{2, 4}, Outputs: []uint32{4}}
	witness, err := verifyPluginCex(model, []int32{1})
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(witness.Bits) != "[0 1]" || witness.Output != 0 {
		t.Fatal(witness)
	}
	if _, err := verifyPluginCex(model, []int32{0, 0}); err == nil {
		t.Fatal("accepted nonfailing witness")
	}
	constant := &ioa.Model{MaxVar: 1, Inputs: []uint32{2}, Outputs: []uint32{0}}
	if _, err := expandPluginCex(constant, []int32{1}); err == nil {
		t.Fatal("would loop with no used PIs")
	}
}

type testPluginBridge struct {
	model   *ioa.Model
	witness *PluginCounterexample
}

func (b *testPluginBridge) Export(_ *abcmain.Frame, format string) (any, error) {
	if format == "gia-aiger" {
		return b.model, nil
	}
	return nil, fmt.Errorf("unexpected export %s", format)
}
func (b *testPluginBridge) Import(_ *abcmain.Frame, format string, value any) error {
	switch format {
	case "gia-counterexample":
		b.witness = value.(*PluginCounterexample)
	case "gia-aiger":
		b.model = value.(*ioa.Model)
	default:
		return fmt.Errorf("unexpected import %s", format)
	}
	return nil
}
func TestPluginProcessAndBridge(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX executable fixture")
	}
	f, out := testFrame(t)
	bridge := &testPluginBridge{model: &ioa.Model{MaxVar: 1, Inputs: []uint32{2}, Outputs: []uint32{2}}}
	f.Bridge = bridge
	binary := filepath.Join(t.TempDir(), "fake plugin")
	script := "#!/bin/sh\nif [ \"$2\" = -list-commands ]; then printf 'prove\\n'; exit 0; fi\nfor arg in \"$@\"; do case \"$arg\" in -output=*) output=${arg#-output=};; esac; done\nprintf 'result: failed\\nbug-free-depth: 0\\ncounter-example: 1\\n' > \"$output\"\n"
	if err := os.WriteFile(binary, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	if status := Cmd_CommandAbcLoadPlugIn(f, 3, []string{"load_plugin", binary, "test"}); status != 0 {
		t.Fatal(status, out.String())
	}
	if status := Cmd_CommandExecute(f, "prove"); status != 0 {
		t.Fatal(status, out.String())
	}
	if f.Status != 0 || bridge.witness == nil || bridge.witness.Output != 0 {
		t.Fatal(f.Status, bridge.witness, out.String())
	}
}
func TestPluginCounterexampleBridgeData(t *testing.T) {
	witness := &PluginCounterexample{Bits: []int32{0, 1}, Registers: 1, Inputs: 1, Frame: 0, Output: 2}
	bits, registers, inputs, frame, output := witness.CounterexampleData()
	bits[0] = 1
	if witness.Bits[0] != 0 || registers != 1 || inputs != 1 || frame != 0 || output != 2 {
		t.Fatal("bridge data aliased or corrupted")
	}
	var empty *PluginCounterexample
	if bits, registers, inputs, frame, output := empty.CounterexampleData(); bits != nil || registers != 0 || inputs != 0 || frame != 0 || output != 0 {
		t.Fatal("nil witness not safe")
	}
}
