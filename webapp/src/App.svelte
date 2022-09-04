<script lang="ts">
import { Button, Dialog, Modal } from "attractions";

import DailyLog from "./lib/components/DailyLog.svelte";
import DaySelector from "./lib/components/DaySelector.svelte";
import LogEntryInterface from "./lib/components/LogEntryInterface.svelte";

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
      <Button on:click={() => modalOpen = true} filled>Add Log</Button>
    </div>
    <Modal bind:open={modalOpen} let:closeCallback>
      <Dialog title="Add Calorie Log" {closeCallback}>
        {#if modalOpen}
            <LogEntryInterface logTime={dateWithCurrentTime(logDay)} on:success={() => {
              closeCallback()
              logDay = logDay
              }}
              on:error={(event) => console.error(event.detail)}/>
        {/if}
      </Dialog>
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
</style>