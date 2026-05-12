<script lang="ts">
	type Props = {
		visible?: boolean;
		url?: string;
		input?: string;
		canGoBack?: boolean;
		canGoForward?: boolean;
		frameKey?: number;
		onSetInput?: (value: string) => void;
		onNavigate?: (url: string) => void;
		onBack?: () => void;
		onForward?: () => void;
		onReload?: () => void;
		onClose?: () => void;
	};

	let {
		visible = false,
		url = '/static/tutorial/kenmcmil.github.io/ivy/language.html',
		input = url,
		canGoBack = false,
		canGoForward = false,
		frameKey = 0,
		onSetInput = () => {},
		onNavigate = () => {},
		onBack = () => {},
		onForward = () => {},
		onReload = () => {},
		onClose = () => {}
	}: Props = $props();

	function handleUrlKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter') {
			onNavigate((event.currentTarget as HTMLInputElement | null)?.value ?? '');
		}
	}
</script>

{#if visible}
	<section id="tutorial-container" aria-label="Tutorial">
		<div class="panel-header tutorial-url-bar">
			<button id="tutorial-back" type="button" class="tutorial-nav-btn" title="Back" disabled={!canGoBack} onclick={onBack}>Back</button>
			<button id="tutorial-fwd" type="button" class="tutorial-nav-btn" title="Forward" disabled={!canGoForward} onclick={onForward}>Forward</button>
			<button id="tutorial-reload" type="button" class="tutorial-nav-btn" title="Reload" onclick={onReload}>Reload</button>
			<input
				id="tutorial-url"
				class="tutorial-url-input"
				value={input}
				aria-label="Tutorial URL"
				spellcheck="false"
				oninput={(event) => onSetInput(event.currentTarget.value)}
				onkeydown={handleUrlKeydown}
			/>
			<button id="tutorial-close" type="button" class="tutorial-close-btn" title="Close tutorial" onclick={onClose}>x</button>
		</div>
		{#key frameKey}
			<iframe id="tutorial-iframe" class="tutorial-iframe" title="Ivy tutorial" src={url}></iframe>
		{/key}
	</section>
{/if}
