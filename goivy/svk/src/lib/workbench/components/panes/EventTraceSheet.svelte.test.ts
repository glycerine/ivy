import { page } from 'vitest/browser';
import { describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import EventTraceSheet from './EventTraceSheet.svelte';
import type { TraceEventSheet } from '$lib/types';

describe('EventTraceSheet', () => {
	it('renders event controls and dispatches pattern commands', async () => {
		expect.hasAssertions();
		const onCommand = vi.fn();
		const sheet: TraceEventSheet = {
			id: 'events-1',
			projectId: 'project-1',
			sessionId: 'session-1',
			label: 'Trace',
			events: [{ id: 'event-1', text: 'root(a)', address: '0' }],
			patterns: [],
			expandedAddresses: [],
			revision: 1
		};

		render(EventTraceSheet, { sheet, onCommand });
		await page.getByLabelText('Event pattern').fill('root');
		await page.getByRole('button', { name: 'Filter' }).click();
		await page.getByRole('button', { name: 'Save' }).click();

		await expect.element(page.getByText('root(a)')).toBeInTheDocument();
		expect(onCommand).toHaveBeenCalledWith('filterEventTrace', 'root');
		expect(onCommand).toHaveBeenCalledWith('saveEventPatterns');
	});
});
