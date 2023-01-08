import crypto from "node:crypto";
import { CreateFoodLogEntry, isNotFoundError } from "./types";
import fs from "node:fs";
import { parse } from "csv-parse/sync";
import { configs } from "./_testConfigs";
import { FoodLogEntry } from "../types";

describe.each(configs)(
  "$name Food Log Storage Integration Tests",
  ({ config }) => {
    let testClient: any;
    beforeAll(async () => {
      testClient = await config.beforeAllSetup();
    });
    afterAll(async () => {
      await config.afterAllTeardown(testClient);
    });
    test("Happy Path :: Bad Retreives, Creates, Retreives, Edits, Reretrieves, Deletes, Fails Retreive", async () => {
      const testUserId = crypto.randomUUID();

      const randomId = crypto.randomUUID();

      const badResult = await (config.storage.foodLog.retrieveFoodLog as any)(
        testUserId,
        randomId,
        testClient
      );

      expect(badResult.isErr()).toBeTruthy();
      expect(isNotFoundError(badResult._unsafeUnwrapErr())).toBeTruthy();

      const input: CreateFoodLogEntry = {
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

      input.labels.sort((a, b) => a.localeCompare(b))

      const result = await (config.storage.foodLog.storeFoodLog as any)(
        testUserId,
        input,
        testClient
      );

      expect(result.isOk()).toBeTruthy();
      const testItemId = result._unsafeUnwrap();
      expect(testItemId.length).toBeGreaterThan(0);

      const storedItemResult = await (
        config.storage.foodLog.retrieveFoodLog as any
      )(testUserId, testItemId, testClient);

      expect(storedItemResult.isOk()).toBeTruthy();
      const storedItem: FoodLogEntry = storedItemResult._unsafeUnwrap();

      storedItem.labels.sort((a, b) => a.localeCompare(b))
      expect(storedItem).toEqual({ id: testItemId, ...input });

      storedItem.name = "Modified Food Log";
      storedItem.metrics = {
        calories: 400,
      };

      const modifiedResult = await (config.storage.foodLog.editFoodLog as any)(
        testUserId,
        storedItem,
        testClient
      );
      expect(modifiedResult.isOk()).toBeTruthy();
      const modified: FoodLogEntry = modifiedResult._unsafeUnwrap();

      modified.labels.sort((a, b) => a.localeCompare(b))
      expect(modified).toEqual(storedItem);

      const reretrievedItemResult = await (
        config.storage.foodLog.retrieveFoodLog as any
      )(testUserId, testItemId, testClient);

      expect(reretrievedItemResult.isOk()).toBeTruthy();
      const reretreived: FoodLogEntry = reretrievedItemResult._unsafeUnwrap();
      reretreived.labels.sort((a, b) => a.localeCompare(b))
      expect(reretreived).toEqual(storedItem);

      const deleteResult = await (config.storage.foodLog.deleteFoodLog as any)(
        testUserId,
        testItemId,
        testClient
      );

      expect(deleteResult.isOk()).toBeTruthy();
      expect(deleteResult._unsafeUnwrap()).toBeTruthy();

      const postDeleteRetrieve = await (
        config.storage.foodLog.retrieveFoodLog as any
      )(testUserId, testItemId, testClient);

      expect(postDeleteRetrieve.isErr()).toBeTruthy();
      expect(
        isNotFoundError(postDeleteRetrieve._unsafeUnwrapErr())
      ).toBeTruthy();

      const redeleteResult = await (
        config.storage.foodLog.deleteFoodLog as any
      )(testUserId, testItemId, testClient);

      expect(redeleteResult.isOk()).toBeTruthy();
      expect(redeleteResult._unsafeUnwrap()).toBeTruthy();
    });

    test.skip("Queries :: can add some logs, and get expected query results", async () => {
      const testUserId = crypto.randomUUID();

      const pastLog: CreateFoodLogEntry = {
        name: "My Food Log",
        labels: [],
        time: {
          start: new Date(1999, 10, 10),
          end: new Date(1999, 10, 11),
        },
        metrics: {
          calories: 500,
        },
      };

      const centerLog: CreateFoodLogEntry = {
        name: "My Food Log",
        labels: [],
        time: {
          start: new Date(1999, 10, 15),
          end: new Date(1999, 10, 16),
        },
        metrics: {
          calories: 500,
        },
      };

      const futureLog: CreateFoodLogEntry = {
        name: "My Food Log",
        labels: [],
        time: {
          start: new Date(1999, 10, 20),
          end: new Date(1999, 10, 21),
        },
        metrics: {
          calories: 500,
        },
      };

      const past = await (config.storage.foodLog.storeFoodLog as any)(
        testUserId,
        pastLog,
        testClient
      );
      const pastItemId = past._unsafeUnwrap();

      const result = await (config.storage.foodLog.storeFoodLog as any)(
        testUserId,
        centerLog,
        testClient
      );
      const centerItemId = result._unsafeUnwrap();

      const future = await (config.storage.foodLog.storeFoodLog as any)(
        testUserId,
        futureLog,
        testClient
      );
      const futureItemId = future._unsafeUnwrap();

      const startingQueryResult = await (
        config.storage.foodLog.queryFoodLogs as any
      )(testUserId, new Date(1999, 10, 15), new Date(1999, 10, 16), testClient);

      expect(startingQueryResult.isOk()).toBeTruthy();
      const firstTest = startingQueryResult._unsafeUnwrap();
      expect(firstTest.length).toBe(1);
      expect(firstTest[0].id).toBe(centerItemId);

      const pastQueryResult = await (
        config.storage.foodLog.queryFoodLogs as any
      )(testUserId, new Date(1999, 10, 9), new Date(1999, 10, 16), testClient);

      expect(pastQueryResult.isOk()).toBeTruthy();
      const secondTest = pastQueryResult._unsafeUnwrap();
      expect(secondTest.length).toBe(2);
      expect(secondTest.map((x: any) => x.id).sort()).toEqual(
        [pastItemId, centerItemId].sort()
      );

      const futureQueryResult = await (
        config.storage.foodLog.queryFoodLogs as any
      )(testUserId, new Date(1999, 10, 15), new Date(1999, 10, 30), testClient);

      expect(futureQueryResult.isOk()).toBeTruthy();
      const thirdTest = futureQueryResult._unsafeUnwrap();
      expect(thirdTest.length).toBe(2);
      expect(thirdTest.map((x: any) => x.id).sort()).toEqual(
        [centerItemId, futureItemId].sort()
      );

      const wildQueryResult = await (
        config.storage.foodLog.queryFoodLogs as any
      )(testUserId, new Date(2012, 0, 1), new Date(2012, 11, 31), testClient);

      expect(startingQueryResult.isOk()).toBeTruthy();
      const wildTest = wildQueryResult._unsafeUnwrap();
      expect(wildTest.length).toBe(0);
    });

    test.skip("Bulk Actions :: Can dump logs to temp file", async () => {
      const testUserId = crypto.randomUUID();
      let logs: any[] = [];

      for (let i = 0; i < 1000; i++) {
        const pastLog: CreateFoodLogEntry = {
          name: "My Food Log " + i,
          labels: ["some-label-" + i],
          time: {
            start: new Date(1999, 10, 10),
            end: new Date(1999, 10, 11),
          },
          metrics: {
            calories: 500 + 1,
          },
        };

        const past = await (config.storage.foodLog.storeFoodLog as any)(
          testUserId,
          pastLog,
          testClient
        );
        const pastItemId = past._unsafeUnwrap();
        logs.push({
          id: pastItemId,
          name: pastLog.name,
          labels: pastLog.labels,
          timeStart: pastLog.time.start.toISOString(),
          timeEnd: pastLog.time.end.toISOString(),
          metrics: pastLog.metrics,
        });
      }

      const bulkFilepathResult = await (
        config.storage.foodLog.bulkExportFoodLogs as any
      )(testUserId, testClient);
      const tempFilename = bulkFilepathResult._unsafeUnwrap();
      const filedata = fs.readFileSync(tempFilename);
      const filedataString = filedata.toString("utf8");
      const records = parse(filedataString, {
        columns: true,
        skip_empty_lines: true,
      }).map((x: any) => {
        return {
          ...x,
          metrics: JSON.parse(x.metrics),
          labels: JSON.parse(x.labels),
        };
      }) as any[];
      records.sort((a, b) =>
        a.name.normalize().localeCompare(b.name.normalize())
      );
      logs.sort((a, b) => a.name.normalize().localeCompare(b.name.normalize()));
      expect(logs.length).toBe(1000);
      expect(records.length).toBe(1000);
      expect(records).toStrictEqual(logs);
    });
  }
);
