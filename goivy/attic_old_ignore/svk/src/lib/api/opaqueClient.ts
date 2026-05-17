import { OpaqueClient, OpaqueID, getOpaqueConfig } from '@cloudflare/opaque-ts';

export type OpaqueFlowResponse = {
	flowId?: string;
	message?: string;
	complete?: boolean;
};

export type OpaqueHttpClient = {
	registrationStart(userId: string, message: Uint8Array): Promise<OpaqueFlowResponse>;
	registrationFinish(flowId: string, message: Uint8Array): Promise<OpaqueFlowResponse>;
	loginStart(userId: string, message: Uint8Array): Promise<OpaqueFlowResponse>;
	loginFinish(flowId: string, message: Uint8Array): Promise<OpaqueFlowResponse>;
};

export function createCloudflareOpaqueClient(opaqueId: OpaqueID = OpaqueID.OPAQUE_P256) {
	return new OpaqueClient(getOpaqueConfig(opaqueId));
}

export function createOpaqueHttpClient(fetcher: typeof fetch = fetch): OpaqueHttpClient {
	return {
		registrationStart(userId, message) {
			return postOpaque(fetcher, '/auth/opaque/register/start', { userId, message: bytesToBase64(message) });
		},
		registrationFinish(flowId, message) {
			return postOpaque(fetcher, '/auth/opaque/register/finish', { flowId, message: bytesToBase64(message) });
		},
		loginStart(userId, message) {
			return postOpaque(fetcher, '/auth/opaque/login/start', { userId, message: bytesToBase64(message) });
		},
		loginFinish(flowId, message) {
			return postOpaque(fetcher, '/auth/opaque/login/finish', { flowId, message: bytesToBase64(message) });
		}
	};
}

export function bytesToBase64(bytes: Uint8Array): string {
	let binary = '';
	for (const byte of bytes) {
		binary += String.fromCharCode(byte);
	}
	return btoa(binary);
}

export function base64ToBytes(value: string): Uint8Array {
	const binary = atob(value);
	const bytes = new Uint8Array(binary.length);
	for (let index = 0; index < binary.length; index += 1) {
		bytes[index] = binary.charCodeAt(index);
	}
	return bytes;
}

async function postOpaque(fetcher: typeof fetch, path: string, body: Record<string, string>): Promise<OpaqueFlowResponse> {
	if (Object.hasOwn(body, 'password')) {
		throw new Error('OPAQUE transport body must not contain raw passwords');
	}
	const response = await fetcher(path, {
		method: 'POST',
		credentials: 'include',
		headers: { accept: 'application/json', 'content-type': 'application/json' },
		body: JSON.stringify(body)
	});
	if (!response.ok) {
		throw new Error(`OPAQUE request failed with ${response.status}`);
	}
	return (await response.json()) as OpaqueFlowResponse;
}
