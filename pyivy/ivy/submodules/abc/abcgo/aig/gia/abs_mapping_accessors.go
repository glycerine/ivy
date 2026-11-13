package gia

// Gia_ManMapping returns the borrowed mapping vector used by abstraction cut
// records. Callers must keep the manager alive and honor the record layout.
func Gia_ManMapping(p *Gia_Man_t) *Vec_Int_t { return p.vMapping }
