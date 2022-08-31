<script lang="ts">
    import { DatePicker, TimePicker, FormField } from "attractions";
    export let logTime = new Date();
    export let duration = 1;
    export let calories = 0;
</script>

<form>
    <FormField
        id="name"
        name="Log Name"
        help="Quick description of the food/meal"
        required
        >
        <input type="text" id="name"/>
    </FormField>
    <label for="date-picker">Date</label>
    <DatePicker value={logTime} closeOnSelection />
    <label for="time">Time</label>
    <TimePicker value={logTime} />
    <FormField
        id="duration"
        name="Duration (minutes)"
        help="Quick description of the food/meal"
        required
        >
        <input type="number" min="1" bind:value={duration} id="duration"/>
    </FormField>
    <FormField
        id="calories"
        name="Calories"
        help="Quick description of the food/meal"
        required
        >
        <input type="number" min="0" bind:value={calories} id="calories"/>
    </FormField>
</form>