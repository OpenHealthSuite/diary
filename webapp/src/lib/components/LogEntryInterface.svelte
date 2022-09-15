<script lang="ts">
    import { createEventDispatcher } from 'svelte';
    import type { FoodLogEntry } from '../types/FoodLogEntry';
    import { apiFetch } from "../utilities";
    export let logTime = new Date();
    export let log: FoodLogEntry = {
        id: undefined,
        name: "",
        labels: [],
        time: {
            start: undefined,
            end: undefined
        },
        metrics: {
            calories: 0
        }
    }

    const dispatch = createEventDispatcher();

    let name = log.name;
    let startTime = log.time.start ? new Date(log.time.start) : logTime;

    let duration = log.time.end && log.time.start ? 
        ((new Date(log.time.end).getTime()) - (new Date(log.time.start).getTime())) / 1000 / 60 : 1;
    let calories = log.metrics.calories ?? 0;

    const submitLog = () => {
        const endTime = new Date(startTime)
        endTime.setMinutes(endTime.getMinutes() + duration)
        apiFetch(log.id ? '/logs/' + log.id : '/logs', {
            method: log.id ? 'PUT' : 'POST',
            body: JSON.stringify({
                id: log.id,
                name,
                labels: [],
                time: {
                    start: startTime.toISOString(),
                    end: endTime.toISOString()
                },
                metrics: {
                    calories
                }
            })
        }).then((response) => { 
            if (response.status === 200) {
                return response.text()
            }
            throw new Error() 
        })
            .then(guid => dispatch('success', guid))
            .catch(err => dispatch('error', "An unknown error occured"));
    }

    const dateUpdater = (event: any) => {
        startTime = new Date(event.detail.value);
    }
    const timeUpdater = (event: any) => {
        const [hour, min] = event.detail.value.split(':');
        startTime.setHours(hour, min, 0)
        startTime = startTime
    }
</script>

<div>
    <fieldset class="log-datetime-selectors">
        <label for="log-entry-day">Log Day</label>
        <input type="date" id="log-entry-day" name="log-entry-day" 
            value={startTime}
            on:change={dateUpdater}> 
        <label for="log-entry-day">Log Time</label>
        <input type="time" id="log-entry-day" name="log-entry-day" 
            value={startTime.toISOString().split('T')[1].slice(0, 5)}
            on:change={timeUpdater}>
    </fieldset>


    <fieldset>
        <label for="name">Log Name</label>
        <input id="name" name="name" bind:value={name} />
        {#if name.length < 1}
            <div>Must add a log name</div>
        {/if}
    </fieldset>
    <fieldset>
        <label for="calories">Calories</label>
        <input type="number" id="calories" name="calories" bind:value={calories} min={0} />
        {#if calories < 0}
            <div>Must have min zero calories</div>
        {/if}
    </fieldset>
    <fieldset>
        <label for="duration">Duration (minutes)</label>
        <input type="number" id="duration" name="duration" bind:value={duration} min={1} />
        {#if duration < 1}
            <div>Must have a positive duration</div>
        {/if}
    </fieldset>
    <button disabled={!name} on:click={() => name && submitLog()}>Submit</button>
</div>

<style lang="scss">
    .log-datetime-selectors {
        margin-bottom: 2em;
    }
</style>