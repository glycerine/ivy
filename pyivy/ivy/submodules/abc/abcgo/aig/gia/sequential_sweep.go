package gia

import "fmt"

// giaTernaryStates overapproximates reachable states from the all-zero reset.
// Primary inputs are unknown; after 300 rounds changing registers are retired
// to unknown periodically, matching the source convergence strategy.
func giaTernaryStates(p *Gia_Man_t) [][]byte {
	const unknown byte = 2
	if Gia_ManRegNum(p) <= 0 {
		panic("gia: ternary simulation requires registers")
	}
	state := make([]byte, Gia_ManRegNum(p))
	states := [][]byte{append([]byte(nil), state...)}
	seen := map[string]int{string(state): 0}
	var previous []byte
	values := make([]byte, p.nObjs)
	for iteration := 0; ; iteration++ {
		values[0] = 0
		for i := int32(0); i < Gia_ManPiNum(p); i++ {
			values[Gia_ManPi(p, i).id] = unknown
		}
		for i, v := range state {
			values[Gia_ManRo(p, int32(i)).id] = v
		}
		read := func(lit int32) byte {
			v := values[lit>>1]
			if v != unknown && lit&1 != 0 {
				v ^= 1
			}
			return v
		}
		for id := int32(1); id < p.nObjs; id++ {
			a, b, c, kind := p.Gate(id)
			switch kind {
			case 1:
				continue
			case 2:
				x, y := read(a), read(b)
				if x == 0 || y == 0 {
					values[id] = 0
				} else if x == 1 && y == 1 {
					values[id] = 1
				} else {
					values[id] = unknown
				}
			case 3:
				x, y := read(a), read(b)
				if x == unknown || y == unknown {
					values[id] = unknown
				} else {
					values[id] = x ^ y
				}
			case 4:
				x, y, z := read(a), read(b), read(c)
				if x == 1 {
					values[id] = y
				} else if x == 0 {
					values[id] = z
				} else if y == z {
					values[id] = y
				} else {
					values[id] = unknown
				}
			case 5, 6:
				values[id] = read(a)
			default:
				panic("gia: invalid ternary simulation gate")
			}
		}
		next := make([]byte, len(state))
		for i := range next {
			next[i] = values[Gia_ManRi(p, int32(i)).id]
		}
		states = append(states, append([]byte(nil), next...))
		if index, ok := seen[string(next)]; ok {
			p.nTerStates = int32(len(states))
			p.nTerLoop = int32(index)
			return states
		}
		seen[string(next)] = len(states) - 1
		state = append(state[:0], next...)
		if iteration >= 300 && iteration%10 == 0 && previous != nil {
			for i := range state {
				if next[i] != previous[i] {
					state[i] = unknown
				}
			}
		}
		previous = next
	}
}
func Gia_ManReduceConst(p *Gia_Man_t, verbose int32) *Gia_Man_t {
	// The compact frontier simulator implements ordinary AND/inverter AIGs.
	// Keep the general evaluator for graphs with explicit XOR, MUX or buffers.
	plain := p.pMuxes == nil
	for _, obj := range p.pObjs {
		if Gia_ObjIsXor(obj) != 0 || Gia_ObjIsBuf(obj) != 0 {
			plain = false
			break
		}
	}
	if plain {
		manager := Gia_ManTerSimulate(p, verbose)
		mapping := Gia_ManTerCreateMap(manager, verbose)
		Gia_ManTerDelete(manager)
		return Gia_ManDupDfsCiMap(p, mapping, nil)
	}
	states := giaTernaryStates(p)
	mapping := make([]int32, Gia_ManCiNum(p))
	for i := range mapping {
		mapping[i] = -1
	}
	signatures := map[string]int32{}
	constants, equivalent := 0, 0
	for i := int32(0); i < Gia_ManRegNum(p); i++ {
		trace := make([]byte, len(states))
		allZero, noUnknown := true, true
		for j, state := range states {
			trace[j] = state[i]
			if state[i] != 0 {
				allZero = false
			}
			if state[i] == 2 {
				noUnknown = false
			}
		}
		if allZero {
			mapping[Gia_ManPiNum(p)+i] = 0
			constants++
			continue
		}
		if noUnknown {
			key := string(trace)
			if repr, ok := signatures[key]; ok {
				mapping[Gia_ManPiNum(p)+i] = Gia_ManRo(p, repr).id << 1
				equivalent++
			} else {
				signatures[key] = i
			}
		}
	}
	if verbose != 0 {
		fmt.Printf("Transforming %d const and %d equiv registers.\n", constants, equivalent)
	}
	return Gia_ManDupDfsCiMap(p, mapping, nil)
}
func Gia_ManReduceEquiv(p *Gia_Man_t, verbose int32) *Gia_Man_t {
	mapping := make([]int32, Gia_ManCiNum(p))
	for i := range mapping {
		mapping[i] = -1
	}
	drivers := map[int32]int32{}
	changes := 0
	for i := int32(0); i < Gia_ManRegNum(p); i++ {
		ri := Gia_ManRi(p, i)
		driver := Gia_ObjFaninLit0(ri, ri.id)
		if driver == 0 {
			mapping[Gia_ManPiNum(p)+i] = 0
			changes++
			continue
		}
		if earlier, ok := drivers[driver]; ok {
			mapping[Gia_ManPiNum(p)+i] = Gia_ManRo(p, earlier).id << 1
			changes++
		} else {
			drivers[driver] = i
		}
	}
	if changes == 0 {
		return p
	}
	return Gia_ManDupDfsCiMap(p, mapping, nil)
}
func Gia_ManSeqStructSweep(src *Gia_Man_t, constant, equivalent, verbose int32) *Gia_Man_t {
	if Gia_ManRegNum(src) == 0 {
		return Gia_ManCleanup(src)
	}
	if verbose != 0 {
		fmt.Println("Performing sequential cleanup.")
	}
	current := Gia_ManSeqCleanup(src)
	if constant != 0 && Gia_ManRegNum(current) > 0 {
		next := Gia_ManReduceConst(current, verbose)
		Gia_ManStop(current)
		current = next
	}
	if equivalent != 0 {
		for {
			next := Gia_ManSeqCleanup(current)
			Gia_ManStop(current)
			current = next
			if Gia_ManRegNum(current) == 0 {
				break
			}
			next = Gia_ManReduceEquiv(current, verbose)
			if next == current {
				break
			}
			Gia_ManStop(current)
			current = next
		}
	}
	return current
}
