import { describe, expect, it } from 'vitest';
import { WEBUI_PARITY_ITEMS, paritySummary } from './webuiParityInventory';

describe('webui parity inventory', () => {
	it('keeps stable unique ids for every parity item', () => {
		expect.hasAssertions();
		const ids = WEBUI_PARITY_ITEMS.map((item) => item.id);
		expect(new Set(ids).size).toBe(ids.length);
		expect(ids.every((id) => /^[a-z]+[a-z0-9.-]*$/.test(id))).toBe(true);
	});

	it('records source and implementation status for every item', () => {
		expect.hasAssertions();
		for (const item of WEBUI_PARITY_ITEMS) {
			expect(item.source, item.id).toBeTruthy();
			expect(['implemented', 'planned']).toContain(item.status);
			if (item.status === 'implemented') {
				expect(item.implementation, item.id).toBeTruthy();
			}
		}
	});

	it('covers the main old workbench surfaces', () => {
		expect.hasAssertions();
		const ids = new Set(WEBUI_PARITY_ITEMS.map((item) => item.id));
		expect(ids).toContain('component.menubar');
		expect(ids).toContain('component.editor-pane');
		expect(ids).toContain('component.state-relations-pane');
		expect(ids).toContain('component.event-trace-sheet');
		expect(ids).toContain('menu.toolbar.check');
		expect(ids).toContain('menu.arg.check-induction');
		expect(ids).toContain('command.registry');
		expect(ids).toContain('workflow.local-first-wasm');
	});

	it('summarizes current planned and implemented work', () => {
		expect.hasAssertions();
		const summary = paritySummary();
		expect(summary.total).toBe(WEBUI_PARITY_ITEMS.length);
		expect(summary.byStatus.planned).toBeGreaterThan(0);
		expect(summary.byStatus.implemented).toBeGreaterThan(0);
		expect(summary.byKind.component).toBeGreaterThan(10);
	});
});
