#!/usr/bin/env node
import { mkdir, readFile, readdir, rm, stat, writeFile, copyFile } from "node:fs/promises";
import { dirname, join, relative } from "node:path";
import { fileURLToPath } from "node:url";

const sourceRoot = "/usr/local/go1.26.4/test";
const scriptDir = dirname(fileURLToPath(import.meta.url));
const destinationRoot = join(scriptDir, "go-toolchain", "test");
const manifestPath = join(scriptDir, "go-toolchain", "MANIFEST.json");
const readmePath = join(scriptDir, "go-toolchain", "README.md");

const unsupportedRules = [
  { name: "go statement", pattern: /\bgo\b/u },
  { name: "select statement", pattern: /\bselect\b/u },
  { name: "channel type", pattern: /\bchan\b/u },
  { name: "channel operation", pattern: /<-/u },
  { name: "recover builtin", pattern: /\brecover\b/u }
];

await rm(join(scriptDir, "go-toolchain"), { recursive: true, force: true });
await mkdir(destinationRoot, { recursive: true });

const includedGo = [];
const skippedGo = [];
const copiedSupport = [];

for await (const sourcePath of walk(sourceRoot)) {
  const rel = relative(sourceRoot, sourcePath);
  const destinationPath = join(destinationRoot, rel);
  if (!sourcePath.endsWith(".go")) {
    await mkdir(dirname(destinationPath), { recursive: true });
    await copyFile(sourcePath, destinationPath);
    copiedSupport.push(rel);
    continue;
  }

  const source = await readFile(sourcePath, "utf8");
  const stripped = stripCommentsAndLiterals(source);
  const reasons = unsupportedRules
    .filter((rule) => rule.pattern.test(stripped))
    .map((rule) => rule.name);

  if (reasons.length > 0) {
    skippedGo.push({ path: rel, reasons });
    continue;
  }

  await mkdir(dirname(destinationPath), { recursive: true });
  await writeFile(destinationPath, source);
  includedGo.push(rel);
}

includedGo.sort();
skippedGo.sort((left, right) => left.path.localeCompare(right.path));
copiedSupport.sort();

await mkdir(dirname(manifestPath), { recursive: true });
await writeFile(manifestPath, `${JSON.stringify({
  sourceRoot,
  destinationRoot,
  filter: {
    included: "Go files without unsupported Go-junior concurrency/recover tokens after comment/string stripping.",
    skippedRules: unsupportedRules.map((rule) => rule.name),
    supportFiles: "All non-.go files are copied as support artifacts."
  },
  counts: {
    includedGo: includedGo.length,
    skippedGo: skippedGo.length,
    copiedSupport: copiedSupport.length
  },
  includedGo,
  skippedGo,
  copiedSupport
}, null, 2)}\n`);

await writeFile(readmePath, `# Go Toolchain Test Corpus

Imported from \`${sourceRoot}\`.

This directory is a copied corpus. A small selected subset is wired into the
active Go-junior test runner through \`gojr/tests/goToolchainCorpus.test.ts\`.

The importer copies Go files from the Go distribution \`test/\` tree when they
do not lexically use unsupported Go-junior concurrency/recover constructs:

- \`go\`
- \`select\`
- \`chan\`
- \`<-\`
- \`recover\`

Comments and string/rune literals are stripped before applying that filter, so
ordinary prose and import paths do not cause false skips. Non-Go support files
are copied unchanged.

See \`MANIFEST.json\` for exact included/skipped file lists.

Current active coverage executes a small selected \`// run\` set first:

- \`alias1.go\`
- \`decl.go\`
- \`helloworld.go\`
- \`for.go\`
- \`closure1.go\`
- \`closure2.go\`
- \`compos.go\`
- \`const8.go\`
- \`func6.go\`
- \`intcvt.go\`
- \`range4.go\`
- \`typeswitch1.go\`
- \`abi/convF_criteria.go\`
- \`abi/convT64_criteria.go\`
- \`abi/defer_aggregate.go\`
- \`abi/double_nested_addressed_struct.go\`
- \`abi/double_nested_struct.go\`
- \`abi/f_ret_z_not.go\`
- \`iota.go\`
- \`ken/for.go\`
- \`ken/intervar.go\`
- \`ken/litfun.go\`
- \`ken/ptrvar.go\`
- \`ken/robfor.go\`
- \`ken/simparray.go\`
- \`ken/simpbool.go\`
- \`ken/simpfun.go\`
- \`ken/strvar.go\`
- \`ken/string.go\`

This smoke set is deliberately small so new failures point at specific missing
Go semantics. It should grow as the runtime gains more standard-library and
dynamic-type fidelity.
`);

console.log(`imported ${includedGo.length} Go files`);
console.log(`skipped ${skippedGo.length} Go files`);
console.log(`copied ${copiedSupport.length} support files`);
console.log(`manifest: ${manifestPath}`);

async function* walk(root) {
  for (const entry of await readdir(root, { withFileTypes: true })) {
    const path = join(root, entry.name);
    if (entry.isDirectory()) {
      yield* walk(path);
      continue;
    }
    if (entry.isFile()) {
      yield path;
      continue;
    }
    if (entry.isSymbolicLink()) {
      const info = await stat(path);
      if (info.isFile()) yield path;
    }
  }
}

function stripCommentsAndLiterals(source) {
  let out = "";
  let index = 0;
  while (index < source.length) {
    const char = source[index] ?? "";
    const next = source[index + 1] ?? "";

    if (char === "/" && next === "/") {
      out += "  ";
      index += 2;
      while (index < source.length && source[index] !== "\n") {
        out += " ";
        index += 1;
      }
      continue;
    }

    if (char === "/" && next === "*") {
      out += "  ";
      index += 2;
      while (index < source.length) {
        const current = source[index] ?? "";
        const following = source[index + 1] ?? "";
        if (current === "*" && following === "/") {
          out += "  ";
          index += 2;
          break;
        }
        out += current === "\n" ? "\n" : " ";
        index += 1;
      }
      continue;
    }

    if (char === "`") {
      out += " ";
      index += 1;
      while (index < source.length) {
        const current = source[index] ?? "";
        out += current === "\n" ? "\n" : " ";
        index += 1;
        if (current === "`") break;
      }
      continue;
    }

    if (char === "\"" || char === "'") {
      const quote = char;
      out += " ";
      index += 1;
      while (index < source.length) {
        const current = source[index] ?? "";
        out += current === "\n" ? "\n" : " ";
        index += 1;
        if (current === "\\") {
          if (index < source.length) {
            out += source[index] === "\n" ? "\n" : " ";
            index += 1;
          }
          continue;
        }
        if (current === quote) break;
      }
      continue;
    }

    out += char;
    index += 1;
  }
  return out;
}
