package ivy2cpp

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type vsInfo struct {
	InstallDir   string
	ToolsVersion string
	BinDir       string
	IncludeDirs  []string
	LibDirs      []string
	Env          []string
}

var findVSFunc = defaultFindVS

func findVS() (vsInfo, error) {
	return findVSFunc()
}

func parseVSWhereInstallationPath(data []byte) (string, error) {
	text := strings.TrimSpace(string(data))
	if text == "" {
		return "", errors.New("vswhere returned no installationPath")
	}
	var rawString string
	if err := json.Unmarshal([]byte(text), &rawString); err == nil {
		if path := strings.TrimSpace(rawString); path != "" {
			return path, nil
		}
	}
	var rawObject map[string]any
	if err := json.Unmarshal([]byte(text), &rawObject); err == nil {
		if path, ok := jsonStringField(rawObject, "installationPath"); ok {
			return path, nil
		}
	}
	var rawArray []map[string]any
	if err := json.Unmarshal([]byte(text), &rawArray); err == nil {
		for _, item := range rawArray {
			if path, ok := jsonStringField(item, "installationPath"); ok {
				return path, nil
			}
		}
	}
	lines := strings.Split(text, "\n")
	if len(lines) == 1 && looksLikeVSInstallPath(lines[0]) {
		return strings.TrimSpace(lines[0]), nil
	}
	return "", fmt.Errorf("vswhere output did not contain installationPath: %s", text)
}

func jsonStringField(item map[string]any, name string) (string, bool) {
	if item == nil {
		return "", false
	}
	value, ok := item[name]
	if !ok {
		return "", false
	}
	text, ok := value.(string)
	if !ok {
		return "", false
	}
	text = strings.TrimSpace(text)
	return text, text != ""
}

func looksLikeVSInstallPath(path string) bool {
	path = strings.TrimSpace(path)
	return strings.Contains(strings.ToLower(path), "visual studio") || strings.Contains(path, `\VC\`) || strings.Contains(path, `/VC/`)
}

func vsInfoFromInstall(install string) (vsInfo, error) {
	install = filepath.Clean(strings.TrimSpace(install))
	if install == "." || install == "" {
		return vsInfo{}, errors.New("empty Visual Studio installation path")
	}
	toolsRoot := filepath.Join(install, "VC", "Tools", "MSVC")
	entries, err := os.ReadDir(toHostPath(toolsRoot))
	if err != nil {
		return vsInfo{}, fmt.Errorf("Visual Studio MSVC tools not found under %s: %w", toolsRoot, err)
	}
	var versions []string
	for _, entry := range entries {
		if entry.IsDir() {
			versions = append(versions, entry.Name())
		}
	}
	if len(versions) == 0 {
		return vsInfo{}, fmt.Errorf("Visual Studio MSVC tools not found under %s", toolsRoot)
	}
	sort.Slice(versions, func(i, j int) bool {
		return compareDottedVersion(versions[i], versions[j]) < 0
	})
	version := versions[len(versions)-1]
	toolRoot := filepath.Join(toolsRoot, version)
	info := vsInfo{
		InstallDir:   install,
		ToolsVersion: version,
		BinDir:       filepath.Join(toolRoot, "bin", "Hostx64", "x64"),
		IncludeDirs:  []string{filepath.Join(toolRoot, "include")},
		LibDirs:      []string{filepath.Join(toolRoot, "lib", "x64")},
	}
	return info, nil
}

func compareDottedVersion(a, b string) int {
	as := strings.Split(a, ".")
	bs := strings.Split(b, ".")
	n := len(as)
	if len(bs) > n {
		n = len(bs)
	}
	for i := 0; i < n; i++ {
		av := 0
		bv := 0
		if i < len(as) {
			fmt.Sscanf(as[i], "%d", &av)
		}
		if i < len(bs) {
			fmt.Sscanf(bs[i], "%d", &bv)
		}
		if av < bv {
			return -1
		}
		if av > bv {
			return 1
		}
	}
	return strings.Compare(a, b)
}

func msvcToolchainEnv(vs vsInfo) []string {
	env := append([]string(nil), os.Environ()...)
	if len(vs.Env) != 0 {
		env = mergeEnv(env, vs.Env)
	}
	if strings.TrimSpace(vs.BinDir) != "" {
		env = prependEnvPath(env, "PATH", vs.BinDir)
	}
	if len(vs.IncludeDirs) != 0 {
		env = prependEnvPath(env, "INCLUDE", strings.Join(vs.IncludeDirs, ";"))
	}
	if len(vs.LibDirs) != 0 {
		env = prependEnvPath(env, "LIB", strings.Join(vs.LibDirs, ";"))
	}
	return env
}

func mergeEnv(base, overrides []string) []string {
	out := append([]string(nil), base...)
	index := envIndex(out)
	for _, item := range overrides {
		name, _, ok := strings.Cut(item, "=")
		if !ok || name == "" {
			continue
		}
		key := strings.ToUpper(name)
		if pos, ok := index[key]; ok {
			out[pos] = item
		} else {
			index[key] = len(out)
			out = append(out, item)
		}
	}
	return out
}

func prependEnvPath(env []string, name, prefix string) []string {
	prefix = strings.Trim(prefix, ";")
	if prefix == "" {
		return env
	}
	out := append([]string(nil), env...)
	index := envIndex(out)
	key := strings.ToUpper(name)
	if pos, ok := index[key]; ok {
		_, old, _ := strings.Cut(out[pos], "=")
		if old == "" {
			out[pos] = name + "=" + prefix
		} else {
			out[pos] = name + "=" + prefix + ";" + old
		}
		return out
	}
	return append(out, name+"="+prefix)
}

func envIndex(env []string) map[string]int {
	out := map[string]int{}
	for i, item := range env {
		name, _, ok := strings.Cut(item, "=")
		if !ok || name == "" {
			continue
		}
		out[strings.ToUpper(name)] = i
	}
	return out
}

func envValue(env []string, name string) string {
	key := strings.ToUpper(name)
	for _, item := range env {
		k, v, ok := strings.Cut(item, "=")
		if ok && strings.ToUpper(k) == key {
			return v
		}
	}
	return ""
}

func toHostPath(path string) string {
	if filepath.Separator == '\\' {
		return path
	}
	return strings.ReplaceAll(path, `\`, string(filepath.Separator))
}
