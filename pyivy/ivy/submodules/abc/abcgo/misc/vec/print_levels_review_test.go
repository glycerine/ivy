package vec

import (
	"os"
	"slices"
	"testing"
)

func vecPrintReviewCapture(t *testing.T, run func()) string {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "level-report")
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = file
	defer func() { os.Stdout = old; file.Close() }()
	run()
	data, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestIntLevelPrintersPreserveSourceEmptyPolicies(t *testing.T) {
	groups := [][]int32{nil, {7}, {-1, 3, 3}, nil, {9}, {2, 4}, nil}
	cases := []struct {
		name string
		wec  bool
		skip int
		want string
	}{
		{"pointer levels/all", false, 0, "    1 : { 7 }\n    2 : { -1 3 3 }\n    4 : { 9 }\n    5 : { 2 4 }\n"},
		{"pointer levels/skip singles", false, 1, "    2 : { -1 3 3 }\n    5 : { 2 4 }\n"},
		{"embedded levels/all", true, 0, "    0 : { }\n    1 : { 7 }\n    2 : { -1 3 3 }\n    3 : { }\n    4 : { 9 }\n    5 : { 2 4 }\n    6 : { }\n"},
		{"embedded levels/skip singles", true, 1, "    0 : { }\n    2 : { -1 3 3 }\n    3 : { }\n    5 : { 2 4 }\n    6 : { }\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pointers := Vec_VecAlloc(0)
			Vec_VecExpandInt(pointers, len(groups)-1)
			embedded := Vec_WecStart(len(groups))
			for i, group := range groups {
				for _, value := range group {
					Vec_VecPushInt(pointers, i, value)
					Vec_WecPush(embedded, i, value)
				}
			}
			got := vecPrintReviewCapture(t, func() {
				if tc.wec {
					Vec_WecPrint(embedded, tc.skip)
				} else {
					Vec_VecPrintInt(pointers, tc.skip)
				}
			})
			if got != tc.want {
				t.Errorf("printed%q want%q", got, tc.want)
			}
			for i, want := range groups {
				if !slices.Equal(Vec_IntArray(Vec_VecEntryInt(pointers, i)), want) || !slices.Equal(Vec_IntArray(Vec_WecEntry(embedded, i)), want) {
					t.Fatalf("printing changed level%d", i)
				}
			}
		})
	}
}
