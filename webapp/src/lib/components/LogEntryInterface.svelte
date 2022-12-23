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

    let [dateString] = startTime.toISOString().split('T')

    let timeString = `${String(startTime.getHours()).padStart(2, '0')}:${String(startTime.getMinutes()).padStart(2, '0')}`;

    let loading = false;

    const submitLog = () => {
        const logTime = new Date(dateString+'T'+timeString)
        const endTime = new Date(logTime)
        endTime.setMinutes(endTime.getMinutes() + duration)
        loading = true;
        apiFetch(log.id ? '/logs/' + log.id : '/logs', {
            method: log.id ? 'PUT' : 'POST',
            body: JSON.stringify({
                id: log.id,
                name,
                labels: [],
                time: {
                    start: logTime,
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
            .catch(err => dispatch('error', "An unknown error occured"))
            .finally(() => {
                loading = false;
            });
    }

    const deleteLog = (logId: string) => {
        loading = true;
        apiFetch('/logs/' + logId , {
            method: 'DELETE',
        }).then((response) => { 
            if (response.status === 204) {
                return;
            }
            throw new Error() 
        })
        .then(() => dispatch('success', undefined))
        .catch(() => dispatch('error', "An unknown error occured"))
        .finally(() => {
            loading = false;
        });
    }
</script>

<div class="form-container">
    <fieldset class="log-datetime-selectors">
        <label for="log-entry-day">Log Day
            <input type="date" id="log-entry-day" name="log-entry-day" 
            pattern="\d{4}-\d{2}-\d{2}"
            bind:value={dateString}> 
        </label>
        <label for="log-entry-day">Log Time
            <input type="time" id="log-entry-day" name="log-entry-day" 
            bind:value={timeString}>
        </label>
    </fieldset>


    <fieldset>
        <label for="name" hidden>Log Name</label>
        <input style="width: 95%" id="name" name="name" placeholder="Log Name" bind:value={name} />
        {#if name.length < 1}
            <div class="input-error">Must add a log name</div>
        {/if}
    </fieldset>
    <fieldset class="left-right-field">
        <label for="calories">Calories</label>
        <input type="number" id="calories" name="calories" bind:value={calories} min={0} />
    </fieldset>
    {#if calories < 0}
        <div class="input-error">Must have min zero calories</div>
    {/if}
    <fieldset class="left-right-field">
        <label for="duration">Duration (minutes)</label>
        <input type="number" id="duration" name="duration" bind:value={duration} min={1} />
    </fieldset>
    {#if duration < 1}
        <div class="input-error">Must have a positive duration</div>
    {/if}
    <button class="submit-button" disabled={!name || loading} on:click={() => name && submitLog()}>Submit</button>
    {#if log.id}
        <button class="submit-button" disabled={loading} on:click={() => deleteLog(log.id)}>Delete</button>
    {/if}
</div>

<style lang="scss">
    fieldset {
        border: none;
        padding: 0;
    }
    .log-datetime-selectors {
        margin-bottom: 1em;
        display: flex;
        flex-direction: column;
        label {
            margin: 0.3em 0;
        }
    }
    label {
        font-size: 0.8em;
    }
    .input-error {
        font-size: 0.7em;
    }
    .form-container {
        display: flex;
        flex-direction: column;
    }
    .submit-button {
        margin-top: 2em;
        width: 80%;
        align-self: center;
        padding: 0.5em;
        color: black;
        background-color: rgba(0,0,0,0.2);
        border: none;
        border-radius: 0.5em;
    }
    .left-right-field {
        display: flex;
        flex-direction: row;
        align-items: center;
        justify-content: space-between;
        margin-top: 1em;
        label {
            width: 45%;
        }
        input {
            width: 40%;
        }
    }
</style>