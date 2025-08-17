import { Result } from "neverthrow";
import { Configuration } from "../../types";
import { StorageError } from "./StorageErrors";

export type UpsertConfiguration = Configuration;

export type StoreConfigurationFunction = (
  userId: string,
  configuration: UpsertConfiguration
) => Promise<Result<Configuration["id"], StorageError>>;

export type QueryUsersConfigurationFunction = (
  userId: string
) => Promise<Result<Configuration[], StorageError>>;

export type RetrieveUsersConfigurationFunction = (
  userId: string,
  configurationId: Configuration["id"]
) => Promise<Result<Configuration, StorageError>>;

export type DeleteUserConfigurationFunction = (
  userId: string,
  configurationId: Configuration["id"]
) => Promise<Result<boolean, StorageError>>;

export interface ConfigurationStorage {
  storeConfiguration: StoreConfigurationFunction;
  queryUserConfiguration: QueryUsersConfigurationFunction;
  retrieveUserConfiguration: RetrieveUsersConfigurationFunction;
  deleteUserConfiguration: DeleteUserConfigurationFunction;
}
