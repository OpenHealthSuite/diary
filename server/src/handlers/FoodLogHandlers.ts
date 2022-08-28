import { Express, NextFunction, Request, Response } from 'express';
import { Result } from 'neverthrow';
import { isValidationError, StorageError } from '../storage';
import { CreateFoodLogEntry } from '../types';
import { OFDLocals } from '../middlewares';

export function addHandlers(app: Express) {
  app.post('/logs', createFoodLogHandler)
}

type StoreFoodLogFunction = (userId: string, logEntry: CreateFoodLogEntry) => Promise<Result<string, StorageError>> 

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