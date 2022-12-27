<script lang="ts">
import type { FoodLogEntry } from "src/lib/types/FoodLogEntry";
import Modal from "../Modal.svelte";
import { parse } from "csv-parse/sync"
import { apiFetch } from "src/lib/utilities";

let modalOpen = false;

let files: FileList = undefined;

let preparing = false;
let uploading = false;

let uploaded = 0;
let errors = 0;

let logs: FoodLogEntry[] = [];

let newLogs: FoodLogEntry[] = [];
let editLogs: FoodLogEntry[] = [];

const filesUpdated = async () => {
  preparing = true;
  logs = [];
  newLogs = [];
  editLogs = [];
  uploaded = 0;
  errors = 0;
  for (let file of files) {
      const filedata = await file.text();
      const records = parse(filedata, {
        columns: true,
        skip_empty_lines: true,
      }).map((x: any) => {
        let log = {
          ...x,
          time: {
            start: x.timeStart,
            end: x.timeEnd
          },
          metrics: JSON.parse(x.metrics),
          labels: JSON.parse(x.labels),
        };
        delete log.timeStart;
        delete log.timeEnd;
        return log;
      }) as FoodLogEntry[];
      logs = [...logs, ...records];
  }
  for (let log of logs) {
    if (log.id) {
      const existing = await apiFetch("/logs/" + log.id);
      if (existing.status === 200) {
        editLogs.push(log);
      } else {
        newLogs.push(log);
      }
    } else {
      newLogs.push(log);
    }
  }
  preparing = false;
}

const uploadLogs = async () => {
  uploading = true;
  uploaded = 0;
  errors = 0;
  for (let newLog of newLogs) {
    apiFetch('/logs', {
            method: 'POST',
            body: JSON.stringify(newLog)
        }).then((response) => { 
            if (response.status === 200) {
                uploaded += 1;
            } else {
                errors += 1;
            }
        });
  }
  for (let editLog of editLogs) {
    apiFetch('/logs/' + editLog.id , {
            method: 'PUT',
            body: JSON.stringify(editLog)
        }).then((response) => { 
            if (response.status === 200) {
                uploaded += 1;
            } else {
                errors += 1;
            }
        });
  }
  uploading = false;
}

</script>

<button on:click={() => modalOpen = true}>Upload Logs</button>

<Modal bind:open={modalOpen}>
  <input type="file" id="input" bind:files={files} on:change={filesUpdated}/>
  {#if preparing}
    <div>Processing...</div>
  {/if}
  {#if !preparing && !uploading && logs.length > 0}
    <ul>
      <li>{logs.length} logs to upload</li>
      <li>{newLogs.length} logs are new (these logs will have new ids)</li>
      <li>{editLogs.length} logs are updates</li>
    </ul>
    <button on:click={uploadLogs}>Upload Logs</button>
  {/if}
  {#if uploading}
    <ul>
      <li>Uploaded {uploaded}/{logs.length}</li>
      <li>Errors: {errors}</li>
    </ul>
  {/if}
  {#if !uploading && (uploaded > 0 || errors > 0)}
    <h3>Done!</h3>
    <ul>
      <li>Uploaded {uploaded}/{logs.length}</li>
      <li>Errors: {errors}</li>
    </ul>
  {/if}
</Modal>

<style lang="scss">

</style>