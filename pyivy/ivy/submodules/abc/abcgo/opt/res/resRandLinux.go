package res

import "sync"

// resCRand mirrors Linux/glibc srand/rand (TYPE_3, degree 31). ABC seeds
// libc's process-global rand in Res_SimAdjust; this shared stream models calls
// made by RES, but cannot track libc rand calls from external packages.
// Other C runtimes use different rand algorithms and need their own port.
type resCRand struct {
	mu          sync.Mutex
	state       [31]uint32
	front, rear int
}

func (r *resCRand) seed(seed uint32) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if seed == 0 {
		seed = 1
	}
	r.front, r.rear = 3, 0
	r.state[0] = seed
	value := int64(int32(seed))
	for i := 1; i < len(r.state); i++ {
		value = (16807 * value) % 2147483647
		if value < 0 {
			value += 2147483647
		}
		r.state[i] = uint32(value)
	}
	for i := 0; i < 310; i++ {
		r.nextUnlocked()
	}
}
func (r *resCRand) nextUnlocked() uint32 {
	r.state[r.front] += r.state[r.rear]
	value := r.state[r.front] >> 1
	r.front = (r.front + 1) % len(r.state)
	r.rear = (r.rear + 1) % len(r.state)
	return value
}
func (r *resCRand) next() uint32 { r.mu.Lock(); defer r.mu.Unlock(); return r.nextUnlocked() }

var resGlobalRand resCRand

// Abc_InfoRandom uses the exact 3-rand expression in abc.h on Linux/glibc.
func Abc_InfoRandom(p []uint32, nWords int32) {
	for i := nWords - 1; i >= 0; i-- {
		p[i] = (resGlobalRand.next() << 24) ^ (resGlobalRand.next() << 12) ^ resGlobalRand.next()
	}
}
