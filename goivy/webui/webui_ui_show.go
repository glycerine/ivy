package webui

// Full port of ivy_show.py — entry point functions for launching the
// verification UI and checking modules.

import (
	"fmt"
	goivy "github.com/glycerine/ivy/goivy"
	"os"
)

// ShowVerification is the main entry point for showing verification results.
// It initializes the module, creates the isolate, and launches the UI.
// (Python: check_module in ivy_show.py).
func ShowVerification(cfg *goivy.Config, filePath string) (*Session, error) {
	if filePath == "" {
		return nil, fmt.Errorf("empty file path")
	}

	// Python: ivy_init.open_read(sys.argv[1])
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Python: iu.set_parameters({'show_compiled':'true'})
	cfg.IsolateCfg.ShowCompiled = true

	// Python: ivy_init.source_file(...) + check_module()
	// LoadFileContent runs the full pipeline: parse → compile (with
	// CreateIsolate) → extract sorts/symbols → concept sessions → AG.
	sess := NewSession(cfg, "show_"+filePath)
	if err := sess.LoadFileContent(filePath, content); err != nil {
		return nil, fmt.Errorf("failed to compile file: %w", err)
	}

	return sess, nil
}

// LaunchUI starts the web-based verification UI for an analysis graph.
// This replaces the Python tk_ui.ui_main_loop.
// (Python: ui_main_loop in ivy_ui.py).
//
// not sure this is goroutine safe... or used for that matter.
// might just be left-over from the port from python.
// ivyweb uses NewServer() directly... but also
// does not use a Session, which it might want to,
// so keep around.
func LaunchUI(cfg *goivy.Config, sess *Session, addr string) (*Server, error) {
	if sess == nil {
		return nil, fmt.Errorf("nil session")
	}

	be := NewGoBackend(cfg)
	be.mu.Lock()
	be.sessions[sess.ID] = sess
	be.mu.Unlock()

	srv := NewServer(cfg, addr, be)
	return srv, nil
}

// DiagnoseMode: now on module.Config.Diagnose for multi-tenancy.

// CompileKwargs holds extra keyword arguments for compilation.
// (Python: compile_kwargs = {'ext':'ext'} in ivy_ui_cti.py).
var CompileKwargs = map[string]string{
	"ext": "ext",
}

// CheckModuleAndShow loads an Ivy file, checks the module, and
// shows the result in the UI.
// (Python: main() in ivy_show.py).
func CheckModuleAndShow(cfg *goivy.Config, filePath string, addr string) error {
	sess, err := ShowVerification(cfg, filePath)
	if err != nil {
		return err
	}

	srv, err := LaunchUI(cfg, sess, addr)
	if err != nil {
		return err
	}

	// Python: ui_main_loop() — blocks until shutdown.
	return srv.Start()
}
