export type PasskeyOptionsResponse = {
	challenge: string;
};

export function base64UrlToBytes(value: string): Uint8Array {
	const padded = value.replace(/-/g, '+').replace(/_/g, '/').padEnd(Math.ceil(value.length / 4) * 4, '=');
	const binary = atob(padded);
	const bytes = new Uint8Array(binary.length);
	for (let index = 0; index < binary.length; index += 1) {
		bytes[index] = binary.charCodeAt(index);
	}
	return bytes;
}

export function bytesToBase64Url(bytes: Uint8Array): string {
	let binary = '';
	for (const byte of bytes) {
		binary += String.fromCharCode(byte);
	}
	return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

export function bytesToArrayBuffer(bytes: Uint8Array): ArrayBuffer {
	const copy = new ArrayBuffer(bytes.byteLength);
	new Uint8Array(copy).set(bytes);
	return copy;
}

export async function createPasskeyCredential(userId: string, challenge: string): Promise<Credential | null> {
	if (!navigator.credentials?.create) {
		throw new Error('Passkeys are unavailable in this browser');
	}
	return navigator.credentials.create({
		publicKey: {
			challenge: bytesToArrayBuffer(base64UrlToBytes(challenge)),
			rp: { name: 'SVK' },
			user: {
				id: bytesToArrayBuffer(new TextEncoder().encode(userId)),
				name: userId,
				displayName: userId
			},
			pubKeyCredParams: [{ type: 'public-key', alg: -7 }],
			authenticatorSelection: {
				residentKey: 'preferred',
				userVerification: 'preferred'
			},
			timeout: 60_000
		}
	});
}

export async function getPasskeyCredential(challenge: string): Promise<Credential | null> {
	if (!navigator.credentials?.get) {
		throw new Error('Passkeys are unavailable in this browser');
	}
	return navigator.credentials.get({
		publicKey: {
			challenge: bytesToArrayBuffer(base64UrlToBytes(challenge)),
			userVerification: 'preferred',
			timeout: 60_000
		}
	});
}
