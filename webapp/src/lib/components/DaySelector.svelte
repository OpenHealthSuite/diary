<script lang="ts">
    import { createEventDispatcher } from 'svelte';
    export let day: Date = new Date();

    const dispatch = createEventDispatcher();

    const dateUpdater = (event: any) => {
        day = new Date(event.target.value);
        dispatch('dateChange', day)
    }
</script>
<div class="control-row">
    <button data-testid="backward-button"
        class="control-button" 
        on:click={() => {
            day.setDate(day.getDate() - 1)
            day = day
            dispatch('dateChange', day)
        }}>&lt;</button>
    <div>
        <label for="day-selected" hidden>Day Selected</label>
        <input type="date" id="day-selected" 
            name="day-selected"
            pattern="\d{4}-\d{2}-\d{2}"
            class="day-selector"
            value={day.toISOString().split('T')[0]}
            on:change={dateUpdater}> 
    </div>
    <button data-testid="forward-button"  
        class="control-button" 
        on:click={() => {
            day.setDate(day.getDate() + 1)
            day = day
            dispatch('dateChange', day)
        }}>&gt;</button>
</div>
<style lang="scss">
    .control-row {
        display: flex;
        align-items: center;
        justify-content: center;
    }
    .control-button {
        margin: 0.5em 1em;
        padding: 1.2em 0.6em;
        border-radius: 0.5em;
        border: none;
        line-height: 0;
        font-weight: 700;
        font-size: 1.3em;
        background-color: rgba(0,0,0,0.2);
        color: black;      
        cursor: pointer;
        &:hover {
            background-color: rgba(0, 0, 0, 0.3);
        }
    }
    .day-selector {
        font-size: 1.2em;
        padding: 0.5em;
        border: 0;
        border-bottom: 1px solid black;
        background-color: rgba(0,0,0,0.05);
        color: black;
    }
</style>