import { Result } from "neverthrow";
import { FoodLogEntry } from "../../types";
import { StorageError } from "./StorageErrors";

export type CreateFoodLogEntry = Pick<FoodLogEntry, "name" | "labels" | "time" | "metrics" >

export type EditFoodLogEntry = Partial<Pick<FoodLogEntry, "id" | "name" | "labels" | "time" | "metrics" >>

export type StoreFoodLogFunction = (userId: string, logEntry: CreateFoodLogEntry) => Promise<Result<string, StorageError>>

export type QueryFoodLogFunction = (userId: string, dateStart: Date, dateEnd: Date) => Promise<Result<FoodLogEntry[], StorageError>>

export type RetrieveFoodLogFunction = (userId: string, logId: string) => Promise<Result<FoodLogEntry, StorageError>>

export type EditFoodLogFunction = (userId: string, logEntry: EditFoodLogEntry) => Promise<Result<FoodLogEntry, StorageError>>

export type DeleteFoodLogFunction = (userId: string, logId: string) => Promise<Result<boolean, StorageError>>
