import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';

const testDir = path.dirname(fileURLToPath(import.meta.url));
const frontendSrcDir = path.resolve(testDir, '../frontend/src');

function methodBlocks(source) {
  const lines = source.split(/\n/);
  const blocks = [];
  for (let i = 0; i < lines.length; i++) {
    const match = lines[i].match(/^\s*(?:async\s+)?([A-Za-z_$][\w$]*)\([^)]*\)\s*\{/);
    if (!match) continue;

    let depth = 0;
    let started = false;
    let end = i;
    for (let j = i; j < lines.length; j++) {
      const line = lines[j].replace(/\/\/.*$/, '');
      for (const ch of line) {
        if (ch === '{') {
          depth++;
          started = true;
        } else if (ch === '}') {
          depth--;
          if (started && depth === 0) {
            end = j;
            break;
          }
        }
      }
      if (started && depth === 0) break;
    }

    blocks.push({
      name: match[1],
      startLine: i + 1,
      endLine: end + 1,
      body: lines.slice(i, end + 1).join('\n'),
    });
    i = end;
  }
  return blocks;
}

describe('bundled frontend self binding', () => {
  it('binds self in methods before callbacks use it', () => {
    const files = [
      path.join(frontendSrcDir, 'legacyAppController.js'),
      path.join(frontendSrcDir, 'legacyPersist.js'),
    ];
    const missingBindings = [];

    for (const file of files) {
      const source = fs.readFileSync(file, 'utf8');
      for (const block of methodBlocks(source)) {
        if (/\bself\b/.test(block.body) && !/\bvar\s+self\s*=\s*this\b/.test(block.body)) {
          missingBindings.push(`${path.relative(frontendSrcDir, file)}:${block.startLine}-${block.endLine} ${block.name}()`);
        }
      }
    }

    expect(missingBindings).toEqual([]);
  });
});
