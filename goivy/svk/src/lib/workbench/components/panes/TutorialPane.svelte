<script lang="ts">
	type Props = {
		visible?: boolean;
		url?: string;
	};

	let { visible = false, url = '/static/tutorial/kenmcmil.github.io/ivy/language.html' }: Props = $props();
</script>

{#if visible}
	<section id="tutorial-container" aria-label="Tutorial">
		<div class="panel-header tutorial-url-bar">
			<button id="tutorial-back" type="button" class="tutorial-nav-btn">Back</button>
			<button id="tutorial-forward" type="button" class="tutorial-nav-btn">Forward</button>
			<input class="tutorial-url-input" value={url} aria-label="Tutorial URL" />
			<button id="tutorial-close" type="button" class="tutorial-close-btn">x</button>
		</div>
		<iframe id="tutorial-iframe" class="tutorial-iframe" title="Ivy tutorial" src={url}></iframe>
	</section>
{/if}
