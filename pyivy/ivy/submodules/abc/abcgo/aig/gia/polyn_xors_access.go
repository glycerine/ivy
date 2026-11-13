package gia

import "github.com/glycerine/abcgo/misc/vec"

// PolynXors returns the temporary top-level XOR collection, if any.
func (p *Gia_Man_t) PolynXors() *vec.Vec_Int_t { return p.vXors }

// SetPolynXors replaces the temporary top-level XOR collection.
func (p *Gia_Man_t) SetPolynXors(xors *vec.Vec_Int_t) { p.vXors = xors }
