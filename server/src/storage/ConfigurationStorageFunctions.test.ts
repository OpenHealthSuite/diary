import { sqlite3 } from "./sqlite3";
import { isValidationError } from "./types";
import { ConfigurationStorage } from "./types/Configuration";
import { Configuration } from "../types";
import { cassandra } from "./cassandra";

const configs = [
  { name: "sqlite3", config: sqlite3.configuration },
  { name: "cassandra", config: cassandra.configuration },
];

describe.each(configs)(
  "$name ConfigurationStorageFunctions",
  ({ config: storeConfig }: { config: ConfigurationStorage }) => {
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
        // {
        //   reason: "bad key",
        //   config: {
        //     id: "metrics",
        //     value: {
        //       "calories in": {
        //         label: "Calories",
        //         priority: 0,
        //       },
        //     },
        //   },
        // },
        // {
        //   reason: "bad key",
        //   config: {
        //     id: "metrics",
        //     value: {
        //       Calories: {
        //         label: "Calories",
        //         priority: 0,
        //       },
        //     },
        //   },
        // },
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
