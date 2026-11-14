package gia

import "github.com/glycerine/abcgo/misc/vec"

// Gia_ManSetMapping replaces the manager's abstraction cut records.
func Gia_ManSetMapping(p *Gia_Man_t, mapping *vec.Vec_Int_t) { p.VMapping = mapping }
