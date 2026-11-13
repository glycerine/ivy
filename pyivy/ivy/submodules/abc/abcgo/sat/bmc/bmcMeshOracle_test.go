package bmc

import (
	"io"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

type meshCell struct {
	node, time int
	set        bool
}

// captureMeshOutput is deliberately used serially: the legacy APIs print the
// placement instead of returning it, and replace process-wide stdout here.
func captureMeshOutput(t *testing.T, run func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old; r.Close() }()
	run()
	if err = w.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// checkMeshWitness checks the printed solution independently of the SAT
// encoding. Every placed gate must find both fanins at time t-1 on distinct
// adjacent cells; any input in the interior must be copied from a neighbor.
func checkMeshWitness(t *testing.T, output string, xsize, ysize, inputs int, fanins [][2]int) {
	t.Helper()
	lines := strings.Split(output, "\n")
	start := -1
	for i, line := range lines {
		if strings.Contains(line, "Y\\X") {
			start = i + 1
			break
		}
	}
	if start < 0 || start+ysize > len(lines) {
		t.Fatalf("missing placement grid:\n%s", output)
	}
	grid := make([][]meshCell, ysize)
	for y := 0; y < ysize; y++ {
		line := lines[start+y]
		if len(line) < 5+5*xsize {
			t.Fatalf("short grid row %q", line)
		}
		grid[y] = make([]meshCell, xsize)
		for x := 0; x < xsize; x++ {
			field := strings.TrimSpace(line[5+5*x : 5+5*(x+1)])
			if field == "" || field == "*" {
				continue
			}
			if len(field) < 2 {
				t.Fatalf("bad grid field %q", field)
			}
			tm, err := strconv.Atoi(field[1:])
			if err != nil {
				t.Fatal(err)
			}
			grid[y][x] = meshCell{node: int(field[0] - 'a'), time: tm, set: true}
		}
	}
	root := grid[1][1]
	if !root.set || root.node != inputs+len(fanins)-1 {
		t.Fatalf("incorrect root %+v\n%s", root, output)
	}
	for y := 0; y < ysize; y++ {
		for x := 0; x < xsize; x++ {
			cell := grid[y][x]
			if !cell.set {
				continue
			}
			if cell.node < 0 || cell.node >= inputs+len(fanins) {
				t.Fatalf("invalid node %+v", cell)
			}
			if x == 0 || y == 0 || x == xsize-1 || y == ysize-1 {
				if cell.time != 0 || cell.node >= inputs {
					t.Fatalf("bad boundary cell %+v", cell)
				}
				continue
			}
			if cell.time <= 0 {
				t.Fatalf("bad interior time %+v", cell)
			}
			neighbors := []meshCell{grid[y][x-1], grid[y-1][x], grid[y][x+1], grid[y+1][x]}
			has := func(node int, except int) bool {
				for j, n := range neighbors {
					if j != except && n.set && n.node == node && n.time == cell.time-1 {
						return true
					}
				}
				return false
			}
			if has(cell.node, -1) {
				// Every computed signal, including an internal gate output,
				// can be forwarded through a buffer cell.
				continue
			}
			if cell.node < inputs {
				t.Fatalf("invalid buffer at %d,%d: %+v\n%s", x, y, cell, output)
			}
			pair := fanins[cell.node-inputs]
			valid := false
			for j, n := range neighbors {
				if n.set && n.node == pair[0] && n.time == cell.time-1 && has(pair[1], j) {
					valid = true
				}
			}
			if !valid {
				t.Fatalf("missing gate fanins at %d,%d: %+v, want %v\n%s", x, y, cell, pair, output)
			}
		}
	}
}

func TestMeshBackendsAgainstSmallPlacementOracle(t *testing.T) {
	backends := []struct {
		name string
		run  func(*gia.Gia_Man_t, int32, int32, int32, int32)
	}{{"satoko", Bmc_MeshTest}, {"bsat", Bmc_MeshTest2}}
	for _, backend := range backends {
		for _, tc := range []struct {
			name                        string
			x, y, latency               int32
			chain, complement, possible bool
		}{
			{"one_gate", 3, 3, 2, false, false, true},
			{"complemented_inputs", 3, 3, 2, false, true, true},
			{"chain_too_small", 3, 3, 2, true, false, false},
			{"chain_no_buffer_space", 4, 3, 3, true, false, false},
			{"chain_fits", 4, 4, 3, true, false, true},
		} {
			t.Run(backend.name+"/"+tc.name, func(t *testing.T) {
				p := gia.Gia_ManStart(12)
				a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
				if tc.complement {
					a ^= 1
					b ^= 1
				}
				first := gia.Gia_ManAppendAnd(p, a, b)
				root := first
				fanins := [][2]int{{0, 1}}
				if tc.chain {
					root = gia.Gia_ManAppendAnd(p, first, a)
					fanins = append(fanins, [2]int{2, 0})
				}
				gia.Gia_ManAppendCo(p, root)
				output := captureMeshOutput(t, func() { backend.run(p, tc.x, tc.y, tc.latency, 0) })
				if tc.possible {
					if !strings.Contains(output, "Satisfying solution found") {
						t.Fatalf("expected solution:\n%s", output)
					}
					checkMeshWitness(t, output, int(tc.x), int(tc.y), 2, fanins)
				} else if !strings.Contains(output, "Problem has no solution") {
					t.Fatalf("expected unsatisfiable placement:\n%s", output)
				}
			})
		}
	}
}
