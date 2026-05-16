#!/usr/bin/env node

import fs from 'node:fs';
import path from 'node:path';

const [, , inputRoot, outputFile] = process.argv;

if (!inputRoot || !outputFile) {
  console.error('usage: node webui/scripts/build-include-tree.mjs <include-root> <output-json>');
  process.exit(2);
}

const root = path.resolve(inputRoot);
const files = [];

function walk(dir) {
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      walk(full);
      continue;
    }
    if (!entry.isFile() || !entry.name.endsWith('.ivy')) continue;
    files.push({
      path: path.relative(root, full).split(path.sep).join('/'),
      data: fs.readFileSync(full, 'utf8'),
    });
  }
}

walk(root);
files.sort((a, b) => a.path.localeCompare(b.path));

fs.mkdirSync(path.dirname(outputFile), { recursive: true });
fs.writeFileSync(outputFile, `${JSON.stringify({ root: '/include', files })}\n`);
