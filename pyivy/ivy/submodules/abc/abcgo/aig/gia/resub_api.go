package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"sync"
)

func resubFormatLit(gates []int32, nVars, lit int32) string {
	id, phase := lit>>1, lit&1
	if id < nVars {
		name := fmt.Sprintf("i%d", id-2)
		if nVars < 26 {
			name = string(rune('a' + id - 2))
		}
		if phase != 0 {
			return "~" + name
		}
		return name
	}
	index := int(2 * (id - nVars))
	if index < 0 || index+1 >= len(gates) {
		panic("gia: invalid resubstitution print node")
	}
	a, b := gates[index], gates[index+1]
	op := "&"
	prefix := ""
	if a > b {
		op = "^"
		if phase != 0 {
			prefix = "~"
			phase = 0
		}
	} else if phase != 0 {
		op = "|"
	}
	return prefix + "(" + resubFormatLit(gates, nVars, a^phase) + " " + op + " " + resubFormatLit(gates, nVars, b^phase) + ")"
}
func Gia_ManResubPrintNode(gates *Vec_Int_t, nVars, node, phase int32) {
	fmt.Print(resubFormatLit(vec.Vec_IntArray(gates), nVars, 2*(nVars+node)+(phase&1)))
}
func Gia_ManResubPrintLit(gates *Vec_Int_t, nVars, lit int32) {
	fmt.Print(resubFormatLit(vec.Vec_IntArray(gates), nVars, lit))
}
func Gia_ManResubPrint(gates *Vec_Int_t, nVars int32) int32 {
	data := vec.Vec_IntArray(gates)
	if len(data) == 0 {
		n, _ := fmt.Print("none")
		return int32(n)
	}
	if len(data)%2 != 1 {
		panic("gia: invalid resubstitution gate list")
	}
	top := data[len(data)-1]
	if top == 0 {
		n, _ := fmt.Print("const0")
		return int32(n)
	}
	if top == 1 {
		n, _ := fmt.Print("const1")
		return int32(n)
	}
	Gia_ManResubPrintLit(gates, nVars, top)
	return 0
}
func Gia_ManResubOne(divs *Vec_Ptr_t, nWords, limit, divsMax, choice, useXor, debug, verbose int32, function []word) *Vec_Int_t {
	p := Gia_ResbAlloc(nWords)
	defer Gia_ResbFree(p)
	Gia_ManResubPerform(p, divs, nWords, limit, divsMax, choice, useXor, debug, verbose)
	if verbose != 0 {
		Gia_ManResubPrint(p.vGates, int32(vec.Vec_PtrSize(divs)))
		fmt.Println()
	}
	if Gia_ManResubVerify(p, function) == 0 {
		Gia_ManResubPrint(p.vGates, int32(vec.Vec_PtrSize(divs)))
		fmt.Println("Verification FAILED.")
	}
	return vec.Vec_IntDup(p.vGates)
}

var resubShared struct {
	sync.Mutex
	manager *Gia_ResbMan_t
}

func Abc_ResubPrepareManager(nWords int32) {
	resubShared.Lock()
	defer resubShared.Unlock()
	if resubShared.manager != nil {
		Gia_ResbFree(resubShared.manager)
		resubShared.manager = nil
	}
	if nWords > 0 {
		resubShared.manager = Gia_ResbAlloc(nWords)
	}
}
func Abc_ResubComputeFunction(divisors [][]word, nWords, limit, divsMax, choice, useXor, debug, verbose int32) []int32 {
	resubShared.Lock()
	defer resubShared.Unlock()
	p := resubShared.manager
	if p == nil {
		panic("gia: resubstitution manager is not prepared")
	}
	divs := vec.Vec_PtrAlloc(len(divisors))
	for _, d := range divisors {
		vec.Vec_PtrPush(divs, d)
	}
	detail := int32(0)
	if verbose == 2 {
		detail = 1
	}
	Gia_ManResubPerform(p, divs, nWords, limit, divsMax, choice, useXor, debug, detail)
	if verbose != 0 && vec.Vec_IntSize(p.vGates) > 1 {
		fmt.Printf("      Gain = %2d  Gates = %2d  __________  F = ", limit+1-int32(vec.Vec_IntSize(p.vGates)/2), vec.Vec_IntSize(p.vGates)/2)
		Gia_ManResubPrint(p.vGates, int32(len(divisors)))
		fmt.Println()
	}
	if debug != 0 && Gia_ManResubVerify(p, nil) == 0 {
		fmt.Println("Verification FAILED.")
	}
	return append([]int32(nil), vec.Vec_IntArray(p.vGates)...)
}
func Abc_ResubDumpProblem(path string, divisors [][]word, nWords int32) {
	if nWords <= 0 {
		panic("gia: invalid resubstitution word count")
	}
	sims := vec.Vec_WrdAlloc(len(divisors) * int(nWords))
	for _, d := range divisors {
		if len(d) < int(nWords) {
			panic("gia: short resubstitution divisor")
		}
		for _, v := range d[:nWords] {
			vec.Vec_WrdPush(sims, v)
		}
	}
	vec.Vec_WrdDumpHex(path, sims, int(nWords), 1)
}
func Gia_ManDeriveDivs(sims *Vec_Wrd_t, nWords int32) *Vec_Ptr_t {
	if nWords <= 0 || vec.Vec_WrdSize(sims)%int(nWords) != 0 {
		panic("gia: invalid resubstitution simulations")
	}
	words := vec.Vec_WrdArray(sims)
	result := vec.Vec_PtrAlloc(len(words) / int(nWords))
	for i := 0; i < len(words); i += int(nWords) {
		vec.Vec_PtrPush(result, words[i:i+int(nWords)])
	}
	return result
}
func Gia_ManCheckResub(divs *Vec_Ptr_t, nWords int32) {
	if vec.Vec_PtrSize(divs) < 5 {
		panic("gia: resubstitution check needs three divisors")
	}
	off, on := resubDiv(divs, 0, nWords), resubDiv(divs, 1, nWords)
	values := [8]int8{-1, -1, -1, -1, -1, -1, -1, -1}
	fmt.Println("Verifying resub:")
	for i := int32(0); i < 64*nWords; i++ {
		mask := word(1) << uint(i&63)
		o, f := off[i>>6]&mask != 0, on[i>>6]&mask != 0
		if !o && !f {
			continue
		}
		mint := 0
		for v := int32(0); v < 3; v++ {
			if resubDiv(divs, v+2, nWords)[i>>6]&mask != 0 {
				mint |= 1 << uint(v)
			}
		}
		value := int8(0)
		if f {
			value = 1
		}
		if values[mint] < 0 {
			values[mint] = value
		} else if values[mint] != value {
			fmt.Printf("Mismatch in pattern %d\n", i)
		}
	}
	fmt.Println("Finished verifying resub.")
}
func Gia_ManResub1(path string, nNodes, nSupp, nDivs, choice, useXor, verbose, veryVerbose int32) *Gia_Man_t {
	_, _, _, _, _ = nNodes, nSupp, nDivs, verbose, veryVerbose
	var nWords int
	sims := vec.Vec_WrdReadHex(path, &nWords, 1)
	if sims == nil || nWords <= 0 {
		return nil
	}
	divs := Gia_ManDeriveDivs(sims, int32(nWords))
	if vec.Vec_PtrSize(divs) >= 1<<14 {
		vec.Vec_PtrShrink(divs, (1<<14)-1)
	}
	p := Gia_ResbAlloc(int32(nWords))
	defer Gia_ResbFree(p)
	Gia_ManResubPerform(p, divs, int32(nWords), 100, 50, choice, useXor, 1, 1)
	if vec.Vec_IntSize(p.vGates) == 0 {
		fmt.Println("Decomposition did not succeed.")
		return nil
	}
	funcs := vec.Vec_WecAlloc(1)
	level := vec.Vec_WecPushLevel(funcs)
	for _, lit := range vec.Vec_IntArray(p.vGates) {
		vec.Vec_IntPush(level, lit)
	}
	return Gia_ManConstructFromGates(funcs, int32(vec.Vec_PtrSize(divs)))
}

// Gia_ManResub2 is unimplemented in the original C source.
func Gia_ManResub2(g *Gia_Man_t, nNodes, nSupp, nDivs, choice, useXor, verbose, veryVerbose int32) *Gia_Man_t {
	return nil
}
