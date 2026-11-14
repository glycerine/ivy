package abci

import (
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
	"strings"
)

// abcNtkESOP translates the PLA cubes directly, retaining negative literals
// and output indices from exor.c's network-to-ESOP path.
func abcNtkESOP(n *Abc_Ntk_t) *Vec_Wec_t {
	rows := vec.Vec_WecAlloc(16)
	inputs := make(map[*abc.Object]int32)
	for i := int32(0); i < abc.Abc_NtkCiNum(n); i++ {
		inputs[abc.Abc_NtkCi(n, i)] = i
	}
	for i := int32(0); i < abc.Abc_NtkCoNum(n); i++ {
		co := abc.Abc_NtkCo(n, i)
		driver := abc.Abc_ObjFanin0(co)
		if abc.Abc_NtkIsNetlist(n) != 0 {
			driver = abc.Abc_ObjFanin0(driver)
		}
		if abc.Abc_ObjIsCi(driver) != 0 {
			row := vec.Vec_WecPushLevel(rows)
			vec.Vec_IntPush(row, inputs[driver]*2)
			vec.Vec_IntPush(row, -i-1)
			continue
		}
		sop, ok := driver.Data.(string)
		if !ok {
			vec.Vec_WecFree(rows)
			return nil
		}
		for _, cube := range strings.Split(strings.TrimSuffix(sop, "\n"), "\n") {
			if len(cube) < len(driver.Fanins())+2 {
				vec.Vec_WecFree(rows)
				return nil
			}
			if cube[len(driver.Fanins())+1] == '0' {
				continue
			}
			row := vec.Vec_WecPushLevel(rows)
			for pin, fanin := range driver.Fanins() {
				object := fanin.Object
				if abc.Abc_NtkIsNetlist(n) != 0 {
					object = abc.Abc_ObjFanin0(object)
				}
				id, exists := inputs[object]
				if !exists {
					vec.Vec_WecFree(rows)
					return nil
				}
				switch cube[pin] {
				case '0':
					vec.Vec_IntPush(row, id*2+1)
				case '1':
					vec.Vec_IntPush(row, id*2)
				}
			}
			vec.Vec_IntPush(row, -i-1)
		}
	}
	return rows
}

func abcESOPRows(rows *Vec_Wec_t) [][]int32 {
	out := make([][]int32, vec.Vec_WecSize(rows))
	for i := range out {
		out[i] = vec.Vec_IntArray(vec.Vec_WecEntry(rows, i))
	}
	return out
}
