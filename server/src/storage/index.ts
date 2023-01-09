export * from "./types/StorageErrors";

import { cassandra } from "./cassandra";
import { StorageType } from "./interfaces";
import { sqlite3 } from "./sqlite3";
import { storageConfig as neo4j } from "./neo4j"

export interface StorageDrivers {
  [key: string]: StorageType;
}

const AVAILABLE_STORAGES: StorageDrivers = {
  cassandra,
  sqlite3,
  neo4j
};

let setupStorage = AVAILABLE_STORAGES.sqlite3;

switch (process.env.OPENFOODDIARY_STORAGE_PROVIDER) {
  case "cassandra":
    setupStorage = AVAILABLE_STORAGES.cassandra;
    break;
  case "neo4j":
    setupStorage = AVAILABLE_STORAGES.neo4j;
    break;
}

export const STORAGE = setupStorage;
