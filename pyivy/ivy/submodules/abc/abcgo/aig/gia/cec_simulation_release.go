package gia

// Gia_ManCecSimFreeAll releases both object signatures and saved CI patterns.
func Gia_ManCecSimFreeAll(p *Gia_Man_t) {
	Gia_ManCecSimFree(p)
	p.VSimsPi = nil
}
