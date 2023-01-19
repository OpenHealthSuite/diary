<script lang="ts">
import Modal from "../Modal.svelte";
import { apiFetch } from "src/lib/utilities";

let modalOpen = false;

let purging = false;
let error: string | undefined = undefined;

const purgeLogs = () => {
  purging = true;
  apiFetch("/logs", { method: "DELETE"}).then((res) => {
    if (res.status === 200 || res.status === 204) {
      purging = false;
      modalOpen = false;
    } else {
      purging = false;
      res.text().then(err => {
        error = err;
      })
    }
  })
}

</script>

<button on:click={() => modalOpen = true}>Purge Logs</button>

<Modal bind:open={modalOpen}>
  {#if error !== undefined}
    <p>Something went wrong purging logs - please try again.</p>
  {/if}
  {#if !purging}
    <h1>Warning!</h1>
    <p>This will <em>irretrievably</em> purge your food logs from the system.</p>
    <button on:click={purgeLogs}>Purge Logs</button>
  {/if}
  {#if purging}
    <p>Deleting all log data...</p>
  {/if}
</Modal>

<style lang="scss">

</style>