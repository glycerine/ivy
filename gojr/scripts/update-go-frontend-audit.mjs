import fs from "node:fs";
import path from "node:path";
import ts from "typescript";

const repo = path.resolve(new URL("../..", import.meta.url).pathname);
const inventoryPath = process.argv[2] ?? "/private/tmp/go_frontend_inventory.json";
const inventory = JSON.parse(fs.readFileSync(inventoryPath, "utf8"));

const targetFiles = [
  "gojr/src/front/token.ts",
  "gojr/src/front/scanner.ts",
  "gojr/src/front/ast.ts",
  "gojr/src/front/parser.ts",
  "gojr/src/front/checker.ts",
  "gojr/src/front/types.ts",
  "gojr/src/frontToAst.ts",
  "gojr/src/go/format.ts"
];

const symbols = [];

function add(name, kind, file, extra = "") {
  if (!name) return;
  const rel = path.relative(repo, file);
  symbols.push({ name, kind, desc: `${kind} ${name}${extra ? ` ${extra}` : ""} (${rel})` });
}

function visitNode(node, file) {
  if (ts.isClassDeclaration(node) && node.name) {
    add(node.name.text, "ClassDeclaration", file);
    for (const member of node.members) {
      if (ts.isMethodDeclaration(member) && member.name && ts.isIdentifier(member.name)) {
        add(member.name.text, "Method", file, `on ${node.name.text}`);
        add(`${node.name.text}.${member.name.text}`, "Method", file);
      }
      if (ts.isPropertyDeclaration(member) && member.name && ts.isIdentifier(member.name)) {
        add(member.name.text, "Property", file, `on ${node.name.text}`);
        add(`${node.name.text}.${member.name.text}`, "Property", file);
      }
    }
  } else if (ts.isInterfaceDeclaration(node) && node.name) {
    add(node.name.text, "InterfaceDeclaration", file);
  } else if (ts.isTypeAliasDeclaration(node) && node.name) {
    add(node.name.text, "TypeAliasDeclaration", file);
  } else if (ts.isEnumDeclaration(node) && node.name) {
    add(node.name.text, "EnumDeclaration", file);
    for (const member of node.members) {
      if (member.name && ts.isIdentifier(member.name)) add(member.name.text, "EnumMember", file, `in ${node.name.text}`);
    }
  } else if (ts.isFunctionDeclaration(node) && node.name) {
    add(node.name.text, "FunctionDeclaration", file);
  } else if (ts.isVariableStatement(node)) {
    for (const decl of node.declarationList.declarations) {
      if (ts.isIdentifier(decl.name)) add(decl.name.text, "Variable", file);
    }
  } else if (ts.isModuleDeclaration(node) && ts.isIdentifier(node.name) && node.body && ts.isModuleBlock(node.body)) {
    for (const stmt of node.body.statements) {
      if (ts.isFunctionDeclaration(stmt) && stmt.name) add(`${node.name.text}.${stmt.name.text}`, "Method", file, "namespace function");
    }
  }
  ts.forEachChild(node, (child) => visitNode(child, file));
}

for (const rel of targetFiles) {
  const file = path.join(repo, rel);
  if (!fs.existsSync(file)) continue;
  const source = ts.createSourceFile(file, fs.readFileSync(file, "utf8"), ts.ScriptTarget.Latest, true);
  visitNode(source, file);
}

const byName = new Map();
for (const symbol of symbols) {
  if (!byName.has(symbol.name)) byName.set(symbol.name, []);
  byName.get(symbol.name).push(symbol);
}

const kindSets = {
  const: new Set(["EnumMember", "Variable"]),
  var: new Set(["Variable", "Property"]),
  type: new Set(["ClassDeclaration", "InterfaceDeclaration", "TypeAliasDeclaration", "EnumDeclaration"]),
  func: new Set(["FunctionDeclaration"]),
  method: new Set(["Method"]),
  production: new Set(["Method", "FunctionDeclaration"])
};

function hitsFor(name, category) {
  return (byName.get(name) ?? []).filter((symbol) => kindSets[category].has(symbol.kind)).map((symbol) => symbol.desc);
}

function presentFor(name, category, recv = "") {
  if ((category === "method" || category === "production") && recv) {
    return hitsFor(`${recv.replace(/^\*/, "")}.${name}`, category);
  }
  return hitsFor(name, category);
}

function rowName(row) {
  return typeof row === "string" ? row : row.name;
}

function rowKind(row) {
  return typeof row === "string" ? "" : row.kind;
}

function shapeText(shape) {
  if (!shape) return "declaration-only";
  return ["if", "for", "range", "switch", "typeSwitch", "select", "branch", "assign", "return", "defer", "go", "call", "binary", "unary", "composite", "funcLiteral"]
    .map((key) => `${key}=${shape[key] ?? 0}`)
    .join(", ");
}

function linesForGroup(title, rows, fmt) {
  if (!rows || rows.length === 0) return [];
  const out = [`#### ${title}`];
  for (const row of rows) out.push(fmt(row));
  out.push("");
  return out;
}

let total = 0;
let missing = 0;
const out = [];

out.push(
  "# Go Frontend Transliteration Audit",
  "",
  "This file is the audit baseline for fixing the existing Go-junior frontend in place against `/usr/local/go1.26.4/src/go/scanner`, `/usr/local/go1.26.4/src/go/ast`, and `/usr/local/go1.26.4/src/go/parser`.",
  "",
  "Target TypeScript files are the original implementation only: `gojr/src/front/token.ts`, `gojr/src/front/scanner.ts`, `gojr/src/front/ast.ts`, `gojr/src/front/parser.ts`, `gojr/src/front/checker.ts`, `gojr/src/front/types.ts`, `gojr/src/frontToAst.ts`, and `gojr/src/go/format.ts`. No parallel `gojr/src/std/go` port is part of this plan.",
  "",
  "Rule for this pass: every upstream declaration must have a corresponding same-kind TypeScript declaration in the original frontend, and function bodies must be audited for 1:1 control-flow correspondence before being marked complete. `Present` means a same-kind exact or receiver-qualified symbol name exists; it does not yet prove faithful transliteration. `Missing` means the original frontend has not yet been brought into symbol-level correspondence.",
  "",
  `Generated from Go source using ${inventoryPath}; TypeScript symbols were collected with the TypeScript compiler API from the target files above.`,
  ""
);

for (const pkg of inventory) {
  out.push(`## go/${pkg.package}`, "");
  for (const file of pkg.files) {
    const rel = file.path.replace("/usr/local/go1.26.4/src/go/", "");
    out.push(`### ${rel}`, "");
    out.push(...linesForGroup("Constants", file.consts, (row) => {
      const name = rowName(row);
      total += 1;
      const hits = presentFor(name, "const");
      if (!hits.length) missing += 1;
      return `- \`${name}\` - ${hits.length ? `Present: ${hits.join("; ")}` : "Missing from original frontend same-kind exact-symbol inventory"} - logic: declaration-only`;
    }));
    out.push(...linesForGroup("Variables", file.vars, (row) => {
      const name = rowName(row);
      total += 1;
      const hits = presentFor(name, "var");
      if (!hits.length) missing += 1;
      return `- \`${name}\` - ${hits.length ? `Present: ${hits.join("; ")}` : "Missing from original frontend same-kind exact-symbol inventory"} - logic: declaration-only`;
    }));
    out.push(...linesForGroup("Types", file.types, (row) => {
      const name = rowName(row);
      total += 1;
      const hits = presentFor(name, "type");
      if (!hits.length) missing += 1;
      return `- \`${name}\`${rowKind(row) ? ` (${rowKind(row)})` : ""} - ${hits.length ? `Present: ${hits.join("; ")}` : "Missing from original frontend same-kind exact-symbol inventory"} - structure: not yet 1:1-audited`;
    }));
    out.push(...linesForGroup("Functions", file.funcs, (row) => {
      total += 1;
      const hits = presentFor(row.name, "func");
      if (!hits.length) missing += 1;
      return `- \`${row.name}\` @ ${row.position} - ${hits.length ? `Present: ${hits.join("; ")}` : "Missing from original frontend same-kind exact-symbol inventory"} - control-flow shape: ${shapeText(row.shape)} - logic: not yet 1:1-audited`;
    }));
    out.push(...linesForGroup("Methods", file.methods, (row) => {
      total += 1;
      const hits = presentFor(row.name, "method", row.recv);
      if (!hits.length) missing += 1;
      return `- \`${row.recv}.${row.name}\` @ ${row.position} - ${hits.length ? `Present: ${hits.join("; ")}` : "Missing from original frontend same-kind exact-symbol inventory"} - control-flow shape: ${shapeText(row.shape)} - logic: not yet 1:1-audited`;
    }));
    const productions = file.productions ?? [];
    if (productions.length) {
      out.push("#### Parser Productions");
      for (const row of productions) {
        total += 1;
        const hits = presentFor(row.name, "production", row.recv);
        if (!hits.length) missing += 1;
        out.push(`- \`${row.recv}.${row.name}\` @ ${row.position} - ${hits.length ? `Present: ${hits.join("; ")}` : "Missing from original frontend same-kind exact-symbol inventory"} - control-flow shape: ${shapeText(row.shape)} - logic: not yet 1:1-audited`);
      }
      out.push("");
    }
  }
}

out.splice(6, 0, `Current same-kind exact-symbol audit: ${total - missing}/${total} upstream declarations or productions have a matching symbol in the original frontend inventory; ${missing} are missing by exact symbol name.`);
out.splice(7, 0, "");

fs.writeFileSync(path.join(repo, "gojr/docs/GO_FRONTEND_TRANSLITERATION_AUDIT.md"), out.join("\n"));
console.log(`wrote gojr/docs/GO_FRONTEND_TRANSLITERATION_AUDIT.md (${total - missing}/${total} present, ${missing} missing)`);
