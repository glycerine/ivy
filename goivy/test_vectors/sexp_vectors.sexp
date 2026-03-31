;; sexp_vectors.sexp — Golden test vectors for Go/Python S-expression agreement.
;; Each entry specifies the exact sexp output for a constructed node.
;; Both Go and Python tests validate their nodes produce the expected string.

;; ============================================================
;; Sort types
;; ============================================================

(vector id:uninterp_sort_basic type:UninterpretedSort
  expected:"(UninterpretedSort name:S)")

(vector id:boolean_sort type:BooleanSort
  expected:"(BooleanSort)")

(vector id:func_sort_binary type:FunctionSort
  expected:"(FunctionSort sorts:[(UninterpretedSort name:S) (UninterpretedSort name:S) (BooleanSort)])")

(vector id:func_sort_unary type:FunctionSort
  expected:"(FunctionSort sorts:[(UninterpretedSort name:S) (BooleanSort)])")

(vector id:enum_sort type:EnumeratedSort
  expected:"(EnumeratedSort name:Color ext:[red,green,blue])")

(vector id:enum_sort_single type:EnumeratedSort
  expected:"(EnumeratedSort name:X ext:[a])")

(vector id:enum_sort_empty type:EnumeratedSort
  expected:"(EnumeratedSort name:E ext:[])")

(vector id:range_sort type:RangeSort
  expected:"(RangeSort name:idx lb:0 ub:10)")

(vector id:top_sort_default type:TopSort
  expected:"(TopSort name:TopSort)")

(vector id:top_sort_named type:TopSort
  expected:"(TopSort name:Alpha)")

;; ============================================================
;; Term types
;; ============================================================

(vector id:variable type:Variable
  expected:"(Variable name:X sort:(UninterpretedSort name:S))")

(vector id:symbol_constant type:Symbol
  expected:"(Symbol name:c sort:(UninterpretedSort name:S))")

(vector id:symbol_unary_func type:Symbol
  expected:"(Symbol name:f sort:(FunctionSort sorts:[(UninterpretedSort name:S) (BooleanSort)]))")

(vector id:apply_binary type:Apply
  expected:"(Apply func:(Symbol name:leq sort:(FunctionSort sorts:[(UninterpretedSort name:S) (UninterpretedSort name:S) (BooleanSort)])) terms:[(Variable name:X sort:(UninterpretedSort name:S)) (Variable name:Y sort:(UninterpretedSort name:S))])")

(vector id:apply_nullary type:Apply
  expected:"(Apply func:(Symbol name:f sort:(FunctionSort sorts:[(BooleanSort)])) terms:[])")

(vector id:apply_nested type:Apply
  expected:"(Apply func:(Symbol name:g sort:(FunctionSort sorts:[(BooleanSort) (BooleanSort)])) terms:[(Apply func:(Symbol name:f sort:(FunctionSort sorts:[(UninterpretedSort name:S) (BooleanSort)])) terms:[(Variable name:X sort:(UninterpretedSort name:S))])])")

;; ============================================================
;; Formula types
;; ============================================================

(vector id:eq type:Eq
  expected:"(Eq t1:(Variable name:X sort:(UninterpretedSort name:S)) t2:(Variable name:Y sort:(UninterpretedSort name:S)))")

(vector id:not type:Not
  expected:"(Not body:(Eq t1:(Variable name:X sort:(UninterpretedSort name:S)) t2:(Variable name:Y sort:(UninterpretedSort name:S))))")

(vector id:and_two type:And
  expected:"(And terms:[(Eq t1:(Variable name:X sort:(UninterpretedSort name:S)) t2:(Variable name:Y sort:(UninterpretedSort name:S))) (Eq t1:(Variable name:X sort:(UninterpretedSort name:S)) t2:(Variable name:Y sort:(UninterpretedSort name:S)))])")

(vector id:and_empty type:And
  expected:"(And terms:[])")

(vector id:or_two type:Or
  expected:"(Or terms:[(Eq t1:(Variable name:X sort:(UninterpretedSort name:S)) t2:(Variable name:Y sort:(UninterpretedSort name:S))) (Eq t1:(Variable name:X sort:(UninterpretedSort name:S)) t2:(Variable name:Y sort:(UninterpretedSort name:S)))])")

(vector id:or_empty type:Or
  expected:"(Or terms:[])")

(vector id:implies type:Implies
  expected:"(Implies t1:(Eq t1:(Variable name:X sort:(UninterpretedSort name:S)) t2:(Variable name:Y sort:(UninterpretedSort name:S))) t2:(Eq t1:(Variable name:X sort:(UninterpretedSort name:S)) t2:(Variable name:Y sort:(UninterpretedSort name:S))))")

(vector id:iff type:Iff
  expected:"(Iff t1:(Eq t1:(Variable name:X sort:(UninterpretedSort name:S)) t2:(Variable name:Y sort:(UninterpretedSort name:S))) t2:(Eq t1:(Variable name:X sort:(UninterpretedSort name:S)) t2:(Variable name:Y sort:(UninterpretedSort name:S))))")

(vector id:ite type:Ite
  expected:"(Ite cond:(Eq t1:(Variable name:X sort:(UninterpretedSort name:S)) t2:(Variable name:Y sort:(UninterpretedSort name:S))) then:(Variable name:X sort:(UninterpretedSort name:S)) else:(Variable name:Y sort:(UninterpretedSort name:S)))")

(vector id:globally_nil_env type:Globally
  expected:"(Globally environ:nil body:(Eq t1:(Variable name:X sort:(UninterpretedSort name:S)) t2:(Variable name:Y sort:(UninterpretedSort name:S))))")

(vector id:globally_with_env type:Globally
  expected:"(Globally environ:env1 body:(Eq t1:(Variable name:X sort:(UninterpretedSort name:S)) t2:(Variable name:Y sort:(UninterpretedSort name:S))))")

(vector id:eventually_nil_env type:Eventually
  expected:"(Eventually environ:nil body:(Eq t1:(Variable name:X sort:(UninterpretedSort name:S)) t2:(Variable name:Y sort:(UninterpretedSort name:S))))")

(vector id:eventually_with_env type:Eventually
  expected:"(Eventually environ:env1 body:(Eq t1:(Variable name:X sort:(UninterpretedSort name:S)) t2:(Variable name:Y sort:(UninterpretedSort name:S))))")

(vector id:when_operator type:WhenOperator
  expected:"(WhenOperator name:when t1:(Variable name:X sort:(UninterpretedSort name:S)) t2:(Eq t1:(Variable name:X sort:(UninterpretedSort name:S)) t2:(Variable name:Y sort:(UninterpretedSort name:S))))")

(vector id:cond type:Cond
  expected:"(Cond t1:(Variable name:X sort:(UninterpretedSort name:S)) t2:(Variable name:Y sort:(UninterpretedSort name:S)))")

;; ============================================================
;; Quantifier types
;; ============================================================

(vector id:forall_single type:ForAll
  expected:"(ForAll vars:[(Variable name:X sort:(UninterpretedSort name:S))] body:(Eq t1:(Variable name:X sort:(UninterpretedSort name:S)) t2:(Variable name:Y sort:(UninterpretedSort name:S))))")

(vector id:forall_multi_sorted type:ForAll
  expected:"(ForAll vars:[(Variable name:A sort:(UninterpretedSort name:S)) (Variable name:Z sort:(UninterpretedSort name:S))] body:(Eq t1:(Variable name:A sort:(UninterpretedSort name:S)) t2:(Variable name:Z sort:(UninterpretedSort name:S))))")

(vector id:exists type:Exists
  expected:"(Exists vars:[(Variable name:X sort:(UninterpretedSort name:S))] body:(Eq t1:(Variable name:X sort:(UninterpretedSort name:S)) t2:(Variable name:Y sort:(UninterpretedSort name:S))))")

(vector id:lambda type:Lambda
  expected:"(Lambda vars:[(Variable name:X sort:(UninterpretedSort name:S))] body:(Variable name:X sort:(UninterpretedSort name:S)))")

(vector id:named_binder_nil_env type:NamedBinder
  expected:"(NamedBinder name:nb environ:nil vars:[(Variable name:X sort:(UninterpretedSort name:S))] body:(Eq t1:(Variable name:X sort:(UninterpretedSort name:S)) t2:(Variable name:Y sort:(UninterpretedSort name:S))))")

(vector id:named_binder_with_env type:NamedBinder
  expected:"(NamedBinder name:nb environ:e1 vars:[(Variable name:X sort:(UninterpretedSort name:S))] body:(Eq t1:(Variable name:X sort:(UninterpretedSort name:S)) t2:(Variable name:Y sort:(UninterpretedSort name:S))))")

;; ============================================================
;; Definition types
;; ============================================================

(vector id:definition type:Definition
  expected:"(Def lhs:(Variable name:X sort:(UninterpretedSort name:S)) rhs:(Variable name:Y sort:(UninterpretedSort name:S)))")

(vector id:definition_schema type:DefinitionSchema
  expected:"(DefSchema lhs:(Variable name:X sort:(UninterpretedSort name:S)) rhs:(Variable name:Y sort:(UninterpretedSort name:S)))")

;; ============================================================
;; ivylogic types
;; ============================================================

(vector id:some_basic type:Some
  expected:"(Some params:[(Variable name:X sort:(UninterpretedSort name:S))] fmla:(Eq t1:(Variable name:X sort:(UninterpretedSort name:S)) t2:(Variable name:Y sort:(UninterpretedSort name:S))) ifVal:nil elseVal:nil)")

(vector id:some_with_else type:Some
  expected:"(Some params:[(Variable name:X sort:(UninterpretedSort name:S))] fmla:(Eq t1:(Variable name:X sort:(UninterpretedSort name:S)) t2:(Variable name:Y sort:(UninterpretedSort name:S))) ifVal:(Variable name:X sort:(UninterpretedSort name:S)) elseVal:(Variable name:Y sort:(UninterpretedSort name:S)))")

(vector id:let type:Let
  expected:"(Let defs:[(Def lhs:(Variable name:X sort:(UninterpretedSort name:S)) rhs:(Variable name:Y sort:(UninterpretedSort name:S)))] body:(Variable name:X sort:(UninterpretedSort name:S)))")

(vector id:literal_pos type:Literal
  expected:"(Literal polarity:1 atom:(Eq t1:(Variable name:X sort:(UninterpretedSort name:S)) t2:(Variable name:Y sort:(UninterpretedSort name:S))))")

(vector id:literal_neg type:Literal
  expected:"(Literal polarity:0 atom:(Eq t1:(Variable name:X sort:(UninterpretedSort name:S)) t2:(Variable name:Y sort:(UninterpretedSort name:S))))")

;; ============================================================
;; Clauses
;; ============================================================

(vector id:clauses_basic type:Clauses
  expected:"(clauses fmlas:[(Eq t1:(Variable name:X sort:(UninterpretedSort name:S)) t2:(Variable name:Y sort:(UninterpretedSort name:S)))] defs:[(Def lhs:(Variable name:X sort:(UninterpretedSort name:S)) rhs:(Variable name:Y sort:(UninterpretedSort name:S)))])")

(vector id:clauses_empty type:Clauses
  expected:"(clauses fmlas:[] defs:[])")
