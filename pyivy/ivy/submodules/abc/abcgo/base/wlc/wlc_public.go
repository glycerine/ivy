package wlc

import "github.com/glycerine/abcgo/misc/util"

type Abc_Nam_t = util.Abc_Nam_t

// Public metadata accessors for clients that need to inspect a word-level
// network without depending on the generated C-layout fields.
func Wlc_NtkName(p *Wlc_Ntk_t) string             { return p.pName }
func Wlc_NtkSpec(p *Wlc_Ntk_t) string             { return p.pSpec }
func Wlc_NtkSetSpec(p *Wlc_Ntk_t, spec string)    { p.pSpec = spec }
func Wlc_NtkSmtLib(p *Wlc_Ntk_t) bool             { return p.fSmtLib != 0 }
func Wlc_NtkInits(p *Wlc_Ntk_t) string            { return p.pInits }
func Wlc_NtkNameMan(p *Wlc_Ntk_t) *util.Abc_Nam_t { return p.pManName }
func Wlc_ObjSigned(p *Wlc_Obj_t) bool             { return p.Signed != 0 }
func Wlc_ObjEnd(p *Wlc_Obj_t) int32               { return p.End }
func Wlc_ObjBeg(p *Wlc_Obj_t) int32               { return p.Beg }
