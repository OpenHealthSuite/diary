import { ok, Result } from "neverthrow"
import { FoodLogEntry } from "../../types";
import { CreateFoodLogEntry, StorageError, StoreFoodLogFunction } from "../types"

const MEMORY_STORAGE: { [key: string] : FoodLogEntry } = {};

export const createFoodLog: StoreFoodLogFunction = (userId: string, logEntry: CreateFoodLogEntry) : Promise<Result<string, StorageError>> => {
    return Promise.resolve(ok(""));
}