import { TokenKind } from "../../front/token.js";
import { Token as GoToken } from "../token/index.js";
export const SEND = 1 << 0;
export const RECV = 1 << 1;
export function isWhitespace(ch) {
    return ch === " " || ch === "\t" || ch === "\n" || ch === "\r";
}
export function stripTrailingWhitespace(s) {
    let i = s.length;
    while (i > 0 && isWhitespace(s[i - 1] ?? "")) {
        i -= 1;
    }
    return s.slice(0, i);
}
export function isDirective(c) {
    if (c.startsWith("line ") || c.startsWith("extern ") || c.startsWith("export ")) {
        return true;
    }
    const colon = c.indexOf(":");
    if (colon <= 0 || colon + 1 >= c.length) {
        return false;
    }
    for (let i = 0; i <= colon + 1; i += 1) {
        if (i === colon) {
            continue;
        }
        const b = c[i] ?? "";
        if (!(("a" <= b && b <= "z") || ("0" <= b && b <= "9"))) {
            return false;
        }
    }
    return true;
}
export function commentGroupText(group) {
    if (!group) {
        return "";
    }
    const comments = group.list.map((comment) => comment.text);
    const lines = [];
    for (let comment of comments) {
        switch (comment[1]) {
            case "/":
                comment = comment.slice(2);
                if (comment.length === 0) {
                    break;
                }
                if (comment[0] === " ") {
                    comment = comment.slice(1);
                    break;
                }
                if (isDirective(comment)) {
                    continue;
                }
                break;
            case "*":
                comment = comment.slice(2, -2);
                break;
            default:
                break;
        }
        for (const line of comment.split("\n")) {
            lines.push(stripTrailingWhitespace(line));
        }
    }
    let n = 0;
    for (const line of lines) {
        if (line !== "" || (n > 0 && lines[n - 1] !== "")) {
            lines[n] = line;
            n += 1;
        }
    }
    lines.length = n;
    if (n > 0 && lines[n - 1] !== "") {
        lines.push("");
    }
    return lines.join("\n");
}
export var CommentGroup;
(function (CommentGroup) {
    function Pos(group) {
        const list = group ? astList(group, "List", "list") : [];
        return PosOf(list[0]);
    }
    CommentGroup.Pos = Pos;
    function End(group) {
        const list = group ? astList(group, "List", "list") : [];
        return EndOf(list[list.length - 1]);
    }
    CommentGroup.End = End;
    function Text(group) {
        return commentGroupText(group);
    }
    CommentGroup.Text = Text;
})(CommentGroup || (CommentGroup = {}));
export function ident(name, span) {
    return { kind: "Ident", name, ...(span ? { span } : {}) };
}
export function NewIdent(name) {
    const id = ident(name);
    NormalizeAst(id);
    return id;
}
export function IsExported(name) {
    const first = Array.from(name)[0] ?? "";
    return /^\p{Lu}$/u.test(first);
}
export var Ident;
(function (Ident) {
    function Pos(id) {
        return nodePos(id);
    }
    Ident.Pos = Pos;
    function End(id) {
        return nodePos(id) + identNameValue(id).length;
    }
    Ident.End = End;
    function exprNode(_id) {
    }
    Ident.exprNode = exprNode;
    function IsExported(id) {
        return id ? astIsExported(identNameValue(id)) : false;
    }
    Ident.IsExported = IsExported;
    function String(id) {
        if (id) {
            return identNameValue(id);
        }
        return "<nil>";
    }
    Ident.String = String;
})(Ident || (Ident = {}));
export var BadExpr;
(function (BadExpr) {
    function Pos(x) { return nodePos(x); }
    BadExpr.Pos = Pos;
    function End(x) { return nodeEnd(x); }
    BadExpr.End = End;
    function exprNode(_x) { }
    BadExpr.exprNode = exprNode;
})(BadExpr || (BadExpr = {}));
export var Ellipsis;
(function (Ellipsis) {
    function Pos(x) { return nodePos(x); }
    Ellipsis.Pos = Pos;
    function End(x) {
        const element = astField(x, "Elt", "element");
        return element ? EndOf(element) : nodeEnd(x) || nodePos(x) + 3;
    }
    Ellipsis.End = End;
    function exprNode(_x) { }
    Ellipsis.exprNode = exprNode;
})(Ellipsis || (Ellipsis = {}));
export var BasicLit;
(function (BasicLit) {
    function Pos(x) { return nodePos(x); }
    BasicLit.Pos = Pos;
    function End(x) {
        return nodeEnd(x) || nodePos(x) + (x.Value ?? x.value).length;
    }
    BasicLit.End = End;
    function exprNode(_x) { }
    BasicLit.exprNode = exprNode;
})(BasicLit || (BasicLit = {}));
export var FuncLit;
(function (FuncLit) {
    function Pos(x) { return PosOf(astField(x, "Type", "type")); }
    FuncLit.Pos = Pos;
    function End(x) { return EndOf(astField(x, "Body", "body")); }
    FuncLit.End = End;
    function exprNode(_x) { }
    FuncLit.exprNode = exprNode;
})(FuncLit || (FuncLit = {}));
export var CompositeLit;
(function (CompositeLit) {
    function Pos(x) {
        const typ = astField(x, "Type", "type");
        return typ ? PosOf(typ) : nodePos(x);
    }
    CompositeLit.Pos = Pos;
    function End(x) { return nodeEnd(x); }
    CompositeLit.End = End;
    function exprNode(_x) { }
    CompositeLit.exprNode = exprNode;
})(CompositeLit || (CompositeLit = {}));
export var ParenExpr;
(function (ParenExpr) {
    function Pos(x) { return nodePos(x); }
    ParenExpr.Pos = Pos;
    function End(x) { return nodeEnd(x); }
    ParenExpr.End = End;
    function exprNode(_x) { }
    ParenExpr.exprNode = exprNode;
})(ParenExpr || (ParenExpr = {}));
export var SelectorExpr;
(function (SelectorExpr) {
    function Pos(x) { return PosOf(astField(x, "X", "object")); }
    SelectorExpr.Pos = Pos;
    function End(x) { return EndOf(astField(x, "Sel", "selector")); }
    SelectorExpr.End = End;
    function exprNode(_x) { }
    SelectorExpr.exprNode = exprNode;
})(SelectorExpr || (SelectorExpr = {}));
export var IndexExpr;
(function (IndexExpr) {
    function Pos(x) { return PosOf(astField(x, "X", "object")); }
    IndexExpr.Pos = Pos;
    function End(x) { return nodeEnd(x); }
    IndexExpr.End = End;
    function exprNode(_x) { }
    IndexExpr.exprNode = exprNode;
})(IndexExpr || (IndexExpr = {}));
export var IndexListExpr;
(function (IndexListExpr) {
    function Pos(x) { return PosOf(astField(x, "X", "object")); }
    IndexListExpr.Pos = Pos;
    function End(x) { return nodeEnd(x); }
    IndexListExpr.End = End;
    function exprNode(_x) { }
    IndexListExpr.exprNode = exprNode;
})(IndexListExpr || (IndexListExpr = {}));
export var SliceExpr;
(function (SliceExpr) {
    function Pos(x) { return PosOf(astField(x, "X", "object")); }
    SliceExpr.Pos = Pos;
    function End(x) { return nodeEnd(x); }
    SliceExpr.End = End;
    function exprNode(_x) { }
    SliceExpr.exprNode = exprNode;
})(SliceExpr || (SliceExpr = {}));
export var TypeAssertExpr;
(function (TypeAssertExpr) {
    function Pos(x) { return PosOf(astField(x, "X", "object")); }
    TypeAssertExpr.Pos = Pos;
    function End(x) { return nodeEnd(x); }
    TypeAssertExpr.End = End;
    function exprNode(_x) { }
    TypeAssertExpr.exprNode = exprNode;
})(TypeAssertExpr || (TypeAssertExpr = {}));
export var CallExpr;
(function (CallExpr) {
    function Pos(x) { return PosOf(astField(x, "Fun", "fun")); }
    CallExpr.Pos = Pos;
    function End(x) { return nodeEnd(x); }
    CallExpr.End = End;
    function exprNode(_x) { }
    CallExpr.exprNode = exprNode;
})(CallExpr || (CallExpr = {}));
export var StarExpr;
(function (StarExpr) {
    function Pos(x) { return nodePos(x); }
    StarExpr.Pos = Pos;
    function End(x) { return EndOf(astField(x, "X", "expr")); }
    StarExpr.End = End;
    function exprNode(_x) { }
    StarExpr.exprNode = exprNode;
})(StarExpr || (StarExpr = {}));
export var UnaryExpr;
(function (UnaryExpr) {
    function Pos(x) { return nodePos(x); }
    UnaryExpr.Pos = Pos;
    function End(x) { return EndOf(astField(x, "X", "expr")); }
    UnaryExpr.End = End;
    function exprNode(_x) { }
    UnaryExpr.exprNode = exprNode;
})(UnaryExpr || (UnaryExpr = {}));
export var BinaryExpr;
(function (BinaryExpr) {
    function Pos(x) { return PosOf(astField(x, "X", "left")); }
    BinaryExpr.Pos = Pos;
    function End(x) { return EndOf(astField(x, "Y", "right")); }
    BinaryExpr.End = End;
    function exprNode(_x) { }
    BinaryExpr.exprNode = exprNode;
})(BinaryExpr || (BinaryExpr = {}));
export var KeyValueExpr;
(function (KeyValueExpr) {
    function Pos(x) { return PosOf(astField(x, "Key", "key")); }
    KeyValueExpr.Pos = Pos;
    function End(x) { return EndOf(astField(x, "Value", "value")); }
    KeyValueExpr.End = End;
    function exprNode(_x) { }
    KeyValueExpr.exprNode = exprNode;
})(KeyValueExpr || (KeyValueExpr = {}));
export var ArrayType;
(function (ArrayType) {
    function Pos(x) { return nodePos(x); }
    ArrayType.Pos = Pos;
    function End(x) { return EndOf(astField(x, "Elt", "element")); }
    ArrayType.End = End;
    function exprNode(_x) { }
    ArrayType.exprNode = exprNode;
})(ArrayType || (ArrayType = {}));
export var StructType;
(function (StructType) {
    function Pos(x) { return nodePos(x); }
    StructType.Pos = Pos;
    function End(x) { return FieldList.End(astField(x, "Fields", "fields")); }
    StructType.End = End;
    function exprNode(_x) { }
    StructType.exprNode = exprNode;
})(StructType || (StructType = {}));
export var FuncType;
(function (FuncType) {
    function Pos(x) { return nodePos(x) || FieldList.Pos(astField(x, "Params", "params")); }
    FuncType.Pos = Pos;
    function End(x) {
        const results = astField(x, "Results", "results");
        return results ? FieldList.End(results) : FieldList.End(astField(x, "Params", "params"));
    }
    FuncType.End = End;
    function exprNode(_x) { }
    FuncType.exprNode = exprNode;
})(FuncType || (FuncType = {}));
export var InterfaceType;
(function (InterfaceType) {
    function Pos(x) { return nodePos(x); }
    InterfaceType.Pos = Pos;
    function End(x) { return FieldList.End(astField(x, "Methods", "methods")); }
    InterfaceType.End = End;
    function exprNode(_x) { }
    InterfaceType.exprNode = exprNode;
})(InterfaceType || (InterfaceType = {}));
export var MapType;
(function (MapType) {
    function Pos(x) { return nodePos(x); }
    MapType.Pos = Pos;
    function End(x) { return EndOf(astField(x, "Value", "value")); }
    MapType.End = End;
    function exprNode(_x) { }
    MapType.exprNode = exprNode;
})(MapType || (MapType = {}));
export var ChanType;
(function (ChanType) {
    function Pos(x) { return nodePos(x); }
    ChanType.Pos = Pos;
    function End(x) { return EndOf(astField(x, "Value", "value")); }
    ChanType.End = End;
    function exprNode(_x) { }
    ChanType.exprNode = exprNode;
})(ChanType || (ChanType = {}));
export var BadStmt;
(function (BadStmt) {
    function Pos(x) { return nodePos(x); }
    BadStmt.Pos = Pos;
    function End(x) { return nodeEnd(x); }
    BadStmt.End = End;
    function stmtNode(_x) { }
    BadStmt.stmtNode = stmtNode;
})(BadStmt || (BadStmt = {}));
export var DeclStmt;
(function (DeclStmt) {
    function Pos(x) { return PosOf(astField(x, "Decl", "decl")); }
    DeclStmt.Pos = Pos;
    function End(x) { return EndOf(astField(x, "Decl", "decl")); }
    DeclStmt.End = End;
    function stmtNode(_x) { }
    DeclStmt.stmtNode = stmtNode;
})(DeclStmt || (DeclStmt = {}));
export var EmptyStmt;
(function (EmptyStmt) {
    function Pos(x) { return nodePos(x); }
    EmptyStmt.Pos = Pos;
    function End(x) { return x.implicit ? nodePos(x) : nodeEnd(x) || nodePos(x) + 1; }
    EmptyStmt.End = End;
    function stmtNode(_x) { }
    EmptyStmt.stmtNode = stmtNode;
})(EmptyStmt || (EmptyStmt = {}));
export var LabeledStmt;
(function (LabeledStmt) {
    function Pos(x) { return PosOf(astField(x, "Label", "label")); }
    LabeledStmt.Pos = Pos;
    function End(x) { return EndOf(astField(x, "Stmt", "stmt")); }
    LabeledStmt.End = End;
    function stmtNode(_x) { }
    LabeledStmt.stmtNode = stmtNode;
})(LabeledStmt || (LabeledStmt = {}));
export var ExprStmt;
(function (ExprStmt) {
    function Pos(x) { return PosOf(astField(x, "X", "expr")); }
    ExprStmt.Pos = Pos;
    function End(x) { return EndOf(astField(x, "X", "expr")); }
    ExprStmt.End = End;
    function stmtNode(_x) { }
    ExprStmt.stmtNode = stmtNode;
})(ExprStmt || (ExprStmt = {}));
export var SendStmt;
(function (SendStmt) {
    function Pos(x) { return PosOf(astField(x, "Chan", "channel")); }
    SendStmt.Pos = Pos;
    function End(x) { return EndOf(astField(x, "Value", "value")); }
    SendStmt.End = End;
    function stmtNode(_x) { }
    SendStmt.stmtNode = stmtNode;
})(SendStmt || (SendStmt = {}));
export var IncDecStmt;
(function (IncDecStmt) {
    function Pos(x) { return PosOf(astField(x, "X", "expr")); }
    IncDecStmt.Pos = Pos;
    function End(x) { return nodeEnd(x) || EndOf(astField(x, "X", "expr")) + 2; }
    IncDecStmt.End = End;
    function stmtNode(_x) { }
    IncDecStmt.stmtNode = stmtNode;
})(IncDecStmt || (IncDecStmt = {}));
export var AssignStmt;
(function (AssignStmt) {
    function Pos(x) { return PosOf(astList(x, "Lhs", "lhs")[0]); }
    AssignStmt.Pos = Pos;
    function End(x) {
        const rhs = astList(x, "Rhs", "rhs");
        return EndOf(rhs[rhs.length - 1]);
    }
    AssignStmt.End = End;
    function stmtNode(_x) { }
    AssignStmt.stmtNode = stmtNode;
})(AssignStmt || (AssignStmt = {}));
export var GoStmt;
(function (GoStmt) {
    function Pos(x) { return nodePos(x); }
    GoStmt.Pos = Pos;
    function End(x) { return EndOf(astField(x, "Call", "call")); }
    GoStmt.End = End;
    function stmtNode(_x) { }
    GoStmt.stmtNode = stmtNode;
})(GoStmt || (GoStmt = {}));
export var DeferStmt;
(function (DeferStmt) {
    function Pos(x) { return nodePos(x); }
    DeferStmt.Pos = Pos;
    function End(x) { return EndOf(astField(x, "Call", "call")); }
    DeferStmt.End = End;
    function stmtNode(_x) { }
    DeferStmt.stmtNode = stmtNode;
})(DeferStmt || (DeferStmt = {}));
export var ReturnStmt;
(function (ReturnStmt) {
    function Pos(x) { return nodePos(x); }
    ReturnStmt.Pos = Pos;
    function End(x) {
        const results = astList(x, "Results", "results");
        return results.length > 0 ? EndOf(results[results.length - 1]) : nodeEnd(x) || nodePos(x) + 6;
    }
    ReturnStmt.End = End;
    function stmtNode(_x) { }
    ReturnStmt.stmtNode = stmtNode;
})(ReturnStmt || (ReturnStmt = {}));
export var BranchStmt;
(function (BranchStmt) {
    function Pos(x) { return nodePos(x); }
    BranchStmt.Pos = Pos;
    function End(x) {
        const label = astField(x, "Label", "label");
        if (label)
            return EndOf(label);
        const tok = x.Tok;
        if (typeof tok === "number") {
            return nodePos(x) + GoToken.String(tok).length;
        }
        return nodeEnd(x);
    }
    BranchStmt.End = End;
    function stmtNode(_x) { }
    BranchStmt.stmtNode = stmtNode;
})(BranchStmt || (BranchStmt = {}));
export var BlockStmt;
(function (BlockStmt) {
    function Pos(x) { return nodePos(x); }
    BlockStmt.Pos = Pos;
    function End(x) {
        const list = astList(x, "List", "statements");
        return nodeEnd(x) || (list.length > 0 ? EndOf(list[list.length - 1]) : nodePos(x) + 1);
    }
    BlockStmt.End = End;
    function stmtNode(_x) { }
    BlockStmt.stmtNode = stmtNode;
})(BlockStmt || (BlockStmt = {}));
export var IfStmt;
(function (IfStmt) {
    function Pos(x) { return nodePos(x); }
    IfStmt.Pos = Pos;
    function End(x) {
        const elseStmt = astField(x, "Else", "else");
        return elseStmt ? EndOf(elseStmt) : EndOf(astField(x, "Body", "body"));
    }
    IfStmt.End = End;
    function stmtNode(_x) { }
    IfStmt.stmtNode = stmtNode;
})(IfStmt || (IfStmt = {}));
export var CaseClause;
(function (CaseClause) {
    function Pos(x) { return nodePos(x); }
    CaseClause.Pos = Pos;
    function End(x) {
        const body = astList(x, "Body", "body");
        return body.length > 0 ? EndOf(body[body.length - 1]) : nodeEnd(x);
    }
    CaseClause.End = End;
    function stmtNode(_x) { }
    CaseClause.stmtNode = stmtNode;
})(CaseClause || (CaseClause = {}));
export var SwitchStmt;
(function (SwitchStmt) {
    function Pos(x) { return nodePos(x); }
    SwitchStmt.Pos = Pos;
    function End(x) { return nodeEnd(x); }
    SwitchStmt.End = End;
    function stmtNode(_x) { }
    SwitchStmt.stmtNode = stmtNode;
})(SwitchStmt || (SwitchStmt = {}));
export var TypeSwitchStmt;
(function (TypeSwitchStmt) {
    function Pos(x) { return nodePos(x); }
    TypeSwitchStmt.Pos = Pos;
    function End(x) { return nodeEnd(x); }
    TypeSwitchStmt.End = End;
    function stmtNode(_x) { }
    TypeSwitchStmt.stmtNode = stmtNode;
})(TypeSwitchStmt || (TypeSwitchStmt = {}));
export var CommClause;
(function (CommClause) {
    function Pos(x) { return nodePos(x); }
    CommClause.Pos = Pos;
    function End(x) {
        const body = astList(x, "Body", "body");
        return body.length > 0 ? EndOf(body[body.length - 1]) : nodeEnd(x);
    }
    CommClause.End = End;
    function stmtNode(_x) { }
    CommClause.stmtNode = stmtNode;
})(CommClause || (CommClause = {}));
export var SelectStmt;
(function (SelectStmt) {
    function Pos(x) { return nodePos(x); }
    SelectStmt.Pos = Pos;
    function End(x) { return nodeEnd(x); }
    SelectStmt.End = End;
    function stmtNode(_x) { }
    SelectStmt.stmtNode = stmtNode;
})(SelectStmt || (SelectStmt = {}));
export var ForStmt;
(function (ForStmt) {
    function Pos(x) { return nodePos(x); }
    ForStmt.Pos = Pos;
    function End(x) { return EndOf(astField(x, "Body", "body")); }
    ForStmt.End = End;
    function stmtNode(_x) { }
    ForStmt.stmtNode = stmtNode;
})(ForStmt || (ForStmt = {}));
export var RangeStmt;
(function (RangeStmt) {
    function Pos(x) { return nodePos(x); }
    RangeStmt.Pos = Pos;
    function End(x) { return EndOf(astField(x, "Body", "body")); }
    RangeStmt.End = End;
    function stmtNode(_x) { }
    RangeStmt.stmtNode = stmtNode;
})(RangeStmt || (RangeStmt = {}));
export var Comment;
(function (Comment) {
    function Pos(x) { return nodePos(x); }
    Comment.Pos = Pos;
    function End(x) { return nodeEnd(x) || nodePos(x) + x.text.length; }
    Comment.End = End;
})(Comment || (Comment = {}));
export var Field;
(function (Field) {
    function Pos(x) {
        const names = astList(x, "Names", "names");
        return names.length > 0 ? PosOf(names[0]) : PosOf(astField(x, "Type", "type"));
    }
    Field.Pos = Pos;
    function End(x) {
        const tag = astField(x, "Tag", "tag");
        if (tag)
            return EndOf(tag);
        const typ = astField(x, "Type", "type");
        if (typ)
            return EndOf(typ);
        const names = astList(x, "Names", "names");
        return EndOf(names[names.length - 1]);
    }
    Field.End = End;
})(Field || (Field = {}));
export var FieldList;
(function (FieldList) {
    function Pos(x) { return nodePos(x) || PosOf(x ? astList(x, "List", "fields")[0] : undefined); }
    FieldList.Pos = Pos;
    function End(x) {
        if (!x)
            return 0;
        const list = astList(x, "List", "fields");
        return nodeEnd(x) || EndOf(list[list.length - 1]);
    }
    FieldList.End = End;
    function NumFields(x) {
        if (!x)
            return 0;
        let n = 0;
        for (const field of astList(x, "List", "fields")) {
            n += astList(field, "Names", "names").length || 1;
        }
        return n;
    }
    FieldList.NumFields = NumFields;
})(FieldList || (FieldList = {}));
export var ImportSpec;
(function (ImportSpec) {
    function Pos(x) {
        const name = astField(x, "Name", "name");
        return name ? PosOf(name) : PosOf(astField(x, "Path", "path"));
    }
    ImportSpec.Pos = Pos;
    function End(x) {
        const endPos = numberField(x, "EndPos");
        if (endPos !== 0)
            return endPos;
        return EndOf(astField(x, "Path", "path"));
    }
    ImportSpec.End = End;
    function specNode(_x) { }
    ImportSpec.specNode = specNode;
})(ImportSpec || (ImportSpec = {}));
export var ValueSpec;
(function (ValueSpec) {
    function Pos(x) { return PosOf(astList(x, "Names", "names")[0]); }
    ValueSpec.Pos = Pos;
    function End(x) {
        const values = astList(x, "Values", "values");
        if (values.length > 0)
            return EndOf(values[values.length - 1]);
        const typ = astField(x, "Type", "type");
        if (typ)
            return EndOf(typ);
        const names = astList(x, "Names", "names");
        return EndOf(names[names.length - 1]);
    }
    ValueSpec.End = End;
    function specNode(_x) { }
    ValueSpec.specNode = specNode;
})(ValueSpec || (ValueSpec = {}));
export var TypeSpec;
(function (TypeSpec) {
    function Pos(x) { return PosOf(astField(x, "Name", "name")); }
    TypeSpec.Pos = Pos;
    function End(x) { return EndOf(astField(x, "Type", "type")); }
    TypeSpec.End = End;
    function specNode(_x) { }
    TypeSpec.specNode = specNode;
})(TypeSpec || (TypeSpec = {}));
export var BadDecl;
(function (BadDecl) {
    function Pos(x) { return nodePos(x); }
    BadDecl.Pos = Pos;
    function End(x) { return nodeEnd(x); }
    BadDecl.End = End;
    function declNode(_x) { }
    BadDecl.declNode = declNode;
})(BadDecl || (BadDecl = {}));
export var GenDecl;
(function (GenDecl) {
    function Pos(x) { return nodePos(x); }
    GenDecl.Pos = Pos;
    function End(x) {
        const rparen = numberField(x, "Rparen");
        if (rparen !== 0)
            return rparen + 1;
        const specs = astList(x, "Specs", "specs");
        return EndOf(specs[0]);
    }
    GenDecl.End = End;
    function declNode(_x) { }
    GenDecl.declNode = declNode;
})(GenDecl || (GenDecl = {}));
export var FuncDecl;
(function (FuncDecl) {
    function Pos(x) { return PosOf(astField(x, "Type", "type")); }
    FuncDecl.Pos = Pos;
    function End(x) {
        const body = astField(x, "Body", "body");
        return body ? EndOf(body) : EndOf(astField(x, "Type", "type"));
    }
    FuncDecl.End = End;
    function declNode(_x) { }
    FuncDecl.declNode = declNode;
})(FuncDecl || (FuncDecl = {}));
export var File;
(function (File) {
    function Pos(x) { return numberField(x, "Package"); }
    File.Pos = Pos;
    function End(x) {
        const decls = astList(x, "Decls", "declarations");
        return decls.length > 0 ? EndOf(decls[decls.length - 1]) : EndOf(astField(x, "Name", "name"));
    }
    File.End = End;
})(File || (File = {}));
export var Package;
(function (Package) {
    function Pos(_x) { return 0; }
    Package.Pos = Pos;
    function End(_x) { return 0; }
    Package.End = End;
})(Package || (Package = {}));
export function PosOf(node) {
    switch (node?.kind) {
        case "BadExpr": return BadExpr.Pos(node);
        case "Ident": return Ident.Pos(node);
        case "Ellipsis": return Ellipsis.Pos(node);
        case "BasicLit": return BasicLit.Pos(node);
        case "FuncLit": return FuncLit.Pos(node);
        case "CompositeLit": return CompositeLit.Pos(node);
        case "ParenExpr": return ParenExpr.Pos(node);
        case "SelectorExpr": return SelectorExpr.Pos(node);
        case "IndexExpr": return IndexExpr.Pos(node);
        case "IndexListExpr": return IndexListExpr.Pos(node);
        case "SliceExpr": return SliceExpr.Pos(node);
        case "TypeAssertExpr": return TypeAssertExpr.Pos(node);
        case "CallExpr": return CallExpr.Pos(node);
        case "StarExpr": return StarExpr.Pos(node);
        case "UnaryExpr": return UnaryExpr.Pos(node);
        case "BinaryExpr": return BinaryExpr.Pos(node);
        case "KeyValueExpr": return KeyValueExpr.Pos(node);
        case "ArrayType": return ArrayType.Pos(node);
        case "StructType": return StructType.Pos(node);
        case "FuncType": return FuncType.Pos(node);
        case "InterfaceType": return InterfaceType.Pos(node);
        case "MapType": return MapType.Pos(node);
        case "ChanType": return ChanType.Pos(node);
        case "BadStmt": return BadStmt.Pos(node);
        case "DeclStmt": return DeclStmt.Pos(node);
        case "EmptyStmt": return EmptyStmt.Pos(node);
        case "LabeledStmt": return LabeledStmt.Pos(node);
        case "ExprStmt": return ExprStmt.Pos(node);
        case "SendStmt": return SendStmt.Pos(node);
        case "IncDecStmt": return IncDecStmt.Pos(node);
        case "AssignStmt": return AssignStmt.Pos(node);
        case "GoStmt": return GoStmt.Pos(node);
        case "DeferStmt": return DeferStmt.Pos(node);
        case "ReturnStmt": return ReturnStmt.Pos(node);
        case "BranchStmt": return BranchStmt.Pos(node);
        case "BlockStmt": return BlockStmt.Pos(node);
        case "IfStmt": return IfStmt.Pos(node);
        case "CaseClause": return CaseClause.Pos(node);
        case "SwitchStmt": return SwitchStmt.Pos(node);
        case "TypeSwitchStmt": return TypeSwitchStmt.Pos(node);
        case "CommClause": return CommClause.Pos(node);
        case "SelectStmt": return SelectStmt.Pos(node);
        case "ForStmt": return ForStmt.Pos(node);
        case "RangeStmt": return RangeStmt.Pos(node);
        case "Comment": return Comment.Pos(node);
        case "CommentGroup": return CommentGroup.Pos(node);
        case "Field": return Field.Pos(node);
        case "FieldList": return FieldList.Pos(node);
        case "ImportSpec": return ImportSpec.Pos(node);
        case "ValueSpec": return ValueSpec.Pos(node);
        case "TypeSpec": return TypeSpec.Pos(node);
        case "BadDecl": return BadDecl.Pos(node);
        case "GenDecl": return GenDecl.Pos(node);
        case "FuncDecl": return FuncDecl.Pos(node);
        case "File": return File.Pos(node);
        case "Package": return Package.Pos(node);
        default: return nodePos(node);
    }
}
export function EndOf(node) {
    switch (node?.kind) {
        case "BadExpr": return BadExpr.End(node);
        case "Ident": return Ident.End(node);
        case "Ellipsis": return Ellipsis.End(node);
        case "BasicLit": return BasicLit.End(node);
        case "FuncLit": return FuncLit.End(node);
        case "CompositeLit": return CompositeLit.End(node);
        case "ParenExpr": return ParenExpr.End(node);
        case "SelectorExpr": return SelectorExpr.End(node);
        case "IndexExpr": return IndexExpr.End(node);
        case "IndexListExpr": return IndexListExpr.End(node);
        case "SliceExpr": return SliceExpr.End(node);
        case "TypeAssertExpr": return TypeAssertExpr.End(node);
        case "CallExpr": return CallExpr.End(node);
        case "StarExpr": return StarExpr.End(node);
        case "UnaryExpr": return UnaryExpr.End(node);
        case "BinaryExpr": return BinaryExpr.End(node);
        case "KeyValueExpr": return KeyValueExpr.End(node);
        case "ArrayType": return ArrayType.End(node);
        case "StructType": return StructType.End(node);
        case "FuncType": return FuncType.End(node);
        case "InterfaceType": return InterfaceType.End(node);
        case "MapType": return MapType.End(node);
        case "ChanType": return ChanType.End(node);
        case "BadStmt": return BadStmt.End(node);
        case "DeclStmt": return DeclStmt.End(node);
        case "EmptyStmt": return EmptyStmt.End(node);
        case "LabeledStmt": return LabeledStmt.End(node);
        case "ExprStmt": return ExprStmt.End(node);
        case "SendStmt": return SendStmt.End(node);
        case "IncDecStmt": return IncDecStmt.End(node);
        case "AssignStmt": return AssignStmt.End(node);
        case "GoStmt": return GoStmt.End(node);
        case "DeferStmt": return DeferStmt.End(node);
        case "ReturnStmt": return ReturnStmt.End(node);
        case "BranchStmt": return BranchStmt.End(node);
        case "BlockStmt": return BlockStmt.End(node);
        case "IfStmt": return IfStmt.End(node);
        case "CaseClause": return CaseClause.End(node);
        case "SwitchStmt": return SwitchStmt.End(node);
        case "TypeSwitchStmt": return TypeSwitchStmt.End(node);
        case "CommClause": return CommClause.End(node);
        case "SelectStmt": return SelectStmt.End(node);
        case "ForStmt": return ForStmt.End(node);
        case "RangeStmt": return RangeStmt.End(node);
        case "Comment": return Comment.End(node);
        case "CommentGroup": return CommentGroup.End(node);
        case "Field": return Field.End(node);
        case "FieldList": return FieldList.End(node);
        case "ImportSpec": return ImportSpec.End(node);
        case "ValueSpec": return ValueSpec.End(node);
        case "TypeSpec": return TypeSpec.End(node);
        case "BadDecl": return BadDecl.End(node);
        case "GenDecl": return GenDecl.End(node);
        case "FuncDecl": return FuncDecl.End(node);
        case "File": return File.End(node);
        case "Package": return Package.End(node);
        default: return nodeEnd(node);
    }
}
export function astIsExported(name) {
    return IsExported(name);
}
export function IsGenerated(file) {
    return generator(file)[1];
}
export function generator(file) {
    const packageOffset = numberField(file, "Package") || PosOf(astField(file, "Name", "name")) || Number.POSITIVE_INFINITY;
    for (const group of astList(file, "Comments", "comments")) {
        for (const comment of astList(group, "List", "list")) {
            if (PosOf(comment) > packageOffset) {
                break;
            }
            const prefix = "// Code generated ";
            const text = comment.Text ?? comment.text;
            if (text.includes(prefix)) {
                for (const line of text.split("\n")) {
                    if (line.startsWith(prefix) && line.endsWith(" DO NOT EDIT.")) {
                        return [line.slice(prefix.length, -" DO NOT EDIT.".length), true];
                    }
                }
            }
        }
    }
    return ["", false];
}
export function Unparen(e) {
    let expr = e;
    while (expr.kind === "ParenExpr") {
        expr = astField(expr, "X", "expr");
    }
    return expr;
}
export function exportFilter(name) {
    return IsExported(name);
}
export function FileExports(src) {
    return filterFile(src, exportFilter, true);
}
export function PackageExports(pkg) {
    return filterPackage(pkg, exportFilter, true);
}
export function filterIdentList(list, f) {
    let j = 0;
    for (const x of list) {
        if (f(identNameValue(x))) {
            list[j] = x;
            j += 1;
        }
    }
    list.length = j;
    return list;
}
export function fieldName(x) {
    switch (x.kind) {
        case "Ident":
            return x;
        case "SelectorExpr":
            if (astField(x, "X", "object")?.kind === "Ident") {
                return astField(x, "Sel", "selector");
            }
            break;
        case "StarExpr":
            return fieldName(astField(x, "X", "expr"));
    }
    return undefined;
}
export function filterFieldList(fields, filter, exportOnly) {
    if (!fields) {
        return false;
    }
    const list = astList(fields, "List", "fields");
    let j = 0;
    let removedFields = false;
    for (const field of list) {
        let keepField = false;
        const names = astList(field, "Names", "names");
        const typ = astField(field, "Type", "type");
        if (names.length === 0) {
            const name = typ ? fieldName(typ) : undefined;
            keepField = name !== undefined && filter(identNameValue(name));
        }
        else {
            const n = names.length;
            field.names = filterIdentList(names, filter);
            if (field.Names !== undefined) {
                field.Names = field.names;
            }
            if (field.names.length < n) {
                removedFields = true;
            }
            keepField = field.names.length > 0;
        }
        if (keepField) {
            if (exportOnly) {
                filterType(typ, filter, exportOnly);
            }
            list[j] = field;
            j += 1;
        }
    }
    if (j < list.length) {
        removedFields = true;
    }
    list.length = j;
    return removedFields;
}
export function filterCompositeLit(lit, filter, exportOnly) {
    const n = astList(lit, "Elts", "elements").length;
    lit.elements = filterExprList(astList(lit, "Elts", "elements"), filter, exportOnly);
    if (lit.Elts !== undefined) {
        lit.Elts = lit.elements;
    }
    if (lit.elements.length < n) {
        lit.Incomplete = true;
    }
}
export function filterExprList(list, filter, exportOnly) {
    let j = 0;
    for (const exp of list) {
        switch (exp.kind) {
            case "CompositeLit":
                filterCompositeLit(exp, filter, exportOnly);
                break;
            case "KeyValueExpr":
                {
                    const key = astField(exp, "Key", "key");
                    const value = astField(exp, "Value", "value");
                    if (key?.kind === "Ident" && !filter(identNameValue(key))) {
                        continue;
                    }
                    if (value?.kind === "CompositeLit") {
                        filterCompositeLit(value, filter, exportOnly);
                    }
                }
                break;
        }
        list[j] = exp;
        j += 1;
    }
    list.length = j;
    return list;
}
export function filterParamList(fields, filter, exportOnly) {
    if (!fields) {
        return false;
    }
    let found = false;
    for (const field of astList(fields, "List", "fields")) {
        if (filterType(astField(field, "Type", "type"), filter, exportOnly)) {
            found = true;
        }
    }
    return found;
}
export function filterType(typ, f, exportOnly) {
    if (!typ) {
        return false;
    }
    switch (typ.kind) {
        case "Ident":
            return f(identNameValue(typ));
        case "ParenExpr":
            return filterType(astField(typ, "X", "expr"), f, exportOnly);
        case "ArrayType":
            return filterType(astField(typ, "Elt", "element"), f, exportOnly);
        case "StructType": {
            const fields = astField(typ, "Fields", "fields");
            if (filterFieldList(fields, f, exportOnly)) {
                typ.Incomplete = true;
            }
            return fields ? astList(fields, "List", "fields").length > 0 : false;
        }
        case "FuncType": {
            const b1 = filterParamList(astField(typ, "Params", "params"), f, exportOnly);
            const b2 = filterParamList(astField(typ, "Results", "results"), f, exportOnly);
            return b1 || b2;
        }
        case "InterfaceType": {
            const methods = astField(typ, "Methods", "methods");
            if (filterFieldList(methods, f, exportOnly)) {
                typ.Incomplete = true;
            }
            return methods ? astList(methods, "List", "fields").length > 0 : false;
        }
        case "MapType": {
            const b1 = filterType(astField(typ, "Key", "key"), f, exportOnly);
            const b2 = filterType(astField(typ, "Value", "value"), f, exportOnly);
            return b1 || b2;
        }
        case "ChanType":
            return filterType(astField(typ, "Value", "value"), f, exportOnly);
    }
    return false;
}
export function filterSpec(spec, f, exportOnly) {
    switch (spec.kind) {
        case "ValueSpec": {
            spec.names = filterIdentList(astList(spec, "Names", "names"), f);
            if (spec.Names !== undefined) {
                spec.Names = spec.names;
            }
            spec.values = filterExprList(astList(spec, "Values", "values"), f, exportOnly);
            if (spec.Values !== undefined) {
                spec.Values = spec.values;
            }
            if (spec.names.length > 0) {
                if (exportOnly) {
                    filterType(astField(spec, "Type", "type"), f, exportOnly);
                }
                return true;
            }
            break;
        }
        case "TypeSpec": {
            if (f(identNameValue(astField(spec, "Name", "name")))) {
                if (exportOnly) {
                    filterType(astField(spec, "Type", "type"), f, exportOnly);
                }
                return true;
            }
            if (!exportOnly) {
                return filterType(astField(spec, "Type", "type"), f, exportOnly);
            }
            break;
        }
    }
    return false;
}
export function filterSpecList(list, f, exportOnly) {
    let j = 0;
    for (const spec of list) {
        if (filterSpec(spec, f, exportOnly)) {
            list[j] = spec;
            j += 1;
        }
    }
    list.length = j;
    return list;
}
export function FilterDecl(decl, f) {
    return filterDecl(decl, f, false);
}
export function filterDecl(decl, f, exportOnly) {
    switch (decl.kind) {
        case "GenDecl":
            decl.specs = filterSpecList(astList(decl, "Specs", "specs"), f, exportOnly);
            if (decl.Specs !== undefined) {
                decl.Specs = decl.specs;
            }
            return decl.specs.length > 0;
        case "FuncDecl":
            return f(identNameValue(astField(decl, "Name", "name")));
    }
    return false;
}
export function FilterFile(src, f) {
    return filterFile(src, f, false);
}
export function filterFile(src, f, exportOnly) {
    let j = 0;
    const declarations = astList(src, "Decls", "declarations");
    for (const declaration of declarations) {
        if (filterDecl(declaration, f, exportOnly)) {
            declarations[j] = declaration;
            j += 1;
        }
    }
    declarations.length = j;
    return j > 0;
}
export function FilterPackage(pkg, f) {
    return filterPackage(pkg, f, false);
}
export function filterPackage(pkg, f, exportOnly) {
    let hasDecls = false;
    for (const src of fileValues((pkg.Files ?? pkg.files))) {
        if (filterFile(src, f, exportOnly)) {
            hasDecls = true;
        }
    }
    return hasDecls;
}
export const FilterFuncDuplicates = 1 << 0;
export const FilterUnassociatedComments = 1 << 1;
export const FilterImportDuplicates = 1 << 2;
export const separator = { kind: "Comment", Slash: 0, Text: "//", text: "//" };
export function nameOf(f) {
    const receiver = astField(f, "Recv", "receiver");
    const fields = receiver ? astList(receiver, "List", "fields") : [];
    if (fields.length === 1) {
        let typ = astField(fields[0], "Type", "type");
        if (typ?.kind === "StarExpr") {
            typ = astField(typ, "X", "expr");
        }
        if (typ?.kind === "Ident") {
            return `${identNameValue(typ)}.${identNameValue(astField(f, "Name", "name"))}`;
        }
    }
    return identNameValue(astField(f, "Name", "name"));
}
export function MergePackageFiles(pkg, mode) {
    const entries = fileEntries((pkg.Files ?? pkg.files))
        .sort(([left], [right]) => left.localeCompare(right));
    let ndocs = 0;
    let ncomments = 0;
    let ndecls = 0;
    let minPos = 0;
    let maxPos = 0;
    for (let i = 0; i < entries.length; i += 1) {
        const file = entries[i][1];
        const doc = astField(file, "Doc", "doc");
        if (doc) {
            ndocs += astList(doc, "List", "list").length + 1;
        }
        ncomments += astList(file, "Comments", "comments").length;
        ndecls += astList(file, "Decls", "declarations").length;
        const fileStart = numberField(file, "FileStart") || PosOf(file);
        const fileEnd = numberField(file, "FileEnd") || EndOf(file);
        if (i === 0 || fileStart < minPos) {
            minPos = fileStart;
        }
        if (i === 0 || fileEnd > maxPos) {
            maxPos = fileEnd;
        }
    }
    let doc;
    let pos = 0;
    if (ndocs > 0) {
        const list = [];
        for (const [, file] of entries) {
            const fileDoc = astField(file, "Doc", "doc");
            if (!fileDoc) {
                continue;
            }
            if (list.length > 0) {
                list.push(separator);
            }
            list.push(...astList(fileDoc, "List", "list"));
            const packagePos = numberField(file, "Package");
            if (packagePos > pos) {
                pos = packagePos;
            }
        }
        doc = { kind: "CommentGroup", List: list, list };
    }
    let declarations = [];
    if (ndecls > 0) {
        const funcs = new Map();
        const maybeDecls = [];
        for (const [, file] of entries) {
            for (let declaration of astList(file, "Decls", "declarations")) {
                if ((mode & FilterFuncDuplicates) !== 0 && declaration.kind === "FuncDecl") {
                    const name = nameOf(declaration);
                    const index = funcs.get(name);
                    if (index !== undefined) {
                        const existing = maybeDecls[index];
                        if (existing?.kind === "FuncDecl" && astField(existing, "Doc", "doc") === undefined) {
                            maybeDecls[index] = undefined;
                        }
                        else {
                            declaration = undefined;
                        }
                    }
                    else {
                        funcs.set(name, maybeDecls.length);
                    }
                }
                maybeDecls.push(declaration);
            }
        }
        declarations = maybeDecls.filter((decl) => decl !== undefined);
    }
    let imports = [];
    if ((mode & FilterImportDuplicates) !== 0) {
        const seen = new Set();
        for (const [, file] of entries) {
            for (const imp of astList(file, "Imports", "imports")) {
                const path = astField(imp, "Path", "path");
                const key = path?.Value ?? path?.value ?? "";
                if (!seen.has(key)) {
                    imports.push(imp);
                    seen.add(key);
                }
            }
        }
    }
    else {
        for (const [, file] of entries) {
            imports.push(...astList(file, "Imports", "imports"));
        }
    }
    let comments = [];
    if ((mode & FilterUnassociatedComments) === 0) {
        comments = new Array(ncomments);
        let i = 0;
        for (const [, file] of entries) {
            for (const comment of astList(file, "Comments", "comments")) {
                comments[i] = comment;
                i += 1;
            }
        }
    }
    const name = NewIdent(pkg.Name ?? pkg.name);
    const scope = pkg.Scope ?? pkg.scope;
    const file = {
        kind: "File",
        ...(doc ? { Doc: doc, doc } : {}),
        Package: pos,
        Name: name,
        Decls: declarations,
        FileStart: minPos,
        FileEnd: maxPos,
        ...(scope ? { Scope: scope, scope } : {}),
        Imports: imports,
        Unresolved: [],
        Comments: comments,
        GoVersion: "",
        name,
        declarations,
        imports,
        unresolved: [],
        comments
    };
    return file;
}
function sortComments(list) {
    list.sort((a, b) => CommentGroup.Pos(a) - CommentGroup.Pos(b));
}
export class CommentMap extends Map {
    addComment(n, c) {
        let list = this.get(n);
        if (list === undefined || list.length === 0) {
            list = [c];
        }
        else {
            list = [...list, c];
        }
        this.set(n, list);
    }
    Update(old, replacement) {
        const list = this.get(old);
        if (list !== undefined && list.length > 0) {
            this.delete(old);
            this.set(replacement, [...(this.get(replacement) ?? []), ...list]);
        }
        return replacement;
    }
    Filter(node) {
        const umap = new CommentMap();
        Inspect(node, (n) => {
            if (n !== undefined) {
                const groups = this.get(n);
                if (groups !== undefined && groups.length > 0) {
                    umap.set(n, groups);
                }
            }
            return true;
        });
        return umap;
    }
    Comments() {
        const list = [];
        for (const entry of this.values()) {
            list.push(...entry);
        }
        sortComments(list);
        return list;
    }
    String() {
        const nodes = [...this.keys()];
        nodes.sort((a, b) => {
            const r = PosOf(a) - PosOf(b);
            if (r !== 0) {
                return r;
            }
            return EndOf(a) - EndOf(b);
        });
        let out = "CommentMap {\n";
        for (const node of nodes) {
            const comments = this.get(node) ?? [];
            const s = node.kind === "Ident" ? identNameValue(node) : node.kind;
            out += `\t${commentMapNodeAddress(node)}  ${s.padStart(20)}:  ${summary(comments)}\n`;
        }
        out += "}\n";
        return out;
    }
}
export function nodeList(n) {
    const list = [];
    Inspect(n, (node) => {
        if (node === undefined) {
            return false;
        }
        switch (node.kind) {
            case "CommentGroup":
            case "Comment":
                return false;
            default:
                list.push(node);
                return true;
        }
    });
    return list;
}
export class commentListReader {
    fset;
    list;
    index = 0;
    comment;
    pos = { Offset: 0, Line: 0 };
    end = { Offset: 0, Line: 0 };
    constructor(fset, list) {
        this.fset = fset;
        this.list = list;
    }
    eol() {
        return this.index >= this.list.length;
    }
    next() {
        if (!this.eol()) {
            const comment = this.list[this.index];
            if (comment === undefined) {
                return;
            }
            this.comment = comment;
            this.pos = sourcePosition(this.fset, this.comment, false);
            this.end = sourcePosition(this.fset, this.comment, true);
            this.index += 1;
        }
    }
}
export class nodeStack {
    list = [];
    push(n) {
        this.pop(PosOf(n));
        this.list.push(n);
    }
    pop(pos) {
        let top;
        let i = this.list.length;
        while (i > 0 && EndOf(this.list[i - 1]) <= pos) {
            top = this.list[i - 1];
            i -= 1;
        }
        this.list.length = i;
        return top;
    }
}
export function NewCommentMap(fset, node, comments) {
    if (comments.length === 0) {
        return undefined;
    }
    const cmap = new CommentMap();
    const tmp = comments.slice();
    sortComments(tmp);
    const r = new commentListReader(fset, tmp);
    r.next();
    const nodes = nodeList(node);
    nodes.push(undefined);
    let p;
    let pend = { Offset: 0, Line: 0 };
    let pg;
    let pgend = { Offset: 0, Line: 0 };
    const stack = new nodeStack();
    for (const q of nodes) {
        let qpos;
        if (q !== undefined) {
            qpos = sourcePosition(fset, q, false);
        }
        else {
            const infinity = 1 << 30;
            qpos = { Offset: infinity, Line: infinity };
        }
        while (r.comment !== undefined && r.end.Offset <= qpos.Offset) {
            const top = stack.pop(CommentGroup.Pos(r.comment));
            if (top !== undefined) {
                pg = top;
                pgend = sourcePosition(fset, pg, true);
            }
            let assoc;
            if (pg !== undefined &&
                (pgend.Line === r.pos.Line ||
                    (pgend.Line + 1 === r.pos.Line && r.end.Line + 1 < qpos.Line))) {
                assoc = pg;
            }
            else if (p !== undefined &&
                (pend.Line === r.pos.Line ||
                    (pend.Line + 1 === r.pos.Line && r.end.Line + 1 < qpos.Line) ||
                    q === undefined)) {
                assoc = p;
            }
            else {
                if (q === undefined) {
                    throw new globalThis.Error("internal error: no comments should be associated with sentinel");
                }
                assoc = q;
            }
            cmap.addComment(assoc, r.comment);
            if (r.eol()) {
                return cmap;
            }
            r.next();
        }
        if (q === undefined) {
            break;
        }
        p = q;
        pend = sourcePosition(fset, p, true);
        if (isCommentNodeGroup(q)) {
            stack.push(q);
        }
    }
    return cmap;
}
export function summary(list) {
    const maxLen = 40;
    let buf = "";
    commentLoop: for (const group of list) {
        for (const comment of group.list) {
            if (buf.length >= maxLen) {
                break commentLoop;
            }
            buf += comment.text;
        }
    }
    if (buf.length > maxLen) {
        buf = `${buf.slice(0, maxLen - 3)}...`;
    }
    return buf.replace(/[\t\n\r]/g, " ");
}
const commentMapIds = new WeakMap();
let nextCommentMapID = 1;
function commentMapNodeAddress(node) {
    let id = commentMapIds.get(node);
    if (id === undefined) {
        id = nextCommentMapID;
        nextCommentMapID += 1;
        commentMapIds.set(node, id);
    }
    return `0x${id.toString(16)}`;
}
function sourcePosition(fset, node, end) {
    const pos = end ? EndOf(node) : PosOf(node);
    const fsetPosition = fset?.Position;
    if (typeof fsetPosition === "function") {
        const got = fsetPosition.call(fset, pos);
        const offset = got.Offset ?? got.offset ?? pos;
        const line = got.Line ?? got.line ?? fallbackLine(node, end);
        return { Offset: offset, Line: line };
    }
    return { Offset: pos, Line: fallbackLine(node, end) };
}
function fallbackLine(node, end) {
    if (node.kind === "CommentGroup") {
        const comment = end ? node.list[node.list.length - 1] : node.list[0];
        return comment ? fallbackLine(comment, end) : 0;
    }
    if (node.kind === "Comment" && end) {
        return (node.span?.line ?? 0) + countLinesAfterFirst(node.text);
    }
    return node.span?.line ?? 0;
}
function countLinesAfterFirst(text) {
    let n = 0;
    for (const ch of text) {
        if (ch === "\n") {
            n += 1;
        }
    }
    return n;
}
function isCommentNodeGroup(node) {
    switch (node.kind) {
        case "File":
        case "Field":
        case "ImportSpec":
        case "ValueSpec":
        case "TypeSpec":
        case "BadDecl":
        case "GenDecl":
        case "FuncDecl":
        case "BadStmt":
        case "DeclStmt":
        case "EmptyStmt":
        case "LabeledStmt":
        case "ExprStmt":
        case "AssignStmt":
        case "IncDecStmt":
        case "ReturnStmt":
        case "BranchStmt":
        case "BlockStmt":
        case "IfStmt":
        case "CaseClause":
        case "CommClause":
        case "SwitchStmt":
        case "TypeSwitchStmt":
        case "SelectStmt":
        case "ForStmt":
        case "RangeStmt":
        case "DeferStmt":
        case "GoStmt":
        case "SendStmt":
            return true;
        default:
            return false;
    }
}
export class posSpan {
    Start;
    End;
    constructor(Start, End) {
        this.Start = Start;
        this.End = End;
    }
}
export class cgPos {
    left;
    cg;
    constructor(left, cg) {
        this.left = left;
        this.cg = cg;
    }
}
export function SortImports(fset, f) {
    for (const decl of astList(f, "Decls", "declarations")) {
        if (decl.kind !== "GenDecl" || genDeclTokForAst(decl) !== GoToken.IMPORT) {
            break;
        }
        if (!genDeclGroupedForAst(decl)) {
            continue;
        }
        let i = 0;
        const specs = [];
        const declSpecs = astList(decl, "Specs", "specs");
        for (let j = 0; j < declSpecs.length; j += 1) {
            const spec = declSpecs[j];
            const prev = declSpecs[j - 1];
            if (spec && prev && j > i && lineAt(fset, PosOf(spec)) > 1 + lineAt(fset, EndOf(prev))) {
                specs.push(...sortSpecs(fset, f, decl, declSpecs.slice(i, j)));
                i = j;
            }
        }
        specs.push(...sortSpecs(fset, f, decl, declSpecs.slice(i)));
        decl.specs = specs;
        if (decl.Specs !== undefined) {
            decl.Specs = specs;
        }
        if (specs.length > 0 && decl.Rparen !== undefined && decl.Rparen !== 0) {
            const lastSpec = specs[specs.length - 1];
            const lastLine = lineAt(fset, PosOf(lastSpec));
            let rparenLine = lineAt(fset, decl.Rparen);
            while (rparenLine > lastLine + 1) {
                rparenLine -= 1;
                mergeLine(fset, decl.Rparen, rparenLine);
            }
        }
    }
    const imports = astList(f, "Imports", "imports");
    imports.length = 0;
    for (const decl of astList(f, "Decls", "declarations")) {
        if (decl.kind === "GenDecl" && genDeclTokForAst(decl) === GoToken.IMPORT) {
            for (const spec of astList(decl, "Specs", "specs")) {
                if (spec.kind === "ImportSpec") {
                    imports.push(spec);
                }
            }
        }
    }
}
export function lineAt(fset, pos) {
    if (pos && typeof pos === "object" && "kind" in pos) {
        return lineAt(fset, PosOf(pos));
    }
    if (typeof pos === "number") {
        if (!Number.isFinite(pos) || pos === 0)
            return 0;
        const fileSet = fset;
        const positioned = fileSet?.PositionFor?.(pos, false) ?? fileSet?.Position?.(pos);
        return positioned?.Line ?? positioned?.line ?? 0;
    }
    return pos?.line ?? 0;
}
export function importPath(s) {
    if (s.kind !== "ImportSpec") {
        return "";
    }
    const path = astField(s, "Path", "path");
    try {
        return unquoteGoString(path?.Value ?? path?.value ?? "");
    }
    catch {
        return "";
    }
}
export function importName(s) {
    if (s.kind !== "ImportSpec") {
        return "";
    }
    return identNameValue(astField(s, "Name", "name"));
}
export function importComment(s) {
    if (s.kind !== "ImportSpec") {
        return "";
    }
    return CommentGroup.Text(astField(s, "Comment", "comment"));
}
export function collapse(prev, next) {
    if (importPath(next) !== importPath(prev) || importName(next) !== importName(prev)) {
        return false;
    }
    return prev.kind === "ImportSpec" && astField(prev, "Comment", "comment") === undefined;
}
export function sortSpecs(fset, f, d, specs) {
    if (specs.length <= 1) {
        return specs;
    }
    const pos = new Array(specs.length);
    for (let i = 0; i < specs.length; i += 1) {
        const spec = specs[i];
        pos[i] = new posSpan(PosOf(spec), EndOf(spec));
    }
    const begSpecs = pos[0]?.Start ?? 0;
    const endSpecs = pos[pos.length - 1]?.End ?? 0;
    const beg = lineStart(fset, f, lineAt(fset, begSpecs));
    const endLine = lineAt(fset, endSpecs);
    let end;
    const endFile = fileForPos(fset, endSpecs);
    if (endLine === maxLine(fset, f, endFile)) {
        end = endSpecs;
    }
    else {
        end = lineStart(fset, f, endLine + 1, endFile);
    }
    const fileComments = astList(f, "Comments", "comments");
    let first = fileComments.length;
    let last = -1;
    for (let i = 0; i < fileComments.length; i += 1) {
        const group = fileComments[i];
        if (!group) {
            continue;
        }
        if (nodeEnd(group) >= end) {
            break;
        }
        if (beg <= nodePos(group)) {
            if (i < first) {
                first = i;
            }
            if (i > last) {
                last = i;
            }
        }
    }
    let comments = [];
    if (last >= 0) {
        comments = fileComments.slice(first, last + 1);
    }
    const importComments = new Map();
    let specIndex = 0;
    for (const group of comments) {
        while (specIndex + 1 < specs.length && (pos[specIndex + 1]?.Start ?? 0) <= nodePos(group)) {
            specIndex += 1;
        }
        let left = false;
        if (specIndex === 0 && (pos[specIndex]?.Start ?? 0) > nodePos(group)) {
            left = true;
        }
        else if (specIndex + 1 < specs.length &&
            lineAt(fset, pos[specIndex]?.Start) + 1 === lineAt(fset, PosOf(group))) {
            specIndex += 1;
            left = true;
        }
        const spec = specs[specIndex];
        if (spec?.kind === "ImportSpec") {
            const list = importComments.get(spec) ?? [];
            list.push(new cgPos(left, group));
            importComments.set(spec, list);
        }
    }
    specs.sort((a, b) => {
        let result = importPath(a).localeCompare(importPath(b));
        if (result !== 0) {
            return result;
        }
        result = importName(a).localeCompare(importName(b));
        if (result !== 0) {
            return result;
        }
        return importComment(a).localeCompare(importComment(b));
    });
    const deduped = [];
    for (let i = 0; i < specs.length; i += 1) {
        const spec = specs[i];
        const next = specs[i + 1];
        if (spec && (!next || !collapse(spec, next))) {
            deduped.push(spec);
        }
        else if (spec) {
            const specStart = PosOf(spec);
            if (lineAt(fset, specStart) !== lineAt(fset, d.Rparen)) {
                mergeLine(fset, specStart, lineAt(fset, specStart));
            }
        }
    }
    specs = deduped;
    for (let i = 0; i < specs.length; i += 1) {
        const spec = specs[i];
        const span = pos[i];
        if (spec?.kind !== "ImportSpec" || !span) {
            continue;
        }
        const name = astField(spec, "Name", "name");
        const path = astField(spec, "Path", "path");
        if (name?.span) {
            name.span = moveSpan(name.span, span.Start);
        }
        if (path)
            updateBasicLitPos(path, span.Start);
        spec.EndPos = span.End;
        for (const group of importComments.get(spec) ?? []) {
            for (const comment of astList(group.cg, "List", "list")) {
                if (comment.span) {
                    comment.span = moveSpan(comment.span, group.left ? span.Start - 1 : span.End);
                }
            }
        }
    }
    comments.sort((a, b) => nodePos(a) - nodePos(b));
    return specs;
}
export function updateBasicLitPos(lit, pos) {
    const length = EndOf(lit) - PosOf(lit);
    lit.ValuePos = pos;
    if (lit.ValueEnd !== undefined && lit.ValueEnd !== 0) {
        lit.ValueEnd = pos + length;
    }
    if (!lit.span) {
        return;
    }
    lit.span = moveSpan(lit.span, pos);
}
export const Bad = 0;
export const Pkg = 1;
export const Con = 2;
export const Typ = 3;
export const Var = 4;
export const Fun = 5;
export const Lbl = 6;
export const objKindStrings = ["bad", "package", "const", "type", "var", "func", "label"];
export class Scope {
    Outer;
    Objects;
    constructor(Outer, objects) {
        this.Outer = Outer;
        this.Objects = objects ?? new Map();
    }
    Lookup(name) {
        return this.Objects.get(name);
    }
    Insert(obj) {
        const alt = this.Objects.get(obj.Name);
        if (alt === undefined) {
            this.Objects.set(obj.Name, obj);
        }
        return alt;
    }
    String() {
        let buf = "scope {";
        if (this !== undefined && this.Objects.size > 0) {
            buf += "\n";
            for (const obj of this.Objects.values()) {
                buf += `\t${ObjKind.String(obj.Kind)} ${obj.Name}\n`;
            }
        }
        buf += "}\n";
        return buf;
    }
}
export function NewScope(outer) {
    return new Scope(outer, new Map());
}
export class Object {
    Kind;
    Name;
    Decl;
    Data;
    Type;
    constructor(Kind, Name) {
        this.Kind = Kind;
        this.Name = Name;
    }
    Pos() {
        const name = this.Name;
        const decl = this.Decl;
        if (isField(decl)) {
            for (const n of astList(decl, "Names", "names")) {
                if (identNameValue(n) === name) {
                    return nodePos(n);
                }
            }
        }
        else if (isImportSpec(decl)) {
            const importNameNode = astField(decl, "Name", "name");
            if (importNameNode && identNameValue(importNameNode) === name) {
                return nodePos(importNameNode);
            }
            return nodePos(astField(decl, "Path", "path"));
        }
        else if (isValueSpec(decl)) {
            for (const n of astList(decl, "Names", "names")) {
                if (identNameValue(n) === name) {
                    return nodePos(n);
                }
            }
        }
        else if (isTypeSpec(decl)) {
            const typeNameNode = astField(decl, "Name", "name");
            if (identNameValue(typeNameNode) === name) {
                return nodePos(typeNameNode);
            }
        }
        else if (isFuncDecl(decl)) {
            const funcNameNode = astField(decl, "Name", "name");
            if (identNameValue(funcNameNode) === name) {
                return nodePos(funcNameNode);
            }
        }
        else if (isLabeledStmt(decl)) {
            const labelNode = astField(decl, "Label", "label");
            if (identNameValue(labelNode) === name) {
                return nodePos(labelNode);
            }
        }
        else if (isAssignStmt(decl)) {
            for (const x of astList(decl, "Lhs", "lhs")) {
                if (x.kind === "Ident" && identNameValue(x) === name) {
                    return nodePos(x);
                }
            }
        }
        else if (decl instanceof Scope) {
        }
        return 0;
    }
}
export function NewObj(kind, name) {
    return new Object(kind, name);
}
export var ObjKind;
(function (ObjKind) {
    function String(kind) {
        return objKindStrings[kind] ?? "";
    }
    ObjKind.String = String;
})(ObjKind || (ObjKind = {}));
export class pkgBuilder {
    fset;
    errors = [];
    error(pos, msg) {
        this.errors.push(new globalThis.Error(`${pos}: ${msg}`));
    }
    errorf(pos, format, ...args) {
        this.error(pos, formatPrintf(format, args));
    }
    declare(scope, altScope, obj) {
        let alt = scope.Insert(obj);
        if (alt === undefined && altScope !== undefined) {
            alt = altScope.Lookup(obj.Name);
        }
        if (alt !== undefined) {
            let prevDecl = "";
            const pos = alt.Pos();
            if (pos !== 0) {
                prevDecl = `\n\tprevious declaration at ${pos}`;
            }
            this.error(obj.Pos(), `${obj.Name} redeclared in this block${prevDecl}`);
        }
    }
}
export function resolve(scope, id) {
    for (; scope !== undefined; scope = scope.Outer) {
        const obj = scope.Lookup(identNameValue(id));
        if (obj !== undefined) {
            id.Obj = obj;
            return true;
        }
    }
    return false;
}
export function NewPackage(_fset, files, importer, universe) {
    const p = new pkgBuilder();
    p.fset = _fset;
    let pkgName = "";
    const pkgScope = NewScope(universe);
    for (const file of fileValues(files)) {
        const fileName = astField(file, "Name", "name");
        const name = identNameValue(fileName);
        switch (true) {
            case pkgName === "":
                pkgName = name;
                break;
            case name !== pkgName:
                p.errorf(nodePos(fileName), "package %s; expected %s", name, pkgName);
                continue;
        }
        for (const obj of (file.Scope ?? file.scope)?.Objects.values() ?? []) {
            p.declare(pkgScope, undefined, obj);
        }
    }
    const imports = new Map();
    for (const file of fileValues(files)) {
        if (identNameValue(astField(file, "Name", "name")) !== pkgName) {
            continue;
        }
        let importErrors = false;
        const fileScope = NewScope(pkgScope);
        for (const spec of astList(file, "Imports", "imports")) {
            if (importer === undefined) {
                importErrors = true;
                continue;
            }
            const path = importPath(spec);
            const [pkg, err] = importer(imports, path);
            if (err !== undefined || pkg === undefined) {
                p.errorf(nodePos(astField(spec, "Path", "path")), "could not import %s (%s)", path, err?.message ?? "unknown error");
                importErrors = true;
                continue;
            }
            let name = pkg.Name;
            const importNameNode = astField(spec, "Name", "name");
            if (importNameNode !== undefined) {
                name = identNameValue(importNameNode);
            }
            if (name === ".") {
                if (pkg.Data instanceof Scope) {
                    for (const obj of pkg.Data.Objects.values()) {
                        p.declare(fileScope, pkgScope, obj);
                    }
                }
            }
            else if (name !== "_") {
                const obj = NewObj(Pkg, name);
                obj.Decl = spec;
                obj.Data = pkg.Data;
                p.declare(fileScope, pkgScope, obj);
            }
        }
        if (importErrors) {
            pkgScope.Outer = undefined;
        }
        let i = 0;
        const unresolved = astList(file, "Unresolved", "unresolved");
        for (const id of unresolved) {
            if (!resolve(fileScope, id)) {
                p.errorf(nodePos(id), "undeclared name: %s", identNameValue(id));
                unresolved[i] = id;
                i += 1;
            }
        }
        unresolved.length = i;
        pkgScope.Outer = universe;
    }
    p.errors.sort((a, b) => a.message.localeCompare(b.message));
    return [{ kind: "Package", name: pkgName, scope: pkgScope, imports, files }, p.errors[0]];
}
export function NotNilFilter(_name, value) {
    switch (kindOf(value)) {
        case "chan":
        case "func":
        case "interface":
        case "map":
        case "pointer":
        case "slice":
            return value !== null && value !== undefined;
    }
    return true;
}
export function Fprint(w, fset, x, f) {
    return fprint(w, fset, x, f);
}
export function fprint(w, fset, x, f) {
    const p = new printer(w, fset, f);
    try {
        if (x === null || x === undefined) {
            p.printf("nil\n");
            return undefined;
        }
        p.print(x);
        p.printf("\n");
        return undefined;
    }
    catch (err) {
        if (err instanceof localError) {
            return err.err;
        }
        throw err;
    }
}
export function Print(fset, x) {
    return Fprint(defaultPrintWriter, fset, x, NotNilFilter);
}
export class printer {
    output;
    fset;
    filter;
    ptrmap = new WeakMap();
    indent = 0;
    last = "\n";
    line = 0;
    constructor(output, fset, filter) {
        this.output = output;
        this.fset = fset;
        this.filter = filter;
    }
    Write(data) {
        const text = typeof data === "string" ? data : new TextDecoder().decode(data);
        let n = 0;
        try {
            for (let i = 0; i < text.length; i += 1) {
                const b = text[i] ?? "";
                if (b === "\n") {
                    this.writeRaw(text.slice(n, i + 1));
                    n = i + 1;
                    this.line += 1;
                }
                else if (this.last === "\n") {
                    this.writeRaw(`${this.line.toString().padStart(6, " ")}  `);
                    for (let j = this.indent; j > 0; j -= 1) {
                        this.writeRaw(indent);
                    }
                }
                this.last = b;
            }
            if (text.length > n) {
                this.writeRaw(text.slice(n));
                n = text.length;
            }
            return [n, undefined];
        }
        catch (err) {
            return [n, err instanceof Error ? err : new globalThis.Error(String(err))];
        }
    }
    printf(format, ...args) {
        const [, err] = this.Write(formatPrintf(format, args));
        if (err) {
            throw new localError(err);
        }
    }
    print(x) {
        if (!NotNilFilter("", x)) {
            this.printf("nil");
            return;
        }
        switch (kindOf(x)) {
            case "interface":
                this.print(x);
                return;
            case "map": {
                const map = x;
                this.printf("%s (len = %d) {", "Map", map.size);
                if (map.size > 0) {
                    this.indent += 1;
                    this.printf("\n");
                    for (const [key, value] of map) {
                        this.print(key);
                        this.printf(": ");
                        this.print(value);
                        this.printf("\n");
                    }
                    this.indent -= 1;
                }
                this.printf("}");
                return;
            }
            case "pointer": {
                const ptr = x;
                this.printf("*");
                const line = this.ptrmap.get(ptr);
                if (line !== undefined) {
                    this.printf("(obj @ %d)", line);
                }
                else {
                    this.ptrmap.set(ptr, this.line);
                    this.printObject(ptr);
                }
                return;
            }
            case "array":
            case "slice": {
                if (x instanceof Uint8Array) {
                    this.printf("%#q", Array.from(x));
                    return;
                }
                const list = x;
                this.printf("%s (len = %d) {", Array.isArray(x) ? "Array" : "Slice", list.length);
                if (list.length > 0) {
                    this.indent += 1;
                    this.printf("\n");
                    for (let i = 0; i < list.length; i += 1) {
                        this.printf("%d: ", i);
                        this.print(list[i]);
                        this.printf("\n");
                    }
                    this.indent -= 1;
                }
                this.printf("}");
                return;
            }
            case "struct":
                this.printObject(x);
                return;
            default:
                if (typeof x === "string") {
                    this.printf("%q", x);
                    return;
                }
                this.printf("%v", x);
        }
    }
    writeRaw(data) {
        this.output.write(data);
    }
    printObject(x) {
        const fields = globalThis.Object.entries(x);
        this.printf("%s {", objectTypeName(x));
        this.indent += 1;
        let first = true;
        for (const [name, value] of fields) {
            if (this.filter === undefined || this.filter(name, value)) {
                if (first) {
                    this.printf("\n");
                    first = false;
                }
                this.printf("%s: ", name);
                this.print(value);
                this.printf("\n");
            }
        }
        this.indent -= 1;
        this.printf("}");
    }
}
export const indent = ".  ";
export class localError extends Error {
    err;
    constructor(err) {
        super(err.message);
        this.err = err;
    }
}
function kindOf(value) {
    if (value === null || value === undefined) {
        return "interface";
    }
    if (value instanceof Map) {
        return "map";
    }
    if (value instanceof Uint8Array) {
        return "slice";
    }
    if (Array.isArray(value)) {
        return "array";
    }
    if (typeof value === "function") {
        return "func";
    }
    if (typeof value === "object") {
        return "struct";
    }
    return "other";
}
function formatPrintf(format, args) {
    let index = 0;
    return format.replace(/%#q|%[sdvq%]/g, (verb) => {
        if (verb === "%%") {
            return "%";
        }
        const arg = args[index];
        index += 1;
        switch (verb) {
            case "%#q":
            case "%q":
                return quoteValue(arg);
            case "%d":
                return typeof arg === "bigint" ? arg.toString() : String(Math.trunc(Number(arg)));
            case "%s":
                return String(arg);
            case "%v":
            default:
                return formatValue(arg);
        }
    });
}
function quoteValue(value) {
    if (typeof value === "string") {
        return JSON.stringify(value);
    }
    if (value instanceof Uint8Array) {
        return JSON.stringify(Array.from(value));
    }
    return JSON.stringify(value);
}
function formatValue(value) {
    if (typeof value === "bigint") {
        return value.toString();
    }
    if (typeof value === "string") {
        return value;
    }
    if (value === undefined) {
        return "<undefined>";
    }
    return String(value);
}
function objectTypeName(value) {
    if ("kind" in value && typeof value.kind === "string") {
        return value.kind;
    }
    const name = value.constructor?.name;
    return name && name !== "Object" ? name : "object";
}
const defaultPrintWriter = {
    write(data) {
        const maybeProcess = globalThis;
        if (maybeProcess.process?.stdout?.write) {
            maybeProcess.process.stdout.write(data);
            return;
        }
        globalThis.console?.log(data);
    }
};
function fileValues(files) {
    if (Array.isArray(files)) {
        return files;
    }
    if (files instanceof Map) {
        return [...files.values()];
    }
    return globalThis.Object.values(files);
}
function fileEntries(files) {
    if (Array.isArray(files)) {
        return files.map((file, index) => [String(index), file]);
    }
    if (files instanceof Map) {
        return [...files.entries()];
    }
    return globalThis.Object.entries(files);
}
function isField(value) {
    return isNodeKind(value, "Field");
}
function isImportSpec(value) {
    return isNodeKind(value, "ImportSpec");
}
function isValueSpec(value) {
    return isNodeKind(value, "ValueSpec");
}
function isTypeSpec(value) {
    return isNodeKind(value, "TypeSpec");
}
function isFuncDecl(value) {
    return isNodeKind(value, "FuncDecl");
}
function isLabeledStmt(value) {
    return isNodeKind(value, "LabeledStmt");
}
function isAssignStmt(value) {
    return isNodeKind(value, "AssignStmt");
}
function isNodeKind(value, kind) {
    return typeof value === "object" && value !== null && "kind" in value && value.kind === kind;
}
function identNameValue(id) {
    return id?.name ?? id?.Name ?? "";
}
function numberField(node, ...names) {
    const object = node;
    if (!object)
        return 0;
    for (const name of names) {
        const value = object[name];
        if (typeof value === "number")
            return value;
    }
    return 0;
}
function genDeclTokForAst(decl) {
    if (typeof decl.Tok === "number")
        return decl.Tok;
    if (decl.token !== undefined)
        return goToken(decl.token);
    return GoToken.ILLEGAL;
}
function genDeclGroupedForAst(decl) {
    return numberField(decl, "Lparen") !== 0 || decl.grouped === true;
}
function nodePos(node) {
    return node?.span?.offset ??
        numberField(node, "NamePos", "From", "Ellipsis", "ValuePos", "Lbrace", "Lparen", "Lbrack", "Star", "OpPos", "Struct", "Func", "Interface", "Map", "Begin", "Semicolon", "TokPos", "If", "Case", "Switch", "Select", "For", "Defer", "Go", "Return", "Opening", "Package", "FileStart", "Slash");
}
function nodeEnd(node) {
    if (node?.span) {
        return node.span.offset + node.span.length;
    }
    const direct = numberField(node, "To", "ValueEnd", "FileEnd", "EndPos");
    if (direct !== 0) {
        return direct;
    }
    const closing = numberField(node, "Rbrace", "Rparen", "Rbrack", "Closing");
    if (closing !== 0) {
        return closing + 1;
    }
    const object = node;
    if (object?.kind === "Ident") {
        return nodePos(node) + identNameValue(node).length;
    }
    if (object?.kind === "BasicLit" && typeof object.Value === "string") {
        return nodePos(node) + object.Value.length;
    }
    if (object?.kind === "Comment" && typeof object.Text === "string") {
        return nodePos(node) + object.Text.length;
    }
    return 0;
}
function moveSpan(span, offset) {
    return {
        ...span,
        offset
    };
}
function fileForPos(fset, pos) {
    return fset?.File?.(pos);
}
function lineStart(fset, file, line, tokenFile) {
    const start = tokenFile?.LineStart(line);
    if (start !== undefined) {
        return start;
    }
    const fileSetFile = fileForPos(fset, PosOf(file));
    const fileSetStart = fileSetFile?.LineStart(line);
    if (fileSetStart !== undefined) {
        return fileSetStart;
    }
    for (const node of childNodes(file)) {
        const span = node.span;
        if (span && span.line === line) {
            return span.offset - Math.max(0, span.column - 1);
        }
    }
    return line;
}
function maxLine(fset, file, tokenFile) {
    const fileLineCount = tokenFile?.LineCount();
    if (fileLineCount !== undefined) {
        return fileLineCount;
    }
    const fileSetFile = fileForPos(fset, PosOf(file));
    const fileSetLineCount = fileSetFile?.LineCount();
    if (fileSetLineCount !== undefined) {
        return fileSetLineCount;
    }
    let line = 0;
    walk(file, (node) => {
        if (node.span) {
            line = Math.max(line, node.span.line);
        }
    });
    return line;
}
function mergeLine(fset, pos, line) {
    fileForPos(fset, pos)?.MergeLine(line);
}
function unquoteGoString(value) {
    if (value.length >= 2 && value.startsWith("`") && value.endsWith("`")) {
        return value.slice(1, -1).replace(/\r/g, "");
    }
    if (value.length >= 2 && value.startsWith("\"") && value.endsWith("\"")) {
        return decodeGoEscaped(value.slice(1, -1));
    }
    throw new globalThis.Error("invalid quoted string");
}
function decodeGoEscaped(value) {
    let decoded = "";
    for (let index = 0; index < value.length; index += 1) {
        const char = value[index] ?? "";
        if (char !== "\\") {
            decoded += char;
            continue;
        }
        const next = value[index + 1] ?? "";
        index += 1;
        switch (next) {
            case "a":
                decoded += "\x07";
                break;
            case "b":
                decoded += "\b";
                break;
            case "f":
                decoded += "\f";
                break;
            case "n":
                decoded += "\n";
                break;
            case "r":
                decoded += "\r";
                break;
            case "t":
                decoded += "\t";
                break;
            case "v":
                decoded += "\x0b";
                break;
            case "\\":
            case "\"":
            case "'":
                decoded += next;
                break;
            case "x":
                decoded += codePointFromEscape(value.slice(index + 1, index + 3), 16);
                index += 2;
                break;
            case "u":
                decoded += codePointFromEscape(value.slice(index + 1, index + 5), 16);
                index += 4;
                break;
            case "U":
                decoded += codePointFromEscape(value.slice(index + 1, index + 9), 16);
                index += 8;
                break;
            default:
                if (/^[0-7]$/u.test(next)) {
                    const digits = next + value.slice(index + 1, index + 3);
                    decoded += codePointFromEscape(digits, 8);
                    index += 2;
                }
                else {
                    decoded += next;
                }
                break;
        }
    }
    return decoded;
}
function codePointFromEscape(digits, radix) {
    const value = Number.parseInt(digits, radix);
    if (!Number.isFinite(value)) {
        return "";
    }
    try {
        return String.fromCodePoint(value);
    }
    catch {
        return "";
    }
}
export class Directive {
    Tool;
    Name;
    Args;
    Slash;
    ArgsPos;
    constructor(Tool, Name, Args, Slash, ArgsPos) {
        this.Tool = Tool;
        this.Name = Name;
        this.Args = Args;
        this.Slash = Slash;
        this.ArgsPos = ArgsPos;
    }
    Pos() {
        return this.Slash;
    }
    End() {
        return this.ArgsPos + this.Args.length;
    }
    ParseArgs() {
        const args = new directiveScanner(this.Args, this.ArgsPos);
        const list = [];
        for (args.skipSpace(); args.str !== ""; args.skipSpace()) {
            let arg;
            const argPos = args.pos;
            switch (args.str[0]) {
                default:
                    arg = args.takeNonSpace();
                    break;
                case "`":
                case "\"": {
                    const quoted = quotedPrefix(args.str);
                    if (!quoted) {
                        return [[], new globalThis.Error(`invalid quoted string in //${this.Tool}:${this.Name}: ${args.str}`)];
                    }
                    arg = unquoteDirectiveArg(args.take(quoted.length));
                    if (args.str !== "" && !/^\s/u.test(args.str)) {
                        return [[], new globalThis.Error(`invalid quoted string in //${this.Tool}:${this.Name}: ${args.str}`)];
                    }
                    break;
                }
            }
            list.push(new DirectiveArg(arg, argPos));
        }
        return [list, undefined];
    }
}
export class DirectiveArg {
    Arg;
    Pos;
    constructor(Arg, Pos) {
        this.Arg = Arg;
        this.Pos = Pos;
    }
}
export function ParseDirective(pos, c) {
    if (!(c.length >= 3 && c[0] === "/" && c[1] === "/" && isalnum(c[2] ?? ""))) {
        return [new Directive("", "", "", 0, 0), false];
    }
    const buf = new directiveScanner(c, pos);
    buf.skip("//".length);
    const colon = buf.str.indexOf(":");
    if (colon <= 0 || colon + 1 >= buf.str.length) {
        return [new Directive("", "", "", 0, 0), false];
    }
    for (let i = 0; i <= colon + 1; i += 1) {
        if (i === colon) {
            continue;
        }
        if (!isalnum(buf.str[i] ?? "")) {
            return [new Directive("", "", "", 0, 0), false];
        }
    }
    const tool = buf.take(colon);
    buf.skip(":".length);
    const name = buf.takeNonSpace();
    buf.skipSpace();
    const argsPos = buf.pos;
    const args = trimRightUnicodeSpace(buf.str);
    return [new Directive(tool, name, args, pos, argsPos), true];
}
export function isalnum(b) {
    return ("a" <= b && b <= "z") || ("0" <= b && b <= "9");
}
class directiveScanner {
    str;
    pos;
    constructor(str, pos) {
        this.str = str;
        this.pos = pos;
    }
    skip(n) {
        this.pos += n;
        this.str = this.str.slice(n);
    }
    take(n) {
        const res = this.str.slice(0, n);
        this.skip(n);
        return res;
    }
    takeNonSpace() {
        const match = /\s/u.exec(this.str);
        let i = match?.index ?? -1;
        if (i === -1) {
            i = this.str.length;
        }
        return this.take(i);
    }
    skipSpace() {
        const trim = this.str.replace(/^\s+/u, "");
        this.skip(this.str.length - trim.length);
    }
}
function quotedPrefix(text) {
    const quote = text[0];
    if (quote === "`") {
        const end = text.indexOf("`", 1);
        return end >= 0 ? text.slice(0, end + 1) : undefined;
    }
    if (quote !== "\"")
        return undefined;
    let escaped = false;
    for (let index = 1; index < text.length; index += 1) {
        const char = text[index] ?? "";
        if (escaped) {
            escaped = false;
            continue;
        }
        if (char === "\\") {
            escaped = true;
            continue;
        }
        if (char === "\"") {
            return text.slice(0, index + 1);
        }
    }
    return undefined;
}
function unquoteDirectiveArg(text) {
    if (text.startsWith("`")) {
        return text.slice(1, -1);
    }
    return JSON.parse(text);
}
function trimRightUnicodeSpace(text) {
    return text.replace(/\s+$/u, "");
}
export function parseCellAddress(raw) {
    const match = /^(\$?)([A-Z]{1,3})(\$?)([1-9][0-9]*)$/.exec(raw);
    if (!match)
        return undefined;
    const [, columnAbs, column, rowAbs, row] = match;
    if (!column || !row)
        return undefined;
    return {
        raw,
        column,
        row: Number(row),
        absoluteColumn: columnAbs === "$",
        absoluteRow: rowAbs === "$"
    };
}
export function walk(node, visit) {
    visit(node);
    for (const child of childNodes(node)) {
        walk(child, visit);
    }
}
export function walkList(v, list) {
    for (const node of list) {
        Walk(v, node);
    }
}
export function Walk(v, node) {
    v = v.Visit(node);
    if (v === undefined) {
        return;
    }
    for (const child of childNodes(node)) {
        Walk(v, child);
    }
    v.Visit(undefined);
}
export class inspector {
    f;
    constructor(f) {
        this.f = f;
    }
    Visit(node) {
        if (this.f(node)) {
            return this;
        }
        return undefined;
    }
}
export function Inspect(node, f) {
    Walk(new inspector(f), node);
}
export function* Preorder(root) {
    yield root;
    for (const child of childNodes(root)) {
        yield* Preorder(child);
    }
}
export function PreorderStack(root, stack, f) {
    const before = stack.length;
    Inspect(root, (node) => {
        if (node !== undefined) {
            if (!f(node, stack)) {
                return false;
            }
            stack.push(node);
        }
        else {
            stack.length = stack.length - 1;
        }
        return true;
    });
    if (stack.length !== before) {
        throw new globalThis.Error("push/pop mismatch");
    }
}
function defineGoField(node, name, value) {
    globalThis.Object.defineProperty(node, name, {
        configurable: true,
        enumerable: false,
        writable: true,
        value
    });
}
function spanPos(node) {
    return node?.span?.offset ?? 0;
}
function spanEnd(node) {
    return node?.span ? node.span.offset + node.span.length : spanPos(node);
}
function closingSpanPos(node) {
    const end = spanEnd(node);
    const start = spanPos(node);
    return end > start ? end - 1 : end;
}
function normalizeAstNode(node) {
    if (node)
        NormalizeAst(node);
}
function normalizeAstNodes(nodes) {
    if (!nodes)
        return;
    for (const node of nodes)
        NormalizeAst(node);
}
function goToken(kind) {
    switch (kind) {
        case TokenKind.Illegal:
            return GoToken.ILLEGAL;
        case TokenKind.EOF:
            return GoToken.EOF;
        case TokenKind.Identifier:
        case TokenKind.CellAddress:
            return GoToken.IDENT;
        case TokenKind.IntLiteral:
            return GoToken.INT;
        case TokenKind.FloatLiteral:
            return GoToken.FLOAT;
        case TokenKind.ImagLiteral:
            return GoToken.IMAG;
        case TokenKind.RuneLiteral:
            return GoToken.CHAR;
        case TokenKind.StringLiteral:
            return GoToken.STRING;
        case TokenKind.Import:
            return GoToken.IMPORT;
        case TokenKind.Package:
            return GoToken.PACKAGE;
        case TokenKind.Func:
            return GoToken.FUNC;
        case TokenKind.Return:
            return GoToken.RETURN;
        case TokenKind.If:
            return GoToken.IF;
        case TokenKind.Else:
            return GoToken.ELSE;
        case TokenKind.Switch:
            return GoToken.SWITCH;
        case TokenKind.Case:
            return GoToken.CASE;
        case TokenKind.Default:
            return GoToken.DEFAULT;
        case TokenKind.Fallthrough:
            return GoToken.FALLTHROUGH;
        case TokenKind.Goto:
            return GoToken.GOTO;
        case TokenKind.For:
            return GoToken.FOR;
        case TokenKind.Range:
            return GoToken.RANGE;
        case TokenKind.Break:
            return GoToken.BREAK;
        case TokenKind.Continue:
            return GoToken.CONTINUE;
        case TokenKind.Defer:
            return GoToken.DEFER;
        case TokenKind.Var:
            return GoToken.VAR;
        case TokenKind.Const:
            return GoToken.CONST;
        case TokenKind.Type:
            return GoToken.TYPE;
        case TokenKind.Struct:
            return GoToken.STRUCT;
        case TokenKind.Interface:
            return GoToken.INTERFACE;
        case TokenKind.Map:
            return GoToken.MAP;
        case TokenKind.Chan:
            return GoToken.CHAN;
        case TokenKind.Go:
            return GoToken.GO;
        case TokenKind.Select:
            return GoToken.SELECT;
        case TokenKind.Ellipsis:
            return GoToken.ELLIPSIS;
        case TokenKind.Define:
            return GoToken.DEFINE;
        case TokenKind.Assign:
            return GoToken.ASSIGN;
        case TokenKind.PlusAssign:
            return GoToken.ADD_ASSIGN;
        case TokenKind.MinusAssign:
            return GoToken.SUB_ASSIGN;
        case TokenKind.StarAssign:
            return GoToken.MUL_ASSIGN;
        case TokenKind.SlashAssign:
            return GoToken.QUO_ASSIGN;
        case TokenKind.PercentAssign:
            return GoToken.REM_ASSIGN;
        case TokenKind.AmpAssign:
            return GoToken.AND_ASSIGN;
        case TokenKind.OrAssign:
            return GoToken.OR_ASSIGN;
        case TokenKind.CaretAssign:
            return GoToken.XOR_ASSIGN;
        case TokenKind.ShlAssign:
            return GoToken.SHL_ASSIGN;
        case TokenKind.ShrAssign:
            return GoToken.SHR_ASSIGN;
        case TokenKind.BitClearAssign:
            return GoToken.AND_NOT_ASSIGN;
        case TokenKind.Equal:
            return GoToken.EQL;
        case TokenKind.NotEqual:
            return GoToken.NEQ;
        case TokenKind.Less:
            return GoToken.LSS;
        case TokenKind.LessEqual:
            return GoToken.LEQ;
        case TokenKind.Greater:
            return GoToken.GTR;
        case TokenKind.GreaterEqual:
            return GoToken.GEQ;
        case TokenKind.AndAnd:
            return GoToken.LAND;
        case TokenKind.OrOr:
            return GoToken.LOR;
        case TokenKind.Arrow:
            return GoToken.ARROW;
        case TokenKind.Plus:
            return GoToken.ADD;
        case TokenKind.PlusPlus:
            return GoToken.INC;
        case TokenKind.Minus:
            return GoToken.SUB;
        case TokenKind.MinusMinus:
            return GoToken.DEC;
        case TokenKind.Star:
            return GoToken.MUL;
        case TokenKind.Slash:
            return GoToken.QUO;
        case TokenKind.Percent:
            return GoToken.REM;
        case TokenKind.Or:
            return GoToken.OR;
        case TokenKind.Caret:
            return GoToken.XOR;
        case TokenKind.Tilde:
            return GoToken.TILDE;
        case TokenKind.Shl:
            return GoToken.SHL;
        case TokenKind.Shr:
            return GoToken.SHR;
        case TokenKind.BitClear:
            return GoToken.AND_NOT;
        case TokenKind.Bang:
            return GoToken.NOT;
        case TokenKind.Amp:
            return GoToken.AND;
        case TokenKind.Dot:
            return GoToken.PERIOD;
        case TokenKind.Comma:
            return GoToken.COMMA;
        case TokenKind.Colon:
            return GoToken.COLON;
        case TokenKind.Semicolon:
            return GoToken.SEMICOLON;
        case TokenKind.LParen:
            return GoToken.LPAREN;
        case TokenKind.RParen:
            return GoToken.RPAREN;
        case TokenKind.LBrace:
            return GoToken.LBRACE;
        case TokenKind.RBrace:
            return GoToken.RBRACE;
        case TokenKind.LBracket:
            return GoToken.LBRACK;
        case TokenKind.RBracket:
            return GoToken.RBRACK;
        case TokenKind.True:
        case TokenKind.False:
        case TokenKind.Nil:
            return GoToken.IDENT;
    }
}
function chanDir(direction) {
    switch (direction) {
        case "send":
            return SEND;
        case "receive":
            return RECV;
        case "both":
            return SEND | RECV;
    }
}
// NormalizeAst mirrors the standard go/ast exported struct field names onto
// the existing GoJr node objects. The aliases are non-enumerable so code that
// already serializes or compares the lower-case compatibility fields keeps the
// same behavior while callers can use the standard Go field names.
export function NormalizeAst(node) {
    if (!node)
        return;
    defineGoField(node, "Pos", () => PosOf(node));
    defineGoField(node, "End", () => EndOf(node));
    switch (node.kind) {
        case "Ident":
            defineGoField(node, "NamePos", spanPos(node));
            defineGoField(node, "Name", node.name);
            defineGoField(node, "Obj", node.Obj);
            break;
        case "BadExpr":
            defineGoField(node, "From", spanPos(node));
            defineGoField(node, "To", spanEnd(node));
            break;
        case "Ellipsis":
            normalizeAstNode(node.element);
            defineGoField(node, "Ellipsis", spanPos(node));
            defineGoField(node, "Elt", node.element);
            break;
        case "BasicLit":
            defineGoField(node, "ValuePos", spanPos(node));
            defineGoField(node, "ValueEnd", spanEnd(node));
            defineGoField(node, "Kind", goToken(node.token));
            defineGoField(node, "Value", node.value);
            break;
        case "FuncLit":
            normalizeAstNode(node.type);
            normalizeAstNode(node.body);
            defineGoField(node, "Type", node.type);
            defineGoField(node, "Body", node.body);
            break;
        case "CompositeLit":
            normalizeAstNode(node.type);
            normalizeAstNodes(node.elements);
            defineGoField(node, "Type", node.type);
            defineGoField(node, "Lbrace", spanPos(node));
            defineGoField(node, "Elts", node.elements);
            defineGoField(node, "Rbrace", closingSpanPos(node));
            defineGoField(node, "Incomplete", false);
            break;
        case "ParenExpr":
            normalizeAstNode(node.expr);
            defineGoField(node, "Lparen", spanPos(node));
            defineGoField(node, "X", node.expr);
            defineGoField(node, "Rparen", closingSpanPos(node));
            break;
        case "SelectorExpr":
            normalizeAstNode(node.object);
            normalizeAstNode(node.selector);
            defineGoField(node, "X", node.object);
            defineGoField(node, "Sel", node.selector);
            break;
        case "IndexExpr":
            normalizeAstNode(node.object);
            normalizeAstNode(node.index);
            defineGoField(node, "X", node.object);
            defineGoField(node, "Lbrack", spanPos(node.index));
            defineGoField(node, "Index", node.index);
            defineGoField(node, "Rbrack", closingSpanPos(node));
            break;
        case "IndexListExpr":
            normalizeAstNode(node.object);
            normalizeAstNodes(node.indices);
            defineGoField(node, "X", node.object);
            defineGoField(node, "Lbrack", node.indices[0] ? spanPos(node.indices[0]) : spanPos(node));
            defineGoField(node, "Indices", node.indices);
            defineGoField(node, "Rbrack", closingSpanPos(node));
            break;
        case "SliceExpr":
            normalizeAstNode(node.object);
            normalizeAstNode(node.low);
            normalizeAstNode(node.high);
            normalizeAstNode(node.max);
            defineGoField(node, "X", node.object);
            defineGoField(node, "Lbrack", spanPos(node));
            defineGoField(node, "Low", node.low);
            defineGoField(node, "High", node.high);
            defineGoField(node, "Max", node.max);
            defineGoField(node, "Slice3", node.max !== undefined);
            defineGoField(node, "Rbrack", closingSpanPos(node));
            break;
        case "TypeAssertExpr":
            normalizeAstNode(node.object);
            normalizeAstNode(node.type);
            defineGoField(node, "X", node.object);
            defineGoField(node, "Lparen", spanPos(node));
            defineGoField(node, "Type", node.type);
            defineGoField(node, "Rparen", closingSpanPos(node));
            break;
        case "CallExpr":
            normalizeAstNode(node.fun);
            normalizeAstNodes(node.args);
            defineGoField(node, "Fun", node.fun);
            defineGoField(node, "Lparen", spanPos(node));
            defineGoField(node, "Args", node.args);
            defineGoField(node, "Ellipsis", node.ellipsis ? spanEnd(node) : 0);
            defineGoField(node, "Rparen", closingSpanPos(node));
            break;
        case "StarExpr":
            normalizeAstNode(node.expr);
            defineGoField(node, "Star", spanPos(node));
            defineGoField(node, "X", node.expr);
            break;
        case "UnaryExpr":
            normalizeAstNode(node.expr);
            defineGoField(node, "OpPos", spanPos(node));
            defineGoField(node, "Op", goToken(node.op));
            defineGoField(node, "X", node.expr);
            break;
        case "BinaryExpr":
            normalizeAstNode(node.left);
            normalizeAstNode(node.right);
            defineGoField(node, "X", node.left);
            defineGoField(node, "OpPos", spanPos(node));
            defineGoField(node, "Op", goToken(node.op));
            defineGoField(node, "Y", node.right);
            break;
        case "KeyValueExpr":
            normalizeAstNode(node.key);
            normalizeAstNode(node.value);
            defineGoField(node, "Key", node.key);
            defineGoField(node, "Colon", spanPos(node));
            defineGoField(node, "Value", node.value);
            break;
        case "CellRefExpr":
            normalizeAstNode(node.namespace);
            break;
        case "RangeRefExpr":
            normalizeAstNode(node.namespace);
            break;
        case "ArrayType":
            normalizeAstNode(node.length);
            normalizeAstNode(node.element);
            defineGoField(node, "Lbrack", spanPos(node));
            defineGoField(node, "Len", node.length);
            defineGoField(node, "Elt", node.element);
            break;
        case "StructType":
            normalizeAstNode(node.fields);
            defineGoField(node, "Struct", spanPos(node));
            defineGoField(node, "Fields", node.fields);
            defineGoField(node, "Incomplete", false);
            break;
        case "FuncType":
            normalizeAstNode(node.typeParams);
            normalizeAstNode(node.params);
            normalizeAstNode(node.results);
            defineGoField(node, "Func", spanPos(node));
            defineGoField(node, "TypeParams", node.typeParams);
            defineGoField(node, "Params", node.params);
            defineGoField(node, "Results", node.results);
            break;
        case "InterfaceType":
            normalizeAstNode(node.methods);
            defineGoField(node, "Interface", spanPos(node));
            defineGoField(node, "Methods", node.methods);
            defineGoField(node, "Incomplete", false);
            break;
        case "MapType":
            normalizeAstNode(node.key);
            normalizeAstNode(node.value);
            defineGoField(node, "Map", spanPos(node));
            defineGoField(node, "Key", node.key);
            defineGoField(node, "Value", node.value);
            break;
        case "ChanType":
            normalizeAstNode(node.value);
            defineGoField(node, "Begin", spanPos(node));
            defineGoField(node, "Arrow", spanPos(node));
            defineGoField(node, "Dir", chanDir(node.direction));
            defineGoField(node, "Value", node.value);
            break;
        case "BadStmt":
            defineGoField(node, "From", spanPos(node));
            defineGoField(node, "To", spanEnd(node));
            break;
        case "DeclStmt":
            normalizeAstNode(node.decl);
            defineGoField(node, "Decl", node.decl);
            break;
        case "EmptyStmt":
            defineGoField(node, "Semicolon", spanPos(node));
            defineGoField(node, "Implicit", node.implicit);
            break;
        case "LabeledStmt":
            normalizeAstNode(node.label);
            normalizeAstNode(node.stmt);
            defineGoField(node, "Label", node.label);
            defineGoField(node, "Colon", spanPos(node));
            defineGoField(node, "Stmt", node.stmt);
            break;
        case "ExprStmt":
            normalizeAstNode(node.expr);
            defineGoField(node, "X", node.expr);
            break;
        case "SendStmt":
            normalizeAstNode(node.channel);
            normalizeAstNode(node.value);
            defineGoField(node, "Chan", node.channel);
            defineGoField(node, "Arrow", spanPos(node));
            defineGoField(node, "Value", node.value);
            break;
        case "IncDecStmt":
            normalizeAstNode(node.expr);
            defineGoField(node, "X", node.expr);
            defineGoField(node, "TokPos", spanPos(node));
            defineGoField(node, "Tok", goToken(node.token));
            break;
        case "AssignStmt":
            normalizeAstNodes(node.lhs);
            normalizeAstNodes(node.rhs);
            defineGoField(node, "Lhs", node.lhs);
            defineGoField(node, "TokPos", spanPos(node));
            defineGoField(node, "Tok", goToken(node.token));
            defineGoField(node, "Rhs", node.rhs);
            break;
        case "GoStmt":
            normalizeAstNode(node.call);
            defineGoField(node, "Go", spanPos(node));
            defineGoField(node, "Call", node.call);
            break;
        case "DeferStmt":
            normalizeAstNode(node.call);
            defineGoField(node, "Defer", spanPos(node));
            defineGoField(node, "Call", node.call);
            break;
        case "ReturnStmt":
            normalizeAstNodes(node.results);
            defineGoField(node, "Return", spanPos(node));
            defineGoField(node, "Results", node.results);
            break;
        case "BranchStmt":
            normalizeAstNode(node.label);
            defineGoField(node, "TokPos", spanPos(node));
            defineGoField(node, "Tok", goToken(node.token));
            defineGoField(node, "Label", node.label);
            break;
        case "BlockStmt":
            normalizeAstNodes(node.statements);
            defineGoField(node, "Lbrace", spanPos(node));
            defineGoField(node, "List", node.statements);
            defineGoField(node, "Rbrace", closingSpanPos(node));
            break;
        case "IfStmt":
            normalizeAstNode(node.init);
            normalizeAstNode(node.condition);
            normalizeAstNode(node.body);
            normalizeAstNode(node.else);
            defineGoField(node, "If", spanPos(node));
            defineGoField(node, "Init", node.init);
            defineGoField(node, "Cond", node.condition);
            defineGoField(node, "Body", node.body);
            defineGoField(node, "Else", node.else);
            break;
        case "CaseClause":
            normalizeAstNodes(node.list);
            normalizeAstNodes(node.body);
            defineGoField(node, "Case", spanPos(node));
            defineGoField(node, "List", node.default ? undefined : node.list);
            defineGoField(node, "Colon", spanPos(node));
            defineGoField(node, "Body", node.body);
            break;
        case "CommClause":
            normalizeAstNode(node.comm);
            normalizeAstNodes(node.body);
            defineGoField(node, "Case", spanPos(node));
            defineGoField(node, "Comm", node.default ? undefined : node.comm);
            defineGoField(node, "Colon", spanPos(node));
            defineGoField(node, "Body", node.body);
            break;
        case "SwitchStmt":
            normalizeAstNode(node.init);
            normalizeAstNode(node.tag);
            normalizeAstNodes(node.body);
            defineGoField(node, "Switch", spanPos(node));
            defineGoField(node, "Init", node.init);
            defineGoField(node, "Tag", node.tag);
            defineGoField(node, "Body", { kind: "BlockStmt", statements: node.body, span: node.span });
            normalizeAstNode(node.Body);
            break;
        case "TypeSwitchStmt":
            normalizeAstNode(node.init);
            normalizeAstNode(node.assign);
            normalizeAstNodes(node.body);
            defineGoField(node, "Switch", spanPos(node));
            defineGoField(node, "Init", node.init);
            defineGoField(node, "Assign", node.assign);
            defineGoField(node, "Body", { kind: "BlockStmt", statements: node.body, span: node.span });
            normalizeAstNode(node.Body);
            break;
        case "SelectStmt":
            normalizeAstNodes(node.body);
            defineGoField(node, "Select", spanPos(node));
            defineGoField(node, "Body", { kind: "BlockStmt", statements: node.body, span: node.span });
            normalizeAstNode(node.Body);
            break;
        case "ForStmt":
            normalizeAstNode(node.init);
            normalizeAstNode(node.condition);
            normalizeAstNode(node.post);
            normalizeAstNode(node.body);
            defineGoField(node, "For", spanPos(node));
            defineGoField(node, "Init", node.init);
            defineGoField(node, "Cond", node.condition);
            defineGoField(node, "Post", node.post);
            defineGoField(node, "Body", node.body);
            break;
        case "RangeStmt":
            normalizeAstNode(node.key);
            normalizeAstNode(node.value);
            normalizeAstNode(node.source);
            normalizeAstNode(node.body);
            defineGoField(node, "For", spanPos(node));
            defineGoField(node, "Key", node.key);
            defineGoField(node, "Value", node.value);
            defineGoField(node, "TokPos", spanPos(node));
            defineGoField(node, "Tok", goToken(node.token));
            defineGoField(node, "Range", spanPos(node.source));
            defineGoField(node, "X", node.source);
            defineGoField(node, "Body", node.body);
            break;
        case "UnsupportedStmt":
            break;
        case "Field":
            normalizeAstNode(node.doc);
            normalizeAstNodes(node.names);
            normalizeAstNode(node.type);
            normalizeAstNode(node.tag);
            normalizeAstNode(node.comment);
            defineGoField(node, "Doc", node.doc);
            defineGoField(node, "Names", node.names);
            defineGoField(node, "Type", node.type);
            defineGoField(node, "Tag", node.tag);
            defineGoField(node, "Comment", node.comment);
            break;
        case "FieldList":
            normalizeAstNodes(node.fields);
            defineGoField(node, "Opening", spanPos(node));
            defineGoField(node, "List", node.fields);
            defineGoField(node, "Closing", closingSpanPos(node));
            break;
        case "ImportSpec":
            normalizeAstNode(node.doc);
            normalizeAstNode(node.name);
            normalizeAstNode(node.path);
            normalizeAstNode(node.comment);
            defineGoField(node, "Doc", node.doc);
            defineGoField(node, "Name", node.name);
            defineGoField(node, "Path", node.path);
            defineGoField(node, "EndPos", spanEnd(node));
            defineGoField(node, "Comment", node.comment);
            break;
        case "ValueSpec":
            normalizeAstNode(node.doc);
            normalizeAstNodes(node.names);
            normalizeAstNode(node.type);
            normalizeAstNodes(node.values);
            normalizeAstNode(node.comment);
            defineGoField(node, "Doc", node.doc);
            defineGoField(node, "Names", node.names);
            defineGoField(node, "Type", node.type);
            defineGoField(node, "Values", node.values);
            defineGoField(node, "Comment", node.comment);
            break;
        case "TypeSpec":
            normalizeAstNode(node.doc);
            normalizeAstNode(node.name);
            normalizeAstNode(node.typeParams);
            normalizeAstNode(node.type);
            normalizeAstNode(node.comment);
            defineGoField(node, "Doc", node.doc);
            defineGoField(node, "Name", node.name);
            defineGoField(node, "TypeParams", node.typeParams);
            defineGoField(node, "Assign", node.alias ? spanPos(node) : 0);
            defineGoField(node, "Type", node.type);
            defineGoField(node, "Comment", node.comment);
            break;
        case "BadDecl":
            defineGoField(node, "From", spanPos(node));
            defineGoField(node, "To", spanEnd(node));
            break;
        case "GenDecl":
            normalizeAstNode(node.doc);
            normalizeAstNodes(node.specs);
            defineGoField(node, "Doc", node.doc);
            defineGoField(node, "TokPos", spanPos(node));
            defineGoField(node, "Tok", goToken(node.token));
            defineGoField(node, "Lparen", node.grouped ? spanPos(node) : 0);
            defineGoField(node, "Specs", node.specs);
            defineGoField(node, "Rparen", node.grouped ? closingSpanPos(node) : 0);
            break;
        case "FuncDecl":
            normalizeAstNode(node.doc);
            normalizeAstNode(node.receiver);
            normalizeAstNode(node.name);
            normalizeAstNode(node.type);
            normalizeAstNode(node.body);
            defineGoField(node, "Doc", node.doc);
            defineGoField(node, "Recv", node.receiver);
            defineGoField(node, "Name", node.name);
            defineGoField(node, "Type", node.type);
            defineGoField(node, "Body", node.body);
            break;
        case "File":
            normalizeAstNode(node.doc);
            normalizeAstNode(node.name);
            normalizeAstNodes(node.declarations);
            normalizeAstNodes(node.imports);
            normalizeAstNodes(node.unresolved);
            normalizeAstNodes(node.comments);
            defineGoField(node, "Doc", node.doc);
            defineGoField(node, "Package", spanPos(node.name));
            defineGoField(node, "Name", node.name);
            defineGoField(node, "Decls", node.declarations);
            defineGoField(node, "FileStart", spanPos(node));
            defineGoField(node, "FileEnd", spanEnd(node));
            defineGoField(node, "Scope", node.scope);
            defineGoField(node, "Imports", node.imports);
            defineGoField(node, "Unresolved", node.unresolved);
            defineGoField(node, "Comments", node.comments);
            defineGoField(node, "GoVersion", "");
            break;
        case "Package":
            normalizeAstNodes(fileValues(node.files));
            defineGoField(node, "Name", node.name);
            defineGoField(node, "Scope", node.scope);
            defineGoField(node, "Imports", node.imports);
            defineGoField(node, "Files", node.files);
            break;
        case "Comment":
            defineGoField(node, "Slash", spanPos(node));
            defineGoField(node, "Text", node.text);
            break;
        case "CommentGroup":
            normalizeAstNodes(node.list);
            defineGoField(node, "List", node.list);
            break;
    }
}
function astField(node, standard, compat) {
    const object = node;
    return (object[standard] ?? (compat ? object[compat] : undefined));
}
function astList(node, standard, compat) {
    const object = node;
    const value = object[standard] ?? (compat ? object[compat] : undefined);
    return Array.isArray(value) ? value : [];
}
function astMaybe(value) {
    return value ? [value] : [];
}
export function childNodes(node) {
    switch (node.kind) {
        case "File":
            return [
                ...astMaybe(astField(node, "Doc", "doc")),
                ...astMaybe(astField(node, "Name", "name")),
                ...astList(node, "Decls", "declarations")
            ];
        case "Package":
            return fileValues((node.Files ?? node.files));
        case "GenDecl":
            return [
                ...astMaybe(astField(node, "Doc", "doc")),
                ...astList(node, "Specs", "specs")
            ];
        case "FuncDecl":
            return [
                ...astMaybe(astField(node, "Doc", "doc")),
                ...astMaybe(astField(node, "Recv", "receiver")),
                ...astMaybe(astField(node, "Name", "name")),
                ...astMaybe(astField(node, "Type", "type")),
                ...astMaybe(astField(node, "Body", "body"))
            ];
        case "FieldList":
            return astList(node, "List", "fields");
        case "Field":
            return [
                ...astMaybe(astField(node, "Doc", "doc")),
                ...astList(node, "Names", "names"),
                ...astMaybe(astField(node, "Type", "type")),
                ...astMaybe(astField(node, "Tag", "tag")),
                ...astMaybe(astField(node, "Comment", "comment"))
            ];
        case "ImportSpec":
            return [
                ...astMaybe(astField(node, "Doc", "doc")),
                ...astMaybe(astField(node, "Name", "name")),
                ...astMaybe(astField(node, "Path", "path")),
                ...astMaybe(astField(node, "Comment", "comment"))
            ];
        case "ValueSpec":
            return [
                ...astMaybe(astField(node, "Doc", "doc")),
                ...astList(node, "Names", "names"),
                ...astMaybe(astField(node, "Type", "type")),
                ...astList(node, "Values", "values"),
                ...astMaybe(astField(node, "Comment", "comment"))
            ];
        case "TypeSpec":
            return [
                ...astMaybe(astField(node, "Doc", "doc")),
                ...astMaybe(astField(node, "Name", "name")),
                ...astMaybe(astField(node, "TypeParams", "typeParams")),
                ...astMaybe(astField(node, "Type", "type")),
                ...astMaybe(astField(node, "Comment", "comment"))
            ];
        case "FuncType":
            return [
                ...astMaybe(astField(node, "TypeParams", "typeParams")),
                ...astMaybe(astField(node, "Params", "params")),
                ...astMaybe(astField(node, "Results", "results"))
            ];
        case "BlockStmt":
            return astList(node, "List", "statements");
        case "DeclStmt":
            return astMaybe(astField(node, "Decl", "decl"));
        case "LabeledStmt":
            return [
                ...astMaybe(astField(node, "Label", "label")),
                ...astMaybe(astField(node, "Stmt", "stmt"))
            ];
        case "ExprStmt":
            return astMaybe(astField(node, "X", "expr"));
        case "AssignStmt":
            return [...astList(node, "Lhs", "lhs"), ...astList(node, "Rhs", "rhs")];
        case "IncDecStmt":
            return astMaybe(astField(node, "X", "expr"));
        case "ReturnStmt":
            return astList(node, "Results", "results");
        case "BranchStmt":
            return astMaybe(astField(node, "Label", "label"));
        case "IfStmt":
            return [
                ...astMaybe(astField(node, "Init", "init")),
                ...astMaybe(astField(node, "Cond", "condition")),
                ...astMaybe(astField(node, "Body", "body")),
                ...astMaybe(astField(node, "Else", "else"))
            ];
        case "CaseClause":
            return [...astList(node, "List", "list"), ...astList(node, "Body", "body")];
        case "CommClause":
            return [...astMaybe(astField(node, "Comm", "comm")), ...astList(node, "Body", "body")];
        case "SwitchStmt":
            return [
                ...astMaybe(astField(node, "Init", "init")),
                ...astMaybe(astField(node, "Tag", "tag")),
                ...childNodes(astField(node, "Body") ?? { kind: "BlockStmt", statements: node.body })
            ];
        case "TypeSwitchStmt":
            return [
                ...astMaybe(astField(node, "Init", "init")),
                ...astMaybe(astField(node, "Assign", "assign")),
                ...childNodes(astField(node, "Body") ?? { kind: "BlockStmt", statements: node.body })
            ];
        case "SelectStmt":
            return childNodes(astField(node, "Body") ?? { kind: "BlockStmt", statements: node.body });
        case "ForStmt":
            return [
                ...astMaybe(astField(node, "Init", "init")),
                ...astMaybe(astField(node, "Cond", "condition")),
                ...astMaybe(astField(node, "Post", "post")),
                ...astMaybe(astField(node, "Body", "body"))
            ];
        case "RangeStmt":
            return [
                ...astMaybe(astField(node, "Key", "key")),
                ...astMaybe(astField(node, "Value", "value")),
                ...astMaybe(astField(node, "X", "source")),
                ...astMaybe(astField(node, "Body", "body"))
            ];
        case "DeferStmt":
            return astMaybe(astField(node, "Call", "call"));
        case "GoStmt":
            return astMaybe(astField(node, "Call", "call"));
        case "SendStmt":
            return [
                ...astMaybe(astField(node, "Chan", "channel")),
                ...astMaybe(astField(node, "Value", "value"))
            ];
        case "Ellipsis":
            return astMaybe(astField(node, "Elt", "element"));
        case "FuncLit":
            return [
                ...astMaybe(astField(node, "Type", "type")),
                ...astMaybe(astField(node, "Body", "body"))
            ];
        case "CompositeLit":
            return [
                ...astMaybe(astField(node, "Type", "type")),
                ...astList(node, "Elts", "elements")
            ];
        case "ParenExpr":
            return astMaybe(astField(node, "X", "expr"));
        case "SelectorExpr":
            return [
                ...astMaybe(astField(node, "X", "object")),
                ...astMaybe(astField(node, "Sel", "selector"))
            ];
        case "IndexExpr":
            return [
                ...astMaybe(astField(node, "X", "object")),
                ...astMaybe(astField(node, "Index", "index"))
            ];
        case "IndexListExpr":
            return [
                ...astMaybe(astField(node, "X", "object")),
                ...astList(node, "Indices", "indices")
            ];
        case "SliceExpr":
            return [
                ...astMaybe(astField(node, "X", "object")),
                ...astMaybe(astField(node, "Low", "low")),
                ...astMaybe(astField(node, "High", "high")),
                ...astMaybe(astField(node, "Max", "max"))
            ];
        case "TypeAssertExpr":
            return [
                ...astMaybe(astField(node, "X", "object")),
                ...astMaybe(astField(node, "Type", "type"))
            ];
        case "CallExpr":
            return [
                ...astMaybe(astField(node, "Fun", "fun")),
                ...astList(node, "Args", "args")
            ];
        case "StarExpr":
            return astMaybe(astField(node, "X", "expr"));
        case "UnaryExpr":
            return astMaybe(astField(node, "X", "expr"));
        case "BinaryExpr":
            return [
                ...astMaybe(astField(node, "X", "left")),
                ...astMaybe(astField(node, "Y", "right"))
            ];
        case "KeyValueExpr":
            return [
                ...astMaybe(astField(node, "Key", "key")),
                ...astMaybe(astField(node, "Value", "value"))
            ];
        case "CellRefExpr":
            return [node.namespace];
        case "RangeRefExpr":
            return [node.namespace];
        case "ArrayType":
            return [
                ...astMaybe(astField(node, "Len", "length")),
                ...astMaybe(astField(node, "Elt", "element"))
            ];
        case "StructType":
            return astMaybe(astField(node, "Fields", "fields"));
        case "InterfaceType":
            return astMaybe(astField(node, "Methods", "methods"));
        case "MapType":
            return [
                ...astMaybe(astField(node, "Key", "key")),
                ...astMaybe(astField(node, "Value", "value"))
            ];
        case "ChanType":
            return astMaybe(astField(node, "Value", "value"));
        default:
            return [];
    }
}
