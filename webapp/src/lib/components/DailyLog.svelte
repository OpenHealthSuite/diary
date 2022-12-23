<script lang="ts">
    import Modal from './Modal.svelte';
    import { apiFetch, DEFAULT_METRICS } from 'src/lib/utilities/index';
    import type { FoodLogEntry } from '../types/FoodLogEntry';
    import LogEntryInterface from './LogEntryInterface.svelte';
    export let day: Date;
    export let metricConfig = DEFAULT_METRICS;

    $: topMetric = Object.entries(metricConfig)
        .sort((a, b) => a[1].priority - b[1].priority)[0]
        
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
<div class="logs-area-container">
    <div class="logs-area">
        {#if dayData.length === 0}
        <div class="raw-text-message">
            No Logs Entered for this day
        </div>
        {:else}
        <div>
            {#if topMetric}
                <h1 class="{topMetric[0]}-total">{dayData.reduce((prev, curr) => prev + (curr.metrics[topMetric[0]] ?? 0), 0).toLocaleString()} {topMetric[1].label} Total</h1>
            {/if}
            {#each dayData as log, i}
            <div class="food-log {i > 0 ? 'top-border' : ''}">
                {#if topMetric}
                    <h2 data-testid="foodlog-{i}-{topMetric[0]}">{log.metrics[topMetric[0]]?.toLocaleString() ?? 0} {topMetric[1].label}</h2>
                {/if}
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
        {#if loading}
        <div data-testid="loading-indicator" class="raw-text-message">
            Loading data
        </div>
        {:else if error}
        <div data-testid="error-indicator" class="raw-text-message">
            Error loading data
        </div>
        {/if}
    </div>
</div>



<Modal bind:open={modalOpen}>
    {#if modalOpen}
        <LogEntryInterface log={editingLog} on:success={() => {
        modalOpen = false
        updateData(day)
        }}
        metricConfig={metricConfig}
        on:error={(event) => console.error(event.detail)}/>
    {/if}
</Modal>

<style lang="scss">
    .logs-area-container {
        display: flex;
        justify-content: center;
        .logs-area {
            padding: 0.5em;
            margin: 0.5em;
            border-radius: 1em;
            border: 1px solid rgba(0,0,0,0.2);
            width: 560px;
            max-width: 100vw;
            .raw-text-message {
                text-align: center;
                padding: 1em 0;
            }
        }

    }
    .calories-total {
        text-align: center;
        line-height: 1em;
    }
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
            border-radius: 0.5em;
            background-color: rgba(0,0,0,0.1);
            color: black;
            border: none;
            padding: 0.4em 0.8em;
            font-weight: 700;
            cursor: pointer;
            &:hover {
                background-color: lightgray;
            }
        }
        &.top-border {
            border-top: 2px rgba(0,0,0,0.1) dashed;
        }
    }
</style>