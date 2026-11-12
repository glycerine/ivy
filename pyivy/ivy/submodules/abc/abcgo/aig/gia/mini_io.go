package gia

import (
	"github.com/glycerine/abcgo/aig/miniaig"
	"github.com/glycerine/abcgo/misc/extra"
)

func Gia_ManReadMiniAig(name string) *Gia_Man_t {
	p := miniaig.Mini_AigLoad(name)
	if p == nil {
		return nil
	}
	g := Gia_ManFromMiniAig(p, nil)
	g.pName = extra.Extra_FileNameGeneric(name)
	return g
}
func Gia_ManWriteMiniAig(g *Gia_Man_t, name string) { miniaig.Mini_AigDump(Gia_ManToMiniAig(g), name) }
func Gia_ManReadMiniLut(name string) *Gia_Man_t {
	p := miniaig.Mini_LutLoad(name)
	if p == nil {
		return nil
	}
	g := Gia_ManFromMiniLut(p, nil)
	g.pName = extra.Extra_FileNameGeneric(name)
	return g
}
func Gia_ManWriteMiniLut(g *Gia_Man_t, name string) { miniaig.Mini_LutDump(Gia_ManToMiniLut(g), name) }
