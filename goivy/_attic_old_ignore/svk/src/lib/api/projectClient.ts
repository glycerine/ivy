import type { Id, ModelDocument, Revision } from '$lib/types';

export type SaveHostedModelInput = {
	projectId: Id;
	filename: string;
	text: string;
	baseRevision?: Revision;
};

export type SaveHostedModelResponse = {
	model: ModelDocument & { revision?: Revision };
};

export class HostedProjectClient {
	constructor(private readonly fetcher: typeof fetch = fetch) {}

	async saveModel(input: SaveHostedModelInput): Promise<SaveHostedModelResponse> {
		const response = await this.fetcher(`/api/projects/${encodeURIComponent(input.projectId)}/models`, {
			method: 'POST',
			credentials: 'include',
			headers: { accept: 'application/json', 'content-type': 'application/json' },
			body: JSON.stringify({
				filename: input.filename,
				text: input.text,
				baseRevision: input.baseRevision ?? 0
			})
		});
		if (response.status === 409) {
			throw new HostedConflictError('model revision conflict');
		}
		if (!response.ok) {
			throw new Error(`save model failed with ${response.status}`);
		}
		return (await response.json()) as SaveHostedModelResponse;
	}
}

export class HostedConflictError extends Error {
	constructor(message: string) {
		super(message);
		this.name = 'HostedConflictError';
	}
}
