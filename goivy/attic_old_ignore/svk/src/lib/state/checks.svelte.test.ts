import { describe, expect, it } from 'vitest';
import { createChecksState } from './checks.svelte';
import type { CheckResult } from '$lib/types';

const result: CheckResult = {
	id: 'check-1',
	jobId: 'job-1',
	sessionId: 'session-1',
	mode: 'induction',
	z3Contacted: true,
	result: 'fail',
	message: 'counterexample found',
	failedConjecture: '~link(X,Y)',
	createdAt: '2026-05-12T00:00:00.000Z'
};

describe('createChecksState', () => {
	it('tracks latest check result by session', () => {
		expect.hasAssertions();

		const checks = createChecksState();
		checks.upsert(result);
		checks.upsert({
			...result,
			id: 'check-2',
			jobId: 'job-2',
			result: 'pass',
			message: 'ok'
		});

		expect(checks.table.order).toEqual(['check-1', 'check-2']);
		expect(checks.latestForSession('session-1')).toMatchObject({ id: 'check-2', result: 'pass' });
		expect(checks.statusText('session-1')).toBe('Check PASS (induction) [Z3: yes]');
	});

	it('has stable empty-session status text', () => {
		expect.hasAssertions();

		const checks = createChecksState();

		expect(checks.latestForSession('missing')).toBeNull();
		expect(checks.statusText('missing')).toBe('No check has run');
	});
});

