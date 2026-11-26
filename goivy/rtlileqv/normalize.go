package rtlileqv

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type designConfig struct {
	section      string
	file, top    string
	clock, reset string
}

var moduleName = regexp.MustCompile(`^[A-Za-z_$][A-Za-z0-9_.$:\-]*$`)

// Check normalizes the two designs in an .ileq file with Yosys and compares
// their output and next-state logic with the native Ivy solver.
func Check(ctx context.Context, configPath string) (Result, error) {
	configs, err := readConfig(configPath)
	if err != nil {
		return Result{}, err
	}
	executable := os.Getenv("IVY_YOSYS")
	if executable == "" {
		executable = "yosys"
	}
	yosys, err := exec.LookPath(executable)
	if err != nil {
		return Result{}, fmt.Errorf("locate Yosys (%s): %w", executable, err)
	}
	yosys, err = filepath.Abs(yosys)
	if err != nil {
		return Result{}, err
	}
	work, err := os.MkdirTemp("", "ivy-rtlil-eqv-")
	if err != nil {
		return Result{}, err
	}
	if os.Getenv("IVY_RTLIL_EQV_KEEP") != "" {
		defer fmt.Fprintf(os.Stderr, "[rtlil_eqv] work dir kept: %s\n", work)
	} else {
		defer os.RemoveAll(work)
	}

	var circuits [2]*aiger
	for i, cfg := range configs {
		circuits[i], err = normalize(ctx, yosys, work, cfg)
		if err != nil {
			return Result{}, err
		}
	}
	return compare(circuits[0], circuits[1], configs[0], configs[1])
}

func readConfig(path string) ([2]designConfig, error) {
	var result [2]designConfig
	file, err := os.Open(path)
	if err != nil {
		return result, err
	}
	defer file.Close()
	values := map[string]map[string]string{}
	var section string
	scan := bufio.NewScanner(file)
	for line := 1; scan.Scan(); line++ {
		s := strings.TrimSpace(scan.Text())
		if s == "" || strings.HasPrefix(s, "#") || strings.HasPrefix(s, ";") {
			continue
		}
		if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
			section = strings.TrimSpace(s[1 : len(s)-1])
			if section == "" || values[section] != nil {
				return result, fmt.Errorf("%s:%d: empty or repeated section", path, line)
			}
			values[section] = make(map[string]string)
			continue
		}
		sep := strings.IndexAny(s, "=:")
		if section == "" || sep < 1 {
			return result, fmt.Errorf("%s:%d: expected a section or setting", path, line)
		}
		key, value := strings.TrimSpace(s[:sep]), strings.TrimSpace(s[sep+1:])
		if _, exists := values[section][key]; exists {
			return result, fmt.Errorf("%s:%d: repeated setting %q", path, line, key)
		}
		values[section][key] = value
	}
	if err := scan.Err(); err != nil {
		return result, err
	}
	for i, name := range []string{"gold", "gate"} {
		settings := values[name]
		if settings == nil {
			return result, fmt.Errorf("configuration is missing section [%s]", name)
		}
		for _, key := range []string{"file", "top", "clk", "rst"} {
			if settings[key] == "" {
				return result, fmt.Errorf("section [%s] is missing setting %q", name, key)
			}
		}
		cfg := designConfig{section: name, file: settings["file"], top: settings["top"], clock: settings["clk"], reset: settings["rst"]}
		cfg.top = strings.TrimPrefix(cfg.top, `\`)
		if !moduleName.MatchString(cfg.top) {
			return result, fmt.Errorf("[%s] invalid top module name %q", name, cfg.top)
		}
		if cfg.clock == cfg.reset {
			return result, fmt.Errorf("[%s] clock and reset must be distinct", name)
		}
		if !filepath.IsAbs(cfg.file) {
			cfg.file = filepath.Join(filepath.Dir(path), cfg.file)
		}
		result[i] = cfg
	}
	return result, nil
}

const yosysScript = `read_rtlil design.il
hierarchy -check -top %s
proc
flatten
memory_map
async2sync
dfflegalize -cell $_DFF_P_ x
simplemap
setundef -zero
aigmap
insbuf o:*
techmap -map output_buffers.v
write_aiger -ascii -map design.map design.aag
`

// Older Yosys AIGER writers cannot map an output directly aliased to an
// input. Give such outputs a distinct driver without changing their value.
const outputBuffers = `module \$_BUF_ (input A, output Y);
  \$_AND_ _TECHMAP_REPLACE_ (.A(A), .B(A), .Y(Y));
endmodule
`

func normalize(ctx context.Context, yosys, work string, cfg designConfig) (*aiger, error) {
	dir := filepath.Join(work, cfg.section)
	if err := os.Mkdir(dir, 0700); err != nil {
		return nil, err
	}
	if err := copyDesign(cfg.file, filepath.Join(dir, "design.il")); err != nil {
		return nil, fmt.Errorf("[%s] read RTLIL: %w", cfg.section, err)
	}
	if err := os.WriteFile(filepath.Join(dir, "normalize.ys"), []byte(fmt.Sprintf(yosysScript, cfg.top)), 0600); err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(dir, "output_buffers.v"), []byte(outputBuffers), 0600); err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(ctx, yosys, "-q", "-s", "normalize.ys")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("Yosys failed on [%s] %q: %w\n%s", cfg.section, cfg.file, err, out)
	}
	graph, err := os.Open(filepath.Join(dir, "design.aag"))
	if err != nil {
		return nil, err
	}
	defer graph.Close()
	names, err := os.Open(filepath.Join(dir, "design.map"))
	if err != nil {
		return nil, err
	}
	defer names.Close()
	result, err := readAiger(graph, names)
	if err != nil {
		return nil, fmt.Errorf("[%s] parse Yosys output: %w", cfg.section, err)
	}
	return result, nil
}

func copyDesign(input, output string) error {
	source, err := os.Open(input)
	if err != nil {
		return err
	}
	defer source.Close()
	dest, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(dest, source)
	closeErr := dest.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}
