import express, { NextFunction, Request, Response, Router } from 'express';
import { isNotFoundError, isValidationError, StorageError } from '../storage';
import { OFDLocals } from '../middlewares';
import { DeleteFoodLogFunction, EditFoodLogFunction, QueryFoodLogFunction, RetrieveFoodLogFunction, StoreFoodLogFunction } from '../storage/types/FoodLog';

import * as storage from '../storage'

const foodStorageProvider = storage.memory.foodLog;

export function buildRouter(router: Router): Router {
  return router.post('/logs', createFoodLogHandler)
    .get('/logs', queryFoodLogHandler)
    .get('/logs/:logId', getFoodLogHandler)
    .put('/logs/:logId', updateFoodLogHandler)
    .delete('/logs/:logId', deleteFoodLogHandler)
}

export const FoodStorageRouter = buildRouter(express.Router())

function errorStatusCodeCalculator(err: StorageError): number {
  if (isValidationError(err))
    return 400
  if (isNotFoundError(err))
    return 404
  return 500
}

export function createFoodLogHandler(
  req: Request,
  res: Response & { locals: OFDLocals },
  next: NextFunction,
  storeFoodLog: StoreFoodLogFunction = foodStorageProvider.storeFoodLog
) {
  storeFoodLog(res.locals.userId, req.body)
    .then(result => result.map(res.send)
      .mapErr(err => res.status(errorStatusCodeCalculator(err)).send(err.message))
    )
}

function validStartEndDateStrings(startDateString: string, endDateString: string): boolean {
  const startDate = new Date(startDateString);
  const endDate = new Date(endDateString);
  return startDate.toString() !== "Invalid Date" && endDate.toString() !== "Invalid Date" && endDate.getTime() >= startDate.getTime()
}

export function queryFoodLogHandler(
  req: Request,
  res: Response & { locals: OFDLocals },
  next: NextFunction,
  queryFoodLog: QueryFoodLogFunction = {} as any
) {
  const { startDate, endDate } = req.query;
  if (typeof startDate === "string" && typeof endDate === "string" && validStartEndDateStrings(startDate, endDate)) {
    queryFoodLog(res.locals.userId, new Date(startDate), new Date(endDate))
      .then(result => result.map(res.send)
        .mapErr(err => res.status(errorStatusCodeCalculator(err)).send(err.message)))
  }
  res.status(400).send("Invalid startDate or endDate")
}

export function getFoodLogHandler(
  req: Request,
  res: Response & { locals: OFDLocals },
  next: NextFunction,
  getFoodLog: RetrieveFoodLogFunction = foodStorageProvider.retrieveFoodLog
) {
  getFoodLog(res.locals.userId, req.params.itemId)
    .then(result => result.map(res.send)
      .mapErr(err => res.status(errorStatusCodeCalculator(err)).send(err.message))
    )
  
}

export function updateFoodLogHandler(
  req: Request,
  res: Response & { locals: OFDLocals },
  next: NextFunction,
  editFoodLog: EditFoodLogFunction = foodStorageProvider.editFoodLog
) {
  editFoodLog(res.locals.userId, { id: req.params.itemId, ...req.body })
    .then(result => result.map(res.send)
      .mapErr(err => res.status(errorStatusCodeCalculator(err)).send(err.message)))
}


export function deleteFoodLogHandler(
  req: Request,
  res: Response & { locals: OFDLocals },
  next: NextFunction,
  deleteFoodLog: DeleteFoodLogFunction = foodStorageProvider.deleteFoodLog
) {
  deleteFoodLog(res.locals.userId, req.params.itemId)
    .then(result => result.map(res.status(204).send)
      .mapErr(err => res.status(errorStatusCodeCalculator(err)).send(err.message)))
}
