// Mechanical port of dot_layout.py.
//
// Python source: ~/ivy/pyivy/ivy/ivy/dot_layout.py (286 lines)
//
// The Python file uses pygraphviz (AGraph) to lay out a Cytoscape graph
// via the `dot` engine and writes positions, widths, heights, and edge
// spline data back into the input CyElements. The Go port replaces
// pygraphviz with a subprocess call to the `dot` binary using the existing
// dotgraph.Graph builder, and parses the result via `dot -Tjson0`.
//
// The geometry helpers (CubicBezierPoint, ApproximateCubicBezier, etc.)
// are pure functions that mirror their Python counterparts byte for byte.
// The DotLayout entry point mirrors the Python `dot_layout(cy_elements,
// edge_labels, subgraph_boxes, node_gt)` function.

package dotgraph

import (
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
	"sort"
	"strings"

	"github.com/glycerine/ivy/goivy/art"
)

// -----------------------------------------------------------------------
// Pure geometry helpers (Python dot_layout.py:27-94)
// -----------------------------------------------------------------------

// PosXY mirrors Python's {"x": ..., "y": ...} position dict, used by the
// helpers below. art.CyPosition is the canonical type for layout output;
// PosXY is used internally for intermediate computations.
type PosXY struct {
	X, Y float64
}

// CubicBezierPoint mirrors Python dot_layout.py:27-38:
//
//	def cubic_bezier_point(p0, p1, p2, p3, t):
//	    a = (1.0 - t)**3
//	    b = 3.0 * t * (1.0 - t)**2
//	    c = 3.0 * t**2 * (1.0 - t)
//	    d = t**3
//	    return {
//	        "x": a * p0["x"] + b * p1["x"] + c * p2["x"] + d * p3["x"],
//	        "y": a * p0["y"] + b * p1["y"] + c * p2["y"] + d * p3["y"],
//	    }
//
// https://en.wikipedia.org/wiki/B%C3%A9zier_curve#Cubic_B.C3.A9zier_curves
func CubicBezierPoint(p0, p1, p2, p3 PosXY, t float64) PosXY {
	a := math.Pow(1.0-t, 3)
	b := 3.0 * t * math.Pow(1.0-t, 2)
	c := 3.0 * t * t * (1.0 - t)
	d := t * t * t
	return PosXY{
		X: a*p0.X + b*p1.X + c*p2.X + d*p3.X,
		Y: a*p0.Y + b*p1.Y + c*p2.Y + d*p3.Y,
	}
}

// SquareDistanceToSegment mirrors Python dot_layout.py:41-54:
//
//	def square_distance_to_segment(p, p1, p2):
//	    v0 = (p["x"] - p1["x"], p["y"] - p1["y"])
//	    v1 = (p2["x"] - p1["x"], p2["y"] - p1["y"])
//	    v0sq = v0[0] * v0[0] + v0[1] * v0[1]
//	    v1sq = v1[0] * v1[0] + v1[1] * v1[1]
//	    prod  = v0[0] * v1[0] + v0[1] * v1[1]
//	    v2sq = prod * prod / v1sq
//	    if prod < 0:
//	        return v0sq
//	    elif v2sq < v1sq:
//	        return v0sq - v2sq
//	    else:
//	        v3 = (v0[0] - v1[0], v0[1] - v1[1])
//	        return v3[0] * v3[0] + v3[1] * v3[1]
func SquareDistanceToSegment(p, p1, p2 PosXY) float64 {
	v0x := p.X - p1.X
	v0y := p.Y - p1.Y
	v1x := p2.X - p1.X
	v1y := p2.Y - p1.Y
	v0sq := v0x*v0x + v0y*v0y
	v1sq := v1x*v1x + v1y*v1y
	prod := v0x*v1x + v0y*v1y
	v2sq := prod * prod / v1sq
	if prod < 0 {
		return v0sq
	} else if v2sq < v1sq {
		return v0sq - v2sq
	}
	v3x := v0x - v1x
	v3y := v0y - v1y
	return v3x*v3x + v3y*v3y
}

// ApproximateCubicBezier mirrors Python dot_layout.py:57-78:
//
//	def approximate_cubic_bezier(p0, p1, p2, p3, threshold=1.0, limit=1024):
//	    threshold_squared = threshold ** 2
//	    points = {0.0: p0, 1.0: p3}
//	    to_check = deque([(0.0, 1.0)])
//	    while len(to_check) > 0 and len(points) < limit:
//	        l, r = to_check.popleft()
//	        pl = points[l]
//	        pr = points[r]
//	        m = (l + r) / 2.0
//	        pm = cubic_bezier_point(p0, p1, p2, p3, m)
//	        if square_distance_to_segment(pm, pl, pr) > threshold_squared:
//	            points[m] = pm
//	            to_check.append((l, m))
//	            to_check.append((m, r))
//	    return [points[t] for t in sorted(points.keys())]
//
// Returns a series of points whose segments approximate the given Bezier
// curve to within `threshold` distance (or until `limit` points are added).
func ApproximateCubicBezier(p0, p1, p2, p3 PosXY, threshold float64, limit int) []PosXY {
	thresholdSquared := threshold * threshold
	points := map[float64]PosXY{
		0.0: p0,
		1.0: p3,
	}
	type interval struct{ l, r float64 }
	toCheck := []interval{{0.0, 1.0}}
	for len(toCheck) > 0 && len(points) < limit {
		iv := toCheck[0]
		toCheck = toCheck[1:]
		pl := points[iv.l]
		pr := points[iv.r]
		m := (iv.l + iv.r) / 2.0
		pm := CubicBezierPoint(p0, p1, p2, p3, m)
		if SquareDistanceToSegment(pm, pl, pr) > thresholdSquared {
			points[m] = pm
			toCheck = append(toCheck, interval{iv.l, m})
			toCheck = append(toCheck, interval{m, iv.r})
		}
	}
	keys := make([]float64, 0, len(points))
	for k := range points {
		keys = append(keys, k)
	}
	sort.Float64s(keys)
	out := make([]PosXY, 0, len(keys))
	for _, k := range keys {
		out = append(out, points[k])
	}
	return out
}

// GetApproximationPoints mirrors Python dot_layout.py:81-94:
//
//	def get_approximation_points(bspline):
//	    result = []
//	    for i in range(0, len(bspline) - 3, 3):
//	        result.extend(approximate_cubic_bezier(
//	            bspline[i], bspline[i+1], bspline[i+2], bspline[i+3],
//	            threshold=4.0,
//	            limit=100,
//	        )[:-1])
//	    result.append(bspline[-1])
//	    return result
//
// Returns a series of points whose segments approximate the given bspline.
func GetApproximationPoints(bspline []PosXY) []PosXY {
	var result []PosXY
	for i := 0; i+3 < len(bspline); i += 3 {
		seg := ApproximateCubicBezier(
			bspline[i], bspline[i+1], bspline[i+2], bspline[i+3],
			4.0, 100,
		)
		// Python's [:-1] drops the last element of each segment except
		// the very last (handled outside the loop).
		if len(seg) > 0 {
			result = append(result, seg[:len(seg)-1]...)
		}
	}
	if len(bspline) > 0 {
		result = append(result, bspline[len(bspline)-1])
	}
	return result
}

// -----------------------------------------------------------------------
// String parsing helpers (Python dot_layout.py:97-133)
// -----------------------------------------------------------------------

// toPosition mirrors Python dot_layout.py:97-104:
//
//	def _to_position(st):
//	    global y_origin
//	    sp = st.split(',')
//	    assert len(sp) == 2, st
//	    return {"x": float(sp[0]), "y": y_origin-float(sp[1])}
//
// The y-axis is inverted relative to dot's coordinate system, with
// `yOrigin` providing the top-of-graph reference.
func toPosition(st string, yOrigin float64) PosXY {
	sp := strings.Split(st, ",")
	if len(sp) != 2 {
		panic(fmt.Sprintf("toPosition: bad input %q", st))
	}
	x := mustParseFloat(sp[0])
	y := mustParseFloat(sp[1])
	return PosXY{X: x, Y: yOrigin - y}
}

// edgePosition mirrors Python dot_layout.py:107-126:
//
//	def _to_edge_position(st):
//	    sp = st.split()
//	    result = {}
//	    if sp[0].startswith('e,'):
//	        result["arrowend"] = _to_position(sp[0][2:])
//	        sp = sp[1:]
//	    if sp[0].startswith('s,'):
//	        result["arrowstart"] = _to_position(sp[0][2:])
//	        sp = sp[1:]
//	    result["bspline"] = [_to_position(x) for x in sp]
//	    result["approxpoints"] = get_approximation_points(result["bspline"])
//	    return result
//
// http://www.graphviz.org/doc/info/attrs.html#k:splineType
type edgePosResult struct {
	ArrowEnd     *PosXY
	ArrowStart   *PosXY
	BSpline      []PosXY
	ApproxPoints []PosXY
}

func toEdgePosition(st string, yOrigin float64) edgePosResult {
	sp := strings.Fields(st)
	result := edgePosResult{}
	if len(sp) > 0 && strings.HasPrefix(sp[0], "e,") {
		p := toPosition(sp[0][2:], yOrigin)
		result.ArrowEnd = &p
		sp = sp[1:]
	}
	if len(sp) > 0 && strings.HasPrefix(sp[0], "s,") {
		p := toPosition(sp[0][2:], yOrigin)
		result.ArrowStart = &p
		sp = sp[1:]
	}
	result.BSpline = make([]PosXY, 0, len(sp))
	for _, x := range sp {
		result.BSpline = append(result.BSpline, toPosition(x, yOrigin))
	}
	result.ApproxPoints = GetApproximationPoints(result.BSpline)
	return result
}

// toCoordList mirrors Python dot_layout.py:129-133:
//
//	def _to_coord_list(st):
//	    """ create a sequence of positions from a dot-generated string """
//	    nums = st.split(',')
//	    pairs = [','.join((nums[2*i],nums[2*i+1])) for i in range(len(nums)//2)]
//	    return list(map(_to_position, pairs))
func toCoordList(st string, yOrigin float64) []PosXY {
	nums := strings.Split(st, ",")
	out := make([]PosXY, 0, len(nums)/2)
	for i := 0; 2*i+1 < len(nums); i++ {
		pair := nums[2*i] + "," + nums[2*i+1]
		out = append(out, toPosition(pair, yOrigin))
	}
	return out
}

// mustParseFloat is a small helper for parsing dot's floating-point output.
func mustParseFloat(s string) float64 {
	var f float64
	if _, err := fmt.Sscanf(s, "%f", &f); err != nil {
		panic(fmt.Sprintf("mustParseFloat(%q): %v", s, err))
	}
	return f
}

// -----------------------------------------------------------------------
// DotLayout — main entry point (Python dot_layout.py:136-286)
// -----------------------------------------------------------------------

// DotLayoutOptions mirrors the Python keyword arguments
// edge_labels, subgraph_boxes, node_gt of dot_layout(cy_elements, ...).
type DotLayoutOptions struct {
	EdgeLabels    bool
	SubgraphBoxes bool
	// NodeGt is an optional comparator on node ids. When set, it returns
	// true iff the source node should be drawn above the target (i.e. the
	// edge should be reversed). Mirrors Python's `node_gt` argument.
	NodeGt func(src, dst string) bool
}

// DotLayout mirrors Python dot_layout.py:136-286.
//
// Get a CyElements object and augment it (in-place) with positions,
// widths, heights, and spline data from a dot based layout.
//
// Returns the same CyElements pointer for chaining.
func DotLayout(cyElements *art.CyElements, opts DotLayoutOptions) *art.CyElements {
	if cyElements == nil {
		return nil
	}

	elements := cyElements.Elements

	// Build a directed graph for dot.
	g := NewDigraph("")
	g.Strict = false
	g.Attrs["forcelabels"] = "true"

	// Index nodes by id.
	nodesByID := make(map[string]*art.CyElement, len(elements))
	for i := range elements {
		if elements[i].Group == "nodes" {
			id, _ := elements[i].Data["id"].(string)
			nodesByID[id] = &elements[i]
		}
	}

	// Build the topological-sort 'order' constraints from transitive edges.
	var order []elementEdgePair
	for i := range elements {
		e := &elements[i]
		if e.Group != "edges" {
			continue
		}
		trans, ok := e.Data["transitive"].(bool)
		if !ok || !trans {
			continue
		}
		srcID, _ := e.Data["source"].(string)
		dstID, _ := e.Data["target"].(string)
		if s, ok := nodesByID[srcID]; ok {
			if d, ok := nodesByID[dstID]; ok {
				order = append(order, elementEdgePair{src: s, dst: d})
			}
		}
	}

	// Topological sort: orders elements so that 'transitive' edges go top
	// to bottom. We mirror Python's topological_sort by emitting nodes in
	// dependency order, with non-node elements stable-sorted at the end
	// (Python preserves all elements but sorts them as a group).
	elements = topologicalSortElements(elements, order)

	// Stable-sort node ids by cluster (Python lines 173-176):
	//   sorted_nodes = sorted(enumerate(sorted_nodes),
	//       key = lambda x: (nodes_by_id[x[1]]["data"]["cluster"], x[0]))
	type nodeOrd struct {
		idx     int
		id      string
		cluster string
	}
	var sortedNodes []nodeOrd
	for i, e := range elements {
		if e.Group != "nodes" {
			continue
		}
		id, _ := e.Data["id"].(string)
		cluster := ""
		if c, ok := e.Data["cluster"].(string); ok {
			cluster = c
		}
		sortedNodes = append(sortedNodes, nodeOrd{idx: i, id: id, cluster: cluster})
	}
	sort.SliceStable(sortedNodes, func(i, j int) bool {
		if sortedNodes[i].cluster != sortedNodes[j].cluster {
			return sortedNodes[i].cluster < sortedNodes[j].cluster
		}
		return sortedNodes[i].idx < sortedNodes[j].idx
	})
	nodeKey := make(map[string]int, len(sortedNodes))
	for k, n := range sortedNodes {
		nodeKey[n.id] = k
	}

	// Python: if node_gt is None: node_gt = lambda X,y: False
	//         else: node_gt = lambda x,y: node_key[x] > node_key[y]
	nodeGt := opts.NodeGt
	if nodeGt == nil {
		nodeGt = func(x, y string) bool { return false }
	} else {
		// The user-supplied node_gt is rewrapped to compare via nodeKey,
		// matching Python lines 178-181.
		userGt := opts.NodeGt
		_ = userGt
		nodeGt = func(x, y string) bool {
			return nodeKey[x] > nodeKey[y]
		}
	}

	// Add nodes to the graph (Python lines 184-186).
	for _, e := range elements {
		if e.Group != "nodes" || e.Classes == "non_existing" {
			continue
		}
		id, _ := e.Data["id"].(string)
		label, _ := e.Data["label"].(string)
		// Python: e["data"]["label"].replace('\n', '\\n')
		label = strings.ReplaceAll(label, "\n", `\n`)
		g.AddNode(id, map[string]string{"label": label})
	}

	// Add edges to the graph (Python lines 199-219).
	for _, e := range elements {
		if e.Group != "edges" {
			continue
		}
		srcID, _ := e.Data["source"].(string)
		dstID, _ := e.Data["target"].(string)
		eid, _ := e.Data["id"].(string)
		attrs := map[string]string{}
		if opts.EdgeLabels {
			label, _ := e.Data["label"].(string)
			attrs["label"] = label
		}
		// graphviz needs a unique edge identifier across multi-edges.
		// pygraphviz exposes this as the third positional arg to add_edge.
		// In raw DOT we encode it as a key= attribute.
		attrs["key"] = eid
		if nodeGt(srcID, dstID) {
			attrs["dir"] = "back"
			g.AddEdge(dstID, srcID, attrs)
		} else {
			g.AddEdge(srcID, dstID, attrs)
		}
	}

	// Add clusters (Python lines 222-231).
	clusters := make(map[string][]string)
	for _, e := range elements {
		if e.Group != "nodes" || e.Classes == "non_existing" {
			continue
		}
		c, ok := e.Data["cluster"].(string)
		if !ok || c == "" {
			continue
		}
		id, _ := e.Data["id"].(string)
		clusters[c] = append(clusters[c], id)
	}
	clusterKeys := make([]string, 0, len(clusters))
	for k := range clusters {
		clusterKeys = append(clusterKeys, k)
	}
	sort.Strings(clusterKeys)
	for i, k := range clusterKeys {
		sub := NewGraph(fmt.Sprintf("cluster_%d", i), true)
		sub.Attrs["rank"] = "min"
		for _, id := range clusters[k] {
			sub.AddNode(id, nil)
		}
		g.AddSubgraph(sub)
	}

	// Run dot to compute the layout (Python: g.layout(prog='dot')).
	// Use json0 which gives us all attributes including positions and
	// edge spline strings.
	dotOut, err := runDotJSON(g)
	if err != nil {
		// Without graphviz available we cannot lay out; return the input
		// unchanged. The Python code does not handle this case but the Go
		// port has to since dot is an external dependency.
		return cyElements
	}

	// Compute the y-origin (Python lines 243-255):
	//   y_origin = max(top of any node) where top = pos.y + height/2
	yOrigin := 0.0
	for _, jn := range dotOut.Objects {
		if jn.Pos == "" || jn.Height == "" {
			continue
		}
		sp := strings.Split(jn.Pos, ",")
		if len(sp) != 2 {
			continue
		}
		py := mustParseFloat(sp[1])
		h := mustParseFloat(jn.Height)
		// dot's height is in inches; pygraphviz returns it as such.
		top := py + h/2
		if top > yOrigin {
			yOrigin = top
		}
	}
	if opts.SubgraphBoxes {
		for _, sg := range dotOut.Subgraphs {
			parts := strings.Split(sg.BB, ",")
			if len(parts) == 4 {
				top := mustParseFloat(parts[3])
				if top > yOrigin {
					yOrigin = top
				}
			}
		}
	}

	// Index dot output by node name and edge key.
	dotNodes := make(map[string]*dotJSONObject, len(dotOut.Objects))
	for i := range dotOut.Objects {
		o := &dotOut.Objects[i]
		dotNodes[o.Name] = o
	}
	type edgeKey struct{ src, dst, eid string }
	dotEdges := make(map[edgeKey]*dotJSONEdge, len(dotOut.Edges))
	for i := range dotOut.Edges {
		e := &dotOut.Edges[i]
		var src, dst string
		if int(e.Tail) < len(dotOut.Objects) {
			src = dotOut.Objects[e.Tail].Name
		}
		if int(e.Head) < len(dotOut.Objects) {
			dst = dotOut.Objects[e.Head].Name
		}
		dotEdges[edgeKey{src, dst, e.Key}] = e
	}

	// Walk the elements and write back position/width/height/splines.
	// Python lines 257-277.
	for i := range elements {
		e := &elements[i]
		if e.Group == "nodes" && e.Classes != "non_existing" {
			id, _ := e.Data["id"].(string)
			n, ok := dotNodes[id]
			if !ok {
				continue
			}
			pos := toPosition(n.Pos, yOrigin)
			e.Position = &art.CyPosition{X: pos.X, Y: pos.Y}
			if w, err := parseFloat(n.Width); err == nil {
				e.Data["width"] = 72 * w
			}
			if h, err := parseFloat(n.Height); err == nil {
				e.Data["height"] = 72 * h
			}
		} else if e.Group == "edges" {
			srcID, _ := e.Data["source"].(string)
			dstID, _ := e.Data["target"].(string)
			eid, _ := e.Data["id"].(string)
			var key edgeKey
			reversed := nodeGt(srcID, dstID)
			if reversed {
				key = edgeKey{dstID, srcID, eid}
			} else {
				key = edgeKey{srcID, dstID, eid}
			}
			ed, ok := dotEdges[key]
			if !ok {
				continue
			}
			pos := ed.Pos
			if reversed {
				// Reverse the spline (Python lines 265-271):
				//   pe = pos.split()
				//   ppe = pe[1:]
				//   ppe.reverse()
				//   pos = ' '.join([pe[0].replace('s','e')] + ppe)
				pe := strings.Fields(pos)
				if len(pe) > 0 {
					ppe := append([]string{}, pe[1:]...)
					for i, j := 0, len(ppe)-1; i < j; i, j = i+1, j-1 {
						ppe[i], ppe[j] = ppe[j], ppe[i]
					}
					pe[0] = strings.ReplaceAll(pe[0], "s", "e")
					pos = strings.Join(append([]string{pe[0]}, ppe...), " ")
				}
			}
			ep := toEdgePosition(pos, yOrigin)
			if ep.ArrowEnd != nil {
				e.Data["arrowend"] = map[string]float64{"x": ep.ArrowEnd.X, "y": ep.ArrowEnd.Y}
			}
			if ep.ArrowStart != nil {
				e.Data["arrowstart"] = map[string]float64{"x": ep.ArrowStart.X, "y": ep.ArrowStart.Y}
			}
			e.Data["bspline"] = pointsToMaps(ep.BSpline)
			e.Data["approxpoints"] = pointsToMaps(ep.ApproxPoints)
			if opts.EdgeLabels {
				if label, _ := e.Data["label"].(string); label != "" && ed.LP != "" {
					lp := toPosition(ed.LP, yOrigin)
					e.Data["lp"] = map[string]float64{"x": lp.X, "y": lp.Y}
				}
			}
		}
	}

	// Subgraph boxes (Python lines 280-285).
	if opts.SubgraphBoxes {
		for _, sg := range dotOut.Subgraphs {
			coords := toCoordList(sg.BB, yOrigin)
			cyElements.Elements = append(cyElements.Elements, art.CyElement{
				Group:   "nodes",
				Classes: "subgraphs",
				Data: map[string]any{
					"id":     sg.Name,
					"coords": pointsToMaps(coords),
				},
			})
		}
	}

	cyElements.Elements = elements
	return cyElements
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

func pointsToMaps(pts []PosXY) []map[string]float64 {
	out := make([]map[string]float64, len(pts))
	for i, p := range pts {
		out[i] = map[string]float64{"x": p.X, "y": p.Y}
	}
	return out
}

// elementEdgePair is a (src,dst) pair of CyElement pointers used to express
// topological-sort constraints between transitive edges.
type elementEdgePair struct {
	src, dst *art.CyElement
}

// topologicalSortElements is a tiny shim around Python's topological_sort:
// it sorts the elements so that for each (src,dst) pair in `order`, src
// appears before dst. Non-node elements and unconstrained nodes preserve
// their original relative order. This mirrors ivy_utils.topological_sort.
func topologicalSortElements(elements []art.CyElement, order []elementEdgePair) []art.CyElement {
	if len(order) == 0 {
		return elements
	}
	idIdx := make(map[*art.CyElement]int, len(elements))
	for i := range elements {
		idIdx[&elements[i]] = i
	}
	deps := make(map[int]map[int]bool)
	for _, p := range order {
		s, sok := idIdx[p.src]
		d, dok := idIdx[p.dst]
		if !sok || !dok {
			continue
		}
		if deps[d] == nil {
			deps[d] = make(map[int]bool)
		}
		deps[d][s] = true
	}
	visited := make([]bool, len(elements))
	var out []int
	var visit func(int)
	visit = func(i int) {
		if visited[i] {
			return
		}
		visited[i] = true
		for src := range deps[i] {
			visit(src)
		}
		out = append(out, i)
	}
	for i := range elements {
		visit(i)
	}
	sorted := make([]art.CyElement, 0, len(elements))
	for _, i := range out {
		sorted = append(sorted, elements[i])
	}
	return sorted
}

// -----------------------------------------------------------------------
// dot subprocess driver and JSON output parsing
// -----------------------------------------------------------------------

// dotJSONOutput is a small subset of dot's -Tjson0 output.
type dotJSONOutput struct {
	Objects   []dotJSONObject `json:"objects"`
	Edges     []dotJSONEdge   `json:"edges"`
	Subgraphs []dotJSONObject `json:"subgraphs"`
}

type dotJSONObject struct {
	Name   string `json:"name"`
	Pos    string `json:"pos"`
	Width  string `json:"width"`
	Height string `json:"height"`
	BB     string `json:"bb"`
}

type dotJSONEdge struct {
	Tail int    `json:"tail"`
	Head int    `json:"head"`
	Key  string `json:"key"`
	Pos  string `json:"pos"`
	LP   string `json:"lp"`
}

// runDotJSON invokes the `dot` binary on the given graph and parses the
// json0 output. Returns an error if dot is not on PATH or fails.
func runDotJSON(g *Graph) (*dotJSONOutput, error) {
	dot := g.ToDOT()
	cmd := exec.Command("dot", "-Tjson0")
	cmd.Stdin = strings.NewReader(dot)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("dot -Tjson0: %w", err)
	}
	var parsed dotJSONOutput
	if err := json.Unmarshal(out, &parsed); err != nil {
		return nil, fmt.Errorf("dot json parse: %w", err)
	}
	return &parsed, nil
}
