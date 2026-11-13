package fra

// Helpers translated from fraClau.c. Literals use ABC's var*2+sign encoding.
import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/cnf"
)

func clauLatchVars(graph *aig.Graph, data *cnf.Cnf_Dat_t, current bool) *vec.Vec_Int_t {
	n := int(graph.RegisterCount())
	inputs, outputs := graph.Inputs(), graph.Outputs()
	result := vec.Vec_IntAlloc(n)
	for i := 0; i < n; i++ {
		edge := outputs[len(outputs)-n+i]
		if current {
			edge = inputs[len(inputs)-n+i]
		}
		vec.Vec_IntPush(result, data.VarNums[edge.NodeID()])
	}
	return result
}
func clauOutputVars(graph *aig.Graph, data *cnf.Cnf_Dat_t) *vec.Vec_Int_t {
	result := vec.Vec_IntAlloc(graph.OutputCount())
	for _, edge := range graph.Outputs() {
		vec.Vec_IntPush(result, data.VarNums[edge.NodeID()])
	}
	return result
}
func clauInputVars(graph *aig.Graph, data *cnf.Cnf_Dat_t, start int) *vec.Vec_Int_t {
	inputs := graph.Inputs()
	if start < 0 || start > len(inputs) {
		panic("invalid starting input")
	}
	result := vec.Vec_IntAlloc(len(inputs) - start)
	for _, edge := range inputs[start:] {
		vec.Vec_IntPush(result, data.VarNums[edge.NodeID()])
	}
	return result
}
func clauMapping(from, to *vec.Vec_Int_t, size int) []int32 {
	if vec.Vec_IntSize(from) != vec.Vec_IntSize(to) || size < 0 {
		panic("invalid clause variable mapping")
	}
	mapping := make([]int32, size)
	for i := range mapping {
		mapping[i] = -1
	}
	for i, v := range vec.Vec_IntArray(from) {
		if v < 0 || int(v) >= size {
			panic("clause variable out of range")
		}
		mapping[v] = vec.Vec_IntEntry(to, i)
	}
	return mapping
}
func clauRemap(mapping []int32, clause *vec.Vec_Int_t, invert bool) *vec.Vec_Int_t {
	out := vec.Vec_IntAlloc(vec.Vec_IntSize(clause))
	for _, lit := range vec.Vec_IntArray(clause) {
		v := lit >> 1
		if lit < 0 || int(v) >= len(mapping) || mapping[v] < 0 {
			panic("unmapped clause variable")
		}
		sign := lit & 1
		if invert {
			sign ^= 1
		}
		vec.Vec_IntPush(out, mapping[v]*2+sign)
	}
	return out
}
func clauReduce(main, other *vec.Vec_Int_t) {
	a, b := vec.Vec_IntArray(main), vec.Vec_IntArray(other)
	if len(a) > len(b) {
		panic("counterexample shorter than clause")
	}
	i, j, k := 0, 0, 0
	for i < len(a) && j < len(b) {
		if a[i]>>1 < b[j]>>1 {
			panic("counterexample omits clause variable")
		}
		if a[i]>>1 > b[j]>>1 {
			j++
			continue
		}
		if a[i] == b[j] {
			a[k] = a[i]
			k++
		}
		i++
		j++
	}
	if i != len(a) {
		panic("counterexample omits clause variable")
	}
	vec.Vec_IntShrink(main, k)
}
