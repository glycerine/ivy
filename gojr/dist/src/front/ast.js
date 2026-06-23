export function ident(name, span) {
    return { kind: "Ident", name, ...(span ? { span } : {}) };
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
            return [...(node.name ? [node.name] : []), node.path];
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
