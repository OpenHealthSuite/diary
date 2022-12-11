import * as sqlite3 from "./sqlite3";
import * as cassandra from "./cassandra";
import knex from "knex";
import { Client } from "cassandra-driver";

export const configs = [
  {
    name: "sqlite3",
    config: {
      beforeAllSetup: async () => {
        const testClient = knex({
          ...sqlite3.DEFAULT_CLIENT_CONFIG,
          connection: { filename: ":memory:" },
        });
        await sqlite3.setupDatabase(testClient);
        return testClient;
      },
      afterAllTeardown: async (testClient: any) => {
        await testClient.destroy();
      },
      storage: sqlite3.sqlite3.foodLog,
    },
  },
  {
    name: "cassandra",
    config: {
      beforeAllSetup: async () => {
        await cassandra.setupDatabase(
          new Client(cassandra.DEFAULT_CLIENT_CONFIG)
        );
        return new Client(cassandra.DEFAULT_CLIENT_CONFIG);
      },
      afterAllTeardown: async (testClient: any) => {
        await testClient.shutdown();
      },
      storage: cassandra.cassandra.foodLog,
    },
  },
];
