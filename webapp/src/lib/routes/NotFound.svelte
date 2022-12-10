<script lang="ts">
  import { link } from "svelte-routing";

</script>
<div class="not-found">
  <h1>Not Found</h1>
  <a href="/" use:link>Back to Logs</a>
</div>
<style lang="scss">
  .not-found {
    width: 100vw;
    height: 100vh;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    a {
      color: #333;
    }
  }
</style>