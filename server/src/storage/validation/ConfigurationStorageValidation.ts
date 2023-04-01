import {
  Configuration,
  MetricsConfiguration,
  SummaryConfiguration
} from "../../types";

export function isValidConfigurationItem (
  configuration: Configuration
): boolean {
  switch (configuration.id) {
    case "metrics":
      return isValidMetrics(configuration);
    case "summaries":
      return isValidSummary(configuration);
    default:
      return false;
  }
}

function isValidMetrics (configuration: MetricsConfiguration): boolean {
  const entries = Object.entries(configuration.value);
  const labels: Set<string> = new Set();
  const priorities: Set<number> = new Set();
  return (
    entries.findIndex(([key, value]) => {
      if (
        !key.match(/([a-z]|-)*/) ||
        labels.has(value.label) ||
        priorities.has(value.priority)
      ) {
        return true;
      }
      labels.add(value.label);
      priorities.add(value.priority);
      return false;
    }) === -1
  );
}

function isValidSummary (configuration: SummaryConfiguration): boolean {
  return (
    configuration.value.showMetricSummary.length ===
      new Set(configuration.value.showMetricSummary).size &&
    configuration.value.showMetricSummary.every((key) =>
      key.match(/([a-z]|-)*/)
    )
  );
}
