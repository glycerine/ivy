package wln

import (
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Wln_NtkSetMetadata(p *Wln_Ntk_t, spec string, smt bool) {
	p.pSpec = spec
	p.fSmtLib = wlnBool(smt)
}
func Wln_NtkSetNameMan(p *Wln_Ntk_t, names *util.Abc_Nam_t) { p.pManName = names }
func Wln_NtkNameMan(p *Wln_Ntk_t) *util.Abc_Nam_t           { return p.pManName }
func Wln_NtkRangeMan(p *Wln_Ntk_t) *vec.Hash_IntMan_t       { return p.pRanges }
func Wln_NtkFfs(p *Wln_Ntk_t) *vec.Vec_Int_t                { return &p.vFfs }
