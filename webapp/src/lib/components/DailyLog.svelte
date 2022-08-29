<script lang="ts">
  import { apiFetch } from 'src/lib/utilities/index'
  export let day: Date
  const dayString = day.toISOString().split("T")[0];
  const startDate = new Date(dayString);
  const endDate = new Date(dayString)
  endDate.setDate(endDate.getDate() + 1);

  apiFetch(`/logs?startDate=${startDate.toISOString()}&endDate=${endDate.toISOString()}`)
</script>

<div>
</div>

<style>

</style>