package io

import (
	"fmt"
	"os"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/bbl"
)

func BBLIFfromNetwork(n *abc.Network) (*bbl.Bbl_Man_t, error) {
	if n == nil {
		return nil, fmt.Errorf("nil network")
	}
	if n.Function != uint8(abc.ABC_FUNC_SOP) {
		return nil, fmt.Errorf("BBLIF requires SOP functions")
	}
	manager := bbl.Bbl_ManStart(n.Name)
	copies := map[*abc.Object]int{}
	for _, input := range n.Inputs() {
		if err := bbl.Bbl_ManCreateObject(manager, bbl.BBL_OBJ_CI, input.ID, 0, ""); err != nil {
			return nil, err
		}
		copies[input] = input.ID
	}
	active := map[*abc.Object]bool{}
	var derive func(*abc.Object) (int, error)
	derive = func(object *abc.Object) (int, error) {
		if id, ok := copies[object]; ok {
			return id, nil
		}
		if active[object] {
			return 0, fmt.Errorf("cyclic ABC netlist")
		}
		active[object] = true
		defer delete(active, object)
		if object.Kind == abc.ObjectNet {
			fans := object.Fanins()
			if len(fans) != 1 || fans[0].Complement {
				return 0, fmt.Errorf("invalid net driver")
			}
			id, err := derive(fans[0].Object)
			if err == nil {
				copies[object] = id
			}
			return id, err
		}
		if object.Kind != abc.ObjectNode {
			return 0, fmt.Errorf("unsupported BBLIF object %d", object.ID)
		}
		sop, ok := object.Data.(string)
		if !ok {
			return 0, fmt.Errorf("node has no SOP")
		}
		fans := object.Fanins()
		ids := make([]int, len(fans))
		for i, fanin := range fans {
			if fanin.Complement {
				return 0, fmt.Errorf("BBLIF requires explicit SOP inverter nodes")
			}
			var err error
			ids[i], err = derive(fanin.Object)
			if err != nil {
				return 0, err
			}
		}
		if err := bbl.Bbl_ManCreateObject(manager, bbl.BBL_OBJ_NODE, object.ID, len(ids), sop); err != nil {
			return 0, err
		}
		for _, id := range ids {
			if err := bbl.Bbl_ManAddFanin(manager, object.ID, id); err != nil {
				return 0, err
			}
		}
		copies[object] = object.ID
		return object.ID, nil
	}
	for _, output := range n.Outputs() {
		fans := output.Fanins()
		if len(fans) != 1 || fans[0].Complement {
			return nil, fmt.Errorf("invalid BBLIF output")
		}
		id, err := derive(fans[0].Object)
		if err != nil {
			return nil, err
		}
		if err = bbl.Bbl_ManCreateObject(manager, bbl.BBL_OBJ_CO, output.ID, 1, ""); err != nil {
			return nil, err
		}
		if err = bbl.Bbl_ManAddFanin(manager, output.ID, id); err != nil {
			return nil, err
		}
	}
	if err := manager.Check(); err != nil {
		return nil, err
	}
	return manager, nil
}
func Bbl_ManFromAbc(n *abc.Network) *bbl.Bbl_Man_t {
	manager, err := BBLIFfromNetwork(n)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	return manager
}
func WriteBBLIF(n *abc.Network, filename string) error {
	manager, err := BBLIFfromNetwork(n)
	if err != nil {
		return err
	}
	defer bbl.Bbl_ManStop(manager)
	return bbl.Bbl_ManDumpBinaryBlif(manager, filename)
}
func Io_WriteBblif(n *abc.Network, filename string) error { return WriteBBLIF(n, filename) }
func init()                                               { RegisterFormat(IO_FILE_BBLIF, nil, WriteBBLIF) }
