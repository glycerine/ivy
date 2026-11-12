// Five-variable disjoint-support decomposition ported from ABC ivyDsd.c.
package ivy

import (
	"fmt"
	"io"
	"math/bits"
	"os"
	"sync/atomic"
)

type Ivy_DecType_t uint32

const (
	IVY_DEC_PI Ivy_DecType_t = iota
	IVY_DEC_CONST1
	IVY_DEC_BUF
	IVY_DEC_AND
	IVY_DEC_EXOR
	IVY_DEC_MUX
	IVY_DEC_MAJ
	IVY_DEC_PRIME
)

// The fixed bounded record retains the original 32-bit tree representation.
type Ivy_Dec_t struct {
	Type          Ivy_DecType_t
	fCompl, nFans uint32
	fans          [6]uint32
}

func Ivy_DecToInt(n Ivy_Dec_t) int32 {
	if n.Type > 15 || n.fCompl > 1 || n.nFans > 6 {
		panic("ivy: invalid decomposition record")
	}
	value := uint32(n.Type) | n.fCompl<<4 | n.nFans<<5
	for i, fan := range n.fans {
		if fan > 15 {
			panic("ivy: decomposition literal exceeds packed representation")
		}
		value |= fan << uint(8+4*i)
	}
	return int32(value)
}
func Ivy_IntToDec(value int32) Ivy_Dec_t {
	v := uint32(value)
	n := Ivy_Dec_t{Type: Ivy_DecType_t(v & 15), fCompl: v >> 4 & 1, nFans: v >> 5 & 7}
	for i := range n.fans {
		n.fans[i] = v >> uint(8+4*i) & 15
	}
	return n
}
func Ivy_DecClear(n *Ivy_Dec_t) { *n = Ivy_Dec_t{} }

var ivyDsdMasks = [6][2]uint32{{0x55555555, 0xaaaaaaaa}, {0x33333333, 0xcccccccc}, {0x0f0f0f0f, 0xf0f0f0f0}, {0x00ff00ff, 0xff00ff00}, {0x0000ffff, 0xffff0000}, {0, 0xffffffff}}

func Ivy_TruthWordCountOnes(value uint32) int32 { return int32(bits.OnesCount32(value)) }
func Ivy_TruthCofactorIsConst(truth uint32, variable, cofactor, constant int32) int32 {
	masked := truth & ivyDsdMasks[variable][cofactor]
	if constant == 0 {
		return ivyBool(masked == 0)
	}
	return ivyBool(masked == ivyDsdMasks[variable][cofactor])
}
func Ivy_TruthCofactorIsOne(truth uint32, variable int32) int32 {
	return ivyBool(truth&ivyDsdMasks[variable][0] == 0)
}

// Literal 2*v selects the positive cofactor; 2*v+1 selects the negative.
func Ivy_TruthCofactor(truth uint32, literal int32) uint32 {
	if literal < 0 || literal >= 10 {
		panic("ivy: invalid cofactor literal")
	}
	cofactor := int32(1)
	if literal&1 != 0 {
		cofactor = 0
	}
	result := truth & ivyDsdMasks[literal>>1][cofactor]
	shift := uint(1) << uint(literal>>1)
	if literal&1 != 0 {
		return result | result<<shift
	}
	return result | result>>shift
}
func Ivy_TruthCofactor2(truth uint32, a, b int32) uint32 {
	return Ivy_TruthCofactor(Ivy_TruthCofactor(truth, a), b)
}
func Ivy_TruthDepends(truth uint32, variable int32) int32 {
	return ivyBool(Ivy_TruthCofactor(truth, variable<<1) != Ivy_TruthCofactor(truth, variable<<1|1))
}
func Ivy_DecSetVar(n *Ivy_Dec_t, index int32, literal uint32) {
	if index < 0 || index >= 6 || literal > 15 {
		panic("ivy: invalid decomposition fanin")
	}
	n.fans[index] = literal
}
func Ivy_DecGetVar(n *Ivy_Dec_t, index int32) uint32 {
	if index < 0 || index >= 6 {
		panic("ivy: invalid decomposition fanin index")
	}
	return n.fans[index]
}
func Ivy_TruthDsd(truth uint32, tree *[]int32) int32 {
	*tree = append((*tree)[:0], 0, 0, 0, 0, 0)
	if truth == 0 || truth == ^uint32(0) {
		*tree = append(*tree, Ivy_DecToInt(Ivy_Dec_t{Type: IVY_DEC_CONST1, fCompl: uint32(ivyBool(truth == 0))}))
		return 1
	}
	root := Ivy_TruthDecompose_rec(truth, tree)
	if root < 0 {
		return 0
	}
	if root>>1 < 5 {
		n := Ivy_Dec_t{Type: IVY_DEC_BUF, fCompl: uint32(root & 1)}
		n.fans[0] = uint32(root &^ 1)
		*tree = append(*tree, Ivy_DecToInt(n))
	} else if root&1 != 0 {
		index := len(*tree) - 1
		n := Ivy_IntToDec((*tree)[index])
		n.fCompl = 1
		(*tree)[index] = Ivy_DecToInt(n)
	}
	if Ivy_TruthDsdCompute(*tree) != truth {
		panic("ivy: decomposition verification failed")
	}
	return 1
}
func Ivy_TruthDecompose_rec(truth uint32, tree *[]int32) int32 {
	var support, zeros, ones, parities []int32
	for variable := int32(0); variable < 5; variable++ {
		switch {
		case Ivy_TruthCofactorIsConst(truth, variable, 0, 0) != 0:
			zeros = append(zeros, variable<<1)
		case Ivy_TruthCofactorIsConst(truth, variable, 1, 0) != 0:
			zeros = append(zeros, variable<<1|1)
		case Ivy_TruthCofactorIsConst(truth, variable, 0, 1) != 0:
			ones = append(ones, variable<<1)
		case Ivy_TruthCofactorIsConst(truth, variable, 1, 1) != 0:
			ones = append(ones, variable<<1|1)
		default:
			a, b := Ivy_TruthCofactor(truth, variable<<1|1), Ivy_TruthCofactor(truth, variable<<1)
			if a == ^b {
				parities = append(parities, variable<<1)
			} else if a != b {
				support = append(support, variable)
			}
		}
	}
	if len(zeros) == 1 && len(support) == 0 {
		return zeros[0]
	}
	if len(zeros)+len(ones)+len(parities) == 0 {
		return Ivy_TruthRecognizeMuxMaj(truth, support, int32(len(support)), tree)
	}
	n := Ivy_Dec_t{Type: IVY_DEC_AND}
	variables := zeros
	compl := int32(0)
	if len(zeros) == 0 {
		if len(ones) > 0 {
			variables = ones
			compl = 1
			truth = ^truth
		} else {
			variables = parities
			n.Type = IVY_DEC_EXOR
		}
	}
	n.nFans = uint32(len(variables))
	if len(support) > 0 {
		n.nFans++
	}
	cofactor := truth
	for i, literal := range variables {
		cofactor = Ivy_TruthCofactor(cofactor, literal)
		Ivy_DecSetVar(&n, int32(i), uint32(literal))
	}
	if n.Type == IVY_DEC_EXOR && n.nFans&1 == 0 {
		compl ^= 1
	}
	if len(support) > 0 {
		child := Ivy_TruthDecompose_rec(cofactor, tree)
		if child < 0 {
			return -1
		}
		if n.Type == IVY_DEC_EXOR && child&1 != 0 {
			compl ^= 1
			child ^= 1
		}
		Ivy_DecSetVar(&n, int32(len(variables)), uint32(child))
	} else if n.Type == IVY_DEC_EXOR && cofactor == 0 {
		compl ^= 1
	}
	*tree = append(*tree, Ivy_DecToInt(n))
	return int32(len(*tree)-1)<<1 | compl
}
func Ivy_TruthRecognizeMuxMaj(truth uint32, support []int32, count int32, tree *[]int32) int32 {
	if count < 3 || int(count) > len(support) {
		return -1
	}
	support = support[:count]
	n := Ivy_Dec_t{Type: IVY_DEC_MUX, nFans: 3}
	for _, control := range support {
		negative, positive := Ivy_TruthCofactor(truth, control<<1|1), Ivy_TruthCofactor(truth, control<<1)
		disjoint := true
		for _, variable := range support {
			if variable != control && Ivy_TruthDepends(negative, variable) != 0 && Ivy_TruthDepends(positive, variable) != 0 {
				disjoint = false
				break
			}
		}
		if !disjoint {
			continue
		}
		a := Ivy_TruthDecompose_rec(negative, tree)
		if a < 0 {
			break
		}
		b := Ivy_TruthDecompose_rec(positive, tree)
		if b < 0 {
			break
		}
		n.fans[0], n.fans[1], n.fans[2] = uint32(control<<1), uint32(b), uint32(a)
		*tree = append(*tree, Ivy_DecToInt(n))
		return int32(len(*tree)-1) << 1
	}
	if count > 3 || bits.OnesCount32(truth) != 16 {
		return -1
	}
	n.Type = IVY_DEC_MAJ
	var counts [3]int32
	for assignment := 0; assignment < 8; assignment++ {
		index := uint(0)
		for k, variable := range support {
			if assignment&(1<<k) != 0 {
				index |= 1 << uint(variable)
			}
		}
		if truth&(1<<index) == 0 {
			continue
		}
		for k := range support {
			if assignment&(1<<k) != 0 {
				counts[k]++
			}
		}
	}
	for k, variable := range support {
		if counts[k] != 1 && counts[k] != 3 {
			return -1
		}
		n.fans[k] = uint32(variable<<1 | ivyBool(counts[k] == 1))
	}
	*tree = append(*tree, Ivy_DecToInt(n))
	return int32(len(*tree)-1) << 1
}

// A packed tree is topological: every child index is strictly smaller.
func ivyDsdNode(tree []int32, index int32) Ivy_Dec_t {
	if index < 0 || int(index) >= len(tree) {
		panic("ivy: decomposition node out of bounds")
	}
	n := Ivy_IntToDec(tree[index])
	if n.Type == IVY_DEC_PI {
		if index >= 5 {
			panic("ivy: invalid decomposition input")
		}
		return n
	}
	if n.Type == IVY_DEC_CONST1 {
		return n
	}
	count := int(n.nFans)
	if n.Type == IVY_DEC_BUF {
		count = 1
	} else if n.Type == IVY_DEC_MUX || n.Type == IVY_DEC_MAJ {
		if count != 3 {
			panic("ivy: invalid ternary decomposition")
		}
	} else if n.Type != IVY_DEC_AND && n.Type != IVY_DEC_EXOR {
		panic("ivy: unsupported decomposition node")
	}
	if count < 1 || count > 6 {
		panic("ivy: invalid decomposition fanin count")
	}
	for _, literal := range n.fans[:count] {
		if int32(literal>>1) >= index {
			panic("ivy: cyclic or non-topological decomposition")
		}
	}
	return n
}
func Ivy_TruthDsdCompute_rec(index int32, tree []int32) uint32 {
	n := ivyDsdNode(tree, index)
	var result uint32
	switch n.Type {
	case IVY_DEC_CONST1:
		result = ^uint32(0)
	case IVY_DEC_PI:
		result = ivyDsdMasks[index][1]
	case IVY_DEC_BUF:
		result = Ivy_TruthDsdCompute_rec(int32(n.fans[0]>>1), tree)
	case IVY_DEC_AND, IVY_DEC_EXOR:
		if n.Type == IVY_DEC_AND {
			result = ^uint32(0)
		}
		for _, literal := range n.fans[:n.nFans] {
			child := Ivy_TruthDsdCompute_rec(int32(literal>>1), tree)
			if literal&1 != 0 {
				child = ^child
			}
			if n.Type == IVY_DEC_AND {
				result &= child
			} else {
				result ^= child
			}
		}
	case IVY_DEC_MUX, IVY_DEC_MAJ:
		var values [3]uint32
		for i, literal := range n.fans[:3] {
			values[i] = Ivy_TruthDsdCompute_rec(int32(literal>>1), tree)
			if literal&1 != 0 {
				values[i] = ^values[i]
			}
		}
		a, b, c := values[0], values[1], values[2]
		if n.Type == IVY_DEC_MUX {
			result = a&b | ^a&c
		} else {
			result = a&b | a&c | b&c
		}
	}
	if n.fCompl != 0 {
		return ^result
	}
	return result
}
func Ivy_TruthDsdCompute(tree []int32) uint32 {
	return Ivy_TruthDsdCompute_rec(int32(len(tree)-1), tree)
}
func Ivy_TruthDsdPrint_rec(out io.Writer, index int32, tree []int32) {
	n := ivyDsdNode(tree, index)
	switch n.Type {
	case IVY_DEC_CONST1:
		fmt.Fprint(out, "Const1")
	case IVY_DEC_PI:
		fmt.Fprintf(out, "%c", 'a'+index)
	case IVY_DEC_BUF:
		Ivy_TruthDsdPrint_rec(out, int32(n.fans[0]>>1), tree)
	default:
		name := map[Ivy_DecType_t]string{IVY_DEC_AND: "AND", IVY_DEC_EXOR: "EXOR", IVY_DEC_MUX: "MUX", IVY_DEC_MAJ: "MAJ"}[n.Type]
		fmt.Fprint(out, name, "(")
		for i, literal := range n.fans[:n.nFans] {
			if i > 0 {
				fmt.Fprint(out, ",")
			}
			Ivy_TruthDsdPrint_rec(out, int32(literal>>1), tree)
			if literal&1 != 0 {
				fmt.Fprint(out, "'")
			}
		}
		fmt.Fprint(out, ")")
	}
	if n.fCompl != 0 {
		fmt.Fprint(out, "'")
	}
}
func Ivy_TruthDsdPrint(out io.Writer, tree []int32) {
	fmt.Fprint(out, "F = ")
	Ivy_TruthDsdPrint_rec(out, int32(len(tree)-1), tree)
	fmt.Fprintln(out)
}
func Ivy_ManDsdConstruct_rec(p *Ivy_Man_t, front []int32, index int32, tree []int32) *Ivy_Obj_t {
	n := ivyDsdNode(tree, index)
	var result *Ivy_Obj_t
	switch n.Type {
	case IVY_DEC_CONST1:
		result = p.pConst1
	case IVY_DEC_PI:
		if int(index) >= len(front) {
			panic("ivy: decomposition frontier too small")
		}
		result = Ivy_ManObj(p, front[index])
	case IVY_DEC_BUF:
		result = Ivy_ManDsdConstruct_rec(p, front, int32(n.fans[0]>>1), tree)
	default:
		children := make([]*Ivy_Obj_t, n.nFans)
		for i, literal := range n.fans[:n.nFans] {
			children[i] = Ivy_NotCond(Ivy_ManDsdConstruct_rec(p, front, int32(literal>>1), tree), int32(literal&1))
		}
		switch n.Type {
		case IVY_DEC_AND:
			result = Ivy_Multi(p, children, int32(len(children)), IVY_AND)
		case IVY_DEC_EXOR:
			result = Ivy_Multi(p, children, int32(len(children)), IVY_EXOR)
		case IVY_DEC_MUX:
			result = Ivy_Mux(p, children[0], children[1], children[2])
		case IVY_DEC_MAJ:
			result = Ivy_Maj(p, children[0], children[1], children[2])
		}
	}
	return Ivy_NotCond(result, int32(n.fCompl))
}
func Ivy_ManDsdConstruct(p *Ivy_Man_t, front []int32, tree []int32) *Ivy_Obj_t {
	for i, leaf := range front {
		front[i] = Ivy_LeafId(leaf)
	}
	return Ivy_ManDsdConstruct_rec(p, front, int32(len(tree)-1), tree)
}
func Ivy_TruthDsdComputePrint(truth uint32) {
	var tree []int32
	if Ivy_TruthDsd(truth, &tree) != 0 {
		Ivy_TruthDsdPrint(os.Stdout, tree)
	} else {
		fmt.Println("Undecomposable")
	}
}

var ivyDsdPrintCounter atomic.Uint64

func Ivy_TruthTestOne(truth uint32) {
	var tree []int32
	if Ivy_TruthDsd(truth, &tree) != 0 {
		fmt.Printf("%5d : %032b  ", ivyDsdPrintCounter.Add(1)-1, truth)
		Ivy_TruthDsdPrint(os.Stdout, tree)
	}
}
