import { Client } from "cassandra-driver";
import { Result } from "neverthrow";
import { FoodLogEntry } from "../../types";
import { StorageError } from "./StorageErrors";

export type CreateFoodLogEntry = Pick<FoodLogEntry, "name" | "labels" | "time" | "metrics" >

export type EditFoodLogEntry = Pick<FoodLogEntry, "id"> & Partial<Pick<FoodLogEntry, "name" | "labels" | "time" | "metrics" >>

export type StoreFoodLogFunction = (userId: string, logEntry: CreateFoodLogEntry, client?: Client) => Promise<Result<string, StorageError>>

export type QueryFoodLogFunction = (userId: string, dateStart: Date, dateEnd: Date, client?: Client) => Promise<Result<FoodLogEntry[], StorageError>>

export type RetrieveFoodLogFunction = (userId: string, logId: string, client?: Client) => Promise<Result<FoodLogEntry, StorageError>>

export type EditFoodLogFunction = (userId: string, logEntry: EditFoodLogEntry, client?: Client) => Promise<Result<FoodLogEntry, StorageError>>

export type DeleteFoodLogFunction = (userId: string, logId: string, client?: Client) => Promise<Result<boolean, StorageError>>
