package abci

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/aig/saig"
)

// These wrappers retain the C API names and return conventions used by the
// translated ABC sources while delegating graph operations to their owners.
func Aig_ManStop(p *aig.Graph) { aig.Aig_ManStop(p) }

func Gia_ManRegNum(p *Gia_Man_t) int32          { return gia.Gia_ManRegNum(p) }
func Gia_ManHasMapping(p *Gia_Man_t) int32      { return gia.Gia_ManHasMapping(p) }
func Gia_ManBufNum(p *Gia_Man_t) int32          { return gia.Gia_ManBufNum(p) }
func Gia_ManLutSizeMax(p *Gia_Man_t) int32      { return gia.Gia_ManLutSizeMax(p) }
func Gia_ManTransferTiming(dst, src *Gia_Man_t) { gia.Gia_ManTransferTiming(dst, src) }

func Gia_ManFromAig(p *aig.Graph) *Gia_Man_t       { return gia.Gia_ManFromAig(p) }
func Gia_ManFromAigSimple(p *aig.Graph) *Gia_Man_t { return gia.Gia_ManFromAigSimple(p) }
func Gia_ManToAigSimple(p *Gia_Man_t) *aig.Graph   { return gia.Gia_ManToAigSimple(p) }

// The Go AIGER reader reports I/O/format failures as errors; the translated C
// call sites use the original nil-on-failure pointer convention.
func Gia_AigerRead(name string, simple, skip, check int32) *Gia_Man_t {
	p, _ := gia.Gia_AigerRead(name, simple, skip, check)
	return p
}

// The C writer reports errors through its status/output path; the source call
// sites intentionally discard the Go writer's error as well.
func Gia_AigerWrite(p *Gia_Man_t, name string, symbols, compact, newline int32) {
	_ = gia.Gia_AigerWrite(p, name, symbols, compact, newline)
}

func Saig_ManPoNum(p *aig.Graph) int32 { return saig.Saig_ManPoNum(p) }
func Saig_ManVerifyCex(p *aig.Graph, c *Abc_Cex_t) int32 {
	return saig.Saig_ManVerifyCex(p, c)
}
