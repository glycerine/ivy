package abc

// Go-only tests for src/bdd/mtr.

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestMTRBasicTreeConstructionCloneAndFormat(t *testing.T) {
	tree, err := NewMTRTree(MTRLimits{})
	if err != nil {
		t.Fatal(err)
	}
	root := tree.Root()
	root.Low = 2
	root.Size = 9
	root.Index = 3
	last, err := tree.CreateLastChild(root)
	if err != nil {
		t.Fatal(err)
	}
	last.Low, last.Size, last.Index = 6, 2, 4
	first, err := tree.CreateFirstChild(root)
	if err != nil {
		t.Fatal(err)
	}
	first.Low, first.Size = 2, 3
	middle, err := tree.newNode()
	if err != nil {
		t.Fatal(err)
	}
	middle.Low, middle.Size = 5, 1
	if err = tree.MakeNextSibling(first, middle); err != nil {
		t.Fatal(err)
	}
	if root.Child() != first || first.Younger() != middle || middle.Elder() != first || middle.Younger() != last || last.Parent() != root {
		t.Fatal("relationship order")
	}
	if err = tree.Validate(); err != nil {
		t.Fatal(err)
	}
	other, _ := NewMTRTree(MTRLimits{})
	if err = tree.MakeLastChild(root, other.Root()); !errors.Is(err, ErrDimension) {
		t.Fatal("cross-tree attachment", err)
	}
	text, err := tree.FormatTree()
	if err != nil {
		t.Fatal(err)
	}
	want := "N=1 C=2 Y=0 E=0 P=0 F=0 L=2 S=9\nN=2 C=0 Y=3 E=0 P=1 F=0 L=2 S=3\nN=3 C=0 Y=4 E=2 P=1 F=0 L=5 S=1\nN=4 C=0 Y=0 E=3 P=1 F=0 L=6 S=2\n"
	if text != want {
		t.Fatalf("%q", text)
	}
	clone, err := tree.Clone(3, MTRLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if clone.Root() == root || clone.Root().Low != 6 || clone.Root().Size != 27 || clone.Root().Index != 9 || clone.Root().Child().Low != 6 || clone.Root().Child().Younger().Low != 15 {
		t.Fatal("clone")
	}
	root.Low++
	if clone.Root().Low != 6 {
		t.Fatal("clone aliases")
	}
	clone.Close()
	clone.Close()
	if clone.Root() != nil || clone.NodeCount() != 0 {
		t.Fatal("close")
	}
}

func TestMTRBasicTreeDoesNotRequireGroupIntervals(t *testing.T) {
	tree, err := NewMTRTree(MTRLimits{})
	if err != nil {
		t.Fatal(err)
	}
	first, err := tree.CreateLastChild(tree.Root())
	if err != nil {
		t.Fatal(err)
	}
	second, err := tree.CreateLastChild(tree.Root())
	if err != nil {
		t.Fatal(err)
	}
	// mtrBasic nodes are not groups: their scalar fields may remain zero or
	// appear in any order without invalidating the relationship tree.
	first.Low, first.Size = 9, 0
	second.Low, second.Size = 1, 0
	if err = tree.Validate(); err != nil {
		t.Fatal(err)
	}
	if _, err = tree.FormatTree(); err != nil {
		t.Fatal(err)
	}
	clone, err := tree.Clone(1, MTRLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if err = clone.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestMTRGroupsMakeFindFormatFlagsAndCutRejection(t *testing.T) {
	tree, err := NewMTRGroupTree(0, 10, MTRLimits{})
	if err != nil {
		t.Fatal(err)
	}
	g24, err := tree.MakeGroup(2, 4, MTRFixed)
	if err != nil {
		t.Fatal(err)
	}
	g35, err := tree.MakeGroup(3, 2, MTRSoft)
	if err != nil {
		t.Fatal(err)
	}
	g02, err := tree.MakeGroup(0, 2, MTRNewNode)
	if err != nil {
		t.Fatal(err)
	}
	g68, err := tree.MakeGroup(6, 2, MTRTerminal)
	if err != nil {
		t.Fatal(err)
	}
	if g24.Child() != g35 || g02.Younger() != g24 || g24.Younger() != g68 || tree.FindGroup(3, 2) != g35 || tree.FindGroup(4, 4) != nil {
		t.Fatal("group layout")
	}
	before := tree.NodeCount()
	if _, err = tree.MakeGroup(1, 3, MTRDefault); !errors.Is(err, ErrDimension) || tree.NodeCount() != before {
		t.Fatalf("cut %v", err)
	}
	if got, _ := tree.FormatGroups(); got != "(0(0,1|N)(2(3,4|S)5|F)(6,7|)9)\n" {
		t.Fatalf("%q", got)
	}
	if _, err = tree.MakeGroup(2, 4, MTRSoft|MTRFixed); err != nil {
		t.Fatal(err)
	}
	if g24.Flags != MTRSoft|MTRFixed {
		t.Fatal(g24.Flags)
	}
	if err = tree.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestMTRSwapAndDissolve(t *testing.T) {
	tree, _ := NewMTRGroupTree(0, 8, MTRLimits{})
	a, _ := tree.MakeGroup(0, 3, MTRFixed)
	aChild, _ := tree.MakeGroup(1, 1, MTRSoft)
	b, _ := tree.MakeGroup(3, 2, MTRNewNode)
	bChild, _ := tree.MakeGroup(3, 1, MTRFixed)
	c, _ := tree.MakeGroup(5, 3, MTRDefault)
	if err := tree.SwapGroups(a, b); err != nil {
		t.Fatal(err)
	}
	if tree.Root().Child() != b || b.Low != 0 || bChild.Low != 0 || a.Low != 2 || aChild.Low != 3 || a.Younger() != c {
		t.Fatal("swap result")
	}
	if err := tree.SwapGroups(b, c); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	parent, err := tree.DissolveGroup(a)
	if err != nil || parent != tree.Root() {
		t.Fatal(parent, err)
	}
	if b.Younger() != aChild || aChild.Younger() != c || a.live || tree.NodeCount() != 5 {
		t.Fatal("dissolve")
	}
	if _, err = tree.DissolveGroup(b); err != nil {
		t.Fatal(err)
	}
	if err = tree.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestMTRSwapTerminalGroupLeavesHiddenDescendantsUnshifted(t *testing.T) {
	tree, _ := NewMTRGroupTree(0, 8, MTRLimits{})
	first, _ := tree.MakeGroup(0, 4, MTRDefault)
	hidden, _ := tree.MakeGroup(1, 2, MTRSoft)
	second, _ := tree.MakeGroup(4, 4, MTRFixed)
	if err := first.Set(MTRTerminal); err != nil {
		t.Fatal(err)
	}
	if err := tree.SwapGroups(first, second); err != nil {
		t.Fatal(err)
	}
	if first.Low != 4 || second.Low != 0 || hidden.Low != 1 {
		t.Fatalf("terminal shift: first=%d second=%d hidden=%d", first.Low, second.Low, hidden.Low)
	}
	if err := tree.Validate(); err != nil {
		t.Fatal(err)
	}
	if got, err := tree.FormatGroups(); err != nil || got != "(0(0,3|F)(4,7|)7)\n" {
		t.Fatalf("FormatGroups() = %q, %v", got, err)
	}
}

func TestMTRParseGroupsOwnershipAndErrors(t *testing.T) {
	data := []byte("0 8 D\n0 4 F\n1 2 S\n4 4 N\n")
	tree, err := ParseMTRGroups(data, 8, MTRLimits{})
	if err != nil {
		t.Fatal(err)
	}
	data[0] = '9'
	if got, _ := tree.FormatGroups(); got != "(0(0(1,2|S)3|F)(4,7|N)7)\n" {
		t.Fatalf("%q", got)
	}
	for _, bad := range []string{"0 2", "x 2 F", "0 0 F", "0 9 F", "0 2 X", "0 5 F 2 5 S"} {
		if parsed, e := ParseMTRGroups([]byte(bad), 8, MTRLimits{}); parsed != nil || e == nil {
			t.Fatalf("%q: %v", bad, e)
		}
	}
	if parsed, e := ParseMTRGroups([]byte("0 1 "+strings.Repeat("D", 33)), 2, MTRLimits{}); parsed != nil || !errors.Is(e, ErrFormat) {
		t.Fatal(parsed, e)
	}
	if _, err = ParseMTRGroups([]byte("0 1 F"), 2, MTRLimits{MaxInputBytes: 2}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = ParseMTRGroups(nil, -1, MTRLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestMTRLimitsCancellationAtomicityAndLifecycle(t *testing.T) {
	if _, err := NewMTRTree(MTRLimits{MaxNodes: -1}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := NewMTRGroupTree(int(MTRMaxHigh)+1, 0, MTRLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := NewMTRTree(MTRLimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	shallowBasic, _ := NewMTRTree(MTRLimits{MaxDepth: 2})
	child, err := shallowBasic.CreateFirstChild(shallowBasic.Root())
	if err != nil {
		t.Fatal(err)
	}
	if _, err = shallowBasic.CreateFirstChild(child); !errors.Is(err, ErrLimit) || shallowBasic.NodeCount() != 2 {
		t.Fatal("basic depth limit was not atomic", err, shallowBasic.NodeCount())
	}
	shallowGroups, _ := NewMTRGroupTree(0, 4, MTRLimits{MaxDepth: 1})
	if _, err = shallowGroups.MakeGroup(0, 2, MTRDefault); !errors.Is(err, ErrLimit) || shallowGroups.NodeCount() != 1 {
		t.Fatal("group depth limit was not atomic", err, shallowGroups.NodeCount())
	}
	tree, _ := NewMTRGroupTree(0, 16, MTRLimits{MaxNodes: 2})
	if _, err := tree.MakeGroup(0, 8, MTRFixed); err != nil {
		t.Fatal(err)
	}
	before, _ := tree.FormatGroups()
	if _, err := tree.MakeGroup(8, 8, MTRFixed); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	after, _ := tree.FormatGroups()
	if after != before {
		t.Fatal("limit mutation")
	}
	stops := 0
	cancel, _ := NewMTRGroupTree(0, 16, MTRLimits{Stop: func() bool { stops++; return stops > 1 }})
	if _, err := cancel.MakeGroup(0, 8, MTRFixed); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if cancel.NodeCount() != 1 {
		t.Fatal("cancel mutation")
	}
	if _, err := tree.Clone(0, MTRLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := tree.Clone(1, MTRLimits{MaxNodes: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	tree.Close()
	if err := tree.Validate(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestMTRMalformedStructuresAndTerminalTraversal(t *testing.T) {
	newTree := func() *MTRTree {
		tree, err := NewMTRGroupTree(0, 8, MTRLimits{})
		if err != nil {
			t.Fatal(err)
		}
		first, err := tree.MakeGroup(0, 4, MTRDefault)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = tree.MakeGroup(4, 4, MTRDefault); err != nil {
			t.Fatal(err)
		}
		if _, err = tree.MakeGroup(1, 2, MTRDefault); err != nil {
			t.Fatal(err)
		}
		if first.child == nil {
			t.Fatal("missing nested group")
		}
		return tree
	}
	breakers := []struct {
		name      string
		breakTree func(*MTRTree)
	}{
		{"parent", func(tree *MTRTree) { tree.root.child.parent = nil }},
		{"elder", func(tree *MTRTree) { tree.root.child.younger.elder = nil }},
		{"cycle", func(tree *MTRTree) { tree.root.child.younger = tree.root.child }},
		{"owner", func(tree *MTRTree) { tree.root.child.owner = nil }},
		{"flags", func(tree *MTRTree) { tree.root.child.Flags = 1 << 31 }},
		{"order", func(tree *MTRTree) { tree.root.child.younger.Low = 2 }},
		{"outside", func(tree *MTRTree) { tree.root.child.Size = 9 }},
	}
	for _, test := range breakers {
		t.Run(test.name, func(t *testing.T) {
			tree := newTree()
			test.breakTree(tree)
			if err := tree.Validate(); !errors.Is(err, ErrInvariant) {
				t.Fatalf("Validate() = %v", err)
			}
		})
	}

	tree := newTree()
	terminal := tree.root.child
	terminal.Flags = MTRTerminal
	text, err := tree.FormatTree()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "N=2 C=3") || strings.Contains(text, "N=3 C=") {
		t.Fatalf("terminal traversal did not retain the hidden child relation: %q", text)
	}
}

type mtrInterval struct{ low, size uint32 }

func mtrModelAccept(groups []mtrInterval, candidate mtrInterval, root mtrInterval) bool {
	if candidate.size == 0 || uint64(candidate.low) < uint64(root.low) || uint64(candidate.low)+uint64(candidate.size) > uint64(root.low)+uint64(root.size) {
		return false
	}
	a0, a1 := uint64(candidate.low), uint64(candidate.low)+uint64(candidate.size)
	for _, group := range groups {
		b0, b1 := uint64(group.low), uint64(group.low)+uint64(group.size)
		if a0 == b0 && a1 == b1 {
			return true
		}
		if a1 <= b0 || b1 <= a0 || a0 <= b0 && b1 <= a1 || b0 <= a0 && a1 <= b1 {
			continue
		}
		return false
	}
	return true
}

func TestMTRIndependentIntervalModelCampaign(t *testing.T) {
	state := uint64(0x9e3779b97f4a7c15)
	for execution := 0; execution < 5000; execution++ {
		tree, _ := NewMTRGroupTree(0, 32, MTRLimits{MaxNodes: 128})
		groups := []mtrInterval{{0, 32}}
		for step := 0; step < 24; step++ {
			state ^= state << 7
			state ^= state >> 9
			state ^= state << 8
			candidate := mtrInterval{uint32(state % 36), uint32((state >> 8) % 12)}
			want := mtrModelAccept(groups, candidate, groups[0])
			node, err := tree.MakeGroup(candidate.low, candidate.size, MTRFlags((state>>20)&15))
			if (err == nil) != want {
				t.Fatalf("execution %d step %d candidate=%+v want=%v err=%v", execution, step, candidate, want, err)
			}
			if want {
				found := false
				for _, g := range groups {
					if g == candidate {
						found = true
					}
				}
				if !found {
					groups = append(groups, candidate)
				}
				if tree.FindGroup(candidate.low, candidate.size) != node {
					t.Fatal("find mismatch")
				}
			}
			if err = tree.Validate(); err != nil {
				t.Fatalf("execution %d step %d: %v", execution, step, err)
			}
		}
	}
}

func FuzzMTRGroupTree(f *testing.F) {
	f.Add([]byte{0, 4, 2, 2, 8, 1})
	f.Add([]byte{2, 4, 3, 1, 6, 2, 0, 8, 4})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 256 {
			return
		}
		tree, _ := NewMTRGroupTree(0, 32, MTRLimits{MaxNodes: 260})
		groups := []mtrInterval{{0, 32}}
		for i := 0; i+2 < len(data); i += 3 {
			candidate := mtrInterval{uint32(data[i] % 36), uint32(data[i+1] % 12)}
			want := mtrModelAccept(groups, candidate, groups[0])
			_, err := tree.MakeGroup(candidate.low, candidate.size, MTRFlags(data[i+2]&15))
			if (err == nil) != want {
				t.Fatalf("%v want=%v err=%v data=%v", candidate, want, err, data)
			}
			if want && !slices.Contains(groups, candidate) {
				groups = append(groups, candidate)
			}
			if err = tree.Validate(); err != nil {
				t.Fatal(err)
			}
		}
		text, err := tree.FormatGroups()
		if err != nil || !strings.HasSuffix(text, "\n") {
			t.Fatal(text, err)
		}
	})
}
