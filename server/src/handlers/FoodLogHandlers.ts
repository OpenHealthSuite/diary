import { Express, NextFunction, Request, Response } from 'express';
import { isValidationError } from '../storage';
import { OFDLocals } from '../middlewares';
import { StoreFoodLogFunction } from '../storage/types/FoodLog';

export function addHandlers(app: Express) {
  app.post('/logs', createFoodLogHandler)
}

export function createFoodLogHandler(
  req: Request,
  res: Response & { locals: OFDLocals },
  next: NextFunction,
  storeFoodLog: StoreFoodLogFunction = {} as any
) {
  storeFoodLog(res.locals.userId, req.body)
    .then(result => result.map(res.send)
      .mapErr(err => res.status(isValidationError(err) ? 400 : 500).send(err.message))
    )
}