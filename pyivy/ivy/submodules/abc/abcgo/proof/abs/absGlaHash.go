package abs

// Saig_ManBmcHashKey uses the source's wrapping 32-bit weighted sum.
func Saig_ManBmcHashKey(key []int32) uint32 {
	if len(key) < 5 {
		panic("abs: short CNF hash key")
	}
	primes := [5]uint32{12582917, 25165843, 50331653, 100663319, 201326611}
	var sum uint32
	for i := range primes {
		sum += uint32(key[i]) * primes[i]
	}
	return sum
}

// Saig_ManBmcLookup returns the hash entry's mutable cached literal.
func Saig_ManBmcLookup(p *Ga2_Man_t, key []int32) *int32 {
	if p.nTable <= 0 || len(p.pTable) < int(p.nTable)*6 {
		panic("abs: invalid CNF hash table")
	}
	offset := int(Saig_ManBmcHashKey(key)%uint32(p.nTable)) * 6
	entry := p.pTable[offset : offset+6]
	match := true
	for i := 0; i < 5; i++ {
		if entry[i] != key[i] {
			match = false
			break
		}
	}
	if match {
		p.nHashHit++
	} else {
		if entry[0] == 0 {
			p.nHashMiss++
		} else {
			p.nHashOver++
		}
		copy(entry[:5], key[:5])
		entry[5] = 0
	}
	return &entry[5]
}
