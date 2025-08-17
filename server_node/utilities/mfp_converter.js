const { parse } = require("csv-parse/sync");
const { stringify } = require("csv-stringify/sync");
const fs = require("node:fs");
const { randomUUID } = require("node:crypto");

// For converting MFP exports to OHS logs

// id,name,labels,timeStart,timeEnd,metrics
// 00110ca6-ceb1-4144-83a3-473f4291fbbf,Snack - Chocolate Croissant,,2022-11-06T12:20:00.000Z,2022-11-06T12:25:00.000Z,"{""calories"":350}"

const data = fs.readFileSync("Nutrition-Summary.csv");
const prsd = parse(data, { columns: true });

const conversionFunction = (input) => {
  const timestartstring = "T12:15:00.000Z";
  const timeendstring = "T12:30:00.000Z";
  const knownMealTimes = {
    Breakfast: ["T08:00:00.000Z", "T08:15:00.000Z"],
    "Mid-morning Snacks": ["T10:00:00.000Z", "T10:15:00.000Z"],
    Lunch: ["T12:00:00.000Z", "T12:15:00.000Z"],
    "Mid Afternoon Snacks": ["T14:00:00.000Z", "T14:15:00.000Z"],
    Dinner: ["T18:00:00.000Z", "T18:15:00.000Z"],
    "Evening Snacks": ["T20:00:00.000Z", "T20:15:00.000Z"]
  };
  return {
    id: randomUUID(),
    name: input.Meal,
    labels: "",
    timeStart: knownMealTimes[input.Meal] ? input.Date + knownMealTimes[input.Meal][0] : input.Date + timestartstring,
    timeEnd: knownMealTimes[input.Meal] ? input.Date + knownMealTimes[input.Meal][1] : input.Date + timeendstring,
    metrics: `{"calories":${input.Calories}}`
  };
};

const filename = "nutrition-summary-converted.csv";
fs.writeFileSync(
  filename,
  stringify([["id", "name", "labels", "timeStart", "timeEnd", "metrics"]]),
  {
    flag: "w"
  }
);

const exportedLogs = prsd.map(conversionFunction).map((log) => [
  log.id,
  log.name,
  log.labels,
  log.timeStart,
  log.timeEnd,
  log.metrics
]);

fs.appendFileSync(filename, stringify(exportedLogs));
