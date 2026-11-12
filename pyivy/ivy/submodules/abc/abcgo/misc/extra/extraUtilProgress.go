package extra

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync/atomic"
)

/**CFile****************************************************************

  FileName    [extraUtilProgress.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [extra]

  Synopsis    [Progress bar.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: extraUtilProgress.c,v 1.0 2003/02/01 00:00:00 alanmi Exp $]

***********************************************************************/

// Progress mode replaces the dependency on ABC's global command frame.
// Applications can update it when their frame/batch settings change.
var progressMode atomic.Uint32

func Extra_ProgressBarSetMode(enabled, batch bool) {
	var mode uint32
	if !enabled {
		mode |= 1
	}
	if batch {
		mode |= 2
	}
	progressMode.Store(mode)
}

type ProgressBarStruct struct {
	nItemsNext, nItemsTotal, posTotal, posCur int32
	pFile                                     io.Writer
	stopped                                   bool
}
type ProgressBar = ProgressBarStruct

func Extra_ProgressBarStart(file io.Writer, total int32) *ProgressBar {
	if progressMode.Load()&1 != 0 {
		return nil
	}
	if total < 0 {
		panic("negative progress total")
	}
	if file == nil {
		file = os.Stdout
	}
	p := &ProgressBar{nItemsTotal: total, posTotal: 78, posCur: 1, pFile: file}
	p.nItemsNext = int32(8 * int64(total) / 78)
	Extra_ProgressBarShow(p, "")
	return p
}
func Extra_ProgressBarUpdate(p *ProgressBar, current int32, message string) {
	if p == nil || p.stopped || current < p.nItemsNext {
		return
	}
	Extra_ProgressBarUpdate_int(p, current, message)
}
func Extra_ProgressBarUpdate_int(p *ProgressBar, current int32, message string) {
	if p == nil || p.stopped || current < p.nItemsNext {
		return
	}
	if current >= p.nItemsTotal {
		p.posCur = 78
		p.nItemsNext = 1<<31 - 1
	} else {
		p.posCur += 7
		p.nItemsNext = int32(int64(7+p.posCur) * int64(p.nItemsTotal) / int64(p.posTotal))
	}
	Extra_ProgressBarShow(p, message)
}
func Extra_ProgressBarShow(p *ProgressBar, message string) {
	if p == nil || p.stopped || progressMode.Load()&2 != 0 {
		return
	}
	var line strings.Builder
	i := 0
	if message != "" {
		line.WriteString(message)
		line.WriteByte(' ')
		i = len(message) + 1
	}
	if i < int(p.posCur) {
		line.WriteString(strings.Repeat("-", int(p.posCur)-i))
		i = int(p.posCur)
	}
	if i == int(p.posCur) {
		line.WriteByte('>')
	}
	i++
	if i <= int(p.posTotal) {
		line.WriteString(strings.Repeat(" ", int(p.posTotal)-i+1))
	}
	line.WriteByte('\r')
	fmt.Fprint(p.pFile, line.String())
}
func Extra_ProgressBarClean(p *ProgressBar) {
	if p == nil || p.stopped || progressMode.Load()&2 != 0 {
		return
	}
	fmt.Fprint(p.pFile, strings.Repeat(" ", int(p.posTotal)+1), "\r")
}
func Extra_ProgressBarStop(p *ProgressBar) {
	if p != nil && !p.stopped {
		Extra_ProgressBarClean(p)
		p.stopped = true
	}
}
