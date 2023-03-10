<script lang="ts">
    import { metricsConfig, type MetricsConfig } from 'src/stores';

    let metrics: MetricsConfig;

    metricsConfig.subscribe(val => {
        metrics = val;
    })

    let newMetric: string | undefined = undefined;

    const saveMetrics = () => {
        metricsConfig.set(metrics)
    }

    const createNewMetric = () => {
        if (newMetric) {
            let newMetricLabel = newMetric;
            const key = newMetric.replaceAll(" ", "_").toLowerCase().replace(/[a-z]^/g, '');
            metrics[key] = {
                label: newMetricLabel,
                priority: Math.max(...Object.values(metrics).map(x => x.priority), 0) + 1
            }
            saveMetrics();
            newMetric = undefined;
        }
    }

    const deleteMetric = (key: string) => {
        delete metrics[key];
        saveMetrics();
    }

    $: metrics_list = Object.entries(metrics).map(([key, val]) => { return { key, val }})
</script>
<div style="width: 100%">
    {#if !metrics}
        <div>Loading...</div>
    {:else}
    <div style="width: 100%">
        {#each metrics_list as { key, val: { label }} (key)}
        <div class="input-row">
            <input bind:value={label} on:change={saveMetrics}/>
            <button on:click={() => deleteMetric(key)}>Delete</button>
        </div>
        {/each}
        <div class="input-row">
            <input placeholder="New metric..." bind:value={newMetric}/>
            <button disabled={!newMetric} on:click={createNewMetric}>Create</button>
        </div>
    </div>
    {/if}
</div>
<style lang="scss">
    .input-row {
        width: 100%;
        display: flex;
        flex-direction: row;
        justify-content: space-evenly;
        margin: 0.25em 0;
        input {
            width: calc(100% - 8em);
        }
        button {
            width: 5em;
        }
    }
</style>