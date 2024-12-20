import { err, ok, Result } from "neverthrow";
import { FoodLogEntry } from "../../types";
import crypto from "node:crypto";
import {
  BulkExportFoodLogEntry,
  CreateFoodLogEntry,
  EditFoodLogEntry,
  NotFoundError,
  StorageError,
  SystemError,
  ValidationError
} from "../types";
import fs from "node:fs";
import { stringify } from "csv-stringify/sync";
import { CASSANDRA_CLIENT } from ".";
import { Client } from "cassandra-driver";
import { isValidCreateLogEntry, isValidEditLogEntry } from "../validation";

type CassandraStoreFoodLogFunction = (
  userId: string,
  logEntry: CreateFoodLogEntry,
  client?: Client
) => Promise<Result<string, StorageError>>;
type CassandraQueryFoodLogFunction = (
  userId: string,
  dateStart: Date,
  dateEnd: Date,
  client?: Client
) => Promise<Result<FoodLogEntry[], StorageError>>;
type CassandraRetrieveFoodLogFunction = (
  userId: string,
  logId: string,
  client?: Client
) => Promise<Result<FoodLogEntry, StorageError>>;
type CassandraEditFoodLogFunction = (
  userId: string,
  logEntry: EditFoodLogEntry,
  client?: Client
) => Promise<Result<FoodLogEntry, StorageError>>;
type CassandraDeleteFoodLogFunction = (
  userId: string,
  logId: string,
  client?: Client
) => Promise<Result<boolean, StorageError>>;
type CassandraBulkExportFoodLogsFunction = (
  userId: string,
  client?: Client
) => Promise<Result<string, StorageError>>;
type CassandraPurgeFoodLogsFunction = (
  userId: string,
  client?: Client
) => Promise<Result<boolean, StorageError>>;

export const storeFoodLog: CassandraStoreFoodLogFunction = async (
  userId: string,
  logEntry: CreateFoodLogEntry,
  cassandraClient = CASSANDRA_CLIENT
): Promise<Result<string, StorageError>> => {
  if (!isValidCreateLogEntry(logEntry)) {
    return Promise.resolve(err(new ValidationError("Invalid Log Entry")));
  }
  const insertEntry: FoodLogEntry & { userId: string } = {
    ...logEntry,
    id: crypto.randomUUID(),
    userId
  };
  try {
    await cassandraClient.execute(
      `INSERT INTO openfooddiary.user_foodlogentry (userId, id, name, labels, time, metrics, timeStart, timeEnd)
            values (?, ?, ?, ?, ?, ?, ?, ?);`,
      [
        userId,
        insertEntry.id,
        insertEntry.name,
        insertEntry.labels,
        insertEntry.time,
        insertEntry.metrics,
        insertEntry.time.start,
        insertEntry.time.end
      ],
      { prepare: true }
    );
    return ok(insertEntry.id);
  } catch (error: any) {
    console.error(error.message);
    return err(new SystemError(error.message));
  }
};

export const retrieveFoodLog: CassandraRetrieveFoodLogFunction = async (
  userId: string,
  logId: string,
  cassandraClient = CASSANDRA_CLIENT
): Promise<Result<FoodLogEntry, StorageError>> => {
  try {
    const result = await cassandraClient.execute(
      `SELECT CAST(id as text) as id, name, labels, time, metrics 
            FROM  openfooddiary.user_foodlogentry 
            WHERE userId = ? AND id = ?;`,
      [userId, logId],
      { prepare: true }
    );
    if (result.rows.length === 0) {
      return err(new NotFoundError("No Log with Id"));
    }
    const item = result.first();
    const constructed: any = {};
    item.keys().forEach((key) => (constructed[key] = item.get(key)));
    return ok(constructed as FoodLogEntry);
  } catch (error: any) {
    console.error(error.message);
    return err(new SystemError(error.message));
  }
};

export const queryFoodLogs: CassandraQueryFoodLogFunction = async (
  userId: string,
  startDate: Date,
  endDate: Date,
  cassandraClient = CASSANDRA_CLIENT
): Promise<Result<FoodLogEntry[], StorageError>> => {
  if (endDate.getTime() < startDate.getTime()) {
    return Promise.resolve(
      err(new ValidationError("startDate is after endDate"))
    );
  }
  try {
    const result = await cassandraClient.execute(
      `SELECT CAST(id as text) as id, name, labels, time, metrics 
            FROM openfooddiary.user_foodlogentry 
            WHERE userId = ? AND timeStart <= ? AND timeEnd >= ?
            ALLOW FILTERING;`,
      [userId, endDate, startDate],
      { prepare: true }
    );
    const constructed: any[] = result.rows.map((row) => {
      const subCon: any = {};
      row.keys().forEach((key) => (subCon[key] = row.get(key)));
      return subCon;
    });
    return ok(constructed as FoodLogEntry[]);
  } catch (error: any) {
    console.error(error.message);
    return err(new SystemError(error.message));
  }
};

const UPDATEABLE_FIELDS = ["name", "labels", "time", "metrics"];
export const editFoodLog: CassandraEditFoodLogFunction = async (
  userId: string,
  logEntry: EditFoodLogEntry,
  cassandraClient = CASSANDRA_CLIENT
) => {
  if (!isValidEditLogEntry(logEntry)) {
    return Promise.resolve(err(new ValidationError("Invalid Log Entry")));
  }
  const updateEntity: any = {
    ...logEntry,
    userId
  };

  const updatedFields: string[] = [];
  const updateValues: any[] = [];

  UPDATEABLE_FIELDS.forEach((field) => {
    if (updateEntity[field] !== undefined) {
      updatedFields.push(field + " = ?");
      updateValues.push(updateEntity[field]);
      if (field === "time") {
        updatedFields.push("timeStart = ?");
        updatedFields.push("timeEnd = ?");
        updateValues.push(updateEntity.time.start);
        updateValues.push(updateEntity.time.end);
      }
    }
  });
  try {
    await cassandraClient.execute(
      `UPDATE openfooddiary.user_foodlogentry
            SET ${updatedFields.join(",")}
            WHERE userId = ? AND id = ?;`,
      [...updateValues, userId, updateEntity.id],
      { prepare: true }
    );
    return await retrieveFoodLog(userId, logEntry.id, cassandraClient);
  } catch (error: any) {
    console.error(error.message);
    return err(new SystemError(error.message));
  }
};

export const deleteFoodLog: CassandraDeleteFoodLogFunction = async (
  userId: string,
  logId: string,
  cassandraClient = CASSANDRA_CLIENT
): Promise<Result<boolean, StorageError>> => {
  try {
    await cassandraClient.execute(
      `DELETE FROM openfooddiary.user_foodlogentry
            WHERE userId = ? AND id = ?;`,
      [userId, logId],
      { prepare: true }
    );
    return ok(true);
  } catch (error: any) {
    console.error(error.message);
    return err(new SystemError(error.message));
  }
};

const TEMP_DIR = process.env.OPENFOODDIARY_TEMP_DIRECTORY ?? "/tmp";

function bulkFromResult (log: FoodLogEntry): BulkExportFoodLogEntry {
  return {
    id: log.id,
    name: log.name,
    labels: log.labels,
    timeStart: log.time.start.toISOString(),
    timeEnd: log.time.end.toISOString(),
    metrics: log.metrics
  };
}

export const bulkExportFoodLogs: CassandraBulkExportFoodLogsFunction = async (
  userId: string,
  cassandraClient = CASSANDRA_CLIENT
): Promise<Result<string, StorageError>> => {
  try {
    const filename = `${TEMP_DIR}/${crypto.randomUUID()}.csv`;
    fs.writeFileSync(
      filename,
      stringify([["id", "name", "labels", "timeStart", "timeEnd", "metrics"]]),
      {
        flag: "w"
      }
    );

    let morelogs = true;
    let pageState: string | Buffer | undefined;
    while (morelogs) {
      let logs: FoodLogEntry[] = [];
      const result = await cassandraClient.execute(
        `SELECT CAST(id as text) as id, name, labels, time, metrics
              FROM openfooddiary.user_foodlogentry
              WHERE userId = ?;`,
        [userId],
        { prepare: true, pageState, fetchSize: 250 }
      );
      logs = result.rows.map((row) => {
        const subCon: any = {};
        row.keys().forEach((key) => (subCon[key] = row.get(key)));
        return subCon;
      }) as FoodLogEntry[];
      const exportedLogs = logs
        .map(bulkFromResult)
        .map((log) => [
          log.id,
          log.name,
          log.labels,
          log.timeStart,
          log.timeEnd,
          log.metrics
        ]);
      fs.appendFileSync(filename, stringify(exportedLogs));
      if (result.pageState == null) {
        morelogs = false;
      } else {
        pageState = result.pageState;
      }
    }
    return ok(filename);
  } catch (error: any) {
    console.error(error.message);
    return err(new SystemError(error.message));
  }
};

export const purgeFoodLogs: CassandraPurgeFoodLogsFunction = async (
  userId: string,
  cassandraClient = CASSANDRA_CLIENT
): Promise<Result<boolean, StorageError>> => {
  try {
    await cassandraClient.execute(
      `DELETE FROM openfooddiary.user_foodlogentry
      WHERE userId = ?;`,
      [userId],
      { prepare: true }
    );
    return ok(true);
  } catch (error: any) {
    console.error(error.message);
    return err(new SystemError(error.message));
  }
};
