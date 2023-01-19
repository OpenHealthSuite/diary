<script lang="ts">

import MetricsConfiguration from "../components/configuration/MetricsConfiguration.svelte";
import BulkLogUpload from "../components/configuration/BulkLogUpload.svelte";
import { apiFetch } from "../utilities";
import PurgeLogs from "../components/configuration/PurgeLogs.svelte";

let logoutEndpoint = undefined;

let [uploadOpen, purgeOpen] = [false, false];

apiFetch("/logout-endpoint").then(res => {
  res.json().then(({ url }: { url: string }) => {
    logoutEndpoint = url;
  })
})
</script>
<h1 style="width: 100%; text-align: center;">Config</h1>
<div class="container">
  <div class="config-container">
    <h2>Metrics</h2>
    <div class="config-content">
      <MetricsConfiguration />
    </div>
  </div>
  <div class="config-container">
      <h2>Data Management</h2>
      <div class="config-content button-stack">
        <a href="/api/logs/export" target="_blank" download><button>Download Logs</button></a>
        <button on:click={() => {
          uploadOpen = true;
          purgeOpen = false;
        }}>Upload Logs</button>
        <button on:click={() => {
          uploadOpen = false;
          purgeOpen = true;
        }}>Purge Logs</button>
      </div>

      <BulkLogUpload bind:modalOpen={uploadOpen}/>
      <PurgeLogs bind:modalOpen={purgeOpen}/>
  </div>
  <div class="config-container">
      <h2>User Actions</h2>
      <div class="config-content button-stack">
        {#if logoutEndpoint != undefined}
        <a href={logoutEndpoint}><button>Logout</button></a>
        {/if}
      </div>
  </div>
</div>
<style lang="scss">
  .container {
    width: 100vw;
    display: flex;
    flex-direction: row;
    justify-content: center;
    flex-wrap: wrap;
  }
  .config-container {
    max-width: 100vw;
    width: 350px;
    margin: 1em;
    padding: 0 1em 1em 1em;
    border-radius: 1em;
    border: 1px solid #000;
    h2 {
      width: 100%;
      text-align: center;
    }
    .config-content {
      width: 100%;
      display: flex;
      flex-direction: column;
      align-items: center;
      &.button-stack > * {
        margin: 0.25em 0;
      }
      button, a {
        width: 100%;
      }
    }
  }
</style>