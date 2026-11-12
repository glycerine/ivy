package espresso

import (
	"math/rand/v2"
	"strings"
	"testing"
)

func TestCoverCostLiteralEnumeration(t *testing.T) {
	savedCube, savedData := cube, cdata
	defer func() { cube, cdata = savedCube, savedData }()
	random := rand.New(rand.NewPCG(11, 73))
	for _, geometry := range []struct {
		binary int32
		multi  []int32
	}{{3, nil}, {2, []int32{4}}, {1, []int32{3, 5, 2}}, {0, []int32{4, 3}}} {
		ConfigureCube(geometry.binary, geometry.multi)
		for variable := range cube.sparse {
			cube.sparse[variable] = int32(variable % 2)
		}
		for trial := 0; trial < 100; trial++ {
			family := sf_new(8, cube.size)
			family.count = int32(random.IntN(9))
			want := cost_t{cubes: family.count}
			for row := int32(0); row < family.count; row++ {
				current := set_clear(GETSET(family, row), cube.size)
				if random.IntN(2) == 0 {
					SET(current, PRIME)
					want.primes++
				}
				for part := int32(0); part < cube.size; part++ {
					present := random.IntN(2) == 0
					if present {
						set_insert(current, part)
					}
					variable := int32(0)
					for cube.last_part[variable] < part {
						variable++
					}
					if variable < cube.num_binary_vars {
						if !present {
							want.in++
						}
					} else if variable == cube.num_vars-1 {
						if present {
							want.out++
						}
					} else if (cube.sparse[variable] != 0) == present {
						want.mv++
					}
				}
			}
			want.total = want.in + want.mv + want.out
			var got, copied cost_t
			cover_cost(family, &got)
			copy_cost(&got, &copied)
			if got != want || copied != want {
				t.Fatalf("cost got %+v want %+v", got, want)
			}
			if !strings.Contains(print_cost(family), "c=") {
				t.Fatal("cost format")
			}
		}
	}
}

func TestCostAccountingAndFatal(t *testing.T) {
	savedCube, savedData := cube, cdata
	defer func() { cube, cdata = savedCube, savedData }()
	ConfigureCube(1, []int32{2})
	family := sf_new(0, cube.size)
	savedCalls, savedTime, savedTrace := total_calls[MV_REDUCE_TIME], total_time[MV_REDUCE_TIME], trace
	defer func() {
		total_calls[MV_REDUCE_TIME], total_time[MV_REDUCE_TIME], trace = savedCalls, savedTime, savedTrace
	}()
	trace = false
	var cost cost_t
	totals(ptime()-10, MV_REDUCE_TIME, family, &cost)
	if total_calls[MV_REDUCE_TIME] != savedCalls+1 || total_time[MV_REDUCE_TIME] < savedTime+10 || cost != (cost_t{}) {
		t.Fatal("execution totals")
	}
	if print_time(1234) != "1.23 sec" {
		t.Fatal(print_time(1234))
	}
	defer func() {
		if recover() != "espresso: test failure" {
			t.Error("fatal error should retain context")
		}
	}()
	fatal("test failure")
}
