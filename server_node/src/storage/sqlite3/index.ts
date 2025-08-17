import knex, { Knex } from "knex";
import * as sqliteFoodLogStorage from "./FoodLogStorageFunctions";
import * as sqliteConfigurationStorage from "./ConfigurationStorageFunctions";

export const DEFAULT_CLIENT_CONFIG: Knex.Config = {
  client: "better-sqlite3",
  connection: {
    filename:
      process.env.OPENFOODDIARY_SQLITE3_FILENAME ??
      ".sqlite/openfooddiary.sqlite"
  },
  useNullAsDefault: true
};

export const knexInstance = knex(DEFAULT_CLIENT_CONFIG);

export async function setupDatabase (knex: Knex = knexInstance) {
  const migrations = [
    `CREATE TABLE IF NOT EXISTS user_foodlogentry 
  (
    user_id TEXT, --UUID 
    id TEXT, --UUID 
    name TEXT, 
    labels TEXT, --set<text>
    metrics TEXT, --map<text,int>
    time_start DATETIME,
    time_end DATETIME
  );`,
    `CREATE TABLE IF NOT EXISTS user_config 
  (
    user_id TEXT, --UUID 
    id TEXT,
    serialised_value TEXT
  );`
  ];
  for (const migration of migrations) {
    await knex.raw(migration);
  }
}

export async function shutdownDatabase (knex: Knex = knexInstance) {
  await knex.destroy();
}

export const sqlite3 = {
  // StorageType
  setupDatabase,
  shutdownDatabase,
  foodLog: sqliteFoodLogStorage,
  configuration: sqliteConfigurationStorage
};
