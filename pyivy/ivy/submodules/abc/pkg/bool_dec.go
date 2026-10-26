package abc

// Factoring, formatting, truth derivation, and network conversion from
// src/bool/dec.

import (
	"strings"
)

// DecFactorLimits bounds the work and retained storage of SOP factoring and
// concrete AIG conversion. Zero values select finite package defaults.
type DecFactorLimits struct {
	MaxVariables int
	MaxCubes     int
	MaxCubeBytes int
	MaxNodes     int
	MaxDepth     int
	Stop         func() bool
}

func checkDecFactorLimits(limits DecFactorLimits) (DecFactorLimits, error) {
	if limits.MaxVariables < 0 || limits.MaxCubes < 0 || limits.MaxCubeBytes < 0 || limits.MaxNodes < 0 || limits.MaxDepth < 0 {
		return limits, ErrDimension
	}
	if limits.MaxVariables == 0 {
		limits.MaxVariables = 1 << 16
	}
	if limits.MaxCubes == 0 {
		limits.MaxCubes = 1 << 20
	}
	if limits.MaxCubeBytes == 0 {
		limits.MaxCubeBytes = 16 << 20
	}
	if limits.MaxNodes == 0 {
		limits.MaxNodes = 1 << 22
	}
	if limits.MaxDepth == 0 {
		limits.MaxDepth = 1 << 16
	}
	if limits.MaxVariables > limits.MaxNodes {
		limits.MaxVariables = limits.MaxNodes
	}
	if limits.MaxVariables >= decoEdgeNodeLimit || limits.MaxNodes > decoEdgeNodeLimit {
		return limits, ErrLimit
	}
	return limits, nil
}

// StartDecoManager implements Dec_ManStart. The NPN arrays have the same
// 65,536-entry layout as Extra_Truth4VarNPN and are independently owned.
func StartDecoManager() (*DecoManagerState, error) {
	npn, err := NewDecManager()
	if err != nil {
		return nil, err
	}
	canons, phases, permutations, classes, err := npn.NPN4Tables()
	closeErr := npn.Close()
	if err != nil {
		return nil, err
	}
	if closeErr != nil {
		return nil, closeErr
	}
	manager := &DecoManagerState{
		MVC:          NewMvcManager(),
		Cubes:        make([]int, 0, 8),
		Literals:     make([]int, 0, 8),
		Canons:       canons,
		Phases:       make([]int8, len(phases)),
		Permutations: make([]int8, len(permutations)),
		ClassMap:     classes,
	}
	for index := range phases {
		manager.Phases[index] = int8(phases[index])
		manager.Permutations[index] = int8(permutations[index])
	}
	return manager, nil
}

// Stop implements Dec_ManStop and releases all manager-owned arrays.
func (manager *DecoManagerState) Stop() error {
	if manager == nil || manager.MVC == nil {
		return ErrInvariant
	}
	if err := manager.MVC.Close(); err != nil {
		return err
	}
	manager.MVC = nil
	manager.Cubes = nil
	manager.Literals = nil
	manager.Canons = nil
	manager.Phases = nil
	manager.Permutations = nil
	manager.ClassMap = nil
	return nil
}

func decSOPCheck(cover SOPCover, limits DecFactorLimits) error {
	if cover.Variables < 0 || len(cover.Cubes) == 0 || !strings.ContainsRune("01nx", rune(cover.Output)) {
		return ErrDimension
	}
	if cover.Variables > limits.MaxVariables || len(cover.Cubes) > limits.MaxCubes || cover.Variables > limits.MaxCubeBytes/len(cover.Cubes) {
		return ErrLimit
	}
	for _, cube := range cover.Cubes {
		if len(cube) != cover.Variables {
			return ErrDimension
		}
		for index := range cube {
			if cube[index] != '-' && cube[index] != '0' && cube[index] != '1' {
				return ErrDimension
			}
		}
	}
	return nil
}

// FactorSOPText is the char-buffer entry point corresponding to Dec_Factor.
func FactorSOPText(text string, limits DecFactorLimits) (*DecoGraph, error) {
	limits, err := checkDecFactorLimits(limits)
	if err != nil {
		return nil, err
	}
	cover, err := ParseSOPCover(text, SOPExpansionOptions{
		MaxVariables:   limits.MaxVariables,
		MaxCubes:       limits.MaxCubes,
		MaxCubeEntries: limits.MaxCubeBytes,
		Stop:           limits.Stop,
	})
	if err != nil {
		return nil, err
	}
	return FactorSOP(cover, limits)
}

// FactorSOP ports Dec_Factor and retains its algebraic MVC factoring order.
// Output codes x/n are treated like source SOP buffers: the cubes are
// factored as an OR cover and n applies output complementation.
func FactorSOP(sop SOPCover, limits DecFactorLimits) (*DecoGraph, error) {
	limits, err := checkDecFactorLimits(limits)
	if err != nil {
		return nil, err
	}
	if err = decSOPCheck(sop, limits); err != nil {
		return nil, err
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	if sop.Variables == 0 {
		if len(sop.Cubes) != 1 || sop.Cubes[0] != "" || (sop.Output != '0' && sop.Output != '1') {
			return nil, ErrDimension
		}
		return NewConstantDecoGraph(sop.Output == '1', DecoGraphLimits{MaxNodes: limits.MaxNodes, Stop: limits.Stop})
	}

	mvcManager := NewMvcManager()
	mvc, err := decConvertSOPToMVC(mvcManager, sop)
	if err != nil {
		_ = mvcManager.Close()
		return nil, err
	}
	defer func() {
		_ = mvc.Close()
		_ = mvcManager.Close()
	}()
	if _, err = mvc.RemoveContained(); err != nil {
		return nil, err
	}
	empty, err := mvc.IsEmpty()
	if err != nil {
		return nil, err
	}
	tautology, err := mvc.IsTautology()
	if err != nil {
		return nil, err
	}
	if empty || tautology {
		return nil, ErrInvariant
	}
	if err = mvc.Invert(); err != nil {
		return nil, err
	}
	graph, err := NewDecoGraph(sop.Variables, DecoGraphLimits{MaxLeaves: limits.MaxVariables, MaxNodes: limits.MaxNodes, Stop: limits.Stop})
	if err != nil {
		return nil, err
	}
	factor := decFactorer{graph: graph, limits: limits}
	root, err := factor.factor(mvc, 0)
	if err != nil {
		graph.Close()
		return nil, err
	}
	if err = graph.SetRoot(root); err != nil {
		graph.Close()
		return nil, err
	}
	if sop.Output == '0' || sop.Output == 'n' {
		if err = graph.Complement(); err != nil {
			graph.Close()
			return nil, err
		}
	}
	return graph, nil
}

func decConvertSOPToMVC(manager *MvcManager, sop SOPCover) (*MvcCover, error) {
	cover, err := NewMvcCover(manager, sop.Variables*2)
	if err != nil {
		return nil, err
	}
	for _, inputCube := range sop.Cubes {
		cube, cubeErr := cover.NewCube()
		if cubeErr != nil {
			_ = cover.Close()
			return nil, cubeErr
		}
		for word := range cube.words {
			cube.words[word] = ^uint32(0)
		}
		cube.cleanUnused()
		for variable := range inputCube {
			switch inputCube[variable] {
			case '0':
				cube.words[(2*variable+1)>>5] &^= uint32(1) << uint((2*variable+1)&31)
			case '1':
				cube.words[(2*variable)>>5] &^= uint32(1) << uint((2*variable)&31)
			}
		}
		if cubeErr = cover.AddTail(cube); cubeErr != nil {
			_ = cube.Close()
			_ = cover.Close()
			return nil, cubeErr
		}
	}
	return cover, nil
}

type decFactorer struct {
	graph  *DecoGraph
	limits DecFactorLimits
	steps  int
}

func (factor *decFactorer) check(depth int) error {
	if depth > factor.limits.MaxDepth || factor.steps >= factor.limits.MaxNodes*8 {
		return ErrLimit
	}
	factor.steps++
	if factor.limits.Stop != nil && factor.limits.Stop() {
		return ErrCancelled
	}
	return nil
}

func (factor *decFactorer) factor(cover *MvcCover, depth int) (DecoEdge, error) {
	if err := factor.check(depth); err != nil {
		return DecoEdge{}, err
	}
	if cover.CubeCount() == 0 {
		return DecoEdge{}, ErrInvariant
	}
	divisor, err := cover.Divisor()
	if err != nil {
		return DecoEdge{}, err
	}
	if divisor == nil {
		return factor.trivial(cover, depth+1)
	}
	quotient, remainder, err := cover.DivideInternal(divisor)
	_ = divisor.Close()
	if err != nil {
		return DecoEdge{}, err
	}
	_ = remainder.Close()
	if quotient.CubeCount() == 1 {
		edge, factorErr := factor.factorLeaf(cover, quotient, depth+1)
		_ = quotient.Close()
		return edge, factorErr
	}
	if err = quotient.MakeCubeFree(); err != nil {
		_ = quotient.Close()
		return DecoEdge{}, err
	}
	divisor, remainder, err = cover.DivideInternal(quotient)
	if err != nil {
		_ = quotient.Close()
		return DecoEdge{}, err
	}
	cubeFree, err := divisor.IsCubeFree()
	if err != nil {
		_ = divisor.Close()
		_ = quotient.Close()
		_ = remainder.Close()
		return DecoEdge{}, err
	}
	if cubeFree {
		edgeDivisor, factorErr := factor.factor(divisor, depth+1)
		if factorErr == nil {
			var edgeQuotient DecoEdge
			edgeQuotient, factorErr = factor.factor(quotient, depth+1)
			if factorErr == nil {
				var edgeAnd DecoEdge
				edgeAnd, factorErr = factor.graph.AddAND(edgeDivisor, edgeQuotient)
				if factorErr == nil {
					if remainder.CubeCount() == 0 {
						edgeDivisor = edgeAnd
					} else {
						var edgeRemainder DecoEdge
						edgeRemainder, factorErr = factor.factor(remainder, depth+1)
						if factorErr == nil {
							edgeDivisor, factorErr = factor.graph.AddOR(edgeAnd, edgeRemainder)
						}
					}
				}
			}
		}
		_ = divisor.Close()
		_ = quotient.Close()
		_ = remainder.Close()
		return edgeDivisor, factorErr
	}
	common, commonErr := divisor.CommonCubeCover()
	_ = divisor.Close()
	_ = quotient.Close()
	_ = remainder.Close()
	if commonErr != nil {
		return DecoEdge{}, commonErr
	}
	edge, factorErr := factor.factorLeaf(cover, common, depth+1)
	_ = common.Close()
	return edge, factorErr
}

func (factor *decFactorer) factorLeaf(cover, simple *MvcCover, depth int) (DecoEdge, error) {
	if err := factor.check(depth); err != nil {
		return DecoEdge{}, err
	}
	divisor, err := cover.BestLiteralCover(simple)
	if err != nil {
		return DecoEdge{}, err
	}
	quotient, remainder, err := cover.DivideByLiteral(divisor)
	if err != nil {
		_ = divisor.Close()
		return DecoEdge{}, err
	}
	head, err := divisor.Head()
	if err != nil || head == nil {
		_ = divisor.Close()
		_ = quotient.Close()
		_ = remainder.Close()
		if err != nil {
			return DecoEdge{}, err
		}
		return DecoEdge{}, ErrInvariant
	}
	edgeDivisor, factorErr := factor.trivialCube(divisor, head, depth+1)
	_ = divisor.Close()
	if factorErr == nil {
		var edgeQuotient DecoEdge
		edgeQuotient, factorErr = factor.factor(quotient, depth+1)
		if factorErr == nil {
			var edgeAnd DecoEdge
			edgeAnd, factorErr = factor.graph.AddAND(edgeDivisor, edgeQuotient)
			if factorErr == nil {
				if remainder.CubeCount() == 0 {
					edgeDivisor = edgeAnd
				} else {
					var edgeRemainder DecoEdge
					edgeRemainder, factorErr = factor.factor(remainder, depth+1)
					if factorErr == nil {
						edgeDivisor, factorErr = factor.graph.AddOR(edgeAnd, edgeRemainder)
					}
				}
			}
		}
	}
	_ = quotient.Close()
	_ = remainder.Close()
	return edgeDivisor, factorErr
}

func (factor *decFactorer) trivial(cover *MvcCover, depth int) (DecoEdge, error) {
	if err := factor.check(depth); err != nil {
		return DecoEdge{}, err
	}
	cubes, err := cover.Cubes()
	if err != nil {
		return DecoEdge{}, err
	}
	edges := make([]DecoEdge, 0, len(cubes))
	for _, cube := range cubes {
		edge, cubeErr := factor.trivialCube(cover, cube, depth+1)
		if cubeErr != nil {
			return DecoEdge{}, cubeErr
		}
		edges = append(edges, edge)
	}
	return factor.balanced(edges, true, depth+1)
}

func (factor *decFactorer) trivialCube(cover *MvcCover, cube *MvcCube, depth int) (DecoEdge, error) {
	if err := factor.check(depth); err != nil {
		return DecoEdge{}, err
	}
	if err := cube.check(); err != nil || cube.cover != cover {
		if err != nil {
			return DecoEdge{}, err
		}
		return DecoEdge{}, ErrInvariant
	}
	edges := make([]DecoEdge, 0, cover.bits)
	for bit := 0; bit < cover.bits; bit++ {
		if cube.words[bit>>5]&(uint32(1)<<uint(bit&31)) != 0 {
			edges = append(edges, DecoEdge{Node: bit / 2, Complement: bit&1 != 0})
		}
	}
	return factor.balanced(edges, false, depth+1)
}

func (factor *decFactorer) balanced(edges []DecoEdge, useOR bool, depth int) (DecoEdge, error) {
	if err := factor.check(depth); err != nil {
		return DecoEdge{}, err
	}
	if len(edges) == 0 {
		return DecoEdge{}, ErrInvariant
	}
	if len(edges) == 1 {
		return edges[0], nil
	}
	middle := len(edges) / 2
	left, err := factor.balanced(edges[:middle], useOR, depth+1)
	if err != nil {
		return DecoEdge{}, err
	}
	right, err := factor.balanced(edges[middle:], useOR, depth+1)
	if err != nil {
		return DecoEdge{}, err
	}
	if useOR {
		return factor.graph.AddOR(left, right)
	}
	return factor.graph.AddAND(left, right)
}

// DecoGraphTruth32 implements Dec_GraphDeriveTruth for at most five leaves.
func DecoGraphTruth32(graph *DecoGraph) (uint32, error) {
	if err := graph.check(); err != nil {
		return 0, err
	}
	if graph.leaves > 5 {
		return 0, ErrDimension
	}
	if graph.stopped() {
		return 0, ErrCancelled
	}
	if graph.constant {
		if graph.root.Complement {
			return 0, nil
		}
		return ^uint32(0), nil
	}
	truths := [...]uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0, 0xff00ff00, 0xffff0000}
	values := make([]uint32, len(graph.nodes))
	copy(values, truths[:graph.leaves])
	for index := graph.leaves; index < len(graph.nodes); index++ {
		if graph.stopped() {
			return 0, ErrCancelled
		}
		node := graph.nodes[index]
		left, right := values[node.Edge0.Node], values[node.Edge1.Node]
		if node.Edge0.Complement {
			left = ^left
		}
		if node.Edge1.Complement {
			right = ^right
		}
		values[index] = left & right
	}
	if graph.root.Node < 0 || graph.root.Node >= len(values) {
		return 0, ErrInvariant
	}
	result := values[graph.root.Node]
	if graph.root.Complement {
		result = ^result
	}
	return result, nil
}

// DecoGraphTruthTable is the owned, backend-independent counterpart of the
// optional CUDD Dec_GraphDeriveBdd routine. It applies the same leaf order,
// complemented edges, AND nodes, and complemented root.
func DecoGraphTruthTable(graph *DecoGraph, maxWords int) (*TruthTable, error) {
	if err := graph.check(); err != nil {
		return nil, err
	}
	if graph.stopped() {
		return nil, ErrCancelled
	}
	if graph.constant {
		return NewConstantTruthTable(graph.leaves, !graph.root.Complement, maxWords)
	}
	values := make([]*TruthTable, len(graph.nodes))
	for variable := 0; variable < graph.leaves; variable++ {
		var err error
		values[variable], err = NewVariableTruthTable(graph.leaves, variable, maxWords)
		if err != nil {
			return nil, err
		}
	}
	for index := graph.leaves; index < len(graph.nodes); index++ {
		if graph.stopped() {
			return nil, ErrCancelled
		}
		node := graph.nodes[index]
		var err error
		values[index], err = values[node.Edge0.Node].And(values[node.Edge1.Node], node.Edge0.Complement, node.Edge1.Complement)
		if err != nil {
			return nil, err
		}
	}
	if graph.root.Node < 0 || graph.root.Node >= len(values) {
		return nil, ErrInvariant
	}
	result := values[graph.root.Node].Clone()
	if graph.root.Complement {
		result = result.Not()
	}
	return result, nil
}

// VerifyDecoFactor maps the optional CUDD Dec_FactorVerify helper using owned
// packed truth tables. maxWords bounds both compared representations.
func VerifyDecoFactor(sop SOPCover, graph *DecoGraph, maxWords int) (bool, error) {
	if graph == nil || graph.leaves != sop.Variables {
		return false, ErrDimension
	}
	limits, err := checkDecFactorLimits(DecFactorLimits{})
	if err != nil {
		return false, err
	}
	if err = decSOPCheck(sop, limits); err != nil {
		return false, err
	}
	actual, err := DecoGraphTruthTable(graph, maxWords)
	if err != nil {
		return false, err
	}
	expected, err := NewConstantTruthTable(sop.Variables, false, maxWords)
	if err != nil {
		return false, err
	}
	for _, cube := range sop.Cubes {
		term, termErr := NewConstantTruthTable(sop.Variables, true, maxWords)
		if termErr != nil {
			return false, termErr
		}
		for variable := range cube {
			if cube[variable] == '-' {
				continue
			}
			literal, literalErr := NewVariableTruthTable(sop.Variables, variable, maxWords)
			if literalErr != nil {
				return false, literalErr
			}
			term, termErr = term.And(literal, false, cube[variable] == '0')
			if termErr != nil {
				return false, termErr
			}
		}
		expected, err = expected.Or(term)
		if err != nil {
			return false, err
		}
	}
	if sop.Output == '0' || sop.Output == 'n' {
		expected = expected.Not()
	}
	return actual.Equal(expected), nil
}

// FormatDecoGraph implements Dec_GraphPrint and returns owned text instead of
// writing through a FILE pointer.
func FormatDecoGraph(graph *DecoGraph, inputNames []string, outputName string) (string, error) {
	if err := graph.check(); err != nil {
		return "", err
	}
	if graph.stopped() {
		return "", ErrCancelled
	}
	if inputNames == nil {
		inputNames = make([]string, graph.leaves)
		for index := range inputNames {
			if graph.leaves < 26 {
				inputNames[index] = string(rune('a' + index))
			} else {
				inputNames[index] = string(rune('a'+index%26)) + string(rune('0'+index/26))
			}
		}
	}
	if len(inputNames) != graph.leaves {
		return "", ErrDimension
	}
	maxLiteral := 0
	for _, name := range inputNames {
		if len(name) > maxLiteral {
			maxLiteral = len(name)
		}
	}
	if maxLiteral > 50 {
		maxLiteral = 20
	}
	if outputName == "" {
		outputName = "F"
	}
	var builder strings.Builder
	if len(outputName) < 6 {
		builder.WriteString(strings.Repeat(" ", 6-len(outputName)))
	}
	builder.WriteString(outputName)
	builder.WriteString(" = ")
	position := 10
	if graph.constant {
		builder.WriteString("Constant ")
		if graph.root.Complement {
			builder.WriteByte('0')
		} else {
			builder.WriteByte('1')
		}
	} else {
		if graph.root.Node < 0 || graph.root.Node >= len(graph.nodes) {
			return "", ErrInvariant
		}
		if err := decFormatNode(&builder, graph, graph.root.Node, graph.root.Complement, inputNames, &position, maxLiteral); err != nil {
			return "", err
		}
	}
	builder.WriteByte('\n')
	return builder.String(), nil
}

// FormatDecoGraphOriginal maps the alternate Dec_GraphPrint2_rec traversal,
// which uses the retained original AND/OR metadata instead of De Morgan form.
func FormatDecoGraphOriginal(graph *DecoGraph, inputNames []string, outputName string) (string, error) {
	if err := graph.check(); err != nil {
		return "", err
	}
	if graph.stopped() {
		return "", ErrCancelled
	}
	if inputNames == nil {
		inputNames = make([]string, graph.leaves)
		for index := range inputNames {
			if graph.leaves < 26 {
				inputNames[index] = string(rune('a' + index))
			} else {
				inputNames[index] = string(rune('a'+index%26)) + string(rune('0'+index/26))
			}
		}
	}
	if len(inputNames) != graph.leaves {
		return "", ErrDimension
	}
	maxLiteral := 0
	for _, name := range inputNames {
		if len(name) > maxLiteral {
			maxLiteral = len(name)
		}
	}
	if maxLiteral > 50 {
		maxLiteral = 20
	}
	if outputName == "" {
		outputName = "F"
	}
	var builder strings.Builder
	if len(outputName) < 6 {
		builder.WriteString(strings.Repeat(" ", 6-len(outputName)))
	}
	builder.WriteString(outputName)
	builder.WriteString(" = ")
	position := 10
	if graph.constant {
		builder.WriteString("Constant ")
		if graph.root.Complement {
			builder.WriteByte('0')
		} else {
			builder.WriteByte('1')
		}
	} else if err := decFormatOriginalNode(&builder, graph, graph.root.Node, graph.root.Complement, inputNames, &position, maxLiteral); err != nil {
		return "", err
	}
	builder.WriteByte('\n')
	return builder.String(), nil
}

func decFormatOriginalNode(builder *strings.Builder, graph *DecoGraph, index int, complement bool, names []string, position *int, maxLiteral int) error {
	if graph.stopped() {
		return ErrCancelled
	}
	if index < 0 || index >= len(graph.nodes) {
		return ErrInvariant
	}
	if index < graph.leaves {
		if complement {
			builder.WriteByte('!')
			*position++
		}
		builder.WriteString(names[index])
		*position += len(names[index])
		return nil
	}
	node := graph.nodes[index]
	writeChild := func(edge DecoEdge, originalComplement bool) error {
		child := graph.nodes[edge.Node]
		if !child.NodeOr {
			return decFormatOriginalNode(builder, graph, edge.Node, originalComplement, names, position, maxLiteral)
		}
		builder.WriteByte('(')
		*position++
		if err := decFormatOriginalNode(builder, graph, edge.Node, originalComplement, names, position, maxLiteral); err != nil {
			return err
		}
		builder.WriteByte(')')
		*position++
		return nil
	}
	if !node.NodeOr {
		if err := writeChild(node.Edge0, node.OriginalComplement0); err != nil {
			return err
		}
		builder.WriteByte(' ')
		*position++
		decFormatUpdatePosition(builder, position, maxLiteral)
		return writeChild(node.Edge1, node.OriginalComplement1)
	}
	if err := decFormatOriginalNode(builder, graph, node.Edge0.Node, node.OriginalComplement0, names, position, maxLiteral); err != nil {
		return err
	}
	builder.WriteString(" + ")
	*position += 3
	decFormatUpdatePosition(builder, position, maxLiteral)
	return decFormatOriginalNode(builder, graph, node.Edge1.Node, node.OriginalComplement1, names, position, maxLiteral)
}

func decFormatNode(builder *strings.Builder, graph *DecoGraph, index int, complement bool, names []string, position *int, maxLiteral int) error {
	if graph.stopped() {
		return ErrCancelled
	}
	if index < 0 || index >= len(graph.nodes) {
		return ErrInvariant
	}
	if index < graph.leaves {
		if complement {
			builder.WriteByte('!')
			*position++
		}
		builder.WriteString(names[index])
		*position += len(names[index])
		return nil
	}
	node := graph.nodes[index]
	left, right := node.Edge0.Node, node.Edge1.Node
	if left >= graph.leaves && right >= graph.leaves {
		leftNode, rightNode := graph.nodes[left], graph.nodes[right]
		l0, l1 := leftNode.Edge0.Node, leftNode.Edge1.Node
		r0, r1 := rightNode.Edge0.Node, rightNode.Edge1.Node
		if (l0 == r0 || l0 == r1) && (l1 == r0 || l1 == r1) {
			builder.WriteByte('(')
			*position++
			if err := decFormatNode(builder, graph, l0, graph.nodes[l0].OriginalComplement0, names, position, maxLiteral); err != nil {
				return err
			}
			builder.WriteString(" # ")
			*position += 3
			if err := decFormatNode(builder, graph, l1, graph.nodes[l1].OriginalComplement1, names, position, maxLiteral); err != nil {
				return err
			}
			builder.WriteByte(')')
			*position++
			return nil
		}
	}
	builder.WriteByte('(')
	*position++
	if complement {
		if err := decFormatNode(builder, graph, left, !node.Edge0.Complement, names, position, maxLiteral); err != nil {
			return err
		}
		builder.WriteString(" + ")
		*position += 3
		decFormatUpdatePosition(builder, position, maxLiteral)
		if err := decFormatNode(builder, graph, right, !node.Edge1.Complement, names, position, maxLiteral); err != nil {
			return err
		}
	} else {
		if err := decFormatNode(builder, graph, left, node.Edge0.Complement, names, position, maxLiteral); err != nil {
			return err
		}
		if err := decFormatNode(builder, graph, right, node.Edge1.Complement, names, position, maxLiteral); err != nil {
			return err
		}
	}
	builder.WriteByte(')')
	*position++
	return nil
}

func decFormatUpdatePosition(builder *strings.Builder, position *int, maxLiteral int) {
	if *position+maxLiteral < 77 {
		return
	}
	builder.WriteByte('\n')
	builder.WriteString("          ")
	*position = 10
}

// DecAIGEdge is a complemented edge in the concrete ABC/HOP/Ivy replacement.
type DecAIGEdge struct {
	Node       int
	Complement bool
}

func (edge DecAIGEdge) Not() DecAIGEdge {
	edge.Complement = !edge.Complement
	return edge
}

type DecAIGNode struct {
	Left, Right DecAIGEdge
	Variable    int
	Level       int
	Marked      bool
	Alive       bool
}

type decAIGKey struct {
	leftNode, rightNode             int
	leftComplement, rightComplement bool
}

// DecAIG is the concrete strashed graph used for the three source network
// backends. Node zero is constant one; variables and ANDs have stable indexes.
type DecAIG struct {
	nodes     []DecAIGNode
	variables []DecAIGEdge
	unique    map[decAIGKey]int
	limits    DecFactorLimits
	closed    bool
}

func NewDecAIG(variableCount int, limits DecFactorLimits) (*DecAIG, error) {
	limits, err := checkDecFactorLimits(limits)
	if err != nil || variableCount < 0 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if variableCount > limits.MaxVariables || variableCount+1 > limits.MaxNodes {
		return nil, ErrLimit
	}
	manager := &DecAIG{nodes: make([]DecAIGNode, 1, variableCount+1), variables: make([]DecAIGEdge, variableCount), unique: make(map[decAIGKey]int), limits: limits}
	manager.nodes[0] = DecAIGNode{Variable: -1, Alive: true}
	for variable := 0; variable < variableCount; variable++ {
		index := len(manager.nodes)
		manager.nodes = append(manager.nodes, DecAIGNode{Variable: variable, Alive: true})
		manager.variables[variable] = DecAIGEdge{Node: index}
	}
	return manager, nil
}

func (manager *DecAIG) check() error {
	if manager == nil || manager.closed || len(manager.nodes) == 0 || manager.nodes[0].Variable != -1 || len(manager.nodes) > manager.limits.MaxNodes {
		return ErrInvariant
	}
	return nil
}

func (manager *DecAIG) stopped() bool { return manager.limits.Stop != nil && manager.limits.Stop() }

func (manager *DecAIG) Constant(value bool) DecAIGEdge { return DecAIGEdge{Complement: !value} }

func (manager *DecAIG) Variable(index int) (DecAIGEdge, error) {
	if err := manager.check(); err != nil {
		return DecAIGEdge{}, err
	}
	if index < 0 || index >= len(manager.variables) {
		return DecAIGEdge{}, ErrDimension
	}
	return manager.variables[index], nil
}

func decAIGPacked(edge DecAIGEdge) uint64 {
	value := uint64(edge.Node) << 1
	if edge.Complement {
		value++
	}
	return value
}

func (manager *DecAIG) validateEdge(edge DecAIGEdge) error {
	if edge.Node < 0 || edge.Node >= len(manager.nodes) || !manager.nodes[edge.Node].Alive {
		return ErrDimension
	}
	return nil
}

func (manager *DecAIG) lookupAND(left, right DecAIGEdge) (DecAIGEdge, bool) {
	// Abc_AigAndLookup performs the same trivial reductions as insertion.
	// Counting these as absent nodes overestimates both gain and level.
	if decAIGPacked(left) == decAIGPacked(right) {
		return left, true
	}
	if left.Node == right.Node && left.Complement != right.Complement {
		return manager.Constant(false), true
	}
	if left.Node == 0 {
		if left.Complement {
			return left, true
		}
		return right, true
	}
	if right.Node == 0 {
		if right.Complement {
			return right, true
		}
		return left, true
	}
	if decAIGPacked(left) > decAIGPacked(right) {
		left, right = right, left
	}
	index, ok := manager.unique[decAIGKey{left.Node, right.Node, left.Complement, right.Complement}]
	return DecAIGEdge{Node: index}, ok && index < len(manager.nodes) && manager.nodes[index].Alive
}

func (manager *DecAIG) and(left, right DecAIGEdge, strash bool) (DecAIGEdge, error) {
	if err := manager.check(); err != nil {
		return DecAIGEdge{}, err
	}
	if err := manager.validateEdge(left); err != nil {
		return DecAIGEdge{}, err
	}
	if err := manager.validateEdge(right); err != nil {
		return DecAIGEdge{}, err
	}
	if manager.stopped() {
		return DecAIGEdge{}, ErrCancelled
	}
	if !strash {
		if len(manager.nodes) >= manager.limits.MaxNodes {
			return DecAIGEdge{}, ErrLimit
		}
		level := 1 + max(manager.nodes[left.Node].Level, manager.nodes[right.Node].Level)
		index := len(manager.nodes)
		manager.nodes = append(manager.nodes, DecAIGNode{Left: left, Right: right, Variable: -2, Level: level, Alive: true})
		return DecAIGEdge{Node: index}, nil
	}
	if decAIGPacked(left) == decAIGPacked(right) {
		return left, nil
	}
	if left.Node == right.Node && left.Complement != right.Complement {
		return manager.Constant(false), nil
	}
	if left.Node == 0 {
		if left.Complement {
			return left, nil
		}
		return right, nil
	}
	if right.Node == 0 {
		if right.Complement {
			return right, nil
		}
		return left, nil
	}
	if decAIGPacked(left) > decAIGPacked(right) {
		left, right = right, left
	}
	key := decAIGKey{left.Node, right.Node, left.Complement, right.Complement}
	if strash {
		if index, ok := manager.unique[key]; ok && manager.nodes[index].Alive {
			return DecAIGEdge{Node: index}, nil
		}
	}
	if len(manager.nodes) >= manager.limits.MaxNodes {
		return DecAIGEdge{}, ErrLimit
	}
	level := 1 + max(manager.nodes[left.Node].Level, manager.nodes[right.Node].Level)
	index := len(manager.nodes)
	manager.nodes = append(manager.nodes, DecAIGNode{Left: left, Right: right, Variable: -2, Level: level, Alive: true})
	if strash {
		manager.unique[key] = index
	}
	return DecAIGEdge{Node: index}, nil
}

// BuildGraph implements Dec_GraphToNetwork, Dec_GraphToAig,
// Dec_GraphToNetworkAig, and Dec_GraphToNetworkIvy. When strash is false it
// implements Dec_GraphToNetworkNoStrash.
func (manager *DecAIG) BuildGraph(graph *DecoGraph, fanins []DecAIGEdge, strash bool) (DecAIGEdge, error) {
	if err := manager.check(); err != nil {
		return DecAIGEdge{}, err
	}
	work := manager.clone()
	result, err := work.buildGraph(graph, fanins, strash)
	if err != nil {
		return DecAIGEdge{}, err
	}
	manager.nodes = work.nodes
	manager.unique = work.unique
	return result, nil
}

func (manager *DecAIG) clone() *DecAIG {
	result := &DecAIG{
		nodes:     append([]DecAIGNode(nil), manager.nodes...),
		variables: append([]DecAIGEdge(nil), manager.variables...),
		unique:    make(map[decAIGKey]int, len(manager.unique)),
		limits:    manager.limits,
	}
	for key, value := range manager.unique {
		result.unique[key] = value
	}
	return result
}

func (manager *DecAIG) buildGraph(graph *DecoGraph, fanins []DecAIGEdge, strash bool) (DecAIGEdge, error) {
	if err := manager.check(); err != nil {
		return DecAIGEdge{}, err
	}
	if err := graph.check(); err != nil {
		return DecAIGEdge{}, err
	}
	if manager.stopped() {
		return DecAIGEdge{}, ErrCancelled
	}
	if len(fanins) != graph.leaves {
		return DecAIGEdge{}, ErrDimension
	}
	for _, edge := range fanins {
		if err := manager.validateEdge(edge); err != nil {
			return DecAIGEdge{}, err
		}
	}
	if graph.constant {
		return manager.Constant(!graph.root.Complement), nil
	}
	values := make([]DecAIGEdge, len(graph.nodes))
	copy(values, fanins)
	for index := graph.leaves; index < len(graph.nodes); index++ {
		node := graph.nodes[index]
		left, right := values[node.Edge0.Node], values[node.Edge1.Node]
		if node.Edge0.Complement {
			left = left.Not()
		}
		if node.Edge1.Complement {
			right = right.Not()
		}
		var err error
		values[index], err = manager.and(left, right, strash)
		if err != nil {
			return DecAIGEdge{}, err
		}
	}
	if graph.root.Node < 0 || graph.root.Node >= len(values) {
		return DecAIGEdge{}, ErrInvariant
	}
	result := values[graph.root.Node]
	if graph.root.Complement {
		result = result.Not()
	}
	return result, nil
}

// FactorSOP builds the HOP-style strashed AIG returned by
// Dec_GraphFactorSop and Dec_SopToAig.
func (manager *DecAIG) FactorSOP(sop SOPCover) (DecAIGEdge, error) {
	graph, err := FactorSOP(sop, manager.limits)
	if err != nil {
		return DecAIGEdge{}, err
	}
	defer graph.Close()
	if graph.leaves > len(manager.variables) {
		return DecAIGEdge{}, ErrDimension
	}
	return manager.BuildGraph(graph, manager.variables[:graph.leaves], true)
}

// CountGraph implements Dec_GraphToNetworkCount without mutating the AIG. It
// returns ErrLimit when either source limit would have returned -1.
func (manager *DecAIG) CountGraph(root DecAIGEdge, graph *DecoGraph, fanins []DecAIGEdge, nodeMax, levelMax int) (int, error) {
	if err := manager.check(); err != nil {
		return 0, err
	}
	if err := graph.check(); err != nil {
		return 0, err
	}
	if err := manager.validateEdge(root); err != nil {
		return 0, err
	}
	if nodeMax < 0 || levelMax < 0 || len(fanins) != graph.leaves {
		return 0, ErrDimension
	}
	if graph.constant || graph.root.Node < graph.leaves {
		return 0, nil
	}
	values := make([]DecAIGEdge, len(graph.nodes))
	present := make([]bool, len(graph.nodes))
	levels := make([]int, len(graph.nodes))
	copy(values, fanins)
	for index := range fanins {
		if err := manager.validateEdge(fanins[index]); err != nil {
			return 0, err
		}
		levels[index] = manager.nodes[fanins[index].Node].Level
		present[index] = true
	}
	counter := 0
	for index := graph.leaves; index < len(graph.nodes); index++ {
		if manager.stopped() {
			return 0, ErrCancelled
		}
		node := graph.nodes[index]
		left, right := values[node.Edge0.Node], values[node.Edge1.Node]
		if node.Edge0.Complement {
			left = left.Not()
		}
		if node.Edge1.Complement {
			right = right.Not()
		}
		found, exists := DecAIGEdge{}, false
		if present[node.Edge0.Node] && present[node.Edge1.Node] {
			found, exists = manager.lookupAND(left, right)
		}
		if exists && found.Node == root.Node {
			return 0, ErrInvariant
		}
		if !exists || manager.nodes[found.Node].Marked {
			counter++
			if counter > nodeMax {
				return 0, ErrLimit
			}
		}
		level := 1 + max(levels[node.Edge0.Node], levels[node.Edge1.Node])
		if exists {
			if found.Node == 0 {
				level = 0
			} else if found.Node == left.Node {
				level = manager.nodes[left.Node].Level
			} else if found.Node == right.Node {
				level = manager.nodes[right.Node].Level
			}
		}
		if level > levelMax {
			return 0, ErrLimit
		}
		levels[index] = level
		if exists {
			values[index] = found
		}
		present[index] = exists
	}
	return counter, nil
}

// UpdateGraph implements Dec_GraphUpdateNetwork for the concrete AIG and
// rewires every live fanout of root to the newly built graph result.
func (manager *DecAIG) UpdateGraph(root DecAIGEdge, graph *DecoGraph, fanins []DecAIGEdge) (DecAIGEdge, error) {
	if err := manager.check(); err != nil {
		return DecAIGEdge{}, err
	}
	if err := graph.check(); err != nil {
		return DecAIGEdge{}, err
	}
	if err := manager.validateEdge(root); err != nil || root.Node == 0 {
		if err != nil {
			return DecAIGEdge{}, err
		}
		return DecAIGEdge{}, ErrDimension
	}
	if len(fanins) != graph.leaves {
		return DecAIGEdge{}, ErrDimension
	}
	var dependsOn func(int, int, []bool) bool
	dependsOn = func(index, target int, seen []bool) bool {
		if index == target {
			return true
		}
		if index < 0 || index >= len(manager.nodes) || seen[index] || !manager.nodes[index].Alive {
			return false
		}
		seen[index] = true
		node := manager.nodes[index]
		if node.Variable >= -1 {
			return false
		}
		return dependsOn(node.Left.Node, target, seen) || dependsOn(node.Right.Node, target, seen)
	}
	for _, fanin := range fanins {
		if err := manager.validateEdge(fanin); err != nil {
			return DecAIGEdge{}, err
		}
		if dependsOn(fanin.Node, root.Node, make([]bool, len(manager.nodes))) {
			return DecAIGEdge{}, ErrInvariant
		}
	}
	work := manager.clone()
	replacement, err := work.buildGraph(graph, fanins, true)
	if err != nil {
		return DecAIGEdge{}, err
	}
	if replacement.Node == root.Node {
		return replacement, nil
	}
	for index := 1; index < len(work.nodes); index++ {
		if work.stopped() {
			return DecAIGEdge{}, ErrCancelled
		}
		if !work.nodes[index].Alive || index == root.Node || work.nodes[index].Variable >= 0 {
			continue
		}
		for side := 0; side < 2; side++ {
			edge := &work.nodes[index].Left
			if side == 1 {
				edge = &work.nodes[index].Right
			}
			if edge.Node == root.Node {
				edge.Node = replacement.Node
				edge.Complement = edge.Complement != root.Complement
				edge.Complement = edge.Complement != replacement.Complement
			}
		}
	}
	work.nodes[root.Node].Alive = false
	work.rebuildUnique()
	manager.nodes, manager.unique = work.nodes, work.unique
	return replacement, nil
}

func (manager *DecAIG) rebuildUnique() {
	manager.unique = make(map[decAIGKey]int)
	for index := 1; index < len(manager.nodes); index++ {
		node := &manager.nodes[index]
		if !node.Alive || node.Variable >= 0 {
			continue
		}
		if decAIGPacked(node.Left) > decAIGPacked(node.Right) {
			node.Left, node.Right = node.Right, node.Left
		}
		manager.unique[decAIGKey{node.Left.Node, node.Right.Node, node.Left.Complement, node.Right.Complement}] = index
	}
	states := make([]uint8, len(manager.nodes))
	var level func(int) (int, bool)
	level = func(index int) (int, bool) {
		if index < 0 || index >= len(manager.nodes) || !manager.nodes[index].Alive {
			return 0, false
		}
		if states[index] == 2 {
			return manager.nodes[index].Level, true
		}
		if states[index] == 1 {
			return 0, false
		}
		states[index] = 1
		node := &manager.nodes[index]
		if index == 0 || node.Variable >= 0 {
			node.Level = 0
		} else {
			left, ok := level(node.Left.Node)
			if !ok {
				return 0, false
			}
			right, ok := level(node.Right.Node)
			if !ok {
				return 0, false
			}
			node.Level = 1 + max(left, right)
		}
		states[index] = 2
		return node.Level, true
	}
	for index := range manager.nodes {
		if manager.nodes[index].Alive {
			_, _ = level(index)
		}
	}
}

func (manager *DecAIG) Evaluate(edge DecAIGEdge, inputs []bool) (bool, error) {
	if err := manager.check(); err != nil {
		return false, err
	}
	if err := manager.validateEdge(edge); err != nil {
		return false, err
	}
	if len(inputs) != len(manager.variables) {
		return false, ErrDimension
	}
	values := make([]bool, len(manager.nodes))
	states := make([]uint8, len(manager.nodes))
	values[0], states[0] = true, 2
	for variable, variableEdge := range manager.variables {
		values[variableEdge.Node], states[variableEdge.Node] = inputs[variable], 2
	}
	var read func(DecAIGEdge) (bool, error)
	read = func(input DecAIGEdge) (bool, error) {
		if err := manager.validateEdge(input); err != nil {
			return false, err
		}
		if manager.stopped() {
			return false, ErrCancelled
		}
		if states[input.Node] == 1 {
			return false, ErrInvariant
		}
		if states[input.Node] == 0 {
			states[input.Node] = 1
			node := manager.nodes[input.Node]
			left, err := read(node.Left)
			if err != nil {
				return false, err
			}
			right, err := read(node.Right)
			if err != nil {
				return false, err
			}
			values[input.Node] = left && right
			states[input.Node] = 2
		}
		value := values[input.Node]
		if input.Complement {
			value = !value
		}
		return value, nil
	}
	return read(edge)
}

func (manager *DecAIG) Nodes() ([]DecAIGNode, error) {
	if err := manager.check(); err != nil {
		return nil, err
	}
	return append([]DecAIGNode(nil), manager.nodes...), nil
}

// MarkNode supplies the traversal mark consumed by CountGraph when estimating
// how many existing nodes belong to the cone being replaced.
func (manager *DecAIG) MarkNode(index int, marked bool) error {
	if err := manager.check(); err != nil {
		return err
	}
	if index < 0 || index >= len(manager.nodes) || !manager.nodes[index].Alive {
		return ErrDimension
	}
	manager.nodes[index].Marked = marked
	return nil
}

func (manager *DecAIG) Close() {
	if manager == nil || manager.closed {
		return
	}
	manager.nodes = nil
	manager.variables = nil
	manager.unique = nil
	manager.closed = true
}
