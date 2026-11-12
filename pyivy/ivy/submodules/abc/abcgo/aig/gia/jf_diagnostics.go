package gia

import "fmt"

func Jf_ObjPrintStore(p *Jf_Man_t, store []*Jf_Cut_t, count int32) {
	for i := int32(0); i < count; i++ {
		cut := store[i]
		fmt.Printf("Flow =%9.5f  Time = %5d   Func = %5d     ", cut.Flow, cut.Time, cut.iFunc)
		Jf_CutPrint(cut.pCut[:])
	}
	fmt.Println()
}
func Jf_ObjCheckPtrs(store []*Jf_Cut_t, count int32) {
	for i := int32(1); i < count; i++ {
		for k := int32(0); k < i; k++ {
			if store[k] == store[i] {
				panic("JF duplicate cut pointers")
			}
		}
	}
}
func Jf_ObjCheckStore(p *Jf_Man_t, store []*Jf_Cut_t, count, id int32) {
	for i := int32(1); i < count; i++ {
		if p.pCutCmp(store[i-1], store[i]) > 0 {
			panic("JF cut store out of order")
		}
	}
	for i := int32(1); i < count; i++ {
		for k := int32(0); k < i; k++ {
			if Jf_CutIsContained1(store[k].pCut[:], store[i].pCut[:]) || Jf_CutIsContained1(store[i].pCut[:], store[k].pCut[:]) {
				panic("JF redundant cut")
			}
		}
	}
}
