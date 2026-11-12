package abc

import (
	"fmt"
	"io"
	"os"

	"github.com/glycerine/abcgo/opt/fxu"
)

// BuildPLA extracts shared logic from output covers and strashes it into an AIG.
func BuildPLA(covers []string, inputs int) (*Network, error) {
	if inputs < 0 {
		return nil, fmt.Errorf("abc: negative PLA input count")
	}
	for i, cover := range covers {
		if !checkSOPSyntax(cover, inputs) {
			return nil, fmt.Errorf("abc: invalid PLA cover %d", i)
		}
	}
	logic := Abc_NtkAlloc(ABC_NTK_LOGIC, ABC_FUNC_SOP, 1)
	defer Abc_NtkDelete(logic)
	logic.Name = "pla"
	for i := 0; i < inputs; i++ {
		logic.NewObject(ObjectPI)
	}
	for range covers {
		logic.NewObject(ObjectPO)
	}
	Abc_NtkAddDummyPiNames(logic)
	Abc_NtkAddDummyPoNames(logic)
	for i, cover := range covers {
		node := logic.NewObject(ObjectNode)
		node.Data = cover
		for _, input := range logic.ObjectsOfKind(ObjectPI) {
			Abc_ObjAddFanin(node, input)
		}
		Abc_ObjAddFanin(logic.ObjectsOfKind(ObjectPO)[i], node)
	}
	if Abc_NtkCheck(logic) == 0 {
		return nil, fmt.Errorf("abc: PLA logic network check failed")
	}
	params := fxu.NewParams()
	fxu.Abc_NtkFastExtract(logic, params)
	if params.LastError != nil {
		return nil, fmt.Errorf("abc: PLA extraction: %w", params.LastError)
	}
	if Abc_NtkSopToAig(logic) == 0 {
		return nil, fmt.Errorf("abc: PLA local AIG conversion failed")
	}
	dst := Abc_NtkStartFrom(logic, ABC_NTK_STRASH, ABC_FUNC_AIG)
	for _, node := range Abc_NtkDfs(logic, 0) {
		setCopy(node, Abc_ConvertAigToAig(dst, node))
	}
	Abc_NtkFinalize(logic, dst)
	return dst, nil
}

func Abc_NtkFromPla(covers []string, inputs, outputs int32) *Network {
	if outputs < 0 || int(outputs) != len(covers) {
		panic("abc: PLA output count mismatch")
	}
	n, err := BuildPLA(covers, int(inputs))
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return n
}

// writePLATestBLIF writes the small demonstration network in BLIF form.
func writePLATestBLIF(w io.Writer, aig *Network) error {
	if _, err := fmt.Fprint(w, ".model pla\n.inputs"); err != nil {
		return err
	}
	for _, o := range aig.ObjectsOfKind(ObjectPI) {
		if _, err := fmt.Fprintf(w, " %s", Abc_ObjName(o)); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprint(w, "\n.outputs"); err != nil {
		return err
	}
	for _, o := range aig.ObjectsOfKind(ObjectPO) {
		if _, err := fmt.Fprintf(w, " %s", Abc_ObjName(o)); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, ".names %s\n1\n", Abc_ObjName(aig.constant)); err != nil {
		return err
	}
	for _, node := range Abc_NtkDfs(aig, 0) {
		if len(node.fanins) != 2 {
			return fmt.Errorf("abc: invalid PLA AIG node")
		}
		a, b := node.fanins[0], node.fanins[1]
		if _, err := fmt.Fprintf(w, ".names %s %s %s\n%s", Abc_ObjName(a.Object), Abc_ObjName(b.Object), Abc_ObjName(node), Abc_SopCreateAnd2(nil, abcBool(a.Complement), abcBool(b.Complement))); err != nil {
			return err
		}
	}
	for _, output := range aig.ObjectsOfKind(ObjectPO) {
		edge := output.fanins[0]
		value := 1
		if edge.Complement {
			value = 0
		}
		if _, err := fmt.Fprintf(w, ".names %s %s\n%d 1\n", Abc_ObjName(edge.Object), Abc_ObjName(output), value); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(w, ".end")
	return err
}

func Abc_NtkFromPlaTest() {
	n := Abc_NtkFromPla([]string{"1000 1\n", "0110 1\n0011 1\n"}, 4, 2)
	if n == nil {
		return
	}
	defer Abc_NtkDelete(n)
	file, err := os.Create("temp.blif")
	if err == nil {
		err = writePLATestBLIF(file, n)
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
	}
	if err != nil {
		fmt.Println(err)
	}
}
