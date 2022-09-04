<script lang="ts">
import { Button, Dialog, Modal } from "attractions";

import DailyLog from "./lib/components/DailyLog.svelte";
import DaySelector from "./lib/components/DaySelector.svelte";
import LogEntryInterface from "./lib/components/LogEntryInterface.svelte";

let modalOpen = false;
let logDay = new Date();

const dateWithCurrentTime = (date: Date) => {
  return new Date(date.toISOString().split('T')[0] + 'T' + (new Date()).toISOString().split('T')[1])
}

</script>

<main>
    <DaySelector day={logDay} on:dateChange={(event) => logDay = event.detail} />
    <Button on:click={() => modalOpen = true}>Add Log</Button>
    <Modal bind:open={modalOpen} let:closeCallback>
      <Dialog title="Hello?" {closeCallback}>
        {#if modalOpen}
            <LogEntryInterface logTime={dateWithCurrentTime(logDay)} on:success={() => closeCallback()}/>
        {/if}
      </Dialog>
    </Modal>
    <DailyLog day={logDay} />
</main>

<style>

</style>