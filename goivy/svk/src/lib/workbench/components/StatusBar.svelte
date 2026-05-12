<script lang="ts">
	import type { CheckResult } from '$lib/types';

	type Props = {
		latestCheck: CheckResult | null;
		sessionLabel: string;
		message: string;
		level?: '' | 'info' | 'success' | 'warning' | 'error';
	};

	let { latestCheck, sessionLabel, message, level = '' }: Props = $props();
</script>

<footer class="statusbar" class:info={level === 'info'} class:success={level === 'success'} class:warning={level === 'warning'} class:error={level === 'error'} data-testid="status-strip">
	<span>
		{#if latestCheck}
			Check {latestCheck.result.toUpperCase()} ({latestCheck.mode}) [Z3: {latestCheck.z3Contacted ? 'yes' : 'no'}] - {latestCheck.message}
		{:else}
			{message}
		{/if}
	</span>
	<span>Session: {sessionLabel}</span>
</footer>
