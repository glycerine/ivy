package bmc

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
	"github.com/glycerine/abcgo/sat/satoko"
)

// The original C headers composed these APIs in one translation unit. Keep
// their names local to BMC while binding them to the native Go package types.
type Aig_Man_t = aig.Aig_Man_t
type Aig_Obj_t = aig.Aig_Obj_t
type Vec_Int_t = vec.Vec_Int_t
type Vec_Ptr_t = vec.Vec_Ptr_t
type Vec_Str_t = vec.Vec_Str_t
type Vec_Wrd_t = vec.Vec_Wrd_t
type Vec_Flt_t = vec.Vec_Flt_t
type Vec_Vec_t = vec.Vec_Vec_t
type sat_solver = bsat.Solver
type satoko_t = satoko.Solver
type Cnf_Dat_t = cnf.Cnf_Dat_t
type Abc_Cex_t = util.Abc_Cex_t
type word = uint64
type abctime = int64

