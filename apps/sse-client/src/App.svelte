<script lang="ts">
    import svelteLogo from "./assets/svelte.svg";
    import viteLogo from "/vite.svg";
    import Counter from "./lib/Counter.svelte";
    import { onDestroy } from "svelte";
    import Keydown from "svelte-keydown";

    let time = "";
    let id = "";
    let eventSource: EventSource;
    let port = 5050;
    async function getTime() {
        const message = Date.now().toString();
        const body = JSON.stringify({
            id,
            email: "test@gmail.com",
            name: "Duong dep try",
        });
        const res = await fetch(`http://localhost:3000/time/${id}`, {
            method: "POST",
            body,
        });
        if (res.status !== 200) {
            console.log("Could not connect to the server");
        } else {
            console.log("OK");
        }
    }

    function onEventSourceMessage(event: MessageEvent<any>) {
        time = event.data;
        console.log({ event });
    }

    function onEventSourceError(e: Event) {
        console.error(e);
    }

    function handleClick() {
        eventSource = new EventSource(`http://localhost:3000/event/${id}`);
        eventSource.onmessage = onEventSourceMessage;
        eventSource.onerror = onEventSourceError;
    }

    function disconnect() {
        eventSource.close();
    }

    onDestroy(disconnect);
</script>

<Keydown on:Enter={disconnect} />
<main>
    <button on:click={getTime}>Get Time</button>
    <p>Time: {time}</p>
    <div>
        <a href="https://vitejs.dev" target="_blank" rel="noreferrer">
            <img src={viteLogo} class="logo" alt="Vite Logo" />
        </a>
        <a href="https://svelte.dev" target="_blank" rel="noreferrer">
            <img src={svelteLogo} class="logo svelte" alt="Svelte Logo" />
        </a>
    </div>
    <h1>Vite + Svelte</h1>
    <label for="id">Id</label>
    <input bind:value={id} type="text" />
    <button on:click={handleClick}>Connect</button>
    <button on:click={disconnect}>Disconnect</button>
    <label for="port">Port</label>
    <input bind:value={port} type="number" />

    <div class="card">
        <Counter />
    </div>

    <p>
        Check out <a
            href="https://github.com/sveltejs/kit#readme"
            target="_blank"
            rel="noreferrer">SvelteKit</a
        >, the official Svelte app framework powered by Vite!
    </p>

    <p class="read-the-docs">
        Click on the Vite and Svelte logos to learn more
    </p>
</main>

<style>
    .logo {
        height: 6em;
        padding: 1.5em;
        will-change: filter;
        transition: filter 300ms;
    }
    .logo:hover {
        filter: drop-shadow(0 0 2em #646cffaa);
    }
    .logo.svelte:hover {
        filter: drop-shadow(0 0 2em #ff3e00aa);
    }
    .read-the-docs {
        color: #888;
    }
</style>
