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
import { Client } from 'cassandra-driver'

// Probably worth finding a better way of doing this but we only have one dataset for now
const MIGRATIONS: string[] =[
    "CREATE KEYSPACE IF NOT EXISTS openfooddiary WITH REPLICATION = {'class':'SimpleStrategy','replication_factor':1};", // TODO: Make this optional?
    `CREATE TYPE IF NOT EXISTS openfooddiary.logTimes (start timestamp, end timestamp);`,
    `CREATE TABLE IF NOT EXISTS openfooddiary.user_foodlogentry (userId UUID, id UUID, name text, labels set<text>,time frozen<openfooddiary.logTimes>,metrics map<text,int>, timeStart timestamp, timeEnd timestamp, PRIMARY KEY ((userId), id));`,
    `CREATE INDEX IF NOT EXISTS ON openfooddiary.user_foodlogentry (timeStart);`,
    `CREATE INDEX IF NOT EXISTS ON openfooddiary.user_foodlogentry (timeEnd);`,
]

const CASSANDRA_CLIENT = new Client({
  contactPoints: process.env.OPENFOODDIARY_CASSANDRA_CONTACT_POINTS ? 
    process.env.OPENFOODDIARY_CASSANDRA_CONTACT_POINTS.split(';') : ['localhost:9042'],
  localDataCenter: process.env.OPENFOODDIARY_CASSANDRA_LOCALDATACENTER ?? 'datacenter1',
  credentials: {
    username: process.env.OPENFOODDIARY_CASSANDRA_USER ?? 'cassandra',
    password: process.env.OPENFOODDIARY_CASSANDRA_PASSWORD ?? 'cassandra'
  }
})

CASSANDRA_CLIENT.connect().then(async () => {
    for(let migration of MIGRATIONS) {
        await CASSANDRA_CLIENT.execute(migration)
    }
})

export async function closeCassandra() {
    await CASSANDRA_CLIENT.shutdown();
}

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


export const storeFoodLog: StoreFoodLogFunction = 
    async (userId: string, logEntry: CreateFoodLogEntry) : Promise<Result<string, StorageError>> => {
        if (!isValidCreateLogEntry(logEntry)) {
            return Promise.resolve(err(new ValidationError("Invalid Log Entry")))
        }
        const insertEntry: FoodLogEntry & { userId: string } = {
            ...logEntry,
            id: crypto.randomUUID(),
            userId
        }
        try {
            await CASSANDRA_CLIENT.execute(`INSERT INTO openfooddiary.user_foodlogentry (userId, id, name, labels, time, metrics, timeStart, timeEnd)
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

export const retrieveFoodLog: RetrieveFoodLogFunction =
    async (userId: string, logId: string) : Promise<Result<FoodLogEntry, StorageError>> => {
        try {
            const result = await CASSANDRA_CLIENT.execute(`SELECT CAST(id as text) as id, name, labels, time, metrics 
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


export const queryFoodLogs: QueryFoodLogFunction =
    async (userId: string, startDate: Date, endDate: Date) : Promise<Result<FoodLogEntry[], StorageError>> => {
        if (endDate.getTime() < startDate.getTime()) {
            return Promise.resolve(err(new ValidationError("startDate is after endDate")))
        }
        try {
            const result = await CASSANDRA_CLIENT.execute(`SELECT CAST(id as text) as id, name, labels, time, metrics 
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
export const editFoodLog: EditFoodLogFunction =
    async (userId: string, logEntry: EditFoodLogEntry) => {
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
            await CASSANDRA_CLIENT.execute(`UPDATE openfooddiary.user_foodlogentry
            SET ${updatedFields.join(',')}
            WHERE userId = ? AND id = ?;`, 
                [
                    ...updateValues,
                    userId,
                    updateEntity.id,
                ], { prepare: true });
            return await retrieveFoodLog(userId, logEntry.id);
        } catch (error: any) {
            return err(new SystemError(error.message))
        }
    }

export const deleteFoodLog: DeleteFoodLogFunction =
    async (userId: string, logId: string): Promise<Result<boolean, StorageError>>  => {
        try {
            await CASSANDRA_CLIENT.execute(`DELETE FROM openfooddiary.user_foodlogentry
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