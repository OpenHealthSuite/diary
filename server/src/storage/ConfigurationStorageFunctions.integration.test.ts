import crypto from "node:crypto";
import { Configuration, MetricsConfiguration, SummaryConfiguration } from "../types";
import { isNotFoundError, isValidationError } from "./types";
import { configs } from "./_testConfigs";
import { ConfigurationStorage } from "./types/Configuration";

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


describe.each(configs)(
  "$name ConfigurationStorageFunctions Validation",
  ({ config: { storage: { configuration : storeConfig  } }}) => {
    describe("UpsertConfiguration", () => {
      const BAD_USER_IDS = [undefined, null, ""];
      const BAD_CONFIGS: { reason: string; config: Configuration }[] = [
        {
          reason: "random key",
          config: {
            id: "not-a-real-config-key",
            value: {},
          } as any,
        },
        {
          reason: "duplicate priority",
          config: {
            id: "metrics",
            value: {
              calories: {
                label: "Calories",
                priority: 0,
              },
              hydration: {
                label: "Hydrations",
                priority: 0,
              },
            },
          },
        },
      ];
      test.each(BAD_USER_IDS)("%s user id rejected", async (userId) => {
        const result = await storeConfig.storeConfiguration(userId as any, {
          id: "metrics",
          value: {
            calories: {
              label: "Calories",
              priority: 0,
            },
          },
        });
        expect(result.isErr()).toBeTruthy();
        expect(isValidationError(result._unsafeUnwrapErr())).toBeTruthy();
        expect(result._unsafeUnwrapErr().message).toBe("Invalid user id");
      });

      test.each(BAD_CONFIGS)(
        "$reason rejected",
        async ({ config }: { config: Configuration }) => {
          const result = await storeConfig.storeConfiguration(
            "test-user-id",
            config
          );
          expect(result.isErr()).toBeTruthy();
          expect(isValidationError(result._unsafeUnwrapErr())).toBeTruthy();
          expect(result._unsafeUnwrapErr().message).toBe(
            "Error with Configuration"
          );
        }
      );
    });
  }
);
