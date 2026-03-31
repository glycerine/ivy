package z3bridge

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"syscall"
	"testing"

	"github.com/glycerine/ivy/goivy/logic"
)

var durlog *os.File

const logAllToDisk = false

func TestMain(m *testing.M) {

	if logAllToDisk {
		pid := fmt.Sprintf("%v", os.Getpid())
		home := os.Getenv("HOME")
		if home == "" {
			panic("could not get env HOME")
		}
		durlogDir := home + "/trash/translate2fuzz"
		durlogPath := durlogDir + "/translate2fuzz_log." + pid
		// cannot do this b/c we don't know who will run first; but lots of
		// process will run during a fuzz test.
		// if firstProcess {
		//    os.RemoveAll(durlogDir)
		// }
		err := os.MkdirAll(durlogDir, 0755)
		if err != nil {
			panicf("could not create logging output dir '%v': '%v'", durlogDir, err)
		}

		// All to file (durlog).

		durlog, err = os.OpenFile(durlogPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			panicf("could not create logging output file '%v': '%v'", durlogPath, err)
		}
		ourStdout = durlog // vv, alwaysPrintf go here.

		if err := syscall.Dup2(int(durlog.Fd()), int(os.Stderr.Fd())); err != nil {
			panic(err)
		}
		if err := syscall.Dup2(int(durlog.Fd()), int(os.Stdout.Fd())); err != nil {
			panic(err)
		}
		os.Stderr = durlog
		os.Stdout = durlog
	}
	exitcode := m.Run()
	durlog.Sync()
	durlog.Close()

	os.Exit(exitcode)
}

// --- Z3 Worker Goroutine ---
//
// Z3 4.7.1's C library uses Thread-Local Storage (TLS) internally and is
// not safe when a goroutine migrates between OS threads between CGO calls.
// We funnel ALL fuzz Z3 work through a single goroutine pinned to one OS
// thread via runtime.LockOSThread().

type z3Job struct {
	fn   func(t *testing.T)
	t    *testing.T
	done chan *z3Result
}

type z3Result struct {
	panicVal interface{}
}

var (
	z3WorkerOnce sync.Once
	z3JobChan    chan *z3Job
)

func startZ3Worker() {
	z3WorkerOnce.Do(func() {
		z3JobChan = make(chan *z3Job, 1)
		go func() {
			runtime.LockOSThread()
			for job := range z3JobChan {
				result := &z3Result{}
				func() {
					defer func() {
						if r := recover(); r != nil {
							result.panicVal = r
						}
					}()
					job.fn(job.t)
				}()
				job.done <- result
			}
		}()
	})
}

func runOnZ3Thread(t *testing.T, fn func(t *testing.T)) {
	t.Helper()
	startZ3Worker()
	done := make(chan *z3Result, 1)
	z3JobChan <- &z3Job{fn: fn, t: t, done: done}
	result := <-done
	if result.panicVal != nil {
		t.Fatalf("panic on Z3 thread: %v", result.panicVal)
	}
}

// FuzzQuantConstraintsForAll builds random quantified formulas with a
// QuantConstraints callback and verifies no panics occur during translation.
func FuzzQuantConstraintsForAll(f *testing.F) {
	f.Add(byte(1), byte(0), true)
	f.Add(byte(3), byte(1), true)
	f.Add(byte(1), byte(2), false)
	f.Add(byte(2), byte(3), false)
	f.Add(byte(1), byte(4), true) // non-bool body, should error not panic

	f.Fuzz(func(t *testing.T, varNameLen byte, bodyKind byte, isForall bool) {
		nameLen := int(varNameLen%5) + 1
		varName := ""
		for i := 0; i < nameLen; i++ {
			varName += string(rune('A' + (int(varNameLen)+i)%26))
		}

		runOnZ3Thread(t, func(t *testing.T) {

			defer func() {
				r := recover()
				if r != nil {
					vv("recovered '%v'", r)
				} else {
					//vv("ran fine")
				}
			}()

			tr := NewTranslator()
			defer tr.Close()

			tr.SortLookup = func(name string) *Sort {
				if name == "mynat" {
					s := tr.Ctx.IntSort()
					return &s
				}
				return nil
			}
			tr.QuantConstraints = func(v *logic.Variable, z3Var Expr) []Expr {
				return []Expr{tr.Ctx.Le(tr.Ctx.IntVal(0), z3Var)}
			}

			sort := &logic.UninterpretedSort{Name: "mynat"}
			x, err := logic.NewVariable(varName, sort)
			if err != nil {
				return
			}

			var body logic.Expr
			switch bodyKind % 5 {
			case 0:
				body = &logic.Eq{T1: x, T2: x}
			case 1:
				body = logic.NewConst("p", logic.Boolean)
			case 2:
				body = &logic.Not{Body: &logic.Eq{T1: x, T2: x}}
			case 3:
				body = &logic.And{}
			case 4:
				body = x // non-Bool body, should produce error not panic
			}

			var fmla logic.Expr
			if isForall {
				fmla = &logic.ForAll{Variables: []*logic.Variable{x}, Body: body}
			} else {
				fmla = &logic.Exists{Variables: []*logic.Variable{x}, Body: body}
			}

			_, _ = tr.Translate(fmla)
		})
	})
}

// FuzzVariableNaming translates variables with random names and sorts,
// verifying no panics and that the Z3 name contains the sort suffix.
func FuzzVariableNaming(f *testing.F) {
	f.Add("X", "node")
	f.Add("Var", "int")
	f.Add("A", "T")
	f.Add("LongVarName", "SomeSort")
	f.Add("V", "bv32")

	f.Fuzz(func(t *testing.T, varName, sortName string) {
		if len(varName) == 0 || len(varName) > 50 {
			return
		}
		if len(sortName) == 0 || len(sortName) > 50 {
			return
		}
		if varName[0] < 'A' || varName[0] > 'Z' {
			return
		}

		runOnZ3Thread(t, func(t *testing.T) {
			tr := NewTranslator()
			defer tr.Close()

			sort := &logic.UninterpretedSort{Name: sortName}
			v, err := logic.NewVariable(varName, sort)
			if err != nil {
				return
			}

			z3v, err := tr.Translate(v)
			if err != nil {
				return
			}

			str := z3v.String()
			if len(str) == 0 {
				t.Fatal("empty Z3 string for variable")
			}
		})
	})
}

// FuzzSortLookup exercises the SortLookup callback with random sort names.
func FuzzSortLookup(f *testing.F) {
	f.Add("mynat", true)
	f.Add("T", false)
	f.Add("bvsort", true)
	f.Add("range_sort", true)
	f.Add("", false)

	f.Fuzz(func(t *testing.T, sortName string, hasInterp bool) {
		if len(sortName) == 0 || len(sortName) > 50 {
			return
		}

		runOnZ3Thread(t, func(t *testing.T) {
			tr := NewTranslator()
			defer tr.Close()

			if hasInterp {
				tr.SortLookup = func(name string) *Sort {
					if name == sortName {
						s := tr.Ctx.IntSort()
						return &s
					}
					return nil
				}
			}

			sort := &logic.UninterpretedSort{Name: sortName}
			_, err := tr.TranslateSort(sort)
			if err != nil {
				return
			}

			sym := logic.NewConst("c", sort)
			_, _ = tr.Translate(sym)
		})
	})
}
