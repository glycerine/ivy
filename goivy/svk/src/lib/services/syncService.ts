import { HostedConflictError, type HostedProjectClient } from '$lib/api/projectClient';
import type { SyncOp } from '$lib/types';
import type { SyncQueue } from '$lib/persistence';

export class SyncService {
	constructor(
		private readonly queue: SyncQueue,
		private readonly client: HostedProjectClient
	) {}

	async uploadPendingModelRevisions(projectId: string): Promise<SyncOp[]> {
		const pending = await this.queue.listPending(projectId);
		const processed: SyncOp[] = [];
		for (const op of pending.filter((candidate) => candidate.entity === 'model_revision')) {
			try {
				const payload = op.payload as { filename: string; text: string };
				await this.client.saveModel({
					projectId,
					filename: payload.filename,
					text: payload.text,
					baseRevision: op.baseRevision
				});
				const synced = await this.queue.markSynced(op.id);
				if (synced) {
					processed.push(synced);
				}
			} catch (error) {
				const failed =
					error instanceof HostedConflictError
						? await this.queue.markConflict(op.id, error.message)
						: await this.queue.markFailed(op.id, error instanceof Error ? error.message : 'sync failed');
				if (failed) {
					processed.push(failed);
				}
			}
		}
		return processed;
	}
}
