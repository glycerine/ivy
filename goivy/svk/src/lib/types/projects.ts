import type { Id } from './ids';

export type AccountKind = 'personal' | 'corporate';
export type BillingStatus = 'trial' | 'active' | 'past_due' | 'disabled';
export type ProjectRole = 'read' | 'write' | 'admin' | 'owner';

export type Account = {
	id: Id;
	name: string;
	slug: string;
	kind: AccountKind;
	billingStatus: BillingStatus;
	createdAt: string;
};

export type Team = {
	id: Id;
	accountId: Id;
	name: string;
	slug: string;
	createdAt: string;
};

export type Project = {
	id: Id;
	accountId: Id;
	ownerKind: 'user' | 'team' | 'account';
	ownerId: Id;
	name: string;
	slug: string;
	storageMode: 'shared_postgres' | 'local_indexeddb';
	createdAt: string;
	updatedAt: string;
};

export type ProjectGrant = {
	id: Id;
	projectId: Id;
	subjectKind: 'user' | 'team' | 'account_users';
	subjectId: Id;
	role: ProjectRole;
};
