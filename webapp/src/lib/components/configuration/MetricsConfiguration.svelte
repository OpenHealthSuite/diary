<script lang="ts">
    import { apiFetch, DEFAULT_METRICS } from 'src/lib/utilities/index';
    import { onMount } from 'svelte';

    let loading = true;
    let error = false;
    // This is the default metrics if they've not been set
    // This should be exportable from elsewhere
    export let metrics: {[key: string]: { label: string; priority: number }} = DEFAULT_METRICS

    let newMetric: string | undefined = undefined;

    const refreshMetricsConfig = () => {
        apiFetch("/config/metrics").then(res => {
            switch (res.status) {
                case 200:
                    error = false;
                    res.json().then(mtrcs => {
                        metrics = mtrcs.value;
                    })
                    return;
                case 404:
                    error = false;
                    return;
                default: 
                    error = true;
            }
        }).catch(err => {
            error = true;
        }).finally(() => {
            loading = false;
        })
    }

    const saveMetrics = () => {
        apiFetch("/config/metrics", { 
            method: "POST",
            body: JSON.stringify(metrics)
        }).finally(refreshMetricsConfig)
    }

    const createNewMetric = () => {
        if (newMetric) {
            let newMetricLabel = newMetric;
            const key = newMetric.replaceAll(" ", "_").toLowerCase().replace(/[a-z]^/g, '');
            metrics[key] = {
                label: newMetricLabel,
                priority: Math.max(...Object.values(metrics).map(x => x.priority)) + 1
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

    onMount(() => refreshMetricsConfig());
</script>
<div style="width: 100%">
    {#if !loading && error}
        <div>Error</div>
    {/if}
    {#if loading}
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