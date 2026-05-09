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
//   - goivy_check_prepare(specLen, metaLen uint32) int32
//   - goivy_check_write_spec(offset, packedBytes, n uint32) int32
//   - goivy_check_write_meta(offset, packedBytes, n uint32) int32
//   - goivy_check_run() int32
//
// It may also import goldweb.heap_profile(elapsedSeconds, ptr, len) or
// goldweb.xtrace_heap_profile(xtraceIndex, ptr, len) to ship runtime/pprof heap
// profiles back to this server during long conformance runs.
//
// meta is JSON: {"filename":"browser_input.ivy","params":{"isolate":"x"}}.
// The command writes its normal output to stdout/stderr; goldweb does not use a
// special trace callback.

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/base64"
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
	writeWait          = 10 * time.Minute
	pongWait           = 60 * time.Second
	pingPeriod         = 30 * time.Second
	readLimit          = 64 << 20
	sendQueueLen       = 8192
	spoolFlushBytes    = 256 << 10
	spoolFlushInterval = 5 * time.Second
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
	version    string
	httpServer *http.Server

	mu      sync.Mutex
	clients map[*wsClient]bool
	jobs    map[string]*job

	startupReq     *goivyCheckRequest
	startupStarted bool
}

type wsClient struct {
	app     *app
	conn    *websocket.Conn
	version string
	send    chan []byte
	done    chan struct{}
	closed  sync.Once
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
	ElapsedSec  int               `json:"elapsed_seconds,omitempty"`
	ChunkIndex  int               `json:"chunk_index,omitempty"`
	ChunkCount  int               `json:"chunk_count,omitempty"`
	ProfileSize int               `json:"profile_size,omitempty"`
	Code        int               `json:"code,omitempty"`
	Error       string            `json:"error,omitempty"`
	Status      string            `json:"status,omitempty"`
	Message     string            `json:"message,omitempty"`
	Version     string            `json:"version,omitempty"`
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

	pySpool        *lineSpool
	browserSpool   *lineSpool
	browserBuf     streamLineBuffer
	profiles       map[string]*profileAssembly
	compareDone    chan struct{}
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
	closePython  sync.Once
	stopOnce     sync.Once
}

type lineEvent struct {
	FD   int
	Line string
}

type streamLineBuffer struct {
	partial string
	fd      int
}

type profileAssembly struct {
	label       string
	profileSize int
	chunks      [][]byte
	received    int
}

type lineSpool struct {
	side      string
	path      string
	writeFile *os.File
	readFile  *os.File
	writer    *bufio.Writer
	reader    *bufio.Reader
	flushStop chan struct{}
	flushDone chan struct{}

	mu            sync.Mutex
	cond          *sync.Cond
	flushStopOnce sync.Once
	seq           int64
	unflushed     int
	done          bool
	closed        bool
	err           error
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
		version:    newID(),
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
	mux.Handle("/node_modules/@bjorn3/browser_wasi_shim/", noStore(http.StripPrefix(
		"/node_modules/@bjorn3/browser_wasi_shim/",
		http.FileServer(http.Dir(filepath.Join(a.webvueDir, "node_modules", "@bjorn3", "browser_wasi_shim"))),
	)))

	a.httpServer = &http.Server{
		Addr:              a.listen,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("goldweb listening on http://%s", a.listen)
	log.Printf("goldweb browser asset version %s", a.version)
	log.Printf("POST JSON to http://%s/goivy_check to start a browser-backed check", a.listen)
	log.Printf("serving Ivy include tree from %s as the same browser WASI path", a.includeDir)
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
	setNoStore(w)
	html := strings.ReplaceAll(indexHTML, "__GOLDWEB_ASSET_VERSION__", a.version)
	_, _ = io.WriteString(w, html)
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
		setNoStore(w)
		http.ServeFile(w, r, path)
	}
}

func noStore(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setNoStore(w)
		next.ServeHTTP(w, r)
	})
}

func setNoStore(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store, max-age=0, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
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
		app:     a,
		conn:    conn,
		version: r.URL.Query().Get("v"),
		send:    make(chan []byte, sendQueueLen),
		done:    make(chan struct{}),
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
	if c.version != a.version {
		msg := fmt.Sprintf("stale goldweb page version %q; server is %q; reload required", c.version, a.version)
		_ = c.sendJSON(wsEnvelope{Type: "hello", Status: "stale", Message: msg, Version: a.version})
		_ = c.sendJSON(wsEnvelope{Type: "reload", Status: "stale", Message: msg, Version: a.version})
		log.Printf("%s", msg)
		return
	}
	_ = c.sendJSON(wsEnvelope{Type: "hello", Status: "ready", Message: "goldweb connected", Version: a.version})
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
		if c.version == a.version {
			return c
		}
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
	case "heap_profile":
		j := a.lookupJob(msg.ID)
		if j == nil {
			log.Printf("heap profile for unknown job %q", msg.ID)
			return
		}
		if err := j.addHeapProfileChunk(msg); err != nil {
			log.Printf("heap profile for job %s: %v", msg.ID, err)
		}
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
	case "cancel":
		j := a.lookupJob(msg.ID)
		if j == nil {
			log.Printf("cancel for unknown job %q", msg.ID)
			return
		}
		reason := msg.Message
		if reason == "" {
			reason = "browser requested stop"
		}
		log.Printf("browser requested cancel for job %s: %s", msg.ID, reason)
		j.cancel(reason)
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
		http.Error(w, "no current-version browser websocket client connected; open / first", http.StatusServiceUnavailable)
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
		id:          newID(),
		filename:    req.Filename,
		sourcePath:  req.SourcePath,
		spec:        req.Spec,
		params:      cloneStringMap(req.Params),
		app:         a,
		client:      c,
		compareDone: make(chan struct{}),
		status:      "running",
	}
}

func (j *job) openXTraceLogs() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	browserPath := filepath.Join(cwd, "browser.xtrace.log")
	pyPath := filepath.Join(cwd, "py.xtrace.log")
	browserSpool, err := newLineSpool("browser", browserPath)
	if err != nil {
		return err
	}
	pySpool, err := newLineSpool("python", pyPath)
	if err != nil {
		_ = browserSpool.close()
		return err
	}

	j.browserSpool = browserSpool
	j.pySpool = pySpool
	j.browserLogPath = browserPath
	j.pyLogPath = pyPath
	log.Printf("writing browser XTRACE log to %s", browserPath)
	log.Printf("writing python XTRACE log to %s", pyPath)
	return nil
}

func (j *job) closeXTraceLogs() {
	if j.browserSpool != nil {
		if err := j.browserSpool.close(); err != nil {
			log.Printf("close %s: %v", j.browserLogPath, err)
		}
	}
	if j.pySpool != nil {
		if err := j.pySpool.close(); err != nil {
			log.Printf("close %s: %v", j.pyLogPath, err)
		}
	}
}

func (j *job) addHeapProfileChunk(msg wsEnvelope) error {
	if msg.ChunkCount <= 0 {
		return fmt.Errorf("invalid chunk_count=%d", msg.ChunkCount)
	}
	if msg.ChunkIndex < 0 || msg.ChunkIndex >= msg.ChunkCount {
		return fmt.Errorf("invalid chunk_index=%d chunk_count=%d", msg.ChunkIndex, msg.ChunkCount)
	}
	label := heapProfileLabel(msg)
	chunk, err := base64.StdEncoding.DecodeString(msg.Data)
	if err != nil {
		return fmt.Errorf("decode chunk %d/%d for %s: %w", msg.ChunkIndex+1, msg.ChunkCount, label, err)
	}
	var profile []byte
	j.mu.Lock()
	if j.profiles == nil {
		j.profiles = make(map[string]*profileAssembly)
	}
	assembly := j.profiles[label]
	if assembly == nil || len(assembly.chunks) != msg.ChunkCount || assembly.profileSize != msg.ProfileSize {
		assembly = &profileAssembly{
			label:       label,
			profileSize: msg.ProfileSize,
			chunks:      make([][]byte, msg.ChunkCount),
		}
		j.profiles[label] = assembly
	}
	if assembly.chunks[msg.ChunkIndex] == nil {
		assembly.received++
	}
	assembly.chunks[msg.ChunkIndex] = chunk
	if assembly.received == len(assembly.chunks) {
		total := 0
		for _, part := range assembly.chunks {
			total += len(part)
		}
		profile = make([]byte, 0, total)
		for _, part := range assembly.chunks {
			profile = append(profile, part...)
		}
		delete(j.profiles, label)
	}
	j.mu.Unlock()

	if profile == nil {
		return nil
	}
	if msg.ProfileSize >= 0 && len(profile) != msg.ProfileSize {
		return fmt.Errorf("assembled profile %s has %d bytes, expected %d", label, len(profile), msg.ProfileSize)
	}
	path, err := webMemprofPath(label)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, profile, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	log.Printf("wrote browser heap profile %s (%d bytes, %d chunks)", path, len(profile), msg.ChunkCount)
	return nil
}

func heapProfileLabel(msg wsEnvelope) string {
	if msg.XTraceIndex != nil {
		xtraceIndex := *msg.XTraceIndex
		if xtraceIndex < 0 {
			xtraceIndex = 0
		}
		return fmt.Sprintf("xtrace%d", xtraceIndex)
	}
	elapsedSec := msg.ElapsedSec
	if elapsedSec < 0 {
		elapsedSec = 0
	}
	return fmt.Sprintf("%d", elapsedSec)
}

func webMemprofPath(label string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	label = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			return r
		}
		return '_'
	}, label)
	if label == "" {
		label = "0"
	}
	return filepath.Join(cwd, "web.memprof."+label), nil
}

func newLineSpool(side, path string) (*lineSpool, error) {
	return newLineSpoolWithFlushInterval(side, path, spoolFlushInterval)
}

func newLineSpoolWithFlushInterval(side, path string, flushInterval time.Duration) (*lineSpool, error) {
	writeFile, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("create %s: %w", path, err)
	}
	readFile, err := os.Open(path)
	if err != nil {
		_ = writeFile.Close()
		return nil, fmt.Errorf("open %s for comparison: %w", path, err)
	}
	spool := &lineSpool{
		side:      side,
		path:      path,
		writeFile: writeFile,
		readFile:  readFile,
		writer:    bufio.NewWriterSize(writeFile, spoolFlushBytes),
		reader:    bufio.NewReaderSize(readFile, spoolFlushBytes),
		flushStop: make(chan struct{}),
		flushDone: make(chan struct{}),
	}
	spool.cond = sync.NewCond(&spool.mu)
	if flushInterval > 0 {
		go spool.flushPeriodically(flushInterval)
	} else {
		close(spool.flushDone)
	}
	return spool, nil
}

func (s *lineSpool) flushPeriodically(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	defer close(s.flushDone)

	for {
		select {
		case <-ticker.C:
			if err := s.flush(); err != nil {
				log.Printf("periodic flush %s: %v", s.path, err)
			}
		case <-s.flushStop:
			return
		}
	}
}

func (s *lineSpool) addLine(line string) error {
	if s == nil {
		return errors.New("nil line spool")
	}
	line = xtracer.NormalizeLine(line)
	if !strings.HasSuffix(line, "\n") {
		line += "\n"
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.done {
		return nil
	}
	if s.err != nil {
		return s.err
	}
	n, err := s.writer.WriteString(line)
	s.unflushed += n
	if err != nil {
		s.err = err
		s.seq++
		s.cond.Broadcast()
		return err
	}
	if s.unflushed >= spoolFlushBytes {
		return s.flushLocked()
	}
	return nil
}

func (s *lineSpool) markDone() {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.done {
		return
	}
	_ = s.flushLocked()
	s.done = true
	s.seq++
	s.cond.Broadcast()
}

func (s *lineSpool) nextLine() (string, bool, error) {
	if s == nil {
		return "", false, errors.New("nil line spool")
	}
	for {
		s.mu.Lock()
		seen := s.seq
		err := s.err
		done := s.done
		s.mu.Unlock()
		if err != nil {
			return "", false, err
		}

		line, readErr := s.reader.ReadString('\n')
		if readErr == nil {
			return line, true, nil
		}
		if !errors.Is(readErr, io.EOF) {
			return "", false, readErr
		}
		if line != "" {
			return line, true, nil
		}
		if done {
			return "", false, nil
		}

		s.mu.Lock()
		for !s.done && s.err == nil && s.seq == seen {
			s.cond.Wait()
		}
		s.mu.Unlock()
	}
}

func (s *lineSpool) close() error {
	if s == nil {
		return nil
	}
	s.stopPeriodicFlush()
	s.mu.Lock()
	if !s.done {
		_ = s.flushLocked()
		s.done = true
		s.seq++
		s.cond.Broadcast()
	}
	if s.closed {
		err := s.err
		s.mu.Unlock()
		return err
	}
	s.closed = true
	writeFile := s.writeFile
	readFile := s.readFile
	err := s.err
	s.mu.Unlock()

	if writeFile != nil {
		if closeErr := writeFile.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}
	if readFile != nil {
		if closeErr := readFile.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}
	return err
}

func (s *lineSpool) stopPeriodicFlush() {
	if s.flushStop == nil {
		return
	}
	s.flushStopOnce.Do(func() {
		close(s.flushStop)
		<-s.flushDone
	})
}

func (s *lineSpool) flush() error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.done {
		return s.err
	}
	return s.flushLocked()
}

func (s *lineSpool) flushLocked() error {
	if s.unflushed == 0 && s.err == nil {
		return nil
	}
	if err := s.writer.Flush(); err != nil {
		s.err = err
		s.seq++
		s.cond.Broadcast()
		return err
	}
	s.unflushed = 0
	s.seq++
	s.cond.Broadcast()
	return nil
}

func (j *job) runPythonIvyCheck() {
	defer j.setPythonDone()

	path := j.sourcePath
	if path == "" {
		tmp, err := os.MkdirTemp("", "goldweb-ivy-*")
		if err != nil {
			j.fail("create temp dir: " + err.Error())
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
		j.addPythonData(scanner.Text() + "\n")
	}
	if err := scanner.Err(); err != nil {
		j.setPythonErr("scan python output: " + err.Error())
	}
}

func (j *job) compareLoop() {
	defer j.closeXTraceLogs()
	defer close(j.compareDone)
	for {
		browser, browserOK := j.nextXTrace("browser", j.browserSpool)
		python, pythonOK := j.nextXTrace("python", j.pySpool)

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

func (j *job) nextXTrace(side string, spool *lineSpool) (string, bool) {
	if spool == nil {
		j.fail("missing " + side + " XTRACE spool")
		return "", false
	}
	for {
		line, ok, err := spool.nextLine()
		if err != nil {
			j.fail(fmt.Sprintf("read %s XTRACE spool: %v", side, err))
			return "", false
		}
		if !ok {
			return "", false
		}
		j.remember(side, line)
		if strings.HasPrefix(line, "XTRACE:") {
			return line, true
		}
	}
}

func (j *job) addBrowserData(fd int, data string) {
	lines := j.browserBuf.add(fd, data)
	for _, line := range lines {
		if err := j.browserSpool.addLine(line.Line); err != nil {
			j.fail(fmt.Sprintf("write %s: %v", j.browserLogPath, err))
			return
		}
	}
}

func (j *job) addPythonData(data string) {
	if err := j.pySpool.addLine(data); err != nil {
		j.fail(fmt.Sprintf("write %s: %v", j.pyLogPath, err))
	}
}

func (j *job) setBrowserDone(code int) {
	j.mu.Lock()
	j.browserCode = code
	j.mu.Unlock()
	lines := j.browserBuf.flush()
	for _, line := range lines {
		if err := j.browserSpool.addLine(line.Line); err != nil {
			j.fail(fmt.Sprintf("write %s: %v", j.browserLogPath, err))
			break
		}
	}
	j.closeBrowser.Do(func() {
		if j.browserSpool != nil {
			j.browserSpool.markDone()
		}
	})
}

func (j *job) setPythonDone() {
	j.closePython.Do(func() {
		if j.pySpool != nil {
			j.pySpool.markDone()
		}
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

func (j *job) cancel(reason string) {
	if reason == "" {
		reason = "cancelled"
	}
	j.mu.Lock()
	if j.status == "running" {
		j.status = "cancelled"
		j.message = reason
	}
	j.mu.Unlock()
	j.cancelPython()
	j.setBrowserDone(-1)
}

func (j *job) stopProducers(message string, xtraceIndex *int) {
	j.stopOnce.Do(func() {
		j.cancelPython()
		if j.client == nil {
			return
		}
		_ = j.client.sendJSON(wsEnvelope{
			Type:        "cancel",
			ID:          j.id,
			Message:     message,
			XTraceIndex: copyIntPtr(xtraceIndex),
		})
	})
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
	var shouldStop bool
	j.mu.Lock()
	if j.status == "running" {
		mismatchAt = j.xtraceCount
		j.mismatchAt = &mismatchAt
		j.status = "mismatch"
		j.message = fmt.Sprintf("%s at XTRACE index %d after %d matching XTRACE lines\nbrowser: %.500s\npython : %.500s", reason, mismatchAt, j.xtraceCount, strings.TrimSpace(browser), strings.TrimSpace(python))
		message = j.message
		shouldStop = true
	} else {
		message = j.message
		if j.mismatchAt != nil {
			mismatchAt = *j.mismatchAt
		} else {
			mismatchAt = j.xtraceCount
		}
	}
	j.mu.Unlock()
	if shouldStop {
		j.stopProducers(message, &mismatchAt)
	}
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
    .display-control {
      align-items: center;
      color: #52606d;
      display: inline-flex;
      font-size: 0.8rem;
      gap: 0.35rem;
    }
    #display-skip { width: 10rem; }
    #display-skip-value {
      color: #102a43;
      display: inline-block;
      font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, "Liberation Mono", monospace;
      font-variant-numeric: tabular-nums;
      min-width: 4ch;
      text-align: right;
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
    button:disabled {
      cursor: not-allowed;
      opacity: 0.55;
    }
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
        <button id="stop-job" type="button" disabled>Stop job</button>
        <label class="display-control" title="Show one output line, then skip this many output lines">
          <span>skip</span>
          <input id="display-skip" type="range" min="0" max="1000" step="1" value="0">
          <output id="display-skip-value" for="display-skip">0</output>
        </label>
        <span>rolling local window; full stream goes over the websocket</span>
      </div>
    </div>
    <pre id="stream-log" tabindex="0"></pre>
    <div id="line-number">last line number : 0</div>
  </div>

  <script type="text/plain" id="worker-source">
const textEncoder = new TextEncoder();
const goldwebAssetVersion = '__GOLDWEB_ASSET_VERSION__';

function assetURL(assetBaseURL, path, version) {
  return assetBaseURL + path + '?v=' + encodeURIComponent(version || goldwebAssetVersion);
}

function requireExport(exports, name) {
  const fn = exports[name];
  if (typeof fn !== 'function') {
    throw new Error('Go Ivy wasm is missing export ' + name);
  }
  return fn;
}

function writeBytesToGo(exports, exportName, bytes) {
  const write = requireExport(exports, exportName);
  let offset = 0;
  while (offset < bytes.length) {
    const n = Math.min(4, bytes.length - offset);
    let word = 0;
    for (let i = 0; i < n; i += 1) {
      word |= bytes[offset + i] << (8 * i);
    }
    const code = write(offset >>> 0, word >>> 0, n >>> 0) | 0;
    if (code !== 0) {
      throw new Error(exportName + ' rejected bytes at offset ' + offset + ' with code ' + code);
    }
    offset += n;
  }
}

function bytesToBase64(bytes) {
  let binary = '';
  const blockSize = 0x8000;
  for (let offset = 0; offset < bytes.length; offset += blockSize) {
    binary += String.fromCharCode.apply(null, bytes.subarray(offset, offset + blockSize));
  }
  return btoa(binary);
}

function postHeapProfile(memory, elapsedSeconds, ptr, len, xtraceIndex) {
  if (!memory) {
    throw new Error('goldweb.heap_profile called before Go wasm memory was available');
  }
  const n = len >>> 0;
  const offset = ptr >>> 0;
  const copy = new Uint8Array(n);
  if (n > 0) {
    copy.set(new Uint8Array(memory.buffer, offset, n));
  }
  memory = null; // don't let it linger on our stack.

  const chunkSize = 768 * 1024;
  const chunkCount = Math.max(1, Math.ceil(copy.length / chunkSize));
  for (let i = 0; i < chunkCount; i += 1) {
    const start = i * chunkSize;
    const end = Math.min(start + chunkSize, copy.length);
    const message = {
      type: 'heap_profile',
      elapsed_seconds: elapsedSeconds >>> 0,
      chunk_index: i,
      chunk_count: chunkCount,
      profile_size: copy.length,
      data: bytesToBase64(copy.subarray(start, end))
    };
    if (Number.isInteger(xtraceIndex)) {
      message.xtrace_index = xtraceIndex;
    }
    self.postMessage(message);
  }
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

async function loadIncludeTree(wasiShim, assetBaseURL, version) {
  const response = await fetch(assetURL(assetBaseURL, '/ivy-include-tree.json', version), { cache: 'no-store' });
  if (!response.ok) {
    throw new Error('could not fetch /ivy-include-tree.json: HTTP ' + response.status);
  }
  const tree = await response.json();
  return {
    root: String(tree.root || 'include'),
    directory: includeDirectoryFromTree(wasiShim, tree)
  };
}

self.onmessage = async (event) => {
  const command = event.data;
  let z3;
  try {
    const assetBaseURL = self.location.origin;
    const commandAssetVersion = goldwebAssetVersion + '-' + String(command.id || Date.now());
    const z3Imports = await import(assetURL(assetBaseURL, '/src/workers/smtZ3Imports.js', commandAssetVersion));
    const wasiShim = await import(assetURL(assetBaseURL, '/node_modules/@bjorn3/browser_wasi_shim/dist/index.js', commandAssetVersion));
    importScripts(assetURL(assetBaseURL, '/z3-471-api.js', commandAssetVersion));
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
          return assetURL(assetBaseURL, '/z3-471-api.wasm', commandAssetVersion);
        }
        return assetURL(assetBaseURL, '/' + file, commandAssetVersion);
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
    const includeTree = await loadIncludeTree(wasiShim, assetBaseURL, commandAssetVersion);
    const wasi = new wasiShim.WASI(
      ['goivy_check_wasip1'],
      [
        'GOIVY_INCLUDE=' + includeTree.root,
        'GOIVY_WASM_HEAPPROFILE_INTERVAL=10',
      ],
      [
        new wasiShim.OpenFile(new wasiShim.File(new Uint8Array())),
        new wasiShim.ConsoleStdout(emitStdout),
        new wasiShim.ConsoleStdout(emitStderr),
        new wasiShim.PreopenDirectory(includeTree.root, includeTree.directory.contents),
      ],
    );

    let wasmMemory;
    const imports = {
      goldweb: {
        heap_profile(elapsedSeconds, ptr, len) {
          postHeapProfile(wasmMemory, elapsedSeconds, ptr, len);
        },
        xtrace_heap_profile(xtraceIndex, ptr, len) {
          postHeapProfile(wasmMemory, 0, ptr, len, xtraceIndex);
        },
      },
      smt_z3: z3Imports.createSmtZ3Imports({ z3, getGoMemory: () => wasmMemory }),
      wasi_snapshot_preview1: wasi.wasiImport,
    };

    const response = await fetch(assetURL(assetBaseURL, '/goivy-check.wasm', commandAssetVersion), { cache: 'no-store' });
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

    const prepare = requireExport(instance.exports, 'goivy_check_prepare');
    const run = requireExport(instance.exports, 'goivy_check_run');
    const specBytes = textEncoder.encode(command.spec || '');
    const metaBytes = textEncoder.encode(JSON.stringify({
      filename: command.filename || 'browser_input.ivy',
      params: command.params || {}
    }));
    const prepareCode = prepare(specBytes.length >>> 0, metaBytes.length >>> 0) | 0;
    if (prepareCode !== 0) {
      throw new Error('goivy_check_prepare failed with code ' + prepareCode);
    }
    writeBytesToGo(instance.exports, 'goivy_check_write_spec', specBytes);
    writeBytesToGo(instance.exports, 'goivy_check_write_meta', metaBytes);
    const code = run() | 0;
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
const pageVersion = '__GOLDWEB_ASSET_VERSION__';
const statusEl = document.getElementById('status');
const jobIdEl = document.getElementById('job-id');
const stdoutBytesEl = document.getElementById('stdout-bytes');
const stderrBytesEl = document.getElementById('stderr-bytes');
const xtraceCountEl = document.getElementById('xtrace-count');
const streamLogEl = document.getElementById('stream-log');
const copyLogEl = document.getElementById('copy-log');
const stopJobEl = document.getElementById('stop-job');
const displaySkipEl = document.getElementById('display-skip');
const displaySkipValueEl = document.getElementById('display-skip-value');
const lineNumberEl = document.getElementById('line-number');
const workerSource = document.getElementById('worker-source').textContent;
const workers = new Map();
const uiLineBuffers = new Map();
const uiTextEncoder = new TextEncoder();
const preserveFirstStreamLines = 30;
const maxRollingVisibleLogLines = 8000;
const lineNumberWidth = 10;
const unnumberedPrefix = ' '.repeat(lineNumberWidth);
const preservedVisibleLogLines = [];
const visibleLogLines = [];
const pendingVisibleLogLines = [];
let stdoutBytes = 0;
let stderrBytes = 0;
let xtraceCount = 0;
let visibleLineNumber = 0;
let activeJobID = '';
let activeWS = null;
let displaySkip = 0;
let displaySkipCountdown = 0;
let visibleLogRenderScheduled = false;
let visibleLogDirty = false;

function log(message) {
  appendVisibleLog('[goldweb] ' + message + '\n', true);
}

function formatVisibleLogLine(text, lineNumber) {
  const prefix = Number.isInteger(lineNumber)
    ? String(lineNumber).padStart(lineNumberWidth, ' ')
    : unnumberedPrefix;
  return prefix + ' ' + text;
}

function splitVisibleLogText(text) {
  return String(text).match(/[^\n]*\n|[^\n]+/g) || [];
}

function shouldDisplayOutputLine() {
  if (displaySkip <= 0) {
    return true;
  }
  if (displaySkipCountdown <= 0) {
    displaySkipCountdown = displaySkip;
    return true;
  }
  displaySkipCountdown -= 1;
  return false;
}

function appendVisibleLog(text, force, lineNumber) {
  const pieces = splitVisibleLogText(text);
  const preserve = Number.isInteger(lineNumber) && lineNumber < preserveFirstStreamLines;
  if (!force && !preserve && !shouldDisplayOutputLine()) {
    return;
  }
  for (const piece of pieces) {
    const formatted = formatVisibleLogLine(piece, lineNumber);
    if (preserve) {
      preservedVisibleLogLines.push(formatted);
    } else {
      pendingVisibleLogLines.push(formatted);
    }
  }
  visibleLogDirty = true;
  scheduleVisibleLogRender();
}

function scheduleVisibleLogRender() {
  if (visibleLogRenderScheduled) {
    return;
  }
  visibleLogRenderScheduled = true;
  requestAnimationFrame(() => {
    visibleLogRenderScheduled = false;
    renderVisibleLog(false);
  });
}

function renderVisibleLog(forceTail) {
  if (!pendingVisibleLogLines.length && !visibleLogDirty) {
    return;
  }
  const distanceFromBottom = streamLogEl.scrollHeight - streamLogEl.scrollTop - streamLogEl.clientHeight;
  const shouldTail = forceTail || distanceFromBottom < 8;
  for (const line of pendingVisibleLogLines) {
    visibleLogLines.push(line);
  }
  pendingVisibleLogLines.length = 0;
  if (visibleLogLines.length > maxRollingVisibleLogLines) {
    visibleLogLines.splice(0, visibleLogLines.length - maxRollingVisibleLogLines);
  }
  streamLogEl.textContent = preservedVisibleLogLines.concat(visibleLogLines).join('');
  if (shouldTail) {
    streamLogEl.scrollTop = streamLogEl.scrollHeight;
  }
  visibleLogDirty = false;
}

function flushVisibleLog(forceTail) {
  renderVisibleLog(forceTail);
}

function updateCounters() {
  stdoutBytesEl.textContent = String(stdoutBytes);
  stderrBytesEl.textContent = String(stderrBytes);
  xtraceCountEl.textContent = String(xtraceCount);
  lineNumberEl.textContent = 'last line number: ' + (visibleLineNumber > 0 ? visibleLineNumber - 1 : -1);
}

function setActiveJob(id, ws) {
  activeJobID = id || '';
  activeWS = activeJobID ? ws : null;
  stopJobEl.disabled = activeJobID === '';
}

function clearActiveJob(id) {
  if (!id || activeJobID === id) {
    setActiveJob('', null);
  }
}

function terminateJob(id) {
  const running = workers.get(id);
  if (!running) {
    clearActiveJob(id);
    return false;
  }
  running.worker.terminate();
  URL.revokeObjectURL(running.url);
  workers.delete(id);
  clearActiveJob(id);
  return true;
}

function terminateAllJobs() {
  for (const id of Array.from(workers.keys())) {
    terminateJob(id);
  }
}

function scrollToStartOfLog() {
  flushVisibleLog(false);
  streamLogEl.scrollTop = 0;
}

function scrollToEndOfLog() {
  flushVisibleLog(true);
  streamLogEl.scrollTop = streamLogEl.scrollHeight;
}

function checkKey(event) {
  if (!event.shiftKey) {
    return;
  }
  switch (event.key) {
    case 'ArrowUp':
      event.preventDefault();
      scrollToStartOfLog();
      break;
    case 'ArrowDown':
      event.preventDefault();
      scrollToEndOfLog();
      break;
  }
}

function reloadForServerVersion(version) {
  terminateAllJobs();
  const nextVersion = encodeURIComponent(version || String(Date.now()));
  location.replace('/?v=' + nextVersion);
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
    appendVisibleLog('[' + label + '] ' + line, false, visibleLineNumber);
    visibleLineNumber += 1;
    pending = pending.slice(idx + 1);
    idx = pending.indexOf('\n');
  }
  uiLineBuffers.set(key, pending);
  if (pending.length > 4096) {
    appendVisibleLog('[' + label + '] ' + pending + '\n', false, visibleLineNumber);
    visibleLineNumber += 1;
    uiLineBuffers.set(key, '');
  }
  updateCounters();
}

copyLogEl.addEventListener('click', async () => {
  flushVisibleLog(false);
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

displaySkipEl.addEventListener('input', () => {
  displaySkip = Number(displaySkipEl.value) || 0;
  displaySkipCountdown = 0;
  displaySkipValueEl.textContent = String(displaySkip);
});

stopJobEl.addEventListener('click', () => {
  const id = activeJobID;
  const ws = activeWS;
  if (!id) {
    return;
  }
  const stopped = terminateJob(id);
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify({
      type: 'cancel',
      id,
      message: 'browser stop button'
    }));
  }
  log((stopped ? 'stop requested for ' : 'stop requested for already-finished ') + id);
});

document.addEventListener('keydown', checkKey);

function connect() {
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
  const ws = new WebSocket(proto + '//' + location.host + '/ws?v=' + encodeURIComponent(pageVersion));

  ws.onopen = () => {
    statusEl.textContent = 'connected; waiting for goldweb commands';
    ws.send(JSON.stringify({ type: 'ready' }));
  };

  ws.onclose = () => {
    statusEl.textContent = 'disconnected; retrying';
    terminateAllJobs();
    setTimeout(connect, 1000);
  };

  ws.onerror = () => {
    statusEl.textContent = 'websocket error';
  };

  ws.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    if (msg.type === 'hello') {
      if (msg.version && msg.version !== pageVersion) {
        log(msg.message || 'server has a newer goldweb page; reloading');
        reloadForServerVersion(msg.version);
        return;
      }
      log(msg.message || 'hello');
      return;
    }
    if (msg.type === 'reload') {
      log(msg.message || 'server requested reload');
      reloadForServerVersion(msg.version);
      return;
    }
    if (msg.type === 'cancel') {
      const stopped = terminateJob(msg.id);
      if (stopped) {
        const at = Number.isInteger(msg.xtrace_index) ? ' at XTRACE index ' + msg.xtrace_index : '';
        log('cancelled ' + msg.id + at + ': ' + (msg.message || ''));
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
  preservedVisibleLogLines.length = 0;
  visibleLogLines.length = 0;
  pendingVisibleLogLines.length = 0;
  visibleLogDirty = false;
  displaySkipCountdown = 0;
  streamLogEl.textContent = '';
  updateCounters();
  log('starting goivy_check job ' + command.id + ' filename=' + command.filename);
  const url = URL.createObjectURL(new Blob([workerSource], { type: 'text/javascript' }));
  const worker = new Worker(url);
  workers.set(command.id, { worker, url });
  setActiveJob(command.id, ws);

  worker.onmessage = (event) => {
    const msg = Object.assign({ id: command.id }, event.data);
    ws.send(JSON.stringify(msg));
    if (msg.type === 'stream') {
      showStream(msg.fd, msg.data);
    }
    if (msg.type === 'heap_profile' && msg.chunk_index === msg.chunk_count - 1) {
      const where = Number.isInteger(msg.xtrace_index)
        ? 'xtrace=' + msg.xtrace_index
        : 't=' + msg.elapsed_seconds + 's';
      log('sent heap profile ' + where + ' bytes=' + msg.profile_size + ' chunks=' + msg.chunk_count);
    }
    if (msg.type === 'error') {
      appendVisibleLog('[worker error] ' + (msg.error || 'unknown worker error') + '\n', true);
      updateCounters();
    }
    if (msg.type === 'done' || msg.type === 'error') {
      terminateJob(command.id);
      if (msg.type === 'done') {
        log('finished ' + command.id + ' code=' + msg.code);
      } else {
        log('finished ' + command.id + ' type=error');
      }
    }
  };

  worker.onerror = (error) => {
    appendVisibleLog('[worker onerror] ' + (error.message || String(error)) + '\n', true);
    updateCounters();
    ws.send(JSON.stringify({
      type: 'error',
      id: command.id,
      error: error.message || String(error)
    }));
    terminateJob(command.id);
  };

  worker.postMessage(command);
}

connect();
  </script>
</body>
</html>
`
