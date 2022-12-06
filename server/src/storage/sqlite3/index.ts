import knex, { Knex } from "knex";
import { StorageType } from "../interfaces";
import * as sqliteFoodLogStorage from "./FoodLogStorageFunctions";

export const DEFAULT_CLIENT_CONFIG: Knex.Config = {
  client: "better-sqlite3",
  connection: {
    filename: ".sqlite/openfooddiary.sqlite",
  },
  useNullAsDefault: true,
};

export const knexInstance = knex(DEFAULT_CLIENT_CONFIG);

export async function setupDatabase(knex: Knex = knexInstance) {
  const setup = `CREATE TABLE IF NOT EXISTS user_foodlogentry 
  (
    user_id TEXT, --UUID 
    id TEXT, --UUID 
    name TEXT, 
    labels TEXT, --set<text>
    metrics TEXT, --map<text,int>
    time_start DATETIME,
    time_end DATETIME
  );`;

  await knex.raw(setup);
}

export const sqlite3: StorageType = {
  setupDatabase,
  shutdownDatabase: () => Promise.resolve(),
  foodLog: sqliteFoodLogStorage,
};
