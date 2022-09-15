<script lang="ts">

import DailyLog from "./lib/components/DailyLog.svelte";
import DaySelector from "./lib/components/DaySelector.svelte";
import LogEntryInterface from "./lib/components/LogEntryInterface.svelte";
import Modal from "./lib/components/Modal.svelte";

let modalOpen = false;
let logDay = new Date();

const dateChange = (event) => logDay = event.detail;

const dateWithCurrentTime = (date: Date) => {
  return new Date(date.toISOString().split('T')[0] + 'T' + (new Date()).toISOString().split('T')[1])
}

</script>

<main>
    <DaySelector day={logDay} on:dateChange={dateChange} />
    <div class="controls-row">
      <button on:click={() => modalOpen = true}>Add Log</button>
    </div>

    <Modal bind:open={modalOpen}>
      {#if modalOpen}
      <div class="form-wrapper">
        <LogEntryInterface logTime={dateWithCurrentTime(logDay)} on:success={() => {
          modalOpen = false
          logDay = logDay
          }}
          on:error={(event) => console.error(event.detail)}/>
      </div>
      {/if}
    </Modal>

    <DailyLog day={logDay} />
</main>

<style>
  .controls-row {
    margin: 1em;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .form-wrapper {
    max-height: 70vh;
    overflow-y: scroll;
  }
</style>