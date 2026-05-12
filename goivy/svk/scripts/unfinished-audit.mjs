import { readdirSync, readFileSync, statSync } from 'node:fs';
import { join, relative, resolve } from 'node:path';

const root = resolve(import.meta.dirname, '..');
const scanRoots = ['src', 'server'].map((path) => resolve(root, path));
const blocked = [
	/\bTODO\b/i,
	/\bstub\b/i,
	/\bplaceholder\b/i,
	/\bcanned\b/i,
	/Fake engine/,
	/Starting local fake engine/,
	/Check FAILED \(induction\) \[Z3: yes\] - counterexample found/
];
const excludedPathParts = [
	'.test.',
	'.spec.',
	'/demo/',
	'/vitest-examples/',
	'/workers/smtZ3Imports.js',
	'/workers/goivyNodeFS.js',
	'/engines/fakeEngine.ts'
];
const excludedExtensions = new Set(['.map', '.wasm']);

function extname(path) {
	const dot = path.lastIndexOf('.');
	return dot >= 0 ? path.slice(dot) : '';
}

function walk(dir, files = []) {
	for (const entry of readdirSync(dir)) {
		const path = join(dir, entry);
		const stat = statSync(path);
		if (stat.isDirectory()) {
			if (entry === 'node_modules' || entry === '.svelte-kit') continue;
			walk(path, files);
		} else {
			files.push(path);
		}
	}
	return files;
}

function isExcluded(path) {
	const normalized = `/${relative(root, path).replaceAll('\\', '/')}`;
	return excludedPathParts.some((part) => normalized.includes(part)) || excludedExtensions.has(extname(path));
}

const findings = [];
for (const scanRoot of scanRoots) {
	for (const file of walk(scanRoot)) {
		if (isExcluded(file)) continue;
		const text = readFileSync(file, 'utf8');
		const lines = text.split(/\r?\n/);
		lines.forEach((line, index) => {
			for (const pattern of blocked) {
				if (pattern.test(line)) {
					findings.push(`${relative(root, file)}:${index + 1}: ${line.trim()}`);
					break;
				}
			}
		});
	}
}

if (findings.length > 0) {
	console.error(`unfinished:audit: ${findings.length} unfinished markers found`);
	for (const finding of findings) console.error(finding);
	process.exit(1);
}

console.log('unfinished:audit: no unfinished production markers found');
