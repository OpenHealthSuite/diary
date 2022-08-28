import { Express, NextFunction, Request, Response } from 'express';
import { isNotFoundError, isValidationError, StorageError } from '../storage';
import { OFDLocals } from '../middlewares';
import { DeleteFoodLogFunction, EditFoodLogFunction, RetrieveFoodLogFunction, StoreFoodLogFunction } from '../storage/types/FoodLog';

import * as storage from '../storage'

const foodStorageProvider = storage.memory.foodLog;

export function addHandlers(app: Express) {
  app.post('/logs', createFoodLogHandler)
  app.get('/logs/:logId', getFoodLogHandler)
  app.put('/logs/:logId', updateFoodLogHandler)
  app.delete('/logs/:logId', deleteFoodLogHandler)
}

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
