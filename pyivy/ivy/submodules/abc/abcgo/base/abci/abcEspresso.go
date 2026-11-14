package abci

import (
	"fmt"
	"strings"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/espresso"
	"github.com/glycerine/abcgo/misc/util"
)

type pset_family = *espresso.Family

// Abc_NtkEspresso minimizes the SOP attached to each internal logic node.
func Abc_NtkEspresso(n *Abc_Ntk_t, verbose int32) {
	_ = verbose // The C routine accepts this option but does not inspect it.
	if n == nil || abc.Abc_NtkIsLogic(n) == 0 {
		panic("Abc_NtkEspresso requires a logic network")
	}
	if abc.Abc_NtkHasMapping(n) != 0 {
		abc.Abc_NtkMapToSop(n)
	} else if abc.Abc_NtkHasBdd(n) != 0 && abc.Abc_NtkBddToSop(n, -1, util.ABC_INFINITY) == 0 {
		fmt.Printf("Abc_NtkEspresso(): Converting to SOPs has failed.\n")
		return
	}
	for _, node := range abc.Abc_NtkNodes(n) {
		Abc_NodeEspresso(node)
	}
}

func Abc_NodeEspresso(node *Abc_Obj_t) {
	if node == nil || abc.Abc_ObjIsNode(node) == 0 {
		panic("Abc_NodeEspresso requires an internal node")
	}
	sop, ok := abc.Abc_ObjData(node).(string)
	if !ok {
		return
	}
	nVars := abc.Abc_SopGetVarNum(sop)
	espresso.ConfigureCube(nVars, nil)
	complemented := abc.Abc_SopIsComplement(sop) != 0
	cover := Abc_EspressoMinimize(Abc_SopToEspresso(sop), nil)
	result := Abc_SopFromEspresso(cover)
	if complemented {
		result = abc.Abc_SopComplement(result)
	}
	abc.Abc_ObjSetData(node, result)
}

// Abc_SopToEspresso maps each input literal to the two-bit Espresso variable
// part. An omitted literal leaves both values in the cube.
func Abc_SopToEspresso(sop string) pset_family {
	if abc.Abc_SopIsConst0(sop) != 0 {
		return espresso.NewFamily(0, 0)
	}
	if abc.Abc_SopIsConst1(sop) != 0 {
		f := espresso.NewFamily(1, 0)
		f.Add(espresso.NewSet(0))
		return f
	}
	nVars := abc.Abc_SopGetVarNum(sop)
	espresso.ConfigureCube(nVars, nil)
	rows := strings.Split(strings.TrimSuffix(sop, "\n"), "\n")
	f := espresso.NewFamily(int32(len(rows)), 2*nVars)
	for _, row := range rows {
		fields := strings.Fields(row)
		if len(fields) == 0 {
			continue
		}
		cube := fields[0]
		set := espresso.NewSet(2 * nVars)
		for v := int32(0); v < 2*nVars; v++ {
			set.Insert(v)
		}
		for v := int32(0); v < nVars && int(v) < len(cube); v++ {
			switch cube[v] {
			case '0':
				set.Remove(2*v + 1)
			case '1':
				set.Remove(2 * v)
			}
		}
		f.Add(set)
	}
	return f
}

func Abc_SopFromEspresso(cover pset_family) string {
	if cover == nil {
		return ""
	}
	if cover.Count() == 0 {
		return abc.Abc_SopCreateConst0(nil)
	}
	nVars := cover.SetSize() / 2
	result := abc.Abc_SopStart(nil, cover.Count(), nVars)
	bytes := []byte(result)
	lineSize := int(nVars + 3)
	for i := int32(0); i < cover.Count(); i++ {
		set := cover.Row(i)
		for v := int32(0); v < nVars; v++ {
			zero, one := set.Contains(2*v), set.Contains(2*v+1)
			if !zero && one {
				bytes[int(i)*lineSize+int(v)] = '0'
			} else if zero && !one {
				bytes[int(i)*lineSize+int(v)] = '1'
			}
		}
	}
	return string(bytes)
}

func Abc_EspressoMinimize(onset, dcset pset_family) pset_family {
	return espresso.AbcMinimize(onset, dcset)
}
