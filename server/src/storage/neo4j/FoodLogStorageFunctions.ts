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
import { FoodLogEntry } from "../../types";
import {
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
      const labels = logEntry.labels.reduce((acc, lbl, i) => {
        return { ...acc, [`fll${i}`]: lbl };
      }, {});
      const id = randomUUID();
      await session.run(
        `MERGE (u:User {userid: $userid})
          MERGE (fl:FoodLog { 
            id: $id, 
            name: $name,
            metrics: $metrics,
            startTime: datetime($startDate),
            endTime: datetime($endDate)
          })
          MERGE (u)-[:ENTERED]->(fl)
          ${Object.keys(labels).map(fll => {
            return `MERGE (${fll}:FoodLogLabel { value: $${fll} })
            MERGE (fl)<-[:LABELED_WITH]-(${fll})`;
          }).join("\n")}
          RETURN fl.id`,
        {
          userid: userId,
          id: id,
          name: logEntry.name,
          metrics: JSON.stringify(logEntry.metrics),
          startDate: logEntry.time.start.toISOString(),
          endDate: logEntry.time.end.toISOString(),
          ...labels
        }
      );
      

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
        `MATCH (:User {userid: $userid})-[r:ENTERED]-(fl:FoodLog {id:$logid})
         OPTIONAL MATCH (fl)-[:LABELED_WITH]-(fll: FoodLogLabel)
          RETURN fl.id as id, 
            fl.name as name, 
            collect(fll.value) as labels, 
            fl.metrics as metrics,
            apoc.temporal.format(fl.startTime, "ISO_DATE_TIME") as startTime,
            apoc.temporal.format(fl.endTime, "ISO_DATE_TIME") as endTime`,
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
              start: new Date(rec.get("startTime")),
              end: new Date(rec.get("endTime")),
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
  editFoodLog: async function (
    userId: string,
    logEntry: EditFoodLogEntry
  ): Promise<Result<FoodLogEntry, StorageError>> {
    if (!isValidEditLogEntry(logEntry)) {
      return Promise.resolve(err(new ValidationError("Invalid Log Entry")));
    }
    const session = NEO4J_DRIVER.session();
    try {
      const labels = (logEntry.labels ?? []).reduce((acc, lbl, i) => {
        return { ...acc, [`fll${i}`]: lbl };
      }, {});
      let props: any = {
        userid: userId,
        id: logEntry.id,
        ...labels
      }
      if (logEntry.name) {
        props = {
          ...props,
          name: logEntry.name,
        }
      }
      if (logEntry.metrics) {
        props = {
          ...props,
          metrics: JSON.stringify(logEntry.metrics),
        }

      }
      if (logEntry.time) {
        props = {
          ...props,
          startDate: logEntry.time!.start.toISOString(),
          endDate: logEntry.time!.end.toISOString(),
        }
      }
      // probably a smarter way of doing this, but lets just make it work
      await session.run(
        `MERGE (u:User {userid: $userid})
          MERGE (fl:FoodLog { 
            id: $id
          })
          ON MATCH SET 
            ${logEntry.name ? "fl.name = $name," : ""}
            ${logEntry.metrics ? "fl.metrics = $metrics," : ""}
            ${logEntry.time ? "fl.startTime = datetime($startDate)," : ""}
            ${logEntry.time ? "fl.endTime = datetime($endDate)" : ""}
          MERGE (u)-[:ENTERED]->(fl)
          ${Object.keys(labels).map(fll => {
            return `MERGE (${fll}:FoodLogLabel { value: $${fll} })
            MERGE (fl)<-[:LABELED_WITH]-(${fll})`;
          }).join("\n")}
          RETURN fl.id`,
          props
      );
      
      return await this.retrieveFoodLog(userId, logEntry.id);
    } catch (error: any) {
      console.error(error);
      return err(new SystemError(error.message));
    }
  },
  deleteFoodLog: async function (
    userId: string,
    logId: string
  ): Promise<Result<boolean, StorageError>> {
    const session = NEO4J_DRIVER.session();
    try {
      const res = await session.run<any>(
        `MATCH (:User {userid: $userid})-[r:ENTERED]-(fl:FoodLog {id:$logid})
        DETACH DELETE fl`,
        {
          userid: userId,
          logid: logId,
        }
      );
      await session.close();

      return ok(true
      );
    } catch (error: any) {
      await session.close();
      console.error(error);
      return err(new SystemError(error.message));
    }
  },
  queryFoodLogs: async function (
    userId: string,
    dateStart: Date,
    dateEnd: Date
  ): Promise<Result<FoodLogEntry[], StorageError>> {
    if (dateEnd.getTime() < dateStart.getTime()) {
      return Promise.resolve(
        err(new ValidationError("startDate is after endDate"))
      );
    }
    const session = NEO4J_DRIVER.session();
    try {
      const res = await session.run<any>(
        `MATCH (:User {userid: $userid})-[r:ENTERED]-(fl:FoodLog)
          WHERE fl.startTime <= datetime($endTime) AND
          fl.endTime >= datetime($startTime)
         OPTIONAL MATCH (fl)-[:LABELED_WITH]-(fll: FoodLogLabel)
          RETURN fl.id as id, 
            fl.name as name, 
            collect(fll.value) as labels, 
            fl.metrics as metrics,
            apoc.temporal.format(fl.startTime, "ISO_DATE_TIME") as startTime,
            apoc.temporal.format(fl.endTime, "ISO_DATE_TIME") as endTime`,
        {
          userid: userId,
          startTime: dateStart.toISOString(),
          endTime: dateEnd.toISOString()
        }
      );
      await session.close();

      return ok(
        res.records.map((rec) => {
          return {
            id: rec.get("id"),
            name: rec.get("name"),
            labels: rec.get("labels"),
            time: {
              start: new Date(rec.get("startTime")),
              end: new Date(rec.get("endTime")),
            },
            metrics: JSON.parse(rec.get("metrics")),
          };
        })
      );
    } catch (error: any) {
      await session.close();
      console.error(error);
      return err(new SystemError(error.message));
    }
  },
  bulkExportFoodLogs: function (
    userId: string
  ): Promise<Result<string, StorageError>> {
    // lets not touch this...
    throw new Error("Function not implemented.");
  },
};
