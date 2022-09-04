<script lang="ts">
    import { Accordion, AccordionSection, Button, Card, Dialog, Modal } from 'attractions';
    import { apiFetch } from 'src/lib/utilities/index'
    import type { FoodLogEntry } from '../types/FoodLogEntry';
    import LogEntryInterface from './LogEntryInterface.svelte';
    export let day: Date;

    let loading = true;
    let error = false;
    let modalOpen = false;
    let editingLog: FoodLogEntry;

    let dayData: FoodLogEntry[] = [];

    const getStartDate = (inDate: Date) => {
        return new Date(inDate.toISOString().split("T")[0]);
    }

    const getEndDate = (inDate: Date) => {
        const endDate = new Date(inDate.toISOString().split("T")[0])
        endDate.setDate(endDate.getDate() + 1);
        return endDate;
    }

    function updateData(inputDay: Date) {
        loading = true;
        error = false;
        apiFetch(`/logs?startDate=${getStartDate(inputDay).toISOString()}&endDate=${getEndDate(inputDay).toISOString()}`)
            .then(res => res.status === 200 ? res.json() : new Error())
            .then(data => {
                dayData = data.sort((a, b) => new Date(a.time.start).getTime() - new Date(b.time.start).getTime())
            })
            .catch(() => error = true)
            .finally(() => loading = false)
    }

    $: updateData(day)
</script>

<Card>
    {#if loading}
    <div data-testid="loading-indicator">
        Loading data
    </div>
    {:else if error}
    <div data-testid="error-indicator">
        Error loading data
    </div>
    {:else}
    {#if dayData.length === 0}
    <div>
        No Logs Entered for this day
    </div>
    {:else}
    <div>
        {#each dayData as log, i}
            <Button on:click={() => {
                modalOpen = true;
                editingLog = log;
            }} outline>
            <span data-testid="foodlog-{i}">{log.name}</span> - <span data-testid="foodlog-{i}-calories">{log.metrics['calories']}</span> Calories
            </Button>
        {/each}
    </div>
    {/if}
    {/if}
</Card>
<Modal bind:open={modalOpen} let:closeCallback>
    <Dialog title="Edit Calorie Log" {closeCallback}>
      {#if modalOpen}
          <LogEntryInterface log={editingLog} on:success={() => {
            closeCallback()
            updateData(day)
            }}
            on:error={(event) => console.error(event.detail)}/>
      {/if}
    </Dialog>
</Modal>

<style>

</style>