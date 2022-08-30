<script lang="ts">
    import { apiFetch } from 'src/lib/utilities/index'
    import type { FoodLogEntry } from '../types/FoodLogEntry';
    export let day: Date
    const dayString = day.toISOString().split("T")[0];
    const startDate = new Date(dayString);
    const endDate = new Date(dayString)
    endDate.setDate(endDate.getDate() + 1);
    
    let loading = true;
    let error = false;

    let dayData: FoodLogEntry[] = [];

    apiFetch(`/logs?startDate=${startDate.toISOString()}&endDate=${endDate.toISOString()}`)
        .then(res => res.json())
        .then(data => dayData = data.sort((a, b) => a.time.start.getTime() - b.time.start.getTime()))
        .catch(() => error = true)
        .finally(() => loading = false)
</script>

{#if loading}
<div data-testid="loading-indicator">
</div>
{:else if error}
<div data-testid="error-indicator">
</div>
{:else}
{#if dayData.length === 0}
<div>
    No Logs Entered for this day
</div>
{:else}
<div>
    <dl>
        {#each dayData as log, i}
            <dt data-testid="foodlog-{i}">{log.name}<dt>
            <dd data-testid="foodlog-{i}-calories">{log.metrics['calories']}</dd>
        {/each}
    </dl>
</div>
{/if}
{/if}

<style>

</style>