<script lang="ts">

import MetricsConfiguration from "../components/configuration/MetricsConfiguration.svelte";
import BulkLogUpload from "../components/configuration/BulkLogUpload.svelte";
  
</script>
<div class="container">
  <h1>Config</h1>
  <div class="config-container">
    <div class="config-item">
      <h2>Metrics</h2>
      <MetricsConfiguration />
    </div>
  </div>
  <div class="config-container">
    <div class="config-item">
      <h2>Data Management</h2>
      <a href="/api/logs/export" target="_blank" download><button>Download logs</button></a>
      <BulkLogUpload />
    </div>
  </div>
</div>
<style lang="scss">
  .container {
    width: 100vw;
    display: flex;
    flex-direction: column;
    align-items: center;
  }
  .config-container {
    width: 100%;
    display: flex;
    flex-direction: row;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-evenly;
    .config-item {
      margin: 1em;
      padding: 1em;
      border-radius: 1em;
      border: 1px solid #000;
    }
  }
</style>