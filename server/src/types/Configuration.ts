export type MetricsConfiguration = {
  id: "metrics";
  value: { [key: string]: { label: string; priority: number } };
};

export type SummaryConfiguration = {
  id: "summaries";
  value: { showMetricSummary: string[] };
};

export type Configuration = MetricsConfiguration | SummaryConfiguration;
