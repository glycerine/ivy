package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func checkGiaOneHotAssignment(t *testing.T, g *GiaManager, skips, variables, assignment int) {
	t.Helper()
	inputs := make([]bool, skips+variables)
	count := 0
	for i := range inputs {
		inputs[i] = assignment&(1<<i) != 0
		if i >= skips && inputs[i] {
			count++
		}
	}
	outputs, err := g.Evaluate(inputs)
	if err != nil {
		t.Fatal(err)
	}
	violation := false
	for _, value := range outputs {
		violation = violation || value
	}
	if violation != (count != 1) || outputs[len(outputs)-1] != (count == 0) {
		t.Fatalf("assignment %d count %d outputs %v", assignment, count, outputs)
	}
}

func TestGiaOneHotNativeAndExhaustive(t *testing.T) {
	raw, err := os.ReadFile("test_vectors/oracle/gia_onehot.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Objects int
		Drivers []GiaLit
	}
	if err = json.Unmarshal(raw, &rows); err != nil || len(rows) != 7 {
		t.Fatal(err, len(rows))
	}
	for variables := 1; variables <= 7; variables++ {
		g, err := NewGiaOneHot(2, variables, GiaCofactorOptions{})
		if err != nil {
			t.Fatal(err)
		}
		drivers := make([]GiaLit, 0, g.OutputCount())
		for _, id := range g.cos {
			driver, err := g.objects[id].Fanin0(id)
			if err != nil {
				t.Fatal(err)
			}
			drivers = append(drivers, driver)
		}
		row := rows[variables-1]
		if g.ObjectCount() != row.Objects || !slices.Equal(drivers, row.Drivers) {
			t.Fatal(variables, g.ObjectCount(), drivers, row)
		}
		if g.InputCount() != variables+2 || g.OutputCount() != variables || g.Name != "onehot" || g.RegisterCount() != 0 || g.ConstraintCount() != 0 {
			t.Fatal("metadata", variables)
		}
		for assignment := 0; assignment < 1<<(variables+2); assignment++ {
			checkGiaOneHotAssignment(t, g, 2, variables, assignment)
		}
		exact, err := NewGiaOneHot(2, variables, GiaCofactorOptions{MaxObjects: row.Objects})
		if err != nil || exact == nil {
			t.Fatal("exact capacity", err)
		}
		failed, err := NewGiaOneHot(2, variables, GiaCofactorOptions{MaxObjects: row.Objects - 1})
		if failed != nil || !errors.Is(err, ErrLimit) {
			t.Fatal("capacity", err)
		}
	}
}

func TestGiaOneHotLimitsAndCancellation(t *testing.T) {
	for _, dims := range [][2]int{{-1, 2}, {0, 0}, {0, -1}} {
		g, err := NewGiaOneHot(dims[0], dims[1], GiaCofactorOptions{})
		if g != nil || !errors.Is(err, ErrDimension) {
			t.Fatal(dims, err)
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: -1}, {MaxMappingEntries: -1}, {MaxObjects: 1}, {MaxMappingEntries: 4}} {
		g, err := NewGiaOneHot(2, 5, options)
		if g != nil || err == nil {
			t.Fatal(options, err)
		}
	}
	if _, err := NewGiaOneHot(2, 5, GiaCofactorOptions{MaxMappingEntries: 8}); err != nil {
		t.Fatal(err)
	}
	for _, threshold := range []int{1, 3, 8, 12} {
		calls := 0
		g, err := NewGiaOneHot(2, 5, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if g != nil || !errors.Is(err, ErrCancelled) {
			t.Fatal(threshold, calls, err)
		}
	}
}

func FuzzGiaOneHot(f *testing.F) {
	f.Add(uint8(5), uint8(2), uint16(42))
	f.Add(uint8(0), uint8(0), uint16(0))
	f.Fuzz(func(t *testing.T, size, ignored uint8, assignment uint16) {
		variables, skips := 1+int(size%12), int(ignored%4)
		g, err := NewGiaOneHot(skips, variables, GiaCofactorOptions{MaxObjects: 128, MaxMappingEntries: 16})
		if err != nil {
			t.Fatal(err)
		}
		checkGiaOneHotAssignment(t, g, skips, variables, int(assignment))
	})
}
