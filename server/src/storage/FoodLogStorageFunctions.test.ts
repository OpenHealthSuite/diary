import { sqlite3 } from "./sqlite3";
import { cassandra } from "./cassandra";
import crypto from "node:crypto";
import {
  CreateFoodLogEntry,
  EditFoodLogEntry,
  isValidationError,
} from "./types";
import { FoodLogStorage } from "./types/FoodLog";

const configs = [
  { name: "sqlite3", config: sqlite3.foodLog },
  { name: "cassandra", config: cassandra.foodLog },
];

describe.each(configs)(
  "$name FoodLogStorageFunctions",
  ({ config }: { config: FoodLogStorage }) => {
    describe("CreateFoodLog", () => {
      describe("Validation Errors", () => {
        const GoldInput: CreateFoodLogEntry = {
          name: "My Food Log",
          labels: ["Some Label", "Some other label"],
          time: {
            start: new Date(1999, 10, 10),
            end: new Date(1999, 10, 11),
          },
          metrics: {
            calories: 500,
          },
        };

        const testUserId = crypto.randomUUID();

        const { name, ...nameless } = structuredClone(GoldInput);
        const { labels, ...labelless } = structuredClone(GoldInput);
        const { metrics, ...metricless } = structuredClone(GoldInput);
        let weirdMetric: any = structuredClone(GoldInput);
        weirdMetric.metrics.calories = "This is not a number";
        const { time, ...timeless } = structuredClone(GoldInput);
        let startTimeLess: any = structuredClone(GoldInput);
        delete startTimeLess.time.start;
        let endTimeLess: any = structuredClone(GoldInput);
        delete endTimeLess.time.end;
        let endBeforeStart = structuredClone(GoldInput);
        endBeforeStart.time.start = new Date(1999, 10, 10);
        endBeforeStart.time.end = new Date(1999, 8, 10);

        const BadValues: any[] = [
          ["Empty", {}],
          ["WithId", { ...GoldInput, id: crypto.randomUUID() }],
          ["No Name", nameless],
          ["No Labels", labelless],
          ["No Metrics", metricless],
          ["Non-number Metric", weirdMetric],
          ["No Times", timeless],
          ["No Start Time", startTimeLess],
          ["No End Time", endTimeLess],
          ["End before start", endBeforeStart],
        ];

        it.each(BadValues)(
          "Rejects Bad Test Case %s",
          async (name: string, badValue: any) => {
            const result = await config.storeFoodLog(testUserId, badValue);

            expect(result.isErr()).toBeTruthy();
            expect(isValidationError(result._unsafeUnwrapErr())).toBeTruthy();
          }
        );
      });
    });

    describe("QueryFoodLog", () => {
      describe("Validation Errors", () => {
        it("Rejects disordered dates", async () => {
          const result = await config.queryFoodLogs(
            crypto.randomUUID(),
            new Date(1997, 10, 1),
            new Date(1987, 10, 1)
          );

          expect(result.isErr()).toBeTruthy();
          expect(isValidationError(result._unsafeUnwrapErr())).toBeTruthy();
        });
      });
    });

    describe("EditFoodLog", () => {
      describe("Validation Errors", () => {
        const GoldInput: EditFoodLogEntry = {
          id: crypto.randomUUID(),
          name: "My Food Log",
          labels: ["Some Label", "Some other label"],
          time: {
            start: new Date(),
            end: new Date(),
          },
          metrics: {
            calories: 500,
          },
        };

        const testUserId = crypto.randomUUID();

        let weirdMetric: any = structuredClone(GoldInput);
        weirdMetric.metrics.calories = "This is not a number";
        let startTimeLess: any = structuredClone(GoldInput);
        delete startTimeLess.time.start;
        let endTimeLess: any = structuredClone(GoldInput);
        delete endTimeLess.time.end;
        let endBeforeStart = structuredClone(GoldInput);
        endBeforeStart.time!.start = new Date(1999, 10, 10);
        endBeforeStart.time!.end = new Date(1999, 8, 10);

        const BadValues: any[] = [
          ["Empty", {}],
          ["WithoutId", { ...GoldInput, id: undefined }],
          ["Non-number Metric", weirdMetric],
          ["No Start Time", startTimeLess],
          ["No End Time", endTimeLess],
          ["End before start", endBeforeStart],
        ];

        it.each(BadValues)(
          "Rejects Bad Test Case %s",
          async (name: string, badValue: any) => {
            const result = await config.editFoodLog(testUserId, badValue);

            expect(result.isErr()).toBeTruthy();
            expect(isValidationError(result._unsafeUnwrapErr())).toBeTruthy();
          }
        );
      });
    });
  }
);
