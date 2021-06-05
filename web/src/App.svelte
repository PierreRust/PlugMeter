<script>
	import { onMount } from "svelte";
	import Plug from './Plug.svelte';
	export let name;

	async function getPlugs() {
		let target = "http://127.0.0.1:3000/api/v1/plugs";
		const res = await fetch(target);
		const plugs = await res.json();

		if (res.ok) {
			return plugs;
		} else {
			throw new Error(res.text);
		}
	}
	// let plugs_promise  = getPlugs();
	let plugs = [];

	onMount(() => {
		async function fetchPlugs() {
			console.log("Updating plug list");
			plugs = await getPlugs();
		}
		fetchPlugs();
		const interval = setInterval(fetchPlugs, 3000);

		return () => clearInterval(interval);
	});
</script>

<main>
	<h1>Plug metering</h1>

	<p>{plugs.length} Plugs detected</p>

	<div class="plugs">
		{#each plugs as plug (plug.Id)}
			<Plug plug={plug} />
		{/each}
	</div>
</main>

<style>
	main {
		/* text-align: center; */
		padding: 1em;
		/* max-width: 240px; */
		margin: 0 auto;
	}

	h1 {
		color: #ff3e00;
		text-transform: uppercase;
		font-size: 4em;
		font-weight: 100;
	}

	/* @media (min-width: 240px) {
		main {
			max-width: none;
		}
	} */

	div.plugs {
		display: flex;
	}
	div.plug {
		border: 1px red solid;
	}
</style>
