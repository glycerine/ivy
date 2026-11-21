package core

import "sync"

// CUDD's public generator and all reordering heuristics draw from one stream.
var cuddRandomState struct {
	sync.Mutex
	first, second, selector int64
	table                   [64]int64
}

func Cudd_Srandom(seed int64) {
	cuddRandomState.Lock()
	defer cuddRandomState.Unlock()
	seedCuddRandom(seed)
}

func seedCuddRandom(seed int64) {
	s := &cuddRandomState
	if seed < 0 {
		seed = -seed
	}
	if seed == 0 {
		seed = 1
	}
	s.first, s.second = seed, seed
	for i := range len(s.table) + 11 {
		s.first = cuddFirstRandomStep(s.first)
		s.table[i%len(s.table)] = s.first
	}
	s.selector = s.table[1]
}

func cuddFirstRandomStep(value int64) int64 {
	quotient := value / 53668
	value = 40014*(value-quotient*53668) - quotient*12211
	if value < 0 {
		value += 2147483563
	}
	return value
}

func Cudd_Random() int64 {
	s := &cuddRandomState
	s.Lock()
	defer s.Unlock()
	if s.first == 0 {
		seedCuddRandom(1)
	}
	s.first = cuddFirstRandomStep(s.first)
	quotient := s.second / 52774
	s.second = 40692*(s.second-quotient*52774) - quotient*3791
	if s.second < 0 {
		s.second += 2147483399
	}
	i := s.selector / 33554431
	s.selector = s.table[i] - s.second
	s.table[i] = s.first
	if s.selector < 1 {
		s.selector += 2147483562
	}
	return s.selector - 1
}
