import type { CheckMode, CheckResult } from '$lib/types';
import { asBoolean, asString, asStringArray, nowIso, stableId } from './primitives';

export type CheckNormalizeOptions = {
	jobId: string;
	sessionId: string;
	mode: CheckMode;
	now?: Date;
};

export function normalizeCheckResult(payload: unknown, options: CheckNormalizeOptions): CheckResult {
	const data = payload && typeof payload === 'object' ? (payload as Record<string, unknown>) : {};
	const result = asString(data.result);
	return {
		id: stableId('check', options.jobId),
		jobId: options.jobId,
		sessionId: options.sessionId,
		mode: options.mode,
		z3Contacted: asBoolean(data.z3_contacted),
		result: result === 'pass' || result === 'fail' || result === 'error' ? result : 'error',
		message: asString(data.message),
		failedConjecture: asString(data.failed_conjecture) || undefined,
		failedLabel: asString(data.failed_label) || undefined,
		usedRelations: asStringArray(data.used_relations),
		counterexampleTrace: asString(data.counterexample_trace) || undefined,
		counterexampleDetails: asString(data.counterexample_details) || undefined,
		createdAt: nowIso(options.now)
	};
}
