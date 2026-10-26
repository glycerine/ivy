package abc

// Tests for owned MFFC transactions from aig/aig/aigMffc.c.

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"slices"
	"testing"
)

func aigMFFCFixture(t testing.TB, seed int) (*AigManager, []AigLit) {
	t.Helper()
	g := NewAigManager()
	ci := make([]AigLit, 6)
	leaf := make([]AigLit, 6)
	for i := range ci {
		ci[i], _ = g.CreateCI()
		leaf[i] = ci[i].NotIf(seed&(1<<uint(i%5)) != 0)
	}
	node := make([]AigLit, 7)
	node[0], _ = g.And(leaf[0], leaf[1])
	node[1], _ = g.And(leaf[2], leaf[3])
	node[2], _ = g.And(node[0], node[1])
	node[3], _ = g.And(node[0], leaf[4])
	node[4], _ = g.And(node[2], node[3])
	node[5], _ = g.And(node[1], leaf[5])
	node[6], _ = g.And(node[4], node[5])
	g.CreateCO(node[6])
	for _, index := range []int{0, 1, 3, 5} {
		g.CreateCO(node[index])
	}
	if seed&1 != 0 {
		g.CreateCO(node[2])
	}
	if seed&2 != 0 {
		g.CreateCO(node[4])
	}
	return g, node
}

func aigMFFCMemberIDs(members []bool) []AigID {
	var result []AigID
	for id, member := range members {
		if member {
			result = append(result, AigID(id))
		}
	}
	return result
}

func TestAigMFFCNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/aig_mffc.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Seed, LevelMinimum, Size, LabelSize, CutSize int
		Power                                        float32
		Support, Members, Cut, CutMembers, Extended  []AigID
		ExtendedExists                               int
	}
	if err = json.Unmarshal(data, &rows); err != nil || len(rows) != 32 {
		t.Fatal(len(rows), err)
	}
	for _, row := range rows {
		g, nodes := aigMFFCFixture(t, row.Seed)
		before := g.Clone()
		root := nodes[6].ID()
		result, err := g.MFFCSupport(root, row.LevelMinimum, AigMFFCOptions{})
		if err != nil || result.Size != row.Size || !slices.Equal(result.Support, row.Support) {
			t.Fatal(row.Seed, result, err, row.Size, row.Support)
		}
		probabilities := make([]float32, len(g.objects))
		for i := range probabilities {
			probabilities[i] = float32((i*17+row.Seed*11)%97+1) / 100
		}
		result, err = g.MFFCLabel(root, probabilities, AigMFFCOptions{})
		if err != nil || result.Size != row.LabelSize || result.Power != row.Power || !slices.Equal(aigMFFCMemberIDs(result.Members), row.Members) {
			t.Fatal(row.Seed, result.Size, result.Power, aigMFFCMemberIDs(result.Members), err, row)
		}
		cut, err := g.MFFCSupport(root, 1, AigMFFCOptions{})
		if err != nil || !slices.Equal(cut.Support, row.Cut) {
			t.Fatal(row.Seed, cut, err, row.Cut)
		}
		result, err = g.MFFCLabelCut(root, cut.Support, AigMFFCOptions{})
		if err != nil || result.Size != row.CutSize || !slices.Equal(aigMFFCMemberIDs(result.Members), row.CutMembers) {
			t.Fatal(row.Seed, result, aigMFFCMemberIDs(result.Members), err, row)
		}
		extended, exists, err := g.ExtendMFFCCut(root, cut.Support, AigMFFCOptions{})
		if err != nil || exists != (row.ExtendedExists != 0) || !slices.Equal(extended, row.Extended) {
			t.Fatal(row.Seed, extended, exists, err, row.Extended)
		}
		if !reflect.DeepEqual(g, before) {
			t.Fatal(row.Seed, "source mutation")
		}
	}
}

func TestAigMFFCErrorsLimitsAndDeep(t *testing.T) {
	g, nodes := aigMFFCFixture(t, 0)
	root := nodes[6].ID()
	ciResult, err := g.MFFCSupport(g.cis[0], 0, AigMFFCOptions{})
	if err != nil || ciResult.Size != 0 || !slices.Equal(ciResult.Support, []AigID{g.cis[0]}) {
		t.Fatal(ciResult, err)
	}
	for _, level := range []int{-1} {
		if result, err := g.MFFCSupport(root, level, AigMFFCOptions{}); result.Support != nil || !errors.Is(err, ErrDimension) {
			t.Fatal(result, err)
		}
	}
	for _, options := range []AigMFFCOptions{{MaxVisits: -1}, {MaxVisits: 1}, {Stop: func() bool { return true }}} {
		if result, err := g.MFFCSupport(root, 0, options); result.Support != nil || !(errors.Is(err, ErrDimension) || errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(options, result, err)
		}
	}
	if result, err := g.MFFCLabel(root, make([]float32, 1), AigMFFCOptions{}); result.Members != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	if result, err := g.MFFCLabelCut(root, []AigID{InvalidAigID}, AigMFFCOptions{}); result.Members != nil || !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(result, err)
	}
	if result, err := (*AigManager)(nil).MFFCSupport(root, 0, AigMFFCOptions{}); result.Support != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	if extended, exists, err := g.ExtendMFFCCut(root, []AigID{g.cis[0], g.cis[1]}, AigMFFCOptions{}); extended != nil || exists || err != nil {
		t.Fatal(extended, exists, err)
	}
	buffered := NewAigManager()
	a, _ := buffered.CreateCI()
	b, _ := buffered.CreateCI()
	buffer, _ := buffered.CreateBuffer(a)
	bufferRoot, _ := buffered.And(buffer, b)
	buffered.CreateCO(bufferRoot)
	if result, err := buffered.MFFCSupport(bufferRoot.ID(), 0, AigMFFCOptions{}); err != nil || result.Size != 1 || !slices.Equal(result.Support, []AigID{b.ID(), buffer.ID()}) {
		t.Fatal(result, err)
	}
	buffered = NewAigManager()
	a, _ = buffered.CreateCI()
	b, _ = buffered.CreateCI()
	internal, _ := buffered.And(a, b)
	buffer, _ = buffered.CreateBuffer(internal)
	c, _ := buffered.CreateCI()
	bufferRoot, _ = buffered.And(buffer, c)
	buffered.CreateCO(bufferRoot)
	if result, err := buffered.MFFCSupport(bufferRoot.ID(), 0, AigMFFCOptions{}); result.Support != nil || !errors.Is(err, ErrUnsupported) {
		t.Fatal(result, err)
	}

	const depth = 10000
	deep := NewAigManager()
	chain, _ := deep.CreateCI()
	for i := 0; i < depth; i++ {
		input, _ := deep.CreateCI()
		chain, _ = deep.And(chain, input)
	}
	deep.CreateCO(chain)
	result, err := deep.MFFCSupport(chain.ID(), 0, AigMFFCOptions{MaxVisits: 100000})
	if err != nil || result.Size != depth || len(result.Support) != depth+1 {
		t.Fatal(result.Size, len(result.Support), err)
	}
}

func aigMFFCReference(g *AigManager, root AigID, levelMinimum uint32) (int, []AigID) {
	references := aigMFFCReferences(g)
	var deref func(AigID) int
	deref = func(id AigID) int {
		object := g.objects[id]
		if object.Type == AigCI {
			return 0
		}
		count := 0
		for _, literal := range []AigLit{object.Fanin0, object.Fanin1} {
			fanin := literal.ID()
			references[fanin]--
			if references[fanin] == 0 && (levelMinimum == 0 || g.objects[fanin].Level > levelMinimum) {
				count += deref(fanin)
			}
		}
		return count + 1
	}
	size := deref(root)
	seen := make([]bool, len(g.objects))
	var support []AigID
	var visit func(AigID, bool)
	visit = func(id AigID, top bool) {
		if seen[id] {
			return
		}
		seen[id] = true
		object := g.objects[id]
		if !top && (object.Type == AigCI || references[id] > 0 || object.Level <= levelMinimum) {
			support = append(support, id)
			return
		}
		visit(object.Fanin0.ID(), false)
		visit(object.Fanin1.ID(), false)
	}
	visit(root, true)
	return size, support
}

func FuzzAigMFFC(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6, 7})
	f.Fuzz(func(t *testing.T, data []byte) {
		g := NewAigManager()
		signals := make([]AigLit, 6)
		for i := range signals {
			signals[i], _ = g.CreateCI()
		}
		var roots []AigLit
		for i := 0; i+1 < len(data) && len(signals) < 70; i += 2 {
			left := signals[int(data[i]&0x3f)%len(signals)].NotIf(data[i]&0x80 != 0)
			right := signals[int(data[i+1]&0x3f)%len(signals)].NotIf(data[i+1]&0x80 != 0)
			before := g.AndCount()
			literal, err := g.And(left, right)
			if err == nil {
				signals = append(signals, literal)
				if g.AndCount() > before {
					roots = append(roots, literal)
				}
			}
		}
		if len(roots) == 0 {
			return
		}
		root := roots[len(roots)-1]
		g.CreateCO(root)
		for i, literal := range roots[:len(roots)-1] {
			if len(data) != 0 && data[i%len(data)]&4 != 0 {
				g.CreateCO(literal)
			}
		}
		level := 0
		if len(data) != 0 {
			level = int(data[0] & 3)
		}
		wantSize, wantSupport := aigMFFCReference(g, root.ID(), uint32(level))
		got, err := g.MFFCSupport(root.ID(), level, AigMFFCOptions{})
		if err != nil || got.Size != wantSize || !slices.Equal(got.Support, wantSupport) {
			t.Fatal(got, err, wantSize, wantSupport)
		}
		label, err := g.MFFCLabel(root.ID(), nil, AigMFFCOptions{})
		if err != nil || label.Size != aigMFFCMemberCount(label.Members) {
			t.Fatal(label, err)
		}
	})
}

func aigMFFCMemberCount(members []bool) int {
	count := 0
	for _, member := range members {
		if member {
			count++
		}
	}
	return count
}
