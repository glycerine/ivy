package gia

import (
	"github.com/glycerine/abcgo/misc/tim"
	"testing"
)

func TestBoxTimingAccess(t *testing.T) {
	p, extra := Gia_ManStart(4), Gia_ManStart(4)
	tm := tim.Tim_ManStart(0, 0)
	p.SetBoxTiming(tm, extra)
	a, b := p.BoxTiming()
	if a != tm || b != extra {
		t.Fatal("timing metadata not stored")
	}
	p.SetBoxTiming(nil, nil)
	a, b = p.BoxTiming()
	if a != nil || b != nil {
		t.Fatal("timing metadata not cleared")
	}
}

func TestClearLevels(t *testing.T) {
	p := Gia_ManStart(4)
	Gia_ManCleanLevels(p, 1)
	if p.vLevels == nil {
		t.Fatal("levels were not allocated")
	}
	Gia_ManClearLevels(p)
	if p.vLevels != nil {
		t.Fatal("levels retained")
	}
}

func TestHasLevels(t *testing.T) {
	p := Gia_ManStart(4)
	if Gia_ManHasLevels(p) {
		t.Fatal("unexpected levels")
	}
	Gia_ManCleanLevels(p, 1)
	if !Gia_ManHasLevels(p) {
		t.Fatal("missing levels")
	}
	Gia_ManClearLevels(p)
	if Gia_ManHasLevels(p) {
		t.Fatal("levels retained")
	}
}
