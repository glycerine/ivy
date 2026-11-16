package gia_test

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	if_ "github.com/glycerine/abcgo/map/if"
	"github.com/glycerine/abcgo/opt/dau"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

var mappingDSDReviewMemoryLine = regexp.MustCompile(`(?m)^(Memory used for [^=]+=) +[0-9.]+ MB\.$`)

func mappingDSDReviewReportNormal(text string) string {
	return mappingDSDReviewMemoryLine.ReplaceAllString(text, "${1} <storage> MB.")
}
func mappingDSDReviewCapture(t *testing.T, run func()) string {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "report-stdout")
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = file
	defer func() { os.Stdout = original; file.Close() }()
	run()
	output, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(output)
}
func mappingDSDReviewReportEqual(t *testing.T, kind, got, want string) {
	t.Helper()
	a, b := strings.Split(mappingDSDReviewReportNormal(got), "\n"), strings.Split(mappingDSDReviewReportNormal(want), "\n")
	for i := 0; i < max(len(a), len(b)); i++ {
		if i >= len(a) || i >= len(b) {
			t.Fatalf("%s report length Go%d C%d", kind, len(a), len(b))
		}
		if a[i] != b[i] {
			t.Fatalf("%s report line%d Go%q C%q", kind, i+1, a[i], b[i])
		}
	}
}
func TestMappingDSDNativeReportOptionsAndRouting(t *testing.T) {
	executable := gia.CompileSourceOracleForTest(t, "if_dsd_reports_review.c")
	for verbose := 0; verbose < 2; verbose++ {
		for number := 0; number <= 2; number += 2 {
			for support := 0; support <= 3; support += 3 {
				for occurs := 0; occurs < 2; occurs++ {
					for fileMode := 0; fileMode < 2; fileMode++ {
						t.Run(fmt.Sprintf("verbose%d/number%d/support%d/occurs%d/file%d", verbose, number, support, occurs, fileMode), func(t *testing.T) {
							path := "-"
							if fileMode != 0 {
								path = filepath.Join(t.TempDir(), "source-report.txt")
							}
							output, err := exec.Command(executable, strconv.Itoa(verbose), strconv.Itoa(number), strconv.Itoa(support), strconv.Itoa(occurs), path).CombinedOutput()
							if err != nil {
								t.Fatalf("source reports: %v %s", err, output)
							}
							p := if_.If_DsdManAlloc(4, 2)
							defer if_.If_DsdManFree(p, 0)
							texts := []string{"(abc)", "E8{abc}", "E8{a(bc)d}", "D3C5{abcd}", "<a(bc)d>"}
							sizes := []int32{3, 3, 4, 4, 4}
							for i, text := range texts {
								for j := 0; j <= i; j++ {
									if_.If_DsdManCompute(p, dau.Dau_DsdToTruth(text, sizes[i]), sizes[i], make([]byte, sizes[i]), "")
								}
							}
							target := ""
							if fileMode != 0 {
								target = filepath.Join(t.TempDir(), "go-report.txt")
							}
							actual := mappingDSDReviewCapture(t, func() { if_.If_DsdManPrint(p, target, int32(number), int32(support), int32(occurs), 0, int32(verbose)) })
							mappingDSDReviewReportEqual(t, "stdout", actual, string(output))
							if fileMode != 0 {
								a, err := os.ReadFile(target)
								if err != nil {
									t.Fatal(err)
								}
								b, err := os.ReadFile(path)
								if err != nil {
									t.Fatal(err)
								}
								mappingDSDReviewReportEqual(t, "file", string(a), string(b))
							}
						})
					}
				}
			}
		}
	}
}
