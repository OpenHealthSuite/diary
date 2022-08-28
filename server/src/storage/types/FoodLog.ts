import { Result } from "neverthrow";
import { FoodLogEntry } from "../../types";
import { StorageError } from "./StorageErrors";

export type CreateFoodLogEntry = Pick<FoodLogEntry, "name" | "labels" | "time" | "metrics" >

export type EditFoodLogEntry = Pick<FoodLogEntry, "id" | "name" | "labels" | "time" | "metrics" >

export type StoreFoodLogFunction = (userId: string, logEntry: CreateFoodLogEntry) => Promise<Result<string, StorageError>> 
