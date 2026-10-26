package abc

// Boolean formula, expression-AIG, and SOP core from src/map/mio/exp.h,
// mioForm.c, mioFunc.c, mioParse.c, and mioSop.c.

import (
	"fmt"
	"slices"
	"strings"
)

const (
	mioExprConst0 MioExprLit = -1
	mioExprConst1 MioExprLit = -2
)

// MioFormulaLimits bounds parser and derived-representation work. Zero selects
// a conservative default. Stop is sampled during parsing, truth evaluation,
// and cover construction.
type MioFormulaLimits struct {
	MaxVariables int
	MaxNodes     int
	MaxTokens    int
	MaxDepth     int
	MaxWords     int
	MaxCubes     int
	Stop         func() bool
}

func normalizeMioFormulaLimits(limits MioFormulaLimits) (MioFormulaLimits, error) {
	if limits.MaxVariables < 0 || limits.MaxNodes < 0 || limits.MaxTokens < 0 || limits.MaxDepth < 0 || limits.MaxWords < 0 || limits.MaxCubes < 0 {
		return MioFormulaLimits{}, ErrDimension
	}
	if limits.MaxVariables == 0 {
		limits.MaxVariables = 16
	}
	if limits.MaxNodes == 0 {
		limits.MaxNodes = 1 << 16
	}
	if limits.MaxTokens == 0 {
		limits.MaxTokens = 1 << 16
	}
	if limits.MaxDepth == 0 {
		limits.MaxDepth = 1024
	}
	if limits.MaxWords == 0 {
		limits.MaxWords = 1 << 20
	}
	if limits.MaxCubes == 0 {
		limits.MaxCubes = 1 << 16
	}
	return limits, nil
}

// MioExprLit uses the ABC literal convention for variables and AND nodes.
// Constants retain exp.h's two negative sentinels.
type MioExprLit int

func MioExprConstant(value bool) MioExprLit {
	if value {
		return mioExprConst1
	}
	return mioExprConst0
}

func (literal MioExprLit) Complement() MioExprLit {
	if literal == mioExprConst0 {
		return mioExprConst1
	}
	if literal == mioExprConst1 {
		return mioExprConst0
	}
	return literal ^ 1
}

func (literal MioExprLit) IsConstant() bool {
	return literal == mioExprConst0 || literal == mioExprConst1
}
func (literal MioExprLit) IsComplemented() bool {
	return literal >= 0 && literal&1 != 0
}
func (literal MioExprLit) Index() int {
	if literal < 0 {
		return -1
	}
	return int(literal) >> 1
}

type MioExprNode struct {
	First, Second MioExprLit
}

// MioExpression is an owned topological AND/inverter graph. Names retain the
// caller's variable order; returned snapshots cannot mutate it.
type MioExpression struct {
	variableNames []string
	nodes         []MioExprNode
	root          MioExprLit
}

func (expression *MioExpression) VariableNames() []string {
	if expression == nil {
		return nil
	}
	return slices.Clone(expression.variableNames)
}
func (expression *MioExpression) Nodes() []MioExprNode {
	if expression == nil {
		return nil
	}
	return slices.Clone(expression.nodes)
}
func (expression *MioExpression) Root() MioExprLit {
	if expression == nil {
		return mioExprConst0
	}
	return expression.root
}
func (expression *MioExpression) NodeCount() int {
	if expression == nil {
		return 0
	}
	return len(expression.nodes)
}

func (expression *MioExpression) Check() error {
	if expression == nil || len(expression.variableNames) > 16 {
		return ErrDimension
	}
	seen := make(map[string]struct{}, len(expression.variableNames))
	for _, name := range expression.variableNames {
		if !mioValidVariableName(name) {
			return ErrDimension
		}
		if _, found := seen[name]; found {
			return ErrDimension
		}
		seen[name] = struct{}{}
	}
	valid := func(literal MioExprLit, available int) bool {
		return literal.IsConstant() || literal >= 0 && literal.Index() < available
	}
	for index, node := range expression.nodes {
		if !valid(node.First, len(expression.variableNames)+index) || !valid(node.Second, len(expression.variableNames)+index) {
			return ErrInvalidLiteral
		}
	}
	if !valid(expression.root, len(expression.variableNames)+len(expression.nodes)) {
		return ErrInvalidLiteral
	}
	return nil
}

func mioValidVariableName(name string) bool {
	if name == "" || name == "0" || name == "1" {
		return false
	}
	for index := 0; index < len(name); index++ {
		if name[index] == 0 || mioFormulaDelimiter(name[index]) {
			return false
		}
	}
	return true
}

func mioFormulaDelimiter(value byte) bool {
	switch value {
	case ' ', '\t', '\r', '\n', '(', ')', '!', '\'', '*', '&', '^', '+', '|':
		return true
	}
	return false
}

type mioFormulaParser struct {
	text      string
	position  int
	names     []string
	nameIndex map[string]int
	nodes     []MioExprNode
	limits    MioFormulaLimits
	tokens    int
	depth     int
}

func (parser *mioFormulaParser) charge() error {
	if parser.limits.Stop != nil && parser.limits.Stop() {
		return ErrCancelled
	}
	if parser.tokens == parser.limits.MaxTokens {
		return ErrLimit
	}
	parser.tokens++
	return nil
}

func (parser *mioFormulaParser) skipSpace() {
	for parser.position < len(parser.text) {
		switch parser.text[parser.position] {
		case ' ', '\t', '\r', '\n':
			parser.position++
		default:
			return
		}
	}
}

func (parser *mioFormulaParser) addAnd(first, second MioExprLit) (MioExprLit, error) {
	if parser.limits.Stop != nil && parser.limits.Stop() {
		return 0, ErrCancelled
	}
	if len(parser.nodes) == parser.limits.MaxNodes {
		return 0, ErrLimit
	}
	result := MioExprLit(2 * (len(parser.names) + len(parser.nodes)))
	parser.nodes = append(parser.nodes, MioExprNode{First: first, Second: second})
	return result, nil
}

func (parser *mioFormulaParser) addBinary(operator byte, first, second MioExprLit) (MioExprLit, error) {
	switch operator {
	case '*', '&':
		return parser.addAnd(first, second)
	case '+', '|':
		and, err := parser.addAnd(first.Complement(), second.Complement())
		return and.Complement(), err
	case '^':
		left, err := parser.addAnd(first.Complement(), second)
		if err != nil {
			return 0, err
		}
		right, err := parser.addAnd(first, second.Complement())
		if err != nil {
			return 0, err
		}
		bothFalse, err := parser.addAnd(right.Complement(), left.Complement())
		return bothFalse.Complement(), err
	}
	return 0, ErrDimension
}

func (parser *mioFormulaParser) parseOr() (MioExprLit, error) {
	left, err := parser.parseXor()
	if err != nil {
		return 0, err
	}
	for {
		parser.skipSpace()
		if parser.position == len(parser.text) || parser.text[parser.position] != '+' && parser.text[parser.position] != '|' {
			return left, nil
		}
		operator := parser.text[parser.position]
		parser.position++
		if err = parser.charge(); err != nil {
			return 0, err
		}
		right, nextErr := parser.parseXor()
		if nextErr != nil {
			return 0, nextErr
		}
		left, err = parser.addBinary(operator, left, right)
		if err != nil {
			return 0, err
		}
	}
}

func (parser *mioFormulaParser) parseXor() (MioExprLit, error) {
	left, err := parser.parseAnd()
	if err != nil {
		return 0, err
	}
	for {
		parser.skipSpace()
		if parser.position == len(parser.text) || parser.text[parser.position] != '^' {
			return left, nil
		}
		parser.position++
		if err = parser.charge(); err != nil {
			return 0, err
		}
		right, nextErr := parser.parseAnd()
		if nextErr != nil {
			return 0, nextErr
		}
		left, err = parser.addBinary('^', left, right)
		if err != nil {
			return 0, err
		}
	}
}

func (parser *mioFormulaParser) startsUnary() bool {
	parser.skipSpace()
	if parser.position == len(parser.text) {
		return false
	}
	value := parser.text[parser.position]
	return value == '!' || value == '(' || value == '0' || value == '1' || !mioFormulaDelimiter(value)
}

func (parser *mioFormulaParser) parseAnd() (MioExprLit, error) {
	left, err := parser.parseUnary()
	if err != nil {
		return 0, err
	}
	for {
		parser.skipSpace()
		explicit := parser.position < len(parser.text) && (parser.text[parser.position] == '*' || parser.text[parser.position] == '&')
		if explicit {
			parser.position++
			if err = parser.charge(); err != nil {
				return 0, err
			}
		} else if !parser.startsUnary() {
			return left, nil
		}
		right, nextErr := parser.parseUnary()
		if nextErr != nil {
			return 0, nextErr
		}
		left, err = parser.addBinary('&', left, right)
		if err != nil {
			return 0, err
		}
	}
}

func (parser *mioFormulaParser) parseUnary() (MioExprLit, error) {
	parser.skipSpace()
	complement := false
	for parser.position < len(parser.text) && parser.text[parser.position] == '!' {
		if err := parser.charge(); err != nil {
			return 0, err
		}
		complement = !complement
		parser.position++
		parser.skipSpace()
	}
	literal, err := parser.parsePrimary()
	if err != nil {
		return 0, err
	}
	for {
		parser.skipSpace()
		if parser.position == len(parser.text) || parser.text[parser.position] != '\'' {
			break
		}
		if err = parser.charge(); err != nil {
			return 0, err
		}
		complement = !complement
		parser.position++
	}
	if complement {
		literal = literal.Complement()
	}
	return literal, nil
}

func (parser *mioFormulaParser) parsePrimary() (MioExprLit, error) {
	parser.skipSpace()
	if parser.position == len(parser.text) {
		return 0, ErrDimension
	}
	if err := parser.charge(); err != nil {
		return 0, err
	}
	value := parser.text[parser.position]
	if value == '(' {
		parser.position++
		parser.depth++
		if parser.depth > parser.limits.MaxDepth {
			return 0, ErrLimit
		}
		literal, err := parser.parseOr()
		parser.depth--
		if err != nil {
			return 0, err
		}
		parser.skipSpace()
		if parser.position == len(parser.text) || parser.text[parser.position] != ')' {
			return 0, ErrDimension
		}
		parser.position++
		return literal, nil
	}
	if value == '0' || value == '1' {
		parser.position++
		return MioExprConstant(value == '1'), nil
	}
	start := parser.position
	for parser.position < len(parser.text) && !mioFormulaDelimiter(parser.text[parser.position]) {
		if parser.text[parser.position] == 0 {
			return 0, ErrDimension
		}
		parser.position++
	}
	if start == parser.position {
		return 0, ErrDimension
	}
	index, found := parser.nameIndex[parser.text[start:parser.position]]
	if !found {
		return 0, ErrDimension
	}
	return MioExprLit(2 * index), nil
}

// ParseMioFormula parses ABC/MIO Boolean formula syntax. AND binds tighter
// than XOR, which binds tighter than OR. Juxtaposition is implicit AND; ! and
// apostrophe are prefix and postfix complement operators.
func ParseMioFormula(formula string, variableNames []string, limits MioFormulaLimits) (*MioExpression, error) {
	limits, err := normalizeMioFormulaLimits(limits)
	if err != nil || formula == "" {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if len(variableNames) > limits.MaxVariables {
		return nil, ErrLimit
	}
	parser := mioFormulaParser{text: formula, names: slices.Clone(variableNames), nameIndex: make(map[string]int, len(variableNames)), limits: limits}
	for index, name := range parser.names {
		if !mioValidVariableName(name) {
			return nil, ErrDimension
		}
		if _, found := parser.nameIndex[name]; found {
			return nil, ErrDimension
		}
		parser.nameIndex[name] = index
	}
	root, err := parser.parseOr()
	if err != nil {
		return nil, err
	}
	parser.skipSpace()
	if parser.position != len(parser.text) {
		return nil, ErrDimension
	}
	expression := &MioExpression{variableNames: parser.names, nodes: parser.nodes, root: root}
	if err = expression.Check(); err != nil {
		return nil, err
	}
	return expression, nil
}

// MioCollectFormulaNames returns distinct names in first-encounter order.
// Boolean constants are not names. Invalid syntax bytes are rejected.
func MioCollectFormulaNames(formula string, maximum int) ([]string, error) {
	if maximum < 0 || formula == "" {
		return nil, ErrDimension
	}
	if maximum == 0 {
		maximum = 1 << 16
	}
	seen := make(map[string]struct{})
	result := make([]string, 0)
	for position := 0; position < len(formula); {
		value := formula[position]
		if value == 0 {
			return nil, ErrDimension
		}
		if value == '0' || value == '1' {
			// The native gate-name collector sees a delimiter-free token
			// beginning with a digit as one pin name, while its parser consumes
			// the digit as a constant. Native gate compilation therefore rejects
			// this ambiguity. Do the same instead of compiling another function.
			if position+1 < len(formula) && !mioFormulaDelimiter(formula[position+1]) {
				return nil, ErrDimension
			}
			position++
			continue
		}
		if mioFormulaDelimiter(value) {
			position++
			continue
		}
		start := position
		for position < len(formula) && !mioFormulaDelimiter(formula[position]) {
			if formula[position] == 0 {
				return nil, ErrDimension
			}
			position++
		}
		name := formula[start:position]
		if _, found := seen[name]; found {
			continue
		}
		if len(result) == maximum {
			return nil, ErrLimit
		}
		seen[name] = struct{}{}
		result = append(result, name)
	}
	return result, nil
}

func (expression *MioExpression) literalWords(literal MioExprLit, values [][]uint64, wordCount int) ([]uint64, error) {
	if literal == mioExprConst0 || literal == mioExprConst1 {
		result := make([]uint64, wordCount)
		if literal == mioExprConst1 {
			for index := range result {
				result[index] = ^uint64(0)
			}
		}
		return result, nil
	}
	index := literal.Index()
	if index < 0 || index >= len(values) {
		return nil, ErrInvalidLiteral
	}
	result := slices.Clone(values[index])
	if literal.IsComplemented() {
		for index := range result {
			result[index] = ^result[index]
		}
	}
	return result, nil
}

// Truth derives the expression's ABC-packed truth table.
func (expression *MioExpression) Truth(limits MioFormulaLimits) (*TruthTable, error) {
	if err := expression.Check(); err != nil {
		return nil, err
	}
	limits, err := normalizeMioFormulaLimits(limits)
	if err != nil {
		return nil, err
	}
	count, err := TruthTableWordCount(len(expression.variableNames))
	if err != nil || count > limits.MaxWords {
		if err != nil {
			return nil, err
		}
		return nil, ErrLimit
	}
	values := make([][]uint64, 0, len(expression.variableNames)+len(expression.nodes))
	for variable := range expression.variableNames {
		if limits.Stop != nil && limits.Stop() {
			return nil, ErrCancelled
		}
		table, tableErr := NewVariableTruthTable(len(expression.variableNames), variable, limits.MaxWords)
		if tableErr != nil {
			return nil, tableErr
		}
		values = append(values, table.Words())
	}
	for _, node := range expression.nodes {
		if limits.Stop != nil && limits.Stop() {
			return nil, ErrCancelled
		}
		first, firstErr := expression.literalWords(node.First, values, count)
		if firstErr != nil {
			return nil, firstErr
		}
		second, secondErr := expression.literalWords(node.Second, values, count)
		if secondErr != nil {
			return nil, secondErr
		}
		for word := range first {
			first[word] &= second[word]
		}
		values = append(values, first)
	}
	words, err := expression.literalWords(expression.root, values, count)
	if err != nil {
		return nil, err
	}
	return TruthTableFromWords(len(expression.variableNames), words, limits.MaxWords)
}

func (expression *MioExpression) Truth6() (uint64, error) {
	if expression == nil || len(expression.variableNames) > 6 {
		return 0, ErrDimension
	}
	table, err := expression.Truth(MioFormulaLimits{MaxWords: 1})
	if err != nil {
		return 0, err
	}
	return table.words[0], nil
}

// Text returns the source's node-list notation without writing global output.
func (expression *MioExpression) Text() (string, error) {
	if err := expression.Check(); err != nil {
		return "", err
	}
	literalText := func(literal MioExprLit) string {
		if literal == mioExprConst0 {
			return "Const0"
		}
		if literal == mioExprConst1 {
			return "Const1"
		}
		prefix := " "
		if literal.IsComplemented() {
			prefix = "!"
		}
		index := literal.Index()
		if index < len(expression.variableNames) {
			return prefix + expression.variableNames[index]
		}
		return fmt.Sprintf("%s%d", prefix, index)
	}
	var text strings.Builder
	for index, node := range expression.nodes {
		fmt.Fprintf(&text, "%2d = %s & %s\n", len(expression.variableNames)+index, literalText(node.First), literalText(node.Second))
	}
	fmt.Fprintf(&text, " F = %s\n", literalText(expression.root))
	return text.String(), nil
}

// MioSOPCube uses two bits per variable: 1 requires zero, 2 requires one.
type MioSOPCube uint32

func MioSOPLiteral(variable int, value bool) (MioSOPCube, error) {
	if variable < 0 || variable >= 16 {
		return 0, ErrDimension
	}
	shift := uint(2 * variable)
	if value {
		return MioSOPCube(2 << shift), nil
	}
	return MioSOPCube(1 << shift), nil
}

func (cube MioSOPCube) Empty() bool { return uint32(cube)&(uint32(cube)>>1)&0x55555555 != 0 }
func (cube MioSOPCube) Contains(other MioSOPCube) bool {
	return cube|other == other
}
func (cube MioSOPCube) Matches(assignment uint64, variables int) (bool, error) {
	if variables < 0 || variables > 16 || assignment >= uint64(1)<<uint(variables) && variables < 64 {
		return false, ErrDimension
	}
	for variable := 0; variable < variables; variable++ {
		pair := uint32(cube) >> uint(2*variable) & 3
		if pair == 3 {
			return false, nil
		}
		value := assignment>>uint(variable)&1 != 0
		if pair == 1 && value || pair == 2 && !value {
			return false, nil
		}
	}
	return true, nil
}

// MioSOP represents either an on-set cover (OnValue true) or an off-set cover.
// A matching cube yields OnValue; no matching cube yields its complement.
type MioSOP struct {
	Variables int
	Cubes     []MioSOPCube
	OnValue   bool
}

func (cover MioSOP) Check() error {
	if cover.Variables < 0 || cover.Variables > 16 {
		return ErrDimension
	}
	validMask := uint32(0)
	if cover.Variables == 16 {
		validMask = ^uint32(0)
	} else if cover.Variables > 0 {
		validMask = 1<<uint(2*cover.Variables) - 1
	}
	for _, cube := range cover.Cubes {
		if uint32(cube)&^validMask != 0 || cube.Empty() {
			return ErrDimension
		}
	}
	return nil
}

func (cover MioSOP) Evaluate(assignment uint64) (bool, error) {
	if err := cover.Check(); err != nil {
		return false, err
	}
	for _, cube := range cover.Cubes {
		match, err := cube.Matches(assignment, cover.Variables)
		if err != nil {
			return false, err
		}
		if match {
			return cover.OnValue, nil
		}
	}
	return !cover.OnValue, nil
}

func (cover MioSOP) Format() (string, error) {
	if err := cover.Check(); err != nil {
		return "", err
	}
	if len(cover.Cubes) == 0 {
		if cover.OnValue {
			return " 0\n", nil
		}
		return " 1\n", nil
	}
	if len(cover.Cubes) == 1 && cover.Cubes[0] == 0 {
		if cover.OnValue {
			return " 1\n", nil
		}
		return " 0\n", nil
	}
	var result strings.Builder
	for _, cube := range cover.Cubes {
		for variable := 0; variable < cover.Variables; variable++ {
			switch uint32(cube) >> uint(2*variable) & 3 {
			case 1:
				result.WriteByte('0')
			case 2:
				result.WriteByte('1')
			default:
				result.WriteByte('-')
			}
		}
		if cover.OnValue {
			result.WriteString(" 1\n")
		} else {
			result.WriteString(" 0\n")
		}
	}
	return result.String(), nil
}

func mioSOPPush(cubes []MioSOPCube, cube MioSOPCube, maximum int) ([]MioSOPCube, error) {
	result := cubes[:0]
	for _, existing := range cubes {
		if existing.Contains(cube) {
			return cubes, nil
		}
		if cube.Contains(existing) {
			continue
		}
		result = append(result, existing)
	}
	if len(result) == maximum {
		return nil, ErrLimit
	}
	return append(result, cube), nil
}

func mioSOPOr(first, second []MioSOPCube, limits MioFormulaLimits) ([]MioSOPCube, error) {
	result := slices.Clone(first)
	for work, cube := range second {
		if limits.Stop != nil && limits.Stop() {
			return nil, ErrCancelled
		}
		if work == limits.MaxTokens {
			return nil, ErrLimit
		}
		var err error
		result, err = mioSOPPush(result, cube, limits.MaxCubes)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func mioSOPAnd(first, second []MioSOPCube, limits MioFormulaLimits) ([]MioSOPCube, error) {
	result := make([]MioSOPCube, 0)
	work := 0
	for _, left := range first {
		for _, right := range second {
			if limits.Stop != nil && limits.Stop() {
				return nil, ErrCancelled
			}
			if work == limits.MaxTokens {
				return nil, ErrLimit
			}
			work++
			cube := left | right
			if cube.Empty() {
				continue
			}
			var err error
			result, err = mioSOPPush(result, cube, limits.MaxCubes)
			if err != nil {
				return nil, err
			}
		}
	}
	return result, nil
}

func mioSOPSelect(literal MioExprLit, positive, negative [][]MioSOPCube) ([]MioSOPCube, error) {
	if literal.IsConstant() {
		if literal == mioExprConst1 {
			return []MioSOPCube{0}, nil
		}
		return []MioSOPCube{}, nil
	}
	index := literal.Index()
	if index < 0 || index >= len(positive) {
		return nil, ErrInvalidLiteral
	}
	if literal.IsComplemented() {
		return negative[index], nil
	}
	return positive[index], nil
}

func (expression *MioExpression) SOP(limits MioFormulaLimits) (MioSOP, error) {
	if err := expression.Check(); err != nil {
		return MioSOP{}, err
	}
	limits, err := normalizeMioFormulaLimits(limits)
	if err != nil {
		return MioSOP{}, err
	}
	if len(expression.variableNames) > 16 {
		return MioSOP{}, ErrDimension
	}
	positive := make([][]MioSOPCube, 0, len(expression.variableNames)+len(expression.nodes))
	negative := make([][]MioSOPCube, 0, cap(positive))
	for variable := range expression.variableNames {
		zero, _ := MioSOPLiteral(variable, false)
		one, _ := MioSOPLiteral(variable, true)
		negative = append(negative, []MioSOPCube{zero})
		positive = append(positive, []MioSOPCube{one})
	}
	for _, node := range expression.nodes {
		firstPositive, firstErr := mioSOPSelect(node.First, positive, negative)
		if firstErr != nil {
			return MioSOP{}, firstErr
		}
		secondPositive, secondErr := mioSOPSelect(node.Second, positive, negative)
		if secondErr != nil {
			return MioSOP{}, secondErr
		}
		firstNegative, _ := mioSOPSelect(node.First.Complement(), positive, negative)
		secondNegative, _ := mioSOPSelect(node.Second.Complement(), positive, negative)
		on, combineErr := mioSOPAnd(firstPositive, secondPositive, limits)
		if combineErr != nil {
			return MioSOP{}, combineErr
		}
		off, combineErr := mioSOPOr(firstNegative, secondNegative, limits)
		if combineErr != nil {
			return MioSOP{}, combineErr
		}
		positive = append(positive, on)
		negative = append(negative, off)
	}
	on, err := mioSOPSelect(expression.root, positive, negative)
	if err != nil {
		return MioSOP{}, err
	}
	off, err := mioSOPSelect(expression.root.Complement(), positive, negative)
	if err != nil {
		return MioSOP{}, err
	}
	result := MioSOP{Variables: len(expression.variableNames), Cubes: slices.Clone(on), OnValue: true}
	if len(off) < len(on) {
		result.Cubes = slices.Clone(off)
		result.OnValue = false
	}
	return result, result.Check()
}

type MioFormulaPin struct {
	Name     string
	MaxDelay float64
}

type MioFormulaGate struct {
	Name       string
	Formula    string
	Pins       []MioFormulaPin
	Expression *MioExpression
	SOP        MioSOP
	Truth      *TruthTable
	MaxDelay   float64
}

// CompileMioFormulaGate atomically derives the owned formula artifacts. A
// single pin named "*" expands to names in first formula encounter order.
func CompileMioFormulaGate(gate MioFormulaGate, limits MioFormulaLimits) (MioFormulaGate, error) {
	result := MioFormulaGate{Name: gate.Name, Formula: gate.Formula, Pins: slices.Clone(gate.Pins)}
	names, err := MioCollectFormulaNames(gate.Formula, 16)
	if err != nil {
		return MioFormulaGate{}, err
	}
	if len(result.Pins) == 0 {
		if gate.Formula != "0" && gate.Formula != "1" {
			return MioFormulaGate{}, ErrDimension
		}
		names = nil
	} else if len(result.Pins) == 1 && result.Pins[0].Name == "*" {
		if len(names) == 0 {
			return MioFormulaGate{}, ErrDimension
		}
		generic := result.Pins[0]
		result.Pins = make([]MioFormulaPin, len(names))
		for index, name := range names {
			result.Pins[index] = generic
			result.Pins[index].Name = name
		}
	} else {
		if len(names) != len(result.Pins) {
			return MioFormulaGate{}, ErrDimension
		}
		found := make(map[string]struct{}, len(names))
		for _, name := range names {
			found[name] = struct{}{}
		}
		names = names[:0]
		for _, pin := range result.Pins {
			if _, ok := found[pin.Name]; !ok || !mioValidVariableName(pin.Name) {
				return MioFormulaGate{}, ErrDimension
			}
			names = append(names, pin.Name)
			delete(found, pin.Name)
			if pin.MaxDelay > result.MaxDelay {
				result.MaxDelay = pin.MaxDelay
			}
		}
		if len(found) != 0 {
			return MioFormulaGate{}, ErrDimension
		}
	}
	for _, pin := range result.Pins {
		if pin.MaxDelay > result.MaxDelay {
			result.MaxDelay = pin.MaxDelay
		}
	}
	result.Expression, err = ParseMioFormula(gate.Formula, names, limits)
	if err != nil {
		return MioFormulaGate{}, err
	}
	result.SOP, err = result.Expression.SOP(limits)
	if err != nil {
		return MioFormulaGate{}, err
	}
	result.Truth, err = result.Expression.Truth(limits)
	if err != nil {
		return MioFormulaGate{}, err
	}
	return result, nil
}

type MioFormulaLibrary struct {
	Gates        []MioFormulaGate
	ConstantZero int
	ConstantOne  int
}

// CompileMioFormulaLibrary compiles every gate into new owned state. The input
// is unchanged on success and failure.
func CompileMioFormulaLibrary(gates []MioFormulaGate, limits MioFormulaLimits) (MioFormulaLibrary, error) {
	result := MioFormulaLibrary{Gates: make([]MioFormulaGate, len(gates)), ConstantZero: -1, ConstantOne: -1}
	for index, gate := range gates {
		if limits.Stop != nil && limits.Stop() {
			return MioFormulaLibrary{}, ErrCancelled
		}
		compiled, err := CompileMioFormulaGate(gate, limits)
		if err != nil {
			return MioFormulaLibrary{}, err
		}
		result.Gates[index] = compiled
		if len(compiled.Pins) == 0 && compiled.Formula == "0" {
			result.ConstantZero = index
		}
		if len(compiled.Pins) == 0 && compiled.Formula == "1" {
			result.ConstantOne = index
		}
	}
	return result, nil
}
