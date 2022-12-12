import { Client } from "cassandra-driver";
import { StorageType } from "../interfaces";
import { ConfigurationStorage } from "../types/Configuration";
import * as cassandraFoodLogStorage from "./FoodLogStorageFunctions";
import * as cassandraConfigStorage from "./ConfigurationStorageFunctions";

export const DEFAULT_CLIENT_CONFIG = {
  contactPoints: process.env.OPENFOODDIARY_CASSANDRA_CONTACT_POINTS
    ? process.env.OPENFOODDIARY_CASSANDRA_CONTACT_POINTS.split(";")
    : ["localhost:9042"],
  localDataCenter:
    process.env.OPENFOODDIARY_CASSANDRA_LOCALDATACENTER ?? "datacenter1",
  credentials: {
    username: process.env.OPENFOODDIARY_CASSANDRA_USER ?? "cassandra",
    password: process.env.OPENFOODDIARY_CASSANDRA_PASSWORD ?? "cassandra",
  },
};

export const CASSANDRA_CLIENT: Client = new Client(DEFAULT_CLIENT_CONFIG);

export async function shutdownDatabase() {
  await CASSANDRA_CLIENT.shutdown();
}

// Probably worth looking into this more formally
const MIGRATIONS: string[] = [
  "CREATE KEYSPACE IF NOT EXISTS openfooddiary WITH REPLICATION = {'class':'SimpleStrategy','replication_factor':1};", // TODO: Make this optional?
  `CREATE TYPE IF NOT EXISTS openfooddiary.logTimes (start timestamp, end timestamp);`,
  `CREATE TABLE IF NOT EXISTS openfooddiary.user_foodlogentry (userId text, id UUID, name text, labels set<text>,time frozen<openfooddiary.logTimes>,metrics map<text,int>, timeStart timestamp, timeEnd timestamp, PRIMARY KEY ((userId), id));`,
  `CREATE INDEX IF NOT EXISTS ON openfooddiary.user_foodlogentry (timeStart);`,
  `CREATE INDEX IF NOT EXISTS ON openfooddiary.user_foodlogentry (timeEnd);`,
  `CREATE TABLE IF NOT EXISTS openfooddiary.user_config (user_id text, id text, serialised_value text, PRIMARY KEY ((user_id), id));`,
];

export async function setupDatabase(
  migrationClient: Client = new Client(DEFAULT_CLIENT_CONFIG)
) {
  await migrationClient.connect().then(async () => {
    for (let migration of MIGRATIONS) {
      await migrationClient.execute(migration);
    }
    return migrationClient.shutdown();
  });
}

export const cassandra: StorageType = {
  setupDatabase,
  shutdownDatabase,
  foodLog: cassandraFoodLogStorage,
  configuration: cassandraConfigStorage,
};
