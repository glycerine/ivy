package mpm

import (
	"encoding/binary"
	"fmt"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"os"
)

func Mpm_ManPrintDsdStats(p *Mpm_Man_t) {
	absent := 0
	for i, n := range p.nCountDsd {
		if n == 0 {
			absent++
			continue
		}
		if p.pPars.FVeryVerbose != 0 {
			fmt.Printf("%5d  :  %-20s   %8d \n", i, p.pDsd6[i].pStr, n)
		}
	}
	fmt.Printf("Unused classes = %d (%.2f %%).  Non-DSD cuts = %d (%.2f %%).  No-match cuts = %d (%.2f %%).\n", absent, 100*float64(absent)/595, p.nNonDsd, 100*float64(p.nNonDsd)/float64(p.nCutsMergedAll), p.nNoMatch, 100*float64(p.nNoMatch)/float64(p.nCutsMergedAll))
}
func Mpm_ManBuildHashTable(out **vec.Vec_Int_t) *vec.Hsh_IntMan_t {
	data, err := os.ReadFile("dsdfuncs6.dat")
	if err != nil {
		panic(err)
	}
	n := len(data) / 12
	truths := vec.Vec_WrdStart(n)
	configs := vec.Vec_IntStart(n)
	for i := 0; i < n; i++ {
		vec.Vec_WrdWriteEntry(truths, i, binary.NativeEndian.Uint64(data[8*i:]))
		vec.Vec_IntWriteEntry(configs, i, int32(binary.NativeEndian.Uint32(data[8*n+4*i:])))
	}
	hash := vec.Hsh_WrdManHashArrayStart(truths, 1)
	if out != nil {
		*out = configs
	} else {
		vec.Vec_IntFree(configs)
	}
	vec.Vec_WrdFree(truths)
	return hash
}
func Mpm_ManPrintPerm(s uint32) {
	for i := 0; i < 6; i++ {
		fmt.Printf("%d ", s>>(3*i)&7)
	}
	fmt.Print("  ")
}
func Mpm_ManPrecomputePerms(p *Mpm_Man_t) {
	start := util.Abc_Clock()
	assert(p.pDsd6 == nil, "DSD initialized once")
	p.pDsd6 = s_DsdClass6[:]
	perm := extra.Extra_PermSchedule(6)
	p.vMap2Perm = vec.Vec_IntStartFull(1 << 18)
	var one int32
	for x := int32(0); x < 6; x++ {
		p.Perm6[0][x] = byte(x)
		one |= x << (3 * x)
	}
	original := one
	for k := 0; k < 720; k++ {
		if k > 0 {
			p.Perm6[k] = p.Perm6[k-1]
		}
		j := perm[k]
		p.Perm6[k][j], p.Perm6[k][j+1] = p.Perm6[k][j+1], p.Perm6[k][j]
		n := (one>>(3*j) ^ one>>(3*(j+1))) & 7
		one ^= n << (3 * j)
		one ^= n << (3 * (j + 1))
		vec.Vec_IntWriteEntry(p.vMap2Perm, int(one), int32(k))
	}
	assert(one == original, "permutation schedule")
	for i := 0; i < vec.Vec_IntSize(p.vMap2Perm); i++ {
		one = int32(i)
		var used [6]bool
		count, k := 0, 0
		for ; k < 6; k++ {
			v := (one >> (3 * k)) & 7
			if v >= 6 && v < 7 {
				break
			}
			if v != 7 {
				if used[v] {
					break
				}
				used[v] = true
				count++
			}
		}
		if k < 6 || count == 6 {
			continue
		}
		var missing []int32
		for x := 0; x < 6; x++ {
			if !used[x] {
				missing = append(missing, int32(x))
			}
		}
		x := 0
		for k = 0; k < 6; k++ {
			if (one>>(3*k))&7 == 7 {
				one ^= (missing[x] ^ 7) << (3 * k)
				x++
			}
		}
		assert(x == len(missing), "missing permutation variables")
		n := vec.Vec_IntEntry(p.vMap2Perm, int(one))
		assert(n != -1, "permutation exists")
		vec.Vec_IntWriteEntry(p.vMap2Perm, i, n)
	}
	assert(p.vPerm6 == nil, "permutations initialized once")
	p.vPerm6 = vec.Vec_WrdAlloc(720 * 595)
	for _, class := range s_DsdClass6 {
		truth := class.uTruth
		for k := 0; k < 720; k++ {
			truth = util.Abc_Tt6SwapAdjacent(truth, perm[k])
			vec.Vec_WrdPush(p.vPerm6, truth)
		}
		assert(truth == class.uTruth, "permutation cycle")
	}
	p.pHash = Mpm_ManBuildHashTable(&p.vConfgRes)
	util.Abc_PrintTime(1, "Setting up DSD information", util.Abc_Clock()-start)
}
func Mpm_CutTruthFromDsd(p *Mpm_Man_t, cut *Mpm_Cut_t, class int32) word {
	truth := p.pDsd6[class].uTruth
	assert(p.pDsd6[class].nVars == int32(cut.nLeaves), "class support")
	for i := int32(0); i < int32(cut.nLeaves); i++ {
		if util.Abc_LitIsCompl(cut.pLeaves[i]) != 0 {
			truth = util.Abc_Tt6Flip(truth, i)
		}
	}
	return truth
}
func Mpm_CutCheckDsd6(p *Mpm_Man_t, truth word) int32 {
	complement := int32(truth & 1)
	if complement != 0 {
		truth = ^truth
	}
	entry := *vec.Hsh_IntManLookup(p.pHash, []int32{int32(truth), int32(truth >> 32)})
	if entry == -1 {
		return -1
	}
	config := vec.Vec_IntEntry(p.vConfgRes, int(entry))
	if complement != 0 {
		config ^= 1 << 16
	}
	return config
}
func Mpm_CutComputeDsd6(p *Mpm_Man_t, cut, a, b, c *Mpm_Cut_t, ca, cb, cc, typ int32) int32 {
	truth0 := p.pDsd6[int32(a.iFunc)>>1].uTruth
	perm1 := vec.Vec_IntEntry(p.vMap2Perm, int(p.uPermMask[1]))
	truth1 := vec.Vec_WrdEntry(p.vPerm6, int(int32(b.iFunc>>1)*720+perm1))
	for i := int32(0); i < 6; i++ {
		if p.uComplMask[1]>>i&1 != 0 {
			truth1 = util.Abc_Tt6Flip(truth1, i)
		}
	}
	if ca^int32(a.fCompl)^int32(a.iFunc&1) != 0 {
		truth0 = ^truth0
	}
	if cb^int32(b.fCompl)^int32(b.iFunc&1) != 0 {
		truth1 = ^truth1
	}
	var truth word
	if c == nil {
		switch typ {
		case 1:
			truth = truth0 & truth1
		case 2:
			truth = truth0 ^ truth1
		default:
			panic("invalid DSD cut type")
		}
	} else {
		permc := vec.Vec_IntEntry(p.vMap2Perm, int(p.uPermMask[2]))
		control := vec.Vec_WrdEntry(p.vPerm6, int(int32(c.iFunc>>1)*720+permc))
		for i := int32(0); i < 6; i++ {
			if p.uComplMask[2]>>i&1 != 0 {
				control = util.Abc_Tt6Flip(control, i)
			}
		}
		if cc^int32(c.fCompl)^int32(c.iFunc&1) != 0 {
			control = ^control
		}
		truth = control&truth1 | ^control&truth0
	}
	config := Mpm_CutCheckDsd6(p, truth)
	if config == -1 {
		p.nNonDsd++
		return 0
	}
	class, compl := config>>17, (config>>16)&1
	config &= 0xffff
	if p.pPars.FMap4Gates != 0 && vec.Vec_IntSize(vec.Vec_WecEntry(p.vNpnConfigs, int(class))) == 0 {
		p.nNoMatch++
		return 0
	}
	cut.iFunc = uint32(util.Abc_Var2Lit(class, compl))
	assert(config>>6 < 720, "DSD permutation")
	leaves := [6]int32{-1, -1, -1, -1, -1, -1}
	for i := int32(0); i < int32(cut.nLeaves); i++ {
		leaves[p.Perm6[config>>6][i]] = util.Abc_LitNotCond(cut.pLeaves[i], config>>i&1)
	}
	cut.nLeaves = uint32(p.pDsd6[class].nVars)
	for i := 0; i < int(cut.nLeaves); i++ {
		assert(leaves[i] != -1, "DSD support leaf")
		cut.pLeaves[i] = leaves[i]
	}
	p.nCountDsd[class]++
	if cut.nLeaves < 2 {
		p.nSmallSupp++
	}
	return 1
}
