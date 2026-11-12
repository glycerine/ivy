package abc

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

func fanEval(o *Object, inputs map[*Object]bool) bool {
	if o.Kind == ObjectConst1 {
		return true
	}
	if o.IsCI() {
		return inputs[o]
	}
	args := make([]bool, len(o.fanins))
	for i, edge := range o.fanins {
		args[i] = fanEval(edge.Object, inputs) != edge.Complement
	}
	text := o.Data.(string)
	polarity := byte('1')
	matched := false
	for _, line := range strings.Split(text, "\n") {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		pattern, output := "", fields[0]
		if len(fields) > 1 {
			pattern, output = fields[0], fields[1]
		}
		polarity = output[0]
		if polarity == 'x' || polarity == 'n' {
			value := false
			for _, arg := range args {
				value = value != arg
			}
			return value != (polarity == 'n')
		}
		match := true
		for i, c := range pattern {
			if c != '-' && (c == '1') != args[i] {
				match = false
				break
			}
		}
		matched = matched || match
	}
	return matched == (polarity == '1')
}
func fanTruth(o *Object, pis []*Object) []bool {
	truth := make([]bool, 1<<len(pis))
	for assignment := range truth {
		values := make(map[*Object]bool)
		for i, pi := range pis {
			values[pi] = assignment&(1<<i) != 0
		}
		truth[assignment] = fanEval(o, values)
	}
	return truth
}
func fanFixture(text string) (*Network, *Object, []*Object) {
	n := NewNetwork("fanin")
	n.Type = uint8(ABC_NTK_LOGIC)
	n.Function = uint8(ABC_FUNC_SOP)
	pis := []*Object{n.NewObject(ObjectPI), n.NewObject(ObjectPI), n.NewObject(ObjectPI), n.NewObject(ObjectPI)}
	o := n.NewObject(ObjectNode)
	o.Data = text
	for i := len(pis) - 1; i >= 0; i-- {
		Abc_ObjAddFanin(o, Edge{Object: pis[i], Complement: i == 1})
	}
	return n, o, pis
}
func TestFaninOrdersPreserveFunctionAndEdgePolarity(t *testing.T) {
	for name, transform := range map[string]func(*Network){"id": Abc_NtkOrderFaninsById, "columns": Abc_NtkOrderFaninsBySortingColumns, "literals": Abc_NtkOrderFaninsByLitCount, "cube-number": Abc_NtkOrderFaninsByLitCountAndCubeCount, "readable": Abc_NtkSortSops} {
		t.Run(name, func(t *testing.T) {
			n, o, pis := fanFixture("1-01 0\n-10- 0\n--11 0\n")
			before := fanTruth(o, pis)
			transform(n)
			if !reflect.DeepEqual(before, fanTruth(o, pis)) {
				t.Fatal("function changed", o.Data)
			}
			if err := n.Check(); err != nil {
				t.Fatal(err)
			}
			if name == "id" {
				for i, edge := range o.fanins {
					if edge.Object != pis[i] || edge.Complement != (i == 1) {
						t.Fatal(o.fanins)
					}
				}
			}
		})
	}
}
func TestSOPTransposeAndSelectionTies(t *testing.T) {
	var columns []string
	var store []byte
	Abc_NtkSopTranspose("1-0 1\n0-1 1\n", 3, &columns, &store)
	if !reflect.DeepEqual(columns, []string{"10", "--", "01"}) {
		t.Fatal(columns)
	}
	var order []int32
	Vec_StrSelectSortCost(columns, 3, &order)
	if !reflect.DeepEqual(order, []int32{1, 2, 0}) {
		t.Fatal(order)
	}
	if !reflect.DeepEqual(store, []byte{'1', '0', 0, '-', '-', 0, '0', '1', 0}) {
		t.Fatal(store)
	}
}
func TestSplitLargeOnAndOffSets(t *testing.T) {
	for _, polarity := range []string{"0", "1"} {
		n, o, pis := fanFixture("1-01 " + polarity + "\n-10- " + polarity + "\n--11 " + polarity + "\n")
		before := fanTruth(o, pis)
		Abc_NodeSplitLarge(o)
		if len(o.fanins) != 2 || len(n.ObjectsOfKind(ObjectNode)) != 3 {
			t.Fatal("not split")
		}
		if !reflect.DeepEqual(before, fanTruth(o, pis)) {
			t.Fatal("split changed function", polarity, o.Data)
		}
		if err := n.Check(); err != nil {
			t.Fatal(err)
		}
	}
}
func TestSOPLegitimizationRandomTruthPreservation(t *testing.T) {
	rng := rand.New(rand.NewSource(901))
	for trial := 0; trial < 100; trial++ {
		var text strings.Builder
		polarity := byte('0' + trial%2)
		for row := 0; row < 8; row++ {
			for variable := 0; variable < 4; variable++ {
				text.WriteByte("-01"[rng.Intn(3)])
			}
			fmt.Fprintf(&text, " %c\n", polarity)
		}
		n, o, pis := fanFixture(text.String())
		before := fanTruth(o, pis)
		Abc_NtkMakeLegit(n)
		if !reflect.DeepEqual(before, fanTruth(o, pis)) {
			t.Fatal("normalization changed function", trial, text.String(), o.Data)
		}
		if Abc_NodeMakeLegit(o) != 0 {
			t.Fatal("not idempotent", o.Data)
		}
		Abc_NodeCheckDist1Free(o)
		rows := fanNodeRows(o)
		for i := range rows {
			for j := i + 1; j < len(rows); j++ {
				if Abc_CubeContain(rows[i].pattern, rows[j].pattern, 4) != 0 {
					t.Fatal("containment remains", o.Data)
				}
			}
		}
	}
}
