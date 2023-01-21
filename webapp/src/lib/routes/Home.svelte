<script lang="ts">
    import { metricsConfig, type MetricsConfig } from "src/stores";
  import DailyLog from "../components/DailyLog.svelte";
  import DaySelector from "../components/DaySelector.svelte";
  import LogEntryInterface from "../components/LogEntryInterface.svelte";
  import Modal from "../components/Modal.svelte";
    import { DEFAULT_METRICS } from "src/stores";

  let modalOpen = false;
  let logDay = new Date();

  let metricConfig: MetricsConfig = DEFAULT_METRICS;

  metricsConfig.subscribe(val => {
    metricConfig = val;
  })

  const dateChange = (event) => logDay = event.detail;

  const dateWithCurrentTime = (date: Date) => {
    return new Date(date.toISOString().split('T')[0] + 'T' + (new Date()).toISOString().split('T')[1])
  }
</script>

<DaySelector day={logDay} on:dateChange={dateChange} />
<div class="controls-row">
  <button class="add-log-button" 
    on:click={() => modalOpen = true}>Add Log</button>
</div>

<Modal bind:open={modalOpen}>
  {#if modalOpen}
  <div class="form-wrapper">
    <LogEntryInterface logTime={dateWithCurrentTime(logDay)} on:success={() => {
      modalOpen = false
      }}
      on:error={(event) => console.error(event.detail)}/>
  </div>
  {/if}
</Modal>

<DailyLog day={logDay} />

<style lang="scss">
  .controls-row {
    margin: 1em;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .form-wrapper {
    max-height: 70vh;
    overflow-y: scroll;
    scrollbar-width: none;
  }

  .add-log-button {
      padding: 0.5em 1em;
      border-radius: 0.5em;
      border: none;
      font-weight: 700;
      font-size: 1.2em;
      background-color: rgba(0,0,0,0.2);
      color: black;
      cursor: pointer;
      &:hover {
        background-color: rgba(0, 0, 0, 0.3);
      }
  }
</style>