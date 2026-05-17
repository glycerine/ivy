import type { LocalBootstrapMetadata, LocalBootstrapMetadataStore } from '$lib/services/appBootstrapService';

export class LocalStorageMetadataStore implements LocalBootstrapMetadataStore {
	constructor(private readonly key = 'svk.localMetadata.v1') {}

	async load(): Promise<LocalBootstrapMetadata | null> {
		const value = localStorage.getItem(this.key);
		if (!value) {
			return null;
		}
		return JSON.parse(value) as LocalBootstrapMetadata;
	}

	async save(metadata: LocalBootstrapMetadata): Promise<void> {
		localStorage.setItem(this.key, JSON.stringify(metadata));
	}

	async clear(): Promise<void> {
		localStorage.removeItem(this.key);
	}
}
