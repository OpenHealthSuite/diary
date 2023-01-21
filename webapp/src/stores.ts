import { writable } from 'svelte/store';
import { DEFAULT_METRICS, apiFetch } from './lib/utilities';

export type MetricsConfig = {[key: string]: { label: string; priority: number }};

export const metricsConfig = writable({} as MetricsConfig);

apiFetch("/config/metrics").then(res => {
    switch (res.status) {
        case 200:
            res.json().then(mtrcs => {
                metricsConfig.set(mtrcs.value);
            })
            break;
        case 404:
            metricsConfig.set(DEFAULT_METRICS)
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
            })
        }
    })
})