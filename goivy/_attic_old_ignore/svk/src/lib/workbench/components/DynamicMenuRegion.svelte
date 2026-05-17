<script lang="ts">
	import type { MenuDescriptorItem } from '$lib/workbench/state';

	type Props = {
		region: string;
		items?: MenuDescriptorItem[];
		onDispatch?: (region: string, item: MenuDescriptorItem) => void | Promise<void>;
	};

	let { region, items = [], onDispatch }: Props = $props();
</script>

{#if items.length > 0}
	<div class="dynamic-menu-root" data-menu-region={region}>
		{#each items as item (item.id)}
			{#if item.separator}
				<div class="dropdown-sep"></div>
			{:else}
				<button type="button" class:disabled={item.enabled === false} disabled={item.enabled === false} onclick={() => void onDispatch?.(region, item)}>
					{item.label}
				</button>
			{/if}
		{/each}
	</div>
{/if}
