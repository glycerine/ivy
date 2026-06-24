import { TokenKind } from "./token.js";
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
        return PosOf(group?.list[0]);
    }
    CommentGroup.Pos = Pos;
    function End(group) {
        return EndOf(group?.list[group.list.length - 1]);
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
    return ident(name);
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
        return nodePos(id) + (id?.name.length ?? 0);
    }
    Ident.End = End;
    function exprNode(_id) {
    }
    Ident.exprNode = exprNode;
    function IsExported(id) {
        return id ? astIsExported(id.name) : false;
    }
    Ident.IsExported = IsExported;
    function String(id) {
        if (id) {
            return id.name;
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
    function End(x) { return x.element ? EndOf(x.element) : nodeEnd(x) || nodePos(x) + 3; }
    Ellipsis.End = End;
    function exprNode(_x) { }
    Ellipsis.exprNode = exprNode;
})(Ellipsis || (Ellipsis = {}));
export var BasicLit;
(function (BasicLit) {
    function Pos(x) { return nodePos(x); }
    BasicLit.Pos = Pos;
    function End(x) { return nodeEnd(x) || nodePos(x) + x.value.length; }
    BasicLit.End = End;
    function exprNode(_x) { }
    BasicLit.exprNode = exprNode;
})(BasicLit || (BasicLit = {}));
export var FuncLit;
(function (FuncLit) {
    function Pos(x) { return PosOf(x.type); }
    FuncLit.Pos = Pos;
    function End(x) { return EndOf(x.body); }
    FuncLit.End = End;
    function exprNode(_x) { }
    FuncLit.exprNode = exprNode;
})(FuncLit || (FuncLit = {}));
export var CompositeLit;
(function (CompositeLit) {
    function Pos(x) { return x.type ? PosOf(x.type) : nodePos(x); }
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
    function Pos(x) { return PosOf(x.object); }
    SelectorExpr.Pos = Pos;
    function End(x) { return EndOf(x.selector); }
    SelectorExpr.End = End;
    function exprNode(_x) { }
    SelectorExpr.exprNode = exprNode;
})(SelectorExpr || (SelectorExpr = {}));
export var IndexExpr;
(function (IndexExpr) {
    function Pos(x) { return PosOf(x.object); }
    IndexExpr.Pos = Pos;
    function End(x) { return nodeEnd(x); }
    IndexExpr.End = End;
    function exprNode(_x) { }
    IndexExpr.exprNode = exprNode;
})(IndexExpr || (IndexExpr = {}));
export var IndexListExpr;
(function (IndexListExpr) {
    function Pos(x) { return PosOf(x.object); }
    IndexListExpr.Pos = Pos;
    function End(x) { return nodeEnd(x); }
    IndexListExpr.End = End;
    function exprNode(_x) { }
    IndexListExpr.exprNode = exprNode;
})(IndexListExpr || (IndexListExpr = {}));
export var SliceExpr;
(function (SliceExpr) {
    function Pos(x) { return PosOf(x.object); }
    SliceExpr.Pos = Pos;
    function End(x) { return nodeEnd(x); }
    SliceExpr.End = End;
    function exprNode(_x) { }
    SliceExpr.exprNode = exprNode;
})(SliceExpr || (SliceExpr = {}));
export var TypeAssertExpr;
(function (TypeAssertExpr) {
    function Pos(x) { return PosOf(x.object); }
    TypeAssertExpr.Pos = Pos;
    function End(x) { return nodeEnd(x); }
    TypeAssertExpr.End = End;
    function exprNode(_x) { }
    TypeAssertExpr.exprNode = exprNode;
})(TypeAssertExpr || (TypeAssertExpr = {}));
export var CallExpr;
(function (CallExpr) {
    function Pos(x) { return PosOf(x.fun); }
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
    function End(x) { return EndOf(x.expr); }
    StarExpr.End = End;
    function exprNode(_x) { }
    StarExpr.exprNode = exprNode;
})(StarExpr || (StarExpr = {}));
export var UnaryExpr;
(function (UnaryExpr) {
    function Pos(x) { return nodePos(x); }
    UnaryExpr.Pos = Pos;
    function End(x) { return EndOf(x.expr); }
    UnaryExpr.End = End;
    function exprNode(_x) { }
    UnaryExpr.exprNode = exprNode;
})(UnaryExpr || (UnaryExpr = {}));
export var BinaryExpr;
(function (BinaryExpr) {
    function Pos(x) { return PosOf(x.left); }
    BinaryExpr.Pos = Pos;
    function End(x) { return EndOf(x.right); }
    BinaryExpr.End = End;
    function exprNode(_x) { }
    BinaryExpr.exprNode = exprNode;
})(BinaryExpr || (BinaryExpr = {}));
export var KeyValueExpr;
(function (KeyValueExpr) {
    function Pos(x) { return PosOf(x.key); }
    KeyValueExpr.Pos = Pos;
    function End(x) { return EndOf(x.value); }
    KeyValueExpr.End = End;
    function exprNode(_x) { }
    KeyValueExpr.exprNode = exprNode;
})(KeyValueExpr || (KeyValueExpr = {}));
export var ArrayType;
(function (ArrayType) {
    function Pos(x) { return nodePos(x); }
    ArrayType.Pos = Pos;
    function End(x) { return EndOf(x.element); }
    ArrayType.End = End;
    function exprNode(_x) { }
    ArrayType.exprNode = exprNode;
})(ArrayType || (ArrayType = {}));
export var StructType;
(function (StructType) {
    function Pos(x) { return nodePos(x); }
    StructType.Pos = Pos;
    function End(x) { return FieldList.End(x.fields); }
    StructType.End = End;
    function exprNode(_x) { }
    StructType.exprNode = exprNode;
})(StructType || (StructType = {}));
export var FuncType;
(function (FuncType) {
    function Pos(x) { return nodePos(x) || FieldList.Pos(x.params); }
    FuncType.Pos = Pos;
    function End(x) { return x.results ? FieldList.End(x.results) : FieldList.End(x.params); }
    FuncType.End = End;
    function exprNode(_x) { }
    FuncType.exprNode = exprNode;
})(FuncType || (FuncType = {}));
export var InterfaceType;
(function (InterfaceType) {
    function Pos(x) { return nodePos(x); }
    InterfaceType.Pos = Pos;
    function End(x) { return FieldList.End(x.methods); }
    InterfaceType.End = End;
    function exprNode(_x) { }
    InterfaceType.exprNode = exprNode;
})(InterfaceType || (InterfaceType = {}));
export var MapType;
(function (MapType) {
    function Pos(x) { return nodePos(x); }
    MapType.Pos = Pos;
    function End(x) { return EndOf(x.value); }
    MapType.End = End;
    function exprNode(_x) { }
    MapType.exprNode = exprNode;
})(MapType || (MapType = {}));
export var ChanType;
(function (ChanType) {
    function Pos(x) { return nodePos(x); }
    ChanType.Pos = Pos;
    function End(x) { return EndOf(x.value); }
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
    function Pos(x) { return PosOf(x.decl); }
    DeclStmt.Pos = Pos;
    function End(x) { return EndOf(x.decl); }
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
    function Pos(x) { return PosOf(x.label); }
    LabeledStmt.Pos = Pos;
    function End(x) { return EndOf(x.stmt); }
    LabeledStmt.End = End;
    function stmtNode(_x) { }
    LabeledStmt.stmtNode = stmtNode;
})(LabeledStmt || (LabeledStmt = {}));
export var ExprStmt;
(function (ExprStmt) {
    function Pos(x) { return PosOf(x.expr); }
    ExprStmt.Pos = Pos;
    function End(x) { return EndOf(x.expr); }
    ExprStmt.End = End;
    function stmtNode(_x) { }
    ExprStmt.stmtNode = stmtNode;
})(ExprStmt || (ExprStmt = {}));
export var SendStmt;
(function (SendStmt) {
    function Pos(x) { return PosOf(x.channel); }
    SendStmt.Pos = Pos;
    function End(x) { return EndOf(x.value); }
    SendStmt.End = End;
    function stmtNode(_x) { }
    SendStmt.stmtNode = stmtNode;
})(SendStmt || (SendStmt = {}));
export var IncDecStmt;
(function (IncDecStmt) {
    function Pos(x) { return PosOf(x.expr); }
    IncDecStmt.Pos = Pos;
    function End(x) { return nodeEnd(x) || EndOf(x.expr) + 2; }
    IncDecStmt.End = End;
    function stmtNode(_x) { }
    IncDecStmt.stmtNode = stmtNode;
})(IncDecStmt || (IncDecStmt = {}));
export var AssignStmt;
(function (AssignStmt) {
    function Pos(x) { return PosOf(x.lhs[0]); }
    AssignStmt.Pos = Pos;
    function End(x) { return EndOf(x.rhs[x.rhs.length - 1]); }
    AssignStmt.End = End;
    function stmtNode(_x) { }
    AssignStmt.stmtNode = stmtNode;
})(AssignStmt || (AssignStmt = {}));
export var GoStmt;
(function (GoStmt) {
    function Pos(x) { return nodePos(x); }
    GoStmt.Pos = Pos;
    function End(x) { return EndOf(x.call); }
    GoStmt.End = End;
    function stmtNode(_x) { }
    GoStmt.stmtNode = stmtNode;
})(GoStmt || (GoStmt = {}));
export var DeferStmt;
(function (DeferStmt) {
    function Pos(x) { return nodePos(x); }
    DeferStmt.Pos = Pos;
    function End(x) { return EndOf(x.call); }
    DeferStmt.End = End;
    function stmtNode(_x) { }
    DeferStmt.stmtNode = stmtNode;
})(DeferStmt || (DeferStmt = {}));
export var ReturnStmt;
(function (ReturnStmt) {
    function Pos(x) { return nodePos(x); }
    ReturnStmt.Pos = Pos;
    function End(x) { return x.results.length > 0 ? EndOf(x.results[x.results.length - 1]) : nodeEnd(x) || nodePos(x) + 6; }
    ReturnStmt.End = End;
    function stmtNode(_x) { }
    ReturnStmt.stmtNode = stmtNode;
})(ReturnStmt || (ReturnStmt = {}));
export var BranchStmt;
(function (BranchStmt) {
    function Pos(x) { return nodePos(x); }
    BranchStmt.Pos = Pos;
    function End(x) { return x.label ? EndOf(x.label) : nodeEnd(x); }
    BranchStmt.End = End;
    function stmtNode(_x) { }
    BranchStmt.stmtNode = stmtNode;
})(BranchStmt || (BranchStmt = {}));
export var BlockStmt;
(function (BlockStmt) {
    function Pos(x) { return nodePos(x); }
    BlockStmt.Pos = Pos;
    function End(x) { return nodeEnd(x) || (x.statements.length > 0 ? EndOf(x.statements[x.statements.length - 1]) : nodePos(x) + 1); }
    BlockStmt.End = End;
    function stmtNode(_x) { }
    BlockStmt.stmtNode = stmtNode;
})(BlockStmt || (BlockStmt = {}));
export var IfStmt;
(function (IfStmt) {
    function Pos(x) { return nodePos(x); }
    IfStmt.Pos = Pos;
    function End(x) { return x.else ? EndOf(x.else) : EndOf(x.body); }
    IfStmt.End = End;
    function stmtNode(_x) { }
    IfStmt.stmtNode = stmtNode;
})(IfStmt || (IfStmt = {}));
export var CaseClause;
(function (CaseClause) {
    function Pos(x) { return nodePos(x); }
    CaseClause.Pos = Pos;
    function End(x) { return x.body.length > 0 ? EndOf(x.body[x.body.length - 1]) : nodeEnd(x); }
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
    function End(x) { return x.body.length > 0 ? EndOf(x.body[x.body.length - 1]) : nodeEnd(x); }
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
    function End(x) { return EndOf(x.body); }
    ForStmt.End = End;
    function stmtNode(_x) { }
    ForStmt.stmtNode = stmtNode;
})(ForStmt || (ForStmt = {}));
export var RangeStmt;
(function (RangeStmt) {
    function Pos(x) { return nodePos(x); }
    RangeStmt.Pos = Pos;
    function End(x) { return EndOf(x.body); }
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
    function Pos(x) { return x.names.length > 0 ? PosOf(x.names[0]) : PosOf(x.type); }
    Field.Pos = Pos;
    function End(x) { return x.tag ? EndOf(x.tag) : EndOf(x.type); }
    Field.End = End;
})(Field || (Field = {}));
export var FieldList;
(function (FieldList) {
    function Pos(x) { return nodePos(x) || PosOf(x?.fields[0]); }
    FieldList.Pos = Pos;
    function End(x) { return x ? nodeEnd(x) || EndOf(x.fields[x.fields.length - 1]) : 0; }
    FieldList.End = End;
    function NumFields(x) {
        if (!x)
            return 0;
        let n = 0;
        for (const field of x.fields) {
            n += field.names.length || 1;
        }
        return n;
    }
    FieldList.NumFields = NumFields;
})(FieldList || (FieldList = {}));
export var ImportSpec;
(function (ImportSpec) {
    function Pos(x) { return x.name ? PosOf(x.name) : PosOf(x.path); }
    ImportSpec.Pos = Pos;
    function End(x) { return x.comment ? CommentGroup.End(x.comment) : EndOf(x.path); }
    ImportSpec.End = End;
    function specNode(_x) { }
    ImportSpec.specNode = specNode;
})(ImportSpec || (ImportSpec = {}));
export var ValueSpec;
(function (ValueSpec) {
    function Pos(x) { return PosOf(x.names[0]); }
    ValueSpec.Pos = Pos;
    function End(x) { return x.values.length > 0 ? EndOf(x.values[x.values.length - 1]) : x.type ? EndOf(x.type) : EndOf(x.names[x.names.length - 1]); }
    ValueSpec.End = End;
    function specNode(_x) { }
    ValueSpec.specNode = specNode;
})(ValueSpec || (ValueSpec = {}));
export var TypeSpec;
(function (TypeSpec) {
    function Pos(x) { return PosOf(x.name); }
    TypeSpec.Pos = Pos;
    function End(x) { return EndOf(x.type); }
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
    function End(x) { return nodeEnd(x) || EndOf(x.specs[x.specs.length - 1]); }
    GenDecl.End = End;
    function declNode(_x) { }
    GenDecl.declNode = declNode;
})(GenDecl || (GenDecl = {}));
export var FuncDecl;
(function (FuncDecl) {
    function Pos(x) { return nodePos(x) || PosOf(x.name); }
    FuncDecl.Pos = Pos;
    function End(x) { return x.body ? EndOf(x.body) : EndOf(x.type); }
    FuncDecl.End = End;
    function declNode(_x) { }
    FuncDecl.declNode = declNode;
})(FuncDecl || (FuncDecl = {}));
export var File;
(function (File) {
    function Pos(x) { return nodePos(x) || PosOf(x.name); }
    File.Pos = Pos;
    function End(x) { return nodeEnd(x) || EndOf(x.declarations[x.declarations.length - 1]); }
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
    const packageOffset = file.name?.span?.offset ?? Number.POSITIVE_INFINITY;
    for (const group of file.comments) {
        for (const comment of group.list) {
            if ((comment.span?.offset ?? 0) > packageOffset) {
                break;
            }
            const prefix = "// Code generated ";
            if (comment.text.includes(prefix)) {
                for (const line of comment.text.split("\n")) {
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
        expr = expr.expr;
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
        if (f(x.name)) {
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
            if (x.object.kind === "Ident") {
                return x.selector;
            }
            break;
        case "StarExpr":
            return fieldName(x.expr);
    }
    return undefined;
}
export function filterFieldList(fields, filter, exportOnly) {
    if (!fields) {
        return false;
    }
    const list = fields.fields;
    let j = 0;
    let removedFields = false;
    for (const field of list) {
        let keepField = false;
        if (field.names.length === 0) {
            const name = fieldName(field.type);
            keepField = name !== undefined && filter(name.name);
        }
        else {
            const n = field.names.length;
            field.names = filterIdentList(field.names, filter);
            if (field.names.length < n) {
                removedFields = true;
            }
            keepField = field.names.length > 0;
        }
        if (keepField) {
            if (exportOnly) {
                filterType(field.type, filter, exportOnly);
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
    lit.elements = filterExprList(lit.elements, filter, exportOnly);
}
export function filterExprList(list, filter, exportOnly) {
    let j = 0;
    for (const exp of list) {
        switch (exp.kind) {
            case "CompositeLit":
                filterCompositeLit(exp, filter, exportOnly);
                break;
            case "KeyValueExpr":
                if (exp.key.kind === "Ident" && !filter(exp.key.name)) {
                    continue;
                }
                if (exp.value.kind === "CompositeLit") {
                    filterCompositeLit(exp.value, filter, exportOnly);
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
    for (const field of fields.fields) {
        if (filterType(field.type, filter, exportOnly)) {
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
            return f(typ.name);
        case "ParenExpr":
            return filterType(typ.expr, f, exportOnly);
        case "ArrayType":
            return filterType(typ.element, f, exportOnly);
        case "StructType":
            filterFieldList(typ.fields, f, exportOnly);
            return typ.fields.fields.length > 0;
        case "FuncType": {
            const b1 = filterParamList(typ.params, f, exportOnly);
            const b2 = filterParamList(typ.results, f, exportOnly);
            return b1 || b2;
        }
        case "InterfaceType":
            filterFieldList(typ.methods, f, exportOnly);
            return typ.methods.fields.length > 0;
        case "MapType": {
            const b1 = filterType(typ.key, f, exportOnly);
            const b2 = filterType(typ.value, f, exportOnly);
            return b1 || b2;
        }
        case "ChanType":
            return filterType(typ.value, f, exportOnly);
    }
    return false;
}
export function filterSpec(spec, f, exportOnly) {
    switch (spec.kind) {
        case "ValueSpec":
            spec.names = filterIdentList(spec.names, f);
            spec.values = filterExprList(spec.values, f, exportOnly);
            if (spec.names.length > 0) {
                if (exportOnly) {
                    filterType(spec.type, f, exportOnly);
                }
                return true;
            }
            break;
        case "TypeSpec":
            if (f(spec.name.name)) {
                if (exportOnly) {
                    filterType(spec.type, f, exportOnly);
                }
                return true;
            }
            if (!exportOnly) {
                return filterType(spec.type, f, exportOnly);
            }
            break;
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
            decl.specs = filterSpecList(decl.specs, f, exportOnly);
            return decl.specs.length > 0;
        case "FuncDecl":
            return f(decl.name.name);
    }
    return false;
}
export function FilterFile(src, f) {
    return filterFile(src, f, false);
}
export function filterFile(src, f, exportOnly) {
    let j = 0;
    for (const declaration of src.declarations) {
        if (filterDecl(declaration, f, exportOnly)) {
            src.declarations[j] = declaration;
            j += 1;
        }
    }
    src.declarations.length = j;
    return j > 0;
}
export function FilterPackage(pkg, f) {
    return filterPackage(pkg, f, false);
}
export function filterPackage(pkg, f, exportOnly) {
    let hasDecls = false;
    for (const src of pkg.files) {
        if (filterFile(src, f, exportOnly)) {
            hasDecls = true;
        }
    }
    return hasDecls;
}
export const FilterFuncDuplicates = 1 << 0;
export const FilterUnassociatedComments = 1 << 1;
export const FilterImportDuplicates = 1 << 2;
export const separator = { kind: "Comment", text: "//" };
export function nameOf(f) {
    if (f.receiver && f.receiver.fields.length === 1) {
        let typ = f.receiver.fields[0]?.type;
        if (typ?.kind === "StarExpr") {
            typ = typ.expr;
        }
        if (typ?.kind === "Ident") {
            return `${typ.name}.${f.name.name}`;
        }
    }
    return f.name.name;
}
export function MergePackageFiles(pkg, mode) {
    const files = [...pkg.files].sort((left, right) => (left.name?.name ?? "").localeCompare(right.name?.name ?? ""));
    const declarations = [];
    const comments = [];
    const funcs = new Set();
    const imports = new Set();
    for (const file of files) {
        if ((mode & FilterUnassociatedComments) === 0) {
            comments.push(...file.comments);
        }
        for (const declaration of file.declarations) {
            if ((mode & FilterFuncDuplicates) !== 0 && declaration.kind === "FuncDecl") {
                const name = nameOf(declaration);
                if (funcs.has(name)) {
                    continue;
                }
                funcs.add(name);
            }
            if ((mode & FilterImportDuplicates) !== 0 && declaration.kind === "GenDecl" && declaration.token === TokenKind.Import) {
                const before = declaration.specs.length;
                declaration.specs = declaration.specs.filter((spec) => {
                    if (spec.kind !== "ImportSpec")
                        return true;
                    const key = spec.path.value;
                    if (imports.has(key))
                        return false;
                    imports.add(key);
                    return true;
                });
                if (before > 0 && declaration.specs.length === 0) {
                    continue;
                }
            }
            declarations.push(declaration);
        }
    }
    return {
        kind: "File",
        name: ident(pkg.name),
        declarations,
        imports: declarations.flatMap((decl) => decl.kind === "GenDecl" ? decl.specs.filter((spec) => spec.kind === "ImportSpec") : []),
        unresolved: [],
        comments
    };
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
            const s = node.kind === "Ident" ? node.name : node.kind;
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
    for (const decl of f.declarations) {
        if (decl.kind !== "GenDecl" || decl.token !== TokenKind.Import) {
            break;
        }
        if (!decl.grouped) {
            continue;
        }
        let i = 0;
        const specs = [];
        for (let j = 0; j < decl.specs.length; j += 1) {
            const spec = decl.specs[j];
            const prev = decl.specs[j - 1];
            if (spec && prev && j > i && lineAt(fset, spec) > 1 + lineAt(fset, prev)) {
                specs.push(...sortSpecs(fset, f, decl, decl.specs.slice(i, j)));
                i = j;
            }
        }
        specs.push(...sortSpecs(fset, f, decl, decl.specs.slice(i)));
        decl.specs = specs;
    }
    f.imports.length = 0;
    for (const decl of f.declarations) {
        if (decl.kind === "GenDecl" && decl.token === TokenKind.Import) {
            for (const spec of decl.specs) {
                if (spec.kind === "ImportSpec") {
                    f.imports.push(spec);
                }
            }
        }
    }
}
export function lineAt(_fset, pos) {
    if (typeof pos === "number") {
        return Number.isFinite(pos) ? pos : 0;
    }
    if (pos && "kind" in pos) {
        return pos.span?.line ?? 0;
    }
    return pos?.line ?? 0;
}
export function importPath(s) {
    if (s.kind !== "ImportSpec") {
        return "";
    }
    try {
        return unquoteGoString(s.path.value);
    }
    catch {
        return "";
    }
}
export function importName(s) {
    if (s.kind !== "ImportSpec") {
        return "";
    }
    return s.name?.name ?? "";
}
export function importComment(s) {
    if (s.kind !== "ImportSpec") {
        return "";
    }
    return CommentGroup.Text(s.comment);
}
export function collapse(prev, next) {
    if (importPath(next) !== importPath(prev) || importName(next) !== importName(prev)) {
        return false;
    }
    return prev.kind === "ImportSpec" && prev.comment === undefined;
}
export function sortSpecs(fset, f, d, specs) {
    if (specs.length <= 1) {
        return specs;
    }
    const pos = new Array(specs.length);
    for (let i = 0; i < specs.length; i += 1) {
        const spec = specs[i];
        pos[i] = new posSpan(nodePos(spec), nodeEnd(spec));
    }
    const begSpecs = pos[0]?.Start ?? 0;
    const endSpecs = pos[pos.length - 1]?.End ?? 0;
    const beg = lineStart(f, begSpecs);
    const endLine = lineAt(fset, specs[specs.length - 1]);
    let end;
    if (endLine === maxLine(f)) {
        end = endSpecs;
    }
    else {
        end = lineStart(f, endLine + 1);
    }
    let first = f.comments.length;
    let last = -1;
    for (let i = 0; i < f.comments.length; i += 1) {
        const group = f.comments[i];
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
        comments = f.comments.slice(first, last + 1);
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
            lineAt(fset, specs[specIndex]) + 1 === lineAt(fset, group)) {
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
            if (lineAt(fset, spec) !== lineAt(fset, d)) {
                mergeLine(f, lineAt(fset, spec));
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
        if (spec.name?.span) {
            spec.name.span = moveSpan(spec.name.span, span.Start);
        }
        updateBasicLitPos(spec.path, span.Start);
        for (const group of importComments.get(spec) ?? []) {
            for (const comment of group.cg.list) {
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
            for (const n of decl.names) {
                if (n.name === name) {
                    return nodePos(n);
                }
            }
        }
        else if (isImportSpec(decl)) {
            if (decl.name && decl.name.name === name) {
                return nodePos(decl.name);
            }
            return nodePos(decl.path);
        }
        else if (isValueSpec(decl)) {
            for (const n of decl.names) {
                if (n.name === name) {
                    return nodePos(n);
                }
            }
        }
        else if (isTypeSpec(decl)) {
            if (decl.name.name === name) {
                return nodePos(decl.name);
            }
        }
        else if (isFuncDecl(decl)) {
            if (decl.name.name === name) {
                return nodePos(decl.name);
            }
        }
        else if (isLabeledStmt(decl)) {
            if (decl.label.name === name) {
                return nodePos(decl.label);
            }
        }
        else if (isAssignStmt(decl)) {
            for (const x of decl.lhs) {
                if (x.kind === "Ident" && x.name === name) {
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
        const obj = scope.Lookup(id.name);
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
        const name = file.name?.name ?? "";
        switch (true) {
            case pkgName === "":
                pkgName = name;
                break;
            case name !== pkgName:
                p.errorf(nodePos(file.name), "package %s; expected %s", name, pkgName);
                continue;
        }
        for (const obj of file.scope?.Objects.values() ?? []) {
            p.declare(pkgScope, undefined, obj);
        }
    }
    const imports = new Map();
    for (const file of fileValues(files)) {
        if ((file.name?.name ?? "") !== pkgName) {
            continue;
        }
        let importErrors = false;
        const fileScope = NewScope(pkgScope);
        for (const spec of file.imports) {
            if (importer === undefined) {
                importErrors = true;
                continue;
            }
            const path = importPath(spec);
            const [pkg, err] = importer(imports, path);
            if (err !== undefined || pkg === undefined) {
                p.errorf(nodePos(spec.path), "could not import %s (%s)", path, err?.message ?? "unknown error");
                importErrors = true;
                continue;
            }
            let name = pkg.Name;
            if (spec.name !== undefined) {
                name = spec.name.name;
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
        for (const id of file.unresolved) {
            if (!resolve(fileScope, id)) {
                p.errorf(nodePos(id), "undeclared name: %s", id.name);
                file.unresolved[i] = id;
                i += 1;
            }
        }
        file.unresolved.length = i;
        pkgScope.Outer = universe;
    }
    p.errors.sort((a, b) => a.message.localeCompare(b.message));
    return [{ kind: "Package", name: pkgName, scope: pkgScope, imports, files: fileValues(files) }, p.errors[0]];
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
    if (files instanceof Map) {
        return [...files.values()];
    }
    return globalThis.Object.values(files);
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
function nodePos(node) {
    return node?.span?.offset ?? 0;
}
function nodeEnd(node) {
    if (!node?.span) {
        return 0;
    }
    return node.span.offset + node.span.length;
}
function moveSpan(span, offset) {
    return {
        ...span,
        offset
    };
}
function lineStart(file, line) {
    for (const node of childNodes(file)) {
        const span = node.span;
        if (span && span.line === line) {
            return span.offset - Math.max(0, span.column - 1);
        }
    }
    return line;
}
function maxLine(file) {
    let line = 0;
    walk(file, (node) => {
        if (node.span) {
            line = Math.max(line, node.span.line);
        }
    });
    return line;
}
function mergeLine(_file, _line) {
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
export function childNodes(node) {
    switch (node.kind) {
        case "File":
            return [...(node.name ? [node.name] : []), ...node.declarations, ...node.imports, ...node.unresolved, ...node.comments];
        case "Package":
            return node.files;
        case "GenDecl":
            return node.specs;
        case "FuncDecl":
            return [...(node.receiver ? [node.receiver] : []), node.name, node.type, ...(node.body ? [node.body] : [])];
        case "FieldList":
            return node.fields;
        case "Field":
            return [...node.names, node.type, ...(node.tag ? [node.tag] : [])];
        case "ImportSpec":
            return [...(node.doc ? [node.doc] : []), ...(node.name ? [node.name] : []), node.path, ...(node.comment ? [node.comment] : [])];
        case "ValueSpec":
            return [...node.names, ...(node.type ? [node.type] : []), ...node.values];
        case "TypeSpec":
            return [node.name, ...(node.typeParams ? [node.typeParams] : []), node.type];
        case "FuncType":
            return [...(node.typeParams ? [node.typeParams] : []), node.params, ...(node.results ? [node.results] : [])];
        case "BlockStmt":
            return node.statements;
        case "DeclStmt":
            return [node.decl];
        case "LabeledStmt":
            return [node.label, node.stmt];
        case "ExprStmt":
            return [node.expr];
        case "AssignStmt":
            return [...node.lhs, ...node.rhs];
        case "IncDecStmt":
            return [node.expr];
        case "ReturnStmt":
            return node.results;
        case "BranchStmt":
            return node.label ? [node.label] : [];
        case "IfStmt":
            return [...(node.init ? [node.init] : []), node.condition, node.body, ...(node.else ? [node.else] : [])];
        case "CaseClause":
            return [...node.list, ...node.body];
        case "CommClause":
            return [...(node.comm ? [node.comm] : []), ...node.body];
        case "SwitchStmt":
            return [...(node.init ? [node.init] : []), ...(node.tag ? [node.tag] : []), ...node.body];
        case "TypeSwitchStmt":
            return [...(node.init ? [node.init] : []), node.assign, ...node.body];
        case "SelectStmt":
            return [...node.body];
        case "ForStmt":
            return [...(node.init ? [node.init] : []), ...(node.condition ? [node.condition] : []), ...(node.post ? [node.post] : []), node.body];
        case "RangeStmt":
            return [...(node.key ? [node.key] : []), ...(node.value ? [node.value] : []), node.source, node.body];
        case "DeferStmt":
            return [node.call];
        case "GoStmt":
            return [node.call];
        case "SendStmt":
            return [node.channel, node.value];
        case "Ellipsis":
            return node.element ? [node.element] : [];
        case "FuncLit":
            return [node.type, node.body];
        case "CompositeLit":
            return [...(node.type ? [node.type] : []), ...node.elements];
        case "ParenExpr":
            return [node.expr];
        case "SelectorExpr":
            return [node.object, node.selector];
        case "IndexExpr":
            return [node.object, node.index];
        case "IndexListExpr":
            return [node.object, ...node.indices];
        case "SliceExpr":
            return [node.object, ...(node.low ? [node.low] : []), ...(node.high ? [node.high] : []), ...(node.max ? [node.max] : [])];
        case "TypeAssertExpr":
            return [node.object, ...(node.type ? [node.type] : [])];
        case "CallExpr":
            return [node.fun, ...node.args];
        case "StarExpr":
            return [node.expr];
        case "UnaryExpr":
            return [node.expr];
        case "BinaryExpr":
            return [node.left, node.right];
        case "KeyValueExpr":
            return [node.key, node.value];
        case "CellRefExpr":
            return [node.namespace];
        case "RangeRefExpr":
            return [node.namespace];
        case "ArrayType":
            return [...(node.length ? [node.length] : []), node.element];
        case "StructType":
            return [node.fields];
        case "InterfaceType":
            return [node.methods];
        case "MapType":
            return [node.key, node.value];
        case "ChanType":
            return [node.value];
        default:
            return [];
    }
}
