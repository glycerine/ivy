package acb

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestSourceSimpleLexFileEntry(t *testing.T) {
	names := util.Abc_NamStart(100, 16)
	for _, keyword := range []string{"module", "endmodule", "input", "output", "wire", "buf", "not", "and", "nand", "or", "nor", "xor", "xnor", "_HMUX", "_DC"} {
		util.Abc_NamStrFindOrAdd(names, keyword, nil)
	}
	filename := filepath.Join(t.TempDir(), "design.v")
	if err := os.WriteFile(filename, []byte("module top;\ninput a,b;\noutput z;\nand gate (z,a,b); // comment\nendmodule\n"), 0600); err != nil {
		t.Fatal(err)
	}
	result := Acb_VerilogSimpleLex(filename, names)
	if result == nil {
		t.Fatal("failed to read input")
	}
	want := []int32{1, 16, 3, 17, 18, 4, 19, 8, 19, 17, 18, 2}
	if got := vec.Vec_IntArray(result); !reflect.DeepEqual(got, want) {
		t.Fatalf("tokens = %v, want %v", got, want)
	}
	if got := Acb_VerilogSimpleLex(filename+"-missing", names); got != nil {
		t.Fatal("missing file accepted")
	}
}
