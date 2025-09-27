<script lang="ts">
import type { FoodLogEntry } from "src/lib/types/FoodLogEntry";
import Modal from "../Modal.svelte";
import { parse } from "csv-parse/sync";
import { apiFetch } from "src/lib/utilities";

export let modalOpen = false;

// eslint-disable-next-line no-undef
let files: FileList;

let preparing = false;
let parseError;

let uploaded = 0;
let errors = 0;

let logs: FoodLogEntry[] = [];

let newLogs: FoodLogEntry[] = [];
let editLogs: FoodLogEntry[] = [];

const filesUpdated = async () => {
  try {
    preparing = true;
    parseError = undefined;
    logs = [];
    newLogs = [];
    editLogs = [];
    uploaded = 0;
    errors = 0;
    for (const file of files) {
      const filedata = await file.text();
      const records = parse(filedata, {
        columns: true,
        skip_empty_lines: true
      }).map((x: any) => {
        const log = {
          ...x,
          time: {
            start: x.timeStart,
            end: x.timeEnd
          },
          metrics: x.metrics ? JSON.parse(x.metrics) : {},
          labels: x.labels ? JSON.parse(x.labels) : []
        };
        delete log.timeStart;
        delete log.timeEnd;
        return log;
      }) as FoodLogEntry[];
      logs = [...logs, ...records];
    }
    for (const log of logs) {
      if (log.id) {
        const existing = await apiFetch("/logs/" + log.id);
        if (existing.status === 200) {
          editLogs = [...editLogs, log];
        } else {
          delete log.id;
          newLogs = [...newLogs, log];
        }
      } else {
        delete log.id;
        newLogs = [...newLogs, log];
      }
    }
  } catch (ex: any) {
    parseError = ex.message;
  }
  preparing = false;
};

const uploadLogs = async () => {
  uploaded = 0;
  errors = 0;
  for (const newLog of newLogs) {
    apiFetch("/logs", {
      method: "POST",
      body: JSON.stringify(newLog)
    }).then((response) => {
      if (response.status === 200) {
        uploaded += 1;
      } else {
        errors += 1;
      }
    });
  }
  for (const editLog of editLogs) {
    apiFetch("/logs/" + editLog.id, {
      method: "PUT",
      body: JSON.stringify(editLog)
    }).then((response) => {
      if (response.status === 200) {
        uploaded += 1;
      } else {
        errors += 1;
      }
    });
  }
};

</script>

<Modal bind:open={modalOpen}>
  <input type="file" id="input" bind:files={files} on:change={filesUpdated}/>
  {#if preparing}
    <div>Processing... (Log {editLogs.length + newLogs.length}/{logs.length})</div>
  {/if}
  {#if parseError}
    <div>Error parsing file!</div>
    <div>{parseError}</div>
  {/if}
  {#if !preparing && !parseError && uploaded === 0 && logs.length > 0}
    <ul>
      <li>{logs.length} logs to upload</li>
      <li>{newLogs.length} logs are new (these logs will have new ids)</li>
      <li>{editLogs.length} logs are updates</li>
    </ul>
    <button on:click={uploadLogs}>Upload Logs</button>
  {/if}
  {#if logs.length > 0 && uploaded + errors === logs.length}
    <h3>Done!</h3>
  {/if}
  {#if uploaded > 0}
    <ul>
      <li>Uploaded {uploaded}/{logs.length}</li>
      <li>Errors: {errors}</li>
    </ul>
  {/if}
</Modal>

<style lang="scss">

</style>