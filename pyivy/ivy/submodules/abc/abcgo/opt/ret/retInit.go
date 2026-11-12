package ret

import (
	"fmt"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/sat/bsat"
	"strings"
)

func retSop(o *abc.Object) string {
	s, ok := o.Data.(string)
	if !ok {
		panic("ret: expected SOP node")
	}
	return s
}
func Abc_ObjSopSimulate(o *abc.Object) int32 {
	s := retSop(o)
	if abc.Abc_SopIsExorType(s) != 0 {
		panic("ret: XOR SOP is unsupported")
	}
	n := int(abc.Abc_SopGetVarNum(s))
	value := false
	for _, cube := range strings.Split(strings.TrimSuffix(s, "\n"), "\n") {
		if len(cube) < n+2 {
			panic("ret: malformed SOP cube")
		}
		term := true
		for i := 0; i < n; i++ {
			v := retFanValue(o.Fanins()[i])
			switch cube[i] {
			case '0':
				term = term && v == 0
			case '1':
				term = term && v != 0
			}
		}
		value = value || term
	}
	if abc.Abc_SopGetPhase(s) == 0 {
		value = !value
	}
	return retBool(value)
}
func retFanValue(e abc.Edge) int32 {
	v := e.Object.Temp
	if e.Object.Kind == abc.ObjectConst1 {
		v = 1
	}
	if e.Complement {
		v ^= 1
	}
	return int32(v) & 1
}

// Direct SOP Tseitin encoding preserves the target-cone SAT obligation without
// converting mutable networks through the legacy ABC command infrastructure.
func Abc_NtkRetimeInitialValues(cone *abc.Network, values []int32, verbose int32) []int32 {
	if cone == nil {
		return append([]int32(nil), values...)
	}
	pis, pos := cone.ObjectsOfKind(abc.ObjectPI), cone.ObjectsOfKind(abc.ObjectPO)
	if len(pos) != len(values) {
		panic("ret: target output count differs")
	}
	solver := bsat.NewSolver()
	solver.SetLimits(500000, 50000000)
	vars := make(map[*abc.Object]bsat.Literal)
	for _, o := range cone.Objects() {
		vars[o] = bsat.Positive(solver.NewVariable())
	}
	solver.AddClause(vars[cone.One().Object])
	lit := func(e abc.Edge) bsat.Literal {
		v := vars[e.Object]
		if e.Complement {
			v ^= 1
		}
		return v
	}
	for _, o := range retNodes(cone) {
		s := retSop(o)
		if abc.Abc_SopIsExorType(s) != 0 {
			panic("ret: XOR SOP is unsupported")
		}
		n := int(abc.Abc_SopGetVarNum(s))
		if n != len(o.Fanins()) {
			panic("ret: SOP arity differs")
		}
		var terms []bsat.Literal
		for _, cube := range strings.Split(strings.TrimSuffix(s, "\n"), "\n") {
			if len(cube) < n+2 {
				panic("ret: malformed SOP")
			}
			term := bsat.Positive(solver.NewVariable())
			var args []bsat.Literal
			for i, e := range o.Fanins() {
				if cube[i] == '-' {
					continue
				}
				a := lit(e)
				if cube[i] == '0' {
					a ^= 1
				}
				args = append(args, a)
				solver.AddClause(term^1, a)
			}
			clause := []bsat.Literal{term}
			for _, a := range args {
				clause = append(clause, a^1)
			}
			solver.AddClause(clause...)
			terms = append(terms, term)
		}
		z := vars[o]
		if abc.Abc_SopGetPhase(s) == 0 {
			z ^= 1
		}
		for _, term := range terms {
			solver.AddClause(term^1, z)
		}
		clause := []bsat.Literal{z ^ 1}
		clause = append(clause, terms...)
		solver.AddClause(clause...)
	}
	for i, o := range pos {
		target := lit(o.Fanins()[0])
		if values[i] == 0 {
			target ^= 1
		}
		solver.AddClause(target)
	}
	status := solver.Solve()
	if status != bsat.Satisfiable {
		if verbose != 0 {
			fmt.Println("Abc_NtkRetimeInitialValues(): Target has no computed state; DC latch values are used.")
		}
		return nil
	}
	model := solver.Model()
	result := make([]int32, len(pis))
	for i, o := range pis {
		result[i] = retBool(model[int(vars[o]>>1)])
	}
	if Abc_NtkRetimeVerifyModel(cone, values, result) == 0 {
		panic("ret: initial-state SAT model failed verification")
	}
	return result
}
func Abc_NtkRetimeVerifyModel(cone *abc.Network, values, model []int32) int32 {
	pis, pos := cone.ObjectsOfKind(abc.ObjectPI), cone.ObjectsOfKind(abc.ObjectPO)
	if len(pis) != len(model) || len(pos) != len(values) {
		return 0
	}
	inputs := make(map[*abc.Object]bool)
	for i, o := range pis {
		inputs[o] = model[i] != 0
	}
	for i, o := range pos {
		got, err := cone.Evaluate(o.Signal(), inputs)
		if err != nil || got != (values[i] != 0) {
			return 0
		}
	}
	return 1
}
func Abc_NtkRetimeTranferToCopy(n *abc.Network) {
	for _, o := range retLatches(n) {
		o.Temp = int(retLatchValue(o))
	}
}
func Abc_NtkRetimeTranferFromCopy(n *abc.Network) {
	for _, o := range retLatches(n) {
		if o.Temp != 0 {
			abc.Abc_LatchSetInit1(o)
		} else {
			abc.Abc_LatchSetInit0(o)
		}
	}
}
func Abc_NtkRetimeCollectLatchValues(n *abc.Network) []int32 {
	values := make([]int32, 0)
	for _, o := range retLatches(n) {
		values = append(values, retLatchValue(o))
	}
	return values
}
func Abc_NtkRetimeInsertLatchValues(n *abc.Network, values []int32) {
	latches := retLatches(n)
	if values != nil && len(values) != len(latches) {
		panic("ret: latch value count differs")
	}
	for i, o := range latches {
		o.Temp = i
		if values == nil {
			abc.Abc_LatchSetInitDc(o)
		} else if values[i] != 0 {
			abc.Abc_LatchSetInit1(o)
		} else {
			abc.Abc_LatchSetInit0(o)
		}
	}
}
func Abc_NtkRetimeBackwardInitialStart(n *abc.Network) *abc.Network {
	cone := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	for _, o := range retLatches(n) {
		o.Copy = abc.Abc_NtkCreatePo(cone)
	}
	return cone
}
func Abc_NtkRetimeBackwardInitialFinish(n, cone *abc.Network, oldValues []int32, verbose int32) {
	for _, o := range retLatches(n) {
		if o.Copy == nil {
			panic("ret: missing initial-state copy")
		}
		retAdd(o.Copy, abc.Abc_NtkCreatePi(cone).Signal())
	}
	Abc_NtkRetimeInsertLatchValues(n, Abc_NtkRetimeInitialValues(cone, oldValues, verbose))
}
func Abc_NtkCycleInitStateSop(n *abc.Network, nFrames, verbose int32) {
	random := newRetRandom(0x12341234)
	pis := n.ObjectsOfKind(abc.ObjectPI)
	for _, o := range pis {
		o.Temp = int(random.next() & 1)
	}
	for _, o := range retLatches(n) {
		o.Temp = int(retLatchValue(o))
		for _, bo := range o.Fanouts() {
			bo.Temp = o.Temp
		}
	}
	nodes := abc.Abc_NtkDfs(n, 0)
	for frame := int32(0); frame < nFrames; frame++ {
		for _, o := range nodes {
			o.Temp = int(Abc_ObjSopSimulate(o))
		}
		for _, o := range append(n.ObjectsOfKind(abc.ObjectPO), n.ObjectsOfKind(abc.ObjectBI)...) {
			o.Temp = int(retFanValue(o.Fanins()[0]))
		}
		for _, o := range pis {
			o.Temp = int(random.next() & 1)
		}
		for _, o := range retLatches(n) {
			for _, bo := range o.Fanouts() {
				bo.Temp = retFan0(o).Temp
			}
		}
	}
	for _, o := range retLatches(n) {
		if o.Fanouts()[0].Temp != 0 {
			abc.Abc_LatchSetInit1(o)
		} else {
			abc.Abc_LatchSetInit0(o)
		}
	}
}

// ABC's source uses libc rand after a fixed seed. Reproduce the glibc additive
// generator locally so cycling does not mutate process-global random state.
type retRandom struct {
	state       [31]uint32
	front, rear int
}

func newRetRandom(seed uint32) *retRandom {
	r := &retRandom{front: 3}
	if seed == 0 {
		seed = 1
	}
	r.state[0] = seed
	value := int64(int32(seed))
	for i := 1; i < 31; i++ {
		value = (16807 * value) % 2147483647
		if value < 0 {
			value += 2147483647
		}
		r.state[i] = uint32(value)
	}
	for i := 0; i < 310; i++ {
		r.next()
	}
	return r
}
func (r *retRandom) next() uint32 {
	r.state[r.front] += r.state[r.rear]
	value := r.state[r.front] >> 1
	r.front = (r.front + 1) % 31
	r.rear = (r.rear + 1) % 31
	return value
}
