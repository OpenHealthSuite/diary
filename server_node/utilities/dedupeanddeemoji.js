const { parse } = require("csv-parse/sync");
const { stringify } = require("csv-stringify/sync");
const fs = require("node:fs");

// For deduping records on export to clean up some old data mistakes of mine from a migration

const data = fs.readFileSync("raw.csv");
const prsd = parse(data, { columns: true });
const totalrows = prsd.length;

let deduped = [];

const containsEmoji = str => /[\p{Emoji_Presentation}\p{Extended_Pictographic}]/u.test(str);

for (const itm of prsd) {
  const { id, ...res } = itm;

  if (deduped.findIndex(({ id: _, ...x }) => {
    return x.name === res.name && x.timeStart === res.timeStart && x.timeEnd === res.timeEnd;
  }) === -1 && !containsEmoji(res.name)) {
    deduped = [...deduped, { id, ...res }];
  }
}

console.log("totalrows: ", totalrows);
console.log("dedupedrows: ", deduped.length);

const filename = "deduped.csv";
fs.writeFileSync(
  filename,
  stringify([["id", "name", "labels", "timeStart", "timeEnd", "metrics"]]),
  {
    flag: "w"
  }
);

const exportedLogs = deduped.map((log) => [
  log.id,
  log.name,
  log.labels,
  log.timeStart,
  log.timeEnd,
  log.metrics
]);

fs.appendFileSync(filename, stringify(exportedLogs));
