// SOP/LUT balancing and pin delays from ifDelay.c.
package if_

import "github.com/glycerine/abcgo/misc/vec"

const IF_MAX_CUBES = 70

var ifSopGateDelays = [20]float64{1, 1, 2, 2.58, 3, 3.32, 3.58, 3.81, 4, 4.17, 4.32, 4.46, 4.58, 4.70, 4.81, 4.91, 5, 5.09, 5.17, 5.25}

func If_CutMaxCubeSize(cover *Vec_Int_t, n int32) int32 {
	if n < 0 || n > 16 {
		panic("SOP cube support outside 0..16")
	}
	var maximum int32
	for _, entry := range vec.Vec_IntArray(cover) {
		var literals int32
		for pin := int32(0); pin < n; pin++ {
			literal := uint32(entry) >> uint(pin*2) & 3
			if literal == 1 || literal == 2 {
				literals++
			}
		}
		maximum = max(maximum, literals)
	}
	return maximum
}
func ifDelayCover(p *If_Man_t, c *If_Cut_t) (*Vec_Int_t, bool) {
	if p == nil || c == nil || int(c.nLeaves) >= len(p.vTtIsops) || p.vTtIsops[c.nLeaves] == nil {
		return nil, false
	}
	id := int(If_CutTruthLit(c) >> 1)
	levels := p.vTtIsops[c.nLeaves]
	if id < 0 || id >= vec.Vec_WecSize(levels) {
		return nil, false
	}
	phase := false
	cacheIndex := c.nLeaves
	if cacheIndex < 6 && p.vTtIsops[cacheIndex] == p.vTtIsops[6] {
		cacheIndex = 6
	}
	if id < len(p.vTtIsopComplement[cacheIndex]) {
		phase = p.vTtIsopComplement[cacheIndex][id]
	}
	return vec.Vec_WecEntry(levels, id), phase
}
func ifDelayTimes(p *If_Man_t, c *If_Cut_t) []int32 {
	times := make([]int32, c.nLeaves)
	for pin := range times {
		times[pin] = int32(If_ObjCutBest(If_CutLeaf(p, c, int32(pin))).Delay)
	}
	return times
}
func ifDelayPermutation(c *If_Cut_t) []byte {
	if len(c.permutation) < int(c.nLeaves) {
		c.permutation = make([]byte, c.nLeaves)
	}
	return c.permutation[:c.nLeaves]
}
func If_CutDelaySop(p *If_Man_t, c *If_Cut_t) int32 {
	c.fUser = 1
	if c.nLeaves == 0 {
		return 0
	}
	times := ifDelayTimes(p, c)
	if c.nLeaves == 1 {
		return times[0]
	}
	cover, _ := ifDelayCover(p, c)
	if cover == nil || vec.Vec_IntSize(cover) == 0 || p.pPars == nil || int32(vec.Vec_IntSize(cover)) > p.pPars.NGateSize {
		return -1
	}
	maxCube := If_CutMaxCubeSize(cover, int32(c.nLeaves))
	delay := ifSopGateDelays[c.nLeaves]
	cost := vec.Vec_IntSize(cover)
	if cost > 1 {
		delay += ifSopGateDelays[maxCube]
		cost++
	}
	gateDelay := int32(delay + .5)
	permutation := ifDelayPermutation(c)
	maximum := int32(0)
	for pin, time := range times {
		permutation[pin] = byte(gateDelay)
		maximum = max(maximum, time+gateDelay)
	}
	c.Cost = uint32(cost)
	return maximum
}
func ifValidSop(cover *Vec_Int_t, times []int32, n int32) bool {
	if cover == nil || n < 1 || n > IF_MAX_FUNC_LUTSIZE || len(times) < int(n) || vec.Vec_IntSize(cover) == 0 || vec.Vec_IntSize(cover) > IF_MAX_CUBES {
		return false
	}
	for _, time := range times[:n] {
		if time < 0 {
			return false
		}
	}
	seen := make(map[int32]bool)
	var used uint32
	for _, entry := range vec.Vec_IntArray(cover) {
		if seen[entry] {
			return false
		}
		seen[entry] = true
		literalCount := 0
		for pin := int32(0); pin < n; pin++ {
			literal := uint32(entry) >> uint(2*pin) & 3
			if literal == 3 {
				return false
			}
			if literal != 0 {
				used |= 1 << uint(pin)
				literalCount++
			}
		}
		if literalCount == 0 {
			return false
		}
	}
	return used == (1<<uint(n))-1
}
func If_CutSopBalancePinDelaysInt(cover *Vec_Int_t, times []int32, faninProfiles []uint64, n int32, result *uint64) int32 {
	if !ifValidSop(cover, times, n) || len(faninProfiles) < int(n) || result == nil {
		return -1
	}
	andTimes, andProfiles := make([]int32, n), make([]uint64, n)
	orTimes, orProfiles := make([]int32, IF_MAX_CUBES), make([]uint64, IF_MAX_CUBES)
	var orCount, delay int32
	for _, entry := range vec.Vec_IntArray(cover) {
		var andCount int32
		for pin := int32(0); pin < n; pin++ {
			literal := uint32(entry) >> uint(pin*2) & 3
			if literal == 1 || literal == 2 {
				delay = If_LogCounterPinDelays(andTimes, &andCount, andProfiles, times[pin], faninProfiles[pin], n, 0)
			}
		}
		profile := If_LogPinDelaysMulti(andProfiles, andCount, n, 0)
		delay = If_LogCounterPinDelays(orTimes, &orCount, orProfiles, delay, profile, n, 0)
	}
	*result = If_LogPinDelaysMulti(orProfiles, orCount, n, 0)
	return delay
}
func If_CutSopBalancePinDelaysIntInt(cover *Vec_Int_t, times []int32, n int32, permutation []byte) int32 {
	if n < 0 || n > IF_MAX_FUNC_LUTSIZE || len(permutation) < int(n) {
		return -1
	}
	profiles := make([]uint64, n)
	for pin := range profiles {
		profiles[pin] = If_CutPinDelayInit(int32(pin))
	}
	var result uint64
	delay := If_CutSopBalancePinDelaysInt(cover, times, profiles, n, &result)
	if delay < 0 {
		return -1
	}
	translated := make([]byte, n)
	for pin := int32(0); pin < n; pin++ {
		depth := If_CutPinDelayGet(result, pin)
		if depth == 0 {
			return -1
		}
		translated[pin] = byte(depth - 1)
	}
	copy(permutation, translated)
	return delay
}
func If_CutSopBalancePinDelays(p *If_Man_t, c *If_Cut_t, permutation []byte) int32 {
	if c.nLeaves == 0 {
		return 0
	}
	if len(permutation) < int(c.nLeaves) {
		return -1
	}
	times := ifDelayTimes(p, c)
	if c.nLeaves == 1 {
		permutation[0] = 0
		return times[0]
	}
	cover, _ := ifDelayCover(p, c)
	return If_CutSopBalancePinDelaysIntInt(cover, times, int32(c.nLeaves), permutation)
}

// Optional coverComplement replaces the C attribute hidden inside Vec_Int.nCap.
func If_CutSopBalanceEvalInt(cover *Vec_Int_t, times, faninLiterals []int32, aig *Vec_Int_t, result *int32, n int32, area *int32, coverComplement ...bool) int32 {
	return ifCutSopBalanceEvalInt(cover, times, faninLiterals, aig, result, n, n, area, coverComplement...)
}

func ifCutSopBalanceEvalInt(cover *Vec_Int_t, times, faninLiterals []int32, aig *Vec_Int_t, result *int32, n, support int32, area *int32, coverComplement ...bool) int32 {
	if !ifValidSop(cover, times, n) || aig != nil && len(faninLiterals) < int(n) {
		return -1
	}
	andTimes, andLiterals := make([]int32, n), make([]int32, n)
	orTimes, orLiterals := make([]int32, IF_MAX_CUBES), make([]int32, IF_MAX_CUBES)
	var orCount, delay int32
	gates := int32(0)
	before := 0
	if aig != nil {
		before = vec.Vec_IntSize(aig)
	}
	for _, entry := range vec.Vec_IntArray(cover) {
		var andCount, literals int32
		for pin := int32(0); pin < n; pin++ {
			literal := uint32(entry) >> uint(pin*2) & 3
			if literal == 0 {
				continue
			}
			lit := int32(-1)
			if aig != nil {
				lit = faninLiterals[pin]
				if literal == 1 {
					lit ^= 1
				}
			}
			literals++
			delay = If_LogCounterAddAig(andTimes, &andCount, andLiterals, times[pin], lit, aig, support, 0, 0)
		}
		root := int32(0)
		if aig != nil {
			root = If_LogCreateAndXorMulti(aig, andLiterals, andCount, support, 0)
		} else {
			gates += literals - 1
		}
		orLit := int32(-1)
		if aig != nil {
			orLit = root ^ 1
		}
		delay = If_LogCounterAddAig(orTimes, &orCount, orLiterals, delay, orLit, aig, support, 0, 0)
	}
	if aig != nil {
		root := If_LogCreateAndXorMulti(aig, orLiterals, orCount, support, 0) ^ 1
		if len(coverComplement) > 0 && coverComplement[0] {
			root ^= 1
		}
		if result != nil {
			*result = root
		}
		gates = int32((vec.Vec_IntSize(aig) - before) / 2)
	} else {
		gates += int32(vec.Vec_IntSize(cover)) - 1
	}
	if area != nil {
		*area += gates
	}
	return delay
}
func If_CutSopBalanceEvalIntInt(cover *Vec_Int_t, leaves int32, times []int32, aig *Vec_Int_t, complement int32, area *int32, coverComplement ...bool) int32 {
	literals := make([]int32, max(leaves, 0))
	for pin := range literals {
		literals[pin] = int32(2 * pin)
	}
	root := int32(0)
	delay := If_CutSopBalanceEvalInt(cover, times, literals, aig, &root, leaves, area, coverComplement...)
	if delay < 0 {
		return -1
	}
	if aig != nil {
		if root>>1 != leaves+int32(vec.Vec_IntSize(aig)/2)-1 {
			panic("balanced AIG root is not the final node")
		}
		if leaves == 1 && vec.Vec_IntSize(aig) == 0 {
			vec.Vec_IntPush(aig, 0)
		}
		vec.Vec_IntPush(aig, root&1^complement)
	}
	return delay
}
func If_CutSopBalanceEval(p *If_Man_t, c *If_Cut_t, aig *Vec_Int_t) int32 {
	c.fUser = 1
	if aig != nil {
		vec.Vec_IntClear(aig)
	}
	if c.nLeaves == 0 {
		if aig != nil {
			vec.Vec_IntPush(aig, If_CutTruthLit(c)&1)
		}
		c.Cost = 0
		return 0
	}
	times := ifDelayTimes(p, c)
	if c.nLeaves == 1 {
		if aig != nil {
			vec.Vec_IntPush(aig, 0)
			vec.Vec_IntPush(aig, If_CutTruthLit(c)&1)
		}
		c.Cost = 0
		return times[0]
	}
	cover, coverPhase := ifDelayCover(p, c)
	var area int32
	delay := If_CutSopBalanceEvalIntInt(cover, int32(c.nLeaves), times, aig, If_CutTruthLit(c)&1^int32(c.fCompl), &area, coverPhase)
	if delay >= 0 {
		c.Cost = uint32(area)
	}
	return delay
}
func ifDelayLutSize(p *If_Man_t) int32 {
	if p == nil || p.pPars == nil || len(p.pPars.PLutStruct) == 0 {
		return 0
	}
	value := p.pPars.PLutStruct[0]
	if value < '2' || value > '9' {
		return 0
	}
	return int32(value - '0')
}
func ifDelayDsdPermutation(p *If_Man_t, c *If_Cut_t) []byte {
	permutation := If_CutDsdPerm(p, c)
	if len(permutation) < int(c.nLeaves) {
		panic("short DSD permutation")
	}
	seen := make(map[int32]bool)
	for _, literal := range permutation[:c.nLeaves] {
		variable := int32(literal) >> 1
		if variable >= int32(c.nLeaves) || seen[variable] {
			panic("invalid DSD permutation")
		}
		seen[variable] = true
	}
	return permutation[:c.nLeaves]
}
func If_CutLutBalancePinDelays(p *If_Man_t, c *If_Cut_t, permutation []byte) int32 {
	if c.nLeaves == 0 {
		return 0
	}
	if len(permutation) < int(c.nLeaves) {
		return -1
	}
	times := ifDelayTimes(p, c)
	if c.nLeaves == 1 {
		permutation[0] = 0
		return times[0]
	}
	size := ifDelayLutSize(p)
	if size == 0 {
		return -1
	}
	for pin := range times {
		permutation[pin] = 1
	}
	if int32(c.nLeaves) > size {
		if c.uMaskFunc == 0 {
			return -1
		}
		order := ifDelayDsdPermutation(p, c)
		for pin, literal := range order {
			if uint32(c.uMaskFunc)>>uint(2*pin)&1 != 0 {
				permutation[literal>>1] = 2
			}
		}
	}
	maximum := int32(0)
	for pin, time := range times {
		maximum = max(maximum, time+int32(permutation[pin]))
	}
	return maximum
}
func If_CutLutBalanceEval(p *If_Man_t, c *If_Cut_t) int32 {
	c.fUser = 1
	c.Cost = 0
	c.uMaskFunc = 0
	if c.nLeaves == 0 {
		return 0
	}
	times := ifDelayTimes(p, c)
	if c.nLeaves == 1 {
		return times[0]
	}
	size := ifDelayLutSize(p)
	if size == 0 {
		return -1
	}
	c.Cost = 1
	maximum := int32(-1)
	for _, time := range times {
		maximum = max(maximum, time)
	}
	if int32(c.nLeaves) <= size {
		return maximum + 1
	}
	if p.pIfDsdMan == nil {
		return -1
	}
	c.Cost = 2
	order := ifDelayDsdPermutation(p, c)
	count := int32(0)
	var latest uint32
	for pin, literal := range order {
		if times[literal>>1] == maximum {
			count++
			latest |= 1 << uint(2*pin)
		}
	}
	dsd := If_CutDsdLit(p, c)
	if count <= size-1 {
		mask := If_DsdManCheckXY(p.pIfDsdMan, dsd, size, 1, latest, 0, 0)
		if mask != 0 {
			c.uMaskFunc = int32(mask)
			return maximum + 1
		}
	}
	mask := If_DsdManCheckXY(p.pIfDsdMan, dsd, size, 1, 0, 0, 0)
	if mask == 0 {
		return -1
	}
	c.uMaskFunc = int32(mask)
	return maximum + 2
}
