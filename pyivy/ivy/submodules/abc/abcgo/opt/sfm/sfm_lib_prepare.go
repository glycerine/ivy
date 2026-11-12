package sfm

import (
	"fmt"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau"
	"time"
)

func Sfm_LibPrepare(nVars, two, delay, verbose, libVerbose int32) *Sfm_Lib_t {
	cells := mio.Mio_CollectRootsNewDefault2(min(6, nVars), nil, 0)
	return sfmLibPrepareCells(nVars, two, delay, verbose, libVerbose, cells)
}
func sfmLibPrepareCells(nVars, two, delay, verbose, libVerbose int32, cells []mio.Mio_Cell2_t) *Sfm_Lib_t {
	start := time.Now()
	p := Sfm_LibStart(nVars, delay, libVerbose)
	p.pCells = cells
	for i := 4; i < len(cells); i++ {
		c := &cells[i]
		c.Type = 0
		truth := c.Truth()
		if util.Abc_Tt6IsAndType(truth, c.Fanins()) != 0 || util.Abc_Tt6IsOrType(truth, c.Fanins()) != 0 {
			c.Type = 1
		} else {
			var text []byte
			if dau.Dau_DsdDecompose([]uint64{truth}, c.Fanins(), 0, 0, &text) <= 3 {
				c.Type = 2
			} else if libVerbose != 0 {
				fmt.Printf("Skipping gate %q with non-DSD function %s\n", c.Name(), text)
			}
		}
	}
	enumerate := func(base []uint64, n int32, bottom, top *mio.Mio_Cell2_t, pin int32) {
		current := append([]uint64(nil), base...)
		permutation := make([]int32, n)
		for i := range permutation {
			permutation[i] = int32(i)
		}
		schedule := extra.Extra_PermSchedule(n)
		for _, swap := range schedule {
			Sfm_LibPrepareAdd(p, current, permutation, bottom, top, pin)
			if top != nil && n > 5 {
				break
			}
			util.Abc_TtSwapAdjacent(current, p.nWords, swap)
			permutation[swap], permutation[swap+1] = permutation[swap+1], permutation[swap]
		}
	}
	for i := 4; i < len(cells); i++ {
		c := &cells[i]
		n := c.Fanins()
		truth := make([]uint64, p.nWords)
		for j := range truth {
			truth[j] = c.Truth()
		}
		enumerate(truth, n, c, nil, -1)
	}
	if two != 0 {
		for i := 4; i < len(cells); i++ {
			bottom := &cells[i]
			if bottom.Type == 0 {
				continue
			}
			for j := 4; j < len(cells); j++ {
				top := &cells[j]
				n := bottom.Fanins() + top.Fanins() - 1
				if top.Type == 0 || n > nVars {
					continue
				}
				for pin := int32(0); pin < top.Fanins(); pin++ {
					truth := make([]uint64, max(4, p.nWords))
					if nVars > 6 {
						Sfm_LibTruth8Two(bottom, top, pin, truth)
					} else {
						truth[0] = Sfm_LibTruth6Two(bottom, top, pin)
					}
					enumerate(truth[:p.nWords], n, bottom, top, pin)
				}
			}
		}
	}
	if verbose != 0 {
		fmt.Printf("Library processing: Var = %d. Cell = %d. Fun = %d. Obj = %d. Skip = %d. Rem = %d. Time = %.2f sec\n", nVars, len(cells), vec.Vec_MemEntryNum(p.vTtMem)-2, len(p.pObjs)-int(p.nObjRemoved), p.nObjSkipped, p.nObjRemoved, time.Since(start).Seconds())
	}
	return p
}
