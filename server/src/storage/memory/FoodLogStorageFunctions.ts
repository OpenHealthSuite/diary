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
    ValidationError
} from "../types"

const MEMORY_STORAGE: { [key: string] : FoodLogEntry[] } = {};

// Making this TOO robust is probably a mistake - this will do for now.

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
    (userId: string, logEntry: CreateFoodLogEntry) : Promise<Result<string, StorageError>> => {
        if (!isValidCreateLogEntry(logEntry)) {
            return Promise.resolve(err(new ValidationError("Invalid Log Entry")))
        }
        const newItem = {
            ...logEntry,
            id: crypto.randomUUID(),
        }
        if (MEMORY_STORAGE[userId] !== undefined) {
            MEMORY_STORAGE[userId].push(newItem)
        } else {
            MEMORY_STORAGE[userId] = [newItem]
        }
        return Promise.resolve(ok(newItem.id));
    }

export const retrieveFoodLog: RetrieveFoodLogFunction =
    (userId: string, logId: string) : Promise<Result<FoodLogEntry, StorageError>> => {
        const log = MEMORY_STORAGE[userId]?.find(x => x && x.id === logId)
        if (log === undefined) {
            return Promise.resolve(err(new NotFoundError("No Log Found")));
        }
        return Promise.resolve(ok(log));
    }


export const queryFoodLogs: QueryFoodLogFunction =
    (userId: string, startDate: Date, endDate: Date) : Promise<Result<FoodLogEntry[], StorageError>> => {
        if (endDate.getTime() < startDate.getTime()) {
            return Promise.resolve(err(new ValidationError("startDate is after endDate")))
        }
        if (!MEMORY_STORAGE[userId]) {
            return Promise.resolve(ok([] as FoodLogEntry[]));
        }
        return Promise.resolve(ok(MEMORY_STORAGE[userId].filter(x => new Date(x.time.start).getTime() >= startDate.getTime() && new Date(x.time.end).getTime() <= endDate.getTime())));
    }

export const editFoodLog: EditFoodLogFunction =
    (userId: string, logEntry: EditFoodLogEntry) => {
        if (!isValidEditLogEntry(logEntry)) {
            return Promise.resolve(err(new ValidationError("Invalid Log Entry")))
        }
        const log = MEMORY_STORAGE[userId]?.findIndex(x => x && x.id === logEntry.id)
        if (log === undefined || log === -1) {
            return Promise.resolve(err(new NotFoundError("No Log Found")));
        }
        MEMORY_STORAGE[userId][log] = {
            ...MEMORY_STORAGE[userId][log],
            ...logEntry
        }
        return Promise.resolve(ok(MEMORY_STORAGE[userId][log]))
    }

export const deleteFoodLog: DeleteFoodLogFunction =
    (userId: string, logId: string): Promise<Result<boolean, StorageError>>  => {
        const log = MEMORY_STORAGE[userId]?.findIndex(x => x && x.id === logId)
        if (log === undefined || log === -1) {
            return Promise.resolve(ok(false));
        }
        MEMORY_STORAGE[userId].splice(log, 1);
        return Promise.resolve(ok(true))
    }