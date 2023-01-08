import * as neo4j from "neo4j-driver";
import { StorageType } from "../interfaces";
import {
  CreateFoodLogEntry,
  EditFoodLogEntry,
  FoodLogStorage,
  NotFoundError,
  StorageError,
  SystemError,
  ValidationError,
} from "../types";
import { Result, err, ok } from "neverthrow";
import { Configuration, FoodLogEntry } from "../../types";
import { ConfigurationStorage } from "../types/Configuration";
import {
  isValidConfigurationItem,
  isValidCreateLogEntry,
  isValidEditLogEntry,
} from "../validation";
import { randomUUID } from "node:crypto";
import { NEO4J_DRIVER } from ".";
import { OGM } from "@neo4j/graphql-ogm";

const typeDefs = `
  type User {
    userid: String!
    foodLogs: [FoodLog!]! @relationship(type: "ENTERED", direction: OUT)
  }

  type FoodLog {
    id: String!
    name: String!
    metrics: [FoodLogMetric!]! @relationship(type: "LOG_METRIC", direction: IN)
    labels: [String!]!
    time: StartAndEndTime! @relationship(type: "LOG_TIME", direction: IN)
  }

  type FoodLogMetric {
    key: String!
    value: Int!
  }

  type StartAndEndTime {
    startTime: DateTime!
    endTime: DateTime!
  }
`;

const ogm = new OGM({
  typeDefs,
  driver: NEO4J_DRIVER,
});

const User = ogm.model("User");

export const foodLog: FoodLogStorage = {
  storeFoodLog: async function (
    userId: string,
    logEntry: CreateFoodLogEntry
  ): Promise<Result<string, StorageError>> {
    if (!isValidCreateLogEntry(logEntry)) {
      return Promise.resolve(err(new ValidationError("Invalid Log Entry")));
    }
    try {
      const metrics = Object.entries(logEntry.metrics).map(([key, value]) => {
        return { key, value };
      });
      const id = randomUUID();

      await ogm.init().then(() => {
        User.create({
          input: [
            {
              userid: userId,
              foodLogs: {
                create: [
                  {
                    node: {
                      id: id,
                      name: logEntry.name,
                      labels: logEntry.labels,
                      metrics: {
                        create: metrics.map((val) => {
                          return {
                            node: val,
                          };
                        }),
                      },
                      time: {
                        create: [
                          {
                            node: {
                              startTime: logEntry.time.start,
                              endTime: logEntry.time.end,
                            },
                          },
                        ],
                      },
                    },
                  },
                ],
              },
            },
          ],
        });
      });

      return ok(id);
    } catch (error: any) {
      console.error(error);
      return err(new SystemError(error.message));
    }
  },
  retrieveFoodLog: async function (
    userId: string,
    logId: string
  ): Promise<Result<FoodLogEntry, StorageError>> {
    const session = NEO4J_DRIVER.session();
    try {
      const res = await session.run<any>(
        `MATCH (:User {userid: $userid})-[r:ENTERED]-(fl:FoodLog {id: $logid})
         MATCH (t: StartAndEndTime)-[:LOG_TIME]-(fl)
         MATCH (m: FoodLogMetric)-[:LOG_METRIC]-(fl)
          RETURN fl.id as id, 
            fl.name as name, 
            collect(fl.metrics) as raw_metrics, 
            fl.labels as labels,
            st.startTime as startTime,
            et.endTime as endTime`,
        {
          userid: userId,
          logid: logId,
        }
      );
      await session.close();
      if (res.records.length == 0) {
        return err(new NotFoundError("Not Found"));
      }

      return ok(
        res.records.map((rec) => {
          return {
            id: rec.get("id"),
            name: rec.get("name"),
            labels: rec.get("labels"),
            time: {
              start: rec.get("startTime"),
              end: rec.get("endTime"),
            },
            metrics: rec
              .get("metrics")
              .reduce((acc: any, curr: { key: string; value: number }) => {
                return { ...acc, [curr.key]: curr.value };
              }, {}),
          };
        })[0]
      );
    } catch (error: any) {
      await session.close();
      console.error(error);
      return err(new SystemError(error.message));
    }
  },
  editFoodLog: function (
    userId: string,
    logEntry: EditFoodLogEntry
  ): Promise<Result<FoodLogEntry, StorageError>> {
    if (!isValidEditLogEntry(logEntry)) {
      return Promise.resolve(err(new ValidationError("Invalid Log Entry")));
    }
    throw new Error("Function not implemented.");
  },
  deleteFoodLog: function (
    userId: string,
    logId: string
  ): Promise<Result<boolean, StorageError>> {
    throw new Error("Function not implemented.");
  },
  queryFoodLogs: function (
    userId: string,
    dateStart: Date,
    dateEnd: Date
  ): Promise<Result<FoodLogEntry[], StorageError>> {
    if (dateEnd.getTime() < dateStart.getTime()) {
      return Promise.resolve(
        err(new ValidationError("startDate is after endDate"))
      );
    }
    throw new Error("Function not implemented.");
  },
  bulkExportFoodLogs: function (
    userId: string
  ): Promise<Result<string, StorageError>> {
    // lets not touch this...
    throw new Error("Function not implemented.");
  },
};
