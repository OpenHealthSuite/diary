import * as neo4j from "neo4j-driver";
import { StorageType } from "../interfaces";
import {
  CreateFoodLogEntry,
  EditFoodLogEntry,
  FoodLogStorage,
  NotFoundError,
  StorageError,
  SystemError,
  ValidationError,
} from "../types";
import { Result, err, ok } from "neverthrow";
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
    const session = NEO4J_DRIVER.session();
    try {
      await session.run(
        `MERGE (u:User {userid: $userid})
          MERGE (c:Config { value: $value })
          MERGE (u)-[r:IS_CONFIG_FOR { type: $type }]->(c)
          RETURN r.type`,
        {
          userid: userId,
          value: JSON.stringify(configuration.value),
          type: configuration.id,
        }
      );

      await session.close();
      return ok(configuration.id);
    } catch (error: any) {
      await session.close();
      console.error(error);
      return err(new SystemError(error.message));
    }
  },
  queryUserConfiguration: async function (
    userId: string
  ): Promise<Result<Configuration[], StorageError>> {
    const session = NEO4J_DRIVER.session();
    try {
      const res = await session.run<{ id: string; value: string }>(
        `MATCH (:User {userid: $userid})-[r:IS_CONFIG_FOR]-(c:Config)
          RETURN r.type as id, c.value as value`,
        {
          userid: userId,
        }
      );
      await session.close();

      return ok(
        res.records.map((rec) => {
          return {
            id: rec.get("id"),
            value: JSON.parse(rec.get("value")),
          } as Configuration;
        })
      );
    } catch (error: any) {
      await session.close();
      console.error(error);
      return err(new SystemError(error.message));
    }
  },
  retrieveUserConfiguration: async function (
    userId: string,
    configurationId: "metrics" | "summaries"
  ): Promise<Result<Configuration, StorageError>> {
    const session = NEO4J_DRIVER.session();
    try {
      const res = await session.run<{ id: string; value: string }>(
        `MATCH (:User {userid: $userid})-[r:IS_CONFIG_FOR { type: $type }]-(c:Config)
          RETURN r.type as id, c.value as value
          LIMIT 1`,
        {
          userid: userId,
          type: configurationId,
        }
      );
      await session.close();
      if (res.records.length == 0) {
        return err(new NotFoundError("Not Found"));
      }

      return ok(
        res.records.map((rec) => {
          return {
            id: rec.get("id"),
            value: JSON.parse(rec.get("value")),
          } as Configuration;
        })[0]
      );
    } catch (error: any) {
      await session.close();
      console.error(error);
      return err(new SystemError(error.message));
    }
  },
  deleteUserConfiguration: async function (
    userId: string,
    configurationId: "metrics" | "summaries"
  ): Promise<Result<boolean, StorageError>> {
    const session = NEO4J_DRIVER.session();
    try {
      const res = await session.run<{ id: string; value: string }>(
        `MATCH (:User {userid: $userid})-[:IS_CONFIG_FOR { type: $type }]-(c:Config)
          DETACH DELETE c`,
        {
          userid: userId,
          type: configurationId,
        }
      );
      await session.close();
      return ok(true);
    } catch (error: any) {
      await session.close();
      console.error(error);
      return err(new SystemError(error.message));
    }
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
