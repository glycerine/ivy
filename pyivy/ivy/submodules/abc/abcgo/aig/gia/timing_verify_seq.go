package gia

import "time"

// giaVerifySequentialMiter explores zero-initialized states exactly. It returns
// undecided when its state cap or deadline is reached; it never infers a proof
// from an incomplete traversal.
func giaVerifySequentialMiter(miter *Gia_Man_t, maxStates int, deadline time.Time) int32 {
	inputs, regs, outputs := int(Gia_ManPiNum(miter)), int(Gia_ManRegNum(miter)), int(Gia_ManPoNum(miter))
	if regs <= 0 || inputs < 0 || outputs < 0 || maxStates <= 0 {
		return -1
	}
	// Bound explicit input enumeration as well as the number of stored states.
	const maxTransitions = 1 << 24
	transitions := 0
	stateBytes := (regs + 7) / 8
	queue := [][]byte{make([]byte, stateBytes)}
	seen := map[string]struct{}{string(queue[0]): {}}
	for head := 0; head < len(queue); head++ {
		state := queue[head]
		assignment := make([]bool, inputs+regs)
		for i := 0; i < regs; i++ {
			assignment[inputs+i] = state[i>>3]&(1<<uint(i&7)) != 0
		}
		for {
			if !deadline.IsZero() && time.Now().After(deadline) {
				return -1
			}
			if transitions >= maxTransitions {
				return -1
			}
			transitions++
			out, err := miter.Evaluate(assignment)
			if err != nil || len(out) != outputs+regs {
				return -1
			}
			for i := 0; i < outputs; i++ {
				if out[i] {
					return 0
				}
			}
			next := make([]byte, stateBytes)
			for i := 0; i < regs; i++ {
				if out[outputs+i] {
					next[i>>3] |= 1 << uint(i&7)
				}
			}
			key := string(next)
			if _, ok := seen[key]; !ok {
				if len(seen) >= maxStates {
					return -1
				}
				seen[key] = struct{}{}
				queue = append(queue, next)
			}
			i := 0
			for i < inputs && assignment[i] {
				assignment[i] = false
				i++
			}
			if i == inputs {
				break
			}
			assignment[i] = true
		}
	}
	return 1
}
