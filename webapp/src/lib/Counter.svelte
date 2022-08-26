<script lang="ts">
  import { Button } from 'attractions';
  let count: number = 0
  const increment = () => {
    count += 1
  }
</script>

<Button on:click={increment}>
  count is {count}
</Button>
