<script lang="ts">
	import type { ContextMenuItem } from '$lib/workbench/state';

	type Props = {
		visible?: boolean;
		x?: number;
		y?: number;
		items?: ContextMenuItem[];
		onSelect?: (item: ContextMenuItem) => void;
	};

	let { visible = false, x = 0, y = 0, items = [], onSelect }: Props = $props();
</script>

{#if visible}
	<nav class="context-menu" style={`left:${x}px;top:${y}px`} aria-label="Context menu">
		{#each items as item (item.id)}
			{#if item.separator}
				<div class="context-menu-separator"></div>
			{:else}
				<button type="button" disabled={item.enabled === false} onclick={() => onSelect?.(item)}>{item.label}</button>
			{/if}
		{/each}
	</nav>
{/if}
