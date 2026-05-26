package ivy2cpp

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

func (g *Generator) variantIndex(super, sub goivy.Sort) int {
	if g == nil || g.Mod == nil {
		return -1
	}
	return g.Mod.VariantIndex(super, sub)
}

func (g *Generator) variantIsaExpr(superExpr string, super, sub goivy.Sort) string {
	return fmt.Sprintf("((%s).tag == %d)", superExpr, g.variantIndex(super, sub))
}

func (g *Generator) variantClassName(className string) string {
	if className != "" || g == nil {
		return className
	}
	return g.ClassName
}

func (g *Generator) variantDowncastExpr(superExpr string, super, sub goivy.Sort, className string) string {
	className = g.variantClassName(className)
	superType := cppScalarTypeWith(g, super, className)
	subType := cppScalarTypeWith(g, sub, className)
	return fmt.Sprintf("%s::unwrap< %s >(%s)", superType, subType, superExpr)
}

func (g *Generator) variantUpcastExpr(super, sub goivy.Sort, expr string, className string) string {
	className = g.variantClassName(className)
	idx := g.variantIndex(super, sub)
	superType := cppScalarTypeWith(g, super, className)
	subType := cppScalarTypeWith(g, sub, className)
	return fmt.Sprintf("%s(%d, new %s::twrap<%s>(%s))", superType, idx, superType, subType, expr)
}

func (g *Generator) maybeVariantUpcast(target, value goivy.Sort, expr string, className string) string {
	if g != nil && g.Mod != nil && g.Mod.IsVariant(target, value) {
		return g.variantUpcastExpr(target, value, expr, className)
	}
	return expr
}

func variantSolverRelationName(super, sub goivy.Sort) string {
	return "*>:" + sortName(super) + ":" + sortName(sub)
}

func (g *Generator) emitPythonTestVariantConstraintAdd(w *cppWriter, smt string) bool {
	if g.Config.Target != "test" || !strings.Contains(smt, "|*>:") {
		return false
	}
	if g.Mod == nil || g.Mod.Sig == nil {
		return false
	}
	for _, superName := range g.Mod.SortOrder {
		variants := g.Mod.Variants[superName]
		if len(variants) != 2 {
			continue
		}
		super, ok := g.Mod.Sig.Sorts.Get2(superName)
		if !ok {
			continue
		}
		sub0 := variants[0]
		sub1 := variants[1]
		sup := sortName(super)
		a := sortName(sub0)
		b := sortName(sub1)
		relA := variantSolverRelationName(super, sub0)
		relB := variantSolverRelationName(super, sub1)
		if !strings.Contains(smt, "|"+relA+"|") || !strings.Contains(smt, "|"+relB+"|") {
			continue
		}
		w.linef("add(\"(assert (let ((a!1 (forall ((|X:%s| %s) (|Y:%s| %s) (|Z:%s| %s)) \"", sup, sup, a, a, a, a)
		w.linef("\"             (=> (and (|%s| |X:%s| |Y:%s|) \"", relA, sup, a)
		w.linef("\"                      (|%s| |X:%s| |Z:%s|)) \"", relA, sup, a)
		w.linef("\"                 (= |Y:%s| |Z:%s|)))) \"", a, a)
		w.linef("\"      (a!2 (forall ((|X:%s| %s) (|Y:%s| %s) (|Z:%s| %s)) \"", sup, sup, b, b, b, b)
		w.linef("\"             (=> (and (|%s| |X:%s| |Y:%s|) \"", relB, sup, b)
		w.linef("\"                      (|%s| |X:%s| |Z:%s|)) \"", relB, sup, b)
		w.linef("\"                 (= |Y:%s| |Z:%s|)))) \"", b, b)
		w.linef("\"      (a!3 (forall ((|X:%s| %s) (|Y:%s| %s) (|Z:%s| %s)) \"", sup, sup, sup, sup, a, a)
		w.linef("\"             (=> (and (|%s| |X:%s| |Z:%s|) \"", relA, sup, a)
		w.linef("\"                      (|%s| |Y:%s| |Z:%s|)) \"", relA, sup, a)
		w.linef("\"                 (= |X:%s| |Y:%s|)))) \"", sup, sup)
		w.linef("\"      (a!4 (forall ((|X:%s| %s) (|Y:%s| %s) (|Z:%s| %s)) \"", sup, sup, sup, sup, b, b)
		w.linef("\"             (=> (and (|%s| |X:%s| |Z:%s|) \"", relB, sup, b)
		w.linef("\"                      (|%s| |Y:%s| |Z:%s|)) \"", relB, sup, b)
		w.linef("\"                 (= |X:%s| |Y:%s|)))) \"", sup, sup)
		w.linef("\"      (a!5 (forall ((|X:%s| %s) (|Y:%s| %s) (|Z:%s| %s)) \"", sup, sup, b, b, a, a)
		w.linef("\"             (not (and (|%s| |X:%s| |Y:%s|) \"", relB, sup, b)
		w.linef("\"                       (|%s| |X:%s| |Z:%s|)))))) \"", relA, sup, a)
		w.line("\"  (and a!1 a!2 a!3 a!4 a!5)))\");")
		return true
	}
	return false
}

func (g *Generator) emitVariantWrapperDecl(w *cppWriter, name string) {
	variants := g.Mod.Variants[name]
	typeName := varName(name)
	w.open(fmt.Sprintf("class %s {", typeName))
	w.line("public:")
	w.indent++
	w.open("struct wrap {")
	w.line("virtual wrap *dup() = 0;")
	w.line("virtual bool deref() = 0;")
	w.line("virtual ~wrap() {}")
	w.close(";")
	w.open("template <typename T> struct twrap : public wrap {")
	w.line("unsigned refs;")
	w.line("T item;")
	w.line("twrap(const T &item) : refs(1), item(item) {}")
	w.line("virtual wrap *dup() {refs++; return this;}")
	w.line("virtual bool deref() {return (--refs) != 0;}")
	w.close(";")
	w.line("int tag;")
	w.line("wrap *ptr;")
	w.open(fmt.Sprintf("%s() {", typeName))
	w.line("tag = -1;")
	w.line("ptr = 0;")
	w.close("")
	w.linef("%s(int tag, wrap *ptr) : tag(tag), ptr(ptr) {}", typeName)
	w.open(fmt.Sprintf("%s(const %s &other) {", typeName, typeName))
	w.line("tag = other.tag;")
	w.line("ptr = other.ptr ? other.ptr->dup() : 0;")
	w.close(";")
	w.open(fmt.Sprintf("%s &operator=(const %s &other) {", typeName, typeName))
	w.line("tag = other.tag;")
	w.line("ptr = other.ptr ? other.ptr->dup() : 0;")
	w.line("return *this;")
	w.close(";")
	w.open(fmt.Sprintf("~%s() {", typeName))
	w.line("if (ptr) { if (!ptr->deref()) delete ptr; }")
	w.close("")
	w.line("static int temp_counter;")
	w.line("static void prepare() { temp_counter = 0; }")
	w.line("static void cleanup() {}")
	w.open("size_t __hash() const {")
	w.open("switch (tag) {")
	superSort := g.Mod.Sig.Sorts.Get(name)
	for i, v := range variants {
		subType := cppScalarTypeWith(g, v, g.variantClassName(""))
		downcast := g.variantDowncastExpr("(*this)", superSort, v, "")
		w.linef("case %d: return %d + hash_space::hash<%s>()(%s);", i, i, subType, downcast)
	}
	w.close("")
	w.line("return 0;")
	w.close("")
	w.open(fmt.Sprintf("template <typename T> static const T &unwrap(const %s &x) {", typeName))
	w.line("return ((static_cast<const twrap<T> *>(x.ptr))->item);")
	w.close("")
	w.open(fmt.Sprintf("template <typename T> static T &unwrap(%s &x) {", typeName))
	w.line("twrap<T> *p = static_cast<twrap<T> *>(x.ptr);")
	w.open("if (p->refs > 1) {")
	w.line("p = new twrap<T>(p->item);")
	w.close("")
	w.line("return ((static_cast<twrap<T> *>(p))->item);")
	w.close("")
	w.indent--
	w.close(";")
}

func (g *Generator) emitVariantImpls(w *cppWriter) {
	if g == nil || g.Mod == nil || g.Mod.Sig == nil {
		return
	}
	for _, name := range g.Mod.SortOrder {
		if !g.isVariantSuperName(name) {
			continue
		}
		g.emitVariantImpl(w, name)
	}
}

func (g *Generator) emitVariantEqualityForwardDecls(w *cppWriter) {
	if g == nil || g.Mod == nil {
		return
	}
	for _, name := range g.Mod.SortOrder {
		if !g.isVariantSuperName(name) {
			continue
		}
		typeName := g.ClassName + "::" + varName(name)
		w.linef("inline bool operator ==(const %s &s, const %s &t);;", typeName, typeName)
	}
}

func (g *Generator) emitVariantEqualityInlines(w *cppWriter) {
	if g == nil || g.Mod == nil || g.Mod.Sig == nil {
		return
	}
	for _, name := range g.Mod.SortOrder {
		if !g.isVariantSuperName(name) {
			continue
		}
		super := g.Mod.Sig.Sorts.Get(name)
		typeName := g.ClassName + "::" + varName(name)
		g.emitVariantEqualityImpl(w, super, typeName)
	}
}

func (g *Generator) emitVariantImpl(w *cppWriter, name string) {
	super := g.Mod.Sig.Sorts.Get(name)
	typeName := g.ClassName + "::" + varName(name)
	w.linef("int %s::temp_counter = 0;", typeName)
	w.blank()
	g.emitVariantStreamImpl(w, super, typeName)
	g.emitVariantArgImpl(w, super, typeName)
	g.emitVariantSerImpl(w, super, typeName)
	g.emitVariantDeserImpl(w, super, typeName)
	g.emitVariantZ3Impl(w, super, typeName)
}

func (g *Generator) emitVariantSubArgImpl(w *cppWriter, sub goivy.Sort) {
	subType := cppScalarTypeWith(g, sub, g.ClassName)
	w.linef("template <> %s _arg<%s>(std::vector<ivy_value> &args, unsigned idx, long long bound) {", subType, subType)
	w.indent++
	w.linef("return %s(_arg<long long>(args, idx, bound));", subType)
	w.indent--
	w.line("}")
	w.blank()
}

func (g *Generator) emitVariantEqualityImpl(w *cppWriter, super goivy.Sort, typeName string) {
	w.open(fmt.Sprintf("bool operator==(const %s &s, const %s &t) {", typeName, typeName))
	w.line("if (s.tag != t.tag) return false;")
	w.open("switch (s.tag) {")
	for i, sub := range g.Mod.Variants[sortName(super)] {
		lhs := g.variantDowncastExpr("s", super, sub, g.ClassName)
		rhs := g.variantDowncastExpr("t", super, sub, g.ClassName)
		w.linef("case %d: return %s == %s;", i, lhs, rhs)
	}
	w.close("")
	w.line("return true;")
	w.close("")
	w.blank()
}

func (g *Generator) emitVariantStreamImpl(w *cppWriter, super goivy.Sort, typeName string) {
	w.open(fmt.Sprintf("std::ostream &operator<<(std::ostream &s, const %s &t) {", typeName))
	w.line(`s << "{";`)
	w.open("switch (t.tag) {")
	for i, sub := range g.Mod.Variants[sortName(super)] {
		downcast := g.variantDowncastExpr("t", super, sub, g.ClassName)
		w.linef("case %d: s << %s << %s; break;", i, strconv.Quote(sortName(sub)+":"), downcast)
	}
	w.close("")
	w.line(`s << "}";`)
	w.line("return s;")
	w.close("")
	w.blank()
}

func (g *Generator) emitVariantArgImpl(w *cppWriter, super goivy.Sort, typeName string) {
	sortText := sortName(super)
	w.open(fmt.Sprintf("template <> %s _arg<%s>(std::vector<ivy_value> &args, unsigned idx, long long bound) {", typeName, typeName))
	w.line("if (args[idx].atom.size())")
	w.indent++
	w.linef(`throw out_of_bounds("unexpected value for sort %s: " + args[idx].atom, args[idx].pos);`, escapeString(sortText))
	w.indent--
	w.line("if (args[idx].fields.size() == 0)")
	w.indent++
	w.linef("return %s();", typeName)
	w.indent--
	w.line("if (args[idx].fields.size() != 1)")
	w.indent++
	w.linef(`throw out_of_bounds("too many fields for sort %s (expected one)", args[idx].pos);`, escapeString(sortText))
	w.indent--
	for _, sub := range g.Mod.Variants[sortText] {
		upcast := g.variantUpcastExpr(super, sub, g.argExprForSortBound("args[idx].fields[0].fields", "0", sub, "0"), g.ClassName)
		w.linef("if (args[idx].fields[0].atom == %s) return %s;", strconv.Quote(sortName(sub)), upcast)
	}
	w.line(`throw out_of_bounds("unexpected field sort SORTNAME: " + args[idx].fields[0].atom, args[idx].pos);`)
	w.close("")
	w.blank()
}

func (g *Generator) emitVariantSerImpl(w *cppWriter, super goivy.Sort, typeName string) {
	w.open(fmt.Sprintf("template <> void __ser<%s>(ivy_ser &res, const %s &inp) {", typeName, typeName))
	for i, sub := range g.Mod.Variants[sortName(super)] {
		downcast := g.variantDowncastExpr("inp", super, sub, g.ClassName)
		w.linef("if (inp.tag == %d) { res.open_tag(%d, %s); __ser(res, %s); res.close_tag(); }", i, i, strconv.Quote(sortName(sub)), downcast)
	}
	w.close("")
	w.blank()
}

func (g *Generator) emitVariantDeserImpl(w *cppWriter, super goivy.Sort, typeName string) {
	w.open(fmt.Sprintf("template <> void __deser<%s>(ivy_deser &res, %s &inp) {", typeName, typeName))
	w.line("std::vector<std::string> tags;")
	for _, sub := range g.Mod.Variants[sortName(super)] {
		w.linef("tags.push_back(%s);", strconv.Quote(sortName(sub)))
	}
	w.line("int tag = res.open_tag(tags);")
	w.open("switch (tag) {")
	for i, sub := range g.Mod.Variants[sortName(super)] {
		subType := cppScalarTypeWith(g, sub, g.ClassName)
		upcast := g.variantUpcastExpr(super, sub, "tmp", g.ClassName)
		w.linef("case %d: { %s tmp; __deser(res, tmp); inp = %s; break; }", i, subType, upcast)
	}
	w.close("")
	w.line("res.close_tag();")
	w.close("")
	w.blank()
}

func (g *Generator) emitVariantZ3Impl(w *cppWriter, super goivy.Sort, typeName string) {
	sortText := sortName(super)
	w.line("#ifdef Z3PP_H_")
	w.open(fmt.Sprintf("template <> void __from_solver<%s>(gen &g, const z3::expr &v, %s &res) {", typeName, typeName))
	for _, sub := range g.Mod.Variants[sortText] {
		subType := cppScalarTypeWith(g, sub, g.ClassName)
		relName := variantSolverRelationName(super, sub)
		w.line("{")
		w.indent++
		w.linef("z3::sort sort = g.sort(%s);", strconv.Quote(sortName(sub)))
		w.linef("z3::func_decl pto = g.ctx.function(%s, g.sort(%s), g.sort(%s), g.ctx.bool_sort());", strconv.Quote(relName), strconv.Quote(sortText), strconv.Quote(sortName(sub)))
		w.line("Z3_ast_vector av = Z3_model_get_sort_universe(g.ctx, g.model, sort);")
		w.open("if (av) {")
		w.line("z3::expr_vector univ(g.ctx, av);")
		w.open("for (unsigned i = 0; i < univ.size(); i++) {")
		w.open("if (eq(g.model.eval(pto(v, univ[i]), true), g.ctx.bool_val(true))) {")
		w.linef("%s tmp;", subType)
		w.line("__from_solver(g, univ[i], tmp);")
		w.linef("res = %s;", g.variantUpcastExpr(super, sub, "tmp", g.ClassName))
		w.close("")
		w.close("")
		w.close("")
		w.indent--
		w.line("}")
	}
	w.close("")
	valParam := fmt.Sprintf("%s &val", typeName)
	existsName := "exists"
	forallName := "forall"
	if g.usesZ3() && g.Config.Target != "test" {
		valParam = fmt.Sprintf("const %s &val", typeName)
		existsName = "::exists"
		forallName = "::forall"
	}
	w.open(fmt.Sprintf("template <> z3::expr __to_solver<%s>(gen &g, const z3::expr &v, %s) {", typeName, valParam))
	for i, sub := range g.Mod.Variants[sortText] {
		subType := cppScalarTypeWith(g, sub, g.ClassName)
		relName := variantSolverRelationName(super, sub)
		w.open(fmt.Sprintf("if (val.tag == %d) {", i))
		w.linef("z3::func_decl pto = g.ctx.function(%s, g.sort(%s), g.sort(%s), g.ctx.bool_sort());", strconv.Quote(relName), strconv.Quote(sortText), strconv.Quote(sortName(sub)))
		w.linef("z3::expr X = g.ctx.constant(\"X\", g.sort(%s));", strconv.Quote(sortName(sub)))
		w.linef("%s tmp = %s;", subType, g.variantDowncastExpr("val", super, sub, g.ClassName))
		w.linef("return %s(X, pto(v, X) && __to_solver(g, X, tmp));", existsName)
		w.close("")
	}
	w.line("z3::expr conj = g.ctx.bool_val(false);")
	for _, sub := range g.Mod.Variants[sortText] {
		relName := variantSolverRelationName(super, sub)
		w.line("{")
		w.indent++
		w.linef("z3::func_decl pto = g.ctx.function(%s, g.sort(%s), g.sort(%s), g.ctx.bool_sort());", strconv.Quote(relName), strconv.Quote(sortText), strconv.Quote(sortName(sub)))
		w.linef("z3::expr Y = g.ctx.constant(\"Y\", g.sort(%s));", strconv.Quote(sortName(sub)))
		w.linef("conj = conj && %s(Y, !pto(v, Y));", forallName)
		w.indent--
		w.line("}")
	}
	w.line("return conj;")
	w.close("")
	w.open(fmt.Sprintf("template <> void __randomize<%s>(gen &g, const z3::expr &apply_expr, const std::string &sort_name) {", typeName))
	w.line("std::ostringstream os;")
	w.linef("os << %s << %s::temp_counter++;", strconv.Quote("__"+sortText+"__tmp"), typeName)
	w.line("std::string temp = os.str();")
	w.line("z3::sort range = apply_expr.get_sort();")
	w.line("z3::expr disj = g.ctx.bool_val(false);")
	w.linef("int tag = __chacha8c_rng.Rand() %% %d;", len(g.Mod.Variants[sortText]))
	for i, sub := range g.Mod.Variants[sortText] {
		subType := cppScalarTypeWith(g, sub, g.ClassName)
		relName := variantSolverRelationName(super, sub)
		w.open(fmt.Sprintf("if (tag == %d) {", i))
		w.linef("z3::func_decl pto = g.ctx.function(%s, g.sort(%s), g.sort(%s), g.ctx.bool_sort());", strconv.Quote(relName), strconv.Quote(sortText), strconv.Quote(sortName(sub)))
		w.linef("z3::expr X = g.ctx.constant(temp.c_str(), g.sort(%s));", strconv.Quote(sortName(sub)))
		w.line("z3::expr pred = pto(apply_expr, X);")
		w.line("g.add_alit(pred);")
		w.linef("__randomize<%s>(g, X, %s);", subType, strconv.Quote(sortName(sub)))
		w.close("")
	}
	w.close("")
	w.line("#endif")
	w.blank()
}
