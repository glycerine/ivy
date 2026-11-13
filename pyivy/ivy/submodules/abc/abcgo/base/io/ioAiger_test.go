package io

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/ioa"
	"github.com/glycerine/abcgo/base/abc"
)

func TestAigerReadWriteReferenceFixtures(t *testing.T) {
	for _, filename := range []string{"aiger_ascii_extended.aag", "aiger_binary_extended.aig"} {
		t.Run(filename, func(t *testing.T) {
			path := filepath.Join("test_vectors", filename)
			network, err := ReadAiger(path, true)
			if err != nil {
				t.Fatal(err)
			}
			if abc.Abc_NtkPiNum(network) != 1 || abc.Abc_NtkLatchNum(network) != 1 || abc.Abc_NtkPoNum(network) != 3 {
				t.Fatalf("unexpected network shape: PI=%d latch=%d PO=%d", abc.Abc_NtkPiNum(network), abc.Abc_NtkLatchNum(network), abc.Abc_NtkPoNum(network))
			}
			if network.ObjectsOfKind(abc.ObjectPI)[0].Name() != "enable" || network.ObjectsOfKind(abc.ObjectPO)[0].Name() != "out" {
				t.Fatal("symbol names were not attached to ABC objects")
			}
			original, err := aigerModelFromNetwork(network)
			if err != nil {
				t.Fatal(err)
			}
			if len(original.Bad) != 1 || len(original.Constraints) != 1 || len(original.Justice) != 0 {
				t.Fatalf("AIGER auxiliary properties were not retained: %+v", original)
			}
			for _, options := range []ioa.WriteOptions{{Symbols: true}, {ASCII: true, Symbols: true}} {
				encoded, err := ioa.Marshal(original, options)
				if err != nil {
					t.Fatal(err)
				}
				decoded, err := ioa.Read(encoded)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(original, decoded) {
					t.Fatalf("AIGER roundtrip changed model:\n%+v\n%+v", original, decoded)
				}
			}
		})
	}
}

func TestAigerFileCompressionAndMalformedInput(t *testing.T) {
	network, err := ReadAiger(filepath.Join("test_vectors", "aiger_ascii_extended.aag"), true)
	if err != nil {
		t.Fatal(err)
	}
	for _, suffix := range []string{".aig.gz", ".aig.bz2"} {
		path := filepath.Join(t.TempDir(), "roundtrip"+suffix)
		if err = WriteAigerWithOptions(network, path, ioa.WriteOptions{Symbols: true}); err != nil {
			t.Fatal(err)
		}
		decoded, err := ReadAiger(path, true)
		if err != nil {
			t.Fatal(err)
		}
		before, _ := aigerModelFromNetwork(network)
		after, _ := aigerModelFromNetwork(decoded)
		if !reflect.DeepEqual(before, after) {
			t.Fatalf("compressed roundtrip differs for %s", suffix)
		}
	}
	bad := filepath.Join(t.TempDir(), "bad.aig")
	if err = os.WriteFile(bad, []byte("aig 1 1 0 1 0\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = ReadAiger(bad, true); err == nil {
		t.Fatal("truncated AIGER file was accepted")
	}
}

func TestAigerNativeDispatchAndCounterexampleTrace(t *testing.T) {
	network, err := ReadNetwork(filepath.Join("test_vectors", "aiger_ascii_extended.aag"), IO_FILE_AIGER, true)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "dispatch.aag")
	if err = WriteNetwork(network, path, IO_FILE_AIGER); err != nil {
		t.Fatal(err)
	}
	if _, err = ReadAiger(path, true); err != nil {
		t.Fatal(err)
	}
	cex := &Abc_Cex_t{IFrame: 1, NPis: 1, NRegs: 1, NBits: 3, PData: []uint32{0b110}}
	trace := filepath.Join(t.TempDir(), "trace.cex")
	if err = WriteAigerCounterexample(cex, network, nil, trace); err != nil {
		t.Fatal(err)
	}
	if data, readErr := os.ReadFile(trace); readErr != nil || string(data) != "1\n0 1 000 1\n1 1 110 0\n" {
		t.Fatalf("counterexample trace = %q, err=%v", data, readErr)
	}
}

func TestAigerConstantLiteralConvention(t *testing.T) {
	model := &ioa.Model{MaxVar: 0, Outputs: []uint32{0, 1}}
	data, err := ioa.Marshal(model, ioa.WriteOptions{ASCII: true})
	if err != nil {
		t.Fatal(err)
	}
	network, err := ReadAigerFromMemory(data, true)
	if err != nil {
		t.Fatal(err)
	}
	outputs := network.ObjectsOfKind(abc.ObjectPO)
	for i, want := range []bool{false, true} {
		got, evalErr := network.Evaluate(outputs[i].Signal(), nil)
		if evalErr != nil || got != want {
			t.Fatalf("constant output %d = %t, want %t (err=%v)", i, got, want, evalErr)
		}
	}
	roundtrip, err := aigerModelFromNetwork(network)
	if err != nil || !reflect.DeepEqual(roundtrip.Outputs, []uint32{0, 1}) {
		t.Fatalf("constant literals changed on write: %+v, err=%v", roundtrip, err)
	}
}

func TestBaseIoAigerCompactLiteralHelpers(t *testing.T) {
	literals := []uint32{0, 1, 127, 128, 4, 0x7fffffff}
	encoded, err := Io_WriteEncodeLiterals(literals)
	if err != nil {
		t.Fatal(err)
	}
	cursor := string(encoded)
	decoded := Io_WriteDecodeLiterals(&cursor, int32(len(literals)))
	if len(cursor) != 0 || !reflect.DeepEqual(decoded, literals) {
		t.Fatalf("compact literal helper roundtrip: got %v, rest %d bytes", decoded, len(cursor))
	}
	cursor = string(encoded)
	if first := Io_ReadAigerDecode(&cursor); first != literals[0] {
		t.Fatalf("first encoded literal = %d, want %d", first, literals[0])
	}
	if len(cursor) == 0 {
		t.Fatal("single-value decoder consumed the entire compact list")
	}
}

func TestAigerWriterKeepsUninitializedLatchState(t *testing.T) {
	network := abc.NewNetwork("state")
	network.Type, network.Function = uint8(abc.ABC_NTK_STRASH), uint8(abc.ABC_FUNC_AIG)
	pi := network.NewObject(abc.ObjectPI)
	bi, latch, bo := network.NewObject(abc.ObjectBI), network.NewObject(abc.ObjectLatch), network.NewObject(abc.ObjectBO)
	for _, connection := range []struct {
		object *abc.Object
		edge   abc.Edge
	}{{bi, pi.Signal()}, {latch, bi.Signal()}, {bo, latch.Signal()}} {
		if err := network.AddFanin(connection.object, connection.edge); err != nil {
			t.Fatal(err)
		}
	}
	model, err := aigerModelFromNetwork(network)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := model.Latches[0].Init, model.Latches[0].Literal; got != want {
		t.Fatalf("uninitialized latch reset = %d, want its current-state literal %d", got, want)
	}
}
