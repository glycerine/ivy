package goivy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const traceArchiveFormat = "ivy-trace-1"

type traceArchive struct {
	Format   string         `json:"format"`
	Detailed string         `json:"detailed"`
	Brief    string         `json:"brief"`
	Tracking *traceTracking `json:"tracking,omitempty"`
}

// ReplayTrace prints either view of an archived trace.
func ReplayTrace(path string, out io.Writer, detailed bool, track string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	var archive traceArchive
	if err := json.NewDecoder(file).Decode(&archive); err != nil {
		return fmt.Errorf("read trace %s: %w", path, err)
	}
	if archive.Format != traceArchiveFormat {
		return fmt.Errorf("unsupported trace format %q in %s", archive.Format, path)
	}
	if detailed && track != "" {
		if archive.Tracking == nil {
			return fmt.Errorf("trace %s has no recorded values for tracking", path)
		}
		return archive.Tracking.replay(out, track)
	}
	view := archive.Brief
	if detailed {
		view = archive.Detailed
	}
	_, err = io.WriteString(out, view)
	return err
}

type traceFiles struct {
	written map[string]bool
	err     error
}

func openTraceFiles(dir string) (*traceFiles, error) {
	if dir != "" {
		if err := os.Mkdir(dir, 0o700); err != nil {
			if errors.Is(err, os.ErrExist) {
				return nil, fmt.Errorf("trace_dir %q already exists", dir)
			}
			return nil, fmt.Errorf("create trace directory: %w", err)
		}
	}
	return &traceFiles{written: make(map[string]bool)}, nil
}

func (files *traceFiles) write(path string, trace *TraceBase, exclusive bool) {
	if files.err != nil {
		return
	}
	archive := traceArchive{
		Format: traceArchiveFormat, Detailed: trace.format(true), Brief: trace.format(false),
		Tracking: recordTraceTracking(trace),
	}
	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if exclusive {
		flags = os.O_WRONLY | os.O_CREATE | os.O_EXCL
	}
	file, err := os.OpenFile(path, flags, 0o600)
	if err == nil {
		err = json.NewEncoder(file).Encode(archive)
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
	}
	if err != nil {
		files.err = fmt.Errorf("write trace %s: %w", path, err)
		return
	}
	fmt.Printf("        wrote counterexample to %s (view with \"goivy_check replay %s\")\n", path, path)
}

func traceArchiveName(name string) string {
	name = strings.NewReplacer("/", "_", "\\", "_").Replace(name)
	if name == "" {
		return "trace"
	}
	return name
}

func (files *traceFiles) writeNamed(dir, name string, trace *TraceBase) {
	name = traceArchiveName(name)
	if files.written[name] {
		return
	}
	files.write(filepath.Join(dir, name+".a2g"), trace, true)
	files.written[name] = files.err == nil
}

func traceEachFailure(mod *Module, history *History, clauses *Clauses, checkers []Checker) {
	files := mod.Cfg.traceFiles
	for _, checker := range checkers {
		if checker.Assume() {
			checker.Start()
			clauses = AndClausesTyped(clauses, checker.Cond())
			continue
		}
		model, solver := SmallModelClauses(clauses, []FinalCond{checker}, true, mod)
		if model == nil || !checker.Failed() || files.err != nil {
			continue
		}
		name := "trace"
		if lf := checker.GetLF(); lf != nil && lf.Label != nil {
			name = lfLabelName(lf)
		} else if base, ok := checker.(*BaseChecker); ok && base.TraceName != "" {
			name = base.TraceName
		}
		if files.written[traceArchiveName(name)] {
			continue
		}
		failed := []Checker{checker}
		trace := buildCheckTrace(mod, history, clauses, failed, failed, solver, model)
		files.writeNamed(mod.Cfg.TraceDir, name, trace.TraceBase)
	}
}

func guaranteeTraceName(action ActionsAction) string {
	if name := guaranteeName(action); name != "" {
		return name
	}
	loc := action.GetLineno()
	for loc.Reference != nil {
		loc = *loc.Reference
	}
	if loc.Line == 0 {
		return "assert"
	}
	stem := "assert"
	if loc.Filename != "" {
		stem = strings.TrimSuffix(filepath.Base(loc.Filename), filepath.Ext(loc.Filename))
	}
	return fmt.Sprintf("assert_%s_%d", stem, loc.Line)
}
