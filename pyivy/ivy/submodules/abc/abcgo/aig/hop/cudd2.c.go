// Recording AIGs for BDD operations, ported from ABC cudd2.c (Alan Mishchenko).
package hop

import (
	"fmt"
	"reflect"
	"sync"
)

type Aig_CuddMan_t struct {
	pAig   *Hop_Man_t
	pTable map[any]*Hop_Obj_t
}

var s_pCuddMan *Aig_CuddMan_t
var cuddRecorderMu sync.Mutex

func cuddIdentity(value any) (any, int32) {
	switch v := value.(type) {
	case BDDRef:
		if v.Identity != nil && !reflect.TypeOf(v.Identity).Comparable() {
			panic("BDD identity must be comparable")
		}
		return v.Identity, hopBool(v.Complement)
	case uintptr:
		return v &^ uintptr(1), int32(v & 1)
	}
	if value != nil && !reflect.TypeOf(value).Comparable() {
		panic("BDD identity must be comparable")
	}
	return value, 0
}
func cuddArg(arg any) *Hop_Obj_t {
	if s_pCuddMan == nil {
		panic("Cudd2 recorder is not initialized")
	}
	key, c := cuddIdentity(arg)
	o, ok := s_pCuddMan.pTable[key]
	if !ok {
		fmt.Println("Cudd2_GetArg(): An argument BDD is not in the hash table.")
		return nil
	}
	return Hop_NotCond(o, c)
}
func cuddSet(o *Hop_Obj_t, result any) {
	if s_pCuddMan == nil {
		panic("Cudd2 recorder is not initialized")
	}
	key, c := cuddIdentity(result)
	if _, exists := s_pCuddMan.pTable[key]; !exists {
		s_pCuddMan.pTable[key] = Hop_NotCond(o, c)
	}
}
func Cudd2_Init(numVars, numVarsZ, numSlots, cacheSize uint32, maxMemory uint64, manager any) {
	cuddRecorderMu.Lock()
	defer cuddRecorderMu.Unlock()
	if s_pCuddMan != nil {
		return
	}
	p := Hop_ManStart()
	for i := uint32(0); i < numVars; i++ {
		Hop_ObjCreatePi(p)
	}
	s_pCuddMan = &Aig_CuddMan_t{p, map[any]*Hop_Obj_t{}}
}
func Cudd2_Quit(manager any) {
	cuddRecorderMu.Lock()
	defer cuddRecorderMu.Unlock()
	if s_pCuddMan == nil {
		panic("Cudd2 recorder is not initialized")
	}
	Hop_ManDumpBlif(s_pCuddMan.pAig, "aig_temp.blif")
	Hop_ManStop(s_pCuddMan.pAig)
	s_pCuddMan = nil
}
func Cudd2_GetArg(arg any) *Hop_Obj_t {
	cuddRecorderMu.Lock()
	defer cuddRecorderMu.Unlock()
	return cuddArg(arg)
}
func Cudd2_SetArg(o *Hop_Obj_t, result any) {
	cuddRecorderMu.Lock()
	defer cuddRecorderMu.Unlock()
	cuddSet(o, result)
}
func Cudd2_bddOne(manager, result any) {
	cuddRecorderMu.Lock()
	defer cuddRecorderMu.Unlock()
	cuddSet(Hop_ManConst1(s_pCuddMan.pAig), result)
}
func Cudd2_bddIthVar(manager any, i int32, result any) {
	cuddRecorderMu.Lock()
	defer cuddRecorderMu.Unlock()
	cuddSet(Hop_IthVar(s_pCuddMan.pAig, i), result)
}
func cuddBinary(a, b, result any, op func(*Hop_Man_t, *Hop_Obj_t, *Hop_Obj_t) *Hop_Obj_t, neg bool) {
	cuddRecorderMu.Lock()
	defer cuddRecorderMu.Unlock()
	x, y := cuddArg(a), cuddArg(b)
	if x == nil || y == nil {
		return
	}
	o := op(s_pCuddMan.pAig, x, y)
	if neg {
		o = Hop_Not(o)
	}
	cuddSet(o, result)
}
func Cudd2_bddAnd(manager, a, b, result any)  { cuddBinary(a, b, result, Hop_And, false) }
func Cudd2_bddOr(manager, a, b, result any)   { cuddBinary(a, b, result, Hop_Or, false) }
func Cudd2_bddNand(manager, a, b, result any) { cuddBinary(a, b, result, Hop_And, true) }
func Cudd2_bddNor(manager, a, b, result any)  { cuddBinary(a, b, result, Hop_Or, true) }
func Cudd2_bddXor(manager, a, b, result any)  { cuddBinary(a, b, result, Hop_Exor, false) }
func Cudd2_bddXnor(manager, a, b, result any) { cuddBinary(a, b, result, Hop_Exor, true) }
func Cudd2_bddIte(manager, c, t, e, result any) {
	cuddRecorderMu.Lock()
	defer cuddRecorderMu.Unlock()
	a, b, d := cuddArg(c), cuddArg(t), cuddArg(e)
	if a != nil && b != nil && d != nil {
		cuddSet(Hop_Mux(s_pCuddMan.pAig, a, b, d), result)
	}
}
func Cudd2_bddCompose(manager, a, b any, v int32, result any) {
	cuddRecorderMu.Lock()
	defer cuddRecorderMu.Unlock()
	x, y := cuddArg(a), cuddArg(b)
	if x != nil && y != nil {
		cuddSet(Hop_Compose(s_pCuddMan.pAig, x, y, v), result)
	}
}
func Cudd2_bddLeq(manager, a, b any, result int32) {
	cuddRecorderMu.Lock()
	defer cuddRecorderMu.Unlock()
	x, y := cuddArg(a), cuddArg(b)
	if x != nil && y != nil {
		Hop_ObjCreatePo(s_pCuddMan.pAig, Hop_And(s_pCuddMan.pAig, x, Hop_Not(y)))
	}
}
func Cudd2_bddEqual(manager, a, b any, result int32) {
	cuddRecorderMu.Lock()
	defer cuddRecorderMu.Unlock()
	x, y := cuddArg(a), cuddArg(b)
	if x != nil && y != nil {
		Hop_ObjCreatePo(s_pCuddMan.pAig, Hop_Exor(s_pCuddMan.pAig, x, y))
	}
}
