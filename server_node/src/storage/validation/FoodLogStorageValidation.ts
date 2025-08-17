import { METRIC_MAX } from "../../config";
import { CreateFoodLogEntry, EditFoodLogEntry } from "../types";

export function isValidCreateLogEntry (logEntry: CreateFoodLogEntry): boolean {
  return (
    (logEntry as any).id === undefined &&
    logEntry.name !== undefined &&
    logEntry.labels !== undefined && // Check the labels are a set?
    logEntry.metrics !== undefined &&
    logEntry.metrics !== null &&
    !Object.values(logEntry.metrics).some((m) => isNaN(m) || m > METRIC_MAX) &&
    logEntry.time !== undefined &&
    logEntry.time.start !== undefined &&
    logEntry.time.end !== undefined &&
    logEntry.time.end.getTime() >= logEntry.time.start.getTime()
  );
}

export function isValidEditLogEntry (logEntry: EditFoodLogEntry): boolean {
  return (
    logEntry.id !== undefined &&
    (logEntry.metrics === undefined ||
      !Object.values(logEntry.metrics).some(
        (m) => isNaN(m) || m > METRIC_MAX
      )) &&
    (logEntry.time === undefined ||
      (logEntry.time.start !== undefined &&
        logEntry.time.end !== undefined &&
        logEntry.time.end.getTime() >= logEntry.time.start.getTime()))
  );
}
