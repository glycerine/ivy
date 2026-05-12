<script lang="ts">
	import type { CheckResult, GraphNode, VerificationJob } from '$lib/types';

	type Props = {
		latestCheck: CheckResult | null;
		selectedNode: GraphNode | null;
		jobs: VerificationJob[];
	};

	let { latestCheck, selectedNode, jobs }: Props = $props();

	function formatLongInfo(value: unknown) {
		return typeof value === 'string' ? value : JSON.stringify(value, null, 2);
	}
</script>

<section class="pane details-pane" aria-label="Details and checks">
	<div class="pane-title compact-title">
		<strong>DETAILS</strong>
	</div>
	<div class="details-body" data-testid="details-pane">
		<p>Verification Result</p>
		{#if latestCheck}
			<p>{latestCheck.result.toUpperCase()} [Z3: {latestCheck.z3Contacted ? 'yes' : 'no'}]: {latestCheck.message}</p>
			{#if latestCheck.failedConjecture}
				<p>{latestCheck.failedConjecture}</p>
			{/if}
			{#if latestCheck.counterexampleTrace}
				<p>{latestCheck.counterexampleTrace}</p>
			{/if}
		{:else}
			<p>No verification result yet.</p>
		{/if}
		{#if selectedNode}
			<h2>{selectedNode.label}</h2>
			{#if selectedNode.shortInfo}
				<p>{selectedNode.shortInfo}</p>
			{/if}
			<p>{selectedNode.obj}</p>
			{#if selectedNode.longInfo}
				<pre>{formatLongInfo(selectedNode.longInfo)}</pre>
			{:else}
				<p>{selectedNode.classes.join(', ')}</p>
			{/if}
		{/if}
		{#if latestCheck?.counterexampleTrace}
			<button type="button" class="trace-action">View error trace</button>
		{/if}
	</div>
	<div class="job-strip" data-testid="job-strip">
		{#each jobs as job (job.id)}
			<div class="job-row">
				<span>{job.kind}</span>
				<strong>{job.status}</strong>
			</div>
		{/each}
	</div>
</section>
