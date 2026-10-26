package abc

import (
	"encoding/json"
	"os"
	"regexp"
	"slices"
	"strconv"
	"testing"
)

func TestGiaSequentialSignalsNativeSummary(t *testing.T) {
	raw, err := os.ReadFile("test_vectors/oracle/gia_enable.txt")
	if err != nil {
		t.Fatal(err)
	}
	rows := regexp.MustCompile(`Flops with set/reset =\s*(\d+)\. Flops with enable =\s*(\d+)\.`).FindAllStringSubmatch(string(raw), -1)
	if len(rows) != 2 {
		t.Fatal("native diagnostic fixture", string(raw))
	}
	fanoutRaw, err := os.ReadFile("test_vectors/oracle/gia_enable_fanout.json")
	if err != nil {
		t.Fatal(err)
	}
	var nativeFanout [][]GiaID
	if err = json.Unmarshal(fanoutRaw, &nativeFanout); err != nil || len(nativeFanout) != 6 {
		t.Fatal(err)
	}
	for mode := 0; mode < 2; mode++ {
		g := NewGiaManager()
		control, _ := g.AppendCI()
		d0, _ := g.AppendCI()
		d1, _ := g.AppendCI()
		r0, _ := g.AppendCI()
		r1, _ := g.AppendCI()
		var drivers []GiaLit
		if mode == 0 {
			n, err := g.AppendAnd(control, d0)
			if err != nil {
				t.Fatal(err)
			}
			drivers = []GiaLit{n, n}
		} else {
			if err := g.HashStart(); err != nil {
				t.Fatal(err)
			}
			first, err := g.HashMux(control, r0, d0)
			if err != nil {
				t.Fatal(err)
			}
			second, err := g.HashMux(control, r1, d1)
			if err != nil {
				t.Fatal(err)
			}
			drivers = []GiaLit{first, second}
		}
		for _, l := range drivers {
			if _, err := g.AppendCO(l); err != nil {
				t.Fatal(err)
			}
		}
		if err := g.SetRegisterCount(2); err != nil {
			t.Fatal(err)
		}
		before := slices.Clone(g.objects)
		result, err := g.DetectSequentialSignals(0)
		if err != nil {
			t.Fatal(err)
		}
		shared, _ := strconv.Atoi(rows[mode][1])
		enables, _ := strconv.Atoi(rows[mode][2])
		if result.RegistersWithSharedSetReset != shared || result.RegistersWithEnable != enables {
			t.Fatal(mode, result, rows[mode])
		}
		if mode == 0 && (result.ResetCandidates[control.ID()] != 2 || result.ResetCandidates[d0.ID()] != 2) {
			t.Fatal("reset frequencies", result)
		}
		if mode == 1 && result.EnableCandidates[control.ID()] != 2 {
			t.Fatal("enable frequencies", result)
		}
		for threshold := 0; threshold < 3; threshold++ {
			got, err := g.SequentialSignalsAboveFanout(threshold, 0)
			if err != nil || !slices.Equal(got, nativeFanout[3*mode+threshold]) {
				t.Fatal("native threshold order", mode, threshold, got, nativeFanout[3*mode+threshold], err)
			}
		}
		if _, err := g.SequentialSignalsAboveFanout(-1, 0); err == nil {
			t.Fatal("negative threshold")
		}
		if !slices.Equal(before, g.objects) {
			t.Fatal("signal detection mutated scratch")
		}
		for _, cap := range []int{-1, g.ObjectCount() - 1} {
			if _, err := g.DetectSequentialSignals(cap); err == nil {
				t.Fatal("invalid cap")
			}
		}
	}
	if _, err := NewGiaManager().DetectSequentialSignals(0); err == nil {
		t.Fatal("zero-register model")
	}
}

func TestGiaSequentialSignalCategoryDeduplication(t *testing.T) {
	g := NewGiaManager()
	common, _ := g.AppendCI()
	resetData, _ := g.AppendCI()
	setData, _ := g.AppendCI()
	for i := 0; i < 4; i++ {
		if _, err := g.AppendCI(); err != nil {
			t.Fatal(err)
		}
	}
	setGate, err := g.AppendAnd(common, setData)
	if err != nil {
		t.Fatal(err)
	}
	resetGate, err := g.AppendAnd(common, resetData)
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range []GiaLit{setGate.Not(), setGate.Not(), resetGate, resetGate} {
		if _, err := g.AppendCO(l); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.SetRegisterCount(4); err != nil {
		t.Fatal(err)
	}
	got, err := g.SequentialSignalsAboveFanout(1, 0)
	if err != nil || !slices.Equal(got, []GiaID{common.ID(), setData.ID(), resetData.ID()}) {
		t.Fatal("category order/uniqueness", got, err)
	}
	got, err = g.SequentialSignalsAboveFanout(2, 0)
	if err != nil || len(got) != 0 {
		t.Fatal("strict threshold", got, err)
	}
}
