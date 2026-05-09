package main

// goldweb is the first browser-conformance harness for Go Ivy.
//
// The intentional narrow waist is WASI stdout/stderr:
//   - the browser worker runs Go Ivy compiled to wasip1 and Z3 compiled to wasm;
//   - every WASI fd_write byte for stdout and stderr is sent to this server;
//   - the server compares XTRACE lines against Python ivy_check, while still
//     retaining and exposing non-XTRACE output for diagnosis.
//
// The Go Ivy wasm command boundary is deliberately small and not the
// cmd/goivy_check CLI. The CLI is only documentation for parameter semantics.
// The browser wasm module is expected to export:
//   - goivy_check_alloc(n uint32) uint32
//   - goivy_check_free(ptr uint32, n uint32)
//   - goivy_check_run(specPtr, specLen, metaPtr, metaLen uint32) int32
//
// meta is JSON: {"filename":"browser_input.ivy","params":{"isolate":"x"}}.
// The command writes its normal output to stdout/stderr; goldweb does not use a
// special trace callback.

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/glycerine/ivy/goivy/xtracer"
	"github.com/gorilla/websocket"
)

const (
	writeWait    = 10 * time.Minute
	pongWait     = 60 * time.Second
	pingPeriod   = 30 * time.Second
	readLimit    = 64 << 20
	sendQueueLen = 8192
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type app struct {
	listen     string
	goivyRoot  string
	webvueDir  string
	staticDir  string
	workerDir  string
	includeDir string
	goivyWasm  string
	ivyCheck   string
	httpServer *http.Server

	mu      sync.Mutex
	clients map[*wsClient]bool
	jobs    map[string]*job

	startupReq     *goivyCheckRequest
	startupStarted bool
}

type wsClient struct {
	app    *app
	conn   *websocket.Conn
	send   chan []byte
	done   chan struct{}
	closed sync.Once
}

type wsEnvelope struct {
	Type        string            `json:"type"`
	ID          string            `json:"id,omitempty"`
	Command     string            `json:"command,omitempty"`
	Filename    string            `json:"filename,omitempty"`
	Spec        string            `json:"spec,omitempty"`
	Params      map[string]string `json:"params,omitempty"`
	FD          int               `json:"fd,omitempty"`
	Data        string            `json:"data,omitempty"`
	Code        int               `json:"code,omitempty"`
	Error       string            `json:"error,omitempty"`
	Status      string            `json:"status,omitempty"`
	Message     string            `json:"message,omitempty"`
	XTraceIndex *int              `json:"xtrace_index,omitempty"`
}

type goivyCheckRequest struct {
	Filename   string            `json:"filename"`
	Spec       string            `json:"spec"`
	Params     map[string]string `json:"params"`
	SourcePath string            `json:"-"`
}

type includeTree struct {
	Root  string        `json:"root"`
	Files []includeFile `json:"files"`
}

type includeFile struct {
	Path string `json:"path"`
	Data string `json:"data"`
}

type job struct {
	id         string
	filename   string
	sourcePath string
	spec       string
	params     map[string]string
	app        *app
	client     *wsClient

	pyLines        chan lineEvent
	browserLines   chan lineEvent
	browserBuf     streamLineBuffer
	compareDone    chan struct{}
	browserLog     *os.File
	pyLog          *os.File
	browserLogPath string
	pyLogPath      string

	mu           sync.Mutex
	status       string
	message      string
	xtraceCount  int
	mismatchAt   *int
	browserCode  int
	pyErr        string
	tail         []string
	cancelPy     context.CancelFunc
	closeBrowser sync.Once
}

type lineEvent struct {
	FD   int
	Line string
}

type streamLineBuffer struct {
	partial string
	fd      int
}

type jobSnapshot struct {
	ID          string            `json:"id"`
	Filename    string            `json:"filename"`
	Params      map[string]string `json:"params,omitempty"`
	Status      string            `json:"status"`
	Message     string            `json:"message,omitempty"`
	XTraceCount int               `json:"xtrace_count"`
	MismatchAt  *int              `json:"mismatch_at,omitempty"`
	BrowserCode int               `json:"browser_code,omitempty"`
	PythonError string            `json:"python_error,omitempty"`
	BrowserLog  string            `json:"browser_log,omitempty"`
	PythonLog   string            `json:"python_log,omitempty"`
	Tail        []string          `json:"tail,omitempty"`
}

func main() {
	rootDefault := defaultGoivyRoot()
	listen := flag.String("listen", "127.0.0.1:8998", "address for the goldweb HTTP/WebSocket server")
	root := flag.String("root", rootDefault, "goivy source root")
	includeDir := flag.String("include-dir", defaultIncludeDir(rootDefault), "Ivy standard-library include directory to mirror into the browser WASI filesystem")
	goivyWasm := flag.String("goivy-wasm", filepath.Join(rootDefault, "webvue", "static", "goivy-check-wasip1.wasm"), "Go Ivy wasip1 wasm file to serve at /goivy-check.wasm")
	ivyCheck := flag.String("ivy-check", "ivy_check", "Python ivy_check executable")
	flag.Parse()

	var startupReq *goivyCheckRequest
	if flag.NArg() > 0 {
		req, err := goivyCheckRequestFromPath(flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}
		startupReq = req
	}

	a := newApp(*listen, *root, *includeDir, *goivyWasm, *ivyCheck, startupReq)
	if err := a.listenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func newApp(listen, root, includeDir, goivyWasm, ivyCheck string, startupReq *goivyCheckRequest) *app {
	root = filepath.Clean(root)
	includeDir = filepath.Clean(includeDir)
	webvueDir := filepath.Join(root, "webvue")
	return &app{
		listen:     listen,
		goivyRoot:  root,
		webvueDir:  webvueDir,
		staticDir:  filepath.Join(webvueDir, "static"),
		workerDir:  filepath.Join(webvueDir, "src", "workers"),
		includeDir: includeDir,
		goivyWasm:  goivyWasm,
		ivyCheck:   ivyCheck,
		clients:    make(map[*wsClient]bool),
		jobs:       make(map[string]*job),
		startupReq: startupReq,
	}
}

func goivyCheckRequestFromPath(path string) (*goivyCheckRequest, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return nil, err
	}
	req := &goivyCheckRequest{
		Filename:   abs,
		SourcePath: abs,
		Spec:       string(data),
		Params:     make(map[string]string),
	}
	if err := normalizeGoivyCheckRequest(req); err != nil {
		return nil, err
	}
	return req, nil
}

func normalizeGoivyCheckRequest(req *goivyCheckRequest) error {
	if strings.TrimSpace(req.Spec) == "" {
		return errors.New("missing spec")
	}
	if req.Filename == "" {
		req.Filename = "browser_input.ivy"
	}
	if !strings.HasSuffix(req.Filename, ".ivy") {
		req.Filename += ".ivy"
	}
	if req.Params == nil {
		req.Params = make(map[string]string)
	}
	return nil
}

func (a *app) listenAndServe() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.serveIndex)
	mux.HandleFunc("/ws", a.serveWS)
	mux.HandleFunc("/goivy_check", a.handleGoivyCheck)
	mux.HandleFunc("/jobs/", a.handleJob)
	mux.HandleFunc("/goivy-check.wasm", serveFile(a.goivyWasm, "application/wasm"))
	mux.HandleFunc("/ivy-include-tree.json", a.serveIncludeTree)
	mux.HandleFunc("/z3-471-api.js", serveFile(filepath.Join(a.staticDir, "z3-471-api.js"), "text/javascript; charset=utf-8"))
	mux.HandleFunc("/z3-471-api.wasm", serveFile(filepath.Join(a.staticDir, "z3-471-api.wasm"), "application/wasm"))
	mux.HandleFunc("/src/workers/smtZ3Imports.js", serveFile(filepath.Join(a.workerDir, "smtZ3Imports.js"), "text/javascript; charset=utf-8"))
	mux.Handle("/node_modules/@bjorn3/browser_wasi_shim/", http.StripPrefix(
		"/node_modules/@bjorn3/browser_wasi_shim/",
		http.FileServer(http.Dir(filepath.Join(a.webvueDir, "node_modules", "@bjorn3", "browser_wasi_shim"))),
	))

	a.httpServer = &http.Server{
		Addr:              a.listen,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("goldweb listening on http://%s", a.listen)
	log.Printf("POST JSON to http://%s/goivy_check to start a browser-backed check", a.listen)
	log.Printf("serving Ivy include tree from %s as browser path include/", a.includeDir)
	if a.startupReq != nil {
		log.Printf("will submit %s to the first browser websocket client", a.startupReq.Filename)
	}
	return a.httpServer.ListenAndServe()
}

func (a *app) serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = io.WriteString(w, indexHTML)
}

func serveFile(path, contentType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if _, err := os.Stat(path); err != nil {
			http.Error(w, fmt.Sprintf("missing asset %s: %v", path, err), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", "no-store")
		http.ServeFile(w, r, path)
	}
}

func (a *app) serveIncludeTree(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	tree, err := readIncludeTree(a.includeDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if r.Method == http.MethodHead {
		return
	}
	_ = json.NewEncoder(w).Encode(tree)
}

func readIncludeTree(root string) (*includeTree, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(absRoot)
	if err != nil {
		return nil, fmt.Errorf("stat Ivy include dir %s: %w", absRoot, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("Ivy include path is not a directory: %s", absRoot)
	}

	tree := &includeTree{Root: absRoot}
	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(absRoot, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		tree.Files = append(tree.Files, includeFile{
			Path: filepath.ToSlash(rel),
			Data: string(data),
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("read Ivy include tree %s: %w", absRoot, err)
	}
	sort.Slice(tree.Files, func(i, j int) bool {
		return tree.Files[i].Path < tree.Files[j].Path
	})
	return tree, nil
}

func (a *app) serveWS(w http.ResponseWriter, r *http.Request) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		return
	}
	c := &wsClient{
		app:  a,
		conn: conn,
		send: make(chan []byte, sendQueueLen),
		done: make(chan struct{}),
	}
	a.register(c)
	go c.writePump()
	c.readPump()
}

func (a *app) register(c *wsClient) {
	a.mu.Lock()
	a.clients[c] = true
	n := len(a.clients)
	a.mu.Unlock()
	log.Printf("browser websocket connected: %s; clients=%d", c.conn.RemoteAddr(), n)
	_ = c.sendJSON(wsEnvelope{Type: "hello", Status: "ready", Message: "goldweb connected"})
	a.maybeStartStartupJob(c)
}

func (a *app) unregister(c *wsClient) {
	c.close()
	a.mu.Lock()
	delete(a.clients, c)
	n := len(a.clients)
	a.mu.Unlock()
	log.Printf("browser websocket disconnected: %s; clients=%d", c.conn.RemoteAddr(), n)
}

func (a *app) firstClient() *wsClient {
	a.mu.Lock()
	defer a.mu.Unlock()
	for c := range a.clients {
		return c
	}
	return nil
}

func (a *app) addJob(j *job) {
	a.mu.Lock()
	a.jobs[j.id] = j
	a.mu.Unlock()
}

func (a *app) lookupJob(id string) *job {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.jobs[id]
}

func (a *app) maybeStartStartupJob(c *wsClient) {
	a.mu.Lock()
	if a.startupReq == nil || a.startupStarted {
		a.mu.Unlock()
		return
	}
	req := *a.startupReq
	a.startupStarted = true
	a.mu.Unlock()

	j, err := a.startGoivyCheck(c, req)
	if err != nil {
		log.Printf("could not start initial goivy_check for %s: %v", req.Filename, err)
		return
	}
	log.Printf("submitted initial goivy_check job %s for %s", j.id, j.filename)
}

func (a *app) handleWSMessage(c *wsClient, data []byte) {
	var msg wsEnvelope
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("bad websocket JSON from browser: %v", err)
		return
	}
	switch msg.Type {
	case "ready":
		return
	case "stream":
		j := a.lookupJob(msg.ID)
		if j == nil {
			log.Printf("stream for unknown job %q", msg.ID)
			return
		}
		j.addBrowserData(msg.FD, msg.Data)
	case "done":
		j := a.lookupJob(msg.ID)
		if j == nil {
			log.Printf("done for unknown job %q", msg.ID)
			return
		}
		j.setBrowserDone(msg.Code)
	case "error":
		j := a.lookupJob(msg.ID)
		if j == nil {
			log.Printf("error for unknown job %q: %s", msg.ID, msg.Error)
			return
		}
		j.fail("browser error: " + msg.Error)
		j.setBrowserDone(msg.Code)
	default:
		log.Printf("unknown websocket message type %q", msg.Type)
	}
}

func (a *app) handleGoivyCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"status":  "ready",
			"message": "POST a JSON body to start a browser-backed goivy_check run.",
			"example": goivyCheckRequest{
				Filename: "browser_input.ivy",
				Spec:     "#lang ivy1.7\n",
				Params:   map[string]string{"isolate": "name"},
			},
		})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	var req goivyCheckRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 128<<20)).Decode(&req); err != nil {
		http.Error(w, "bad JSON request: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := normalizeGoivyCheckRequest(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c := a.firstClient()
	if c == nil {
		http.Error(w, "no browser websocket client connected; open / first", http.StatusServiceUnavailable)
		return
	}

	j, err := a.startGoivyCheck(c, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	writeJSON(w, http.StatusAccepted, j.snapshot())
}

func (a *app) startGoivyCheck(c *wsClient, req goivyCheckRequest) (*job, error) {
	if err := normalizeGoivyCheckRequest(&req); err != nil {
		return nil, err
	}

	j := newJob(a, c, req)
	if err := j.openXTraceLogs(); err != nil {
		return nil, err
	}
	a.addJob(j)
	go j.compareLoop()
	go j.runPythonIvyCheck()

	if err := c.sendJSON(wsEnvelope{
		Type:     "command",
		ID:       j.id,
		Command:  "goivy_check",
		Filename: j.filename,
		Spec:     j.spec,
		Params:   j.params,
	}); err != nil {
		j.fail("send browser command: " + err.Error())
		return nil, err
	}

	return j, nil
}

func (a *app) handleJob(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/jobs/")
	if id == "" || strings.Contains(id, "/") {
		http.NotFound(w, r)
		return
	}
	j := a.lookupJob(id)
	if j == nil {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, j.snapshot())
}

func newJob(a *app, c *wsClient, req goivyCheckRequest) *job {
	return &job{
		id:           newID(),
		filename:     req.Filename,
		sourcePath:   req.SourcePath,
		spec:         req.Spec,
		params:       cloneStringMap(req.Params),
		app:          a,
		client:       c,
		pyLines:      make(chan lineEvent, 8192),
		browserLines: make(chan lineEvent, 8192),
		compareDone:  make(chan struct{}),
		status:       "running",
	}
}

func (j *job) openXTraceLogs() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	browserPath := filepath.Join(cwd, "browser.xtrace.log")
	pyPath := filepath.Join(cwd, "py.xtrace.log")
	browserLog, err := os.Create(browserPath)
	if err != nil {
		return fmt.Errorf("create %s: %w", browserPath, err)
	}
	pyLog, err := os.Create(pyPath)
	if err != nil {
		_ = browserLog.Close()
		return fmt.Errorf("create %s: %w", pyPath, err)
	}

	j.browserLog = browserLog
	j.pyLog = pyLog
	j.browserLogPath = browserPath
	j.pyLogPath = pyPath
	log.Printf("writing browser XTRACE log to %s", browserPath)
	log.Printf("writing python XTRACE log to %s", pyPath)
	return nil
}

func (j *job) closeXTraceLogs() {
	if j.browserLog != nil {
		if err := j.browserLog.Close(); err != nil {
			log.Printf("close %s: %v", j.browserLogPath, err)
		}
		j.browserLog = nil
	}
	if j.pyLog != nil {
		if err := j.pyLog.Close(); err != nil {
			log.Printf("close %s: %v", j.pyLogPath, err)
		}
		j.pyLog = nil
	}
}

func (j *job) runPythonIvyCheck() {
	path := j.sourcePath
	if path == "" {
		tmp, err := os.MkdirTemp("", "goldweb-ivy-*")
		if err != nil {
			j.fail("create temp dir: " + err.Error())
			close(j.pyLines)
			return
		}
		defer os.RemoveAll(tmp)

		filename := filepath.Base(j.filename)
		if filename == "." || filename == string(filepath.Separator) {
			filename = "browser_input.ivy"
		}
		path = filepath.Join(tmp, filename)
		if err := os.WriteFile(path, []byte(j.spec), 0o600); err != nil {
			j.fail("write temp spec: " + err.Error())
			close(j.pyLines)
			return
		}
	}

	args := paramsAsArgs(j.params)
	args = append(args, path)
	ctx, cancel := context.WithCancel(context.Background())
	j.setPythonCancel(cancel)
	defer func() {
		j.setPythonCancel(nil)
		cancel()
	}()
	cmd := exec.CommandContext(ctx, j.app.ivyCheck, args...)
	cmd.Dir = j.app.goivyRoot

	pr, pw := io.Pipe()
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		j.fail("start python ivy_check: " + err.Error())
		_ = pw.Close()
		close(j.pyLines)
		return
	}

	go func() {
		waitErr := cmd.Wait()
		if waitErr != nil {
			j.setPythonErr(waitErr.Error())
		}
		_ = pw.Close()
	}()

	scanner := bufio.NewScanner(pr)
	scanner.Buffer(make([]byte, 0, 16<<20), 1<<30)
	for scanner.Scan() {
		j.pyLines <- lineEvent{FD: 1, Line: scanner.Text() + "\n"}
	}
	if err := scanner.Err(); err != nil {
		j.setPythonErr("scan python output: " + err.Error())
	}
	close(j.pyLines)
}

func (j *job) compareLoop() {
	defer j.closeXTraceLogs()
	defer close(j.compareDone)
	for {
		browser, browserOK := j.nextXTrace("browser", j.browserLines)
		python, pythonOK := j.nextXTrace("python", j.pyLines)

		switch {
		case !browserOK && !pythonOK:
			j.finishMatched()
			return
		case !browserOK:
			j.mismatch("browser ended before python", "", python)
			return
		case !pythonOK:
			j.mismatch("python ended before browser", browser, "")
			return
		}

		if browser != python {
			j.mismatch("xtrace divergence", browser, python)
			return
		}
		j.mu.Lock()
		j.xtraceCount++
		j.mu.Unlock()
	}
}

func (j *job) nextXTrace(side string, ch <-chan lineEvent) (string, bool) {
	for ev := range ch {
		line := xtracer.NormalizeLine(ev.Line)
		j.writeXTraceLogLine(side, line)
		j.remember(side, line)
		if strings.HasPrefix(line, "XTRACE:") {
			return line, true
		}
	}
	return "", false
}

func (j *job) writeXTraceLogLine(side, line string) {
	var f *os.File
	var path string
	switch side {
	case "browser":
		f = j.browserLog
		path = j.browserLogPath
	case "python":
		f = j.pyLog
		path = j.pyLogPath
	default:
		return
	}
	if f == nil {
		return
	}
	if !strings.HasSuffix(line, "\n") {
		line += "\n"
	}
	if _, err := io.WriteString(f, line); err != nil {
		j.fail(fmt.Sprintf("write %s: %v", path, err))
	}
}

func (j *job) addBrowserData(fd int, data string) {
	lines := j.browserBuf.add(fd, data)
	for _, line := range lines {
		j.browserLines <- line
	}
}

func (j *job) setBrowserDone(code int) {
	j.mu.Lock()
	j.browserCode = code
	j.mu.Unlock()
	lines := j.browserBuf.flush()
	for _, line := range lines {
		j.browserLines <- line
	}
	j.closeBrowser.Do(func() {
		close(j.browserLines)
	})
}

func (j *job) setPythonErr(msg string) {
	j.mu.Lock()
	if j.pyErr == "" {
		j.pyErr = msg
	}
	j.mu.Unlock()
}

func (j *job) setPythonCancel(cancel context.CancelFunc) {
	j.mu.Lock()
	j.cancelPy = cancel
	j.mu.Unlock()
}

func (j *job) cancelPython() {
	j.mu.Lock()
	cancel := j.cancelPy
	j.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (j *job) fail(msg string) {
	j.mu.Lock()
	if j.status == "running" {
		j.status = "error"
		j.message = msg
	}
	j.mu.Unlock()
}

func (j *job) finishMatched() {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.status != "running" {
		return
	}
	if j.pyErr != "" {
		j.status = "error"
		j.message = "python ivy_check failed: " + j.pyErr
		return
	}
	if j.browserCode != 0 {
		j.status = "error"
		j.message = fmt.Sprintf("browser goivy_check returned %d", j.browserCode)
		return
	}
	j.status = "matched"
	j.message = "browser Go Ivy and Python ivy_check XTRACE streams matched"
}

func (j *job) mismatch(reason, browser, python string) {
	var message string
	var mismatchAt int
	j.mu.Lock()
	if j.status == "running" {
		mismatchAt = j.xtraceCount
		j.mismatchAt = &mismatchAt
		j.status = "mismatch"
		j.message = fmt.Sprintf("%s at XTRACE index %d after %d matching XTRACE lines\nbrowser: %.500s\npython : %.500s", reason, mismatchAt, j.xtraceCount, strings.TrimSpace(browser), strings.TrimSpace(python))
		message = j.message
	} else {
		message = j.message
		if j.mismatchAt != nil {
			mismatchAt = *j.mismatchAt
		} else {
			mismatchAt = j.xtraceCount
		}
	}
	j.mu.Unlock()
	j.cancelPython()
	_ = j.client.sendJSON(wsEnvelope{Type: "cancel", ID: j.id, Message: message, XTraceIndex: &mismatchAt})
}

func (j *job) remember(side, line string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	entry := fmt.Sprintf("%s: %s", side, strings.TrimRight(line, "\n"))
	j.tail = append(j.tail, entry)
	if len(j.tail) > 200 {
		copy(j.tail, j.tail[len(j.tail)-200:])
		j.tail = j.tail[:200]
	}
}

func (j *job) snapshot() jobSnapshot {
	j.mu.Lock()
	defer j.mu.Unlock()
	return jobSnapshot{
		ID:          j.id,
		Filename:    j.filename,
		Params:      cloneStringMap(j.params),
		Status:      j.status,
		Message:     j.message,
		XTraceCount: j.xtraceCount,
		MismatchAt:  copyIntPtr(j.mismatchAt),
		BrowserCode: j.browserCode,
		PythonError: j.pyErr,
		BrowserLog:  j.browserLogPath,
		PythonLog:   j.pyLogPath,
		Tail:        append([]string(nil), j.tail...),
	}
}

func (b *streamLineBuffer) add(fd int, data string) []lineEvent {
	var lines []lineEvent
	if b.partial == "" {
		b.fd = fd
	}
	s := b.partial + data
	for {
		idx := strings.IndexByte(s, '\n')
		if idx < 0 {
			b.partial = s
			if s != "" {
				b.fd = fd
			}
			return lines
		}
		lines = append(lines, lineEvent{FD: fd, Line: s[:idx+1]})
		s = s[idx+1:]
	}
}

func (b *streamLineBuffer) flush() []lineEvent {
	if b.partial == "" {
		return nil
	}
	line := b.partial
	fd := b.fd
	if fd == 0 {
		fd = 1
	}
	b.partial = ""
	b.fd = 0
	return []lineEvent{{FD: fd, Line: line}}
}

func (c *wsClient) readPump() {
	defer func() {
		c.app.unregister(c)
		_ = c.conn.Close()
	}()
	c.conn.SetReadLimit(readLimit)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket read error: %v", err)
			}
			return
		}
		select {
		case <-c.done:
			return
		default:
			c.app.handleWSMessage(c, message)
		}
	}
}

func (c *wsClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.close()
		_ = c.conn.Close()
	}()
	for {
		select {
		case message := <-c.send:
			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.done:
			return
		}
	}
}

func (c *wsClient) write(mt int, payload []byte) error {
	_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	return c.conn.WriteMessage(mt, payload)
}

func (c *wsClient) sendJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	select {
	case c.send <- data:
		return nil
	case <-c.done:
		return errors.New("websocket client closed")
	case <-time.After(5 * time.Second):
		return errors.New("websocket send queue blocked")
	}
}

func (c *wsClient) close() {
	c.closed.Do(func() {
		close(c.done)
	})
}

func paramsAsArgs(params map[string]string) []string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	args := make([]string, 0, len(keys))
	for _, k := range keys {
		args = append(args, k+"="+params[k])
	}
	return args
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func copyIntPtr(in *int) *int {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func newID() string {
	var buf [12]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf[:])
}

func defaultIncludeDir(root string) string {
	parent := filepath.Dir(filepath.Clean(root))
	candidates := []string{
		filepath.Join(parent, "ivy-lang-examples", "ivy", "include"),
		filepath.Join(parent, "pyivy", "ivy", "ivy", "include"),
		filepath.Join(root, "ivy-lang-examples", "ivy", "include"),
	}
	if home := os.Getenv("HOME"); home != "" {
		candidates = append(candidates,
			filepath.Join(home, "ivy", "ivy-lang-examples", "ivy", "include"),
			filepath.Join(home, "ivy", "pyivy", "ivy", "ivy", "include"),
		)
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	return filepath.Join(parent, "ivy-lang-examples", "ivy", "include")
}

func defaultGoivyRoot() string {
	if wd, err := os.Getwd(); err == nil {
		if looksLikeGoivyRoot(wd) {
			return wd
		}
		if filepath.Base(wd) == "goldweb" && looksLikeGoivyRoot(filepath.Dir(wd)) {
			return filepath.Dir(wd)
		}
	}
	_, file, _, ok := runtime.Caller(0)
	if ok {
		dir := filepath.Dir(file)
		if filepath.Base(dir) == "goldweb" && looksLikeGoivyRoot(filepath.Dir(dir)) {
			return filepath.Dir(dir)
		}
	}
	return "."
}

func looksLikeGoivyRoot(dir string) bool {
	_, err1 := os.Stat(filepath.Join(dir, "go.mod"))
	_, err2 := os.Stat(filepath.Join(dir, "webvue", "static", "z3-471-api.js"))
	return err1 == nil && err2 == nil
}

const indexHTML = `<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>goldweb</title>
  <style>
    body {
      color: #1f2933;
      font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      margin: 2rem;
      max-width: 78rem;
    }
    h1 { font-size: 1.5rem; margin: 0 0 0.5rem; }
    .topline {
      align-items: center;
      display: flex;
      flex-wrap: wrap;
      gap: 0.75rem;
      margin-bottom: 1rem;
    }
    #status { font-weight: 700; }
    .metric {
      background: #eef2f7;
      border: 1px solid #d9e2ec;
      display: inline-flex;
      gap: 0.35rem;
      padding: 0.2rem 0.45rem;
    }
    .metric > span {
      display: inline-block;
      font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, "Liberation Mono", monospace;
      font-variant-numeric: tabular-nums;
      min-width: 15ch;
    }
    .metric #job-id { min-width: 24ch; }
    .log-shell {
      border: 1px solid #bcccdc;
    }
    .log-head {
      align-items: center;
      background: #f0f4f8;
      border-bottom: 1px solid #bcccdc;
      display: flex;
      justify-content: space-between;
      padding: 0.45rem 0.6rem;
    }
    .log-actions {
      align-items: center;
      display: flex;
      gap: 0.6rem;
    }
    button {
      background: #ffffff;
      border: 1px solid #9fb3c8;
      color: #102a43;
      cursor: pointer;
      font: inherit;
      padding: 0.2rem 0.5rem;
    }
    button:active { background: #d9e2ec; }
    .log-head strong { font-size: 0.9rem; }
    .log-head span { color: #52606d; font-size: 0.8rem; }
    #stream-log {
      background: #102a43;
      border: 0;
      box-sizing: border-box;
      color: #f0f4f8;
      display: block;
      font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, "Liberation Mono", monospace;
      font-size: 0.82rem;
      height: 64vh;
      line-height: 1.35;
      margin: 0;
      outline: none;
      overflow: auto;
      padding: 0.75rem;
      white-space: pre;
      user-select: text;
      width: 100%;
    }
    #line-number {
      color: #52606d;
      font-size: 0.9rem;
      padding: 0.45rem 0.6rem;
    }
  </style>
</head>
<body>
  <h1>goldweb</h1>
  <div class="topline">
    <span id="status">connecting</span>
    <span class="metric">job <span id="job-id">none</span></span>
    <span class="metric">stdout <span id="stdout-bytes">0</span>B</span>
    <span class="metric">stderr <span id="stderr-bytes">0</span>B</span>
    <span class="metric">XTRACE <span id="xtrace-count">0</span></span>
  </div>
  <div class="log-shell">
    <div class="log-head">
      <strong>Browser stdout/stderr sent to goldweb</strong>
      <div class="log-actions">
        <button id="copy-log" type="button">Copy visible log</button>
        <span>rolling local window; full stream goes over the websocket</span>
      </div>
    </div>
    <pre id="stream-log" tabindex="0"></pre>
    <div id="line-number">last line number : 0</div>
  </div>

  <script type="text/plain" id="worker-source">
const textEncoder = new TextEncoder();

function requireExport(exports, name) {
  const fn = exports[name];
  if (typeof fn !== 'function') {
    throw new Error('Go Ivy wasm is missing export ' + name);
  }
  return fn;
}

function writeBytes(exports, memory, bytes) {
  const alloc = requireExport(exports, 'goivy_check_alloc');
  const ptr = alloc(bytes.length) >>> 0;
  new Uint8Array(memory.buffer, ptr, bytes.length).set(bytes);
  return ptr;
}

function includeDirectoryFromTree(wasiShim, tree) {
  const root = { dirs: new Map(), files: new Map() };

  function childDir(parent, name) {
    let dir = parent.dirs.get(name);
    if (!dir) {
      dir = { dirs: new Map(), files: new Map() };
      parent.dirs.set(name, dir);
    }
    return dir;
  }

  for (const file of tree.files || []) {
    const parts = String(file.path || '').split('/').filter(Boolean);
    if (!parts.length) {
      continue;
    }
    const filename = parts.pop();
    let dir = root;
    for (const part of parts) {
      dir = childDir(dir, part);
    }
    dir.files.set(filename, textEncoder.encode(String(file.data || '')));
  }

  function materialize(dir) {
    const entries = [];
    for (const [name, child] of Array.from(dir.dirs.entries()).sort()) {
      entries.push([name, materialize(child)]);
    }
    for (const [name, bytes] of Array.from(dir.files.entries()).sort()) {
      entries.push([name, new wasiShim.File(bytes, { readonly: true })]);
    }
    return new wasiShim.Directory(entries);
  }

  return materialize(root);
}

async function loadIncludeDirectory(wasiShim, assetBaseURL) {
  const response = await fetch(assetBaseURL + '/ivy-include-tree.json', { cache: 'no-store' });
  if (!response.ok) {
    throw new Error('could not fetch /ivy-include-tree.json: HTTP ' + response.status);
  }
  return includeDirectoryFromTree(wasiShim, await response.json());
}

self.onmessage = async (event) => {
  const command = event.data;
  let z3;
  try {
    const assetBaseURL = self.location.origin;
    const z3Imports = await import(assetBaseURL + '/src/workers/smtZ3Imports.js');
    const wasiShim = await import(assetBaseURL + '/node_modules/@bjorn3/browser_wasi_shim/dist/index.js');
    importScripts(assetBaseURL + '/z3-471-api.js');
    if (typeof initZ3 !== 'function') {
      throw new Error('z3-471-api.js did not expose initZ3');
    }
    z3 = await initZ3({
      print(text) {
        self.postMessage({ type: 'stream', fd: 1, data: String(text) + '\n' });
      },
      printErr(text) {
        self.postMessage({ type: 'stream', fd: 2, data: String(text) + '\n' });
      },
      locateFile(file) {
        if (file === 'z3-api.wasm') {
          return assetBaseURL + '/z3-471-api.wasm';
        }
        return assetBaseURL + '/' + file;
      }
    });

    const stdoutDecoder = new TextDecoder('utf-8', { fatal: false });
    const stderrDecoder = new TextDecoder('utf-8', { fatal: false });
    const emitStdout = (data) => {
      self.postMessage({ type: 'stream', fd: 1, data: stdoutDecoder.decode(data, { stream: true }) });
    };
    const emitStderr = (data) => {
      self.postMessage({ type: 'stream', fd: 2, data: stderrDecoder.decode(data, { stream: true }) });
    };
    const includeDir = await loadIncludeDirectory(wasiShim, assetBaseURL);
    const wasi = new wasiShim.WASI(
      ['goivy_check_wasip1'],
      [],
      [
        new wasiShim.OpenFile(new wasiShim.File(new Uint8Array())),
        new wasiShim.ConsoleStdout(emitStdout),
        new wasiShim.ConsoleStdout(emitStderr),
        new wasiShim.PreopenDirectory('.', [['include', includeDir]]),
      ],
    );

    let wasmMemory;
    const imports = {
      smt_z3: z3Imports.createSmtZ3Imports({ z3, getGoMemory: () => wasmMemory }),
      wasi_snapshot_preview1: wasi.wasiImport,
    };

    const response = await fetch(assetBaseURL + '/goivy-check.wasm', { cache: 'no-store' });
    if (!response.ok) {
      throw new Error('could not fetch /goivy-check.wasm: HTTP ' + response.status);
    }
    const bytes = await response.arrayBuffer();
    const result = await WebAssembly.instantiate(bytes, imports);
    const instance = result.instance;
    wasmMemory = instance.exports.memory;
    if (!wasmMemory) {
      throw new Error('Go Ivy wasm does not export memory');
    }

    const startCode = wasi.start(instance);
    if (startCode !== 0) {
      throw new Error('Go Ivy wasm _start exited with code ' + startCode);
    }

    const run = requireExport(instance.exports, 'goivy_check_run');
    const free = instance.exports.goivy_check_free;
    const specBytes = textEncoder.encode(command.spec || '');
    const metaBytes = textEncoder.encode(JSON.stringify({
      filename: command.filename || 'browser_input.ivy',
      params: command.params || {}
    }));
    const specPtr = writeBytes(instance.exports, wasmMemory, specBytes);
    const metaPtr = writeBytes(instance.exports, wasmMemory, metaBytes);
    let code = 0;
    try {
      code = run(specPtr, specBytes.length, metaPtr, metaBytes.length) | 0;
    } finally {
      if (typeof free === 'function') {
        free(specPtr, specBytes.length);
        free(metaPtr, metaBytes.length);
      }
    }
    self.postMessage({ type: 'done', code });
  } catch (error) {
    self.postMessage({
      type: 'error',
      error: error && error.stack ? error.stack : String(error)
    });
  }
};
  </script>

  <script>
const statusEl = document.getElementById('status');
const jobIdEl = document.getElementById('job-id');
const stdoutBytesEl = document.getElementById('stdout-bytes');
const stderrBytesEl = document.getElementById('stderr-bytes');
const xtraceCountEl = document.getElementById('xtrace-count');
const streamLogEl = document.getElementById('stream-log');
const copyLogEl = document.getElementById('copy-log');
const lineNumberEl = document.getElementById('line-number');
const workerSource = document.getElementById('worker-source').textContent;
const workers = new Map();
const uiLineBuffers = new Map();
const uiTextEncoder = new TextEncoder();
const maxVisibleLogBytes = 1024 * 1024;
let stdoutBytes = 0;
let stderrBytes = 0;
let xtraceCount = 0;
let visibleLineNumber = 0;

function log(message) {
  appendVisibleLog('[goldweb] ' + message + '\n');
}

function appendVisibleLog(text) {
  const distanceFromBottom = streamLogEl.scrollHeight - streamLogEl.scrollTop - streamLogEl.clientHeight;
  const shouldTail = distanceFromBottom < 8;
  streamLogEl.textContent += text;
  if (streamLogEl.textContent.length > maxVisibleLogBytes) {
    streamLogEl.textContent = streamLogEl.textContent.slice(streamLogEl.textContent.length - maxVisibleLogBytes);
  }
  if (shouldTail) {
    streamLogEl.scrollTop = streamLogEl.scrollHeight;
  }
}

function updateCounters() {
  stdoutBytesEl.textContent = String(stdoutBytes);
  stderrBytesEl.textContent = String(stderrBytes);
  xtraceCountEl.textContent = String(xtraceCount);
  lineNumberEl.textContent = 'last line number: ' + visibleLineNumber;
}

function showStream(fd, data) {
  const nbytes = uiTextEncoder.encode(data).length;
  if (fd === 2) {
    stderrBytes += nbytes;
  } else {
    stdoutBytes += nbytes;
  }

  const label = fd === 2 ? 'stderr' : 'stdout';
  const key = String(fd);
  let pending = (uiLineBuffers.get(key) || '') + data;
  let idx = pending.indexOf('\n');
  while (idx >= 0) {
    const line = pending.slice(0, idx + 1);
    if (line.startsWith('XTRACE:')) {
      xtraceCount += 1;
    }
    appendVisibleLog('[' + label + '] ' + line);
    visibleLineNumber += 1;
    pending = pending.slice(idx + 1);
    idx = pending.indexOf('\n');
  }
  uiLineBuffers.set(key, pending);
  if (pending.length > 4096) {
    appendVisibleLog('[' + label + '] ' + pending + '\n');
    visibleLineNumber += 1;
    uiLineBuffers.set(key, '');
  }
  updateCounters();
}

copyLogEl.addEventListener('click', async () => {
  try {
    await navigator.clipboard.writeText(streamLogEl.textContent);
    log('copied visible log');
  } catch (error) {
    const selection = window.getSelection();
    const range = document.createRange();
    range.selectNodeContents(streamLogEl);
    selection.removeAllRanges();
    selection.addRange(range);
    log('clipboard write failed; selected visible log instead');
  }
});

function connect() {
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
  const ws = new WebSocket(proto + '//' + location.host + '/ws');

  ws.onopen = () => {
    statusEl.textContent = 'connected; waiting for goldweb commands';
    ws.send(JSON.stringify({ type: 'ready' }));
  };

  ws.onclose = () => {
    statusEl.textContent = 'disconnected; retrying';
    for (const worker of workers.values()) {
      worker.terminate();
    }
    workers.clear();
    setTimeout(connect, 1000);
  };

  ws.onerror = () => {
    statusEl.textContent = 'websocket error';
  };

  ws.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    if (msg.type === 'hello') {
      log(msg.message || 'hello');
      return;
    }
    if (msg.type === 'cancel') {
      const worker = workers.get(msg.id);
      if (worker) {
        worker.terminate();
        workers.delete(msg.id);
        log('cancelled ' + msg.id + ': ' + (msg.message || ''));
      }
      return;
    }
    if (msg.type === 'command' && msg.command === 'goivy_check') {
      runGoivyCheck(ws, msg);
      return;
    }
    log('unknown message: ' + event.data);
  };
}

function runGoivyCheck(ws, command) {
  jobIdEl.textContent = command.id;
  stdoutBytes = 0;
  stderrBytes = 0;
  xtraceCount = 0;
  visibleLineNumber = 0;
  uiLineBuffers.clear();
  streamLogEl.textContent = '';
  updateCounters();
  log('starting goivy_check job ' + command.id + ' filename=' + command.filename);
  const url = URL.createObjectURL(new Blob([workerSource], { type: 'text/javascript' }));
  const worker = new Worker(url);
  workers.set(command.id, worker);

  worker.onmessage = (event) => {
    const msg = Object.assign({ id: command.id }, event.data);
    ws.send(JSON.stringify(msg));
    if (msg.type === 'stream') {
      showStream(msg.fd, msg.data);
    }
    if (msg.type === 'error') {
      appendVisibleLog('[worker error] ' + (msg.error || 'unknown worker error') + '\n');
      visibleLineNumber += 1;
      updateCounters();
    }
    if (msg.type === 'done' || msg.type === 'error') {
      workers.delete(command.id);
      URL.revokeObjectURL(url);
      worker.terminate();
      if (msg.type === 'done') {
        log('finished ' + command.id + ' code=' + msg.code);
      } else {
        log('finished ' + command.id + ' type=error');
      }
    }
  };

  worker.onerror = (error) => {
    appendVisibleLog('[worker onerror] ' + (error.message || String(error)) + '\n');
    visibleLineNumber += 1;
    updateCounters();
    ws.send(JSON.stringify({
      type: 'error',
      id: command.id,
      error: error.message || String(error)
    }));
    workers.delete(command.id);
    URL.revokeObjectURL(url);
    worker.terminate();
  };

  worker.postMessage(command);
}

connect();
  </script>
</body>
</html>
`
