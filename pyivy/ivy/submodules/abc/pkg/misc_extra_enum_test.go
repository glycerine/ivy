package abc

// Go-only fixed-data, structural, exhaustive, and fuzz tests for extraUtilEnum.c.

import (
	"errors"
	"slices"
	"testing"
)

func TestExtraEnumerationProblems(t *testing.T) {
	first, second, third := ExtraEnumerationProblemFirst(), ExtraEnumerationProblemSecond(), ExtraEnumerationProblemThird()
	for _, row := range []struct {
		problem                        ExtraEnumerationProblem
		variables, minterms, functions int
	}{{first, 8, 16, 8}, {second, 10, 32, 7}, {third, 8, 16, 7}} {
		if row.problem.Variables != row.variables || len(row.problem.Minterms) != row.minterms || len(row.problem.Functions) != row.functions || len(row.problem.VariableTruths) != row.variables {
			t.Fatal(row.variables, row.problem)
		}
		for minterm, mask := range row.problem.Minterms {
			for variable := 0; variable < row.variables; variable++ {
				fromColumn := row.problem.VariableTruths[variable]&(1<<uint(minterm)) != 0
				fromRow := mask&(1<<uint(variable)) != 0
				if fromColumn != fromRow {
					t.Fatal(row.variables, minterm, variable)
				}
			}
		}
	}
	if first.Minterms[0] != 0x55 || first.VariableTruths[0] != 0x0f0f {
		t.Fatal(first.Minterms[0], first.VariableTruths[0])
	}
	first.Minterms[0] = 0
	if ExtraEnumerationProblemFirst().Minterms[0] != 0x55 {
		t.Fatal("ownership")
	}
	if _, err := newExtraEnumerationProblem([]string{"1x"}, []string{"1"}); !errors.Is(err, ErrFormat) {
		t.Fatal(err)
	}
}

func TestExtraFunctionEnumerationOrderExpressionsAndLimits(t *testing.T) {
	problem := ExtraEnumerationProblem{
		Variables:      2,
		Minterms:       []uint32{0, 1, 2, 3},
		VariableTruths: []uint32{0b1010, 0b1100},
		Functions:      []uint32{0b1000, 0b1110},
	}
	result, err := EnumerateExtraFunctions(problem, 1, ExtraFunctionEnumerationOptions{})
	if err != nil || !slices.Equal(result.LayerTotals, []int{0, 2, 4}) || result.Attempts != 2 || len(result.Discoveries) != 2 {
		t.Fatal(result, err)
	}
	if result.Discoveries[0] != (ExtraFunctionDiscovery{0, 2, 1, "ab"}) || result.Discoveries[1] != (ExtraFunctionDiscovery{1, 3, 1, "a+b"}) {
		t.Fatal(result.Discoveries)
	}
	for index, node := range result.Nodes {
		if index < problem.Variables {
			continue
		}
		want := result.Nodes[node.First].Truth & result.Nodes[node.Second].Truth
		if node.Or {
			want = result.Nodes[node.First].Truth | result.Nodes[node.Second].Truth
		}
		if node.Truth != want {
			t.Fatal(index, node, want)
		}
	}
	if _, err = FormatExtraDecomposition(result.Nodes, len(result.Nodes), 2); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if partial, err := EnumerateExtraFunctions(problem, 1, ExtraFunctionEnumerationOptions{MaxNodes: 3}); !errors.Is(err, ErrLimit) || len(partial.Nodes) != 3 {
		t.Fatal(partial, err)
	}
	if partial, err := EnumerateExtraFunctions(problem, 1, ExtraFunctionEnumerationOptions{MaxAttempts: 1}); !errors.Is(err, ErrLimit) || partial.Attempts != 2 {
		t.Fatal(partial.Attempts, err)
	}
	if _, err := EnumerateExtraFunctions(problem, 1, ExtraFunctionEnumerationOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
}

func TestExtraFunctionEnumerationSecondProblemPrefix(t *testing.T) {
	result, err := EnumerateExtraFunctions(ExtraEnumerationProblemSecond(), 5, ExtraFunctionEnumerationOptions{MaxNodes: 120000, MaxAttempts: 1 << 24})
	wantLayers := []int{0, 10, 88, 640, 4200, 23608, 116764}
	if err != nil || !slices.Equal(result.LayerTotals, wantLayers) {
		t.Fatal(result.LayerTotals, len(result.Nodes), result.Attempts, err)
	}
	if len(result.Discoveries) != 1 || result.Discoveries[0] != (ExtraFunctionDiscovery{Function: 6, Node: 13, Gates: 1, Expression: "a+c"}) {
		t.Fatal(result.Discoveries)
	}
}

func checkExtraNetwork(t testing.TB, network ExtraEnumeratedNetwork) {
	t.Helper()
	patterns := []uint64{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0, 0xff00ff00ff00ff00, 0xffff0000ffff0000, 0xffffffff00000000}
	truths := append([]uint64(nil), patterns[:network.Variables]...)
	refs := make([]int, network.Variables+len(network.Gates))
	for offset, gate := range network.Gates {
		node := network.Variables + offset
		if gate.First < 0 || gate.Second <= gate.First || gate.Second >= node {
			t.Fatal(node, gate)
		}
		x, y := truths[gate.First], truths[gate.Second]
		var truth uint64
		if gate.XOR {
			truth = x ^ y
		} else {
			if gate.ComplementFirst {
				x = ^x
			}
			if gate.ComplementSecond {
				y = ^y
			}
			truth = x & y
		}
		if truth != gate.Truth {
			t.Fatal(node, truth, gate.Truth)
		}
		truths = append(truths, truth)
		refs[gate.First]++
		refs[gate.Second]++
	}
	tops := 0
	for node := range refs {
		if refs[node] == 0 {
			tops++
		}
	}
	if tops != 1 || truths[len(truths)-1] != network.Truth {
		t.Fatal(tops, truths[len(truths)-1], network.Truth)
	}
}

func TestExtraNetworkEnumerationPrimitiveCountsAndStructures(t *testing.T) {
	andOnly, err := EnumerateExtraNetworks(2, 1, ExtraNetworkEnumerationOptions{ReturnNetworks: true})
	if err != nil || !andOnly.Feasible || andOnly.Tries != 1 || andOnly.Builds != 1 || andOnly.Finished != 1 || len(andOnly.Networks) != 1 {
		t.Fatal(andOnly, err)
	}
	checkExtraNetwork(t, andOnly.Networks[0])
	if text, err := FormatExtraEnumeratedNetwork(andOnly.Networks[0]); err != nil || text != "c=ab" {
		t.Fatal(text, err)
	}
	withXOR, err := EnumerateExtraNetworks(2, 1, ExtraNetworkEnumerationOptions{UseXOR: true, ReturnNetworks: true})
	if err != nil || withXOR.Tries != 2 || withXOR.Builds != 2 || withXOR.Finished != 2 || len(withXOR.Networks) != 2 {
		t.Fatal(withXOR, err)
	}
	for _, network := range withXOR.Networks {
		checkExtraNetwork(t, network)
	}
	if text, _ := FormatExtraEnumeratedNetwork(withXOR.Networks[1]); text != "c=a+b" {
		t.Fatal(text)
	}
	three, err := EnumerateExtraNetworks(3, 2, ExtraNetworkEnumerationOptions{ReturnNetworks: true})
	if err != nil || three.Finished == 0 || uint64(len(three.Networks)) != three.Finished {
		t.Fatal(three, err)
	}
	for _, network := range three.Networks {
		checkExtraNetwork(t, network)
	}
	again, err := EnumerateExtraNetworks(3, 2, ExtraNetworkEnumerationOptions{})
	if err != nil || again.Tries != three.Tries || again.Builds != three.Builds || again.Finished != three.Finished {
		t.Fatal(three.Tries, again.Tries, three.Builds, again.Builds, three.Finished, again.Finished, err)
	}
	if infeasible, err := EnumerateExtraNetworks(4, 2, ExtraNetworkEnumerationOptions{}); err != nil || infeasible.Feasible || infeasible.Finished != 0 {
		t.Fatal(infeasible, err)
	}
}

func TestExtraNetworkEnumerationLimitsCancellationAndFormatting(t *testing.T) {
	if result, err := EnumerateExtraNetworks(3, 2, ExtraNetworkEnumerationOptions{MaxTries: 1}); !errors.Is(err, ErrLimit) || result.Tries != 1 {
		t.Fatal(result, err)
	}
	if result, err := EnumerateExtraNetworks(3, 2, ExtraNetworkEnumerationOptions{MaxBuilds: 1}); !errors.Is(err, ErrLimit) || result.Builds != 1 {
		t.Fatal(result, err)
	}
	if result, err := EnumerateExtraNetworks(3, 2, ExtraNetworkEnumerationOptions{MaxFinished: 1}); !errors.Is(err, ErrLimit) || result.Finished != 1 {
		t.Fatal(result, err)
	}
	if _, err := EnumerateExtraNetworks(3, 2, ExtraNetworkEnumerationOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	for _, row := range [][2]int{{1, 1}, {7, 6}, {2, 0}, {6, 10}} {
		if _, err := EnumerateExtraNetworks(row[0], row[1], ExtraNetworkEnumerationOptions{}); !errors.Is(err, ErrDimension) {
			t.Fatal(row, err)
		}
	}
	bad := ExtraEnumeratedNetwork{Variables: 2, Gates: []ExtraEnumeratedGate{{First: 1, Second: 0}}}
	if _, err := FormatExtraEnumeratedNetwork(bad); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func FuzzExtraEnumerations(f *testing.F) {
	f.Add(uint8(2), uint32(0x96), uint32(0x3c))
	f.Add(uint8(1), uint32(0xaa), uint32(0xcc))
	f.Fuzz(func(t *testing.T, flags uint8, firstTruth, secondTruth uint32) {
		minterms := 4 + int(flags%5)
		mask := uint32(1<<uint(minterms)) - 1
		firstTruth &= mask
		secondTruth &= mask
		if firstTruth == secondTruth {
			secondTruth ^= 1
		}
		problem := ExtraEnumerationProblem{Variables: 2, Minterms: make([]uint32, minterms), VariableTruths: []uint32{firstTruth, secondTruth}}
		result, err := EnumerateExtraFunctions(problem, 2, ExtraFunctionEnumerationOptions{MaxNodes: 1 << 12})
		if err != nil {
			t.Fatal(err)
		}
		seen := map[uint32]struct{}{}
		for index, node := range result.Nodes {
			if _, duplicate := seen[node.Truth]; duplicate {
				t.Fatal("duplicate", index, node.Truth)
			}
			seen[node.Truth] = struct{}{}
			if index < 2 {
				continue
			}
			want := result.Nodes[node.First].Truth & result.Nodes[node.Second].Truth
			if node.Or {
				want = result.Nodes[node.First].Truth | result.Nodes[node.Second].Truth
			}
			if node.First >= node.Second || node.Truth != want || node.Gates != result.Nodes[node.First].Gates+result.Nodes[node.Second].Gates+1 {
				t.Fatal(index, node, want)
			}
			if _, err := FormatExtraDecomposition(result.Nodes, index, 2); err != nil {
				t.Fatal(err)
			}
		}
		variables := 2 + int(flags&1)
		gates := variables - 1
		networks, err := EnumerateExtraNetworks(variables, gates, ExtraNetworkEnumerationOptions{UseXOR: flags&2 != 0, ReturnNetworks: true, MaxFinished: 1 << 14})
		if err != nil {
			t.Fatal(variables, gates, err)
		}
		if uint64(len(networks.Networks)) != networks.Finished {
			t.Fatal(len(networks.Networks), networks.Finished)
		}
		for _, network := range networks.Networks {
			checkExtraNetwork(t, network)
		}
	})
}
