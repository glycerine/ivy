import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

const root = resolve(import.meta.dirname, '..');
const inventoryPath = resolve(root, 'src/lib/workbench/parity/webuiParityInventory.ts');
const source = readFileSync(inventoryPath, 'utf8');

const ids = [...source.matchAll(/id:\s*'([^']+)'/g)].map((match) => match[1]);
const statuses = [...source.matchAll(/status:\s*'([^']+)'/g)].map((match) => match[1]);
const implemented = [...source.matchAll(/status:\s*'implemented'/g)].length;
const planned = [...source.matchAll(/status:\s*'planned'/g)].length;
const duplicateIds = ids.filter((id, index) => ids.indexOf(id) !== index);
const badStatuses = statuses.filter((status) => !['implemented', 'planned'].includes(status));
const requiredIds = [
	'component.menubar',
	'component.editor-pane',
	'component.state-relations-pane',
	'component.event-trace-sheet',
	'component.tutorial-pane',
	'menu.toolbar.check',
	'menu.arg.check-induction',
	'menu.concept.conjecture',
	'command.registry',
	'workflow.local-first-wasm',
	'workflow.hosted-webui'
];
const missingRequired = requiredIds.filter((id) => !ids.includes(id));

const errors = [];
if (ids.length === 0) errors.push('inventory has no ids');
if (duplicateIds.length > 0) errors.push(`duplicate inventory ids: ${[...new Set(duplicateIds)].join(', ')}`);
if (badStatuses.length > 0) errors.push(`bad statuses: ${[...new Set(badStatuses)].join(', ')}`);
if (missingRequired.length > 0) errors.push(`missing required parity ids: ${missingRequired.join(', ')}`);
if (implemented === 0) errors.push('inventory has no implemented entries');

if (process.env.PARITY_ENFORCE_IMPLEMENTED === '1' && planned > 0) {
	errors.push(`${planned} parity items are still planned`);
}

if (errors.length > 0) {
	for (const error of errors) console.error(`parity:audit: ${error}`);
	process.exit(1);
}

console.log(`parity:audit: ${ids.length} items (${implemented} implemented, ${planned} planned)`);
