<script lang="ts">
    import Modal from './Modal.svelte';
    import { apiFetch } from 'src/lib/utilities/index';
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

<div>
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
        <h1>Total: {dayData.reduce((prev, curr) => prev + curr.metrics.calories, 0)} Calories</h1>
        {#each dayData as log, i}
        <div class="food-log {i > 0 ? 'top-border' : ''}">
            <h2 data-testid="foodlog-{i}-calories">{log.metrics['calories']} Calories</h2>
            <h5 data-testid="foodlog-{i}">{new Date(log.time.start).toTimeString().split(' ')[0]} - {log.name}</h5>
            <button class='log-button'
                on:click={() => {
                    modalOpen = true;
                    editingLog = log;
                }}>
                Edit
            </button>
        </div>
        {/each}
    </div>
    {/if}
    {/if}
</div>


<Modal bind:open={modalOpen}>
    {#if modalOpen}
        <LogEntryInterface log={editingLog} on:success={() => {
        modalOpen = false
        updateData(day)
        }}
        on:error={(event) => console.error(event.detail)}/>
    {/if}
</Modal>

<style lang="scss">
    .food-log {
        position: relative;
        display: flex;
        flex-direction: column;
        padding: 1em;
        h2 {
            line-height: 0;
        }
        h5 {
            line-height: 0;
        }
        button {
            position: absolute;
            top: 1.5em;
            right: 1em;
            border-radius: 1em;
            padding: 0.6em;
            font-weight: 700;
            cursor: pointer;
            &:hover {
                background-color: lightgray;
            }
        }
        &.top-border {
            border-top: 2px lightgray solid;
        }
    }
</style>