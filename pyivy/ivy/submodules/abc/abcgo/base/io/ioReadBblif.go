package io

import (
	"fmt"
	"os"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/bbl"
)

type Bbl_Man_t = bbl.Bbl_Man_t
type Bbl_Obj_t = bbl.Bbl_Obj_t

func NetworkFromBBLIF(manager *bbl.Bbl_Man_t) (*abc.Network, error) {
	if manager == nil {
		return nil, fmt.Errorf("nil BBLIF manager")
	}
	if err := manager.Check(); err != nil {
		return nil, err
	}
	n := abc.NewNetwork(manager.Name)
	n.Type = uint8(abc.ABC_NTK_LOGIC)
	n.Function = uint8(abc.ABC_FUNC_SOP)
	objects := map[*bbl.Bbl_Obj_t]*abc.Object{}
	pi, po := 0, 0
	for _, object := range manager.Objects {
		var kind abc.ObjectKind
		switch object.Type {
		case bbl.BBL_OBJ_CI:
			kind = abc.ObjectPI
		case bbl.BBL_OBJ_CO:
			kind = abc.ObjectPO
		case bbl.BBL_OBJ_NODE:
			kind = abc.ObjectNode
		default:
			return nil, fmt.Errorf("invalid BBLIF object kind")
		}
		o := n.NewObject(kind)
		if kind == abc.ObjectPI {
			o.SetName(fmt.Sprintf("pi%d", pi))
			pi++
		} else if kind == abc.ObjectPO {
			o.SetName(fmt.Sprintf("po%d", po))
			po++
		} else {
			o.Data = object.SOP
		}
		objects[object] = o
	}
	for _, object := range manager.Objects {
		for _, fanin := range object.Fanins {
			if err := n.AddFanin(objects[object], objects[fanin].Signal()); err != nil {
				return nil, err
			}
		}
	}
	return n, n.Check()
}
func Bbl_ManToAbc(manager *bbl.Bbl_Man_t) *abc.Network {
	n, err := NetworkFromBBLIF(manager)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	return n
}
func BBLIFToAIG(manager *bbl.Bbl_Man_t) (*abc.Network, error) {
	if manager == nil {
		return nil, fmt.Errorf("nil BBLIF manager")
	}
	if err := manager.Check(); err != nil {
		return nil, err
	}
	graph := aig.NewGraph()
	graph.Name = manager.Name
	signals := map[*bbl.Bbl_Obj_t]aig.Edge{}
	active := map[*bbl.Bbl_Obj_t]bool{}
	for _, object := range manager.Objects {
		if object.Type == bbl.BBL_OBJ_CI {
			signals[object] = graph.NewInput()
		}
	}
	var derive func(*bbl.Bbl_Obj_t) (aig.Edge, error)
	derive = func(object *bbl.Bbl_Obj_t) (aig.Edge, error) {
		if edge, ok := signals[object]; ok {
			return edge, nil
		}
		if active[object] {
			return aig.InvalidEdge, fmt.Errorf("cyclic BBLIF network")
		}
		if object.Type != bbl.BBL_OBJ_NODE {
			return aig.InvalidEdge, fmt.Errorf("invalid BBLIF fanin")
		}
		active[object] = true
		inputs := make([]aig.Edge, len(object.Fanins))
		for i, fanin := range object.Fanins {
			var err error
			inputs[i], err = derive(fanin)
			if err != nil {
				return aig.InvalidEdge, err
			}
		}
		result := abc.Abc_NtkDeriveFlatGiaSop(graph, inputs, object.SOP)
		active[object] = false
		signals[object] = result
		return result, nil
	}
	for _, object := range manager.Objects {
		if object.Type == bbl.BBL_OBJ_NODE {
			if _, err := derive(object); err != nil {
				return nil, err
			}
		}
	}
	for _, object := range manager.Objects {
		if object.Type == bbl.BBL_OBJ_CO {
			edge, err := derive(object.Fanins[0])
			if err != nil {
				return nil, err
			}
			graph.AddOutput(edge)
		}
	}
	model, err := graph.AigerModel()
	if err != nil {
		return nil, err
	}
	return NetworkFromAigerModel(model)
}
func Bbl_ManToAig(manager *bbl.Bbl_Man_t) *abc.Network {
	n, err := BBLIFToAIG(manager)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	return n
}
func ReadBBLIF(filename string, check bool) (*abc.Network, error) {
	manager, err := bbl.Bbl_ManReadBinaryBlif(filename)
	if err != nil {
		return nil, err
	}
	defer bbl.Bbl_ManStop(manager)
	n, err := BBLIFToAIG(manager)
	if err == nil {
		n.Spec = filename
	}
	return n, err
}
func Io_ReadBblif(filename string, check int32) *abc.Network {
	n, err := ReadBBLIF(filename, check != 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	return n
}
func init() { RegisterFormat(IO_FILE_BBLIF, ReadBBLIF, nil) }
