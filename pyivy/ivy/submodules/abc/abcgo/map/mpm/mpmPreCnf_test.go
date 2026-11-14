package mpm

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestMpmComputeCnfSizeOneEncoding(t *testing.T) {
	cases := []struct {
		name  string
		truth word
		cubes int32
		bytes []byte
	}{
		{name: "false", truth: 0, cubes: 1, bytes: []byte{1, 0xff}},
		{name: "true", truth: ^word(0), cubes: 1, bytes: []byte{0, 0xff}},
		{name: "variable zero", truth: 0xaaaaaaaaaaaaaaaa, cubes: 2, bytes: []byte{1, 4, 0xff, 0, 5, 0xff}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cover := vec.Vec_IntAlloc(16)
			cnf := vec.Vec_StrAlloc(16)
			gotCubes := Mpm_ComputeCnfSizeOne(tc.truth, 2, cover, cnf)
			gotBytes := vec.Vec_StrArray(cnf)
			if gotCubes != tc.cubes || !reflect.DeepEqual(gotBytes, tc.bytes) {
				t.Fatalf("Mpm_ComputeCnfSizeOne(%#016x) = %d, %v; want %d, %v", tc.truth, gotCubes, gotBytes, tc.cubes, tc.bytes)
			}
			vec.Vec_IntFree(cover)
			vec.Vec_StrFree(cnf)
		})
	}
}
