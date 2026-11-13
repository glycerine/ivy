package cnf

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestPackedSopAndIsopConversion(t *testing.T) {
	cover := vec.Vec_IntAlloc(0)
	Cnf_SopConvertToVector(string([]byte{0, 80}), 2, cover)
	if got := vec.Vec_IntArray(cover); !reflect.DeepEqual(got, []int32{0x55, 0}) {
		t.Fatalf("cover %v", got)
	}
	if got := Cnf_SopCountLiterals(string([]byte{0, 80}), 2); got != 4 {
		t.Fatalf("SOP literals %d", got)
	}
	if got := Cnf_IsopCountLiterals(cover, 4); got != 4 {
		t.Fatalf("ISOP literals %d", got)
	}
	vars, literals := []int32{5, 7, 9, 11}, make([]int32, 4)
	if got := Cnf_IsopWriteCube(0x21, 4, vars, literals); got != 2 {
		t.Fatalf("cube length %d", got)
	}
	if !reflect.DeepEqual(literals[:2], []int32{10, 19}) || !reflect.DeepEqual(vars, []int32{5, 7, 9, 11}) {
		t.Fatalf("variables %v literals %v", vars, literals)
	}
}
