"use strict";

const childProcess = require("child_process");
const fs = require("fs");
const os = require("os");
const path = require("path");

function argValue(name, defaultValue) {
  const index = process.argv.indexOf(name);
  if (index === -1) return defaultValue;
  if (index + 1 >= process.argv.length) {
    throw new Error(`missing value for ${name}`);
  }
  return process.argv[index + 1];
}

const modulePath = path.resolve(argValue("--module", "./build/wasm/abc.js"));
const aigtoaigPath = path.resolve(argValue("--aigtoaig", "../../ivy/bin/aigtoaig"));
const initAbc = require(modulePath);
const abcRc = fs.readFileSync(path.resolve("abc.rc"), "utf8");

function convertAagToAig(aagText, name) {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), `abc-wasm-${name}-`));
  const aagPath = path.join(dir, `${name}.aag`);
  const aigPath = path.join(dir, `${name}.aig`);
  fs.writeFileSync(aagPath, aagText, "utf8");
  childProcess.execFileSync(aigtoaigPath, [aagPath, aigPath], { stdio: "pipe" });
  return fs.readFileSync(aigPath);
}

async function runAbc(command, files, readPaths) {
  const out = [];
  const err = [];
  const Module = await initAbc({
    noInitialRun: true,
    print: (s) => out.push(s),
    printErr: (s) => err.push(s),
  });

  Module.FS.writeFile("/abc.rc", abcRc);
  for (const [fileName, contents] of Object.entries(files || {})) {
    Module.FS.writeFile(fileName, contents);
  }

  Module.callMain(["-c", command]);

  const resultFiles = {};
  for (const fileName of readPaths || []) {
    if (Module.FS.analyzePath(fileName).exists) {
      resultFiles[fileName] = Module.FS.readFile(fileName, { encoding: "utf8" });
    }
  }

  return {
    stdout: out.join("\n"),
    stderr: err.join("\n"),
    text: `${out.join("\n")}\n${err.join("\n")}`,
    files: resultFiles,
  };
}

function assertIncludes(text, needle, label) {
  if (!text.includes(needle)) {
    throw new Error(`${label}: expected output to include ${JSON.stringify(needle)}\n${text}`);
  }
}

(async function main() {
  const blif = [
    ".model smoke",
    ".inputs a b",
    ".outputs y",
    ".names a b y",
    "11 1",
    ".end",
    "",
  ].join("\n");
  const blifResult = await runAbc(
    "read_blif /smoke.blif; strash; print_stats",
    { "/smoke.blif": blif },
    []
  );
  assertIncludes(blifResult.text, "i/o =    2/    1", "BLIF smoke");
  assertIncludes(blifResult.text, "and =      1", "BLIF smoke");

  const provedAig = convertAagToAig("aag 2 1 1 1 0\n2\n4 0\n4\n", "proved");
  const provedResult = await runAbc(
    "read_aiger /proved.aig; pdr; write_aiger_cex /proved.out",
    { "/proved.aig": provedAig },
    ["/proved.out"]
  );
  assertIncludes(provedResult.text, "Property proved", "AIGER proved smoke");
  if (provedResult.files["/proved.out"]) {
    throw new Error(`AIGER proved smoke: unexpected CEX file\n${provedResult.files["/proved.out"]}`);
  }

  const cexAig = convertAagToAig("aag 2 1 1 1 0\n2\n4 2\n5\n", "cex");
  const cexResult = await runAbc(
    "read_aiger /cex.aig; pdr; write_aiger_cex /cex.out",
    { "/cex.aig": cexAig },
    ["/cex.out"]
  );
  assertIncludes(cexResult.text, "was asserted in frame", "AIGER CEX smoke");
  if (!cexResult.files["/cex.out"]) {
    throw new Error(`AIGER CEX smoke: expected CEX file\n${cexResult.text}`);
  }

  console.log("abc wasm smoke test passed");
})().catch((error) => {
  console.error(error && error.stack ? error.stack : error);
  process.exit(1);
});
