const { parse } = require("csv-parse/sync");
const { stringify } = require("csv-stringify/sync");
const fs = require("node:fs");

// For creating a spreadsheet of counts

const data = fs.readFileSync("allrecords.csv");
const prsd = parse(data, { columns: true });

const knownmetrics = [];

const aggregated = prsd.reduce((acc, rcrd) => {
  const rcrdday = rcrd.timeStart.split("T")[0];
  const metrics = JSON.parse(rcrd.metrics);
  Object.keys(metrics).forEach((key) => {
    if (!knownmetrics.includes(key)) {
      knownmetrics.push(key);
    }
  });
  if (acc[rcrdday]) {
    Object.entries(metrics).forEach(([key, val]) => {
      if (acc[rcrdday][key]) {
        acc[rcrdday][key] += val;
      } else {
        acc[rcrdday][key] = val;
      }
    });
  } else {
    acc[rcrdday] = metrics;
  }
  return acc;
}, {});

const filename = "dailysummary.csv";
fs.writeFileSync(
  filename,
  stringify([["day", ...knownmetrics]]),
  {
    flag: "w"
  }
);

const exportedLogs = Object.entries(aggregated).map(([day, metrics]) => [
  day,
  ...knownmetrics.map(km => metrics[km] ?? 0)
]).sort(([a], [b]) => a.localeCompare(b));

fs.appendFileSync(filename, stringify(exportedLogs));
