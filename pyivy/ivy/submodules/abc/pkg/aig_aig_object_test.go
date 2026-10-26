package abc

// Tests for the stable-ID-safe object mutations from aig/aig/aigObj.c.

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"slices"
	"testing"
)

type aigObjectOracleRow struct {
	Seed                                 int
	DeleteRoot, PatchOutput, PatchDriver int
	DeletedOutput                        int
	DeleteObjects, PatchObjects          []aigOperationObject
	OutputObjects                        []aigOperationObject
}

func aigObjectVariables(t testing.TB, graph *AigManager, count int) []AigLit {
	t.Helper()
	result := make([]AigLit, count)
	for index := range result {
		var err error
		result[index], err = graph.CreateCI()
		if err != nil {
			t.Fatal(err)
		}
	}
	return result
}

func aigObjectAnd(t testing.TB, graph *AigManager, left, right AigLit) AigLit {
	t.Helper()
	result, err := graph.And(left, right)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func aigObjectDeletionGraph(t testing.TB, seed int) (*AigManager, AigID) {
	t.Helper()
	graph := NewAigManager()
	v := aigObjectVariables(t, graph, 5)
	base := aigObjectAnd(t, graph, v[0].NotIf(seed&1 != 0), v[1].NotIf(seed&2 != 0))
	shared := aigObjectAnd(t, graph, base, v[2].NotIf(seed&4 != 0))
	if _, err := graph.CreateCO(shared); err != nil {
		t.Fatal(err)
	}
	x := aigObjectAnd(t, graph, v[3].NotIf(seed&8 != 0), v[4].NotIf(seed&16 != 0))
	y := aigObjectAnd(t, graph, x, v[0].Not())
	root := aigObjectAnd(t, graph, y, v[2].Not())
	return graph, root.ID()
}

func aigObjectPatchGraph(t testing.TB, seed int) (*AigManager, AigID, AigLit) {
	t.Helper()
	graph := NewAigManager()
	v := aigObjectVariables(t, graph, 5)
	old0 := aigObjectAnd(t, graph, v[0].NotIf(seed&1 != 0), v[1].NotIf(seed&2 != 0))
	old1 := aigObjectAnd(t, graph, old0, v[2].NotIf(seed&4 != 0))
	replacement := aigObjectAnd(t, graph, v[3].NotIf(seed&8 != 0), v[4].NotIf(seed&16 != 0))
	output, err := graph.CreateCO(old1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = graph.CreateCO(replacement); err != nil {
		t.Fatal(err)
	}
	return graph, output, replacement.NotIf(seed&1 != 0)
}

func aigObjectOutputGraph(t testing.TB, seed int) (*AigManager, AigID) {
	t.Helper()
	graph := NewAigManager()
	v := aigObjectVariables(t, graph, 3)
	node := aigObjectAnd(t, graph, v[0].NotIf(seed&1 != 0), v[1].NotIf(seed&2 != 0))
	first, err := graph.CreateCO(node.NotIf(seed&4 != 0))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = graph.CreateCO(v[2].NotIf(seed&8 != 0)); err != nil {
		t.Fatal(err)
	}
	return graph, first
}

func TestAigObjectNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/aig_object.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []aigObjectOracleRow
	if err = json.Unmarshal(data, &rows); err != nil || len(rows) != 32 {
		t.Fatal(len(rows), err)
	}
	for _, row := range rows {
		deleted, root := aigObjectDeletionGraph(t, row.Seed)
		if int(root) != row.DeleteRoot {
			t.Fatal(row.Seed, root, row.DeleteRoot)
		}
		count, err := deleted.DeleteCone(root, AigObjectOptions{})
		if err != nil || count != 3 || !slices.Equal(aigOperationObjects(deleted), row.DeleteObjects) {
			t.Fatal(row.Seed, count, err, aigOperationObjects(deleted), row.DeleteObjects)
		}

		patched, output, driver := aigObjectPatchGraph(t, row.Seed)
		if int(output) != row.PatchOutput || int(driver) != row.PatchDriver {
			t.Fatal(row.Seed, output, driver, row.PatchOutput, row.PatchDriver)
		}
		count, err = patched.PatchOutput(output, driver, AigObjectOptions{})
		if err != nil || count != 2 || !slices.Equal(aigOperationObjects(patched), row.PatchObjects) {
			t.Fatal(row.Seed, count, err, aigOperationObjects(patched), row.PatchObjects)
		}

		outputs, deletedOutput := aigObjectOutputGraph(t, row.Seed)
		if int(deletedOutput) != row.DeletedOutput {
			t.Fatal(row.Seed, deletedOutput, row.DeletedOutput)
		}
		if err = outputs.DeleteOutput(deletedOutput); err != nil || !slices.Equal(aigOperationObjects(outputs), row.OutputObjects) {
			t.Fatal(row.Seed, err, aigOperationObjects(outputs), row.OutputObjects)
		}
		if outputs.Check() != nil || !slices.Equal(outputs.Outputs(), []AigID{6}) {
			t.Fatal(row.Seed, outputs.Outputs(), outputs.Check())
		}
	}
}

func TestAigObjectErrorsAtomicAndChoices(t *testing.T) {
	graph := NewAigManager()
	v := aigObjectVariables(t, graph, 4)
	left := aigObjectAnd(t, graph, v[0], v[1])
	right := aigObjectAnd(t, graph, v[2], v[3])
	root := aigObjectAnd(t, graph, left, right)
	output, err := graph.CreateCO(root)
	if err != nil {
		t.Fatal(err)
	}
	invalid := AigID(graph.ObjectCount() + 4)
	tests := []struct {
		name    string
		apply   func() error
		wantErr error
	}{
		{"nil delete", func() error { _, err := (*AigManager)(nil).DeleteCone(root.ID(), AigObjectOptions{}); return err }, ErrDimension},
		{"invalid delete", func() error { _, err := graph.DeleteCone(invalid, AigObjectOptions{}); return err }, ErrInvalidLiteral},
		{"terminal delete", func() error { _, err := graph.DeleteCone(v[0].ID(), AigObjectOptions{}); return err }, ErrUnsupported},
		{"referenced delete", func() error { _, err := graph.DeleteCone(root.ID(), AigObjectOptions{}); return err }, ErrInvariant},
		{"negative limit", func() error { _, err := graph.PatchOutput(output, v[0], AigObjectOptions{MaxObjects: -1}); return err }, ErrDimension},
		{"small limit", func() error { _, err := graph.PatchOutput(output, v[0], AigObjectOptions{MaxObjects: 1}); return err }, ErrLimit},
		{"cancelled", func() error {
			_, err := graph.PatchOutput(output, v[0], AigObjectOptions{Stop: func() bool { return true }})
			return err
		}, ErrCancelled},
		{"invalid output", func() error { _, err := graph.PatchOutput(v[0].ID(), v[1], AigObjectOptions{}); return err }, ErrInvalidLiteral},
		{"invalid driver", func() error { _, err := graph.PatchOutput(output, AigLit(invalid)<<1, AigObjectOptions{}); return err }, ErrInvalidLiteral},
		{"delete missing output", func() error { return graph.DeleteOutput(v[0].ID()) }, ErrInvalidLiteral},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			before := graph.Clone()
			if err := test.apply(); !errors.Is(err, test.wantErr) || !reflect.DeepEqual(graph, before) {
				t.Fatal(err, test.wantErr, reflect.DeepEqual(graph, before))
			}
		})
	}

	choices := NewAigManager()
	cv := aigObjectVariables(t, choices, 4)
	choice0 := aigObjectAnd(t, choices, cv[0], cv[1])
	choice1 := aigObjectAnd(t, choices, cv[2], cv[3])
	if err = choices.SetChoiceChains([][]AigID{{choice0.ID(), choice1.ID()}}, AigTraversalOptions{}); err != nil {
		t.Fatal(err)
	}
	before := choices.Clone()
	if _, err = choices.DeleteCone(choice0.ID(), AigObjectOptions{}); !errors.Is(err, ErrInvariant) || !reflect.DeepEqual(choices, before) {
		t.Fatal(err)
	}
}

func TestAigObjectPatchDescendantAndPartitions(t *testing.T) {
	graph := NewAigManager()
	v := aigObjectVariables(t, graph, 3)
	child := aigObjectAnd(t, graph, v[0], v[1])
	root := aigObjectAnd(t, graph, child, v[2])
	output, _ := graph.CreateCO(root)
	deleted, err := graph.PatchOutput(output, child.Not(), AigObjectOptions{})
	if err != nil || deleted != 1 || graph.objects[root.ID()].Type != AigNone || graph.objects[child.ID()].Type != AigAND || graph.objects[child.ID()].RefCount != 1 || graph.Check() != nil {
		t.Fatal(deleted, err, graph.objects[root.ID()], graph.objects[child.ID()], graph.Check())
	}

	partitioned := NewAigManager()
	pv := aigObjectVariables(t, partitioned, 4)
	var outputs [4]AigID
	for index := range outputs {
		outputs[index], _ = partitioned.CreateCO(pv[index])
	}
	if err = partitioned.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	if err = partitioned.SetConstraintCount(1); err != nil {
		t.Fatal(err)
	}
	if err = partitioned.DeleteOutput(outputs[2]); err != nil || partitioned.ConstraintCount() != 0 || partitioned.RegisterCount() != 1 {
		t.Fatal(err, partitioned.ConstraintCount(), partitioned.RegisterCount())
	}
	if err = partitioned.DeleteOutput(outputs[3]); err != nil || partitioned.RegisterCount() != 0 {
		t.Fatal(err, partitioned.RegisterCount())
	}
	if err = partitioned.DeleteOutput(outputs[0]); err != nil || !slices.Equal(partitioned.Outputs(), []AigID{outputs[1]}) || partitioned.Check() != nil {
		t.Fatal(err, partitioned.Outputs(), partitioned.Check())
	}
}

func TestAigObjectFormatAndDeepDeletion(t *testing.T) {
	graph := NewAigManager()
	v := aigObjectVariables(t, graph, 2)
	and := aigObjectAnd(t, graph, v[0].Not(), v[1])
	buffer, err := graph.CreateBuffer(and.Not())
	if err != nil {
		t.Fatal(err)
	}
	output, err := graph.CreateCO(buffer)
	if err != nil {
		t.Fatal(err)
	}
	for literal, want := range map[AigLit]string{
		AigConst1:           "Node    0 : constant 1 (refs =   0)",
		v[0].Not():          "Compl Node    1 : PI (refs =   1)",
		and:                 "Node    3 : AND(    1',    2  ) (refs =   1)",
		buffer:              "Node    4 : BUF( 3' ) (refs =   1)",
		AigLit(output) << 1: "Node    5 : PO(    4  ) (refs =   0)",
	} {
		got, err := graph.FormatObject(literal)
		if err != nil || got != want {
			t.Fatal(literal, got, want, err)
		}
	}
	if _, err = graph.FormatObject(InvalidAigLit); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(err)
	}

	const depth = 10000
	deep := NewAigManager()
	variables := aigObjectVariables(t, deep, depth)
	root := variables[0]
	for index := 1; index < depth; index++ {
		root = aigObjectAnd(t, deep, root, variables[index])
	}
	deleted, err := deep.DeleteCone(root.ID(), AigObjectOptions{MaxObjects: depth - 1})
	if err != nil || deleted != depth-1 || deep.AndCount() != 0 || deep.Check() != nil {
		t.Fatal(deleted, err, deep.AndCount(), deep.Check())
	}
}

func FuzzAigObjectMutations(f *testing.F) {
	for _, seed := range []uint64{0, 1, 2, 3, 0x123456789abcdef0, ^uint64(0)} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		graph := NewAigManager()
		count := 3 + int(seed%8)
		v := aigObjectVariables(t, graph, count)
		root := v[0]
		for index := 1; index < count; index++ {
			root = aigObjectAnd(t, graph, root.NotIf(seed&(1<<uint(index%63)) != 0), v[index])
		}
		output, err := graph.CreateCO(root)
		if err != nil {
			t.Fatal(err)
		}
		replacementIndex := int((seed >> 8) % uint64(count))
		replacement := v[replacementIndex].NotIf(seed>>16&1 != 0)
		if _, err = graph.PatchOutput(output, replacement, AigObjectOptions{MaxObjects: count}); err != nil || graph.Check() != nil {
			t.Fatal(seed, err, graph.Check())
		}
		for sample := 0; sample < 8; sample++ {
			inputs := make([]bool, count)
			for index := range inputs {
				inputs[index] = (seed+uint64(sample*17))&(1<<uint(index%63)) != 0
			}
			values, err := graph.Evaluate(inputs)
			want := inputs[replacementIndex] != replacement.Complemented()
			if err != nil || len(values) != 1 || values[0] != want {
				t.Fatal(seed, sample, values, want, err)
			}
		}

		dangling := NewAigManager()
		dv := aigObjectVariables(t, dangling, count)
		danglingRoot := dv[0]
		for index := 1; index < count; index++ {
			danglingRoot = aigObjectAnd(t, dangling, danglingRoot, dv[index].NotIf(seed&(1<<uint(index%63)) != 0))
		}
		deleted, err := dangling.DeleteCone(danglingRoot.ID(), AigObjectOptions{MaxObjects: count})
		if err != nil || deleted != count-1 || dangling.AndCount() != 0 || dangling.Check() != nil {
			t.Fatal(seed, deleted, err, dangling.AndCount(), dangling.Check())
		}
	})
}
