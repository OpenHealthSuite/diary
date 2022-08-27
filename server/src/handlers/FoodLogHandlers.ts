import { Express, NextFunction, Request, Response } from 'express';
import { CreateFoodLogEntry } from '../types/FoodLogEntry';
import { OFDLocals } from '../types/Locals';

export function addHandlers(app: Express) {
  app.post('/logs', createFoodLogHandler)
}

type StoreFoodLogFunction = (userId: string, logEntry: CreateFoodLogEntry) => Promise<string> 

export function createFoodLogHandler(
  req: Request,
  res: Response & { locals: OFDLocals },
  next: NextFunction,
  storeFoodLog: StoreFoodLogFunction = {} as any
) {
  if (res.locals.userId) {
    storeFoodLog(res.locals.userId, req.body)
      .then(res.send)
  }
}