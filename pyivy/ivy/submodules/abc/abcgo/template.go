package abcgo

// Xyz_ParTry_t is the example operation's parameter record from src/template.c.
type Xyz_ParTry_t struct{ Par int32 }

// Xyz_ManTry_t borrows the caller's parameters and AIG without modifying either.
// The source template never accesses the AIG. Keeping this unused borrowed handle
// opaque avoids coupling the example to a particular AIG representation.
type Xyz_ManTry_t struct {
	pPar *Xyz_ParTry_t
	pAig any
}

func Xyz_ManTryAlloc(aig any, parameters *Xyz_ParTry_t) *Xyz_ManTry_t {
	return &Xyz_ManTry_t{pPar: parameters, pAig: aig}
}

// Xyz_ManTryFree releases the example manager to Go's garbage collector. Borrowed
// parameters and AIG remain owned by the caller, matching the C free operation.
func Xyz_ManTryFree(_ *Xyz_ManTry_t) {}

// Xyz_ManPerform preserves the original template's successful no-op operation.
func Xyz_ManPerform(aig any, parameters *Xyz_ParTry_t) int32 {
	manager := Xyz_ManTryAlloc(aig, parameters)
	defer Xyz_ManTryFree(manager)
	return 1
}
