import type { EntityTable, Id, VerificationJob, VerificationJobStatus } from '$lib/types';
import { emptyEntityTable } from '$lib/types';

export function createJobsState(initial: EntityTable<VerificationJob> = emptyEntityTable<VerificationJob>()) {
	const table = $state<EntityTable<VerificationJob>>(structuredClone(initial));

	function bump() {
		table.revision += 1;
	}

	return {
		get table() {
			return table;
		},
		upsert(job: VerificationJob) {
			if (!Object.hasOwn(table.byId, job.id)) {
				table.order.push(job.id);
			}
			table.byId[job.id] = job;
			bump();
		},
		setProgress(jobId: Id, progress: VerificationJob['progress'], updatedAt = new Date().toISOString()) {
			const job = table.byId[jobId];
			if (!job) {
				return;
			}
			job.progress = progress;
			job.updatedAt = updatedAt;
			bump();
		},
		transition(jobId: Id, status: VerificationJobStatus, updatedAt = new Date().toISOString()) {
			const job = table.byId[jobId];
			if (!job) {
				return;
			}
			job.status = status;
			job.updatedAt = updatedAt;
			if (status === 'running' && !job.startedAt) {
				job.startedAt = updatedAt;
			}
			if (status === 'succeeded' || status === 'failed' || status === 'cancelled') {
				job.finishedAt = updatedAt;
			}
			bump();
		},
		fail(jobId: Id, error: string, updatedAt = new Date().toISOString()) {
			const job = table.byId[jobId];
			if (!job) {
				return;
			}
			job.error = error;
			job.status = 'failed';
			job.finishedAt = updatedAt;
			job.updatedAt = updatedAt;
			bump();
		}
	};
}
