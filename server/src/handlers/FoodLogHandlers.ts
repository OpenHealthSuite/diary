import { Express, NextFunction, Request, Response } from 'express';
import { isNotFoundError, isValidationError, StorageError } from '../storage';
import { OFDLocals } from '../middlewares';
import { RetrieveFoodLogFunction, StoreFoodLogFunction } from '../storage/types/FoodLog';

export function addHandlers(app: Express) {
  app.post('/logs', createFoodLogHandler)
  app.get('/logs/:logId', getFoodLogHandler)
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
  storeFoodLog: StoreFoodLogFunction = {} as any
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
  getFoodLog: RetrieveFoodLogFunction = {} as any
) {
  getFoodLog(res.locals.userId, req.params.itemId)
    .then(result => result.map(res.send)
      .mapErr(err => res.status(errorStatusCodeCalculator(err)).send(err.message))
    )
  
}