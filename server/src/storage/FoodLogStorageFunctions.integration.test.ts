import crypto from "node:crypto";
import { CreateFoodLogEntry, isNotFoundError } from "./types";

import { configs } from "./_testConfigs";

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
    afterAll(async () => {});
    test("Happy Path :: Bad Retreives, Creates, Retreives, Edits, Reretrieves, Deletes, Fails Retreive, Redelete succeeds false", async () => {
      const testUserId = crypto.randomUUID();

      const randomId = crypto.randomUUID();

      const badResult = await (config.storage.retrieveFoodLog as any)(
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

      const result = await (config.storage.storeFoodLog as any)(
        testUserId,
        input,
        testClient
      );

      expect(result.isOk()).toBeTruthy();
      const testItemId = result._unsafeUnwrap();
      expect(testItemId.length).toBeGreaterThan(0);

      const storedItemResult = await (config.storage.retrieveFoodLog as any)(
        testUserId,
        testItemId,
        testClient
      );

      expect(storedItemResult.isOk()).toBeTruthy();
      const storedItem = storedItemResult._unsafeUnwrap();
      expect(storedItem).toEqual({ id: testItemId, ...input });

      storedItem.name = "Modified Food Log";
      storedItem.metrics = {
        calories: 400,
      };

      const modifiedResult = await (config.storage.editFoodLog as any)(
        testUserId,
        storedItem,
        testClient
      );
      expect(modifiedResult.isOk()).toBeTruthy();
      const modified = modifiedResult._unsafeUnwrap();
      expect(modified).toEqual(storedItem);

      const reretrievedItemResult = await (
        config.storage.retrieveFoodLog as any
      )(testUserId, testItemId, testClient);

      expect(reretrievedItemResult.isOk()).toBeTruthy();
      const reretreived = reretrievedItemResult._unsafeUnwrap();
      expect(reretreived).toEqual(storedItem);

      const deleteResult = await (config.storage.deleteFoodLog as any)(
        testUserId,
        testItemId,
        testClient
      );

      expect(deleteResult.isOk()).toBeTruthy();
      expect(deleteResult._unsafeUnwrap()).toBeTruthy();

      const postDeleteRetrieve = await (config.storage.retrieveFoodLog as any)(
        testUserId,
        testItemId,
        testClient
      );

      expect(postDeleteRetrieve.isErr()).toBeTruthy();
      expect(
        isNotFoundError(postDeleteRetrieve._unsafeUnwrapErr())
      ).toBeTruthy();

      const redeleteResult = await (config.storage.deleteFoodLog as any)(
        testUserId,
        testItemId,
        testClient
      );

      expect(redeleteResult.isOk()).toBeTruthy();
      expect(redeleteResult._unsafeUnwrap()).toBeTruthy();
    });

    test("Queries :: can add some logs, and get expected query results", async () => {
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

      const past = await (config.storage.storeFoodLog as any)(
        testUserId,
        pastLog,
        testClient
      );
      const pastItemId = past._unsafeUnwrap();

      const result = await (config.storage.storeFoodLog as any)(
        testUserId,
        centerLog,
        testClient
      );
      const centerItemId = result._unsafeUnwrap();

      const future = await (config.storage.storeFoodLog as any)(
        testUserId,
        futureLog,
        testClient
      );
      const futureItemId = future._unsafeUnwrap();

      const startingQueryResult = await (config.storage.queryFoodLogs as any)(
        testUserId,
        new Date(1999, 10, 15),
        new Date(1999, 10, 16),
        testClient
      );

      expect(startingQueryResult.isOk()).toBeTruthy();
      const firstTest = startingQueryResult._unsafeUnwrap();
      expect(firstTest.length).toBe(1);
      expect(firstTest[0].id).toBe(centerItemId);

      const pastQueryResult = await (config.storage.queryFoodLogs as any)(
        testUserId,
        new Date(1999, 10, 9),
        new Date(1999, 10, 16),
        testClient
      );

      expect(pastQueryResult.isOk()).toBeTruthy();
      const secondTest = pastQueryResult._unsafeUnwrap();
      expect(secondTest.length).toBe(2);
      expect(secondTest.map((x: any) => x.id).sort()).toEqual(
        [pastItemId, centerItemId].sort()
      );

      const futureQueryResult = await (config.storage.queryFoodLogs as any)(
        testUserId,
        new Date(1999, 10, 15),
        new Date(1999, 10, 30),
        testClient
      );

      expect(futureQueryResult.isOk()).toBeTruthy();
      const thirdTest = futureQueryResult._unsafeUnwrap();
      expect(thirdTest.length).toBe(2);
      expect(thirdTest.map((x: any) => x.id).sort()).toEqual(
        [centerItemId, futureItemId].sort()
      );

      const wildQueryResult = await (config.storage.queryFoodLogs as any)(
        testUserId,
        new Date(2012, 0, 1),
        new Date(2012, 11, 31),
        testClient
      );

      expect(startingQueryResult.isOk()).toBeTruthy();
      const wildTest = wildQueryResult._unsafeUnwrap();
      expect(wildTest.length).toBe(0);
    });
  }
);
