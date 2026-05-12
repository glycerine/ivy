import { describe, expect, it } from 'vitest';
import { bootstrapApp, type LocalBootstrapMetadata } from './appBootstrapService';
import { createAuthState, createModelsState, createProjectsState, createWorkspaceState } from '$lib/state';
import type { AuthClient } from '$lib/api';
import type { AuthUser, ModelDocument, Project } from '$lib/types';

const createdAt = '2026-05-12T00:00:00.000Z';

const user: AuthUser = {
	id: 'user-1',
	primaryEmail: 'jane@example.test',
	displayName: 'Jane',
	createdAt
};

const project: Project = {
	id: 'project-1',
	accountId: 'account-1',
	ownerKind: 'user',
	ownerId: 'user-1',
	name: 'Personal',
	slug: 'personal',
	storageMode: 'local_indexeddb',
	createdAt,
	updatedAt: createdAt
};

const model: ModelDocument = {
	id: 'model-1',
	projectId: 'project-1',
	filename: 'model.ivy',
	text: 'type t',
	dirty: false,
	parseRevision: 1,
	engineRevision: 1,
	createdAt,
	updatedAt: createdAt
};

function makeStores() {
	return {
		auth: createAuthState(),
		workspace: createWorkspaceState(),
		projects: createProjectsState(),
		models: createModelsState()
	};
}

function makeMetadataStore(metadata: LocalBootstrapMetadata | null, corrupt = false) {
	let saved: LocalBootstrapMetadata | null = null;
	return {
		get saved() {
			return saved;
		},
		async load() {
			if (corrupt) {
				throw new Error('bad metadata');
			}
			return metadata;
		},
		async save(next: LocalBootstrapMetadata) {
			saved = next;
		}
	};
}

function makeRepository(projects: Project[] = [project], models: ModelDocument[] = [model]) {
	return {
		async listProjects() {
			return projects;
		},
		async listModels(projectId: string) {
			return models.filter((candidate) => candidate.projectId === projectId);
		}
	};
}

describe('bootstrapApp', () => {
	it('hydrates an online authenticated user from local project records', async () => {
		expect.hasAssertions();

		const stores = makeStores();
		const metadataStore = makeMetadataStore({ lastProjectId: 'project-1', lastModelId: 'model-1' });
		const authClient: AuthClient = {
			async me() {
				return {
					status: 'authenticated',
					user,
					csrfToken: 'csrf-1',
					sessionExpiresAt: '2026-05-13T00:00:00.000Z'
				};
			}
		};

		const result = await bootstrapApp({
			authClient,
			repository: makeRepository(),
			metadataStore,
			stores,
			isOnline: () => true
		});

		expect(result).toMatchObject({
			authenticated: true,
			offline: false,
			loginRequired: false,
			selectedProjectId: 'project-1',
			selectedModelId: 'model-1'
		});
		expect(stores.auth.current).toMatchObject({ status: 'authenticated', userId: 'user-1' });
		expect(stores.workspace.current).toMatchObject({
			hydrated: true,
			activeProjectId: 'project-1',
			activeModelId: 'model-1'
		});
		expect(metadataStore.saved).toMatchObject({ authenticatedUserId: 'user-1' });
	});

	it('keeps an online anonymous user in login-required state', async () => {
		expect.hasAssertions();

		const stores = makeStores();
		const result = await bootstrapApp({
			authClient: { async me() { return { status: 'anonymous' }; } },
			repository: makeRepository(),
			metadataStore: makeMetadataStore(null),
			stores,
			isOnline: () => true
		});

		expect(result).toMatchObject({ authenticated: false, loginRequired: true });
		expect(stores.auth.current.status).toBe('anonymous');
		expect(stores.workspace.current.hydrated).toBe(true);
		expect(stores.projects.table.order).toHaveLength(0);
	});

	it('unlocks offline mode for a previously authenticated local user', async () => {
		expect.hasAssertions();

		const stores = makeStores();
		const result = await bootstrapApp({
			authClient: { async me() { throw new Error('should not call network offline'); } },
			repository: makeRepository(),
			metadataStore: makeMetadataStore({ authenticatedUserId: 'user-1' }),
			stores,
			isOnline: () => false,
			now: () => '2026-05-12T00:00:01.000Z'
		});

		expect(result).toMatchObject({ authenticated: true, offline: true, loginRequired: false });
		expect(stores.auth.current).toMatchObject({
			status: 'offline-authenticated',
			userId: 'user-1',
			offlineUnlockedAt: '2026-05-12T00:00:01.000Z'
		});
		expect(stores.workspace.current.activeProjectId).toBe('project-1');
	});

	it('requires login offline when no previous authenticated user is available', async () => {
		expect.hasAssertions();

		const stores = makeStores();
		const result = await bootstrapApp({
			authClient: { async me() { throw new Error('should not call network offline'); } },
			repository: makeRepository(),
			metadataStore: makeMetadataStore(null),
			stores,
			isOnline: () => false
		});

		expect(result).toMatchObject({ authenticated: false, offline: true, loginRequired: true });
		expect(stores.auth.current.status).toBe('expired');
		expect(stores.workspace.current.activeProjectId).toBeNull();
	});

	it('reports corrupt local metadata without crashing bootstrap', async () => {
		expect.hasAssertions();

		const stores = makeStores();
		const result = await bootstrapApp({
			authClient: { async me() { return { status: 'anonymous' }; } },
			repository: makeRepository(),
			metadataStore: makeMetadataStore(null, true),
			stores,
			isOnline: () => true
		});

		expect(result).toMatchObject({
			authenticated: false,
			loginRequired: true,
			localMetadataCorrupt: true
		});
		expect(stores.workspace.current.hydrated).toBe(true);
	});
});
