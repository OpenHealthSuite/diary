import { cassandra } from "./cassandra";
import { StorageType } from "./interfaces";
import { sqlite3 } from "./sqlite3";

export * from "./types/StorageErrors";

export interface StorageDrivers {
  [key: string]: StorageType;
}

const AVAILABLE_STORAGES: StorageDrivers = {
  cassandra,
  sqlite3
};

let setupStorage = AVAILABLE_STORAGES.sqlite3;

switch (process.env.OPENFOODDIARY_STORAGE_PROVIDER) {
  case "cassandra":
    setupStorage = AVAILABLE_STORAGES.cassandra;
    break;
}

export const STORAGE = setupStorage;
