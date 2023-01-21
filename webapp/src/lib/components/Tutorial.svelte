<script lang="ts">
  import MetricsConfiguration from "./configuration/MetricsConfiguration.svelte";
  import LogEntryInterface from "./LogEntryInterface.svelte";
  import Modal from "./Modal.svelte";

  export let modalOpen = false;

  // TODO: Really need to elevate this to some global state...
  let tutorialMetrics = undefined;

  let stage = 0;
  const MAX_STAGE = 3;

  let icons = [' - ',' - ',' - ',' - ']

  $: {
    let newIcons = [];
    for (let i = 0; i <= MAX_STAGE; i++) {
      newIcons.push(i <= stage ? ' * ' : ' - ')
    }
    icons = newIcons;
  }

  const progress = () => {
    if (stage < MAX_STAGE) {
      stage = stage + 1
    } else {
      modalOpen = false;
      stage = 0;
    }
  }
</script>

<Modal bind:open={modalOpen}>
  <div class="tutorial-content">
    {#if stage == 0}
      <h2>Welcome!</h2>
      <p>This brief tutorial will run you through the key concepts of OpenFoodDiary.</p>
    {:else if stage == 1}
      <h2>Enter your metrics!</h2>
      <p>Enter the names of metrics you want to track.</p>
      <p class="note">Note: You can choose to track no metrics at all.</p>
      <MetricsConfiguration/>
    {:else if stage == 2}
      <h2>Add a Log!</h2>
      <p class="note">Note: You can delete this log later - it's not permanent.</p>
      <LogEntryInterface on:success={progress}/>
    {:else}
      <h2>You're Ready!</h2>
      <p>You can replay this tutorial at any time from the config page, accessed from the button at the bottom of the screen.</p>
    {/if}
  </div>
  <div class="tutorial-step-controls">
    <button disabled={stage == 0} on:click={() => {stage = stage - 1}}>Back</button>
    <div>{#each icons as icon}{icon}{/each}</div>
    <button on:click={progress}>{#if stage != MAX_STAGE}Next{:else}Close{/if}</button>
  </div>
</Modal>

<style lang="scss">
  .tutorial-content {
    width: 450px;
    max-width: 80vw;
  }
  .tutorial-step-controls {
    display: flex;
    flex-direction: row;
    justify-content: space-between;
    align-items: center;
    margin-top: 2em;
    button {
      width: 7em;
      font-size: 1em;
    }
  }
  p.note {
    font-style: italic;
  }
</style>