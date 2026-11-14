package abci

import (
	"github.com/glycerine/abcgo/base/abc"
	"testing"
)

func TestSourceDebugFaninOrderAndFilters(t *testing.T) {
	for _, outputs := range []int{1, 2} {
		t.Run(map[int]string{1: "single", 2: "multiple"}[outputs], func(t *testing.T) {
			n := abc.NewNetwork("fanin")
			input := abc.Abc_NtkCreatePi(n)
			constant := abc.Abc_NtkCreateNode(n)
			node := abc.Abc_NtkCreateNode(n)
			abc.Abc_ObjAddFanin(node, input)
			abc.Abc_ObjAddFanin(node, constant)
			child := abc.Abc_NtkCreateNode(n)
			abc.Abc_ObjAddFanin(child, node)
			var pos []*abc.Object
			for i := 0; i < outputs; i++ {
				po := abc.Abc_NtkCreatePo(n)
				abc.Abc_ObjAddFanin(po, child)
				pos = append(pos, po)
			}
			expected := [][2]*abc.Object{{node, input}, {child, node}}
			if outputs == 2 {
				expected = append(expected, [2]*abc.Object{pos[0], child}, [2]*abc.Object{pos[1], child})
			}
			if got := Abc_NtkCountFaninsTotal(n); got != int32(len(expected)) {
				t.Fatalf("count %d want %d", got, len(expected))
			}
			for i, pair := range expected {
				var obj, fanin *abc.Object
				if Abc_NtkFindGivenFanin(n, int32(i), &obj, &fanin) != 1 || obj != pair[0] || fanin != pair[1] {
					t.Fatalf("step %d got %v,%v", i, obj, fanin)
				}
			}
			obj, fanin := node, input
			if Abc_NtkFindGivenFanin(n, int32(len(expected)), &obj, &fanin) != 0 || obj != node || fanin != input {
				t.Fatal("missing step modified outputs")
			}
		})
	}
}
