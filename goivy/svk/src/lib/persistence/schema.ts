export const SVK_DB_VERSION = 1;

export const SVK_STORES = [
	'projects',
	'models',
	'model_revisions',
	'jobs',
	'check_results',
	'graph_snapshots',
	'concept_states',
	'trace_event_sheets',
	'sync_ops'
] as const;

export type SvkStoreName = (typeof SVK_STORES)[number];

type StoreIndex = {
	name: string;
	keyPath: string | string[];
	options?: IDBIndexParameters;
};

export type StoreDefinition = {
	name: SvkStoreName;
	options: IDBObjectStoreParameters;
	indexes: StoreIndex[];
};

export const SVK_STORE_DEFINITIONS: StoreDefinition[] = [
	{
		name: 'projects',
		options: { keyPath: 'id' },
		indexes: [{ name: 'by_account', keyPath: 'accountId' }]
	},
	{
		name: 'models',
		options: { keyPath: 'id' },
		indexes: [
			{ name: 'by_project', keyPath: 'projectId' },
			{ name: 'by_project_filename', keyPath: ['projectId', 'filename'], options: { unique: true } }
		]
	},
	{
		name: 'model_revisions',
		options: { keyPath: 'id' },
		indexes: [
			{ name: 'by_project', keyPath: 'projectId' },
			{ name: 'by_model_revision', keyPath: ['modelId', 'revision'], options: { unique: true } }
		]
	},
	{
		name: 'jobs',
		options: { keyPath: 'id' },
		indexes: [
			{ name: 'by_project', keyPath: 'projectId' },
			{ name: 'by_model', keyPath: 'modelId' },
			{ name: 'by_status', keyPath: 'status' }
		]
	},
	{
		name: 'check_results',
		options: { keyPath: 'id' },
		indexes: [
			{ name: 'by_job', keyPath: 'jobId', options: { unique: true } },
			{ name: 'by_session', keyPath: 'sessionId' }
		]
	},
	{
		name: 'graph_snapshots',
		options: { keyPath: 'id' },
		indexes: [
			{ name: 'by_sheet', keyPath: 'sheetId' },
			{ name: 'by_kind', keyPath: 'kind' }
		]
	},
	{
		name: 'concept_states',
		options: { keyPath: 'id' },
		indexes: [
			{ name: 'by_session', keyPath: 'sessionId' },
			{ name: 'by_sheet', keyPath: 'sheetId' }
		]
	},
	{
		name: 'trace_event_sheets',
		options: { keyPath: 'id' },
		indexes: [
			{ name: 'by_project', keyPath: 'projectId' },
			{ name: 'by_session', keyPath: 'sessionId' }
		]
	},
	{
		name: 'sync_ops',
		options: { keyPath: 'id' },
		indexes: [
			{ name: 'by_project', keyPath: 'projectId' },
			{ name: 'by_status', keyPath: 'status' },
			{ name: 'by_project_status', keyPath: ['projectId', 'status'] },
			{ name: 'by_created_at', keyPath: 'createdAt' }
		]
	}
];
