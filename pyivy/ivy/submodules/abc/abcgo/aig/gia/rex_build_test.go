package gia

import "testing"

func rexRun(p *Gia_Man_t, alphabet, input string) []bool {
	state := make([]bool, Gia_ManRegNum(p))
	result := make([]bool, 0, len(input))
	for _, ch := range []byte(input) {
		values := make([]bool, p.nObjs)
		for i := int32(0); i < Gia_ManPiNum(p); i++ {
			values[Gia_ObjId(p, Gia_ManPi(p, i))] = int(i) < len(alphabet) && alphabet[i] == ch
		}
		for i := int32(0); i < Gia_ManRegNum(p); i++ {
			values[Gia_ObjId(p, Gia_ManRo(p, i))] = state[i]
		}
		lit := func(v int32) bool { return values[v>>1] != (v&1 != 0) }
		for id := int32(1); id < p.nObjs; id++ {
			if Gia_ObjIsAnd(Gia_ManObj(p, id)) == 0 {
				continue
			}
			a, b, _, _ := p.Gate(id)
			values[id] = lit(a) && lit(b)
		}
		read := func(obj *Gia_Obj_t) bool { a, _, _, _ := p.Gate(Gia_ObjId(p, obj)); return lit(a) }
		result = append(result, read(Gia_ManPo(p, 0)))
		for i := range state {
			state[i] = read(Gia_ManRi(p, int32(i)))
		}
	}
	return result
}
func TestRexBuildSequence(t *testing.T) {
	for _, tc := range []struct {
		expression, input string
		want              []bool
	}{
		{"ab", "ab", []bool{false, true}},
		{"(a|b)*c", "abac", []bool{false, false, false, true}},
		{"a|b", "b", []bool{true}},
	} {
		p := Gia_ManRex2Gia(tc.expression, 1, 0)
		got := rexRun(p, "abc", tc.input)
		if len(got) != len(tc.want) {
			t.Fatal(got)
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Fatalf("%s %q: got %v want %v", tc.expression, tc.input, got, tc.want)
			}
		}
	}
}
