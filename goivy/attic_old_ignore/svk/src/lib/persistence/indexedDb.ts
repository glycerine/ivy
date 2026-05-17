import { SVK_DB_VERSION, SVK_STORE_DEFINITIONS, type SvkStoreName } from './schema';

export type SvkDatabaseOptions = {
	name: string;
	version?: number;
};

export class IndexedDbUnavailableError extends Error {
	constructor() {
		super('IndexedDB is unavailable in this runtime');
		this.name = 'IndexedDbUnavailableError';
	}
}

export function openSvkDatabase(options: SvkDatabaseOptions): Promise<IDBDatabase> {
	if (!globalThis.indexedDB) {
		throw new IndexedDbUnavailableError();
	}

	return new Promise((resolve, reject) => {
		const request = globalThis.indexedDB.open(options.name, options.version ?? SVK_DB_VERSION);

		request.onupgradeneeded = () => {
			const db = request.result;
			for (const definition of SVK_STORE_DEFINITIONS) {
				const store = db.objectStoreNames.contains(definition.name)
					? request.transaction?.objectStore(definition.name)
					: db.createObjectStore(definition.name, definition.options);

				if (!store) {
					continue;
				}

				for (const index of definition.indexes) {
					if (!store.indexNames.contains(index.name)) {
						store.createIndex(index.name, index.keyPath, index.options);
					}
				}
			}
		};

		request.onsuccess = () => resolve(request.result);
		request.onerror = () => reject(request.error);
		request.onblocked = () => reject(new Error(`Opening IndexedDB database "${options.name}" was blocked`));
	});
}

export function deleteSvkDatabase(name: string): Promise<void> {
	if (!globalThis.indexedDB) {
		throw new IndexedDbUnavailableError();
	}

	return new Promise((resolve, reject) => {
		const request = globalThis.indexedDB.deleteDatabase(name);
		request.onsuccess = () => resolve();
		request.onerror = () => reject(request.error);
		request.onblocked = () => reject(new Error(`Deleting IndexedDB database "${name}" was blocked`));
	});
}

function promisifyRequest<T>(request: IDBRequest<T>): Promise<T> {
	return new Promise((resolve, reject) => {
		request.onsuccess = () => resolve(request.result);
		request.onerror = () => reject(request.error);
	});
}

function promisifyTransaction(transaction: IDBTransaction): Promise<void> {
	return new Promise((resolve, reject) => {
		transaction.oncomplete = () => resolve();
		transaction.onerror = () => reject(transaction.error);
		transaction.onabort = () => reject(transaction.error ?? new Error('IndexedDB transaction aborted'));
	});
}

export class SvkIndexedDb {
	constructor(private readonly db: IDBDatabase) {}

	close() {
		this.db.close();
	}

	put<T>(storeName: SvkStoreName, value: T): Promise<void> {
		const transaction = this.db.transaction(storeName, 'readwrite');
		transaction.objectStore(storeName).put(value);
		return promisifyTransaction(transaction);
	}

	delete(storeName: SvkStoreName, id: IDBValidKey): Promise<void> {
		const transaction = this.db.transaction(storeName, 'readwrite');
		transaction.objectStore(storeName).delete(id);
		return promisifyTransaction(transaction);
	}

	clear(storeName: SvkStoreName): Promise<void> {
		const transaction = this.db.transaction(storeName, 'readwrite');
		transaction.objectStore(storeName).clear();
		return promisifyTransaction(transaction);
	}

	get<T>(storeName: SvkStoreName, id: IDBValidKey): Promise<T | undefined> {
		const transaction = this.db.transaction(storeName, 'readonly');
		return promisifyRequest<T | undefined>(transaction.objectStore(storeName).get(id));
	}

	getByIndex<T>(storeName: SvkStoreName, indexName: string, key: IDBValidKey | IDBKeyRange): Promise<T[]> {
		const transaction = this.db.transaction(storeName, 'readonly');
		return promisifyRequest<T[]>(transaction.objectStore(storeName).index(indexName).getAll(key));
	}

	getAll<T>(storeName: SvkStoreName): Promise<T[]> {
		const transaction = this.db.transaction(storeName, 'readonly');
		return promisifyRequest<T[]>(transaction.objectStore(storeName).getAll());
	}
}

export async function createSvkIndexedDb(options: SvkDatabaseOptions): Promise<SvkIndexedDb> {
	return new SvkIndexedDb(await openSvkDatabase(options));
}
