import {
  NotFoundError,
  StorageError,
  SystemError,
  ValidationError,
} from "../types";
import { Result, err, ok } from "neverthrow";
import { Configuration } from "../../types";
import { ConfigurationStorage } from "../types/Configuration";
import { isValidConfigurationItem } from "../validation";
import { NEO4J_DRIVER } from ".";

export const configuration: ConfigurationStorage = {
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
          MERGE (u)-[r:HAS_CONFIG { type: $type }]->(c)
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
        `MATCH (:User {userid: $userid})-[r:HAS_CONFIG]-(c:Config)
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
        `MATCH (:User {userid: $userid})-[r:HAS_CONFIG { type: $type }]-(c:Config)
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
        `MATCH (:User {userid: $userid})-[:HAS_CONFIG { type: $type }]-(c:Config)
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
