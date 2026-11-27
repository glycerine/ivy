package abci

// Port of ioWriteBlif.c's special writers (write_blif -j / -S / -a):
// Io_WriteBlifSpecial, Io_WriteBlifInt, Io_NtkWriteNodeInt,
// Io_NtkWriteNodeIntStruct, Io_NtkWriteModelIntStruct, Io_NtkDeriveSop.

import (
	"fmt"
	"os"
	"strings"

	"github.com/glycerine/abcgo/base/abc"
	ioPkg "github.com/glycerine/abcgo/base/io"
	"github.com/glycerine/abcgo/bool/kit"
	ifpkg "github.com/glycerine/abcgo/map/if"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/mem"
	"github.com/glycerine/abcgo/misc/vec"
)

func init() { ioPkg.RegisterSpecialBLIFWriter(Io_WriteBlifSpecial) }

// Io_WriteBlifSpecial converts logic network n to SOPs and a netlist and
// writes it with Io_WriteBlifInt.
func Io_WriteBlifSpecial(n *abc.Network, fileName, lutStruct string, useHie bool) error {
	if abc.Abc_NtkIsLogic(n) == 0 {
		return fmt.Errorf("write_blif: the network is not a logic network")
	}
	abc.Abc_NtkToSop(n, -1, 1_000_000_000)
	netlist := abc.Abc_NtkToNetlist(n)
	if netlist == nil {
		fmt.Printf("Writing BLIF has failed.\n")
		return nil
	}
	defer abc.Abc_NtkDelete(netlist)
	return Io_WriteBlifInt(netlist, fileName, lutStruct, lutStruct != "" && useHie)
}

// Io_WriteBlifInt writes the netlist with each node decomposed.
func Io_WriteBlifInt(n *abc.Network, fileName, lutStruct string, useHie bool) error {
	pool, _ := n.FunctionManager.(*mem.Mem_Flex_t)
	cover := vec.Vec_IntAlloc(1 << 16)
	nodes := n.ObjectsOfKind(abc.ObjectNode)
	err := ioPkg.WriteBLIFSpecialFile(n, fileName, func(out *strings.Builder, name func(*abc.Object) string) {
		w := &blifSpecial{out: out, name: name, pool: pool, cover: cover}
		if useHie {
			out.WriteString("\n")
			for _, node := range nodes {
				w.writeNodeSubckt(node)
			}
			out.WriteString(".end\n\n")
			for _, node := range nodes {
				w.writeModelIntStruct(node, lutStruct)
			}
			out.WriteString("\n")
			return
		}
		for _, node := range nodes {
			if lutStruct != "" {
				w.writeNodeIntStruct(node, lutStruct)
			} else {
				w.writeNodeInt(node)
			}
		}
		out.WriteString(".end\n\n")
	})
	if os.IsNotExist(err) || os.IsPermission(err) {
		fmt.Printf("Io_WriteBlifInt(): Cannot open the output file.\n")
		return nil
	}
	return err
}

type blifSpecial struct {
	out   *strings.Builder
	name  func(*abc.Object) string
	pool  *mem.Mem_Flex_t
	cover *vec.Vec_Int_t
}

func (w *blifSpecial) printf(format string, args ...any) { fmt.Fprintf(w.out, format, args...) }

func (w *blifSpecial) outName(node *abc.Object) string { return w.name(abc.Abc_ObjFanout0(node)) }

func (w *blifSpecial) faninName(node *abc.Object, i int32) string {
	return w.name(abc.Abc_ObjFanin(node, i))
}

func (w *blifSpecial) sop(node *abc.Object) string {
	s, _ := node.Data.(string)
	return strings.TrimRight(s, "\x00")
}

// deriveSop is Io_NtkDeriveSop.
func (w *blifSpecial) deriveSop(truth uint64, nVars int32) string {
	var cubes []uint32
	ret := kit.Kit_TruthIsop([]uint32{uint32(truth), uint32(truth >> 32)}, nVars, &cubes, 1)
	assert(ret == 0 || ret == 1, "Io_NtkDeriveSop: ISOP result")
	if len(cubes) == 0 || len(cubes) == 1 && cubes[0] == 0 {
		assert(ret == 0, "Io_NtkDeriveSop: constant cover")
		if len(cubes) == 0 {
			return " 0\n"
		}
		return " 1\n"
	}
	vec.Vec_IntClear(w.cover)
	for _, c := range cubes {
		vec.Vec_IntPush(w.cover, int32(c))
	}
	s := abc.Abc_SopCreateFromIsop(w.pool, nVars, w.cover)
	if ret != 0 {
		s = abc.Abc_SopComplement(s)
	}
	return s
}

// writeNodeSubckt is Io_NtkWriteNodeSubckt with Io_NtkWriteSubcktFanins.
func (w *blifSpecial) writeNodeSubckt(node *abc.Object) {
	w.printf(".subckt")
	lineLength, nameCounter := 6, 0
	w.printf(" m%d", abc.Abc_ObjId(node))
	for i := int32(0); i < int32(node.FaninNum()); i++ {
		name := w.faninName(node, i)
		added := len(name) + 3
		if nameCounter != 0 && lineLength+added+3 > 78 {
			w.printf(" \\\n")
			lineLength, nameCounter = 0, 0
		}
		w.printf(" %c=%s", 'a'+i, name)
		lineLength += added
		nameCounter++
	}
	name := w.outName(node)
	if nameCounter != 0 && lineLength+len(name)+3 > 75 {
		w.printf(" \\\n")
	}
	w.printf(" %c=%s", 'o', name)
	w.printf("\n")
}

// writeNodeInt is Io_NtkWriteNodeInt: nodes of up to 4 inputs as they
// are, larger ones as a MUX of two cofactors or a cascade of two LUTs.
func (w *blifSpecial) writeNodeInt(node *abc.Object) {
	nVars := int32(node.FaninNum())
	if nVars > 7 {
		fmt.Printf("Node \"%s\" has more than 7 inputs. Writing BLIF has failed.\n", w.outName(node))
		return
	}
	w.printf("\n")
	if nVars <= 4 {
		w.printf(".names")
		for i := int32(0); i < nVars; i++ {
			w.printf(" %s", w.faninName(node, i))
		}
		w.printf(" %s\n", w.outName(node))
		w.printf("%s", w.sop(node))
		return
	}
	var uTruth6 uint64
	var uTruth7, cofs6 [2]uint64
	var cofs7 [2][2]uint64
	var nVarsMin [2]int32
	var pVars [2][10]int32
	for i := int32(0); i < nVars; i++ {
		pVars[0][i], pVars[1][i] = i, i
	}
	var iVar int32
	if nVars == 7 {
		abc.Abc_SopToTruth7(w.sop(node), nVars, uTruth7[:])
		iVar = ifpkg.If_Dec7PickBestMux(uTruth7[:], cofs7[0][:], cofs7[1][:])
	} else {
		uTruth6 = abc.Abc_SopToTruth(w.sop(node), nVars)
		iVar = ifpkg.If_Dec6PickBestMux(uTruth6, cofs6[:])
	}
	if iVar >= 0 {
		if nVars == 7 {
			ifpkg.If_Dec7MinimumBase(cofs7[0][:], pVars[0][:], nVars, &nVarsMin[0])
			ifpkg.If_Dec7MinimumBase(cofs7[1][:], pVars[1][:], nVars, &nVarsMin[1])
		} else {
			cofs6[0] = ifpkg.If_Dec6MinimumBase(cofs6[0], pVars[0][:], nVars, &nVarsMin[0])
			cofs6[1] = ifpkg.If_Dec6MinimumBase(cofs6[1], pVars[1][:], nVars, &nVarsMin[1])
		}
		assert(nVarsMin[0] < 5 && nVarsMin[1] < 5, "Io_NtkWriteNodeInt: cofactor support")
		out := w.outName(node)
		w.printf(".names %s %s_cascade0 %s_cascade1 %s\n", w.faninName(node, iVar), out, out, out)
		w.printf("1-1 1\n01- 1\n")
		for c := 0; c < 2; c++ {
			t := cofs6[c]
			if nVars == 7 {
				t = cofs7[c][0]
			}
			sop := w.deriveSop(t, nVarsMin[c])
			w.printf(".names")
			for i := int32(0); i < nVarsMin[c]; i++ {
				w.printf(" %s", w.faninName(node, pVars[c][i]))
			}
			w.printf(" %s_cascade%d\n", out, c)
			w.printf("%s", sop)
		}
		return
	}
	assert(nVars == 6 || nVars == 7, "Io_NtkWriteNodeInt: cascade size")
	var z uint64
	if nVars == 7 {
		z = ifpkg.If_Dec7Perform(uTruth7[:], 1)
	} else {
		z = ifpkg.If_Dec6Perform(uTruth6, 1)
	}
	if z == 0 {
		fmt.Printf("Node \"%s\" is not decomposable. Writing BLIF has failed.\n", w.outName(node))
		return
	}
	out := w.outName(node)
	for c := 1; c >= 0; c-- {
		shift := uint(0)
		if c != 0 {
			shift = 32
		}
		uTruth7[c] = (z >> shift) & 0xffff
		uTruth7[c] |= uTruth7[c] << 16
		uTruth7[c] |= uTruth7[c] << 32
		for i := 0; i < 4; i++ {
			pVars[c][i] = int32((z >> uint(c*32+16+4*i)) & 7)
		}
		cofs6[c] = ifpkg.If_Dec6MinimumBase(uTruth7[c], pVars[c][:], 4, &nVarsMin[c])
		w.printf(".names")
		for i := int32(0); i < nVarsMin[c]; i++ {
			if pVars[c][i] == 7 {
				w.printf(" %s_cascade", out)
			} else {
				w.printf(" %s", w.faninName(node, pVars[c][i]))
			}
		}
		suffix := "_cascade"
		if c != 0 {
			suffix = ""
		}
		w.printf(" %s%s\n", out, suffix)
		w.printf("%s", w.deriveSop(cofs6[c], nVarsMin[c]))
	}
}

// blifTruthVars is the static TruthStore of Io_NtkWriteNodeIntStruct.
var blifTruthVars = func() [][]uint64 {
	const nWords = 1 << 10
	truth6 := [6]uint64{0xAAAAAAAAAAAAAAAA, 0xCCCCCCCCCCCCCCCC, 0xF0F0F0F0F0F0F0F0,
		0xFF00FF00FF00FF00, 0xFFFF0000FFFF0000, 0xFFFFFFFF00000000}
	vars := make([][]uint64, 16)
	for i := range vars {
		vars[i] = make([]uint64, nWords)
		for k := range vars[i] {
			if i < 6 {
				vars[i][k] = truth6[i]
			} else if (k>>(i-6))&1 != 0 {
				vars[i][k] = ^uint64(0)
			}
		}
	}
	return vars
}()

// lutStructParse checks the LUT structure as both struct writers do.
func lutStructParse(nLeaves int32, str string) (leaf, leaf2, root int32, ok bool) {
	length := len(str)
	if length != 2 && length != 3 {
		fmt.Printf("Wrong LUT struct (%s)\n", str)
		return 0, 0, 0, false
	}
	for i := 0; i < length; i++ {
		if d := int32(str[i]) - '0'; d < 3 || d > 6 {
			fmt.Printf("The LUT size (%d) should belong to {3,4,5,6}.\n", d)
			return 0, 0, 0, false
		}
	}
	leaf = int32(str[0]) - '0'
	if length == 3 {
		leaf2 = int32(str[1]) - '0'
	}
	root = int32(str[length-1]) - '0'
	extra2 := int32(0)
	if leaf2 != 0 {
		extra2 = leaf2 - 1
	}
	if nLeaves > leaf-1+extra2+root {
		fmt.Printf("The node size (%d) is too large for the LUT structure %s.\n", nLeaves, str)
		return 0, 0, 0, false
	}
	return leaf, leaf2, root, true
}

type lutDecomp struct {
	lut0, lut1, lut2    [32]byte
	func0, func1, func2 uint64
	constant            int32 // -1: not constant
}

// decomposeStruct derives the node's truth table and splits it into the
// LUT structure; ok is false (with C's messages printed) when it cannot.
func (w *blifSpecial) decomposeStruct(node *abc.Object, nLeaves, leaf, leaf2, root int32, length int) (d lutDecomp, ok bool) {
	var cube, res [1 << 10]uint64
	abc.Abc_SopToTruthBig(w.sop(node), nLeaves, blifTruthVars, cube[:], res[:])
	words := make([]uint32, 2*len(res))
	for i, x := range res {
		words[2*i], words[2*i+1] = uint32(x), uint32(x>>32)
	}
	d.constant = -1
	if kit.Kit_TruthIsConst0(words, nLeaves) != 0 || kit.Kit_TruthIsConst1(words, nLeaves) != 0 {
		d.constant = kit.Kit_TruthIsConst1(words, nLeaves)
		return d, true
	}
	var found int32
	if length == 2 {
		found = ifpkg.If_CluCheckExt(nil, res[:], nLeaves, leaf, root, d.lut0[:], d.lut1[:], &d.func0, &d.func1)
	} else {
		found = ifpkg.If_CluCheckExt3(nil, res[:], nLeaves, leaf, leaf2, root, d.lut0[:], d.lut1[:], d.lut2[:], &d.func0, &d.func1, &d.func2)
	}
	if found == 0 {
		extra.Extra_PrintHex(os.Stdout, words, nLeaves)
		fmt.Printf("    ")
		kit.Kit_DsdPrintFromTruth(words, nLeaves)
		fmt.Printf("\n")
		fmt.Printf("Node \"%s\" is not decomposable. Writing BLIF has failed.\n", w.outName(node))
		return d, false
	}
	return d, true
}

// writeStructLuts writes the leaf LUT(s) and the root of a decomposition;
// in(i) names fanin i, lut(k) names the k-th leaf LUT, out the root.
func (w *blifSpecial) writeStructLuts(d *lutDecomp, nLeaves int32, length int, in func(int32) string, lut func(int) string, out string) {
	w.printf(".names")
	for i := 0; i < int(d.lut1[0]); i++ {
		w.printf(" %s", in(int32(d.lut1[2+i])))
	}
	w.printf(" %s\n", lut(1))
	w.printf("%s", w.deriveSop(d.func1, int32(d.lut1[0])))
	if length == 3 && d.lut2[0] > 0 {
		w.printf(".names")
		for i := 0; i < int(d.lut2[0]); i++ {
			if v := int32(d.lut2[2+i]); v == nLeaves {
				w.printf(" %s", lut(1))
			} else {
				w.printf(" %s", in(v))
			}
		}
		w.printf(" %s\n", lut(2))
		w.printf("%s", w.deriveSop(d.func2, int32(d.lut2[0])))
	}
	w.printf(".names")
	for i := 0; i < int(d.lut0[0]); i++ {
		switch v := int32(d.lut0[2+i]); v {
		case nLeaves:
			w.printf(" %s", lut(1))
		case nLeaves + 1:
			w.printf(" %s", lut(2))
		default:
			w.printf(" %s", in(v))
		}
	}
	w.printf(" %s\n", out)
	w.printf("%s", w.deriveSop(d.func0, int32(d.lut0[0])))
}

// writeNodeIntStruct is Io_NtkWriteNodeIntStruct.
func (w *blifSpecial) writeNodeIntStruct(node *abc.Object, str string) {
	nLeaves := int32(node.FaninNum())
	leaf, leaf2, root, ok := lutStructParse(nLeaves, str)
	if !ok {
		return
	}
	w.printf("\n")
	out := w.outName(node)
	if nLeaves <= max(leaf2, leaf, root) {
		w.printf(".names")
		for i := int32(0); i < nLeaves; i++ {
			w.printf(" %s", w.faninName(node, i))
		}
		w.printf(" %s\n", out)
		w.printf("%s", w.sop(node))
		return
	}
	d, ok := w.decomposeStruct(node, nLeaves, leaf, leaf2, root, len(str))
	if !ok {
		return
	}
	if d.constant >= 0 {
		w.printf(".names %s\n %d\n", out, d.constant)
		return
	}
	w.writeStructLuts(&d, nLeaves, len(str),
		func(i int32) string { return w.faninName(node, i) },
		func(k int) string { return fmt.Sprintf("%s_lut%d", out, k) }, out)
}

// writeModelIntStruct is Io_NtkWriteModelIntStruct.
func (w *blifSpecial) writeModelIntStruct(node *abc.Object, str string) {
	nLeaves := int32(node.FaninNum())
	w.printf("\n")
	w.printf(".model m%d\n", abc.Abc_ObjId(node))
	w.printf(".inputs")
	for i := int32(0); i < nLeaves; i++ {
		w.printf(" %c", 'a'+i)
	}
	w.printf("\n")
	w.printf(".outputs o\n")
	leaf, leaf2, root, ok := lutStructParse(nLeaves, str)
	if !ok {
		return
	}
	in := func(i int32) string { return string(rune('a' + i)) }
	if nLeaves <= max(leaf2, leaf, root) {
		w.printf(".names")
		for i := int32(0); i < nLeaves; i++ {
			w.printf(" %s", in(i))
		}
		w.printf(" %s\n", "o")
		w.printf("%s", w.sop(node))
		w.printf(".end\n")
		return
	}
	d, ok := w.decomposeStruct(node, nLeaves, leaf, leaf2, root, len(str))
	if !ok {
		return
	}
	if d.constant >= 0 {
		w.printf(".names %s\n %d\n", "o", d.constant)
		w.printf(".end\n")
		return
	}
	w.writeStructLuts(&d, nLeaves, len(str), in,
		func(k int) string { return fmt.Sprintf("lut%d", k) }, "o")
	w.printf(".end\n")
}
