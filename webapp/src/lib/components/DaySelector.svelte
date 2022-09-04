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
<div class="control-row">
    <Button data-testid="backward-button" on:click={() => {
        day.setDate(day.getDate() - 1)
        day = day
        dispatch('dateChange', day)
    }} outline>&lt;</Button>
    <DatePicker value={day} 
        format="%Y-%m-%d"
        on:change={dateUpdater}
        closeOnSelection/>
    <Button data-testid="forward-button" on:click={() => {
        day.setDate(day.getDate() + 1)
        day = day
        dispatch('dateChange', day)
    }} outline>&gt;</Button>
</div>
<style lang="scss">
    .control-row {
        display: flex;
        align-items: center;
        justify-content: center;
    }
</style>