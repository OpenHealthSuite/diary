import express, { NextFunction, Request, Response, Router } from "express";
import {
  isNotFoundError,
  isValidationError,
  StorageError,
  STORAGE,
} from "../storage";
import { OFDLocals } from "../middlewares";
import {
  DeleteFoodLogFunction,
  EditFoodLogFunction,
  QueryFoodLogFunction,
  RetrieveFoodLogFunction,
  StoreFoodLogFunction,
} from "../storage/types/FoodLog";

let foodStorageProvider = STORAGE.foodLog;

export function buildRouter(router: Router): Router {
  return router
    .post("/logs", createFoodLogHandler)
    .get("/logs", queryFoodLogHandler)
    .get("/logs/:itemId", getFoodLogHandler)
    .put("/logs/:itemId", updateFoodLogHandler)
    .delete("/logs/:itemId", deleteFoodLogHandler);
}

export const FoodStorageRouter = buildRouter(express.Router());

function errorStatusCodeCalculator(err: StorageError): number {
  if (isValidationError(err)) return 400;
  if (isNotFoundError(err)) return 404;
  return 500;
}

export function createFoodLogHandler(
  req: Request,
  res: Response & { locals: OFDLocals },
  next: NextFunction,
  storeFoodLog: StoreFoodLogFunction = foodStorageProvider.storeFoodLog
) {
  let item = structuredClone(req.body);
  if (
    item &&
    item.time &&
    item.time.start &&
    item.time.end &&
    validStartEndDateStrings(item.time.start, item.time.end)
  ) {
    item.time.start = new Date(item.time.start);
    item.time.end = new Date(item.time.end);
  }
  storeFoodLog(res.locals.userId, item).then((result) =>
    result
      .map((val) => res.send(val))
      .mapErr((err) =>
        res.status(errorStatusCodeCalculator(err)).send(err.message)
      )
  );
}

function validStartEndDateStrings(
  startDateString: string,
  endDateString: string
): boolean {
  const startDate = new Date(startDateString);
  const endDate = new Date(endDateString);
  return (
    startDate.toString() !== "Invalid Date" &&
    endDate.toString() !== "Invalid Date"
  );
}

export function queryFoodLogHandler(
  req: Request,
  res: Response & { locals: OFDLocals },
  next: NextFunction,
  queryFoodLog: QueryFoodLogFunction = foodStorageProvider.queryFoodLogs
) {
  const { startDate, endDate } = req.query;
  if (
    typeof startDate === "string" &&
    typeof endDate === "string" &&
    validStartEndDateStrings(startDate, endDate)
  ) {
    queryFoodLog(
      res.locals.userId,
      new Date(startDate),
      new Date(endDate)
    ).then((result) =>
      result
        .map((val) => res.send(val))
        .mapErr((err) =>
          res.status(errorStatusCodeCalculator(err)).send(err.message)
        )
    );
    return;
  }
  res.status(400).send("Invalid startDate or endDate");
}

export function getFoodLogHandler(
  req: Request,
  res: Response & { locals: OFDLocals },
  next: NextFunction,
  getFoodLog: RetrieveFoodLogFunction = foodStorageProvider.retrieveFoodLog
) {
  getFoodLog(res.locals.userId, req.params.itemId).then((result) =>
    result
      .map((val) => res.send(val))
      .mapErr((err) =>
        res.status(errorStatusCodeCalculator(err)).send(err.message)
      )
  );
}

export function updateFoodLogHandler(
  req: Request,
  res: Response & { locals: OFDLocals },
  next: NextFunction,
  editFoodLog: EditFoodLogFunction = foodStorageProvider.editFoodLog
) {
  let item = structuredClone(req.body);
  if (
    item.time &&
    item.time.start &&
    item.time.end &&
    validStartEndDateStrings(item.time.start, item.time.end)
  ) {
    item.time.start = new Date(item.time.start);
    item.time.end = new Date(item.time.end);
  } else if (item.time) {
    res.status(400).send("Unable to parse dates");
    return;
  }
  editFoodLog(res.locals.userId, { id: req.params.itemId, ...item }).then(
    (result) =>
      result
        .map((val) => res.send(val))
        .mapErr((err) =>
          res.status(errorStatusCodeCalculator(err)).send(err.message)
        )
  );
}

export function deleteFoodLogHandler(
  req: Request,
  res: Response & { locals: OFDLocals },
  next: NextFunction,
  deleteFoodLog: DeleteFoodLogFunction = foodStorageProvider.deleteFoodLog
) {
  deleteFoodLog(res.locals.userId, req.params.itemId).then((result) =>
    result
      .map(() => res.status(204).send())
      .mapErr((err) =>
        res.status(errorStatusCodeCalculator(err)).send(err.message)
      )
  );
}
