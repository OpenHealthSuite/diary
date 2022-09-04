<script lang="ts">
    import { Button, DatePicker } from 'attractions';
    import { createEventDispatcher } from 'svelte';
    export let day: Date = new Date();

    const dispatch = createEventDispatcher();

    const dateUpdater = (event: any) => {
        day = new Date(event.detail.value);
        dispatch('dateChange', day)
    }
</script>
<Button data-testid="backward-button" on:click={() => {
    day.setDate(day.getDate() - 1)
    day = day
    dispatch('dateChange', day)
}}>&lt;</Button>
<DatePicker value={day} 
    on:change={dateUpdater}
    closeOnSelection/>
<Button data-testid="forward-button" on:click={() => {
    day.setDate(day.getDate() + 1)
    day = day
    dispatch('dateChange', day)
}}>&gt;</Button>
<slot day={day}></slot>