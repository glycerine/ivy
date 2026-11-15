package gia

import (
	"math/rand"
	"testing"
)

func TestQuantSupportTypedDagIndependentSets(t *testing.T) {
	for _, subset := range []bool{false, true} {
		g := Gia_ManStart(1024)
		g.EnableMuxes()
		Gia_ManQuantSetSuppStart(g)
		tracked := map[int32]int{}
		var pool []int32
		for i := 0; i < 193; i++ {
			lit := Gia_ManAppendCi(g)
			pool = append(pool, lit)
			Gia_ManQuantSetSuppZero(g)
			if !subset || i%3 != 1 {
				tracked[lit>>1] = len(tracked)
				Gia_ManQuantSetSuppCi(g, Gia_ManObj(g, lit>>1))
			}
		}
		rng := rand.New(rand.NewSource(71))
		for i := 0; i < 256; i++ {
			a, b, c := pool[rng.Intn(len(pool))], pool[rng.Intn(len(pool))], pool[rng.Intn(len(pool))]
			if a>>1 == b>>1 || a>>1 == c>>1 || b>>1 == c>>1 {
				i--
				continue
			}
			a ^= int32(rng.Intn(2))
			b ^= int32(rng.Intn(2))
			c ^= int32(rng.Intn(2))
			var root int32
			switch i % 4 {
			case 0:
				root = Gia_ManAppendAnd(g, a, b)
			case 1:
				root = Gia_ManAppendXorReal(g, a, b)
			case 2:
				root = Gia_ManAppendMuxReal(g, c, a, b)
			case 3:
				root = Gia_ManAppendBuf(g, a)
			}
			pool = append(pool, root)
		}
		expected := make([][]bool, g.nObjs)
		for id := int32(0); id < g.nObjs; id++ {
			expected[id] = make([]bool, len(tracked))
			a, b, c, kind := g.Gate(id)
			if bit, ok := tracked[id]; ok {
				expected[id][bit] = true
			}
			var edges []int32
			switch kind {
			case 2, 3:
				edges = []int32{a, b}
			case 4:
				edges = []int32{a, b, c}
			case 5:
				edges = []int32{a}
			}
			for _, edge := range edges {
				for bit, value := range expected[edge>>1] {
					expected[id][bit] = expected[id][bit] || value
				}
			}
			row := Gia_ManQuantInfoId(g, id)
			for bit := 0; bit < len(row)*64; bit++ {
				want := bit < len(tracked) && expected[id][bit]
				got := row[bit>>6]&(uint64(1)<<uint(bit&63)) != 0
				if got != want {
					t.Fatalf("subset%t node%d bit%d got%t want%t", subset, id, bit, got, want)
				}
			}
		}
		// Cross-check the structural recursive verifier for typed MUX controls too.
		report := hashReviewStdout(t, func() {
			for id := g.nObjs - 8; id < g.nObjs; id++ {
				Gia_ManQuantVerify(g, id)
			}
		})
		if report != "" {
			t.Fatalf("support verifier disagrees: %s", report)
		}
		count := g.iSuppPi
		for bit := int32(0); bit < count; bit++ {
			g.iSuppPi = bit
			root := g.nObjs - 2
			if !expected[root][bit] {
				continue
			}
			reachable := map[int32]bool{}
			stack := []int32{root}
			for len(stack) > 0 {
				id := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if reachable[id] {
					continue
				}
				reachable[id] = true
				a, b, c, kind := g.Gate(id)
				switch kind {
				case 2, 3:
					stack = append(stack, a>>1, b>>1)
				case 4:
					stack = append(stack, a>>1, b>>1, c>>1)
				case 5:
					stack = append(stack, a>>1)
				}
			}
			want := int32(0)
			for id := range reachable {
				_, _, _, kind := g.Gate(id)
				if kind >= 2 && expected[id][bit] {
					want++
				}
			}
			if got := Gia_ManQuantCountUsed(g, root); got != want {
				t.Fatalf("subset%t bit%d affected nodes%d want%d", subset, bit, got, want)
			}
		}
		g.iSuppPi = count
		keep := func(_ any, ci int32) int32 { return giaBool(ci%5 != 1) }
		Gia_ManQuantMarkUsedCis(g, keep, nil)
		marker := Gia_ManQuantInfoId(g, 0)
		for id, bit := range tracked {
			ci := Gia_ObjCioId(g.pObjs[id])
			want := ci%5 == 1
			if got := marker[bit>>6]&(uint64(1)<<uint(bit&63)) != 0; got != want {
				t.Fatalf("marker node%d bit%d got%t want%t", id, bit, got, want)
			}
		}
		Gia_ManStop(g)
	}
}

func TestQuantificationDeepSharedCone(t *testing.T) {
	for mode := 0; mode < 2; mode++ {
		g := Gia_ManStart(25000)
		Gia_ManQuantSetSuppStart(g)
		a := Gia_ManAppendCi(g)
		Gia_ManQuantSetSuppZero(g)
		Gia_ManQuantSetSuppCi(g, Gia_ManObj(g, a>>1))
		b := Gia_ManAppendCi(g)
		Gia_ManQuantSetSuppZero(g)
		Gia_ManQuantSetSuppCi(g, Gia_ManObj(g, b>>1))
		Gia_ManHashAlloc(g)
		root := Gia_ManAppendAnd(g, a, b)
		for i := 0; i < 7000; i++ {
			left := Gia_ManAppendAnd(g, root, a)
			right := Gia_ManAppendAnd(g, root, b)
			root = Gia_ManAppendAnd(g, left, right)
		}
		keepB := func(_ any, ci int32) int32 { return giaBool(ci == 1) }
		var result int32
		if mode == 0 {
			result = Gia_ManQuantExist(g, root, keepB, nil)
		} else {
			result = Gia_ManQuantExist2(g, root, keepB, nil)
		}
		// Every shared level is still a AND b, so eliminating a must leave b.
		for bits := 0; bits < 4; bits++ {
			values := cofReviewValues(g, bits)
			got := values[result>>1] != (result&1 != 0)
			if got != (bits&2 != 0) {
				t.Fatalf("mode%d input%d literal%d got%t", mode, bits, result, got)
			}
		}
		for _, copy := range Vec_IntArray(&g.vCopies) {
			if copy != -1 {
				t.Fatal("left an occupied copy-table slot")
			}
		}
		Gia_ManStop(g)
	}
}
