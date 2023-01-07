import * as neo4j from "neo4j-driver";
import { StorageType } from "../interfaces";
import {
  CreateFoodLogEntry,
  EditFoodLogEntry,
  FoodLogStorage,
  StorageError,
  ValidationError,
} from "../types";
import { Result, err } from "neverthrow";
import { Configuration, FoodLogEntry } from "../../types";
import { ConfigurationStorage } from "../types/Configuration";
import {
  isValidConfigurationItem,
  isValidCreateLogEntry,
  isValidEditLogEntry,
} from "../validation";

// Create a driver instance, for the user `neo4j` with password `password`.
// It should be enough to have a single driver per database per application.
export const NEO4J_DRIVER = neo4j.driver(
  "neo4j://localhost:7687",
  neo4j.auth.basic("neo4j", "s3cr3tly")
);

// Close the driver when application exits.
// This closes all used network connections.

const foodLog: FoodLogStorage = {
  storeFoodLog: function (
    userId: string,
    logEntry: CreateFoodLogEntry
  ): Promise<Result<string, StorageError>> {
    if (!isValidCreateLogEntry(logEntry)) {
      return Promise.resolve(err(new ValidationError("Invalid Log Entry")));
    }
    throw new Error("Function not implemented.");
  },
  retrieveFoodLog: function (
    userId: string,
    logId: string
  ): Promise<Result<FoodLogEntry, StorageError>> {
    throw new Error("Function not implemented.");
  },
  editFoodLog: function (
    userId: string,
    logEntry: EditFoodLogEntry
  ): Promise<Result<FoodLogEntry, StorageError>> {
    if (!isValidEditLogEntry(logEntry)) {
      return Promise.resolve(err(new ValidationError("Invalid Log Entry")));
    }
    throw new Error("Function not implemented.");
  },
  deleteFoodLog: function (
    userId: string,
    logId: string
  ): Promise<Result<boolean, StorageError>> {
    throw new Error("Function not implemented.");
  },
  queryFoodLogs: function (
    userId: string,
    dateStart: Date,
    dateEnd: Date
  ): Promise<Result<FoodLogEntry[], StorageError>> {
    if (dateEnd.getTime() < dateStart.getTime()) {
      return Promise.resolve(
        err(new ValidationError("startDate is after endDate"))
      );
    }
    throw new Error("Function not implemented.");
  },
  bulkExportFoodLogs: function (
    userId: string
  ): Promise<Result<string, StorageError>> {
    throw new Error("Function not implemented.");
  },
};

const configuration: ConfigurationStorage = {
  storeConfiguration: async function (
    userId: string,
    configuration: Configuration
  ): Promise<Result<"metrics" | "summaries", StorageError>> {
    if (!userId) {
      return err(new ValidationError("Invalid user id"));
    }
    if (!isValidConfigurationItem(configuration)) {
      return err(new ValidationError("Error with Configuration"));
    }
    throw new Error("Function not implemented.");
  },
  queryUserConfiguration: function (
    userId: string
  ): Promise<Result<Configuration[], StorageError>> {
    throw new Error("Function not implemented.");
  },
  retrieveUserConfiguration: function (
    userId: string,
    configurationId: "metrics" | "summaries"
  ): Promise<Result<Configuration, StorageError>> {
    throw new Error("Function not implemented.");
  },
  deleteUserConfiguration: function (
    userId: string,
    configurationId: "metrics" | "summaries"
  ): Promise<Result<boolean, StorageError>> {
    throw new Error("Function not implemented.");
  },
};

export const storageConfig: StorageType = {
  setupDatabase: async () => {},
  shutdownDatabase: async () => {
    await NEO4J_DRIVER.close();
  },
  foodLog,
  configuration,
};
