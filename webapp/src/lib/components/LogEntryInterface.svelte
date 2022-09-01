<script lang="ts">
    import { createEventDispatcher } from 'svelte';
    import { Button, DatePicker, TimePicker, FormField } from "attractions";
    import type { FoodLogEntry } from '../types/FoodLogEntry';
    import { apiFetch } from "../utilities";
    export let logTime = new Date();
    export let log: FoodLogEntry = {
        id: undefined,
        name: "",
        labels: new Set(),
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
    let startTime = log.time.start ?? logTime;

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
                labels: new Set(),
                time: {
                    start: startTime.toISOString(),
                    end: endTime.toISOString()
                },
                metrics: {
                    calories
                }
            })
        }).then(response => { 
            if (response.status !== 200) {
                response.text()
                    .then(text => dispatch('error', text))
                    .catch(() => dispatch('error', "An unknown error occured"))
            }
            return response.json() 
        })
            .then(guid => dispatch('success', guid))
            .catch(err => dispatch('error', "An unknown error occured"));
    }
</script>

<div>
    <FormField
        id="name"
        name="Log Name"
        help="Quick description of the food/meal"
        required
        >
        <input type="text" id="name" bind:value={name}/>
    </FormField>
    <label for="date-picker">Date</label>
    <DatePicker value={startTime} closeOnSelection/>
    <label for="time">Time</label>
    <TimePicker value={logTime} />
    <FormField
        id="duration"
        name="Duration (minutes)"
        help="How long you were eating for"
        required
        >
        <input type="number" min="1" bind:value={duration} id="duration"/>
    </FormField>
    <FormField
        id="calories"
        name="Calories"
        help="How many calories was the food/meal"
        required
        >
        <input type="number" min="0" bind:value={calories} id="calories"/>
    </FormField>
    <Button disabled={!name} on:click={() => name && submitLog()}>Submit</Button>
</div>