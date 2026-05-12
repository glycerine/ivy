import { page } from 'vitest/browser';
import { describe, expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import DetailsPane from './DetailsPane.svelte';
import type { CheckResult, GraphNode } from '$lib/types';

describe('DetailsPane', () => {
	it('renders selected graph item metadata and trace action', async () => {
		expect.hasAssertions();
		const selectedNode: GraphNode = {
			id: 'n1',
			obj: 'state_0',
			label: '0',
			classes: ['initial'],
			shortInfo: 'Initial state',
			longInfo: { source: 'init' }
		};
		const latestCheck: CheckResult = {
			id: 'check-1',
			jobId: 'job-1',
			sessionId: 'session-1',
			mode: 'induction',
			z3Contacted: true,
			result: 'fail',
			message: 'counterexample found',
			counterexampleTrace: 'trace-1',
			createdAt: '2026-05-12T00:00:00.000Z'
		};

		render(DetailsPane, { latestCheck, selectedNode, jobs: [] });

		await expect.element(page.getByText('Initial state')).toBeInTheDocument();
		await expect.element(page.getByText('"source": "init"')).toBeInTheDocument();
		await expect.element(page.getByRole('button', { name: 'View error trace' })).toBeVisible();
	});
});
