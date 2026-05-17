import type { Id, Revision, SyncEntity, SyncOp } from '$lib/types';
import type { SvkRepository } from './repositories';

export type SyncQueueOptions = {
	now?: () => string;
	createId?: () => Id;
};

export type EnqueueSyncOp = {
	projectId: Id;
	entity: SyncEntity;
	op: SyncOp['op'];
	payload: unknown;
	baseRevision?: Revision;
};

export class SyncQueue {
	private readonly now: () => string;
	private readonly createId: () => Id;

	constructor(
		private readonly repository: SvkRepository,
		options: SyncQueueOptions = {}
	) {
		this.now = options.now ?? (() => new Date().toISOString());
		this.createId = options.createId ?? (() => crypto.randomUUID());
	}

	async enqueue(input: EnqueueSyncOp): Promise<SyncOp> {
		const timestamp = this.now();
		const op: SyncOp = {
			id: this.createId(),
			projectId: input.projectId,
			entity: input.entity,
			op: input.op,
			baseRevision: input.baseRevision,
			payload: input.payload,
			status: 'pending',
			createdAt: timestamp,
			updatedAt: timestamp
		};

		await this.repository.saveSyncOp(op);
		return op;
	}

	listPending(projectId?: Id) {
		return projectId
			? this.repository.listProjectSyncOps(projectId, 'pending')
			: this.repository.listSyncOps('pending');
	}

	async markSyncing(id: Id): Promise<SyncOp | undefined> {
		return this.transition(id, 'syncing');
	}

	async markSynced(id: Id): Promise<SyncOp | undefined> {
		return this.transition(id, 'synced');
	}

	async markConflict(id: Id, error: string): Promise<SyncOp | undefined> {
		return this.transition(id, 'conflict', error);
	}

	async markFailed(id: Id, error: string): Promise<SyncOp | undefined> {
		return this.transition(id, 'failed', error);
	}

	private async transition(id: Id, status: SyncOp['status'], error?: string): Promise<SyncOp | undefined> {
		const op = await this.repository.getSyncOp(id);
		if (!op) {
			return undefined;
		}

		const updated: SyncOp = {
			...op,
			status,
			error,
			updatedAt: this.now()
		};
		await this.repository.saveSyncOp(updated);
		return updated;
	}
}
