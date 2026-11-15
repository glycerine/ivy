package gia

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestCubeReachabilityNativePrivateReports(t *testing.T) {
	exe := compileGiaSourceOracle(t, "era_cube_review.c")
	native, err := exec.Command(exe, "--reports").CombinedOutput()
	if err != nil {
		t.Fatalf("C reports: %v\n%s", err, native)
	}
	p := Gia_ManAreCreate(eraReviewGraph(1, 3, 0))
	defer Gia_ManAreFree(p)
	a := Gia_ManAreCreateStaInit(p)
	Gia_ManAreCubeAddToList(p, &p.Root, a)
	b := Gia_ManAreCreateSta(p)
	Gia_StaSetValue1(b, 0)
	b.iPrev = Gia_Int2Ptr(1)
	Gia_ManAreCubeAddToList(p, &p.Root, b)
	p.iStaCur = 2
	goText := hashReviewStdout(t, func() {
		fmt.Println("private reports")
		Gia_ManArePrintReport(p, time.Now(), 0)
		Gia_ManArePrintReport(p, time.Now(), 1)
		fmt.Println("private cubes")
		Gia_ManArePrintCube(p, a)
		Gia_ManArePrintCube(p, b)
		Gia_StaSetUnused(a)
		printed := Gia_ManArePrintUsed(p)
		fmt.Printf("print return %d\n", printed)
		fmt.Printf("counts %d %d %d\n", Gia_ManAreDepth(p, 2), Gia_ManAreListCountUsed(p), Gia_ManCountMinterms(p))
	})
	timeField := regexp.MustCompile(`Time =([ ]*[0-9]+\.[0-9]{2}) sec`)
	memField := regexp.MustCompile(`Mem =([ ]*[0-9]+\.[0-9]{2}) MB\.`)
	address := regexp.MustCompile(`0x[0-9a-f]+`)
	normalize := func(text string) string {
		times := timeField.FindAllStringSubmatch(text, -1)
		memories := memField.FindAllStringSubmatch(text, -1)
		if len(times) != 2 || len(memories) != 2 {
			t.Errorf("missing private report fields: %q", text)
		}
		for _, field := range append(times, memories...) {
			if len(field[1]) != 9 {
				t.Errorf("numeric report width%d want9: %q", len(field[1]), field[0])
			}
		}
		text = timeField.ReplaceAllString(text, "Time = <elapsed> sec")
		text = memField.ReplaceAllString(text, "Mem = <storage> MB.")
		return address.ReplaceAllString(text, "<address>")
	}
	if got, want := normalize(goText), normalize(string(native)); got != want {
		t.Fatalf("Go reports\n%q\nC reports\n%q", got, want)
	}
}

func TestCubeReachabilityNativeDriverReports(t *testing.T) {
	exe := compileGiaSourceOracle(t, "era_cube_review.c")
	native, err := exec.Command(exe, "--driver-reports").CombinedOutput()
	if err != nil {
		t.Fatalf("C driver reports: %v\n%s", err, native)
	}
	goText := hashReviewStdout(t, func() {
		for seed := int32(0); seed < 4; seed++ {
			for verbose := int32(0); verbose < 2; verbose++ {
				g := eraReviewGraph(1, 3, seed)
				g.pName = "cube_review"
				fmt.Printf("driver %d %d\n", seed, verbose)
				Gia_ManArePerform(g, 100, 0, verbose)
				Gia_ManStop(g)
			}
		}
	})
	elapsed := regexp.MustCompile(`Time =[ ]*[0-9]+\.[0-9]{2} sec`)
	memory := regexp.MustCompile(`Mem =[ ]*[0-9]+\.[0-9]{2} MB\.`)
	parts := regexp.MustCompile(`(Cofactoring|Containment|Other      |TOTAL      ) =[ ]*[0-9]+\.[0-9]{2} sec \([ ]*[0-9]+\.[0-9]{2} %\)`)
	progress := regexp.MustCompile(`States =[ ]*([0-9]+)\. Reached =[ ]*([0-9]+)\. R = 1\.000\. Depth =[ ]*([0-9]+)\.`)
	normalize := func(text string, goOutput bool) string {
		// C indexes the zero-filled record past nStas on successful completion,
		// reporting a spurious depth one. Go deliberately bounds this lookup.
		text = progress.ReplaceAllStringFunc(text, func(line string) string {
			match := progress.FindStringSubmatch(line)
			if match[1] != match[2] {
				return line
			}
			if goOutput && match[3] != "0" {
				t.Errorf("past-end Go depth should be guarded: %s", line)
			}
			return line[:strings.Index(line, "Depth =")] + "Depth =<past-end>."
		})
		text = elapsed.ReplaceAllString(text, "Time = <elapsed> sec")
		text = memory.ReplaceAllString(text, "Mem = <storage> MB.")
		return parts.ReplaceAllString(text, "$1 = <elapsed> sec (<percent> %)")
	}
	if got, want := normalize(goText, true), normalize(string(native), false); got != want {
		t.Fatalf("Go driver reports\n%q\nC driver reports\n%q", got, want)
	}
}

func TestCubeReachabilityFailureMessageUsesWitness(t *testing.T) {
	g := cubeReviewCounter(6)
	defer Gia_ManStop(g)
	g.pName = "counter"
	text := hashReviewStdout(t, func() {
		if Gia_ManArePerform(g, 100, 1, 0) != 0 {
			t.Fatal("counter did not fail")
		}
	})
	want := "Output 0 of miter \"counter\" was asserted in frame 32.\n"
	if !strings.Contains(text, want) {
		t.Fatalf("missing actual witness diagnostic%q in%q", want, text)
	}
	if !strings.Contains(text, "Generated counter-example verified correctly.") {
		t.Fatal("missing verified-witness diagnostic")
	}
}
