import { describe, expect, it } from 'vitest';
import { emptyEntityTable, upsertEntity } from './ids';

describe('EntityTable helpers', () => {
	it('upserts records while preserving first-seen order', () => {
		expect.hasAssertions();

		let table = emptyEntityTable<{ id: string; name: string }>();
		table = upsertEntity(table, { id: 'a', name: 'A' });
		table = upsertEntity(table, { id: 'b', name: 'B' });
		table = upsertEntity(table, { id: 'a', name: 'A2' });

		expect(table.order).toEqual(['a', 'b']);
		expect(table.byId.a.name).toBe('A2');
		expect(table.revision).toBe(3);
	});
});
