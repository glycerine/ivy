package abcmain

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// LibraryLoader replaces C dlsym/void* casts with typed Go extension loading.
// The default loader resolves registered Go libraries; platform-specific loaders
// may be supplied by the executable without adding C dependencies to the frame.
type Library struct {
	Name      string
	Init, End func(*Frame)
	Symbols   map[string]any
}
type LibraryLoader interface {
	Load(string) (*Library, error)
}

var goLibraries struct {
	sync.RWMutex
	values map[string]*Library
}

func RegisterLibrary(library *Library) error {
	if library == nil || library.Name == "" {
		return errors.New("abcmain: library requires a name")
	}
	goLibraries.Lock()
	defer goLibraries.Unlock()
	if goLibraries.values == nil {
		goLibraries.values = map[string]*Library{}
	}
	if _, exists := goLibraries.values[library.Name]; exists {
		return fmt.Errorf("abcmain: library %s already registered", library.Name)
	}
	goLibraries.values[library.Name] = library
	return nil
}
func registeredLibrary(name string) (*Library, error) {
	goLibraries.RLock()
	defer goLibraries.RUnlock()
	if library := goLibraries.values[name]; library != nil {
		return library, nil
	}
	if library := goLibraries.values[filepath.Base(name)]; library != nil {
		return library, nil
	}
	return nil, fmt.Errorf("abcmain: no Go loader registered for extension %s", name)
}
func (f *Frame) LoadLibraries(names []string) error {
	var loaded []*Library
	for _, name := range names {
		var library *Library
		var err error
		if f.LibraryLoader != nil {
			library, err = f.LibraryLoader.Load(name)
		} else {
			library, err = registeredLibrary(name)
		}
		if err != nil {
			return err
		}
		if library == nil {
			return fmt.Errorf("abcmain: loader returned nil for %s", name)
		}
		loaded = append(loaded, library)
	}
	f.mu.Lock()
	f.libraries = append(f.libraries, loaded...)
	f.mu.Unlock()
	for _, library := range loaded {
		if library.Init != nil {
			library.Init(f)
		}
	}
	return nil
}
func (f *Frame) UnloadLibraries() {
	f.mu.Lock()
	libraries := f.libraries
	f.libraries = nil
	f.mu.Unlock()
	for i := len(libraries) - 1; i >= 0; i-- {
		if libraries[i].End != nil {
			libraries[i].End(f)
		}
	}
}
func libraryPaths() []string {
	var paths []string
	directories := filepath.SplitList(os.Getenv("ABC_LIB_PATH"))
	if len(directories) == 0 {
		directories = []string{"."}
	}
	for _, directory := range directories {
		entries, err := os.ReadDir(directory)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasPrefix(entry.Name(), "libabc_") && strings.HasSuffix(entry.Name(), ".so") {
				paths = append(paths, filepath.Join(directory, entry.Name()))
			}
		}
	}
	sort.Strings(paths)
	return paths
}
func Libs_Init(f *Frame) error {
	names := libraryPaths()
	if configured, exists := f.LookupFlag("libs"); exists {
		names = append(names, strings.Fields(configured)...)
	}
	return f.LoadLibraries(names)
}
func Libs_End(f *Frame) { f.UnloadLibraries() }
func open_libs() error  { return Libs_Init(Abc_FrameGetGlobalFrame()) }
func close_libs()       { Libs_End(Abc_FrameGetGlobalFrame()) }
func get_fnct_ptr(index int32, name string) any {
	f := Abc_FrameGetGlobalFrame()
	f.mu.RLock()
	defer f.mu.RUnlock()
	if index < 0 || int(index) >= len(f.libraries) {
		return nil
	}
	return f.libraries[index].Symbols[name]
}
func call_inits(f *Frame) {
	f.mu.RLock()
	libraries := append([]*Library(nil), f.libraries...)
	f.mu.RUnlock()
	for _, library := range libraries {
		if library.Init != nil {
			library.Init(f)
		}
	}
}
func call_ends(f *Frame) {
	f.mu.RLock()
	libraries := append([]*Library(nil), f.libraries...)
	f.mu.RUnlock()
	for i := len(libraries) - 1; i >= 0; i-- {
		if libraries[i].End != nil {
			libraries[i].End(f)
		}
	}
}
