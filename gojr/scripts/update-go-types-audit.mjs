import fs from "node:fs";
import path from "node:path";
import ts from "typescript";

const repo = path.resolve(new URL("../..", import.meta.url).pathname);
const inventoryPath = process.argv[2] ?? "/private/tmp/go_types_inventory_127.json";
const inventory = JSON.parse(fs.readFileSync(inventoryPath, "utf8"));
const firstInventoryPath = inventory.flatMap((pkg) => pkg.files ?? []).find((file) => file.path)?.path ?? "";
const sourceRootMatch = firstInventoryPath.match(/^(.*\/src\/go\/types)\//);
const sourceRoot = sourceRootMatch ? sourceRootMatch[1] : "/usr/local/go1.27rc1/src/go/types";

function collectTsFiles(dir) {
  if (!fs.existsSync(dir)) return [];
  const out = [];
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const file = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      out.push(...collectTsFiles(file));
    } else if (entry.isFile() && entry.name.endsWith(".ts")) {
      out.push(file);
    }
  }
  return out.sort();
}

const targetFiles = collectTsFiles(path.join(repo, "gojr/src/go/types"));

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
    for (const member of node.members) {
      if (ts.isMethodSignature(member) && member.name && ts.isIdentifier(member.name)) {
        add(`${node.name.text}.${member.name.text}`, "Method", file, "interface method");
      }
      if (ts.isPropertySignature(member) && member.name && ts.isIdentifier(member.name)) {
        add(`${node.name.text}.${member.name.text}`, "Property", file, "interface property");
      }
    }
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
  }
  ts.forEachChild(node, (child) => visitNode(child, file));
}

for (const file of targetFiles) {
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
  method: new Set(["Method"])
};

function rowName(row) {
  return typeof row === "string" ? row : row.name;
}

function rowKind(row) {
  return typeof row === "string" ? "" : row.kind;
}

function receiverNames(recv = "") {
  const bare = recv.replace(/^\*/, "");
  return [...new Set([
    bare,
    bare.replace(/_$/, ""),
    `${bare}Type`,
    `${bare.replace(/_$/, "")}Type`
  ])];
}

function hitsFor(name, category) {
  return (byName.get(name) ?? []).filter((symbol) => kindSets[category].has(symbol.kind)).map((symbol) => symbol.desc);
}

function presentFor(name, category, recv = "") {
  if (category === "method" && recv) {
    return receiverNames(recv).flatMap((receiver) => hitsFor(`${receiver}.${name}`, category));
  }
  return hitsFor(name, category);
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
  "# Go Types Transliteration Audit",
  "",
  `This file is the audit baseline for the fresh symbol-by-symbol TypeScript transliteration of \`${sourceRoot}\`.`,
  "",
  "Target TypeScript files currently included in the inventory are only files under `gojr/src/go/types/`. The legacy frontend checker is deliberately excluded and must not count as the `go/types` port.",
  "",
  "Rule for this pass: every upstream `go/types` declaration must have a corresponding same-kind TypeScript declaration, and function bodies must be ported with visibly isomorphic control flow before being marked complete. `Present` means a same-kind exact or receiver-qualified symbol name exists in the fresh port; it does not prove faithful logic yet. `Missing` means the fresh port has not yet reached symbol-level correspondence.",
  "",
  `Generated from Go source using ${inventoryPath}; TypeScript symbols were collected with the TypeScript compiler API from ${targetFiles.length} fresh-port files.`,
  ""
);

for (const pkg of inventory) {
  out.push(`## go/${pkg.package}`, "");
  for (const file of pkg.files) {
    const rel = file.path.startsWith(`${sourceRoot}/`) ? file.path.slice(sourceRoot.length + 1) : file.path;
    out.push(`### ${rel}`, "");
    out.push(...linesForGroup("Constants", file.consts, (row) => {
      const name = rowName(row);
      total += 1;
      const hits = presentFor(name, "const");
      if (!hits.length) missing += 1;
      return `- \`${name}\` - ${hits.length ? `Present: ${hits.join("; ")}` : "Missing from current checker same-kind exact-symbol inventory"} - logic: declaration-only`;
    }));
    out.push(...linesForGroup("Variables", file.vars, (row) => {
      const name = rowName(row);
      total += 1;
      const hits = presentFor(name, "var");
      if (!hits.length) missing += 1;
      return `- \`${name}\` - ${hits.length ? `Present: ${hits.join("; ")}` : "Missing from current checker same-kind exact-symbol inventory"} - logic: declaration-only`;
    }));
    out.push(...linesForGroup("Types", file.types, (row) => {
      const name = rowName(row);
      total += 1;
      const hits = presentFor(name, "type");
      if (!hits.length) missing += 1;
      return `- \`${name}\`${rowKind(row) ? ` (${rowKind(row)})` : ""} - ${hits.length ? `Present: ${hits.join("; ")}` : "Missing from current checker same-kind exact-symbol inventory"} - structure: not yet 1:1-audited`;
    }));
    out.push(...linesForGroup("Functions", file.funcs, (row) => {
      total += 1;
      const hits = presentFor(row.name, "func");
      if (!hits.length) missing += 1;
      return `- \`${row.name}\` @ ${row.position} - ${hits.length ? `Present: ${hits.join("; ")}` : "Missing from current checker same-kind exact-symbol inventory"} - control-flow shape: ${shapeText(row.shape)} - logic: not yet 1:1-audited`;
    }));
    out.push(...linesForGroup("Methods", file.methods, (row) => {
      total += 1;
      const hits = presentFor(row.name, "method", row.recv);
      if (!hits.length) missing += 1;
      return `- \`${row.recv}.${row.name}\` @ ${row.position} - ${hits.length ? `Present: ${hits.join("; ")}` : "Missing from current checker same-kind exact-symbol inventory"} - control-flow shape: ${shapeText(row.shape)} - logic: not yet 1:1-audited`;
    }));
  }
}

out.splice(6, 0, `Current same-kind exact-symbol audit: ${total - missing}/${total} upstream go/types declarations have a matching symbol in the current TypeScript inventory; ${missing} are missing by exact symbol name.`);
out.splice(7, 0, "");

fs.writeFileSync(path.join(repo, "gojr/docs/GO_TYPES_TRANSLITERATION_AUDIT.md"), out.join("\n"));
console.log(`wrote gojr/docs/GO_TYPES_TRANSLITERATION_AUDIT.md (${total - missing}/${total} present, ${missing} missing)`);
