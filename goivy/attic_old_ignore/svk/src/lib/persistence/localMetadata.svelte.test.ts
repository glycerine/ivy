import { afterEach, describe, expect, it } from 'vitest';
import { LocalStorageMetadataStore } from './localMetadata';

describe('LocalStorageMetadataStore', () => {
	afterEach(() => {
		localStorage.clear();
	});

	it('stores the last authenticated offline shell metadata', async () => {
		expect.hasAssertions();

		const store = new LocalStorageMetadataStore('svk.test.metadata');
		await store.save({
			authenticatedUserId: 'user-1',
			lastProjectId: 'project-1',
			lastModelId: 'model-1'
		});

		await expect(store.load()).resolves.toEqual({
			authenticatedUserId: 'user-1',
			lastProjectId: 'project-1',
			lastModelId: 'model-1'
		});
	});

	it('surfaces corrupt metadata so bootstrap can fall back safely', async () => {
		expect.hasAssertions();

		localStorage.setItem('svk.test.metadata', '{bad json');
		const store = new LocalStorageMetadataStore('svk.test.metadata');

		await expect(store.load()).rejects.toThrow();
	});
});
