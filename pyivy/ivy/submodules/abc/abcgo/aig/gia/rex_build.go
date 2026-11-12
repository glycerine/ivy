package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_ManRex2Gia builds a sequential AIG for the given regular expression.
func Gia_ManRex2Gia(expression string, order, verbose int32) *Gia_Man_t {
	s := Gia_ManRexPreprocess(expression)
	var alpha, stateMap *Vec_Int_t
	nStates := Gia_ManRexNumInputs(s, &alpha, &stateMap)
	if vec.Vec_IntSize(alpha) == 0 {
		panic("gia: regex has no input symbols")
	}
	if order != 0 {
		vec.Vec_IntSort(alpha, 0)
	}
	fmt.Print("Input variable order: ")
	for _, c := range vec.Vec_IntArray(alpha) {
		fmt.Printf("%c", c)
	}
	fmt.Println()
	p := Gia_ManStart(1000)
	p.pName = expression
	for i := int32(0); i < int32(vec.Vec_IntSize(alpha))+nStates; i++ {
		Gia_ManAppendCi(p)
	}
	transitions := rexTransitions(s)
	if verbose != 0 {
		Gia_ManPrintAutom(s, transitions)
	}
	visited := vec.Vec_IntStartFull(len(s) + 1)
	reached := vec.Vec_IntAlloc(len(s) + 1)
	stateLits := make([]int32, nStates+1)
	travID := int32(1)
	Gia_ManHashStart(p)
	for i := 1; i < len(s); i++ {
		if Gia_SymbSpecial(s[i]) != 0 {
			continue
		}
		Gia_ManCollectReached(s, transitions, int32(i), reached, visited, travID)
		travID++
		if verbose != 0 {
			Gia_ManPrintReached(s, int32(i), reached)
		}
		state := vec.Vec_IntEntry(stateMap, i)
		symbol := vec.Vec_IntFind(alpha, int32(s[i]))
		a := Gia_Obj2Lit(p, Gia_ManCi(p, int32(vec.Vec_IntSize(alpha))+state))
		b := Gia_Obj2Lit(p, Gia_ManCi(p, int32(symbol)))
		lit := Gia_ManHashAnd(p, a, b)
		for _, dest := range vec.Vec_IntArray(reached) {
			j := vec.Vec_IntEntry(stateMap, int(dest))
			lit = Gia_ManHashOr(p, lit, stateLits[j])
			stateLits[j] = lit
		}
	}
	inputs := vec.Vec_IntAlloc(vec.Vec_IntSize(alpha))
	for i := 0; i < vec.Vec_IntSize(alpha); i++ {
		vec.Vec_IntPush(inputs, Gia_Obj2Lit(p, Gia_ManCi(p, int32(i))))
	}
	oneHot := Gia_ManBuild1Hot(p, inputs)
	for i := range stateLits {
		stateLits[i] = Gia_ManHashAnd(p, oneHot, stateLits[i])
	}
	Gia_ManHashStop(p)
	Gia_ManCollectReached(s, transitions, 0, reached, visited, travID)
	initial := make([]byte, nStates)
	for i := range initial {
		initial[i] = '0'
	}
	for _, id := range vec.Vec_IntArray(reached) {
		if rexSymbol(s, id) != 0 {
			initial[vec.Vec_IntEntry(stateMap, int(id))] = '1'
		}
	}
	if verbose != 0 {
		Gia_ManPrintReached(s, 0, reached)
		fmt.Printf("Init state = %s\n", initial)
	}
	Gia_ManAppendCo(p, stateLits[nStates])
	for _, lit := range stateLits[:nStates] {
		Gia_ManAppendCo(p, lit)
	}
	Gia_ManSetRegNum(p, nStates)
	p = Gia_ManCleanup(p)
	return Gia_ManDupZeroUndc(p, string(initial), 0, 0, 0)
}
