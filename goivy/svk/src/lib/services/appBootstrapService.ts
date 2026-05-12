import type { AuthClient } from '$lib/api';
import type { Id, ModelDocument, Project } from '$lib/types';
import type { createAuthState } from '$lib/state/auth.svelte';
import type { createModelsState } from '$lib/state/models.svelte';
import type { createProjectsState } from '$lib/state/projects.svelte';
import type { createWorkspaceState } from '$lib/state/workspace.svelte';

export type LocalBootstrapMetadata = {
	authenticatedUserId?: Id;
	lastProjectId?: Id;
	lastModelId?: Id;
};

export interface LocalBootstrapMetadataStore {
	load(): Promise<LocalBootstrapMetadata | null>;
	save?(metadata: LocalBootstrapMetadata): Promise<void>;
}

export interface BootstrapRepository {
	listProjects(): Promise<Project[]>;
	listModels(projectId: Id): Promise<ModelDocument[]>;
}

export type BootstrapStores = {
	auth: ReturnType<typeof createAuthState>;
	workspace: ReturnType<typeof createWorkspaceState>;
	projects: ReturnType<typeof createProjectsState>;
	models: ReturnType<typeof createModelsState>;
};

export type BootstrapOptions = {
	authClient: AuthClient;
	repository: BootstrapRepository;
	metadataStore: LocalBootstrapMetadataStore;
	stores: BootstrapStores;
	isOnline?: () => boolean;
	now?: () => string;
};

export type BootstrapResult = {
	authenticated: boolean;
	offline: boolean;
	loginRequired: boolean;
	localMetadataCorrupt: boolean;
	selectedProjectId: Id | null;
	selectedModelId: Id | null;
};

export async function bootstrapApp(options: BootstrapOptions): Promise<BootstrapResult> {
	const online = options.isOnline?.() ?? navigator.onLine;
	const now = options.now ?? (() => new Date().toISOString());
	const { auth, workspace, projects, models } = options.stores;

	workspace.setOnline(online);

	let metadata: LocalBootstrapMetadata | null = null;
	let localMetadataCorrupt = false;
	try {
		metadata = await options.metadataStore.load();
	} catch {
		localMetadataCorrupt = true;
	}

	let authenticated = false;
	let loginRequired = false;

	if (online) {
		const authMe = await options.authClient.me();
		if (authMe.status === 'authenticated') {
			auth.setAuthenticated(authMe.user, authMe.sessionExpiresAt, authMe.csrfToken);
			authenticated = true;
			await options.metadataStore.save?.({ ...metadata, authenticatedUserId: authMe.user.id });
		} else {
			auth.setAnonymous();
			loginRequired = true;
		}
	} else if (metadata?.authenticatedUserId) {
		auth.setOfflineAuthenticated(metadata.authenticatedUserId, now());
		authenticated = true;
	} else {
		auth.setExpired();
		loginRequired = true;
	}

	if (!authenticated) {
		projects.replaceAll([]);
		workspace.selectProject(null);
		workspace.hydrate();
		return {
			authenticated,
			offline: !online,
			loginRequired,
			localMetadataCorrupt,
			selectedProjectId: null,
			selectedModelId: null
		};
	}

	const localProjects = await options.repository.listProjects();
	projects.replaceAll(localProjects);

	const selectedProjectId =
		(metadata?.lastProjectId && projects.table.byId[metadata.lastProjectId]
			? metadata.lastProjectId
			: projects.table.order[0]) ?? null;

	workspace.selectProject(selectedProjectId);

	let selectedModelId: Id | null = null;
	if (selectedProjectId) {
		const localModels = await options.repository.listModels(selectedProjectId);
		for (const model of localModels) {
			models.upsert(model);
		}
		selectedModelId =
			(metadata?.lastModelId && models.table.byId[metadata.lastModelId]
				? metadata.lastModelId
				: localModels[0]?.id) ?? null;
		workspace.selectModel(selectedModelId);
	}

	workspace.hydrate();
	return {
		authenticated,
		offline: !online,
		loginRequired,
		localMetadataCorrupt,
		selectedProjectId,
		selectedModelId
	};
}
