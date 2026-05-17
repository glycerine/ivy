import { describe, expect, it } from 'vitest';
import { normalizeCheckResult } from './checks';

describe('normalizeCheckResult', () => {
	it('normalizes failed webui check results', () => {
		expect.hasAssertions();

		const result = normalizeCheckResult(
			{
				z3_contacted: true,
				result: 'fail',
				message: 'not inductive',
				failed_conjecture: 'forall X. safe(X)',
				failed_label: 'safety',
				used_relations: ['link'],
				counterexample_trace: 'trace',
				counterexample_details: 'details'
			},
			{
				jobId: 'job-1',
				sessionId: 's1',
				mode: 'induction',
				now: new Date('2026-05-12T00:00:00.000Z')
			}
		);

		expect(result).toMatchObject({
			id: 'check_job-1',
			jobId: 'job-1',
			sessionId: 's1',
			mode: 'induction',
			z3Contacted: true,
			result: 'fail',
			message: 'not inductive',
			failedConjecture: 'forall X. safe(X)',
			failedLabel: 'safety',
			usedRelations: ['link'],
			counterexampleTrace: 'trace',
			counterexampleDetails: 'details',
			createdAt: '2026-05-12T00:00:00.000Z'
		});
	});

	it('treats unknown result values as errors', () => {
		expect.hasAssertions();

		const result = normalizeCheckResult({ result: 'mystery' }, { jobId: 'job-1', sessionId: 's1', mode: 'pdr' });

		expect(result.result).toBe('error');
		expect(result.message).toBe('');
	});
});
