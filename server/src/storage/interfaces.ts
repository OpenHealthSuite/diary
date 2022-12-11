import { FoodLogStorage } from "./types";
import { ConfigurationStorage } from "./types/Configuration";

export interface StorageType {
  setupDatabase: () => Promise<void>;
  shutdownDatabase: () => Promise<void>;
  foodLog: FoodLogStorage;
  configuration: ConfigurationStorage;
}
