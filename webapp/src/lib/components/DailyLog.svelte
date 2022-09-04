<script lang="ts">
    import { Card } from 'attractions';
    import { apiFetch } from 'src/lib/utilities/index'
    import type { FoodLogEntry } from '../types/FoodLogEntry';
    export let day: Date;

    let loading = true;
    let error = false;

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
        <dl>
            {#each dayData as log, i}
                <dt data-testid="foodlog-{i}">{log.name}<dt>
                <dd data-testid="foodlog-{i}-calories">{log.metrics['calories']}</dd>
            {/each}
        </dl>
    </div>
    {/if}
    {/if}
</Card>


<style>

</style>