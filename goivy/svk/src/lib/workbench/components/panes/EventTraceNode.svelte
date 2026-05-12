<script lang="ts">
	import type { TraceEvent } from '$lib/types';
	import EventTraceNode from './EventTraceNode.svelte';

	type Props = {
		node: TraceEvent;
	};

	let { node }: Props = $props();
</script>

<li class="event-tree-node">
	<div class="event-row">{node.text}</div>
	{#if (node.children ?? []).length > 0}
		<ul class="event-tree-list">
			{#each node.children ?? [] as child (child.address)}
				<EventTraceNode node={child} />
			{/each}
		</ul>
	{/if}
</li>
