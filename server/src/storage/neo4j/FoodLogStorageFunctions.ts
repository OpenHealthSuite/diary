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

export const foodLog: FoodLogStorage = {
  storeFoodLog: async function (
    userId: string,
    logEntry: CreateFoodLogEntry
  ): Promise<Result<string, StorageError>> {
    if (!isValidCreateLogEntry(logEntry)) {
      return Promise.resolve(err(new ValidationError("Invalid Log Entry")));
    }
    const session = NEO4J_DRIVER.session();
    try {
      const labelMap = logEntry.labels.reduce((prev, curr, i) => {
        return { ...prev, [`fll${i}`]: curr };
      }, {});
      const id = randomUUID();
      await session.run(
        `MERGE (u:User {userid: $userid})
          MERGE (fl:Foodlog { 
            id: $id,
            name: $name,
            metrics: $metrics
          })
          ${logEntry.labels
            .map((_, i) => `MERGE (fll${i}: FoodLogLabel { value: $fll${i} })`)
            .join("\n")}
          MERGE (st: Time { logTime: datetime($startTime) })
          MERGE (et: Time { logTime: datetime($endTime) })
          MERGE (u)-[:ENTERED_LOG]->(fl)
          MERGE (fl)<-[:START_TIME]-(st)
          MERGE (fl)<-[:END_TIME]-(et)
          ${logEntry.labels
            .map((_, i) => `MERGE (fl)<-[:HAS_LABEL]-(fll${i})\n`)
            .join("\n")}
          RETURN fl.id as id`,
        {
          userid: userId,
          name: logEntry.name,
          id,
          metrics: JSON.stringify(logEntry.metrics),
          startTime: logEntry.time.start.toISOString(),
          endTime: logEntry.time.end.toISOString(),
          ...labelMap,
        }
      );

      await session.close();
      return ok(id);
    } catch (error: any) {
      await session.close();
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
        `MATCH (:User {userid: $userid})-[r:ENTERED_LOG]-(fl:FoodLog {id: $logid})
         MATCH (st: Time)-[:START_TIME]-(fl)
         MATCH (et: Time)-[:END_TIME]-(fl)
         MATCH (fl)-[:HAS_LABEL]-(fll)
          RETURN fl.id as id, 
            fl.name as name, 
            fl.metrics as metrics, 
            collect(fll.value) as labels,
            st.logTime as startTime,
            et.logTime as endTime`,
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
            metrics: JSON.parse(rec.get("metrics")),
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
