package place

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Test-only adapter to the original C implementation. The Go package has no C
// runtime, cgo or compiler dependency; this differential check skips when the
// original source or a C compiler is absent.
const qpReferenceDriver = `
#include <stdio.h>
#include <stdlib.h>
#include "place_qpsolver.h"
int main(void) {
 qps_problem_t p={0}; int ne,pr=0;
 if(scanf("%d %d %d %d",&p.num_cells,&ne,&p.cog_num,&p.loop_num)!=4)return 2;
 int n=p.num_cells;
 p.x=calloc(n,sizeof(float));p.y=calloc(n,sizeof(float));p.fixed=calloc(n,sizeof(int));p.area=calloc(n,sizeof(float));
 for(int i=0;i<n;i++)scanf("%f %f %d %f",p.x+i,p.y+i,p.fixed+i,p.area+i);
 int *a=calloc(ne,sizeof(int)),*b=calloc(ne,sizeof(int));float *w=calloc(ne,sizeof(float));
 for(int i=0;i<ne;i++)scanf("%d %d %f",a+i,b+i,w+i);
 p.connect=calloc(2*ne+n+2,sizeof(int));p.edge_weight=calloc(2*ne+n+2,sizeof(float));
 for(int v=0;v<n;v++){for(int i=0;i<ne;i++){if(a[i]==v||b[i]==v){p.connect[pr]=a[i]==v?b[i]:a[i];p.edge_weight[pr++]=w[i];}}p.connect[pr]=-1;p.edge_weight[pr++]=-1;}
 p.cog_x=calloc(p.cog_num,sizeof(float));p.cog_y=calloc(p.cog_num,sizeof(float));p.cog_list=calloc((n+1)*p.cog_num+2,sizeof(int));pr=0;
 for(int i=0;i<p.cog_num;i++){int k;scanf("%d %f %f",&k,p.cog_x+i,p.cog_y+i);for(int j=0;j<k;j++)scanf("%d",p.cog_list+pr++);p.cog_list[pr++]=-1;}
 p.loop_max=calloc(p.loop_num,sizeof(float));p.loop_penalty=calloc(p.loop_num,sizeof(float));p.loop_list=calloc((n+1)*p.loop_num+2,sizeof(int));pr=0;
 for(int i=0;i<p.loop_num;i++){int k;scanf("%d %f %f",&k,p.loop_max+i,p.loop_penalty+i);for(int j=0;j<k;j++)scanf("%d",p.loop_list+pr++);p.loop_list[pr++]=-1;}
 qps_init(&p);qps_solve(&p);printf("%d %.9g",p.loop_fail,p.f);for(int i=0;i<n;i++)printf(" %.9g %.9g",p.x[i],p.y[i]);puts("");qps_clean(&p);return 0;
}
`

func findQPReference(t *testing.T) string {
	t.Helper()
	if source := os.Getenv("ABC_QP_REFERENCE_SOURCE"); source != "" {
		if _, err := os.Stat(source); err != nil {
			t.Fatal(err)
		}
		return source
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for dir := cwd; ; dir = filepath.Dir(dir) {
		candidate := filepath.Join(dir, "src/phys/place/place_qpsolver.c")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		if filepath.Dir(dir) == dir {
			break
		}
	}
	t.Skip("original QP C source unavailable")
	return ""
}
func TestQPOriginalCComparison(t *testing.T) {
	compiler, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable for optional differential check")
	}
	source := findQPReference(t)
	root := t.TempDir()
	driver := filepath.Join(root, "driver.c")
	if err := os.WriteFile(driver, []byte(qpReferenceDriver), 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(root, "qp-reference")
	args := []string{"-std=c99", "-O0", "-DABC_NAMESPACE_HEADER_START=", "-DABC_NAMESPACE_HEADER_END=", "-DABC_NAMESPACE_IMPL_START=", "-DABC_NAMESPACE_IMPL_END=", "-I", filepath.Dir(source), source, driver, "-lm", "-o", binary}
	if out, err := exec.Command(compiler, args...).CombinedOutput(); err != nil {
		t.Fatalf("reference build: %v: %s", err, out)
	}
	rng := rand.New(rand.NewSource(1107))
	for trial := 0; trial < 40; trial++ {
		n := 3 + rng.Intn(6)
		x, y := make([]float32, n), make([]float32, n)
		fixed := make([]bool, n)
		areas := make([]float32, n)
		for i := range x {
			x[i] = float32(rng.Intn(100)-50) / 7
			y[i] = float32(rng.Intn(100)-50) / 9
			areas[i] = float32(1 + rng.Intn(4))
		}
		fixed[0], fixed[n-1] = true, true
		var edges []QPEdge
		for i := 1; i < n; i++ {
			edges = append(edges, QPEdge{int32(i - 1), int32(i), float32(1 + rng.Intn(5))})
		}
		p, err := NewQPProblem(x, y, fixed, edges)
		if err != nil {
			t.Fatal(err)
		}
		p.SetAreas(areas)
		withCOG := trial%2 != 0
		var members []int32
		if withCOG {
			for i := 1; i < n-1; i++ {
				members = append(members, int32(i))
			}
			p.AddCenterOfGravity(members, 1.25, -.75)
		}
		if err := p.Solve(); err != nil {
			t.Fatal(trial, err)
		}
		var input strings.Builder
		groups := 0
		if withCOG {
			groups = 1
		}
		fmt.Fprintf(&input, "%d %d %d 0\n", n, len(edges), groups)
		for i := range x {
			f := 0
			if fixed[i] {
				f = 1
			}
			fmt.Fprintf(&input, "%.9g %.9g %d %.9g\n", x[i], y[i], f, areas[i])
		}
		for _, e := range edges {
			fmt.Fprintf(&input, "%d %d %.9g\n", e.A, e.B, e.Weight)
		}
		if withCOG {
			fmt.Fprintf(&input, "%d 1.25 -.75", len(members))
			for _, v := range members {
				fmt.Fprintf(&input, " %d", v)
			}
			input.WriteByte('\n')
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		cmd := exec.CommandContext(ctx, binary)
		cmd.Stdin = strings.NewReader(input.String())
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		out, runErr := cmd.Output()
		cancel()
		if runErr != nil {
			t.Fatalf("C case %d: %v: %s", trial, runErr, stderr.String())
		}
		reader := bytes.NewReader(out)
		var flags int32
		var objective float32
		if _, err := fmt.Fscan(reader, &flags, &objective); err != nil {
			t.Fatal(err, string(out))
		}
		if math.Abs(float64(p.f-objective)) > .004*math.Max(1, math.Abs(float64(objective))) {
			t.Fatal(trial, "objective", p.f, objective)
		}
		for i := range x {
			var cx, cy float32
			if _, err := fmt.Fscan(reader, &cx, &cy); err != nil {
				t.Fatal(err, string(out))
			}
			for _, pair := range [][2]float32{{p.x[i], cx}, {p.y[i], cy}} {
				if math.Abs(float64(pair[0]-pair[1])) > .004*math.Max(1, math.Abs(float64(pair[1]))) {
					t.Fatal(trial, i, pair, flags)
				}
			}
		}
	}
}
