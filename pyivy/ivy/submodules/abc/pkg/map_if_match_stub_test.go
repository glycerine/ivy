package abc

// Tests for specialized matching stubs from map/if/ifMatch2.c.

import (
	"encoding/json"
	"os"
	"testing"
)

type batMatchOracleRow struct {
	Manager     uint64   `json:"manager"`
	Truth       []uint32 `json:"truth"`
	Variables   int      `json:"variables"`
	Leaves      int      `json:"leaves"`
	Description string   `json:"description"`
	Result      int      `json:"result"`
}

func TestBatMatchNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/if_match_stub.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []batMatchOracleRow
	if err = json.Unmarshal(data, &rows); err != nil || len(rows) != 16 {
		t.Fatal(len(rows), err)
	}
	BatFunctionTableSetup()
	for _, row := range rows {
		if got := BatCellFunctionLookup(uintptr(row.Manager), row.Truth, row.Variables, row.Leaves, row.Description); got != row.Result {
			t.Fatal(row, got)
		}
	}
	BatFunctionTableSetdown()
}

func TestBatMatchIgnoresMutableArguments(t *testing.T) {
	truth := []uint32{0, 1, ^uint32(0), 0x5a5a5a5a}
	want := append([]uint32(nil), truth...)
	if got := BatCellFunctionLookup(^uintptr(0), truth, -9, -17, "\x00binary\xff"); got != 1 {
		t.Fatal(got)
	}
	for i := range truth {
		if truth[i] != want[i] {
			t.Fatal(i, truth, want)
		}
	}
}

func FuzzBatCellFunctionLookup(f *testing.F) {
	f.Add(uint64(0), uint32(0), 0, 0, "")
	f.Add(^uint64(0), ^uint32(0), -1, -1, "native stub")
	f.Fuzz(func(t *testing.T, manager uint64, word uint32, variables, leaves int, description string) {
		truth := []uint32{word, ^word}
		before := truth[0]
		if BatCellFunctionLookup(uintptr(manager), truth, variables, leaves, description) != 1 {
			t.Fatal(manager, word, variables, leaves)
		}
		if truth[0] != before || truth[1] != ^before {
			t.Fatal(truth, before)
		}
	})
}
