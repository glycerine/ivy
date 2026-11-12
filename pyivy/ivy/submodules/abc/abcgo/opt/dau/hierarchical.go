package dau

import (
	"encoding/binary"
	"github.com/glycerine/abcgo/bool/lucky"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau/canon"
)

const TT_MAX_LEVELS = 5

type dauHieLevel struct {
	ids             map[string]int
	truths          [][]uint64
	representatives []int
}
type Abc_TtHieMan_t struct {
	nWords   int
	levels   []dauHieLevel
	truthIDs []int
	vPhase   *vec.Vec_Int_t
}

func Abc_TtHieManStart(nVars, nLevels int32) *Abc_TtHieMan_t {
	if nLevels > TT_MAX_LEVELS {
		return nil
	}
	if nLevels < 0 {
		nLevels = 0
	}
	p := &Abc_TtHieMan_t{nWords: dauCountWords(nVars), levels: make([]dauHieLevel, nLevels), truthIDs: make([]int, nLevels), vPhase: vec.Vec_IntAlloc(2500)}
	for i := range p.levels {
		p.levels[i].ids = map[string]int{}
	}
	return p
}
func Abc_TtHieManStop(p *Abc_TtHieMan_t) {
	if p != nil {
		*p = Abc_TtHieMan_t{}
	}
}
func dauTruthKey(t []uint64) string {
	b := make([]byte, 8*len(t))
	for i, w := range t {
		binary.LittleEndian.PutUint64(b[i*8:], w)
	}
	return string(b)
}
func Abc_TtHieRetrieveOrInsert(p *Abc_TtHieMan_t, level int32, truth, result []uint64) int32 {
	if p == nil {
		return -1
	}
	i := int(level)
	if i < 0 {
		i += len(p.levels)
	}
	if i < 0 || i >= len(p.levels) {
		return -1
	}
	l := &p.levels[i]
	key := dauTruthKey(truth[:p.nWords])
	spot, ok := l.ids[key]
	if !ok {
		spot = len(l.truths)
		l.ids[key] = spot
		l.truths = append(l.truths, append([]uint64(nil), truth[:p.nWords]...))
		p.truthIDs[i] = spot
		if i < len(p.levels)-1 {
			return 0
		}
	}
	representative := spot
	if i < len(p.levels)-1 {
		representative = l.representatives[spot]
	}
	for earlier := 0; earlier < i; earlier++ {
		row := &p.levels[earlier]
		id := p.truthIDs[earlier]
		for len(row.representatives) <= id {
			row.representatives = append(row.representatives, 0)
		}
		row.representatives[id] = representative
	}
	copy(result[:p.nWords], p.levels[len(p.levels)-1].truths[representative])
	if i < len(p.levels)-1 {
		return 1
	}
	return 0
}
func Abc_TtCanonicize(truth []uint64, nVars int32, perm []byte) uint32 {
	return canon.Canonicalize(truth, nVars, perm)
}
func Abc_TtCanonicizePerm(truth []uint64, nVars int32, perm []byte) uint32 {
	return canon.CanonicalizePermutation(truth, nVars, perm)
}
func Abc_TtCanonicizePhase(truth []uint64, nVars int32) uint32 {
	return canon.CanonicalizePhase(truth, nVars)
}
func Abc_TtCanonicizeHie(p *Abc_TtHieMan_t, truth []uint64, nVars int32, perm []byte, exact int32) uint32 {
	if nVars == 0 {
		truth[0] = 0
		return 0
	}
	hit := false
	var counts []int
	canon.CanonicalizeWithCheckpoints(truth, nVars, perm, func(stage int, c []int) bool {
		counts = append(counts[:0], c...)
		if Abc_TtHieRetrieveOrInsert(p, int32(stage), truth, truth) > 0 {
			hit = true
		}
		return hit
	})
	if hit {
		return 0
	}
	if exact != 0 && counts[nVars] != 0 {
		var groups []int32
		for v := 0; v < int(nVars); {
			end := v + 1
			for end < int(nVars) && counts[end] == counts[v] {
				end++
			}
			groups = append(groups, int32(end-v))
			v = end
		}
		infos := make([]*lucky.PermInfo, nVars+1)
		for _, n := range groups {
			if infos[n] == nil {
				infos[n] = lucky.NewPermInfo(n)
			}
		}
		words := dauCountWords(nVars)
		aux, minimal := make([]uint64, words), make([]uint64, words)
		flipOutput := counts[nVars] == words*32
		flipInput := flipOutput || counts[0] != counts[1] && counts[0] == counts[nVars]-counts[0]
		lucky.SimpleMinimalGroups(truth, aux, minimal, groups, infos, nVars, flipOutput, flipInput)
	}
	Abc_TtHieRetrieveOrInsert(p, 4, truth, truth)
	return 0
}
func Abc_TtCanonicizeWrap(fn TtCanonicizeFunc, p *Abc_TtHieMan_t, truth []uint64, nVars int32, perm []byte, flag int32) uint32 {
	Abc_TtNormalizeSmallTruth(truth, nVars)
	words := dauCountWords(nVars)
	if Abc_TtCountOnesInTruth(truth, nVars) != int32(words*32) {
		return fn(p, truth, nVars, perm, flag)
	}
	other := make([]uint64, words)
	for i, w := range truth[:words] {
		other[i] = ^w
	}
	pOther := make([]byte, nVars)
	phase := fn(p, truth, nVars, perm, flag)
	phaseOther := fn(p, other, nVars, pOther, flag)
	for i := words - 1; i >= 0; i-- {
		if truth[i] < other[i] {
			return phase
		}
		if truth[i] > other[i] {
			copy(truth, other)
			copy(perm, pOther)
			// The second callback saw the complemented input; include that
			// extra output inversion when describing the original input.
			return phaseOther ^ (1 << uint(nVars))
		}
	}
	return phase
}
