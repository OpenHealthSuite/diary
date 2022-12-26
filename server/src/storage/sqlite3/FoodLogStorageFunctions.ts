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
  ValidationError,
} from "../types";
import { stringify } from "csv-stringify/sync";
import fs from "node:fs";
import { knexInstance } from ".";
import { isValidCreateLogEntry, isValidEditLogEntry } from "../validation";
import { Knex } from "knex";

type SqliteStoreFoodLogFunction = (
  userId: string,
  logEntry: CreateFoodLogEntry,
  client?: Knex
) => Promise<Result<string, StorageError>>;
type SqliteQueryFoodLogFunction = (
  userId: string,
  dateStart: Date,
  dateEnd: Date,
  client?: Knex
) => Promise<Result<FoodLogEntry[], StorageError>>;
type SqliteRetrieveFoodLogFunction = (
  userId: string,
  logId: string,
  client?: Knex
) => Promise<Result<FoodLogEntry, StorageError>>;
type SqliteEditFoodLogFunction = (
  userId: string,
  logEntry: EditFoodLogEntry,
  client?: Knex
) => Promise<Result<FoodLogEntry, StorageError>>;
type SqliteDeleteFoodLogFunction = (
  userId: string,
  logId: string,
  client?: Knex
) => Promise<Result<boolean, StorageError>>;
type SqliteBulkExportFoodLogsFunction = (
  userId: string,
  client?: Knex
) => Promise<Result<string, StorageError>>;

interface SqliteFoodLogEntry {
  user_id: string;
  id: string;
  name: string;
  labels: string;
  metrics: string;
  time_start: number;
  time_end: number;
}

function sqlFromGeneral(
  gen: Pick<FoodLogEntry, "name" | "labels" | "metrics" | "time">
): Pick<
  SqliteFoodLogEntry,
  "name" | "labels" | "metrics" | "time_start" | "time_end"
> {
  return {
    name: gen.name,
    labels: JSON.stringify(gen.labels),
    metrics: JSON.stringify(gen.metrics),
    time_start: new Date(gen.time.start).getTime() / 1000,
    time_end: new Date(gen.time.end).getTime() / 1000,
  };
}

function generalFromSql(sql: SqliteFoodLogEntry): FoodLogEntry {
  return {
    id: sql.id,
    name: sql.name,
    labels: JSON.parse(sql.labels),
    time: {
      start: new Date(sql.time_start * 1000),
      end: new Date(sql.time_end * 1000),
    },
    metrics: JSON.parse(sql.metrics),
  };
}

export const storeFoodLog: SqliteStoreFoodLogFunction = async (
  userId: string,
  logEntry: CreateFoodLogEntry,
  client = knexInstance
): Promise<Result<string, StorageError>> => {
  if (!isValidCreateLogEntry(logEntry)) {
    return Promise.resolve(err(new ValidationError("Invalid Log Entry")));
  }
  const insertEntry: SqliteFoodLogEntry = {
    ...sqlFromGeneral(logEntry),
    user_id: userId,
    id: crypto.randomUUID(),
  };
  try {
    await client!("user_foodlogentry").insert(insertEntry);
    return ok(insertEntry.id);
  } catch (error: any) {
    return err(new SystemError(error.message));
  }
};

export const retrieveFoodLog: SqliteRetrieveFoodLogFunction = async (
  userId: string,
  logId: string,
  client = knexInstance
): Promise<Result<FoodLogEntry, StorageError>> => {
  try {
    const log = await client<SqliteFoodLogEntry>("user_foodlogentry")
      .select()
      .where("user_id", userId)
      .andWhere("id", logId)
      .first();
    if (log === undefined) {
      return err(new NotFoundError("Log not found"));
    }
    return ok(generalFromSql(log));
  } catch (error: any) {
    return err(new SystemError(error.message));
  }
};

export const queryFoodLogs: SqliteQueryFoodLogFunction = async (
  userId: string,
  startDate: Date,
  endDate: Date,
  client = knexInstance
): Promise<Result<FoodLogEntry[], StorageError>> => {
  if (endDate.getTime() < startDate.getTime()) {
    return Promise.resolve(
      err(new ValidationError("startDate is after endDate"))
    );
  }
  try {
    const logs = await client<SqliteFoodLogEntry>("user_foodlogentry")
      .select()
      .where("user_id", userId)
      .andWhere("time_start", "<=", endDate.getTime() / 1000)
      .andWhere("time_end", ">=", startDate.getTime() / 1000);
    return ok(logs.map(generalFromSql));
  } catch (error: any) {
    return err(new SystemError(error.message));
  }
};

export const editFoodLog: SqliteEditFoodLogFunction = async (
  userId: string,
  logEntry: EditFoodLogEntry,
  client = knexInstance
) => {
  if (!isValidEditLogEntry(logEntry)) {
    return Promise.resolve(err(new ValidationError("Invalid Log Entry")));
  }
  const logid = logEntry.id;
  const updates: { [key: string]: any } = {};
  if (logEntry.name) {
    updates["name"] = logEntry.name;
  }
  if (logEntry.time) {
    updates["time_start"] = new Date(logEntry.time.start).getTime() / 1000;
    updates["time_end"] = new Date(logEntry.time.end).getTime() / 1000;
  }
  if (logEntry.metrics) {
    updates["metrics"] = JSON.stringify(logEntry.metrics);
  }
  if (logEntry.labels) {
    updates["labels"] = JSON.stringify(logEntry.labels);
  }
  try {
    await client("user_foodlogentry")
      .update(updates)
      .where("user_id", userId)
      .andWhere("id", logid);
    return await retrieveFoodLog(userId, logEntry.id, client);
  } catch (error: any) {
    return err(new SystemError(error.message));
  }
};

export const deleteFoodLog: SqliteDeleteFoodLogFunction = async (
  userId: string,
  logId: string,
  client = knexInstance
): Promise<Result<boolean, StorageError>> => {
  try {
    await client<SqliteFoodLogEntry>("user_foodlogentry")
      .delete()
      .where("user_id", userId)
      .andWhere("id", logId);
    return ok(true);
  } catch (error: any) {
    return err(new SystemError(error.message));
  }
};

const TEMP_DIR = process.env.OPENFOODDIARY_TEMP_DIRECTORY ?? "/tmp";

function bulkFromSql(sql: SqliteFoodLogEntry): BulkExportFoodLogEntry {
  return {
    id: sql.id,
    name: sql.name,
    labels: JSON.parse(sql.labels),
    timeStart: new Date(sql.time_start * 1000).toISOString(),
    timeEnd: new Date(sql.time_end * 1000).toISOString(),
    metrics: JSON.parse(sql.metrics),
  };
}

export const bulkExportFoodLogs: SqliteBulkExportFoodLogsFunction = async (
  userId: string,
  client = knexInstance
): Promise<Result<string, StorageError>> => {
  try {
    const filename = `${TEMP_DIR}/${crypto.randomUUID()}.csv`;
    const logs = await client<SqliteFoodLogEntry>("user_foodlogentry")
      .select()
      .where("user_id", userId);
    const exportedLogs = logs.map(bulkFromSql);
    fs.writeFileSync(
      filename,
      stringify([["id", "name", "labels", "timeStart", "timeEnd", "metrics"]]),
      {
        flag: "w",
      }
    );
    for (const log of exportedLogs) {
      fs.appendFileSync(
        filename,
        stringify([
          [
            log.id,
            log.name,
            log.labels,
            log.timeStart,
            log.timeEnd,
            log.metrics,
          ],
        ])
      );
    }
    return ok(filename);
  } catch (error: any) {
    console.error(error.message);
    return err(new SystemError(error.message));
  }
};
