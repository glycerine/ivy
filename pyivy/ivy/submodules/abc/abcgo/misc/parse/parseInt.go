package parse

// Precedence follows ABC's parser: negation, conjunction, xor, disjunction,
// equivalence, forward implication, backward implication. Equal-precedence
// binary operators associate to the left.
const PARSE_OPER_NEG = 10
const PARSE_OPER_AND = 9
const PARSE_OPER_XOR = 8
const PARSE_OPER_OR = 7
const PARSE_OPER_EQU = 6
const PARSE_OPER_FLR = 5
const PARSE_OPER_FLL = 4
const PARSE_OPER_MARK = 1
