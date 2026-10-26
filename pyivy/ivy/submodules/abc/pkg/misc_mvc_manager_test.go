package abc

// Tests for src/misc/mvc/mvcMan.c.

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
)

type mvcManagerOracle struct {
	Cases, AllNonNull, Distinct int
}

func TestMvcManagerNativeLifecycle(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/mvc_manager.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle mvcManagerOracle
	if err = json.Unmarshal(data, &oracle); err != nil || oracle.Cases != 32 || oracle.AllNonNull != 1 || oracle.Distinct != 1 {
		t.Fatal(oracle, err)
	}
	managers := make([]*MvcManager, oracle.Cases)
	for index := range managers {
		managers[index] = NewMvcManager()
		if managers[index] == nil || managers[index].check() != nil {
			t.Fatal(index)
		}
		for previous := 0; previous < index; previous++ {
			if managers[index] == managers[previous] {
				t.Fatal("aliased manager", index, previous)
			}
		}
	}
	for index, manager := range managers {
		if err := manager.Close(); err != nil {
			t.Fatal(index, err)
		}
		if err := manager.check(); !errors.Is(err, ErrInvariant) {
			t.Fatal(index, err)
		}
	}
	if err := (*MvcManager)(nil).Close(); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	manager := NewMvcManager()
	if err := manager.Close(); err != nil {
		t.Fatal(err)
	}
	if err := manager.Close(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func FuzzMvcManagerLifecycle(f *testing.F) {
	f.Add(uint8(0), uint8(0))
	f.Add(uint8(17), uint8(5))
	f.Fuzz(func(t *testing.T, rawCount, rawEarly uint8) {
		count := int(rawCount)%64 + 1
		early := int(rawEarly) % count
		managers := make([]*MvcManager, count)
		for index := range managers {
			managers[index] = NewMvcManager()
		}
		if err := managers[early].Close(); err != nil {
			t.Fatal(err)
		}
		for index, manager := range managers {
			if index == early {
				if err := manager.Close(); !errors.Is(err, ErrInvariant) {
					t.Fatal(index, err)
				}
				continue
			}
			if manager.check() != nil {
				t.Fatal(index)
			}
			if err := manager.Close(); err != nil {
				t.Fatal(index, err)
			}
		}
	})
}
