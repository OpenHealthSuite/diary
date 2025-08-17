import { err, ok, Result } from "neverthrow";
import { Configuration } from "../../types";
import { knexInstance } from ".";
import { isValidConfigurationItem } from "../validation";
import { UpsertConfiguration } from "../types/Configuration";
import {
  NotFoundError,
  StorageError,
  SystemError,
  ValidationError
} from "../types";

type SqliteConfig = {
  user_id: string;
  id: Configuration["id"];
  serialised_value: string;
};

function sqlFromGeneral (userId: string, gen: Configuration): SqliteConfig {
  return {
    user_id: userId,
    id: gen.id,
    serialised_value: JSON.stringify(gen.value)
  };
}

function generalFromSql (sql: SqliteConfig): Configuration {
  return {
    id: sql.id,
    value: JSON.parse(sql.serialised_value)
  };
}

export async function storeConfiguration (
  userId: string,
  configuration: UpsertConfiguration,
  client = knexInstance
): Promise<Result<Configuration["id"], StorageError>> {
  if (!userId) {
    return err(new ValidationError("Invalid user id"));
  }
  if (!isValidConfigurationItem(configuration)) {
    return err(new ValidationError("Error with Configuration"));
  }
  const insertEntry: SqliteConfig = sqlFromGeneral(userId, configuration);
  try {
    // This is because sqlite knex doesn't support upserts...
    await client!("user_config")
      .delete()
      .where("user_id", userId)
      .andWhere("id", configuration.id);
    await client!("user_config").insert(sqlFromGeneral(userId, configuration));
    return ok(insertEntry.id);
  } catch (error: any) {
    console.error(error.message);
    return err(new SystemError(error.message));
  }
}

export async function queryUserConfiguration (
  userId: string,
  client = knexInstance
): Promise<Result<Configuration[], StorageError>> {
  try {
    const log = await client<SqliteConfig>("user_config")
      .select()
      .where("user_id", userId);
    return ok(log.map(generalFromSql));
  } catch (error: any) {
    console.error(error.message);
    return err(new SystemError(error.message));
  }
}

export async function retrieveUserConfiguration (
  userId: string,
  configurationId: string,
  client = knexInstance
): Promise<Result<Configuration, StorageError>> {
  try {
    const log = await client<SqliteConfig>("user_config")
      .select()
      .where("user_id", userId)
      .andWhere("id", configurationId)
      .first();
    if (log === undefined) {
      return err(new NotFoundError("Log not found"));
    }
    return ok(generalFromSql(log));
  } catch (error: any) {
    console.error(error.message);
    return err(new SystemError(error.message));
  }
}

export async function deleteUserConfiguration (
  userId: string,
  configurationId: Configuration["id"],
  client = knexInstance
): Promise<Result<boolean, StorageError>> {
  try {
    await client<SqliteConfig>("user_config")
      .delete()
      .where("user_id", userId)
      .andWhere("id", configurationId);
    return ok(true);
  } catch (error: any) {
    console.error(error.message);
    return err(new SystemError(error.message));
  }
}
