import type { AuthUser } from '$lib/types';

export type AuthMeResponse =
	| {
			status: 'authenticated';
			user: AuthUser;
			sessionExpiresAt?: string;
			csrfToken?: string;
	  }
	| { status: 'anonymous' };

export interface AuthClient {
	me(): Promise<AuthMeResponse>;
}

export function createAuthClient(fetcher: typeof fetch = fetch): AuthClient {
	return {
		async me() {
			const response = await fetcher('/auth/me', {
				credentials: 'include',
				headers: { accept: 'application/json' }
			});

			if (response.status === 401) {
				return { status: 'anonymous' };
			}
			if (!response.ok) {
				throw new Error(`/auth/me failed with ${response.status}`);
			}
			return (await response.json()) as AuthMeResponse;
		}
	};
}
