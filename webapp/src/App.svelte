<script lang="ts">
import { Button, Dialog, Modal } from "attractions";

import DailyLog from "./lib/components/DailyLog.svelte";
import LogEntryInterface from "./lib/components/LogEntryInterface.svelte";

let modalOpen = false;

</script>

<main>
    <DailyLog day={new Date()} />
    <Button on:click={() => modalOpen = true}>Add Log</Button>
    <Modal bind:open={modalOpen} let:closeCallback>
      <Dialog title="Hello?" {closeCallback}>
        {#if modalOpen}
            <LogEntryInterface on:success={closeCallback}/>
        {/if}
      </Dialog>
    </Modal>

</main>

<style>

</style>