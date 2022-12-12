import { err, ok, Result } from "neverthrow";
import { Configuration } from "../../types";
import { CASSANDRA_CLIENT } from ".";
import { isValidConfigurationItem } from "../validation";
import { UpsertConfiguration } from "../types/Configuration";
import {
  NotFoundError,
  StorageError,
  SystemError,
  ValidationError,
} from "../types";

type CassandraConfig = {
  user_id: string;
  id: Configuration["id"];
  serialised_value: string;
};

function sqlFromGeneral(user_id: string, gen: Configuration): CassandraConfig {
  return {
    user_id,
    id: gen.id,
    serialised_value: JSON.stringify(gen.value),
  };
}

function generalFromSql(sql: CassandraConfig): Configuration {
  return {
    id: sql.id,
    value: JSON.parse(sql.serialised_value),
  };
}

export async function storeConfiguration(
  userId: string,
  configuration: UpsertConfiguration,
  client = CASSANDRA_CLIENT
): Promise<Result<Configuration["id"], StorageError>> {
  if (!userId) {
    return err(new ValidationError("Invalid user id"));
  }
  if (!isValidConfigurationItem(configuration)) {
    return err(new ValidationError("Error with Configuration"));
  }
  const insertEntry: CassandraConfig = sqlFromGeneral(userId, configuration);
  try {
    await client.execute(
      `INSERT INTO openfooddiary.user_config (user_id, id, serialised_value)
      VALUES (?, ?, ?);`,
      [insertEntry.user_id, insertEntry.id, insertEntry.serialised_value],
      { prepare: true }
    );
    return ok(insertEntry.id);
  } catch (error: any) {
    return err(new SystemError(error.message));
  }
}

export async function queryUserConfiguration(
  userId: string,
  client = CASSANDRA_CLIENT
): Promise<Result<Configuration[], StorageError>> {
  try {
    const result = await client.execute(
      `SELECT * FROM openfooddiary.user_config
      WHERE user_id = ?;`,
      [userId],
      { prepare: true }
    );
    const constructed: any[] = result.rows.map((row) => {
      const subCon: any = {};
      row.keys().forEach((key) => (subCon[key] = row.get(key)));
      return subCon;
    });
    return ok(constructed.map(generalFromSql));
  } catch (error: any) {
    return err(new SystemError(error.message));
  }
}

export async function retrieveUserConfiguration(
  userId: string,
  configurationId: string,
  client = CASSANDRA_CLIENT
): Promise<Result<Configuration, StorageError>> {
  try {
    const result = await client.execute(
      `SELECT * FROM openfooddiary.user_config
      WHERE user_id = ? AND id = ?;`,
      [userId, configurationId],
      { prepare: true }
    );
    if (result.rows.length == 0) {
      return err(new NotFoundError("No Config Found"));
    }
    const item = result.first();
    const constructed: any = {};
    item.keys().forEach((key) => (constructed[key] = item.get(key)));
    return ok(generalFromSql(constructed));
  } catch (error: any) {
    return err(new SystemError(error.message));
  }
}

export async function deleteUserConfiguration(
  userId: string,
  configurationId: Configuration["id"],
  client = CASSANDRA_CLIENT
): Promise<Result<boolean, StorageError>> {
  try {
    await client.execute(
      `DELETE FROM openfooddiary.user_config
      WHERE user_id = ? AND id = ?;`,
      [userId, configurationId],
      { prepare: true }
    );
    return ok(true);
  } catch (error: any) {
    return err(new SystemError(error.message));
  }
}
