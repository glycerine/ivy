package gia

// Gia_ManGateClasses and Gia_ManObjClasses return borrowed abstraction maps.
func Gia_ManGateClasses(p *Gia_Man_t) *Vec_Int_t { return p.vGateClasses }
func Gia_ManObjClasses(p *Gia_Man_t) *Vec_Int_t  { return p.vObjClasses }

// Gia_ManSetObjClasses transfers the supplied frame-map vector to the manager.
func Gia_ManSetObjClasses(p *Gia_Man_t, classes *Vec_Int_t) { p.vObjClasses = classes }
