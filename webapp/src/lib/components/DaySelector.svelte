<script lang="ts">
    import { createEventDispatcher } from 'svelte';
    export let day: Date = new Date();

    const dispatch = createEventDispatcher();

    const dateUpdater = (event: any) => {
        day = new Date(event.detail.value);
        dispatch('dateChange', day)
    }
</script>
<div class="control-row">
    <button data-testid="backward-button" on:click={() => {
        day.setDate(day.getDate() - 1)
        day = day
        dispatch('dateChange', day)
    }}>&lt;</button>
    <fieldset>
        <label for="day-selected">Day Selected</label>
        <input type="date" id="day-selected" name="day-selected"
            value={day}
            on:change={dateUpdater}> 
    </fieldset>
    <button data-testid="forward-button" on:click={() => {
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
</style>