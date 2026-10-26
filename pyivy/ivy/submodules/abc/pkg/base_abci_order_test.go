package abc

// Standalone tests for src/base/abci/abcOrder.c.

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func abcCIOrderFixture(t *testing.T) (*ABCNetwork, []int) {
	t.Helper()
	network, err := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{})
	if err != nil {
		t.Fatal(err)
	}
	a, _ := network.CreatePI()
	b, _ := network.CreatePI()
	c, _ := network.CreatePI()
	state, _ := network.CreateBO()
	for id, name := range map[int]string{a: "a", b: "b", c: "c", state: "s"} {
		if err = network.SetObjectName(id, name); err != nil {
			t.Fatal(err)
		}
	}
	left, _ := network.CreateNode()
	right, _ := network.CreateNode()
	if err = network.AddFanin(left, ABCObjectEdge{ID: c}); err != nil {
		t.Fatal(err)
	}
	if err = network.AddFanin(left, ABCObjectEdge{ID: a}); err != nil {
		t.Fatal(err)
	}
	if err = network.AddFanin(right, ABCObjectEdge{ID: state}); err != nil {
		t.Fatal(err)
	}
	if err = network.AddFanin(right, ABCObjectEdge{ID: b}); err != nil {
		t.Fatal(err)
	}
	first, _ := network.CreatePO()
	second, _ := network.CreatePO()
	if err = network.AddFanin(first, ABCObjectEdge{ID: left}); err != nil {
		t.Fatal(err)
	}
	if err = network.AddFanin(second, ABCObjectEdge{ID: right}); err != nil {
		t.Fatal(err)
	}
	return network, []int{a, b, c, state}
}

func TestABCFindAndChangeCIOrder(t *testing.T) {
	network, original := abcCIOrderFixture(t)
	order, err := ABCFindCIOrder(network, false, ABCCIOrderOptions{})
	want := []int{original[2], original[0], original[3], original[1]}
	if err != nil || !slices.Equal(order, want) || !slices.Equal(network.PIs(), []int{original[2], original[0], original[1]}) {
		t.Fatalf("order=%v pis=%v want=%v err=%v", order, network.PIs(), want, err)
	}
	if err = network.Validate(); err != nil {
		t.Fatalf("valid reordered network: %v", err)
	}
	if err = ABCChangeCIOrder(network, original, true, ABCCIOrderOptions{}); err != nil {
		t.Fatal(err)
	}
	want = []int{original[3], original[2], original[1], original[0]}
	if !slices.Equal(network.CIs(), want) || !slices.Equal(network.PIs(), want[1:]) {
		t.Fatalf("reverse cis=%v pis=%v", network.CIs(), network.PIs())
	}
	corrupt := network.Clone()
	corrupt.cis[0] = corrupt.cis[1]
	if err = corrupt.Validate(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("duplicate terminal partition: %v", err)
	}
}

func TestABCImplementCIOrderOwnedBytesAndErrors(t *testing.T) {
	network, original := abcCIOrderFixture(t)
	data := []byte("b\n s\ta c")
	order, err := ABCImplementCIOrder(network, data, false, ABCCIOrderOptions{})
	if err != nil || !slices.Equal(order, []int{original[1], original[3], original[0], original[2]}) || !slices.Equal(network.PIs(), []int{original[1], original[0], original[2]}) {
		t.Fatalf("order=%v pis=%v err=%v", order, network.PIs(), err)
	}
	data[0] = 'x'
	if network.objects[order[0]].Name != "b" {
		t.Fatal("text input was retained")
	}
	for _, bad := range [][]byte{
		[]byte("a b c"), []byte("a b c missing"), []byte("a b c a"), []byte("a b c n"),
	} {
		fresh, ids := abcCIOrderFixture(t)
		node, _ := fresh.CreateNode()
		if err = fresh.SetObjectName(node, "n"); err != nil {
			t.Fatal(err)
		}
		before := fresh.CIs()
		got, gotErr := ABCImplementCIOrder(fresh, bad, false, ABCCIOrderOptions{})
		if got != nil || !errors.Is(gotErr, ErrDimension) || !slices.Equal(fresh.CIs(), before) || !slices.Equal(before, ids) {
			t.Fatalf("bad %q: got=%v err=%v cis=%v", bad, got, gotErr, fresh.CIs())
		}
	}
}

func TestABCCIOrderLimitsCancellationAndAtomicity(t *testing.T) {
	network, original := abcCIOrderFixture(t)
	for _, options := range []ABCCIOrderOptions{{MaxCIs: -1}, {MaxVisited: -1}, {MaxInputBytes: -1}} {
		if order, err := ABCFindCIOrder(network, false, options); order != nil || !errors.Is(err, ErrDimension) {
			t.Fatalf("negative options: order=%v err=%v", order, err)
		}
	}
	if order, err := ABCFindCIOrder(network, false, ABCCIOrderOptions{MaxCIs: 3}); order != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("CI limit: order=%v err=%v", order, err)
	}
	if order, err := ABCFindCIOrder(network, false, ABCCIOrderOptions{MaxVisited: 1}); order != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("visit limit: order=%v err=%v", order, err)
	}
	if order, err := ABCImplementCIOrder(network, []byte("a b c s"), false, ABCCIOrderOptions{MaxInputBytes: 2}); order != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("byte limit: order=%v err=%v", order, err)
	}
	calls := 0
	probe, _ := abcCIOrderFixture(t)
	if _, err := ABCFindCIOrder(probe, false, ABCCIOrderOptions{Stop: func() bool { calls++; return false }}); err != nil {
		t.Fatal(err)
	}
	for stopAt := 1; stopAt <= calls; stopAt++ {
		fresh, _ := abcCIOrderFixture(t)
		before := fresh.CIs()
		count := 0
		order, err := ABCFindCIOrder(fresh, false, ABCCIOrderOptions{Stop: func() bool { count++; return count == stopAt }})
		if order != nil || !errors.Is(err, ErrCancelled) || !slices.Equal(fresh.CIs(), before) {
			t.Fatalf("stop %d/%d: order=%v err=%v cis=%v", stopAt, calls, order, err, fresh.CIs())
		}
	}
	if !slices.Equal(network.CIs(), original) {
		t.Fatal("failed operations mutated source")
	}
}

func FuzzABCCIOrder(f *testing.F) {
	f.Add([]byte{3, 2, 1, 0}, false)
	f.Add([]byte{}, true)
	f.Fuzz(func(t *testing.T, data []byte, reverse bool) {
		count := 2
		if len(data) > 0 {
			count += int(data[0] % 6)
		}
		network, _ := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{})
		ids := make([]int, count)
		for i := range ids {
			ids[i], _ = network.CreatePI()
			if err := network.SetObjectName(ids[i], "i"+string(rune('a'+i))); err != nil {
				t.Fatal(err)
			}
		}
		permutation := slices.Clone(ids)
		for i := count - 1; i > 0; i-- {
			value := byte(i)
			if len(data) > 0 {
				value = data[i%len(data)]
			}
			j := int(value) % (i + 1)
			permutation[i], permutation[j] = permutation[j], permutation[i]
		}
		for _, id := range permutation {
			po, _ := network.CreatePO()
			if err := network.AddFanin(po, ABCObjectEdge{ID: id}); err != nil {
				t.Fatal(err)
			}
		}
		parsed := network.Clone()
		found, err := ABCFindCIOrder(network, reverse, ABCCIOrderOptions{})
		want := slices.Clone(permutation)
		if reverse {
			slices.Reverse(want)
		}
		if err != nil || !slices.Equal(found, want) || !slices.Equal(network.PIs(), want) {
			t.Fatalf("found=%v want=%v err=%v", found, want, err)
		}
		var text strings.Builder
		for i, id := range permutation {
			if i&1 == 0 {
				text.WriteByte('\n')
			} else {
				text.WriteByte('\t')
			}
			text.WriteString(parsed.objects[id].Name)
		}
		implemented, err := ABCImplementCIOrder(parsed, []byte(text.String()), reverse, ABCCIOrderOptions{})
		if err != nil || !slices.Equal(implemented, want) || !slices.Equal(parsed.PIs(), want) {
			t.Fatalf("implemented=%v want=%v err=%v", implemented, want, err)
		}
	})
}
