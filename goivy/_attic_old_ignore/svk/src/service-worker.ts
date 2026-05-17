/// <reference lib="webworker" />
import { build, files, version } from '$service-worker';

const worker = self as unknown as ServiceWorkerGlobalScope;
const appCache = `svk-app-${version}`;
const wasmCache = `svk-wasm-${version}`;
const wasmAssets = [
	'/wasm/goivy-check-js.wasm',
	'/wasm/wasm_exec-go1.25.6.js',
	'/wasm/z3-471-api.js',
	'/wasm/z3-471-api.wasm'
];

worker.addEventListener('install', (event) => {
	event.waitUntil(
		Promise.all([
			caches.open(appCache).then((cache) => cache.addAll([...build, ...files])),
			caches.open(wasmCache).then((cache) => cache.addAll(wasmAssets))
		]).then(() => worker.skipWaiting())
	);
});

worker.addEventListener('activate', (event) => {
	event.waitUntil(
		caches
			.keys()
			.then((keys) =>
				Promise.all(keys.filter((key) => key.startsWith('svk-') && key !== appCache && key !== wasmCache).map((key) => caches.delete(key)))
			)
			.then(() => worker.clients.claim())
	);
});

worker.addEventListener('fetch', (event) => {
	if (event.request.method !== 'GET') {
		return;
	}
	const url = new URL(event.request.url);
	if (url.origin !== worker.location.origin) {
		return;
	}
	event.respondWith(cacheFirst(event.request));
});

async function cacheFirst(request: Request): Promise<Response> {
	const cached = await caches.match(request);
	if (cached) {
		return cached;
	}
	const response = await fetch(request);
	if (response.ok) {
		const cache = await caches.open(request.url.includes('/wasm/') ? wasmCache : appCache);
		await cache.put(request, response.clone());
	}
	return response;
}
