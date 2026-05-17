import { describe, expect, it } from 'vitest';
import { normalizeTraceEventSheet } from './events';

describe('normalizeTraceEventSheet', () => {
	it('normalizes nested trace events with stable address IDs', () => {
		expect.hasAssertions();

		const sheet = normalizeTraceEventSheet(
			[
				{
					text: 'root',
					subs: [{ text: 'child' }]
				}
			],
			{
				id: 'trace-1',
				projectId: 'project-1',
				sessionId: 's1',
				label: 'Trace',
				patterns: ['*'],
				revision: 2
			}
		);

		expect(sheet).toMatchObject({
			id: 'trace-1',
			projectId: 'project-1',
			sessionId: 's1',
			label: 'Trace',
			patterns: ['*'],
			expandedAddresses: [],
			revision: 2
		});
		expect(sheet.events[0]).toMatchObject({
			id: 'event_0',
			text: 'root',
			address: '0'
		});
		expect(sheet.events[0].children?.[0]).toMatchObject({
			id: 'event_0_0',
			text: 'child',
			address: '0/0'
		});
	});
});
