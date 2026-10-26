package abc

// Tests for structural hashing from aig/aig/aigTable.c.

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"slices"
	"testing"
)

type aigTableOracleRow struct {
	Seed, Entries                                        int
	Existing, Missing, Same, Opposite, WithOne, WithZero int
	BeforeSize, AfterSize, ExistingAfter                 int
	BeforeBuckets, AfterBuckets                          [][2]int
}

func aigTableGraph(t testing.TB, seed int) (*AigManager, []AigLit) {
	t.Helper()
	graph := NewAigManager()
	v := aigObjectVariables(t, graph, 6)
	n0 := aigObjectAnd(t, graph, v[0].NotIf(seed&1 != 0), v[1].NotIf(seed&2 != 0))
	n1 := aigObjectAnd(t, graph, n0, v[2].NotIf(seed&4 != 0))
	n2 := aigObjectAnd(t, graph, v[3].NotIf(seed&8 != 0), v[4].NotIf(seed&16 != 0))
	n3 := aigObjectAnd(t, graph, n1, n2)
	if _, err := graph.CreateCO(n3); err != nil {
		t.Fatal(err)
	}
	return graph, v
}

func aigHashNonempty(profile AigHashProfile) [][2]int {
	result := make([][2]int, 0, profile.Entries)
	for bucket, count := range profile.BucketOccupancies {
		if count != 0 {
			result = append(result, [2]int{bucket, count})
		}
	}
	return result
}

func aigLookupValue(result AigLit, found bool) int {
	if !found {
		return -1
	}
	return int(result)
}

func TestAigTableNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/aig_table.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []aigTableOracleRow
	if err = json.Unmarshal(data, &rows); err != nil || len(rows) != 32 {
		t.Fatal(len(rows), err)
	}
	for _, row := range rows {
		graph, v := aigTableGraph(t, row.Seed)
		before := graph.Clone()
		lookup := func(left, right AigLit) int {
			result, found, err := graph.LookupHashedAnd(left, right)
			if err != nil {
				t.Fatal(row.Seed, err)
			}
			return aigLookupValue(result, found)
		}
		got := []int{
			lookup(v[0].NotIf(row.Seed&1 != 0), v[1].NotIf(row.Seed&2 != 0)),
			lookup(v[0], v[5]), lookup(v[2], v[2]), lookup(v[2], v[2].Not()),
			lookup(AigConst1, v[3]), lookup(AigConst0, v[3]),
		}
		want := []int{row.Existing, row.Missing, row.Same, row.Opposite, row.WithOne, row.WithZero}
		if !slices.Equal(got, want) || !reflect.DeepEqual(graph, before) || graph.AndCount() != row.Entries {
			t.Fatal(row.Seed, got, want, graph.AndCount(), row.Entries, reflect.DeepEqual(graph, before))
		}
		profileBefore, err := graph.StructuralHashProfile(row.BeforeSize, AigHashOptions{})
		if err != nil || profileBefore.Entries != row.Entries || !slices.Equal(aigHashNonempty(profileBefore), row.BeforeBuckets) {
			t.Fatal(row.Seed, err, profileBefore, row.BeforeBuckets)
		}
		profileAfter, err := graph.StructuralHashProfile(row.AfterSize, AigHashOptions{})
		if err != nil || profileAfter.Entries != row.Entries || !slices.Equal(aigHashNonempty(profileAfter), row.AfterBuckets) || lookup(v[0].NotIf(row.Seed&1 != 0), v[1].NotIf(row.Seed&2 != 0)) != row.ExistingAfter {
			t.Fatal(row.Seed, err, profileAfter, row.AfterBuckets, row.ExistingAfter)
		}
	}
}

func TestAigTableErrorsDeletionAndOwnership(t *testing.T) {
	graph, v := aigTableGraph(t, 0)
	for _, test := range []struct {
		graph   *AigManager
		size    int
		options AigHashOptions
		want    error
	}{
		{nil, 11, AigHashOptions{}, ErrDimension},
		{graph, 0, AigHashOptions{}, ErrDimension},
		{graph, -1, AigHashOptions{}, ErrDimension},
		{graph, 11, AigHashOptions{MaxObjects: -1}, ErrDimension},
		{graph, 11, AigHashOptions{MaxObjects: graph.ObjectCount() - 1}, ErrLimit},
		{graph, 11, AigHashOptions{Stop: func() bool { return true }}, ErrCancelled},
	} {
		if result, err := test.graph.StructuralHashProfile(test.size, test.options); result.BucketOccupancies != nil || !errors.Is(err, test.want) {
			t.Fatal(result, err, test.want)
		}
	}
	if result, found, err := graph.LookupHashedAnd(AigLit(graph.ObjectCount()+1)<<1, v[0]); result != InvalidAigLit || found || !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(result, found, err)
	}
	profile, err := graph.StructuralHashProfile(11, AigHashOptions{})
	if err != nil {
		t.Fatal(err)
	}
	profile.BucketOccupancies[0] = 99
	again, _ := graph.StructuralHashProfile(11, AigHashOptions{})
	if again.BucketOccupancies[0] == 99 {
		t.Fatal("profile aliases graph state")
	}

	dangling := aigObjectAnd(t, graph, v[0], v[5])
	if graph.AndCount() != 5 {
		t.Fatal(graph.AndCount())
	}
	if _, err = graph.DeleteCone(dangling.ID(), AigObjectOptions{}); err != nil || graph.AndCount() != 4 || graph.Check() != nil {
		t.Fatal(err, graph.AndCount(), graph.Check())
	}
	if result, found, err := graph.LookupHashedAnd(v[0], v[5]); result != InvalidAigLit || found || err != nil {
		t.Fatal(result, found, err)
	}
}

func FuzzAigTable(f *testing.F) {
	for _, seed := range []uint64{0, 1, 2, 3, 0x123456789abcdef0, ^uint64(0)} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		graph := NewAigManager()
		count := 2 + int(seed%9)
		v := aigObjectVariables(t, graph, count)
		pairs := make([][2]AigLit, 0, count*3)
		for index := 0; index < count*3; index++ {
			left := v[index%count].NotIf(seed&(1<<uint(index%63)) != 0)
			right := v[(index+1+int(seed%uint64(count-1)))%count].NotIf(seed&(1<<uint((index+13)%63)) != 0)
			result := aigObjectAnd(t, graph, left, right)
			if result.ID() != 0 && result.ID() >= AigID(count+1) {
				pairs = append(pairs, [2]AigLit{left, right})
			}
		}
		size := 1 + int((seed>>32)%97)
		profile, err := graph.StructuralHashProfile(size, AigHashOptions{})
		if err != nil || profile.Entries != graph.AndCount() {
			t.Fatal(seed, err, profile.Entries, graph.AndCount())
		}
		total := 0
		for _, count := range profile.BucketOccupancies {
			total += count
		}
		if total != graph.AndCount() {
			t.Fatal(seed, total, graph.AndCount())
		}
		for _, pair := range pairs {
			result, found, err := graph.LookupHashedAnd(pair[0], pair[1])
			if err != nil || !found || result == InvalidAigLit {
				t.Fatal(seed, pair, result, found, err)
			}
		}
	})
}
