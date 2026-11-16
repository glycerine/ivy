package gia

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"
)

func mappingReviewReportGraph(variant int) *Gia_Man_t {
	p := Gia_ManStart(64)
	a, b, c, d := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	var roots [4]int32
	roots[0] = Gia_ManAppendMux(p, a^int32(variant&1), b^int32(variant>>1&1), c^int32(variant>>2&1))
	roots[1] = Gia_ManAppendMux(p, d, roots[0], c)
	roots[2] = Gia_ManAppendAnd(p, roots[0], b)
	roots[3] = Gia_ManAppendAnd(p, roots[2], d)
	leaves := [][]int32{{a >> 1, b >> 1, c >> 1}, {d >> 1, roots[0] >> 1, c >> 1}, {roots[0] >> 1, b >> 1}, {roots[2] >> 1, d >> 1}}
	for i, root := range roots {
		Gia_ManAppendCo(p, root^int32(i&1))
	}
	Gia_ManAppendCo(p, a)
	p.SetIdentity("mapping_A", "")
	p.VMapping = Vec_IntStart(int(p.nObjs))
	for i, root := range roots {
		id := root >> 1
		Vec_IntWriteEntry(p.VMapping, int(id), int32(Vec_IntSize(p.VMapping)))
		Vec_IntPush(p.VMapping, int32(len(leaves[i])))
		for _, fanin := range leaves[i] {
			Vec_IntPush(p.VMapping, fanin)
		}
		if i < 3 {
			id = -id
		}
		Vec_IntPush(p.VMapping, id)
	}
	p.VPacking = Vec_IntAlloc(16)
	Vec_IntPush(p.VPacking, int32(variant%4))
	for i := 1; i <= variant%4; i++ {
		Vec_IntPush(p.VPacking, int32(i))
		for k := 0; k < i; k++ {
			Vec_IntPush(p.VPacking, roots[k]>>1)
		}
	}
	return p
}
func TestMappingNativeMuxPackingAndProfiles(t *testing.T) {
	exe := compileGiaSourceOracle(t, "mapping_reports_review.c")
	for variant := 0; variant < 8; variant++ {
		source, err := exec.Command(exe, "0", strconv.Itoa(variant)).CombinedOutput()
		if err != nil {
			t.Fatalf("native mux%d: %v %s", variant, err, source)
		}
		p := mappingReviewReportGraph(variant)
		got := hashReviewStdout(t, func() {
			for id := int32(1); id < p.nObjs; id++ {
				if Gia_ObjIsLut(p, id) != 0 && Gia_ObjLutSize(p, id) == 3 {
					var fs [3]int32
					Gia_ManPrintGetMuxFanins(p, p.pObjs[id], fs[:])
					fmt.Printf("mux %d %d %d %d\n", id, fs[0], fs[1], fs[2])
				}
			}
			count := Gia_ManCountDupLut(p)
			fmt.Printf("duplicates %d\n", count)
			Gia_ManPrintMappingStats(p, "")
			Gia_ManPrintPackingStats(p)
			Gia_ManPrintLutStats(p)
		})
		if string(source) != got {
			t.Logf("variant%d", variant)
			esopReviewCompare(t, source, got)
		}
		for _, o := range p.pObjs {
			if o.fMark0 != 0 || o.fMark1 != 0 {
				t.Fatal("report leaves marks set")
			}
		}
		Gia_ManStop(p)
	}
	for variant := 0; variant < 66; variant++ {
		source, err := exec.Command(exe, "1", strconv.Itoa(variant)).CombinedOutput()
		if err != nil {
			t.Fatalf("native profile%d: %v %s", variant, err, source)
		}
		counts := make([]int32, 33)
		for i := range counts {
			counts[i] = int32((i*17 + variant*13) % 37)
		}
		got := hashReviewStdout(t, func() { Gia_ManPrintNodeProfile(counts, int32(variant%33)) })
		if string(source) != got {
			t.Logf("profile%d", variant)
			esopReviewCompare(t, source, got)
		}
	}
}
func TestMappingNativePersistentDumpAndEmptyLuts(t *testing.T) {
	exe := compileGiaSourceOracle(t, "mapping_reports_review.c")
	dir := t.TempDir()
	sourceFile, goFile := filepath.Join(dir, "source.txt"), filepath.Join(dir, "go.txt")
	source, err := exec.Command(exe, "2", "7", sourceFile).CombinedOutput()
	if err != nil {
		t.Fatalf("native dump: %v %s", err, source)
	}
	giaMappingStatsState.Lock()
	oldName, oldSince := giaMappingStatsState.name, giaMappingStatsState.since
	giaMappingStatsState.name = ""
	giaMappingStatsState.Unlock()
	defer func() {
		giaMappingStatsState.Lock()
		giaMappingStatsState.name, giaMappingStatsState.since = oldName, oldSince
		giaMappingStatsState.Unlock()
	}()
	p := mappingReviewReportGraph(7)
	defer Gia_ManStop(p)
	got := hashReviewStdout(t, func() {
		Gia_ManPrintMappingStats(p, goFile)
		Gia_ManPrintMappingStats(p, goFile)
		p.SetIdentity("mapping_B", "")
		Gia_ManPrintMappingStats(p, goFile)
		p.SetIdentity("mapping_A", "")
		Gia_ManPrintMappingStats(p, goFile)
	})
	if string(source) != got {
		esopReviewCompare(t, source, got)
	}
	want, err := os.ReadFile(sourceFile)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := os.ReadFile(goFile)
	if err != nil {
		t.Fatal(err)
	}
	elapsed := regexp.MustCompile(`[0-9]+\.[0-9]{2}`)
	if elapsed.ReplaceAllString(string(want), "<elapsed>") != elapsed.ReplaceAllString(string(actual), "<elapsed>") {
		t.Fatalf("Go table%q; source%q", actual, want)
	}
	source, err = exec.Command(exe, "3", "0").CombinedOutput()
	if err != nil {
		t.Fatalf("native empty LUTs: %v %s", err, source)
	}
	p.VMapping = Vec_IntStart(int(p.nObjs))
	got = hashReviewStdout(t, func() { Gia_ManPrintLutStats(p) })
	if string(source) != got {
		t.Fatalf("Go empty LUTs%q; source%q", got, source)
	}
}
