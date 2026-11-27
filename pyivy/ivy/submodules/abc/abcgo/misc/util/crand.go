package util

import "sync"

// C's rand()/srand(): one process-wide glibc TYPE_3 stream, as ABC sees it.
// Ports whose C original calls rand() share this state, so a sequence that
// spans several ABC routines (for example the srand(0xABC) of
// Sim_ComputeFunSupp followed by saucy's SIM_RANDOM_UNSIGNED draws) matches
// the native one. The stream starts as if seeded with 1, as glibc's does.
var cRand = struct {
	sync.Mutex
	state  [34]int32
	index  int
	seeded bool
}{}

func cSrandLocked(seed uint32) {
	if seed == 0 {
		seed = 1
	}
	r := &cRand
	var s [31]int32
	word := int64(int32(seed))
	if word < 0 {
		word = -word
	}
	s[0] = int32(word)
	for i := 1; i < 31; i++ {
		hi, lo := word/127773, word%127773
		word = 16807*lo - 2836*hi
		if word < 0 {
			word += 2147483647
		}
		s[i] = int32(word)
	}
	copy(r.state[:31], s[:])
	r.index = 0
	r.seeded = true
	for i := 0; i < 310; i++ {
		cRandNextLocked()
	}
}

func cRandNextLocked() uint32 {
	r := &cRand
	i := r.index % 31
	front := (i + 3) % 31
	value := uint32(r.state[front]) + uint32(r.state[i])
	r.state[front] = int32(value)
	r.index++
	return value >> 1
}

// C_Srand is srand(seed).
func C_Srand(seed uint32) {
	cRand.Lock()
	defer cRand.Unlock()
	cSrandLocked(seed)
}

// C_Rand is rand(): a value in [0, 2^31).
func C_Rand() uint32 {
	cRand.Lock()
	defer cRand.Unlock()
	if !cRand.seeded {
		cSrandLocked(1)
	}
	return cRandNextLocked()
}

// Sim_RandomUnsigned is ABC's SIM_RANDOM_UNSIGNED (sim.h, also
// Ivy_ObjRandomSim): three draws of rand() spread over 32 bits, evaluated
// left to right as gcc does.
func Sim_RandomUnsigned() uint32 {
	return (C_Rand() << 24) ^ (C_Rand() << 12) ^ C_Rand()
}

// C_RAND_MAX is glibc's RAND_MAX.
const C_RAND_MAX = 2147483647
