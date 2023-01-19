<script lang="ts">
  import { Router, Link, Route, navigate } from "svelte-routing";
  import Home from "./lib/routes/Home.svelte";
  import Config from "./lib/routes/Config.svelte";
  import NotFound from "./lib/routes/NotFound.svelte";
  import { onMount } from "svelte";

  export let url = "";

  let pathname = '';
  onMount(() => pathname = window.location.pathname);
</script>

<main>
  <Router url="{url}">
      <Route path="/"><Home /></Route>
      <Route path="/config"><Config /></Route>
      <Route path="*"><NotFound /></Route>
      <div class="navigation">
        {#if pathname != '/'}
          <button on:click={() => {
            pathname = '/'
            navigate('/')
          }}>Logs</button>
        {:else}
          <button on:click={() => {
            pathname = '/config'
            navigate('/config')
          }}>Config</button>
        {/if}
      </div>
  </Router>
</main>

<style lang="scss">
  main {
    margin-bottom: 3em;
  }
  .navigation {
    position: fixed;
    bottom: 0.5em;
    display: flex;
    justify-content: center;
    width: 100vw;
    button {
      padding: 0.5em;
      background-color: #fff;
      border: 1px solid #222;
      border-radius: 1em;
      color: #222;      
      cursor: pointer;
      &:hover {
        background-color: rgba(0, 0, 0, 0.3);
      }
    }
  }
</style>