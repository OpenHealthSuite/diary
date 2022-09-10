import { err, ok, Result } from "neverthrow"
import { FoodLogEntry } from "../../types";
import crypto from 'node:crypto'
import { 
    CreateFoodLogEntry,
    DeleteFoodLogFunction,
    EditFoodLogEntry,
    EditFoodLogFunction,
    NotFoundError,
    QueryFoodLogFunction,
    RetrieveFoodLogFunction,
    StorageError,
    StoreFoodLogFunction,
    SystemError,
    ValidationError
} from "../types"

import { CASSANDRA_CLIENT } from '.'
import { Client } from "cassandra-driver";

type CassandraStoreFoodLogFunction = (userId: string, logEntry: CreateFoodLogEntry, client?: Client) => Promise<Result<string, StorageError>>;
type CassandraQueryFoodLogFunction = (userId: string, dateStart: Date, dateEnd: Date, client?: Client) => Promise<Result<FoodLogEntry[], StorageError>>;
type CassandraRetrieveFoodLogFunction = (userId: string, logId: string, client?: Client) => Promise<Result<FoodLogEntry, StorageError>>;
type CassandraEditFoodLogFunction = (userId: string, logEntry: EditFoodLogEntry, client?: Client) => Promise<Result<FoodLogEntry, StorageError>>;
type CassandraDeleteFoodLogFunction = (userId: string, logId: string, client?: Client) => Promise<Result<boolean, StorageError>>;

function isValidCreateLogEntry(logEntry: CreateFoodLogEntry): boolean {
    return (logEntry as any).id === undefined 
        && logEntry.name !== undefined
        && logEntry.labels !== undefined //Check the labels are a set?
        && logEntry.metrics !== undefined
        && !Object.values(logEntry.metrics).some(isNaN)
        && logEntry.time !== undefined
        && logEntry.time.start !== undefined
        && logEntry.time.end !== undefined
        && logEntry.time.end.getTime() >= logEntry.time.start.getTime()
}

function isValidEditLogEntry(logEntry: EditFoodLogEntry): boolean {
    return logEntry.id !== undefined 
        && (logEntry.metrics === undefined || !Object.values(logEntry.metrics).some(isNaN))
        && (logEntry.time === undefined || (
        logEntry.time.start !== undefined
        && logEntry.time.end !== undefined
        && logEntry.time.end.getTime() >= logEntry.time.start.getTime() ))
}

export const storeFoodLog: CassandraStoreFoodLogFunction = 
    async (userId: string, logEntry: CreateFoodLogEntry, cassandraClient = CASSANDRA_CLIENT) : Promise<Result<string, StorageError>> => {
        if (!isValidCreateLogEntry(logEntry)) {
            return Promise.resolve(err(new ValidationError("Invalid Log Entry")))
        }
        const insertEntry: FoodLogEntry & { userId: string } = {
            ...logEntry,
            id: crypto.randomUUID(),
            userId
        }
        try {
            await cassandraClient.execute(`INSERT INTO openfooddiary.user_foodlogentry (userId, id, name, labels, time, metrics, timeStart, timeEnd)
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
                ], { prepare: true });
            return ok(insertEntry.id);
        } catch (error: any) {
            return err(new SystemError(error.message))
        }
    }

export const retrieveFoodLog: CassandraRetrieveFoodLogFunction =
    async (userId: string, logId: string, cassandraClient = CASSANDRA_CLIENT) : Promise<Result<FoodLogEntry, StorageError>> => {
        try {
            const result = await cassandraClient.execute(`SELECT CAST(id as text) as id, name, labels, time, metrics 
            FROM  openfooddiary.user_foodlogentry 
            WHERE userId = ? AND id = ?;`, [userId, logId], { prepare: true });
            if (result.rows.length == 0) {
                return err(new NotFoundError("No Log with Id"))
            }
            const item = result.first();
            const constructed: any = {};
            item.keys().forEach(key => constructed[key] = item.get(key))
            return ok(constructed as FoodLogEntry);
        } catch (error: any) {
            return err(new SystemError(error.message))
        }
    }


export const queryFoodLogs: CassandraQueryFoodLogFunction =
    async (userId: string, startDate: Date, endDate: Date, cassandraClient = CASSANDRA_CLIENT) : Promise<Result<FoodLogEntry[], StorageError>> => {
        if (endDate.getTime() < startDate.getTime()) {
            return Promise.resolve(err(new ValidationError("startDate is after endDate")))
        }
        try {
            const result = await cassandraClient.execute(`SELECT CAST(id as text) as id, name, labels, time, metrics 
            FROM openfooddiary.user_foodlogentry 
            WHERE userId = ? AND timeStart <= ? AND timeEnd >= ?
            ALLOW FILTERING;`,
            [userId, endDate, startDate], { prepare: true });
            const constructed: any[] = result.rows.map(row => {
                const subCon: any = {};
                row.keys().forEach(key => subCon[key] = row.get(key));
                return subCon;
            })
            return ok(constructed as FoodLogEntry[]);
        } catch (error: any) {
            return err(new SystemError(error.message))
        }
    }

const UPDATEABLE_FIELDS = ["name", "labels", "time", "metrics"]
export const editFoodLog: CassandraEditFoodLogFunction =
    async (userId: string, logEntry: EditFoodLogEntry, cassandraClient = CASSANDRA_CLIENT) => {
        if (!isValidEditLogEntry(logEntry)) {
            return Promise.resolve(err(new ValidationError("Invalid Log Entry")))
        }
        const updateEntity: any = {
            ...logEntry,
            userId
        }

        let updatedFields: string[] = []
        let updateValues: any[] = []

        UPDATEABLE_FIELDS.forEach(field => {
            if (updateEntity[field] !== undefined) {
                updatedFields.push(field + ' = ?');
                updateValues.push(updateEntity[field])
                if (field === 'time') {
                    updatedFields.push('timeStart = ?');
                    updatedFields.push('timeEnd = ?');
                    updateValues.push(updateEntity['time'].start)
                    updateValues.push(updateEntity['time'].end)
                }
            }
        })
        try {
            await cassandraClient.execute(`UPDATE openfooddiary.user_foodlogentry
            SET ${updatedFields.join(',')}
            WHERE userId = ? AND id = ?;`, 
                [
                    ...updateValues,
                    userId,
                    updateEntity.id,
                ], { prepare: true });
            return await retrieveFoodLog(userId, logEntry.id, cassandraClient);
        } catch (error: any) {
            return err(new SystemError(error.message))
        }
    }

export const deleteFoodLog: CassandraDeleteFoodLogFunction =
    async (userId: string, logId: string, cassandraClient = CASSANDRA_CLIENT): Promise<Result<boolean, StorageError>>  => {
        try {
            await cassandraClient.execute(`DELETE FROM openfooddiary.user_foodlogentry
            WHERE userId = ? AND id = ?;`, 
                [
                    userId,
                    logId,
                ], { prepare: true });
            return ok(true);
        } catch (error: any) {
            return err(new SystemError(error.message))
        }
    }