import { writable } from "svelte/store";
import { apiFetch } from "src/lib/utilities";

export const logUpdated = writable(new Date().toISOString());

export const DEFAULT_METRICS = {
  calories: { label: "Calories", priority: 0 }
};

export type MetricsConfig = {
  [key: string]: { label: string; priority: number };
};

export const metricsConfig = writable({} as MetricsConfig);

apiFetch("/config/metrics").then(async (res) => {
  switch (res.status) {
    case 200:
      metricsConfig.set((await res.json()).value);
      break;
    case 404:
      metricsConfig.set(DEFAULT_METRICS);
      break;
    default:
      // TODO: have a think about errors here
      break;
  }
  metricsConfig.subscribe((mrtc) => {
    if (mrtc) {
      apiFetch("/config/metrics", {
        method: "POST",
        body: JSON.stringify(mrtc)
      });
    }
  });
});
