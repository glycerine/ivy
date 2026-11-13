package mpm

import "testing"

func TestMigSourceObjectClassifications(t *testing.T) {
	cases := []struct {
		name                       string
		fans                       [3]uint32
		ci, co, buf, and, xor, mux int32
	}{
		{"ci", [3]uint32{MIG_NONE, MIG_NONE, 0}, 1, 0, 0, 0, 0, 0},
		{"co", [3]uint32{2, MIG_NONE, 0}, 0, 1, 0, 0, 0, 0},
		{"buf", [3]uint32{2, MIG_NONE, MIG_NONE}, 0, 0, 1, 0, 0, 0},
		{"and", [3]uint32{2, 3, MIG_NONE}, 0, 0, 0, 1, 0, 0},
		{"xor", [3]uint32{3, 2, MIG_NONE}, 0, 0, 0, 0, 1, 0},
		{"mux", [3]uint32{2, 3, 4}, 0, 0, 0, 0, 0, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			o := Mig_Obj_t{}
			for i := range tc.fans {
				o.pFans[i].Id = tc.fans[i]
			}
			got := [6]int32{Mig_ObjIsCi(&o), Mig_ObjIsCo(&o), Mig_ObjIsBuf(&o), Mig_ObjIsAnd(&o), Mig_ObjIsXor(&o), Mig_ObjIsMux(&o)}
			want := [6]int32{tc.ci, tc.co, tc.buf, tc.and, tc.xor, tc.mux}
			if got != want {
				t.Fatalf("got %v want %v", got, want)
			}
		})
	}
}
func TestMigFanFields(t *testing.T) {
	o := Mig_Obj_t{}
	if Mig_FanSetId(&o, 0, 42) != 42 || Mig_FanSetCompl(&o, 0, 1) != 1 || Mig_FanId(&o, 0) != 42 || Mig_FanCompl(&o, 0) != 1 {
		t.Fatal("fanin fields did not round trip")
	}
}
