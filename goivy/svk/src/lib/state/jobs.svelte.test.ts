import { describe, expect, it } from 'vitest';
import { createJobsState } from './jobs.svelte';
import type { VerificationJob } from '$lib/types';

const job: VerificationJob = {
	id: 'job-1',
	sessionId: 'session-1',
	engineId: 'engine-1',
	projectId: 'project-1',
	modelId: 'model-1',
	modelRevision: 1,
	kind: 'check-induction',
	status: 'queued',
	createdAt: '2026-05-12T00:00:00.000Z',
	updatedAt: '2026-05-12T00:00:00.000Z'
};

describe('createJobsState', () => {
	it('tracks job lifecycle transitions', () => {
		expect.hasAssertions();

		const jobs = createJobsState();
		jobs.upsert(job);
		jobs.transition('job-1', 'running', '2026-05-12T00:00:01.000Z');
		jobs.setProgress('job-1', { phase: 'checking', done: 1, total: 3 }, '2026-05-12T00:00:02.000Z');
		jobs.transition('job-1', 'succeeded', '2026-05-12T00:00:03.000Z');

		expect(jobs.table.byId['job-1']).toMatchObject({
			status: 'succeeded',
			startedAt: '2026-05-12T00:00:01.000Z',
			finishedAt: '2026-05-12T00:00:03.000Z',
			progress: { phase: 'checking', done: 1, total: 3 }
		});
		expect(jobs.table.revision).toBe(4);
	});

	it('stores failure details', () => {
		expect.hasAssertions();

		const jobs = createJobsState();
		jobs.upsert(job);
		jobs.fail('job-1', 'solver failed', '2026-05-12T00:00:04.000Z');

		expect(jobs.table.byId['job-1']).toMatchObject({
			status: 'failed',
			error: 'solver failed',
			finishedAt: '2026-05-12T00:00:04.000Z'
		});
	});
});
