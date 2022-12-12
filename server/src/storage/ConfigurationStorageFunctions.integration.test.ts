import crypto from "node:crypto";
import { MetricsConfiguration, SummaryConfiguration } from "../types";
import { isNotFoundError } from "./types";
import { configs } from "./_testConfigs";

describe.each(configs)(
  "$name Configuration Storage Integration Tests",
  ({ config }) => {
    let testClient: any;
    beforeAll(async () => {
      testClient = await config.beforeAllSetup();
    });
    afterAll(async () => {
      await config.afterAllTeardown(testClient);
    });
    test("Happy Path :: Bad Retreives, Upserts, Retreives, Upserts, Reretrieves, Deletes, Fails Retreive", async () => {
      const testUserId = crypto.randomUUID();

      const badResult = await (
        config.storage.configuration.retrieveUserConfiguration as any
      )(testUserId, "metrics", testClient);

      expect(badResult.isErr()).toBeTruthy();
      expect(isNotFoundError(badResult._unsafeUnwrapErr())).toBeTruthy();

      const upsertConfig: MetricsConfiguration = {
        id: "metrics",
        value: {
          calories: {
            label: "Calories",
            priority: 0,
          },
        },
      };
      const response = await (
        config.storage.configuration.storeConfiguration as any
      )(testUserId, upsertConfig, testClient);

      expect(response.isOk()).toBeTruthy();
      expect(response._unsafeUnwrap()).toBe("metrics");

      const reretreive = await (
        config.storage.configuration.retrieveUserConfiguration as any
      )(testUserId, "metrics", testClient);

      expect(reretreive.isOk()).toBeTruthy();
      expect(reretreive._unsafeUnwrap()).toStrictEqual(upsertConfig);

      await (config.storage.configuration.deleteUserConfiguration as any)(
        testUserId,
        "metrics",
        testClient
      );

      expect(reretreive.isOk()).toBeTruthy();

      const secBadResult = await (
        config.storage.configuration.retrieveUserConfiguration as any
      )(testUserId, "metrics", testClient);

      expect(secBadResult.isErr()).toBeTruthy();
      expect(isNotFoundError(secBadResult._unsafeUnwrapErr())).toBeTruthy();
    });

    test("Happy Path :: Multistore", async () => {
      const testUserId = crypto.randomUUID();

      const empty = await (
        config.storage.configuration.queryUserConfiguration as any
      )(testUserId, testClient);

      expect(empty.isOk()).toBeTruthy();
      expect(empty._unsafeUnwrap()).toEqual([]);

      const metrics: MetricsConfiguration = {
        id: "metrics",
        value: {
          calories: {
            label: "Calories",
            priority: 0,
          },
        },
      };
      const summary: SummaryConfiguration = {
        id: "summaries",
        value: { showMetricSummary: ["calories"] },
      };
      await (config.storage.configuration.storeConfiguration as any)(
        testUserId,
        metrics,
        testClient
      );
      await (config.storage.configuration.storeConfiguration as any)(
        testUserId,
        summary,
        testClient
      );

      const filled = await (
        config.storage.configuration.queryUserConfiguration as any
      )(testUserId, testClient);

      expect(filled.isOk()).toBeTruthy();
      expect(filled._unsafeUnwrap()).toEqual([metrics, summary]);
    });
  }
);
