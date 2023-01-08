import * as neo4j from "neo4j-driver";
import { StorageType } from "../interfaces";
import { foodLog } from "./FoodLogStorageFunctions";
import { configuration } from "./ConfigurationStorageFunctions";

export const NEO4J_DRIVER = neo4j.driver(
  "neo4j://localhost:7687",
  neo4j.auth.basic("neo4j", "s3cr3tly")
);

export const storageConfig: StorageType = {
  setupDatabase: async () => {},
  shutdownDatabase: async () => {
    await NEO4J_DRIVER.close();
  },
  foodLog,
  configuration,
};
