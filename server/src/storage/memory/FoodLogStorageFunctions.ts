import { err, ok, Result } from "neverthrow"
import { FoodLogEntry } from "../../types";
import crypto from 'node:crypto'
import { CreateFoodLogEntry, NotFoundError, RetrieveFoodLogFunction, StorageError, StoreFoodLogFunction, ValidationError } from "../types"

const MEMORY_STORAGE: { [key: string] : FoodLogEntry[] } = {};

function isValidCreateLogEntry(logEntry: CreateFoodLogEntry): boolean {
    return logEntry.name !== undefined;
}

export const storeFoodLog: StoreFoodLogFunction = 
    (userId: string, logEntry: CreateFoodLogEntry) : Promise<Result<string, StorageError>> => {
        if (!isValidCreateLogEntry(logEntry)) {
            return Promise.resolve(err(new ValidationError("Invalid Log Entry")))
        }
        const newItem = {
            id: crypto.randomUUID(),
            ...logEntry
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
        const log = MEMORY_STORAGE[userId]?.find(x => x.id === logId)
        if (log === undefined) {
            return Promise.resolve(err(new NotFoundError("No Log Found")));
        }
        return Promise.resolve(ok(log));
    }
