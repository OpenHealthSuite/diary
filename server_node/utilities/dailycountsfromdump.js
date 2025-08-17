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

// first date
const firstDate = exportedLogs[0][0];
// last date
const lastDate = exportedLogs[exportedLogs.length - 1][0];

function getDatesBetween (startDateStr, endDateStr) {
  // Parse the date strings into Date objects
  const startDate = new Date(startDateStr);
  const endDate = new Date(endDateStr);

  // Array to store each date
  const dates = [];

  // Iterate over each day
  // eslint-disable-next-line no-unmodified-loop-condition
  for (let date = new Date(startDate); date <= endDate; date.setDate(date.getDate() + 1)) {
    // Add the current date to the array
    dates.push(new Date(date).toISOString().split("T")[0]); // Format as "YYYY-MM-DD"
  }

  return dates;
}

const allDates = getDatesBetween(firstDate, lastDate);

const finalLogs = allDates.map((date) => [
  date,
  ...knownmetrics.map(km => aggregated[date] ? aggregated[date][km] ?? 0 : 0)
]).sort(([a], [b]) => a.localeCompare(b));

fs.appendFileSync(filename, stringify(finalLogs));
