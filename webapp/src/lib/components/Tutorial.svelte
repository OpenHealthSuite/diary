<script lang="ts">
  import MetricsConfiguration from "./configuration/MetricsConfiguration.svelte";
  import LogEntryInterface from "./LogEntryInterface.svelte";
  import Modal from "./Modal.svelte";

  export let modalOpen = false;

  let tutorialMetrics = undefined;

  enum TutorialStages {
    Intro,
    Metrics,
    Log,
    Complete
  }

  let stage = 0;

  let icons = [' - ',' - ',' - ',' - ']

  $: {
    let newIcons = [];
    for (let i = 0; i <= TutorialStages.Complete; i++) {
      newIcons.push(i <= stage ? ' * ' : ' - ')
    }
    icons = newIcons;
  }

  const progress = () => {
    if (stage < TutorialStages.Complete) {
      stage = stage + 1
    } else {
      modalOpen = false;
      stage = 0;
    }
  }

  const nextButtonLabelText = (stg: number) => {
    if (stage === TutorialStages.Complete) {
      return "Close";
    }
    if (stage === TutorialStages.Log) {
      return "Skip";
    }
    return "Next";
  }
</script>

<Modal bind:open={modalOpen}>
  <div class="tutorial-content">
    {#if stage == TutorialStages.Intro}
      <h2>Welcome!</h2>
      <p>This brief tutorial will run you through the key concepts of OpenFoodDiary.</p>
    {:else if stage == TutorialStages.Metrics}
      <h2>Enter your metrics!</h2>
      <p>Enter the metrics you want to track, like Calories, Sugar, Satiation or Enjoyment.</p>
      <p class="note">Note: You can choose to track no metrics at all.</p>
      <MetricsConfiguration/>
    {:else if stage == TutorialStages.Log}
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
    <button on:click={progress} class={stage == TutorialStages.Log ? "skip-button" : ""}>{nextButtonLabelText(stage)}</button>
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
  .skip-button {
    background-color: rgba(0,0,0,0);
    border: 1px solid rgba(0,0,0,1);
  }
</style>