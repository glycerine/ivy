// Error reporting and design cleanup from verCore.c.
package ver

import (
	"bytes"
	"fmt"

	"github.com/glycerine/abcgo/base/abc"
)

func verSetError(p *Ver_Man_t, message string) {
	clear(p.sError[:])
	copy(p.sError[:], message)
}

func Ver_ParseFreeData(p *Ver_Man_t) {
	if p.pDesign != nil {
		abc.Abc_DesFree(p.pDesign, nil)
		p.pDesign = nil
	}
}

func Ver_ParsePrintErrorMessage(p *Ver_Man_t) {
	p.fError = 1
	message := p.sError[:]
	if i := bytes.IndexByte(message, 0); i >= 0 {
		message = message[:i]
	}
	if p.fTopLevel != 0 {
		fmt.Fprintf(p.Output, "%s: %s\n", p.pFileName, message)
	} else {
		fmt.Fprintf(p.Output, "%s (line %d): %s\n", p.pFileName, Ver_StreamGetLineNumber(p.pReader), message)
	}
	Ver_ParseFreeData(p)
}
