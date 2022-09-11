<script lang="ts">
    import { createEventDispatcher } from 'svelte';
    import { Button, DatePicker, TimePicker, FormField, TextField } from "attractions";
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
</script>

<div>
    <div class="log-datetime-selectors">
        <DatePicker 
            value={startTime}
            format="%Y-%m-%d"
            on:change={dateUpdater} 
            closeOnSelection/>
        <TimePicker value={startTime} format="%H:%M"
            on:change={dateUpdater} />
    </div>
    <FormField
        id="name"
        name="Log Name"
        required
        >
        <TextField
            bind:value={name}
            id="name"
            error={name.length < 1 && 'Must add a log name'}
        />
    </FormField>
    <FormField
        id="calories"
        name="Calories"
        required
        >
        <TextField
            bind:value={calories}
            id="calories"
            type="number"
            min="0"
            error={calories < 0 && 'Must have a positive or zero calories'}
        />
    </FormField>
    <FormField
        id="duration"
        name="Duration (minutes)"
        required
        >
        <TextField
            bind:value={duration}
            id="duration"
            type="number"
            min="1"
            error={duration < 1 && 'Must have a positive duration'}
        />
    </FormField>
    <Button disabled={!name} on:click={() => name && submitLog()} outline>Submit</Button>
</div>

<style lang="scss">
    .log-datetime-selectors {
        margin-bottom: 2em;
    }
</style>